package webpage

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestImporterPreservesHTMLAndFetchesImages(t *testing.T) {
	originalHTML, err := os.ReadFile("testdata/article.html")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/start":
			http.Redirect(writer, request, "/essays/quiet-weather", http.StatusFound)
		case "/essays/quiet-weather":
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = writer.Write(originalHTML)
		case "/images/station.jpg":
			writer.Header().Set("Content-Type", "image/jpeg")
			_, _ = writer.Write([]byte("local-image-bytes"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	result, err := NewImporter(server.Client()).Import(context.Background(), server.URL+"/start")
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.RequestedURL != server.URL+"/start" {
		t.Fatalf("requested URL = %q", result.Snapshot.RequestedURL)
	}
	if result.Snapshot.FinalURL != server.URL+"/essays/quiet-weather" {
		t.Fatalf("final URL = %q", result.Snapshot.FinalURL)
	}
	if string(result.Snapshot.OriginalHTML) != string(originalHTML) {
		t.Fatal("original HTML was not preserved byte-for-byte")
	}

	var fetchedImage bool
	for _, resource := range result.Document.Resources {
		if strings.HasSuffix(resource.URL, "/images/station.jpg") {
			fetchedImage = resource.ContentType == "image/jpeg" && string(resource.Data) == "local-image-bytes"
		}
	}
	if !fetchedImage {
		t.Fatalf("image resources = %#v", result.Document.Resources)
	}
	if len(result.Document.Warnings) != 1 || !strings.HasSuffix(result.Document.Warnings[0].URL, "/images/missing.jpg") {
		t.Fatalf("warnings = %#v", result.Document.Warnings)
	}
}

func TestImporterRejectsNonHTMLResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/pdf")
		_, _ = writer.Write([]byte("not a PDF"))
	}))
	defer server.Close()

	_, err := NewImporter(server.Client()).Import(context.Background(), server.URL)
	if err == nil || !strings.Contains(err.Error(), "unsupported content type") {
		t.Fatalf("error = %v", err)
	}
}

func TestImporterEnforcesPageLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte("0123456789"))
	}))
	defer server.Close()

	importer := NewImporter(server.Client())
	importer.MaxPageBytes = 4
	_, err := importer.Import(context.Background(), server.URL)
	if !errors.Is(err, ErrPageTooLarge) {
		t.Fatalf("error = %v, want ErrPageTooLarge", err)
	}
}
