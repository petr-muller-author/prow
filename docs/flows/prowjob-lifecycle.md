---
audience: [contributor]
---

# ProwJob lifecycle: webhook to reported status

## Purpose

This is the core cross-system flow in Prow: a GitHub event turns into one or more running jobs,
and their results make it back to GitHub. Almost every Prow component participates in some part
of it, and understanding it is close to a prerequisite for changing any of them.

This doc covers the same ground as the Hugo site's
[`life-of-a-prow-job.md`](../../site/content/en/docs/life-of-a-prow-job.md), but from the
`contributor` angle — which component owns which step and which files implement it — rather than
the `pr-author` angle of what a contributor to an external repo sees happen to their PR. Read the
Hugo page first if you want the user-visible framing; read this one when you need to trace or
change the underlying mechanism.

## Trigger

A GitHub webhook event: a PR opened/updated, a comment posted (e.g. `/test foo`, `/retest`), or a
push to a branch. (Periodic jobs instead originate from Horologium on a schedule — see
[`../TODO.md`](../TODO.md) for the not-yet-written `horologium.md` system doc — and skip the
first two steps below.)

## Participants

- `ghproxy` (`cmd/ghproxy`) — caches/fronts the GitHub API calls the other participants make.
- `hook` (`cmd/hook`) — receives the webhook, dispatches to plugins.
- `trigger` plugin (`pkg/plugins/trigger`) — decides which presubmits apply and creates ProwJobs.
- `prow-controller-manager` (`cmd/prow-controller-manager`) — watches ProwJobs, schedules pods.
- `crier` (`pkg/crier`) — watches ProwJob status, reports it to GitHub (and other sinks).
- Tide (see [`../systems/tide.md`](../systems/tide.md)) — a downstream consumer of the resulting
  status contexts, not part of this flow itself.

## Step-by-step flow

1. GitHub sends a webhook for the event; it reaches `hook` (typically via `ghproxy` for
   Prow's own outbound calls, though inbound webhooks go directly to `hook`).
2. `hook` dispatches the event to registered plugins. For presubmit-relevant events, this
   includes `trigger` (`handlePullRequest`/`handleGenericCommentEvent`/`handlePush` in
   `pkg/plugins/trigger/trigger.go`).
3. `trigger` determines which presubmits are relevant (`FilterPresubmits`,
   `pkg/plugins/trigger/generic-comment.go`) and whether the PR author is trusted
   (`TrustedPullRequest`/`TrustedUser`, `pkg/plugins/trigger/pull-request.go` /
   `trigger.go`) — untrusted authors need an `/ok-to-test` from an org member before jobs run.
4. For each applicable presubmit, `trigger` creates a `ProwJob` custom resource
   (`pkg/apis/prowjobs/v1`) with `Status.State` starting at `SchedulingState` or
   `TriggeredState`.
5. `prow-controller-manager` watches ProwJobs cluster-wide (`cmd/prow-controller-manager/main.go`,
   controller-runtime `manager.New` with leader election) and, for each new ProwJob, schedules a
   pod to run it, moving `Status.State` through `PendingState`.
6. The pod runs; pod utilities (`clonerefs`, `entrypoint`, `initupload`, `sidecar` — see
   [glossary: Pod Utilities](../glossary.md#pod-utilities)) handle cloning the repo, capturing
   logs, and uploading artifacts independent of the job's own container.
7. On completion, the ProwJob's `Status.State` becomes `SuccessState`/`FailureState`/`ErrorState`
   (`pkg/apis/prowjobs/v1/types.go`).
8. `crier` (`pkg/crier/controller.go`) watches ProwJob status changes and, through its reporters
   (`pkg/crier/reporters/github` and others — Slack, Pub/Sub, GCS, ResultStore, Gerrit), reports
   the result back — for GitHub, as a status context on the PR's head SHA.
9. That status context is what Tide (and branch protection) later reads when deciding
   mergeability — see [`../systems/tide.md`](../systems/tide.md).

## Data and state changes

The `ProwJob` resource is the single piece of state that threads through the whole flow — every
step reads or writes its `Status`, rather than components messaging each other directly. See
[`../architecture/component-decoupling-via-crd.md`](../architecture/component-decoupling-via-crd.md).

## Success behavior

ProwJob reaches `SuccessState`/`FailureState`; a corresponding GitHub status context is set;
Tide can subsequently consider it if it's a required context.

## Failure behavior

- If `trigger` decides a PR is untrusted, no ProwJob is created and (depending on config) a
  comment is posted asking for `/ok-to-test`.
- If pod scheduling fails, `prow-controller-manager` retries; persistent failure surfaces as
  `ErrorState` on the ProwJob rather than leaving it stuck silently.
- If `crier` fails to report (e.g. transient GitHub API error via `ghproxy`), it retries on its
  own reconcile loop rather than dropping the report.

## External dependencies

GitHub API (through `ghproxy`), the Kubernetes API of the ProwJob's target build cluster.

## Invariants and assumptions

- Exactly one `ProwJob` resource represents one job run — components never share ad hoc state
  outside the ProwJob to coordinate.
- Untrusted-author gating happens once, in `trigger`, before a ProwJob is even created — later
  stages don't re-check trust.

## Security and privacy notes

Trust-gating in step 3 is the main access-control point in this flow — it's what prevents
arbitrary GitHub users from running CI (and thus arbitrary code) in Prow's build clusters. See
`TrustedPullRequest`/`TrustedUser` in `pkg/plugins/trigger`.

## Observability and debugging

- ProwJob status (`kubectl get prowjobs`, or Deck's job list) is the primary place to see where a
  given job is stuck.
- `hook` and `trigger` logs show whether an event was received and whether it was filtered out as
  untrusted or not matching any presubmit.
- `crier` logs show reporting attempts/failures per ProwJob.

## Testing notes

Each participant has its own unit tests (`trigger`: `pull-request_test.go`,
`generic-comment_test.go`; `tide`: see [`../systems/tide.md`](../systems/tide.md)); there is no
single end-to-end test covering the whole flow in this repo as far as this doc has traced —
worth confirming before relying on that assumption.

## Source map

- `pkg/plugins/trigger/` — steps 2–4.
- `pkg/apis/prowjobs/v1/types.go` — the ProwJob type threading the whole flow.
- `cmd/prow-controller-manager/main.go` — step 5.
- `pkg/crier/controller.go`, `pkg/crier/reporters/` — steps 8–9.

## Related docs

- [`../systems/tide.md`](../systems/tide.md)
- [`../architecture/component-decoupling-via-crd.md`](../architecture/component-decoupling-via-crd.md)
- Hugo: [`life-of-a-prow-job.md`](../../site/content/en/docs/life-of-a-prow-job.md) — user-facing
  framing of this same flow.
