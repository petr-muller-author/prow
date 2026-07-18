---
audience: [contributor, operator]
---

# Tide

## Purpose

Tide (`pkg/tide`, `cmd/tide`) automatically merges pull requests that meet a repo's configured
merge criteria, so merging doesn't require a human to click "merge" once checks pass.

## Questions this doc answers

- What does Tide actually look at when deciding whether a PR is mergeable?
- What's a "pool" and why does Tide work in terms of them instead of one PR at a time?
- How does Tide interact with required status contexts, and how does it create the
  "tide" status context itself?
- Where would I look if Tide isn't merging a PR I expect it to?

## Scope

Deciding which open PRs are eligible to merge (per-repo `tide` config: labels, missing labels,
required/optional contexts, branch protection context requirements), grouping them into pools,
batching, and performing the merge via the code review provider (GitHub or Gerrit).

## Non-scope

- Branch protection itself is configured/enforced elsewhere (Peribolos / GitHub branch
  protection settings) — Tide *reads* required contexts derived from it, but doesn't set it up.
- Presubmit job execution is [`prow-controller-manager`](../TODO.md)'s responsibility; Tide only
  creates ProwJobs for missing/batch contexts and watches their results, it doesn't run them.
- User-facing configuration reference (merge methods, `tide` config fields) is documented on the
  Hugo site: [`components/core/tide/`](../../site/content/en/docs/components/core/tide/_index.md).

## Key concepts

- **Pool** (`Pool`/`PoolForDeck` in `tide.go`) — the set of PRs in a single org/repo/branch being
  considered together for merge at a point in time.
- **Subpool** (`subpool` in `tide.go`) — Tide's internal working unit per pool: the PRs, their
  contexts, and the base SHA, built fresh each sync.
- **Batch** — a speculative combination of multiple PRs tested together (via a batch ProwJob) to
  merge more than one PR per required-context run, when `tide` batch config is enabled.
- **Action** (`Action` type) — what Tide decided to do for a pool this sync: e.g. merge, trigger
  a batch, wait.

## How the system works

`Controller` (`tide.go`) wraps two sub-controllers: a `syncController` that does the actual
pooling/merging work on a timer, and a status controller that keeps the `tide` GitHub status
context up to date independent of merge timing. `syncController.Sync()` is the core loop:
1. List open PRs matching each configured `tide` query.
2. Group them into subpools per org/repo/branch (`initSubpoolData`).
3. Filter each subpool to PRs that are actually eligible (`filterSubpool`/`filterPR`) — correct
   labels present/absent, mergeable, passing required contexts.
4. Pick what to do with each subpool: merge a PR whose required contexts already pass
   (`pickHighestPriorityPR`, honoring `config.TidePriority`), or trigger missing/batch contexts
   as ProwJobs and wait.
5. Perform merges through the review-provider abstraction (`provider`/`CodeReviewCommon`, with
   GitHub and Gerrit implementations in `github.go`/`gerrit.go`).

## Important flows

Tide's own merge sequence isn't yet documented as a `docs/flows/` doc — see
`pr-merge-via-tide.md` in [`../TODO.md`](../TODO.md). It's a good next flow doc: it crosses
Tide, `prow-controller-manager` (for missing-context ProwJobs), and the review provider's API.

## Data and state

Tide holds no persistent state of its own between syncs — each `Sync()` rebuilds pools from the
live PR list and current ProwJobs. It does keep an in-memory `history.History` (recent actions,
exposed via `Controller.History()`) for Deck's Tide history view, and serves the current
`Pool`/`PoolForDeck` snapshot over HTTP (`Controller.ServeHTTP`) for Deck to poll.

## Interfaces and entry points

- `cmd/tide/main.go` — the binary entry point, wires config/GitHub/Kubernetes clients into
  `tide.NewController`.
- `Controller.Sync()` — invoked on a timer; the actual work.
- `Controller.ServeHTTP()` — HTTP endpoint Deck's Tide UI polls for pool state.

## Dependencies

- GitHub (or Gerrit) API, via `ghproxy` in normal deployments, for PR/status/review data and
  performing merges.
- ProwJob API (Kubernetes), to create missing/batch-context ProwJobs and read their results.
- `pkg/config` for the `tide` section of Prow Config (queries, merge method, batching, priorities).

## Downstream effects

Merging a PR triggers postsubmit ProwJobs (via `prow-controller-manager`/`hook` push handling)
and closes the PR — anything depending on "PR is merged" as a signal depends transitively on
Tide's decision.

## Invariants and assumptions

- A PR is never merged unless all of its currently-required contexts (as computed from branch
  protection + explicit `tide` config) are passing at merge time — Tide re-checks this within the
  same sync rather than trusting stale status.
- Subpools are rebuilt from scratch each sync; Tide does not assume its previous sync's view of
  a pool is still accurate.

## Error handling

Per-PR/per-subpool errors (e.g. a merge conflict discovered at merge time) are logged and recorded
in `history.History` rather than aborting the whole sync; other subpools still get processed.

## Security and privacy notes

Tide only acts on PRs already passing required/branch-protection contexts — it doesn't perform
its own trust/author checks (those belong to `trigger`, see
[glossary: Trigger](../glossary.md#trigger)). Merge credentials are the same GitHub App/token
Prow's other write-path components use.

## Observability and debugging

- `Controller.History()` / the Deck Tide history view — the first place to check why a specific
  PR wasn't merged in a given sync.
- `Controller.ServeHTTP()` pool snapshot — current view of what Tide thinks the pool looks like.
- Standard Prow metrics/logging conventions (structured logrus fields keyed by org/repo/PR).

## Testing notes

`tide_test.go`, `status_test.go`, `github_test.go`, `gerrit_test.go` cover pool filtering,
status-context computation, and provider-specific behavior separately — when changing merge
eligibility logic, check `filterPR`/`filterSubpool` tests specifically.

## Common pitfalls

- Assuming a PR's status context is what Tide merges on — Tide computes its own required-context
  set (`requiredContextsMap`) from config + branch protection, which can differ from what's
  literally shown on the PR if config changed recently.
- Forgetting that batches are speculative: a passing batch does not guarantee each individual PR
  in it would pass alone; see `accumulateBatch`.

## Source map

- `pkg/tide/tide.go` — `Controller`, `syncController.Sync`, pool/subpool construction and
  filtering.
- `pkg/tide/status.go` — the `tide` GitHub status context controller.
- `pkg/tide/github.go`, `pkg/tide/gerrit.go`, `pkg/tide/codereview.go` — provider-specific review
  data access (`CodeReviewCommon` abstraction).
- `pkg/tide/blocklist` and `pkg/tide/blockers` — blocking-issue integration.
- `pkg/tide/history` — in-memory action history.
- `cmd/tide/main.go` — binary entry point.

## Related docs

- [`../glossary.md#tide`](../glossary.md#tide)
- Hugo: [`components/core/tide/`](../../site/content/en/docs/components/core/tide/_index.md) —
  user-facing config reference for `configurator`s.
