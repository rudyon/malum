# Desktop article reader

This document translates the Lunacy frame `Reader / Desktop` into an
implementation reference. Lunacy remains the visual source of truth; this file
records the choices needed to reproduce it in the browser.

## Scope

This slice is the uninterrupted reading surface for one normalized article. It
contains the article's table of contents, title, source metadata, images,
captions, headings, and prose. Application navigation, reading controls,
highlights, notes, and document-management actions are deliberately outside
this slice.

## Desktop geometry

- Reference viewport: 1440 by 1024 pixels.
- Page background: `#090909`.
- Table of contents: fixed to the left edge, 356 pixels wide and viewport high,
  with 16 pixels of horizontal padding and 64 pixels of top padding.
- Document frame: begins 420 pixels from the left edge and is 601 pixels wide.
  Its horizontal padding is 16 pixels, leaving a 569-pixel content measure.
- Document content begins 64 pixels from the top. Consecutive top-level content
  blocks are separated by 16 pixels.
- The page, rather than an inner document panel, scrolls. The table of contents
  remains fixed.

The fixed measurements describe the desktop surface. No mobile reader has been
designed yet.

## Typography and color

- Typeface: Cabin, bundled with the frontend rather than fetched at runtime.
- Article title: 36 pixels, bold, centered, white.
- Metadata and captions: 14 pixels. Primary metadata is `#ededed`; secondary
  metadata and captions are `#868686`. Captions are italic and centered.
- Body: 16 pixels with a 24-pixel line height, `#ededed`.
- The Lunacy body is justified. The browser implementation may fall back to
  left alignment if justification creates distracting rivers or spacing.
- Table-of-contents entries: 16 pixels with approximately 27 pixels between
  baselines. The current article is white; other sections are `#868686`.

## Content behavior

- Images occupy the complete 569-pixel content width, use a 4-pixel radius, and
  preserve their aspect ratio.
- Captions sit 8 pixels below their image.
- Table-of-contents entries link to stable section anchors.
- Content is rendered from a typed document value. The rendering boundary must
  not depend on fixture-specific wording, block counts, or image names.
- The checked-in fixture is original material. It mirrors the density and
  structural variety of the Lunacy reference without copying the referenced
  article or its images.

