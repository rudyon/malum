# Development methodology

Malum uses evidence-led incremental development. Work begins with a real
observed problem, a deliberate design decision, or a concrete technical risk;
it does not begin with a generic feature checklist or a speculative roadmap.

The aim is to keep development connected to the experience of using a personal
reading environment while retaining enough written context for decisions to
survive across sessions and contributors.

## Development loop

Each increment follows this sequence:

```text
use or observe
      |
      v
capture a concrete issue
      |
      v
choose one issue to address
      |
      v
decide behaviour and design
      |
      v
document durable contracts
      |
      v
implement and test one coherent slice
      |
      v
use it, commit it, and close the issue
```

An increment is complete when its agreed behaviour works, its relevant tests
pass, its durable decisions are documented, and it has been checked in the
running application where that is meaningful. A newly discovered concern is a
new issue, not an automatic reason to expand the active increment.

## Issue tracker

GitHub Issues is Malum's durable work queue. An issue may describe a bug, an
agreed capability, an unresolved design question, or maintenance work. It is
created when the observation is concrete enough to preserve useful context;
issues are not required for every passing idea.

The repository uses GitHub's default labels where they fit, plus:

- `ingestion` for retrieval, extraction, and import concerns;
- `design` for unresolved visual, interaction, or product choices; and
- `maintenance` for technical upkeep that protects the product.

The issue template is a prompt, not a form to complete mechanically. Its
fields may be omitted when they do not help. A useful issue records the actual
behaviour or desired outcome, why it matters, a relevant URL or document when
applicable, and any decision that must not be made implicitly.

GitHub Projects, milestones, estimates, points, sprints, and standing
ceremonies are not part of the current process. They should be introduced only
if a demonstrated coordination problem makes one of them useful.

## Choosing work

Keep only one implementation or design increment active at a time. Choose it
from the open issues based on what most impedes using Malum, most threatens the
trustworthiness of the library, or is needed by the next intentionally chosen
capability.

After a meaningful change, use Malum with material that matters to its owner
when possible. Real imports and reading sessions are the primary way the
project discovers its next requirements.

## Before implementation

Before code changes, establish the relevant behaviour. If an issue affects a
visible surface, inspect its current Lunacy design and update or create the
textual interface specification before implementation. If it changes storage,
API, ingestion, or another durable boundary, update the focused contract in
`docs/` in the same increment.

Unresolved choices remain explicit questions for the owner. They are never
silently resolved merely because a framework, component, or implementation
path has a convenient default.

## Branches and pull requests

`main` should remain in a working, committed state. For small, understood
increments, direct work on `main` is appropriate. Use a short-lived branch
when an increment is risky, spans multiple sessions, or benefits from an
isolated experiment.

Pull requests are optional rather than required. They become useful when their
review view, discussion, or CI history materially helps the work; they are not
ceremony required for a single-owner repository.
