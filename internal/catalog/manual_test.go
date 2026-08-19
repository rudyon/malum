package catalog_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rudyon/malum/internal/catalog"
	"github.com/rudyon/malum/internal/ingest/webpage"
	authorstore "github.com/rudyon/malum/internal/storage/author"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

func TestManualMilestonePublication(t *testing.T) {
	if os.Getenv("MALUM_MILESTONE_TEST") == "" {
		t.Skip("set MALUM_MILESTONE_TEST to run the two live milestone imports")
	}

	ctx := context.Background()
	root := t.TempDir()
	catalogue, err := catalog.Open(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	defer catalogue.Close()

	client := &http.Client{Timeout: 90 * time.Second}
	importer := webpage.NewImporter(client)
	documents := documentstore.New(root)
	avatars := authorstore.New(root)
	targets := []string{
		"https://www.alicemaz.com/writing/minecraft.html",
		"https://alicemaz.substack.com/p/one-with-the-machine",
	}

	var authorID string
	for _, target := range targets {
		imported, err := importer.Import(ctx, target)
		if err != nil {
			t.Fatalf("import %s: %v", target, err)
		}
		saved, err := documents.SaveWebpage(imported)
		if err != nil {
			t.Fatalf("store %s: %v", target, err)
		}
		published, err := catalogue.AddDocument(ctx, saved)
		if err != nil {
			t.Fatalf("catalogue %s: %v", target, err)
		}
		if published.Author == nil {
			t.Fatalf("%s has no resolved author", target)
		}
		if authorID == "" {
			authorID = published.Author.ID
		} else if published.Author.ID != authorID {
			t.Fatalf("targets resolved to different authors: %s and %s", authorID, published.Author.ID)
		}
		if published.Author.DisplayName != "alice maz" || published.Author.Handle != "alicemaz" {
			t.Fatalf("author = %#v", published.Author)
		}

		if len(imported.Document.AuthorCandidates) > 0 {
			candidate := imported.Document.AuthorCandidates[0]
			if len(candidate.ImageData) > 0 {
				avatar, err := avatars.SaveAvatar(published.Author.ID, candidate.ImageURL, candidate.ImageContentType, candidate.ImageData)
				if err != nil {
					t.Fatalf("store avatar for %s: %v", target, err)
				}
				if err := catalogue.SetAuthorAvatar(ctx, published.Author.ID, avatar.SourceURL, avatar.Path); err != nil {
					t.Fatalf("catalogue avatar for %s: %v", target, err)
				}
			}
		}
		t.Logf("published title=%q author=@%s resources=%d warnings=%d", published.Title, published.Author.Handle, len(imported.Document.Resources), len(imported.Document.Warnings))
	}

	listed, err := catalogue.ListDocuments(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 {
		t.Fatalf("catalogue documents = %d, want 2", len(listed))
	}
	for _, document := range listed {
		if document.Author == nil || document.Author.ID != authorID {
			t.Fatalf("listed document author = %#v", document.Author)
		}
	}
	if listed[0].Author.AvatarPath == "" {
		t.Fatal("structured Substack avatar was not stored")
	}
}
