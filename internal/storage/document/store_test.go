package document

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rudyon/malum/internal/ingest/webpage"
)

const fixedDocumentID = "12345678-1234-4234-8234-123456789abc"

var fixedStoredAt = time.Date(2026, time.August, 19, 10, 11, 12, 0, time.FixedZone("test", 3*60*60))

func TestSaveWebpageWritesRecoverableBundle(t *testing.T) {
	root := t.TempDir()
	store := fixedStore(root)
	result := storageTestResult()

	saved, err := store.SaveWebpage(result)
	if err != nil {
		t.Fatalf("SaveWebpage() error = %v", err)
	}

	expectedPath := filepath.Join(root, documentsDirectory, fixedDocumentID)
	if saved.ID != fixedDocumentID || saved.Path != expectedPath {
		t.Fatalf("SaveWebpage() = ID %q, path %q; want ID %q, path %q", saved.ID, saved.Path, fixedDocumentID, expectedPath)
	}

	original := readTestFile(t, expectedPath, originalFilename)
	if !bytes.Equal(original, result.Snapshot.OriginalHTML) {
		t.Fatal("original.html does not contain the exact imported response body")
	}

	imageHash := checksum(result.Document.Resources[0].Data)
	imagePath := filepath.ToSlash(filepath.Join(resourcesDirectory, imageHash+".png"))
	article := string(readTestFile(t, expectedPath, articleFilename))
	if strings.Count(article, `src="`+imagePath+`"`) != 2 {
		t.Fatalf("article.html did not rewrite both stored resources to %q:\n%s", imagePath, article)
	}
	if !strings.Contains(article, `src="https://cdn.example/missing.png"`) {
		t.Fatalf("article.html did not retain the unavailable remote resource:\n%s", article)
	}

	resourceEntries, err := os.ReadDir(filepath.Join(expectedPath, resourcesDirectory))
	if err != nil {
		t.Fatalf("ReadDir(resources) error = %v", err)
	}
	if len(resourceEntries) != 1 || resourceEntries[0].Name() != imageHash+".png" {
		t.Fatalf("resources = %v; want one content-hash file", resourceEntries)
	}
	if got := readTestFile(t, expectedPath, imagePath); !bytes.Equal(got, result.Document.Resources[0].Data) {
		t.Fatal("stored resource bytes differ from imported bytes")
	}

	manifestBytes := readTestFile(t, expectedPath, manifestFilename)
	if !bytes.HasSuffix(manifestBytes, []byte("\n")) {
		t.Fatal("document.json does not end with a newline")
	}
	if bytes.Contains(manifestBytes, []byte("contentHtml")) || bytes.Contains(manifestBytes, result.Document.Resources[0].Data) {
		t.Fatal("document.json contains representation or resource bodies")
	}
	if bytes.Contains(manifestBytes, result.Document.AuthorCandidates[0].ImageData) {
		t.Fatal("document.json contains downloaded author avatar bytes")
	}

	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode document.json: %v", err)
	}
	assertManifest(t, manifest, result, imageHash, imagePath, []byte(article))

	documentEntries, err := os.ReadDir(filepath.Join(root, documentsDirectory))
	if err != nil {
		t.Fatalf("ReadDir(documents) error = %v", err)
	}
	if len(documentEntries) != 1 || documentEntries[0].Name() != fixedDocumentID {
		t.Fatalf("documents directory = %v; want only the published bundle", documentEntries)
	}
}

func TestSaveWebpageDoesNotOverwriteExistingDocument(t *testing.T) {
	root := t.TempDir()
	store := fixedStore(root)
	first := storageTestResult()
	if _, err := store.SaveWebpage(first); err != nil {
		t.Fatalf("first SaveWebpage() error = %v", err)
	}

	second := storageTestResult()
	second.Snapshot.OriginalHTML = []byte("replacement")
	if _, err := store.SaveWebpage(second); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second SaveWebpage() error = %v; want existing-ID error", err)
	}

	stored := readTestFile(t, root, documentsDirectory, fixedDocumentID, originalFilename)
	if !bytes.Equal(stored, first.Snapshot.OriginalHTML) {
		t.Fatal("existing document was changed by a colliding save")
	}
}

