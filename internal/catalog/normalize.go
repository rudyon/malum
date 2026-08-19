package catalog

import (
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var foldCase = cases.Fold()

func generatedHandle(displayName string) string {
	var handle strings.Builder
	for _, character := range strings.TrimSpace(displayName) {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			handle.WriteRune(character)
		}
	}
	if handle.Len() == 0 {
		return "author"
	}
	return handle.String()
}

func comparisonKey(value string) string {
	value = foldCase.String(norm.NFKC.String(strings.TrimSpace(value)))
	var key strings.Builder
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsNumber(character) {
			key.WriteRune(character)
		}
	}
	return key.String()
}

func identityKey(kind, value string) string {
	value = strings.TrimSpace(value)
	switch kind {
	case "json-ld-id", "profile-url", "same-as":
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return value
		}
		parsed.Scheme = strings.ToLower(parsed.Scheme)
		parsed.Host = strings.ToLower(parsed.Host)
		if parsed.Path == "/" && parsed.RawQuery == "" && parsed.Fragment == "" {
			parsed.Path = ""
		}
		return parsed.String()
	default:
		return value
	}
}
