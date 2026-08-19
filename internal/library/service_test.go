package library_test

import (
	"context"
	"testing"

	"github.com/rudyon/malum/internal/catalog"
	"github.com/rudyon/malum/internal/ingest/webpage"
	"github.com/rudyon/malum/internal/library"
	authorstore "github.com/rudyon/malum/internal/storage/author"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

func TestImportURLStoresCataloguesAndPublishesAuthorAvatar(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	catalogue, err := catalog.Open(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	defer catalogue.Close()

	result := webpage.Result{
		Snapshot: webpage.Snapshot{
			RequestedURL: "https://example.test/article",
			FinalURL:     "https://example.test/article",
			ContentType:  "text/html",
			OriginalHTML: []byte("<!doctype html><title>Article</title><p>Body</p>"),
		},
		Document: webpage.Document{
			Title:       "Article",
			SiteName:    "Example",
			ContentHTML: "<p>Body</p>",
			Blocks:      []webpage.Block{{Kind: webpage.BlockParagraph, Text: "Body"}},
			AuthorCandidates: []webpage.AuthorCandidate{{
				Name:             "alice maz",
				ImageURL:         "https://example.test/avatar.png",
				ImageContentType: "image/png",
				ImageData:        []byte("avatar bytes"),
				Evidence:         "json-ld",
			}},
		},
	}
	documents := documentstore.New(root)
	authors := authorstore.New(root)
	service := library.New(staticImporter{result: result}, catalogue, documents, authors)

	imported, err := service.ImportURL(ctx, result.Snapshot.RequestedURL)
	if err != nil {
		t.Fatal(err)
	}
	if imported.Document.Author == nil {
		t.Fatal("published document has no author")
	}
	if imported.Document.Author.DisplayName != "alice maz" || imported.Document.Author.Handle != "alicemaz" {
		t.Fatalf("author = %#v", imported.Document.Author)
	}
	if imported.Document.Author.AvatarPath == "" {
		t.Fatal("author avatar was not published")
	}

	readerDocument, err := service.GetDocument(ctx, imported.Document.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(readerDocument.Manifest.Article.Blocks) != 1 || readerDocument.Manifest.Article.Blocks[0].Text != "Body" {
		t.Fatalf("reader blocks = %#v", readerDocument.Manifest.Article.Blocks)
	}
	avatar, err := service.OpenAvatar(ctx, imported.Document.Author.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer avatar.File.Close()
	if avatar.ContentType != "image/png" {
		t.Fatalf("avatar content type = %q", avatar.ContentType)
	}
}

type staticImporter struct {
	result webpage.Result
}

func (i staticImporter) Import(context.Context, string) (webpage.Result, error) {
	return i.result, nil
}
