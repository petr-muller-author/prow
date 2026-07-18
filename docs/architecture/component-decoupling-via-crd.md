---
audience: [contributor]
---

# Component decoupling via the ProwJob CRD

## Purpose

Explain the durable structural pattern behind almost every Prow component interaction: they
coordinate through the `ProwJob` Kubernetes custom resource, not through direct RPC calls to each
other.

## Scope

The fact that `hook`/`trigger`, `prow-controller-manager`, `crier`, Horologium, Sinker, Deck, and
Tide all read and/or write `ProwJob` resources (`pkg/apis/prowjobs/v1`) as their primary means of
coordination, and what that buys the system.

## Non-scope

The step-by-step sequence of a specific job's lifecycle — that's
[`../flows/prowjob-lifecycle.md`](../flows/prowjob-lifecycle.md). This doc is about the pattern,
not a walkthrough.

## The pattern

A `ProwJob` is created once (by `trigger`, Horologium, or Tide, depending on trigger type) and
then read and updated by whichever component needs to act next, keyed entirely off its `Spec`
and `Status`:

- `prow-controller-manager` watches for ProwJobs needing a pod scheduled.
- `crier` watches for ProwJob status transitions and reports them outward.
- Deck reads ProwJobs (and their pods/artifacts) to render UI.
- Sinker watches for old/completed ProwJobs to garbage-collect.
- Tide creates ProwJobs for missing required contexts and reads their results.

No component in this list calls another directly to say "a job finished" or "please schedule
this" — they all communicate by writing to, or reacting to changes on, the shared `ProwJob`
resource via the Kubernetes API (typically through controller-runtime watches/reconcile loops,
e.g. `cmd/prow-controller-manager/main.go`'s `manager.New` + leader election setup).

## Why this shape

- **Decoupling**: components can be deployed, scaled, and restarted independently — none needs
  to know the network address of another, only the shared Kubernetes API.
- **Durability of intent**: a `ProwJob`'s `Spec` is a durable record of "this job should run
  this way," surviving restarts of any individual component; nothing is lost if
  `prow-controller-manager` or `crier` is down when a ProwJob is created — they catch up once
  they're back.
- **Uniform observability**: `kubectl get prowjobs` (or Deck) gives one consistent view of "what
  is Prow doing right now," regardless of which component last touched a given job.
- **Multi-cluster scheduling**: because coordination is via a Kubernetes resource rather than a
  direct connection, the pod a ProwJob runs in can live in a different (build) cluster than the
  ProwJob resource itself — see the not-yet-written `multi-cluster-build-model.md` in
  [`../TODO.md`](../TODO.md).

## Consequences / tradeoffs

- All coordination is eventually-consistent on watch/reconcile latency, not immediate — no
  component can assume another has already reacted to a just-written `ProwJob` change.
- The `ProwJob` schema (`pkg/apis/prowjobs/v1/types.go`) is a de facto shared API surface between
  every component; changing it has wide blast radius and needs to stay backward compatible
  across independently-deployed components.
- Debugging requires understanding "what does the current `ProwJob.Status` mean" as the shared
  vocabulary, rather than tracing direct calls between two components.

## Source map

- `pkg/apis/prowjobs/v1/types.go` — the shared resource type all of this rests on.
- `cmd/prow-controller-manager/main.go` — a representative controller-runtime consumer
  (watch ProwJobs, reconcile pods).
- `pkg/crier/controller.go` — a representative watch-and-report consumer.

## Related docs

- [`../flows/prowjob-lifecycle.md`](../flows/prowjob-lifecycle.md)
- [`../glossary.md#prowjob`](../glossary.md#prowjob)
