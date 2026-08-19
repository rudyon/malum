package catalog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rudyon/malum/internal/ingest/webpage"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

const (
	firstDocumentID  = "11111111-1111-4111-8111-111111111111"
	secondDocumentID = "22222222-2222-4222-8222-222222222222"
	fixedAuthorID    = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

func TestCataloguePublishesDocumentsAndResolvesAuthor(t *testing.T) {
	ctx := context.Background()
	catalogue, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer catalogue.Close()
	catalogue.newID = func() (string, error) { return fixedAuthorID, nil }

	first := testSavedDocument(firstDocumentID, time.Date(2026, 8, 19, 9, 0, 0, 0, time.UTC))
	first.Manifest.Article.Byline = "alice maz"
	first.Manifest.Article.AuthorCandidates = []webpage.AuthorCandidate{
		{Name: "alice maz", Evidence: "readability-byline"},
	}
	firstDocument, err := catalogue.AddDocument(ctx, first)
	if err != nil {
		t.Fatalf("AddDocument(first) error = %v", err)
	}
	if firstDocument.Author == nil || firstDocument.Author.ID != fixedAuthorID || firstDocument.Author.DisplayName != "alice maz" || firstDocument.Author.Handle != "alicemaz" {
		t.Fatalf("first author = %#v", firstDocument.Author)
	}

	catalogue.newID = func() (string, error) {
		t.Fatal("second document created a duplicate author")
		return "", nil
	}
	second := testSavedDocument(secondDocumentID, time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC))
	second.Manifest.Article.Title = "one with the machine"
	second.Manifest.Article.Byline = "alice maz"
	second.Manifest.Article.AuthorCandidates = []webpage.AuthorCandidate{
		{
			Name:       "alice maz",
			ImageURL:   "https://cdn.example/alice.jpg",
			ProfileURL: "https://substack.com/@alicemaz",
			Evidence:   "json-ld",
			Identities: []webpage.AuthorIdentity{
				{Kind: "profile-url", Value: "https://substack.com/@alicemaz"},
				{Kind: "identifier", Value: "user:103721713"},
				{Kind: "same-as", Value: "https://twitter.com/alicemazzy"},
			},
		},
	}
	secondDocument, err := catalogue.AddDocument(ctx, second)
	if err != nil {
		t.Fatalf("AddDocument(second) error = %v", err)
	}
	if secondDocument.Author == nil || secondDocument.Author.ID != fixedAuthorID || secondDocument.Author.DisplayName != "alice maz" || secondDocument.Author.Handle != "alicemaz" {
		t.Fatalf("second author = %#v", secondDocument.Author)
	}
	if secondDocument.Author.AvatarSourceURL != "https://cdn.example/alice.jpg" {
		t.Fatalf("avatar source = %q", secondDocument.Author.AvatarSourceURL)
	}

	var identities int
	if err := catalogue.database.QueryRowContext(ctx, "SELECT COUNT(*) FROM author_identities WHERE author_id = ?", fixedAuthorID).Scan(&identities); err != nil {
		t.Fatal(err)
	}
	if identities != 3 {
		t.Fatalf("author identities = %d, want 3", identities)
	}

	avatarPath := "authors/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/avatars/hash.jpg"
	if err := catalogue.SetAuthorAvatar(ctx, fixedAuthorID, "https://cdn.example/alice.jpg", avatarPath); err != nil {
		t.Fatalf("SetAuthorAvatar() error = %v", err)
	}
	if err := catalogue.SetAuthorAvatar(ctx, fixedAuthorID, "", "documents/unrelated/image.jpg"); err == nil {
		t.Fatal("SetAuthorAvatar() accepted a path outside the author's avatar directory")
	}
	storedSecond, err := catalogue.GetDocument(ctx, secondDocumentID)
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if storedSecond.Author == nil || storedSecond.Author.AvatarPath != avatarPath {
		t.Fatalf("stored author = %#v", storedSecond.Author)
	}

	documents, err := catalogue.ListDocuments(ctx)
	if err != nil {
		t.Fatalf("ListDocuments() error = %v", err)
	}
	if len(documents) != 2 || documents[0].ID != secondDocumentID || documents[1].ID != firstDocumentID {
		t.Fatalf("documents = %#v", documents)
	}
	if documents[0].ThumbnailPath != "resources/lead.png" {
		t.Fatalf("thumbnail path = %q", documents[0].ThumbnailPath)
	}

	if _, err := catalogue.AddDocument(ctx, second); !errors.Is(err, ErrDocumentExists) {
		t.Fatalf("duplicate AddDocument() error = %v", err)
	}
}

func TestCataloguePreservesDisplayNameAndHandleCasing(t *testing.T) {
	ctx := context.Background()
	catalogue, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer catalogue.Close()
	catalogue.newID = func() (string, error) { return fixedAuthorID, nil }

	saved := testSavedDocument(firstDocumentID, time.Now())
	saved.Manifest.Article.AuthorCandidates = []webpage.AuthorCandidate{{Name: "Alice Maz", Evidence: "readability-byline"}}
	document, err := catalogue.AddDocument(ctx, saved)
	if err != nil {
		t.Fatal(err)
	}
	if document.Author == nil || document.Author.DisplayName != "Alice Maz" || document.Author.Handle != "AliceMaz" {
		t.Fatalf("author = %#v", document.Author)
	}
}

func TestCatalogueLeavesMissingAuthorNull(t *testing.T) {
	ctx := context.Background()
	catalogue, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer catalogue.Close()

	saved := testSavedDocument(firstDocumentID, time.Now())
	document, err := catalogue.AddDocument(ctx, saved)
	if err != nil {
		t.Fatal(err)
	}
	if document.Author != nil {
		t.Fatalf("author = %#v, want nil", document.Author)
	}
}

func TestOpenAppliesInitialMigrationAndForeignKeys(t *testing.T) {
	ctx := context.Background()
	catalogue, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer catalogue.Close()

	var version, foreignKeys int
	if err := catalogue.database.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err := catalogue.database.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if version != 1 || foreignKeys != 1 {
		t.Fatalf("schema version = %d, foreign_keys = %d", version, foreignKeys)
	}
}

func testSavedDocument(id string, storedAt time.Time) documentstore.Saved {
	return documentstore.Saved{
		ID: id,
		Manifest: documentstore.Manifest{
			SchemaVersion: documentstore.SchemaVersion,
			DocumentID:    id,
			ReadingKind:   documentstore.ReadingKindArticle,
			StoredAt:      storedAt,
			Acquisition: documentstore.Acquisition{
				Method:       documentstore.AcquisitionURL,
				RequestedURL: "https://example.test/start",
				FinalURL:     "https://example.test/article",
			},
			Original: documentstore.Original{Format: documentstore.OriginalFormatHTML},
			Article: documentstore.Article{
				Title:              "An article",
				SiteName:           "Example Test",
				Excerpt:            "An article description.",
				WordCount:          750,
				ReadingTimeMinutes: 3,
			},
			Resources: []documentstore.Resource{
				{Role: "lead-image", Status: documentstore.ResourceStored, Path: "resources/lead.png"},
			},
		},
	}
}
