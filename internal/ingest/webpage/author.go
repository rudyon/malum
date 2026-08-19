package webpage

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	nethtml "golang.org/x/net/html"
)

const (
	authorEvidenceJSONLD      = "json-ld"
	authorEvidenceReadability = "readability-byline"
)

func extractAuthorCandidates(baseURL *url.URL, originalHTML []byte, byline string) []AuthorCandidate {
	values := jsonLDValues(originalHTML)
	index := make(map[string]map[string]any)
	for _, value := range values {
		indexJSONLDIDs(value, baseURL, index)
	}

	var candidates []AuthorCandidate
	for _, value := range values {
		visitJSONLDObjects(value, func(object map[string]any) {
			authorValue, ok := object["author"]
			if !ok {
				return
			}
			for _, candidate := range authorCandidatesFromValue(authorValue, baseURL, index) {
				candidates = appendUniqueAuthor(candidates, candidate)
			}
		})
	}
	if len(candidates) > 0 {
		return candidates
	}

	byline = strings.TrimSpace(byline)
	if byline == "" {
		return nil
	}
	return []AuthorCandidate{{Name: byline, Evidence: authorEvidenceReadability}}
}

func jsonLDValues(originalHTML []byte) []any {
	document, err := nethtml.Parse(bytes.NewReader(originalHTML))
	if err != nil {
		return nil
	}
	var values []any
	var visit func(*nethtml.Node)
	visit = func(node *nethtml.Node) {
		if node.Type == nethtml.ElementNode && strings.EqualFold(node.Data, "script") &&
			strings.EqualFold(strings.TrimSpace(attribute(node, "type")), "application/ld+json") {
			var value any
			decoder := json.NewDecoder(strings.NewReader(scriptText(node)))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err == nil {
				values = append(values, value)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return values
}

func scriptText(node *nethtml.Node) string {
	var text strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == nethtml.TextNode {
			text.WriteString(child.Data)
		}
	}
	return text.String()
}

func indexJSONLDIDs(value any, baseURL *url.URL, index map[string]map[string]any) {
	visitJSONLDObjects(value, func(object map[string]any) {
		if id := resolvedStringURL(baseURL, stringValue(object["@id"])); id != "" {
			index[id] = object
		}
	})
}

func visitJSONLDObjects(value any, visit func(map[string]any)) {
	switch value := value.(type) {
	case map[string]any:
		visit(value)
		for _, child := range value {
			visitJSONLDObjects(child, visit)
		}
	case []any:
		for _, child := range value {
			visitJSONLDObjects(child, visit)
		}
	}
}

func authorCandidatesFromValue(value any, baseURL *url.URL, index map[string]map[string]any) []AuthorCandidate {
	var candidates []AuthorCandidate
	switch value := value.(type) {
	case string:
		if name := strings.TrimSpace(value); name != "" {
			candidates = append(candidates, AuthorCandidate{Name: name, Evidence: authorEvidenceJSONLD})
		}
	case []any:
		for _, item := range value {
			candidates = append(candidates, authorCandidatesFromValue(item, baseURL, index)...)
		}
	case map[string]any:
		object := value
		if id := resolvedStringURL(baseURL, stringValue(value["@id"])); id != "" {
			if referenced, ok := index[id]; ok {
				object = referenced
			}
		}
		if candidate, ok := authorCandidateFromObject(object, baseURL); ok {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func authorCandidateFromObject(object map[string]any, baseURL *url.URL) (AuthorCandidate, bool) {
	name := strings.TrimSpace(stringValue(object["name"]))
	if name == "" {
		given := strings.TrimSpace(stringValue(object["givenName"]))
		family := strings.TrimSpace(stringValue(object["familyName"]))
		name = strings.TrimSpace(given + " " + family)
	}
	if name == "" {
		return AuthorCandidate{}, false
	}

	candidate := AuthorCandidate{
		Name:       name,
		ImageURL:   imageURL(object["image"], baseURL),
		ProfileURL: resolvedStringURL(baseURL, stringValue(object["url"])),
		Evidence:   authorEvidenceJSONLD,
	}
	if id := resolvedStringURL(baseURL, stringValue(object["@id"])); id != "" {
		candidate.Identities = appendIdentity(candidate.Identities, "json-ld-id", id)
	}
	if candidate.ProfileURL != "" {
		candidate.Identities = appendIdentity(candidate.Identities, "profile-url", candidate.ProfileURL)
	}
	for _, identifier := range identifierValues(object["identifier"], baseURL) {
		candidate.Identities = appendIdentity(candidate.Identities, "identifier", identifier)
	}
	for _, sameAs := range stringValues(object["sameAs"]) {
		if identity := resolvedStringURL(baseURL, sameAs); identity != "" {
			candidate.Identities = appendIdentity(candidate.Identities, "same-as", identity)
		}
	}
	return candidate, true
}

func imageURL(value any, baseURL *url.URL) string {
	switch value := value.(type) {
	case string:
		return resolvedStringURL(baseURL, value)
	case []any:
		for _, item := range value {
			if result := imageURL(item, baseURL); result != "" {
				return result
			}
		}
	case map[string]any:
		for _, key := range []string{"contentUrl", "url", "thumbnailUrl"} {
			if result := resolvedStringURL(baseURL, stringValue(value[key])); result != "" {
				return result
			}
		}
	}
	return ""
}

func identifierValues(value any, baseURL *url.URL) []string {
	var result []string
	switch value := value.(type) {
	case string:
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	case []any:
		for _, item := range value {
			result = append(result, identifierValues(item, baseURL)...)
		}
	case map[string]any:
		propertyID := strings.TrimSpace(stringValue(value["propertyID"]))
		identifier := strings.TrimSpace(stringValue(value["value"]))
		if identifier == "" {
			identifier = resolvedStringURL(baseURL, stringValue(value["url"]))
		}
		if identifier != "" {
			if propertyID != "" {
				identifier = propertyID + ":" + identifier
			}
			result = append(result, identifier)
		}
	}
	return result
}

func appendIdentity(identities []AuthorIdentity, kind, value string) []AuthorIdentity {
	for _, identity := range identities {
		if identity.Kind == kind && identity.Value == value {
			return identities
		}
	}
	return append(identities, AuthorIdentity{Kind: kind, Value: value})
}

func appendUniqueAuthor(candidates []AuthorCandidate, candidate AuthorCandidate) []AuthorCandidate {
	for _, existing := range candidates {
		if existing.Name == candidate.Name && existing.ProfileURL == candidate.ProfileURL {
			return candidates
		}
	}
	return append(candidates, candidate)
}

func resolvedStringURL(baseURL *url.URL, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	resolved := baseURL.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	return resolved.String()
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}

func stringValues(value any) []string {
	switch value := value.(type) {
	case string:
		return []string{value}
	case []any:
		var result []string
		for _, item := range value {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}
