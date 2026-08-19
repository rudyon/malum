package webpage

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestManualWebpageImport(t *testing.T) {
	targetURL := os.Getenv("MALUM_WEBPAGE_TEST_URL")
	if targetURL == "" {
		t.Skip("set MALUM_WEBPAGE_TEST_URL to run a live import")
	}

	client := &http.Client{Timeout: 45 * time.Second}
	result, err := NewImporter(client).Import(context.Background(), targetURL)
	if err != nil {
		t.Fatal(err)
	}
	if result.Document.Title == "" || result.Document.WordCount == 0 || len(result.Document.Blocks) == 0 {
		t.Fatalf("incomplete document: %#v", result.Document)
	}

	fetchedResources := 0
	blockCounts := make(map[BlockKind]int)
	for _, block := range result.Document.Blocks {
		blockCounts[block.Kind]++
	}
	for _, resource := range result.Document.Resources {
		if len(resource.Data) > 0 {
			fetchedResources++
		}
	}
	for _, warning := range result.Document.Warnings {
		t.Logf("warning code=%s url=%s message=%s", warning.Code, warning.URL, warning.Message)
	}
	for _, heading := range result.Document.Outline {
		t.Logf("heading level=%d id=%s title=%q", heading.Level, heading.ID, heading.Title)
	}
	t.Logf("lead image=%s", result.Document.LeadImageURL)
	for _, block := range result.Document.Blocks {
		if block.Image != nil {
			t.Logf("image url=%s captionBytes=%d", block.Image.URL, len(block.Image.Caption))
		}
	}
	t.Logf("block counts=%v", blockCounts)
	t.Logf(
		"title=%q byline=%q site=%q words=%d minutes=%d blocks=%d headings=%d resources=%d/%d warnings=%d originalBytes=%d",
		result.Document.Title,
		result.Document.Byline,
		result.Document.SiteName,
		result.Document.WordCount,
		result.Document.ReadingTimeMinutes,
		len(result.Document.Blocks),
		len(result.Document.Outline),
		fetchedResources,
		len(result.Document.Resources),
		len(result.Document.Warnings),
		len(result.Snapshot.OriginalHTML),
	)
}