func TestSaveWebpageRemovesStagingDirectoryAfterFailure(t *testing.T) {
	root := t.TempDir()
	store := fixedStore(root)
	result := storageTestResult()
	result.Document.Resources[0].ContentType = "application/octet-stream"

	if _, err := store.SaveWebpage(result); err == nil || !strings.Contains(err.Error(), "invalid image content type") {
		t.Fatalf("SaveWebpage() error = %v; want invalid content-type error", err)
	}

	entries, err := os.ReadDir(filepath.Join(root, documentsDirectory))
	if err != nil {
		t.Fatalf("ReadDir(documents) error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("failed save left files behind: %v", entries)
	}
}

func TestSaveWebpageValidatesBeforeCreatingStorage(t *testing.T) {
	root := t.TempDir()
	store := fixedStore(root)
	result := storageTestResult()
	result.Snapshot.OriginalHTML = nil

	if _, err := store.SaveWebpage(result); err == nil || !strings.Contains(err.Error(), "original HTML is empty") {
		t.Fatalf("SaveWebpage() error = %v; want empty-original error", err)
	}
	if _, err := os.Stat(filepath.Join(root, documentsDirectory)); !os.IsNotExist(err) {
		t.Fatalf("invalid result created the documents directory; Stat error = %v", err)
	}
}

func fixedStore(root string) *Store {
	store := New(root)
	store.now = func() time.Time { return fixedStoredAt }
	store.newID = func() (string, error) { return fixedDocumentID, nil }
	return store
}

func storageTestResult() webpage.Result {
	imageData := []byte("original fictional image bytes")
	return webpage.Result{
		Snapshot: webpage.Snapshot{
			RequestedURL: "https://example.test/story",
			FinalURL:     "https://example.test/writing/story",
			ContentType:  "text/html",
			OriginalHTML: []byte("<!doctype html>\r\n<title>A stored article</title>\r\n"),
		},
		Document: webpage.Document{
			SourceURL:          "https://example.test/writing/story",
			Title:              "A stored article",
			Byline:             "Mara Vale",
			SiteName:           "Example Test",
			Language:           "en",
			Excerpt:            "An original test article.",
			WordCount:          500,
			ReadingTimeMinutes: 2,
			LeadImageURL:       "https://cdn.example/lead.png",
			AuthorCandidates: []webpage.AuthorCandidate{
				{
					Name:             "mara vale",
					ImageURL:         "https://cdn.example/mara.png",
					Evidence:         "json-ld",
					ImageContentType: "image/png",
					ImageData:        []byte("fictional author avatar bytes"),
				},
			},
			ContentHTML: `<article>
  <img src="https://cdn.example/lead.png" alt="Lead">
  <p>Original fixture prose.</p>
  <img src="https://cdn.example/repeated.png" alt="Repeated">
  <img src="https://cdn.example/missing.png" alt="Missing">
</article>`,
			Blocks: []webpage.Block{
				{Kind: webpage.BlockImage, Image: &webpage.Image{URL: "https://cdn.example/lead.png", Alt: "Lead"}},
				{Kind: webpage.BlockParagraph, Text: "Original fixture prose."},
			},
			Outline: []webpage.OutlineItem{},
			Resources: []webpage.Resource{
				{URL: "https://cdn.example/lead.png", Role: "lead-image", ContentType: "image/png", Data: imageData},
				{URL: "https://cdn.example/repeated.png", Role: "content-image", ContentType: "image/png", Data: imageData},
				{URL: "https://cdn.example/missing.png", Role: "content-image"},
			},
			Warnings: []webpage.Warning{
				{Code: "resource-fetch-failed", URL: "https://cdn.example/missing.png", Message: "test resource unavailable"},
			},
		},
	}
}

func assertManifest(t *testing.T, manifest Manifest, result webpage.Result, imageHash, imagePath string, articleHTML []byte) {
	t.Helper()
	if manifest.SchemaVersion != SchemaVersion || manifest.DocumentID != fixedDocumentID || manifest.ReadingKind != ReadingKindArticle {
		t.Fatalf("manifest identity fields = %#v", manifest)
	}
	if !manifest.StoredAt.Equal(fixedStoredAt.UTC()) {
		t.Fatalf("manifest storedAt = %v; want %v", manifest.StoredAt, fixedStoredAt.UTC())
	}
	if manifest.Acquisition.Method != AcquisitionURL || manifest.Acquisition.RequestedURL != result.Snapshot.RequestedURL || manifest.Acquisition.FinalURL != result.Snapshot.FinalURL {
		t.Fatalf("manifest acquisition = %#v", manifest.Acquisition)
	}
	if manifest.Original.Format != OriginalFormatHTML || manifest.Original.Path != originalFilename || manifest.Original.SHA256 != checksum(result.Snapshot.OriginalHTML) || manifest.Original.Size != int64(len(result.Snapshot.OriginalHTML)) {
		t.Fatalf("manifest original = %#v", manifest.Original)
	}
	if manifest.Article.Path != articleFilename || manifest.Article.SHA256 != checksum(articleHTML) || manifest.Article.Size != int64(len(articleHTML)) || manifest.Article.Title != result.Document.Title {
		t.Fatalf("manifest article = %#v", manifest.Article)
	}
	if len(manifest.Article.Blocks) != len(result.Document.Blocks) || manifest.Article.Blocks[0].Image.URL != result.Document.Blocks[0].Image.URL {
		t.Fatalf("manifest typed blocks lost source provenance: %#v", manifest.Article.Blocks)
	}
	if len(manifest.Article.AuthorCandidates) != 1 || manifest.Article.AuthorCandidates[0].Name != "mara vale" || len(manifest.Article.AuthorCandidates[0].ImageData) != 0 {
		t.Fatalf("manifest author candidates = %#v", manifest.Article.AuthorCandidates)
	}
	if len(manifest.Resources) != 3 {
		t.Fatalf("manifest resources = %d; want 3", len(manifest.Resources))
	}
	for _, resource := range manifest.Resources[:2] {
		if resource.Status != ResourceStored || resource.Path != imagePath || resource.SHA256 != imageHash || resource.Size != int64(len(result.Document.Resources[0].Data)) {
			t.Fatalf("stored manifest resource = %#v", resource)
		}
	}
	unavailable := manifest.Resources[2]
	if unavailable.Status != ResourceUnavailable || unavailable.SourceURL != "https://cdn.example/missing.png" || unavailable.Path != "" || unavailable.SHA256 != "" {
		t.Fatalf("unavailable manifest resource = %#v", unavailable)
	}
	if len(manifest.Warnings) != 1 || manifest.Warnings[0].Code != "resource-fetch-failed" {
		t.Fatalf("manifest warnings = %#v", manifest.Warnings)
	}
}

func readTestFile(t *testing.T, parts ...string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(parts...))
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", filepath.Join(parts...), err)
	}
	return data
}
