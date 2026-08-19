# Working on Malum

Read [`docs/product.md`](docs/product.md) before planning or changing Malum. It
is the source of truth for the product's purpose, quality bar, technical
direction, rejected starting points, and unresolved product questions. Do not
silently resolve its open questions through scaffolding or implementation
convenience.

This file records working agreements and design context that do not belong in
the product document.

## Collaboration

- Work in small, reviewable increments and explain the role and location of new
  pieces. This is the owner's first full-stack application, so comprehension is
  part of the result.
- Do not generate broad scaffolds, placeholder screens, or speculative
  architecture. Propose or implement only the next agreed slice.
- Do not substitute framework defaults or generic component-library patterns
  for design decisions. In particular, DaisyUI is a source of primitives, not
  Malum's visual identity.
- Preserve existing work and decisions. If a proposed change conflicts with the
  product document or the current design, surface the conflict instead of
  smoothing it over.
- Treat unresolved product and interaction choices as questions for the owner,
  not blanks for an agent to fill automatically.

## Continuity protocol

Project documentation is part of each completed slice, not a retrospective
cleanup task. Whenever work establishes or changes a durable product decision,
contract, or implementation boundary:

1. Update the authoritative focused document in `docs/` in the same working
   increment as the implementation.
2. Update `docs/product.md` when the decision resolves or materially changes
   one of its product-level questions.
3. Update this file when the current phase, required reading, or working
   agreements change.

Keep this file concise. It should orient a fresh conversation and route it to
the authoritative documents; it should not duplicate their detailed contracts
or become a session log. Do not record transient next steps, estimates, or
uncommitted experiments here.

At the beginning of a fresh conversation, read `docs/product.md`, the focused
documents named under **Current project state**, and inspect the repository
status and recent history before proposing work.

## Design source

The current UI source of truth is the Lunacy document, accessed through the
Lunacy MCP integration. Inspect the current design before translating or
implementing it; do not rely on an old screenshot or reconstruct it from
memory.

Lunacy `.free` files are intentionally not stored in this repository. The
repository should instead contain textual implementation specifications and the
resulting code. Those specifications describe the design for implementation;
they do not supersede newer deliberate changes in Lunacy.

The implemented desktop surfaces are the library and article reader. Their
textual implementation references are `docs/library-interface.md` and
`docs/reader-interface.md`. Inspect the corresponding current Lunacy frames
before changing their visual or interaction design.

Do not invent additional navigation, toolbars, controls, cards, or screens
while implementing these surfaces.

## Current project state

The current end-to-end milestone is importing a deliberately saved webpage by
URL, storing it locally, listing it in the library, opening it in the reader,
and returning to the library.

- The desktop library and reader are implemented against typed original
  fixture data. See `docs/library-interface.md` and
  `docs/reader-interface.md`.
- `internal/ingest/webpage` retrieves a URL, preserves the original HTML
  response body in memory, extracts a normalized article, and downloads image
  resources. See `docs/webpage-ingestion.md`.
- `internal/storage/document` writes imported webpages as atomic, recoverable
  document bundles. See `docs/document-storage.md`.
- `internal/ingest/webpage`, `internal/storage/author`, and `internal/catalog`
  provide structured author extraction, author-specific avatar storage, and a
  SQLite catalogue. See `docs/catalogue.md`.
- The Go HTTP API and frontend integration have not yet been implemented.

The two manual ingestion targets are Alice Maz's *Playing to Win* and *One with
the Machine*. Neither article nor its images is checked into the repository.

## Implementation sequence

The initial interface-first sequence was:

1. Inspect the Lunacy design.
2. Translate the agreed surface into a detailed textual implementation
   specification.
3. Implement the frontend against typed, realistic fixture data.
4. Use what the interface actually needs to inform the initial frontend
   contracts and document model.
5. Add ingestion and storage behavior, then the SQLite catalogue and Go API,
   when there is a concrete interface slice for them to support.

This is not permission to build a disconnected collection of mock screens. A
fixture-backed surface should be a deliberate implementation of the real UI and
should be capable of accepting real data later without being discarded.

## Reader fixture content

The Lunacy reader design uses Alice Maz's article *Playing to Win* as a visual
reference. Do not commit that article's full text or images, and do not make the
development UI depend on fetching the author's site.

Create an original fictional fixture that exercises the same useful structure:
a realistic title and metadata, long-form paragraphs, multiple heading levels,
a table of contents, images, and captions. Keep its density and content lengths
representative enough to expose real typography and layout problems.

When URL ingestion exists, the real article may be used as a manual end-to-end
import test through Malum's normal ingestion path. It is not bundled sample
data.
