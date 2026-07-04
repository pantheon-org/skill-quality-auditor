---
title: "ADR-035: Package skill-quality-auditor as a composite GitHub Action"
status: proposed
date: 2026-07-04
context:
  - path: .context/findings/github-action-packaging-2026-07-04.md
---

**Status:** Proposed
**Date:** 2026-07-04

## Context

`skill-quality-auditor` already produces multi-platform release artifacts:
`.goreleaser.yaml` cross-compiles `skill-auditor` for `linux`/`darwin`/`windows`
× `amd64`/`arm64` on every `release-please`-driven release, and
`.github/workflows/skill-quality.yml` hand-rolls the exact quality-gate
sequence (`validate` → `duplication` → `batch --fail-below B` →
`eval --fail-below 0`) by building the binary from source each run. Other
repositories that want the same quality gate today would need to copy that
workflow and vendor a Go build step. See
`.context/findings/github-action-packaging-2026-07-04.md` for the full
comparison of packaging options.

## Decision

Package `skill-auditor` as a **composite GitHub Action** (`action.yml` at
repo root, `runs.using: composite`) that downloads the pre-built release
binary matching the runner's OS/arch, verifies it against
`checksums.txt`, and runs the requested subcommands — rather than a Docker
container action (Linux-only, image-pull overhead, leaves the darwin/windows
goreleaser builds unused) or a JS/TS action (requires `ncc` bundling and a
second versioning surface to keep in sync with the Go binary).

This ADR is **proposed**, not yet accepted: implementation has not started,
and the following are still open before work begins:

- Default version pin (a specific released tag) vs. floating on `latest`.
- Whether `batch`/`eval` results need to be surfaced as Action `outputs`
  (e.g. `score`, `grade`) or whether exit-code gating is sufficient.
- Whether this repo's own `skill-quality.yml` should be migrated to dogfood
  the new action in the same PR, or as a follow-up.

## Consequences

If accepted, other repositories gain a drop-in
`uses: pantheon-org/skill-quality-auditor@v1` step instead of hand-rolling a
Go build + CLI invocation, and this repo's own `skill-quality.yml` could
drop its `go build` step in favor of the packaged binary, cutting CI time.
The tradeoff is an additional release-time contract (`action.yml`'s
asset-naming assumptions must stay in lockstep with `.goreleaser.yaml`) and
a new surface to keep documented as goreleaser's build matrix evolves.
