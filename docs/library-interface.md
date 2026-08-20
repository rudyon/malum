# Desktop library

This document translates the current Lunacy library frames into an
implementation reference. Lunacy remains the visual source of truth.

## Scope

The library surface lists documents known to Malum and provides the first URL
import entry point. This slice includes empty, ready, and importing states; an
Add menu; URL entry; navigation into the reader; and navigation back to the
library.

Search, filtering, sorting, tags, bulk operations, archive behavior, and other
source types are outside this slice.

## Desktop geometry

- Reference viewport: 1440 by 1024 pixels.
- Page background: `#090909`.
- Application sidebar: fixed to the left edge, 300 pixels wide and viewport
  high, with 16 pixels of padding, a `#121212` background, and a one-pixel
  right divider. It can collapse to a 108-pixel rail containing the reopen
  control and selected Library icon.
- Brand banner: 268 pixels wide, approximately 61 pixels high, `#252525`
  background, and 12-pixel radius. The Malum wordmark uses Merriweather Bold at
  36 pixels. Sidebar-close and Add controls are aligned to its right.
- The selected Library navigation item sits below the banner. It is the only
  navigation destination currently designed; no speculative entries should be
  added.
- With both sidebars open, library content begins at x=300 and ends at x=1140,
  with 16 pixels of padding. The right 300 pixels contain document information.
- Application chrome and library rows use Atkinson Hyperlegible. Article
  reading typography remains Cabin.

## Empty state

The empty state is centered within the available library content region. It contains the
muted 24-pixel text `No documents yet.` and `Add something you want to read.`,
followed by the Add document control. The Info sidebar is not present when no
document can be previewed. The large amount of remaining empty space is
intentional.

## Library rows

- With both sidebars open, rows are 809 pixels wide and 126 pixels high with
  eight pixels of inset and an 8-pixel gap between thumbnail and information.
- A ready row has a 110-pixel square image with an 8-pixel radius. Its title is
  Atkinson Hyperlegible Bold at 36 pixels. Description and metadata are 24
  pixels; secondary information is `#868686`.
- A ready row is the navigation target for its reader route.
- The first ready document supplies the initial Info preview. Hovering or
  keyboard-focusing another ready row changes the preview; leaving the row does
  not clear it. Clicking continues to open the document.
- An importing row shows the submitted URL as its temporary title and an
  `Importing...` status. Its appearance is muted, but implementation should use
  explicit subdued colors instead of reducing opacity on the whole row. A
  DaisyUI loading indicator may accompany the status.
- A failed import is not a library document. Its importing row disappears and a
  DaisyUI error toast identifies the failed source and offers a retry action.

The implemented library initially loads completed summaries from
`GET /api/documents`. Submitting the URL dialog immediately adds the transient
row, then replaces that exact row with the summary returned by
`POST /api/documents`. Reloading the page reconstructs the ready library from
SQLite rather than client persistence or fixture data.

## Add flow

Both the sidebar add icon and the empty-state button open the same DaisyUI
dropdown. The dropdown currently contains one option, URL. Choosing URL opens a
DaisyUI modal containing a labelled URL input and Cancel/Add actions.

The controls use DaisyUI behavior and accessibility primitives while taking
their colors, typography, and proportions from Malum rather than a stock
DaisyUI theme.

## Navigation

- `/library` displays the library.
- `/documents/:documentId` displays the selected document in the reader.
- The reader sidebar begins with a back control returning to `/library`, above
  the table of contents.

## Document information

The fixed right sidebar is 300 pixels wide, uses the same `#121212` surface and
divider treatment as the application sidebar, and can be closed independently.
Its header contains the reopen/close affordance and centered `Info` label. The
body shows the document title, source domain, author, type, optional publication
date, length, saved time, and progress when Malum has real progress data. The
initial API does not fabricate progress, so imported documents omit that row.
The displayed source domain comes from the source URL's hostname, not the
page-supplied site name. A leading `www.` is removed for display only; the
stored source URL is unchanged.

An author is a nullable first-class entity with a stable internal handle. A
known author without a supplied image receives a deterministic local DiceBear
Line Face avatar, using the handle as its seed, background `#f4f1ea`, and scale
`1.3`. A null author renders the checked-in Unknown Author SVG together with
`Unknown author` and `@unknown`; it does not create or imply a shared author
record. Optional metadata rows with no value are omitted.
