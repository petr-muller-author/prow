---
status: Accepted
date: 2019-10-06
audience: [contributor, configurator]
---

# In-repo Prow configuration (InRepoConfig)

## Context

This ADR documents an existing decision retroactively — it predates this `docs/` tree. The
`date` above is a best-effort estimate: the earliest commit found (via `git log -S"InRepoConfig"`
on this repo's full history, inherited from `kubernetes/test-infra`) that introduces the feature
is `7265046` ("Allow enabling inrepoconfig", 2019-10-06). The actual proposal/discussion predates
that commit by an unknown amount; treat this date as "feature landed around," not "decision was
made on."

Originally, all Prow job configuration (`config.yaml`) and plugin configuration
(`plugins.yaml`) lived centrally, managed by whoever administers the Prow instance (the
`operator`/instance-admin persona). For an instance serving many repos — as `kubernetes/*` does —
every job addition or change for any repo required a PR against the central config, reviewed and
merged by instance admins. This doesn't scale well as the number of served repos grows: admins
become a bottleneck for routine job changes, and repo owners can't iterate on their own CI
without going through a different team.

## Decision

Allow individual repos to define some of their own Prow job configuration in-repo (`.prow.yaml`
or `.prow/*.yaml` at the repo root), loaded and merged with central config by `pkg/config` at
job-trigger time, gated by an explicit per-repo opt-in in the central config
(`InRepoConfig` allowed-clusters/enablement settings — see
`pkg/plugins/trigger` allowed-cluster checks and `pkg/config`'s `InRepoConfig` types).

## Consequences

- Repo owners (`configurator` persona) can add/change their own presubmits/postsubmits without a
  central-config PR, as long as their repo is enabled for in-repo config.
- Central config remains the source of truth for *whether* a repo is trusted to define its own
  jobs at all, and for anything with instance-wide implications (cluster access, global plugin
  behavior) — in-repo config is deliberately scoped to job definitions, not everything
  `config.yaml` can express.
- `pkg/config` gained meaningfully more complexity: it must fetch, validate, and merge in-repo
  config at trigger time (with caching — see `ProwYAMLCache` in `pkg/config`), rather than
  loading one static file at startup.
- Because in-repo config is fetched from the PR/branch being tested, it's part of the trust
  surface `trigger`'s author-trust check protects — an untrusted PR can't use in-repo config to
  define jobs that run before a trusted `/ok-to-test`.

## Alternatives considered

- **Central-only config (status quo at the time)**: rejected as the driver for this decision —
  didn't scale to the number of repos/orgs a single Prow instance serves.
- **Full delegation (repo fully controls its own Prow behavior)**: not adopted — plugin
  enablement, cluster access, and other instance-wide settings remain centrally controlled, so a
  compromised or careless repo config can't affect other repos or the instance's security
  posture.

## Related code

- `pkg/config` — in-repo config loading, merging, caching (`ProwYAMLCache` and related).
- `pkg/plugins/trigger` — trust gating that in-repo config's job definitions are still subject to.

## Related docs

- [`../../glossary.md#inrepoconfig`](../../glossary.md#inrepoconfig)
- Hugo: [`inrepoconfig.md`](../../../site/content/en/docs/inrepoconfig.md) — user-facing
  reference for `configurator`s on how to actually enable and use this.
