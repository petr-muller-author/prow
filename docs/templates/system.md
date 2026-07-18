---
audience: [] # one or more of: pr-author, configurator, operator, contributor — see ../PERSONAS.md
---

# <System Name>

## Purpose

What this system is responsible for, in a sentence or two.

## Questions this doc answers

- A bullet list of concrete questions a reader could come here to answer, e.g. "how does X
  decide Y?", "what happens when Z fails?"

## Scope

What's inside this system's responsibility.

## Non-scope

What's deliberately *not* covered here (link to where it is covered instead).

## Key concepts

Terms and concepts specific to this system that a reader needs before the rest of the doc makes
sense. Link to [`../glossary.md`](../glossary.md) for terms that are already defined there.

## How the system works

The core explanation — how it's structured, what drives its behavior.

## Important flows

Flows this system participates in that deserve their own doc — link to
[`../flows/`](../flows/), don't restate them here.

## Data and state

What state this system owns, where it lives, who else can read/write it.

## Interfaces and entry points

How other systems or external actors interact with this one.

## Dependencies

What this system depends on (other systems, external services).

## Downstream effects

What depends on this system, and what breaks if it misbehaves.

## Invariants and assumptions

Things that must remain true for this system to work correctly.

## Error handling

How this system handles failure — its own and its dependencies'.

## Security and privacy notes

Anything relevant to security or privacy, if applicable.

## Observability and debugging

Logs, metrics, dashboards, and common debugging entry points.

## Testing notes

How this system is tested, and anything non-obvious about doing so.

## Common pitfalls

Mistakes people commonly make when working on or around this system.

## Source map

Links to the most important files implementing this system. Don't list everything — prioritize
entry points, state definitions, handlers/services, and tests.

- `path/to/entry_point.go` — ...

## Related docs

Links to related system/flow/architecture docs and Hugo pages.
