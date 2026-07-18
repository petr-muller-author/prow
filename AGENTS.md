# AGENTS.md

This file tells agents (and humans) where to look and how to work in this repository. Detailed
documentation about how Prow actually works lives in [`docs/`](docs/README.md), not here — this
file is a routing layer, not a knowledge base. Don't add system/flow/architecture detail here;
add it to the appropriate file under `docs/` and link to it instead.

This is a fork of [kubernetes-sigs/prow](https://github.com/kubernetes-sigs/prow). The `docs/`
tree described below is **fork-local**: it documents this fork's understanding of Prow for
whoever is working in it, and is not intended to be upstreamed.

## Two documentation trees — don't confuse them

- **`site/content/en/docs/`** — the public, user-facing documentation site (Hugo + Docsy,
  deployed via Netlify). Audience: people *outside* this codebase who use, configure, or operate
  a Prow instance. If you're answering "how do I configure X" or "what does this command do",
  the answer likely already lives there — check first, and don't duplicate it.
- **`docs/`** (this tree) — internal, repo-local documentation about how Prow's internals work:
  systems, cross-system flows, architecture, decisions, and a glossary. Audience: primarily
  people and agents *changing Prow's own code*. See [`docs/PERSONAS.md`](docs/PERSONAS.md) for
  the full breakdown of who reads which tree.

When a topic genuinely belongs in both (e.g. a system doc that also has user-facing config
docs), the internal doc should link to the Hugo page rather than restate it.

## Docs map

- [`docs/README.md`](docs/README.md) — start here; explains the whole `docs/` tree.
- [`docs/PERSONAS.md`](docs/PERSONAS.md) — the audiences Prow documentation serves, and which
  tree (this one or the Hugo site) primarily serves each.
- [`docs/STYLE.md`](docs/STYLE.md) — writing conventions for everything under `docs/`.
- [`docs/glossary.md`](docs/glossary.md) — domain terms; read this if a term looks
  Prow-specific and its meaning isn't obvious.
- [`docs/systems/`](docs/systems/) — one file per coherent system (e.g. `tide.md`).
- [`docs/flows/`](docs/flows/) — behavior that crosses systems (e.g. `prowjob-lifecycle.md`).
- [`docs/architecture/`](docs/architecture/README.md) — durable, cross-system architectural
  decisions and constraints.
- [`docs/architecture/decisions/`](docs/architecture/decisions/README.md) — ADRs: point-in-time
  decisions with context, tradeoffs, and status.
- [`docs/TODO.md`](docs/TODO.md) — backlog of docs that don't exist yet.

## Working with these docs

1. Before changing a system, read the relevant doc(s) under `docs/systems/`, `docs/flows/`, and
   `docs/architecture/` if they exist.
2. Inspect the actual source (docs are a map, not a replacement for reading code).
3. Make your change.
4. If the change affects a system's responsibilities, an interface, a flow, an invariant, or a
   documented decision, update the relevant doc in the same change. Docs and code must agree —
   if you find them disagreeing, fix the doc (or flag the mismatch) rather than leaving it.
5. If you wrote a doc for something not yet covered, remove it from `docs/TODO.md` if it was
   listed there.

Most docs are missing right now — `docs/` currently holds one worked example per category
(see `docs/TODO.md` for what's next), not full coverage. Treat gaps as normal, not as errors.
