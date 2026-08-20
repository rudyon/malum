# Desktop article reader

This document translates the Lunacy frame `Reader / Desktop` into an
implementation reference. Lunacy remains the visual source of truth; this file
records the choices needed to reproduce it in the browser.

## Scope

This slice is the reading surface for one normalized article. It contains the
article's hierarchical table of contents, document information, title, source
metadata, images, captions, headings, and prose. Highlights, notes, and
document-management actions remain outside this slice.

## Desktop geometry

- Reference viewport: 1440 by 1024 pixels.
- Page background: `#090909`.
- Table-of-contents sidebar: fixed to the left edge, 300 pixels wide and
  viewport high, with 16 pixels of padding, a `#121212` background, and a
  one-pixel right divider.
- Document frame: begins 420 pixels from the left edge and is 601 pixels wide.
  Its horizontal padding is 16 pixels, leaving a 569-pixel content measure.
- Document information: fixed to the right edge, 300 pixels wide and viewport
  high, mirroring the left sidebar.
- Document content begins 64 pixels from the top. Consecutive top-level content
  blocks are separated by 16 pixels.
- The page, rather than an inner document panel, scrolls. The table of contents
  and Info sidebar remain fixed.
- Closing either sidebar leaves its compact reopen control at the corresponding
  viewport edge. Closing the sidebars does not move the reading column.

The fixed measurements describe the desktop surface. No mobile reader has been
designed yet.

## Typography and color

- Article typeface: Cabin, bundled with the frontend rather than fetched at
  runtime. Sidebar chrome uses Atkinson Hyperlegible.
- Article title: 36 pixels, bold, centered, white.
- Metadata and captions: 14 pixels. Primary metadata is `#ededed`; secondary
  metadata and captions are `#868686`. Captions are italic and centered.
- Body: 16 pixels with a 24-pixel line height, `#ededed`.
- The Lunacy body is justified. The browser implementation may fall back to
  left alignment if justification creates distracting rivers or spacing.
- Table-of-contents entries: 16 pixels. The current article is white and bold;
  other sections are `#868686`. Nested headings are indented beneath their
  parent section.

## Content behavior

- Images occupy the complete 569-pixel content width, use a 4-pixel radius, and
  preserve their aspect ratio.
- Captions sit 8 pixels below their image.
- Table-of-contents entries link to stable section anchors.
- An article with no extracted headings has an empty outline and remains a
  valid readable document; its table of contents contains no section entries.
- The header chevrons expand or collapse all nested table-of-contents headings.
  Parent sections with children may also be expanded or collapsed individually.
  The unavailable all-expand or all-collapse action is disabled.
- Content is rendered from a typed document value. The rendering boundary must
  not depend on fixture-specific wording, block counts, or image names.
- The API-backed reader renders paragraphs, headings, images, lists,
  definitions, quotations, preformatted text, dividers, and tables. Retained
  table HTML is parsed into inert cell text and React elements; extracted HTML
  is never injected into the page.
- API image references that were successfully stored resolve through Malum's
  checked local resource routes. Unavailable resources may retain their remote
  source URL as defined by the ingestion contract.
- The checked-in fixture is original material. It mirrors the density and
  structural variety of the Lunacy reference without copying the referenced
  article or its images.
- The Info sidebar uses the same author and optional-metadata rules documented
  for the library. It is the same component in both surfaces, rather than a
  second document-details implementation.
