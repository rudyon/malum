package webpage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

const (
	defaultMaxPageBytes          int64 = 16 << 20
	defaultMaxResourceBytes      int64 = 20 << 20
	defaultMaxTotalResourceBytes int64 = 100 << 20
	defaultUserAgent                   = "Malum webpage importer"
)

var (
	ErrPageTooLarge     = errors.New("webpage response exceeds size limit")
	ErrResourceTooLarge = errors.New("webpage resource exceeds size limit")
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Importer struct {
	Client                HTTPClient
	MaxPageBytes          int64
	MaxResourceBytes      int64
	MaxTotalResourceBytes int64
	UserAgent             string
}

func NewImporter(client HTTPClient) Importer {
	return Importer{
		Client:                client,
		MaxPageBytes:          defaultMaxPageBytes,
		MaxResourceBytes:      defaultMaxResourceBytes,
		MaxTotalResourceBytes: defaultMaxTotalResourceBytes,
		UserAgent:             defaultUserAgent,
	}
}

func (i Importer) Import(ctx context.Context, rawURL string) (Result, error) {
	if i.Client == nil {
		return Result{}, errors.New("webpage importer requires an HTTP client")
	}
	if _, err := parseHTTPURL(rawURL); err != nil {
		return Result{}, err
	}

	snapshot, err := i.fetchPage(ctx, rawURL)
	if err != nil {
		return Result{}, err
	}
	document, err := Extract(snapshot.FinalURL, snapshot.OriginalHTML)
	if err != nil {
		return Result{}, err
	}
	document.SourceURL = snapshot.FinalURL
	i.fetchResources(ctx, &document)

	return Result{Snapshot: snapshot, Document: document}, nil
}

func (i Importer) fetchPage(ctx context.Context, rawURL string) (Snapshot, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create webpage request: %w", err)
	}
	request.Header.Set("Accept", "text/html, application/xhtml+xml;q=0.9")
	request.Header.Set("User-Agent", i.userAgent())

	response, err := i.Client.Do(request)
	if err != nil {
		return Snapshot{}, fmt.Errorf("fetch webpage: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Snapshot{}, fmt.Errorf("fetch webpage: unexpected HTTP status %s", response.Status)
	}

	contentType := response.Header.Get("Content-Type")
	mediaType, _, parseErr := mime.ParseMediaType(contentType)
	if parseErr != nil {
		return Snapshot{}, fmt.Errorf("parse webpage content type %q: %w", contentType, parseErr)
	}
	if mediaType != "text/html" && mediaType != "application/xhtml+xml" {
		return Snapshot{}, fmt.Errorf("fetch webpage: unsupported content type %q", mediaType)
	}

	originalHTML, err := readLimited(response.Body, i.maxPageBytes(), ErrPageTooLarge)
	if err != nil {
		return Snapshot{}, fmt.Errorf("fetch webpage: %w", err)
	}
	finalURL := rawURL
	if response.Request != nil && response.Request.URL != nil {
		finalURL = response.Request.URL.String()
	}

	return Snapshot{
		RequestedURL: rawURL,
		FinalURL:     finalURL,
		ContentType:  mediaType,
		OriginalHTML: originalHTML,
	}, nil
}

func (i Importer) fetchResources(ctx context.Context, document *Document) {
	var totalBytes int64
	for index := range document.Resources {
		resource := &document.Resources[index]
		remaining := i.maxTotalResourceBytes() - totalBytes
		if remaining <= 0 {
			document.Warnings = append(document.Warnings, Warning{
				Code:    "resource-total-limit",
				URL:     resource.URL,
				Message: "resource was not fetched because the import reached its total image byte limit",
			})
			continue
		}

		limit := min(i.maxResourceBytes(), remaining)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, resource.URL, nil)
		if err != nil {
			i.warnResource(document, resource.URL, err)
			continue
		}
		request.Header.Set("Accept", "image/*")
		request.Header.Set("User-Agent", i.userAgent())

		response, err := i.Client.Do(request)
		if err != nil {
			i.warnResource(document, resource.URL, err)
			continue
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			response.Body.Close()
			i.warnResource(document, resource.URL, fmt.Errorf("unexpected HTTP status %s", response.Status))
			continue
		}

		contentType := response.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || !strings.HasPrefix(mediaType, "image/") {
			response.Body.Close()
			i.warnResource(document, resource.URL, fmt.Errorf("unsupported image content type %q", contentType))
			continue
		}
		data, err := readLimited(response.Body, limit, ErrResourceTooLarge)
		response.Body.Close()
		if err != nil {
			i.warnResource(document, resource.URL, err)
			continue
		}
		resource.ContentType = mediaType
		resource.Data = data
		totalBytes += int64(len(data))
	}
}

func (i Importer) warnResource(document *Document, resourceURL string, err error) {
	document.Warnings = append(document.Warnings, Warning{
		Code:    "resource-fetch-failed",
		URL:     resourceURL,
		Message: err.Error(),
	})
}

func readLimited(reader io.Reader, limit int64, limitError error) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, limitError
	}
	return data, nil
}

func (i Importer) maxPageBytes() int64 {
	if i.MaxPageBytes > 0 {
		return i.MaxPageBytes
	}
	return defaultMaxPageBytes
}

func (i Importer) maxResourceBytes() int64 {
	if i.MaxResourceBytes > 0 {
		return i.MaxResourceBytes
	}
	return defaultMaxResourceBytes
}

func (i Importer) maxTotalResourceBytes() int64 {
	if i.MaxTotalResourceBytes > 0 {
		return i.MaxTotalResourceBytes
	}
	return defaultMaxTotalResourceBytes
}

func (i Importer) userAgent() string {
	if strings.TrimSpace(i.UserAgent) != "" {
		return i.UserAgent
	}
	return defaultUserAgent
}
