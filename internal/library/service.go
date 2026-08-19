// Package library composes ingestion, durable storage, and catalogue publication.
package library

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rudyon/malum/internal/catalog"
	"github.com/rudyon/malum/internal/ingest/webpage"
	authorstore "github.com/rudyon/malum/internal/storage/author"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

var (
	ErrImportFailed    = errors.New("document import failed")
	ErrStorageFailed   = errors.New("document storage failed")
	ErrCatalogueFailed = errors.New("document catalogue publication failed")
)

type Importer interface {
	Import(context.Context, string) (webpage.Result, error)
}

type Catalogue interface {
	AddDocument(context.Context, documentstore.Saved) (catalog.Document, error)
	ListDocuments(context.Context) ([]catalog.Document, error)
	GetDocument(context.Context, string) (catalog.Document, error)
	GetAuthor(context.Context, string) (catalog.Author, error)
	SetAuthorAvatar(context.Context, string, string, string) error
}

type DocumentStore interface {
	SaveWebpage(webpage.Result) (documentstore.Saved, error)
	Load(string) (documentstore.Manifest, error)
	OpenResource(string, string) (documentstore.OpenedResource, error)
}

type AuthorStore interface {
	SaveAvatar(string, string, string, []byte) (authorstore.SavedAvatar, error)
	OpenAvatar(string, string) (authorstore.OpenedAvatar, error)
}

type Service struct {
	importer  Importer
	catalogue Catalogue
	documents DocumentStore
	authors   AuthorStore
}

type ImportedDocument struct {
	Document    catalog.Document
	Warnings    []webpage.Warning
	Diagnostics []error
}

type ReaderDocument struct {
	Document catalog.Document
	Manifest documentstore.Manifest
}

func New(importer Importer, catalogue Catalogue, documents DocumentStore, authors AuthorStore) *Service {
	return &Service{importer: importer, catalogue: catalogue, documents: documents, authors: authors}
}

func (s *Service) ImportURL(ctx context.Context, rawURL string) (ImportedDocument, error) {
	if strings.TrimSpace(rawURL) == "" {
		return ImportedDocument{}, fmt.Errorf("import document: URL is required")
	}
	imported, err := s.importer.Import(ctx, rawURL)
	if err != nil {
		return ImportedDocument{}, fmt.Errorf("%w: %w", ErrImportFailed, err)
	}
	saved, err := s.documents.SaveWebpage(imported)
	if err != nil {
		return ImportedDocument{}, fmt.Errorf("%w: %w", ErrStorageFailed, err)
	}
	published, err := s.catalogue.AddDocument(ctx, saved)
	if err != nil {
		return ImportedDocument{}, fmt.Errorf("%w: %w", ErrCatalogueFailed, err)
	}

	warnings := append([]webpage.Warning(nil), imported.Document.Warnings...)
	var diagnostics []error
	if published.Author != nil && published.Author.AvatarPath == "" && len(imported.Document.AuthorCandidates) > 0 {
		candidate := imported.Document.AuthorCandidates[0]
		if len(candidate.ImageData) > 0 {
			avatar, avatarErr := s.authors.SaveAvatar(
				published.Author.ID,
				candidate.ImageURL,
				candidate.ImageContentType,
				candidate.ImageData,
			)
			if avatarErr == nil {
				avatarErr = s.catalogue.SetAuthorAvatar(ctx, published.Author.ID, avatar.SourceURL, avatar.Path)
			}
			if avatarErr != nil {
				diagnostics = append(diagnostics, fmt.Errorf("store imported author avatar: %w", avatarErr))
				warnings = append(warnings, webpage.Warning{
					Code:    "author-avatar-storage-failed",
					URL:     candidate.ImageURL,
					Message: "The article was saved, but its author avatar could not be stored.",
				})
			} else {
				published.Author.AvatarSourceURL = avatar.SourceURL
				published.Author.AvatarPath = avatar.Path
			}
		}
	}
	return ImportedDocument{Document: published, Warnings: warnings, Diagnostics: diagnostics}, nil
}

func (s *Service) ListDocuments(ctx context.Context) ([]catalog.Document, error) {
	return s.catalogue.ListDocuments(ctx)
}

func (s *Service) GetDocument(ctx context.Context, documentID string) (ReaderDocument, error) {
	document, err := s.catalogue.GetDocument(ctx, documentID)
	if err != nil {
		return ReaderDocument{}, err
	}
	manifest, err := s.documents.Load(documentID)
	if err != nil {
		return ReaderDocument{}, fmt.Errorf("load reader document: %w", err)
	}
	return ReaderDocument{Document: document, Manifest: manifest}, nil
}

func (s *Service) OpenResource(documentID, filename string) (documentstore.OpenedResource, error) {
	return s.documents.OpenResource(documentID, filename)
}

func (s *Service) OpenAvatar(ctx context.Context, authorID string) (authorstore.OpenedAvatar, error) {
	author, err := s.catalogue.GetAuthor(ctx, authorID)
	if err != nil {
		return authorstore.OpenedAvatar{}, err
	}
	if author.AvatarPath == "" {
		return authorstore.OpenedAvatar{}, fmt.Errorf("%w: %s", authorstore.ErrAvatarNotFound, authorID)
	}
	return s.authors.OpenAvatar(authorID, author.AvatarPath)
}
