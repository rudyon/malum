package catalog

import "time"

type Author struct {
	ID              string
	Handle          string
	DisplayName     string
	AvatarSourceURL string
	AvatarPath      string
}

type Document struct {
	ID                 string
	ReadingKind        string
	AcquisitionMethod  string
	OriginalFormat     string
	SourceURL          string
	Title              string
	Description        string
	SiteName           string
	RawByline          string
	Author             *Author
	Language           string
	PublishedAt        *time.Time
	SourceModifiedAt   *time.Time
	WordCount          int
	ReadingTimeMinutes int
	ThumbnailPath      string
	SavedAt            time.Time
}
