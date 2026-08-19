# HTTP API

This document defines the initial HTTP boundary between Malum's React client
and Go application. It supports the first saved-webpage milestone; it is not a
general public API or a promise that later document formats use URL import.

## Server boundary

The Go application is one process. It opens the SQLite catalogue, composes the
webpage importer and filesystem stores, and exposes JSON and stored assets over
HTTP. During development Vite proxies `/api` to this process. A production
build will eventually be served by the same Go process, so the API is
same-origin and does not enable permissive CORS.

The server binds to `127.0.0.1:8080` by default. Listening on another address is
an explicit operator choice; authentication and a supported remote-access
configuration remain outside this milestone. The data root is supplied
explicitly with `--data-dir` and is never silently placed in the repository.

## Import behavior

`POST /api/documents` accepts exactly one URL:

```json
{
  "url": "https://www.alicemaz.com/writing/minecraft.html"
}
```

The request remains open while Malum retrieves, extracts, stores, and
catalogues the article. The frontend owns the transient importing row while
the request is pending. A successful response has status `201 Created` and
contains the completed document summary plus any non-fatal ingestion warnings.
Malum does not create a job table, persist failed rows, or require polling for
this milestone.

The complete document bundle is written before catalogue publication. If
publication fails, the existing recoverable-orphan behavior still applies.
Failure to durably store an available author avatar is non-fatal to an
otherwise completed document and is reported as a warning.

## URL safety

Server-side retrieval accepts only HTTP and HTTPS. By default, every connection
must resolve exclusively to globally routable addresses. Loopback, private,
link-local, multicast, unspecified, and other non-global destinations are
rejected, including redirect destinations and article or author image URLs.
The transport connects to the addresses it validated rather than performing a
second independent DNS lookup.

This default prevents an imported URL from being used to reach services on the
Malum host or its private network. A deliberate opt-in for intranet imports may
be designed later; it is not inferred from self-hosting.

## Routes

```text
POST /api/documents
GET  /api/documents
GET  /api/documents/{document-id}
GET  /api/documents/{document-id}/resources/{filename}
GET  /api/authors/{author-id}/avatar
```

`GET /api/documents` returns completed documents in catalogue order. It does
not fabricate reading progress. `GET /api/documents/{document-id}` adds the
stored typed blocks and heading outline required by the reader.

Document summaries expose source and presentation metadata, an optional author,
and API URLs for an available thumbnail or avatar. Handles are returned without
an `@`; presentation adds it. Absolute server filesystem paths are never
returned.

Stored document resources are addressed by their manifest filename. A resource
request succeeds only when that filename belongs to a stored resource in that
document's manifest. Author avatars likewise require a catalogue author and a
catalogued path within that author's avatar directory. These routes do not
interpret arbitrary filesystem paths supplied by clients.

## Reader data and trust

The reader response transports the typed block projection preserved in the
document manifest, including lists, definitions, quotations, preformatted
text, dividers, and retained HTML fragments. Image URLs with a stored resource
are projected to the corresponding local API resource URL.

HTML fields remain untrusted extracted data. Returning them in JSON is not
permission for the frontend to use `dangerouslySetInnerHTML`. A separate
sanitization and rendering decision is still required before direct HTML
rendering; typed text fields can be rendered without making that decision.

## Errors

API errors use a stable outer shape:

```json
{
  "error": {
    "code": "document_import_failed",
    "message": "Malum could not import a readable article from this URL."
  }
}
```

Client errors use a `4xx` status. Unexpected storage, catalogue, and server
errors use `500` without exposing filesystem paths or internal error chains.
The server logs the underlying error for diagnosis.
