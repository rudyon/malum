// Package webpage retrieves and normalizes readable webpage articles.
package webpage

import "time"

type BlockKind string

const (
	BlockParagraph    BlockKind = "paragraph"
	BlockHeading      BlockKind = "heading"
	BlockImage        BlockKind = "image"
	BlockList         BlockKind = "list"
	BlockDefinitions  BlockKind = "definitions"
	BlockQuote        BlockKind = "quote"
	BlockPreformatted BlockKind = "preformatted"
	BlockDivider      BlockKind = "divider"
	BlockHTML         BlockKind = "html"
)

type Snapshot struct {
	RequestedURL string `json:"requestedUrl"`
	FinalURL     string `json:"finalUrl"`
	ContentType  string `json:"contentType"`
	OriginalHTML []byte `json:"-"`
}

type Document struct {
	SourceURL          string        `json:"sourceUrl"`
	Title              string        `json:"title"`
	Byline             string        `json:"byline,omitempty"`
	SiteName           string        `json:"siteName"`
	Language           string        `json:"language,omitempty"`
	Excerpt            string        `json:"excerpt,omitempty"`
	PublishedAt        *time.Time    `json:"publishedAt,omitempty"`
	ModifiedAt         *time.Time    `json:"modifiedAt,omitempty"`
	WordCount          int           `json:"wordCount"`
	ReadingTimeMinutes int           `json:"readingTimeMinutes"`
	LeadImageURL       string        `json:"leadImageUrl,omitempty"`
	ContentHTML        string        `json:"contentHtml"`
	Blocks             []Block       `json:"blocks"`
	Outline            []OutlineItem `json:"outline"`
	Resources          []Resource    `json:"resources"`
	Warnings           []Warning     `json:"warnings,omitempty"`
}

type Block struct {
	Kind        BlockKind    `json:"kind"`
	ID          string       `json:"id,omitempty"`
	Level       int          `json:"level,omitempty"`
	Text        string       `json:"text,omitempty"`
	HTML        string       `json:"html,omitempty"`
	Image       *Image       `json:"image,omitempty"`
	List        *List        `json:"list,omitempty"`
	Definitions []Definition `json:"definitions,omitempty"`
}

type Image struct {
	URL     string `json:"url"`
	Alt     string `json:"alt,omitempty"`
	Caption string `json:"caption,omitempty"`
}

type List struct {
	Ordered bool       `json:"ordered"`
	Items   []ListItem `json:"items"`
}

type ListItem struct {
	Text string `json:"text"`
	HTML string `json:"html,omitempty"`
}

type Definition struct {
	Term        string `json:"term"`
	Description string `json:"description"`
}

type OutlineItem struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
	Title string `json:"title"`
}

type Resource struct {
	URL         string `json:"url"`
	Role        string `json:"role"`
	ContentType string `json:"contentType,omitempty"`
	Data        []byte `json:"-"`
}

type Warning struct {
	Code    string `json:"code"`
	URL     string `json:"url,omitempty"`
	Message string `json:"message"`
}

type Result struct {
	Snapshot Snapshot `json:"snapshot"`
	Document Document `json:"document"`
}
