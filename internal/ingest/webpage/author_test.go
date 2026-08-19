package webpage

import (
	"net/url"
	"testing"
)

func TestExtractAuthorCandidatesResolvesJSONLDGraphReference(t *testing.T) {
	base, err := url.Parse("https://journal.example/posts/story")
	if err != nil {
		t.Fatal(err)
	}
	html := []byte(`<!doctype html><html><head>
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@graph": [
    {"@type": "NewsArticle", "author": {"@id": "#author"}},
    {
      "@id": "#author",
      "@type": "Person",
      "name": "alice maz",
      "url": "/authors/alice",
      "sameAs": ["https://social.example/alice"],
      "image": {"thumbnailUrl": "/images/alice.jpg"}
    }
  ]
}
</script></head><body></body></html>`)

	candidates := extractAuthorCandidates(base, html, "ignored fallback")
	if len(candidates) != 1 {
		t.Fatalf("candidates = %#v", candidates)
	}
	candidate := candidates[0]
	if candidate.Name != "alice maz" || candidate.ProfileURL != "https://journal.example/authors/alice" || candidate.ImageURL != "https://journal.example/images/alice.jpg" {
		t.Fatalf("candidate = %#v", candidate)
	}
	if len(candidate.Identities) != 3 {
		t.Fatalf("identities = %#v", candidate.Identities)
	}
}

func TestExtractAuthorCandidatesIgnoresMalformedJSONLD(t *testing.T) {
	base, _ := url.Parse("https://example.test/story")
	html := []byte(`<script type="application/ld+json">{broken</script>`)
	candidates := extractAuthorCandidates(base, html, "rudyon")
	if len(candidates) != 1 || candidates[0].Name != "rudyon" || candidates[0].Evidence != "readability-byline" {
		t.Fatalf("candidates = %#v", candidates)
	}
}
