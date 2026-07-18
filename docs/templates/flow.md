---
audience: [] # one or more of: pr-author, configurator, operator, contributor — see ../PERSONAS.md
---

# <Flow Name>

## Purpose

What this flow accomplishes and why it deserves its own doc (crosses systems, has several
steps/state transitions, changes/breaks often, involves external services, or has
security/data-integrity/user-visible implications).

## Trigger

What starts this flow.

## Participants

The systems/components involved, each linking to its [`../systems/`](../systems/) doc if one
exists.

## Step-by-step flow

1. ...
2. ...

## Data and state changes

What state changes as this flow progresses, and where it's persisted.

## Success behavior

What the end state looks like when this flow completes normally.

## Failure behavior

What happens at each likely failure point — retries, partial state, user-visible errors.

## External dependencies

Any external services this flow relies on.

## Invariants and assumptions

Things that must hold true for this flow to behave correctly.

## Security and privacy notes

Anything relevant to security or privacy, if applicable.

## Observability and debugging

How to trace this flow in logs/metrics/traces when debugging it.

## Testing notes

How this flow is tested, and anything non-obvious about doing so.

## Source map

Links to the most important files implementing each step.

## Related docs

Links to related system/architecture docs and Hugo pages (especially if a user-facing
description of this flow already exists there).
