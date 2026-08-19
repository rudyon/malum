// Package document stores durable, recoverable document bundles.
package document

import (
	"time"

	"github.com/rudyon/malum/internal/ingest/webpage"
)

const (
	SchemaVersion = 1

	ReadingKindArticle = "article"
	AcquisitionURL     = "url"
	OriginalFormatHTML = "html"
)

type Manifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	DocumentID    string            `json:"documentId"`
	ReadingKind   string            `json:"readingKind"`
	StoredAt      time.Time         `json:"storedAt"`
	Acquisition   Acquisition       `json:"acquisition"`
	Original      Original          `json:"original"`
	Article       Article           `json:"article"`
	Resources     []Resource        `json:"resources"`
	Warnings      []webpage.Warning `json:"warnings,omitempty"`
}

type Acquisition struct {
	Method       string `json:"method"`
	RequestedURL string `json:"requestedUrl"`
	FinalURL     string `json:"finalUrl"`
}

type Original struct {
	Format      string `json:"format"`
	ContentType string `json:"contentType"`
	Path        string `json:"path"`
	SHA256      string `json:"sha256"`
	Size        int64  `json:"size"`
}

type Article struct {
	Path               string                `json:"path"`
	SHA256             string                `json:"sha256"`
	Size               int64                 `json:"size"`
	Title              string                `json:"title"`
	Byline             string                `json:"byline,omitempty"`
	SiteName           string                `json:"siteName"`
	Language           string                `json:"language,omitempty"`
	Excerpt            string                `json:"excerpt,omitempty"`
	PublishedAt        *time.Time            `json:"publishedAt,omitempty"`
	ModifiedAt         *time.Time            `json:"modifiedAt,omitempty"`
	WordCount          int                   `json:"wordCount"`
	ReadingTimeMinutes int                   `json:"readingTimeMinutes"`
	LeadImageURL       string                `json:"leadImageUrl,omitempty"`
	Blocks             []webpage.Block       `json:"blocks"`
	Outline            []webpage.OutlineItem `json:"outline"`
}

type ResourceStatus string

const (
	ResourceStored      ResourceStatus = "stored"
	ResourceUnavailable ResourceStatus = "unavailable"
)

type Resource struct {
	SourceURL   string         `json:"sourceUrl"`
	Role        string         `json:"role"`
	Status      ResourceStatus `json:"status"`
	Path        string         `json:"path,omitempty"`
	ContentType string         `json:"contentType,omitempty"`
	SHA256      string         `json:"sha256,omitempty"`
	Size        int64          `json:"size,omitempty"`
}

type Saved struct {
	ID       string
	Path     string
	Manifest Manifest
}
