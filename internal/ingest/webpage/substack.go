package webpage

import (
	"bytes"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

var substackProfilePostPath = regexp.MustCompile(`^/@[^/]+/p-(\d+)/?$`)

// substackArticleURL recognizes Substack's profile/share URL, whose HTTP
// response is an application shell, and reads the public article URL from the
// shell's own bootstrap data.
func substackArticleURL(pageURL string, originalHTML []byte) (string, bool) {
	parsedPageURL, err := url.Parse(pageURL)
	if err != nil || !strings.EqualFold(strings.TrimPrefix(parsedPageURL.Hostname(), "www."), "substack.com") {
		return "", false
	}
	pathMatch := substackProfilePostPath.FindStringSubmatch(parsedPageURL.EscapedPath())
	if pathMatch == nil {
		return "", false
	}

	const marker = "window._preloads"
	markerIndex := bytes.Index(originalHTML, []byte(marker))
	if markerIndex < 0 {
		return "", false
	}
	bootstrap := originalHTML[markerIndex+len(marker):]
	parseIndex := bytes.Index(bootstrap, []byte("JSON.parse("))
	if parseIndex < 0 {
		return "", false
	}
	bootstrap = bootstrap[parseIndex+len("JSON.parse("):]

	var encoded string
	if err := json.NewDecoder(bytes.NewReader(bootstrap)).Decode(&encoded); err != nil {
		return "", false
	}
	var preloads struct {
		FeedData struct {
			InitialPost struct {
				Post struct {
					ID           json.Number `json:"id"`
					CanonicalURL string      `json:"canonical_url"`
				} `json:"post"`
			} `json:"initialPost"`
		} `json:"feedData"`
	}
	decoder := json.NewDecoder(strings.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&preloads); err != nil || preloads.FeedData.InitialPost.Post.ID.String() != pathMatch[1] {
		return "", false
	}

	articleURL, err := parseHTTPURL(preloads.FeedData.InitialPost.Post.CanonicalURL)
	if err != nil || articleURL.Hostname() == "" || !strings.HasPrefix(articleURL.EscapedPath(), "/p/") {
		return "", false
	}
	return articleURL.String(), true
}
