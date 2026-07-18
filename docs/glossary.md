---
audience: [contributor, configurator, operator]
---

# Glossary

Domain terms used across Prow's internal documentation and code. When a finished doc refers to
one of these concepts, use the Title Case form for clarity.

### ProwJob

The Kubernetes custom resource (`prowjobs.prow.k8s.io`, `pkg/apis/prowjobs/v1`) representing a
single run of a job (test, build, etc.). Prow's components communicate primarily by
creating/reading/updating ProwJobs rather than calling each other directly — see
[`architecture/component-decoupling-via-crd.md`](architecture/component-decoupling-via-crd.md).

### Plugin

A unit of GitHub-event-driven behavior registered with `hook`, configured per-repo/org in
`plugins.yaml` (e.g. `lgtm`, `approve`, `trigger`). Lives under `pkg/plugins/`.

### Trigger

The plugin (`pkg/plugins/trigger`) responsible for deciding whether a GitHub event (a PR being
opened, a `/test` comment, etc.) should result in one or more ProwJobs being created. See
[`flows/prowjob-lifecycle.md`](flows/prowjob-lifecycle.md).

### Prow Config

The central `config.yaml`, defining jobs, cluster settings, and other instance-wide behavior.
Loaded and typed by `pkg/config`.

### Plugins Config

The central `plugins.yaml`, defining which plugins are enabled per repo/org and their settings.

### Tide

The merge-automation system (`pkg/tide`, `cmd/tide`) that merges PRs meeting configured criteria
(labels, required contexts, branch protection). See [`systems/tide.md`](systems/tide.md).

### Deck

Prow's web frontend (`cmd/deck`), serving job status, PR history, and other UI-facing views.

### Hook

The webhook-handling component (`cmd/hook`) that receives GitHub events and dispatches them to
plugins.

### Prow Controller Manager

The component (`cmd/prow-controller-manager`) that watches for ProwJobs and schedules their
execution as Kubernetes pods. Successor to the deprecated `plank`.

### Crier

The component (`pkg/crier`, `cmd/crier`) that reports ProwJob status back to external systems
(GitHub status contexts, Slack, Pub/Sub, etc.) via pluggable reporters.

### Horologium

The component (`cmd/horologium`) that creates ProwJobs for periodic jobs on their configured
schedule.

### Sinker

The component (`cmd/sinker`) that garbage-collects old ProwJobs and their pods.

### Presubmit

A job type that runs against a PR before merge (e.g. tests gating merge).

### Postsubmit

A job type that runs after a change merges to a branch.

### Periodic

A job type that runs on a schedule, not tied to a specific PR or push, created by Horologium.

### InRepoConfig

The mechanism (`pkg/config`, enabled per-repo) allowing job configuration to live in the repo
being tested (`.prow.yaml` / `.prow/*.yaml`) instead of only in the central Prow Config. See
[`architecture/decisions/in-repo-prow-configuration.md`](architecture/decisions/in-repo-prow-configuration.md).

### Pod Utilities

The small binaries injected into ProwJob pods to handle cloning, artifact upload, and log
capture (`clonerefs`, `entrypoint`, `initupload`, `sidecar`), independent of the job's own
container image.

### Spyglass

Prow's pluggable UI for rendering job artifacts (test results, logs, coverage) in Deck.

### ghProxy

The caching reverse proxy (`cmd/ghproxy`, `pkg/ghcache`) Prow's components use in front of the
GitHub API to reduce rate-limit consumption.
