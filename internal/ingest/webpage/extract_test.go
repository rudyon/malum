package webpage

import (
	"errors"
	"os"
	"testing"
)

func TestExtractProjectsReadableArticle(t *testing.T) {
	originalHTML, err := os.ReadFile("testdata/article.html")
	if err != nil {
		t.Fatal(err)
	}

	document, err := Extract("https://fieldnotes.example/essays/quiet-weather", originalHTML)
	if err != nil {
		t.Fatal(err)
	}

	if document.Title != "A Map of Quiet Weather" {
		t.Fatalf("title = %q", document.Title)
	}
	if document.Byline != "Iris Bell" {
		t.Fatalf("byline = %q", document.Byline)
	}
	if document.SiteName != "Field Notes" {
		t.Fatalf("site name = %q", document.SiteName)
	}
	if document.PublishedAt == nil || document.PublishedAt.Format("2006-01-02") != "2025-01-14" {
		t.Fatalf("published time = %v", document.PublishedAt)
	}
	if document.WordCount < 150 {
		t.Fatalf("word count = %d", document.WordCount)
	}
	if document.ReadingTimeMinutes < 1 {
		t.Fatalf("reading time = %d", document.ReadingTimeMinutes)
	}
	if document.ContentHTML == "" {
		t.Fatal("cleaned HTML is empty")
	}

	wantKinds := map[BlockKind]bool{
		BlockParagraph:    false,
		BlockHeading:      false,
		BlockImage:        false,
		BlockList:         false,
		BlockDefinitions:  false,
		BlockQuote:        false,
		BlockPreformatted: false,
		BlockHTML:         false,
	}
	for _, block := range document.Blocks {
		if _, wanted := wantKinds[block.Kind]; wanted {
			wantKinds[block.Kind] = true
		}
	}
	for kind, found := range wantKinds {
		if !found {
			t.Errorf("missing projected block kind %q", kind)
		}
	}

	if len(document.Outline) != 3 {
		t.Fatalf("outline length = %d, want 3", len(document.Outline))
	}
	if document.Outline[0].ID != "first-observations" ||
		document.Outline[1].ID != "what-the-gauges-missed" ||
		document.Outline[2].ID != "first-observations-2" {
		t.Fatalf("outline IDs = %#v", document.Outline)
	}

	var foundCaptionedImage bool
	for _, block := range document.Blocks {
		if block.Image != nil && block.Image.URL == "https://fieldnotes.example/images/station.jpg" {
			foundCaptionedImage = block.Image.Caption == "The north station after the first long snow."
		}
	}
	if !foundCaptionedImage {
		t.Fatal("captioned relative image was not projected and resolved")
	}
	if len(document.Resources) != 2 {
		t.Fatalf("resources = %#v", document.Resources)
	}
}

func TestExtractRejectsInvalidURL(t *testing.T) {
	_, err := Extract("file:///tmp/article.html", []byte("<html></html>"))
	if err == nil {
		t.Fatal("expected invalid URL error")
	}
}

func TestExtractRejectsUnreadableDocument(t *testing.T) {
	_, err := Extract("https://example.com/", []byte("<html><head><title>Empty</title></head><body></body></html>"))
	if !errors.Is(err, ErrNoReadableContent) {
		t.Fatalf("error = %v, want ErrNoReadableContent", err)
	}
}
