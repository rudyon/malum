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

## Design source

The current UI source of truth is the Lunacy document, accessed through the
Lunacy MCP integration. Inspect the current design before translating or
implementing it; do not rely on an old screenshot or reconstruct it from
memory.

Lunacy `.free` files are intentionally not stored in this repository. The
repository should instead contain textual implementation specifications and the
resulting code. Those specifications describe the design for implementation;
they do not supersede newer deliberate changes in Lunacy.

The first designed surface is the desktop article reader, in the frame named
`Reader / Desktop` (last known Lunacy layer ID
`YaPG5uBNS06iGF3nEX9GfA`). It establishes the reading surface, not the complete
reader application.

Current design facts include:

- A 1440 by 1024 desktop frame with a near-black background.
- A 356-pixel table-of-contents column and a centered reading column whose text
  measure is approximately 569 pixels.
- Cabin typography: approximately 36 pixels for the article title, 16/24 for
  body text, and 14 pixels for metadata and captions.
- Body copy is currently justified. Watch for visibly poor word spacing during
  browser implementation rather than preserving justification mechanically.
- The frame contains only what was considered essential to reading at this
  stage. Missing controls and annotation states are out of scope for this
  surface; their absence does not imply that the eventual product should hide
  or omit them.
- Content must flow and scroll naturally. The fixed design frame is a viewport,
  not a fixed-height document container.

Do not invent additional navigation, toolbars, controls, cards, or screens while
implementing this surface.

## Implementation sequence

For the present phase, work from the interface inward:

1. Inspect the Lunacy design.
2. Translate the agreed surface into a detailed textual implementation
   specification.
3. Implement the frontend against typed, realistic fixture data.
4. Use what the interface actually needs to inform the initial frontend
   contracts and document model.
5. Add the Go API, SQLite schema, and ingestion/storage behavior when there is a
   concrete interface slice for them to support.

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

