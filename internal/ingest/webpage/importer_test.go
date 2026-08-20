package webpage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
		case "/images/iris.jpg":
			writer.Header().Set("Content-Type", "image/jpeg")
			_, _ = writer.Write([]byte("author-image-bytes"))
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
	if len(result.Document.AuthorCandidates) != 1 || result.Document.AuthorCandidates[0].ImageContentType != "image/jpeg" || string(result.Document.AuthorCandidates[0].ImageData) != "author-image-bytes" {
		t.Fatalf("author candidates = %#v", result.Document.AuthorCandidates)
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

func TestImporterResolvesSubstackProfilePostToPublicArticle(t *testing.T) {
	articleHTML, err := os.ReadFile("testdata/article.html")
	if err != nil {
		t.Fatal(err)
	}
	const (
		sharedURL    = "https://substack.com/@alicemaz/p-113366829"
		canonicalURL = "https://alicemaz.substack.com/p/one-with-the-machine"
	)
	preloads, err := json.Marshal(map[string]any{
		"feedData": map[string]any{
			"initialPost": map[string]any{
				"post": map[string]any{"id": 113366829, "canonical_url": canonicalURL},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	shellHTML := `<html><body><script>window._preloads = JSON.parse(` + strconv.Quote(string(preloads)) + `);</script></body></html>`

	var fetched []string
	client := doFunc(func(request *http.Request) (*http.Response, error) {
		fetched = append(fetched, request.URL.String())
		statusCode := http.StatusOK
		contentType := "text/html; charset=utf-8"
		body := shellHTML
		switch request.URL.String() {
		case sharedURL:
		case canonicalURL:
			body = string(articleHTML)
		default:
			statusCode = http.StatusNotFound
			contentType = "text/plain"
			body = "not found"
		}
		return &http.Response{
			StatusCode: statusCode,
			Status:     http.StatusText(statusCode),
			Header:     http.Header{"Content-Type": []string{contentType}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})

	result, err := NewImporter(client).Import(context.Background(), sharedURL)
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.RequestedURL != sharedURL || result.Snapshot.FinalURL != canonicalURL {
		t.Fatalf("snapshot URLs = requested %q, final %q", result.Snapshot.RequestedURL, result.Snapshot.FinalURL)
	}
	if string(result.Snapshot.OriginalHTML) != string(articleHTML) {
		t.Fatal("resolved article HTML was not preserved")
	}
	if len(fetched) < 2 || fetched[0] != sharedURL || fetched[1] != canonicalURL {
		t.Fatalf("fetched URLs = %#v", fetched)
	}
}

type doFunc func(*http.Request) (*http.Response, error)

func (do doFunc) Do(request *http.Request) (*http.Response, error) {
	return do(request)
}
