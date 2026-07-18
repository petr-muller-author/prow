# Architecture

Durable, cross-system structure and constraints — patterns and boundaries that hold across many
systems and don't change with any single feature. If you're documenting the behavior of one
system or one flow, it belongs in [`../systems/`](../systems/) or [`../flows/`](../flows/)
instead; this directory is for the "why is the codebase shaped this way" layer above those.

Appropriate here: system-wide patterns, cross-system boundaries, major integration strategies,
persistence/deployment assumptions, architectural tradeoffs.

Not appropriate here: ordinary implementation notes, single-feature documentation, detailed
step-by-step behavior (that belongs in a flow doc).

[`decisions/`](decisions/README.md) holds the point-in-time decision records (ADRs) behind
specific architectural choices; this directory's own docs describe the resulting structure as it
stands today, independent of when or why each piece was decided.

## Current docs

- [`component-decoupling-via-crd.md`](component-decoupling-via-crd.md) — how Prow's components
  communicate.
