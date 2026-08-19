# SQLite catalogue

This document defines Malum's initial SQLite catalogue and author-resolution
contract. The catalogue makes completed filesystem bundles discoverable and
stores current, user-correctable metadata. It does not store article HTML,
typed article content, or image bodies.

## Database boundary

The catalogue uses Go's `database/sql` with explicit SQL. It does not use an
ORM. SQLite-specific code remains inside `internal/catalog`, whose operations
return ordinary Go values rather than exposing SQL rows to the application.

The database is `<data-root>/malum.db`. Numbered SQL migrations are embedded in
the program and applied transactionally. Document and author identities are
Malum-generated UUIDs rather than SQLite row IDs.

SQLite duplicates the metadata needed to list, identify, and query documents:
title, description, source, reading kind, acquisition method, original format,
site, raw byline, language, source timestamps, length, thumbnail path, saved
time, and nullable author association. This is a queryable projection of the
preserved bundle, not a replacement original.

HTML, typed blocks, outlines, resource manifests, image bodies, and ingestion
warnings remain in the document bundle.

## Initial tables

`documents` contains one row for each bundle published in the library. Its
primary key is the bundle's document ID. Source URLs are deliberately not
unique because duplicate and re-import behavior remains unresolved.

`authors` contains a stable UUID, visible handle, case-folded handle key,
display name, case-folded comparison key, optional avatar source URL, optional
local avatar path, and creation time.

`author_identities` associates an author with strong structured evidence such
as a JSON-LD identifier, author profile URL, or `sameAs` URL. Each normalized
identity value may belong to only one author.

The current singular `documents.author_id` reflects the designed interface.
Ingestion preserves all extracted author candidates even though the catalogue
uses only the first candidate as the primary author for this milestone.

## Display names and handles

An extracted display name is trimmed but never title-cased, lowercased, or
otherwise rewritten for presentation. For example, Alice Maz's chosen display
name remains `alice maz`.

A default handle removes spaces and punctuation from a separate copy of the
display name while preserving its letter casing:

```text
alice maz  -> @alicemaz
Alice Maz  -> @AliceMaz
rudyon     -> @rudyon
```

Handles are unique under an invisible Unicode-aware, case-folded comparison,
so `@alicemaz` and `@AliceMaz` cannot belong to different authors. When a new
author genuinely collides with an occupied handle, Malum appends a numeric
suffix. The visible display name and handle retain their chosen casing.

## Best-effort author resolution

Web metadata is incomplete and fallible. Malum therefore makes reversible
best-effort decisions rather than claiming perfect identity proof.

Resolution follows this order:

1. Reuse an author when a structured identifier, profile URL, or `sameAs` URL
   exactly matches a known normalized identity.
2. Otherwise, compare the invisible normalized name key. If exactly one author
   matches, reuse it.
3. Otherwise, create a new author with an available generated handle.
4. If no credible author name exists, leave `author_id` null and use the agreed
   `Unknown author` / `@unknown` presentation.

Avatar images and names are not used as biometric or probabilistic identity
evidence. Raw bylines and structured candidates remain preserved so later
manual correction can reassign, merge, or split authors without erasing what
ingestion originally observed.

## Author avatars

An author image URL is provenance, not durable presentation. When structured
metadata supplies an image, the importer attempts to download it. Failure is
non-fatal and the author continues to use the deterministic DiceBear fallback.

Downloaded avatars live outside document bundles:

```text
<data-root>/authors/<author-id>/avatars/<sha256>.<extension>
```

The catalogue records both the remote source URL and the relative local path.
An author avatar must not depend on the continued existence of one particular
document bundle.

## Publication and queries

Document storage completes the filesystem bundle first. Catalogue publication
then resolves or creates the primary author and inserts the document row in one
SQLite transaction. An existing document ID is rejected rather than replaced.

The first catalogue operations are:

- publish a completed document bundle;
- list document summaries;
- retrieve one document summary;
- record a locally stored author avatar.

Importing and failed rows remain transient frontend state. Reading progress is
also deferred: a percentage alone cannot provide reliable resume behavior
across HTML, EPUB, and PDF, so the initial catalogue does not enshrine that
incomplete model.

If catalogue publication fails, the complete bundle remains as a recoverable
but unindexed directory. A later reconciliation operation can detect such
orphans by comparing bundle IDs with catalogue IDs.

The two live targets can exercise the complete importer, bundle, catalogue,
resolution, and avatar-storage path without committing their content:

```powershell
$env:MALUM_MILESTONE_TEST='1'
go test ./internal/catalog -run TestManualMilestonePublication -v
```

## Manual correction

The database contains Malum's current presentation metadata. Later interfaces
may edit document metadata, author display names and handles, avatars, and
document-author assignments; they may also merge duplicate authors or split an
incorrect match. Those editing interactions are not part of the current slice,
but the schema must not treat extracted guesses as immutable truth.
