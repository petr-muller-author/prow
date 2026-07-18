# docs/

This is Prow's **repo-local, internal documentation** — how Prow's own systems, flows, and
architecture work, for people and agents changing Prow's code in this fork. It follows the
methodology in [this gist](https://gist.github.com/lukewilson2002/cb48062397d8b51954034d94b8c19d6d):
`AGENTS.md` at the repo root tells agents where to look; `docs/` explains how the application
works.

This tree is **fork-local**, maintained here for this fork's own use — it is not intended to be
upstreamed to `kubernetes-sigs/prow`.

## This is not the public docs site

Prow already has public, user-facing documentation: the Hugo + Docsy site at
[`site/content/en/docs/`](../site/content/en/docs/), deployed to the public Prow docs site. That
tree answers "how do I configure/use/deploy Prow" for people outside this codebase. This tree
(`docs/`) answers "how does Prow work internally" for people changing Prow itself. See
[`AGENTS.md`](../AGENTS.md#two-documentation-trees--dont-confuse-them) for the short version, and
[`PERSONAS.md`](PERSONAS.md) for the full audience breakdown.

Rule of thumb: if a doc here covers a topic that also has a Hugo page, link to the Hugo page
instead of restating its content.

## Structure

| Path | Content | Primary audience |
|---|---|---|
| [`systems/`](systems/) | One doc per coherent system (module/service/feature area) | contributor |
| [`flows/`](flows/) | Behavior that crosses systems | contributor |
| [`architecture/`](architecture/README.md) | Durable, cross-system structure and constraints | contributor |
| [`architecture/decisions/`](architecture/decisions/README.md) | ADRs — point-in-time decisions | contributor |
| [`glossary.md`](glossary.md) | Domain terms | all |
| [`PERSONAS.md`](PERSONAS.md) | Who reads Prow docs, and where they should look | all |
| [`STYLE.md`](STYLE.md) | Writing conventions for this tree | doc authors |
| [`templates/`](templates/) | Starting point for new system/flow/ADR docs | doc authors |
| [`TODO.md`](TODO.md) | Backlog of docs that don't exist yet | doc authors |

## Current state

This tree currently holds structure, guidelines, and exactly **one worked example per
category** (`systems/tide.md`, `flows/prowjob-lifecycle.md`,
`architecture/component-decoupling-via-crd.md`,
`architecture/decisions/in-repo-prow-configuration.md`), not full coverage of Prow. See
[`TODO.md`](TODO.md) for what's planned next — pick an item, copy the matching template from
`templates/`, and write it.

## Maintenance rule

Any change in this fork that alters a system's responsibilities, an interface, a cross-system
flow, an invariant, or a recorded decision should update the relevant doc in the same change.
Documentation is only useful if it's trustworthy — prefer no doc over a stale one, and prefer
marking uncertainty explicitly over guessing.
