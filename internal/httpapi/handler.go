// Package httpapi exposes Malum's application operations over HTTP and JSON.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/rudyon/malum/internal/catalog"
	"github.com/rudyon/malum/internal/identifier"
	"github.com/rudyon/malum/internal/ingest/webpage"
	"github.com/rudyon/malum/internal/library"
	"github.com/rudyon/malum/internal/safefetch"
	authorstore "github.com/rudyon/malum/internal/storage/author"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

const maxRequestBody = 8 << 10

type Library interface {
	ImportURL(context.Context, string) (library.ImportedDocument, error)
	ListDocuments(context.Context) ([]catalog.Document, error)
	GetDocument(context.Context, string) (library.ReaderDocument, error)
	OpenResource(string, string) (documentstore.OpenedResource, error)
	OpenAvatar(context.Context, string) (authorstore.OpenedAvatar, error)
}

type handler struct {
	library Library
	logger  *slog.Logger
}

func New(service Library, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	h := &handler{library: service, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/documents", h.importDocument)
	mux.HandleFunc("GET /api/documents", h.listDocuments)
	mux.HandleFunc("GET /api/documents/{documentID}", h.getDocument)
	mux.HandleFunc("GET /api/documents/{documentID}/resources/{filename}", h.getResource)
	mux.HandleFunc("GET /api/authors/{authorID}/avatar", h.getAvatar)
	return securityHeaders(mux)
}

func (h *handler) importDocument(response http.ResponseWriter, request *http.Request) {
	var input struct {
		URL string `json:"url"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, maxRequestBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", "The request must contain one URL.")
		return
	}
	if err := ensureJSONEnd(decoder); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", "The request must contain one JSON object.")
		return
	}
	if !validImportURL(input.URL) {
		writeError(response, http.StatusBadRequest, "invalid_url", "Enter a valid HTTP or HTTPS URL.")
		return
	}

	result, err := h.library.ImportURL(request.Context(), input.URL)
	if err != nil {
		if request.Context().Err() != nil || errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, safefetch.ErrPrivateDestination) {
			writeError(response, http.StatusBadRequest, "private_network_url", "Malum does not import from private or local network addresses.")
			return
		}
		if errors.Is(err, library.ErrImportFailed) {
			h.logger.Warn("document import failed", "error", err)
			writeError(response, http.StatusUnprocessableEntity, "document_import_failed", "Malum could not import a readable article from this URL.")
			return
		}
		h.logger.Error("document import failed", "error", err)
		writeError(response, http.StatusInternalServerError, "internal_error", "Malum could not finish saving this document.")
		return
	}
	for _, diagnostic := range result.Diagnostics {
		h.logger.Warn("document imported with a non-fatal problem", "error", diagnostic)
	}
	writeJSON(response, http.StatusCreated, map[string]any{
		"document": documentDTO(result.Document, nil),
		"warnings": append([]webpage.Warning{}, result.Warnings...),
	})
}

func (h *handler) listDocuments(response http.ResponseWriter, request *http.Request) {
	documents, err := h.library.ListDocuments(request.Context())
	if err != nil {
		h.internalError(response, request, "list documents", err)
		return
	}
	items := make([]documentResponse, 0, len(documents))
	for _, document := range documents {
		items = append(items, documentDTO(document, nil))
	}
	writeJSON(response, http.StatusOK, map[string]any{"documents": items})
}

func (h *handler) getDocument(response http.ResponseWriter, request *http.Request) {
	documentID := request.PathValue("documentID")
	if !identifier.IsUUID(documentID) {
		writeError(response, http.StatusNotFound, "document_not_found", "This document is not in the library.")
		return
	}
	readerDocument, err := h.library.GetDocument(request.Context(), documentID)
	if err != nil {
		if errors.Is(err, catalog.ErrDocumentNotFound) {
			writeError(response, http.StatusNotFound, "document_not_found", "This document is not in the library.")
			return
		}
		h.internalError(response, request, "load document", err)
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{
		"document": documentDTO(readerDocument.Document, &readerDocument.Manifest),
	})
}

func (h *handler) getResource(response http.ResponseWriter, request *http.Request) {
	documentID := request.PathValue("documentID")
	filename := request.PathValue("filename")
	if !identifier.IsUUID(documentID) {
		writeError(response, http.StatusNotFound, "resource_not_found", "This document resource is unavailable.")
		return
	}
	resource, err := h.library.OpenResource(documentID, filename)
	if err != nil {
		if errors.Is(err, documentstore.ErrDocumentNotFound) || errors.Is(err, documentstore.ErrResourceNotFound) {
			writeError(response, http.StatusNotFound, "resource_not_found", "This document resource is unavailable.")
			return
		}
		h.internalError(response, request, "open document resource", err)
		return
	}
	defer resource.File.Close()
	serveAsset(response, request, filename, resource.ContentType, resource.File)
}

func (h *handler) getAvatar(response http.ResponseWriter, request *http.Request) {
	authorID := request.PathValue("authorID")
	if !identifier.IsUUID(authorID) {
		writeError(response, http.StatusNotFound, "avatar_not_found", "This author avatar is unavailable.")
		return
	}
	avatar, err := h.library.OpenAvatar(request.Context(), authorID)
	if err != nil {
		if errors.Is(err, catalog.ErrAuthorNotFound) || errors.Is(err, authorstore.ErrAvatarNotFound) {
			writeError(response, http.StatusNotFound, "avatar_not_found", "This author avatar is unavailable.")
			return
		}
		h.internalError(response, request, "open author avatar", err)
		return
	}
	defer avatar.File.Close()
	serveAsset(response, request, path.Base(avatar.File.Name()), avatar.ContentType, avatar.File)
}

func (h *handler) internalError(response http.ResponseWriter, request *http.Request, operation string, err error) {
	if request.Context().Err() != nil || errors.Is(err, context.Canceled) {
		return
	}
	h.logger.Error(operation, "error", err)
	writeError(response, http.StatusInternalServerError, "internal_error", "Malum could not complete this request.")
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		next.ServeHTTP(response, request)
	})
}

func serveAsset(response http.ResponseWriter, request *http.Request, name, contentType string, content io.ReadSeeker) {
	response.Header().Set("Content-Type", contentType)
	if contentType == "image/svg+xml" {
		response.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	}
	http.ServeContent(response, request, name, time.Time{}, content)
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("additional JSON value")
		}
		return err
	}
	return nil
}

func validImportURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != ""
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	if err := json.NewEncoder(response).Encode(value); err != nil {
		slog.Default().Error("write JSON response", "error", err)
	}
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

type documentResponse struct {
	ID                 string           `json:"id"`
	ReadingKind        string           `json:"readingKind"`
	AcquisitionMethod  string           `json:"acquisitionMethod"`
	OriginalFormat     string           `json:"originalFormat"`
	Title              string           `json:"title"`
	Description        string           `json:"description,omitempty"`
	Source             sourceResponse   `json:"source"`
	Author             *authorResponse  `json:"author"`
	Language           string           `json:"language,omitempty"`
	PublishedAt        *time.Time       `json:"publishedAt,omitempty"`
	SourceModifiedAt   *time.Time       `json:"sourceModifiedAt,omitempty"`
	WordCount          int              `json:"wordCount"`
	ReadingTimeMinutes int              `json:"readingTimeMinutes"`
	ThumbnailURL       string           `json:"thumbnailUrl,omitempty"`
	SavedAt            time.Time        `json:"savedAt"`
	Content            *contentResponse `json:"content,omitempty"`
}

type sourceResponse struct {
	URL      string `json:"url"`
	SiteName string `json:"siteName"`
}

type authorResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Handle      string `json:"handle"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
}

type contentResponse struct {
	Blocks  []webpage.Block       `json:"blocks"`
	Outline []webpage.OutlineItem `json:"outline"`
}

func documentDTO(document catalog.Document, manifest *documentstore.Manifest) documentResponse {
	response := documentResponse{
		ID:                 document.ID,
		ReadingKind:        document.ReadingKind,
		AcquisitionMethod:  document.AcquisitionMethod,
		OriginalFormat:     document.OriginalFormat,
		Title:              document.Title,
		Description:        document.Description,
		Source:             sourceResponse{URL: document.SourceURL, SiteName: document.SiteName},
		Language:           document.Language,
		PublishedAt:        document.PublishedAt,
		SourceModifiedAt:   document.SourceModifiedAt,
		WordCount:          document.WordCount,
		ReadingTimeMinutes: document.ReadingTimeMinutes,
		SavedAt:            document.SavedAt,
	}
	if document.Author != nil {
		response.Author = &authorResponse{
			ID:          document.Author.ID,
			DisplayName: document.Author.DisplayName,
			Handle:      document.Author.Handle,
		}
		if document.Author.AvatarPath != "" {
			response.Author.AvatarURL = fmt.Sprintf("/api/authors/%s/avatar", document.Author.ID)
		}
	}
	if document.ThumbnailPath != "" {
		response.ThumbnailURL = resourceURL(document.ID, document.ThumbnailPath)
	}
	if manifest != nil {
		blocks := append([]webpage.Block{}, manifest.Article.Blocks...)
		stored := make(map[string]string)
		for _, resource := range manifest.Resources {
			if resource.Status == documentstore.ResourceStored {
				stored[resource.SourceURL] = resourceURL(document.ID, resource.Path)
			}
		}
		for index := range blocks {
			if blocks[index].Image != nil {
				imageCopy := *blocks[index].Image
				if localURL := stored[imageCopy.URL]; localURL != "" {
					imageCopy.URL = localURL
				}
				blocks[index].Image = &imageCopy
			}
		}
		response.Content = &contentResponse{
			Blocks:  blocks,
			Outline: append([]webpage.OutlineItem{}, manifest.Article.Outline...),
		}
	}
	return response
}

func resourceURL(documentID, relativePath string) string {
	return fmt.Sprintf("/api/documents/%s/resources/%s", documentID, url.PathEscape(path.Base(relativePath)))
}
