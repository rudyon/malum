# Malum

Malum is a self-hosted reading environment. It is intended to provide the kind
of coherent reading system offered by Readwise Reader without making that system
dependent on a SaaS company.

This document records the decisions that led to the repository. It is a source
of continuity, not a complete specification or roadmap.

## Purpose

Malum should give one person a single place where material can arrive, be kept,
and be read. That material will eventually include deliberately saved webpages,
books, PDFs, feeds or newsletters, and potentially other sources.

The central requirement is ownership:

> If every company involved disappeared tomorrow, the reading system and
> everything placed in it should continue to work.

That means:

- The server runs on hardware controlled by the user.
- Essential functionality does not require an external cloud service.
- Original documents remain available in ordinary, recoverable formats where
  possible.
- Metadata, reading state, highlights, and notes belong to the user and should
  be exportable.
- The installation can be backed up and kept running without permission from a
  vendor.

Self-hosting is therefore not merely a deployment option. It is part of the
product's identity.

## Product shape

Malum is not primarily a book server, RSS reader, bookmark manager, or Kavita
clone. It is a personal reading environment in which different kinds of reading
material belong to one coherent system.

The broad flow is:

```text
material arrives
      |
      v
inbox / feed / library
      |
      v
reading
      |
      v
progress / highlights / notes / archive
```

The distinction used by Readwise Reader between deliberately saved material and
automatically arriving feed material is important. The exact Malum vocabulary
and behavior have not yet been designed.

Different sources should not create unrelated miniature applications. A saved
webpage and an RSS item are both documents acquired in different ways. EPUB and
PDF require format-specific handling, but should still participate in the same
library, reading-state, search, and annotation system where that is meaningful.

The quality of the reading loop is more important than supporting every format:

```text
open -> continue -> read -> annotate -> close -> resume reliably
```

## Product quality

Malum must not become a technically functional but crude collection of forms,
cards, and dashboards. Visual and interaction quality cannot be postponed to a
final "polish" phase if early interface decisions will become the product's
foundation.

The UI must be designed deliberately. Framework defaults, generated dashboard
layouts, arbitrary component-library styling, and broad AI-generated scaffolds
are not an acceptable substitute for art direction. DaisyUI may provide useful
primitives, but its defaults do not define Malum's appearance.

An early version may have little functionality. It should not use that limited
scope as an excuse to feel disposable, generic, or careless.

## Technical direction

The chosen stack is:

```text
Frontend:    React + TypeScript
UI:          Tailwind CSS + DaisyUI
Build tool:  Vite
Backend:     Go
Database:    SQLite
Protocol:    HTTP + JSON
Deployment:  one self-hosted server
```

React was chosen partly because ecosystem size matters for browser-based EPUB,
PDF, and document tooling. The frontend should remain restrained: additional
state, data-fetching, or application frameworks need a concrete reason to exist.

The core application belongs in Go rather than in a JavaScript server framework.
The intended deployment is one process that can eventually serve the compiled
frontend, handle the API and ingestion work, and use SQLite and local document
storage. A future native client should be able to use the same server without
requiring the core reading system to be rewritten.

The Go module is:

```text
github.com/rudyon/malum
```

## Working principles

- Keep the complete product in view without attempting to build every subsystem
  at once.
- Introduce abstractions because they remove understood mechanical work, not
  because a conventional stack includes them.
- Prefer one coherent program over a collection of services unless separation
  solves a demonstrated problem.
- Build in small, inspectable changes. Do not generate the whole application and
  attempt to repair it afterward.
- Explain what each new piece does and where it runs. Malum is also the owner's
  first full-stack application; understanding the system is part of the work.
- Preserve originals and escape routes rather than trapping the library in
  Malum-specific representations.

## Directions already rejected

The following have been proposed and found inadequate as starting points:

- Reducing Malum to a URL-saving toy or generic EPUB library.
- Attempting Readwise Reader feature parity from the outset.
- Treating Kavita as the product or visual model.
- Beginning with a fake multi-screen client containing hardcoded documents.
- Treating a wireframe as though it were a working piece of the product.
- Allowing a simple server-rendered tutorial UI to determine the final product.
- Starting with a broad framework scaffold or a pile of conventional services.
- Presenting the entire future architecture as work that must be understood or
  implemented immediately.

A minimal HTTP executable was discussed only as a way to make the selected
client/server architecture concrete. It is infrastructure, not a meaningful
Malum feature or a visual prototype.

## Decisions established by the first milestone

The first end-to-end capability is a deliberately saved URL that becomes a
durable article in the library and opens in the reader. Its interface is
documented in `docs/library-interface.md` and `docs/reader-interface.md`.

The first acquisition method is URL retrieval, the first preserved original
format is HTML, and the first normalized reading kind is article. These are
independent concepts rather than one `webpage` type. Retrieval and extraction
are documented in `docs/webpage-ingestion.md`.

Original HTML, normalized article HTML, typed article structure, and downloaded
resources are stored together in a recoverable filesystem bundle. SQLite will
later catalogue those bundles and own mutable Malum state. The boundary is
documented in `docs/document-storage.md`.

## Still unresolved

These questions remain deliberate product decisions rather than blanks to fill
through implementation convenience:

- What are the precise roles and transitions among inbox, feed, library, and
  archive?
- Which additional concepts and states belong in the document model as reading
  progress, annotation, and other formats are introduced?
- How should duplicate URLs, canonical URLs, and re-imports behave?
- How should extracted bylines be resolved to internal authors?
- What sanitization and rendering policy should apply to normalized HTML?
- Which acquisition method or original format should follow saved HTML pages?
