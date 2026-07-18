# Personas

Prow is unusual among the projects the [source methodology](https://gist.github.com/lukewilson2002/cb48062397d8b51954034d94b8c19d6d)
was designed for: it doesn't have one audience of "engineers building the app." It has several
audiences with mostly disjoint concerns, and Prow already has two documentation trees
(`docs/` and the public Hugo site) that split along those lines. This file defines the personas
so that both trees — and the `audience:` frontmatter used throughout `docs/` (see
[`STYLE.md`](STYLE.md)) — refer to the same fixed vocabulary.

Every doc under `systems/`, `flows/`, `architecture/`, and `architecture/decisions/` should
declare which of these personas it's written for via `audience:` frontmatter, using the slugs
below.

## `pr-author`

Someone whose repository is serviced by a Prow instance, interacting with it through PRs: `/lgtm`,
`/retest`, `/hold`, status checks, seeing job results. They don't configure or run Prow — they
just experience its behavior on their PRs.

**Primarily served by:** the Hugo site, e.g.
[`life-of-a-prow-job.md`](../site/content/en/docs/life-of-a-prow-job.md). Internal docs in this
tree should link to the relevant Hugo page for this persona rather than duplicate it.

## `configurator`

Someone configuring Prow's behavior for their repo or org: `config.yaml` job definitions,
`plugins.yaml` plugin enablement, OWNERS-based merge rules, branch protection, in-repo config.
They don't run Prow's control plane, but they make decisions that depend on how Prow's systems
actually behave (e.g. how Tide computes a merge pool).

**Primarily served by:** the Hugo site (e.g. `config.md`, the `components/plugins/` pages,
`inrepoconfig.md`). Internal `docs/systems/` and `docs/architecture/decisions/` pages are a
secondary reference when the "why" behind a config option matters.

## `operator`

Someone deploying and running a Prow instance: cluster setup, scaling, secrets, build clusters,
monitoring, incident response.

**Partly served by:** the Hugo site (`getting-started-deploy.md`, `scaling.md`,
`prow-secrets.md`). Internal runbook/incident-response material has no home yet in this fork —
see the note in [`TODO.md`](TODO.md); when written, it belongs in `docs/` under a new
`docs/runbooks/` category (see the "System Extensibility" note in the source methodology), not
in the Hugo site.

## `contributor`

Someone changing Prow's own code in this repository: fixing bugs, adding systems, refactoring.
This is the primary audience `docs/` (systems, flows, architecture, decisions) exists for.

**Primarily served by:** this tree, in full.

## Multiple personas per doc

A doc can — and often should — declare more than one audience, e.g. `docs/systems/tide.md` is
`[contributor, operator]` because both need to understand merge-pool behavior, just for
different reasons (changing it vs. debugging it in production). List every persona for whom the
doc is a primary reference; don't list a persona just because the topic touches them tangentially.
