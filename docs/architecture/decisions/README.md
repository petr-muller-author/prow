# Architecture Decision Records

ADRs document proposed, active, and historical technical decisions: context, tradeoffs,
consequences, and rejected alternatives. Use one for a decision that shapes the system beyond a
single implementation detail (e.g. "why is job config allowed to live in-repo, not just
centrally"). Routine implementation details, small refactors, bug fixes, and temporary
workarounds don't need one.

## File naming

One file per decision, named as a descriptive slug (no numeric prefix — this fork doesn't use
sequential ADR numbers): `docs/architecture/decisions/<slug>.md`.

## Required frontmatter

```yaml
---
status: Accepted      # Proposed | Accepted | Superseded | Deprecated | Rejected
date: YYYY-MM-DD
# superseded_by: relative/path/to/replacement.md   # required when status: Superseded
audience: [contributor]
---
```

- **Proposed** — under discussion, not yet current guidance.
- **Accepted** — the current decision.
- **Superseded** — replaced by a newer ADR; `superseded_by` must point to it.
- **Deprecated** — discouraged going forward but historically relevant.
- **Rejected** — considered but not adopted; kept so the option doesn't get re-litigated from
  scratch later without knowing why it was turned down.

## Modification rules

Never rewrite an Accepted ADR to describe a different decision. If the decision changes, write a
new ADR, mark the old one Superseded, and link `superseded_by`. Small corrections (typos, broken
links) that don't change the recorded decision are fine to edit in place.

## Backfilled decisions

Some of this fork's ADRs document decisions made long before this `docs/` tree existed. Those
are marked explicitly in their `## Context` section as retroactive documentation, with the ADR's
`date` set to a best-effort estimate (e.g. from `git log`) rather than the actual original
decision date, which this fork's ADR process cannot know for certain.

## Current ADRs

- [`in-repo-prow-configuration.md`](in-repo-prow-configuration.md) — Accepted, backfilled.
