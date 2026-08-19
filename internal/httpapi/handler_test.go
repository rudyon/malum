package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rudyon/malum/internal/catalog"
	"github.com/rudyon/malum/internal/ingest/webpage"
	"github.com/rudyon/malum/internal/library"
	"github.com/rudyon/malum/internal/safefetch"
	authorstore "github.com/rudyon/malum/internal/storage/author"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

const (
	testDocumentID = "12345678-1234-4234-8234-123456789abc"
	testAuthorID   = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

func TestListDocumentsProjectsPublicURLsWithoutFakeProgress(t *testing.T) {
	service := &fakeLibrary{documents: []catalog.Document{testDocument()}}
	response := performRequest(t, service, http.MethodGet, "/api/documents", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, wanted := range []string{`"displayName":"alice maz"`, `"handle":"alicemaz"`, `"thumbnailUrl":"/api/documents/` + testDocumentID + `/resources/cover.png"`, `"avatarUrl":"/api/authors/` + testAuthorID + `/avatar"`} {
		if !strings.Contains(body, wanted) {
			t.Errorf("response does not contain %s: %s", wanted, body)
		}
	}
	for _, forbidden := range []string{"progress", "thumbnailPath", "authors/" + testAuthorID + "/avatars/avatar.png"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("response leaked %q: %s", forbidden, body)
		}
	}
}

func TestGetDocumentProjectsStoredBlockImages(t *testing.T) {
	service := &fakeLibrary{reader: library.ReaderDocument{
		Document: testDocument(),
		Manifest: documentstore.Manifest{
			Article: documentstore.Article{
				Blocks:  []webpage.Block{{Kind: webpage.BlockImage, Image: &webpage.Image{URL: "https://example.test/cover.png"}}},
				Outline: []webpage.OutlineItem{{ID: "start", Level: 2, Title: "Start"}},
			},
			Resources: []documentstore.Resource{{
				SourceURL: "https://example.test/cover.png",
				Status:    documentstore.ResourceStored,
				Path:      "resources/cover.png",
			}},
		},
	}}
	response := performRequest(t, service, http.MethodGet, "/api/documents/"+testDocumentID, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"url":"/api/documents/`+testDocumentID+`/resources/cover.png"`) {
		t.Fatalf("stored image URL was not projected: %s", body)
	}
	if !strings.Contains(body, `"outline":[{"id":"start","level":2,"title":"Start"}]`) {
		t.Fatalf("outline missing: %s", body)
	}
}

func TestImportDocumentValidatesJSONAndReturnsCreatedDocument(t *testing.T) {
	service := &fakeLibrary{imported: library.ImportedDocument{Document: testDocument()}}
	response := performRequest(t, service, http.MethodPost, "/api/documents", strings.NewReader(`{"url":"https://example.test/article"}`))
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.importURL != "https://example.test/article" {
		t.Fatalf("import URL = %q", service.importURL)
	}

	invalid := performRequest(t, service, http.MethodPost, "/api/documents", strings.NewReader(`{"url":"file:///secret"}`))
	if invalid.Code != http.StatusBadRequest || !strings.Contains(invalid.Body.String(), `"code":"invalid_url"`) {
		t.Fatalf("invalid response = %d %s", invalid.Code, invalid.Body.String())
	}
}

func TestImportFailureDoesNotExposeInternalError(t *testing.T) {
	service := &fakeLibrary{importErr: errors.Join(library.ErrImportFailed, errors.New(`open C:\private\document: denied`))}
	response := performRequest(t, service, http.MethodPost, "/api/documents", strings.NewReader(`{"url":"https://example.test/article"}`))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "private") {
		t.Fatalf("response exposed internal error: %s", response.Body.String())
	}
}

func TestImportRejectsPrivateNetworkDestinationClearly(t *testing.T) {
	service := &fakeLibrary{importErr: errors.Join(library.ErrImportFailed, safefetch.ErrPrivateDestination)}
	response := performRequest(t, service, http.MethodPost, "/api/documents", strings.NewReader(`{"url":"http://127.0.0.1/article"}`))
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"private_network_url"`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func performRequest(t *testing.T, service Library, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, body)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	New(service, logger).ServeHTTP(response, request)
	return response
}

func testDocument() catalog.Document {
	return catalog.Document{
		ID:                 testDocumentID,
		ReadingKind:        "article",
		AcquisitionMethod:  "url",
		OriginalFormat:     "html",
		SourceURL:          "https://example.test/article",
		Title:              "playing to win",
		Description:        "A description.",
		SiteName:           "Example",
		Author:             &catalog.Author{ID: testAuthorID, DisplayName: "alice maz", Handle: "alicemaz", AvatarPath: "authors/" + testAuthorID + "/avatars/avatar.png"},
		WordCount:          1200,
		ReadingTimeMinutes: 6,
		ThumbnailPath:      "resources/cover.png",
		SavedAt:            time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC),
	}
}

type fakeLibrary struct {
	documents []catalog.Document
	reader    library.ReaderDocument
	imported  library.ImportedDocument
	importErr error
	importURL string
}

func (f *fakeLibrary) ImportURL(_ context.Context, rawURL string) (library.ImportedDocument, error) {
	f.importURL = rawURL
	return f.imported, f.importErr
}

func (f *fakeLibrary) ListDocuments(context.Context) ([]catalog.Document, error) {
	return f.documents, nil
}

func (f *fakeLibrary) GetDocument(context.Context, string) (library.ReaderDocument, error) {
	return f.reader, nil
}

func (f *fakeLibrary) OpenResource(string, string) (documentstore.OpenedResource, error) {
	return documentstore.OpenedResource{}, documentstore.ErrResourceNotFound
}

func (f *fakeLibrary) OpenAvatar(context.Context, string) (authorstore.OpenedAvatar, error) {
	return authorstore.OpenedAvatar{}, authorstore.ErrAvatarNotFound
}
