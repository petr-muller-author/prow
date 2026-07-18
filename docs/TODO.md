# Documentation backlog

`docs/` currently has one worked example per category. This is a starting backlog of what to
write next, not an exhaustive list — add to it as gaps are found. Pick one item, copy the
matching file from [`templates/`](templates/), write it, then remove the item from this list.

Each item is tagged with its likely `audience:` (see [`PERSONAS.md`](PERSONAS.md)).

## Systems (`docs/systems/`)

- [ ] `hook.md` — webhook ingestion and plugin dispatch (`cmd/hook`, `pkg/plugins`). `[contributor]`
- [ ] `prow-controller-manager.md` — ProwJob scheduling onto pods (`cmd/prow-controller-manager`). `[contributor, operator]`
- [ ] `crier.md` — status reporting (`pkg/crier`). `[contributor]`
- [ ] `horologium.md` — periodic job scheduling (`cmd/horologium`). `[contributor]`
- [ ] `sinker.md` — ProwJob/pod garbage collection (`cmd/sinker`). `[contributor, operator]`
- [ ] `deck.md` — web frontend (`cmd/deck`). `[contributor]`
- [ ] `spyglass.md` — artifact rendering (`pkg/spyglass`). `[contributor]`
- [ ] `ghproxy.md` — GitHub API caching proxy (`cmd/ghproxy`, `pkg/ghcache`). `[contributor, operator]`
- [ ] `config-loading.md` — how `pkg/config` loads/merges/validates central + in-repo config. `[contributor, configurator]`
- [ ] `plugin-framework.md` — how `pkg/plugins` registers and dispatches plugins. `[contributor]`
- [ ] `gerrit.md` — the Gerrit adapter, as an alternative to the GitHub path. `[contributor, operator]`

## Flows (`docs/flows/`)

- [ ] `pr-merge-via-tide.md` — from a PR meeting merge criteria to Tide actually merging it. `[contributor]`
- [ ] `github-webhook-ingestion.md` — a GitHub event's path through ghproxy/hook before reaching plugins. `[contributor]`
- [ ] `build-cluster-registration.md` — how a build cluster gets registered and used for scheduling. `[contributor, operator]`
- [ ] `branch-protection-sync.md` — how Peribolos reconciles configured branch protection. `[contributor, operator]`
- [ ] `secret-rotation.md` — operational flow for rotating Prow's secrets. `[operator]`

## Architecture (`docs/architecture/`)

- [ ] `multi-cluster-build-model.md` — how Prow schedules jobs across multiple build clusters. `[contributor, operator]`
- [ ] `plugin-extensibility-model.md` — the durable design behind plugin registration/config, beyond any one plugin. `[contributor]`
- [ ] `artifact-storage-model.md` — the GCS-based artifact/storage conventions pod utilities rely on. `[contributor, operator]`
- [ ] `ghproxy-caching-architecture.md` — the caching strategy and its tradeoffs, as a durable design (as opposed to `systems/ghproxy.md`, which covers the component itself). `[contributor, operator]`

## Architecture decisions (`docs/architecture/decisions/`)

- [ ] `ghproxy-existence.md` — why a caching proxy exists in front of the GitHub API at all. `[contributor, operator]`
- [ ] `peribolos-vs-manual-branch-protection.md` — why branch protection is reconciled by a controller instead of managed by hand. `[contributor, operator]`
- [ ] `deck-public-private-split.md` — why Deck has a public/private deployment split (see `private-deck.md` on the Hugo site for the user-facing description). `[contributor, operator]`
- [ ] `moonraker-introduction.md` — why Moonraker was introduced and what it took over. `[contributor]`

## Runbooks (new category, not yet created)

`docs/architecture/decisions/` and friends don't have a home for operational procedures
(incident response, secret rotation runbooks, etc.) — per the source methodology's "System
Extensibility" note, a `docs/runbooks/` category should be added once there's a first real
runbook to put in it. Don't create the empty category speculatively.

## Glossary

`glossary.md` is seeded with the most load-bearing terms; extend it opportunistically whenever a
doc uses a Prow-specific term that isn't there yet, rather than as a standalone task.
