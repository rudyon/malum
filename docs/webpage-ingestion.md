# Webpage ingestion

This document records the first webpage-import contract. It describes the
boundary implemented by `internal/ingest/webpage`; it is not a complete storage
or API design.

## Scope

The importer accepts one webpage URL, retrieves the original response, extracts
the primary readable article, describes its remote image resources, and returns
a normalized document. It does not write SQLite rows, choose filesystem paths,
create Malum authors, expose HTTP routes, or update the frontend.

The initial manual targets are Alice Maz's *Playing to Win* and *One with the
Machine*. Those articles and their images are not committed to the repository.
Automated tests use small original HTML fixtures.

## Result boundary

An import result contains three related representations:

1. **Original response** — the exact HTML response-body bytes returned by the
   source, together with requested URL, final URL after redirects, and content
   type. The document storage layer retains these bytes in an ordinary
   recoverable file.
2. **Cleaned article HTML** — the main content selected and cleaned by
   Readability. This preserves inline meaning such as links and emphasis, plus
   structures the current reader does not render yet. It is normalized input,
   not browser-trusted HTML; the frontend must sanitize it before direct HTML
   rendering.
3. **Typed block projection** — ordered headings, paragraphs, images, lists,
   definitions, quotations, preformatted text, dividers, and retained HTML
   fragments. This supplies stable structure for a reader and a table of
   contents without making extraction depend on fixture wording.

The normalized metadata currently includes title, raw extracted byline, site
name, language, excerpt, publication and modification times when present, word
count, estimated reading time, lead image URL, a flat heading outline, and
warnings.

## Authors

The importer preserves the raw Readability byline and separately extracts
structured author candidates. JSON-LD `author` metadata is the preferred
signal. A candidate may contain an exact display name, image URL, profile URL,
and structured identities such as `@id`, `identifier`, or `sameAs`, together
with the evidence source.

When no structured candidate exists, a non-empty Readability byline becomes a
fallback candidate. The importer does not generate an internal handle or decide
whether candidates from different documents are the same person; that belongs
to the catalogue contract in `docs/catalogue.md`.

Display-name casing is preserved. A missing author is a successful import and
renders through the agreed `Unknown author` / `@unknown` presentation.

When a structured candidate provides an image, the importer attempts to fetch
it under the same bounded image policy used for article resources. Failure adds
a warning but does not fail the document.

## Images

Image URLs are resolved against the final page URL and deduplicated into a
resource manifest. The importer attempts to retrieve them and returns their
bytes and content type to its caller. An individual image failure does not
discard an otherwise readable document; it produces a warning and leaves the
remote source URL in the normalized content.

The document storage layer chooses durable filenames, writes the returned
bytes, rewrites normalized references to relative local paths, and records
checksums. Its contract is documented in `docs/document-storage.md`.

## Fetching and limits

The importer accepts an HTTP client from its caller. Redirect policy, proxy
configuration, DNS resolution rules, and whether private-network URLs are
allowed therefore belong to the future API/server boundary rather than being
silently fixed here.

The package itself rejects non-HTTP(S) URLs, non-success responses, non-HTML
documents, oversized HTML responses, and documents from which no readable
content can be obtained. Page, per-image, and total-image byte limits prevent
unbounded reads.

## Extraction engine

Malum uses `codeberg.org/readeck/go-readability/v2`, the maintained successor to
the archived `github.com/go-shiori/go-readability` package. It follows Mozilla
Readability 0.6 and returns both metadata and a cleaned content DOM.

Readability is a strong default, not an assertion that every webpage can be
normalized automatically. A later ingestion system may add site-specific
adapters or a manual repair workflow when concrete failures justify them.

## Deferred decisions

- URL safety policy for the HTTP API.
- Canonical-URL deduplication and re-import behavior.
- HTML sanitization and rendering policy in the frontend.
- Styling and interaction for lists, definitions, quotations, code, tables,
  and other structures not yet designed in the reader.
