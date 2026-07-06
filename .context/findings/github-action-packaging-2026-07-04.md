---
title: "Finding: Packaging skill-quality-auditor as a reusable GitHub Action"
type: finding
status: active
date: 2026-07-04
value: medium
related:
  - ../../.github/workflows/release.yml
  - ../../.github/workflows/skill-quality.yml
  - ../../.goreleaser.yaml
---
# Finding: Packaging skill-quality-auditor as a reusable GitHub Action

> Recommend packaging `skill-auditor` as a composite GitHub Action that downloads the pre-built release binary and runs it, rather than a Docker or JS/TS action. No implementation has happened yet — this is a finding awaiting a decision to proceed.

## Summary

The project already produces everything a reusable GitHub Action needs except the `action.yml` itself: goreleaser cross-compiles binaries for every OS/arch on each release, and `skill-quality.yml` already encodes the exact command sequence (`validate` → `duplication` → `batch --fail-below B` → `eval --fail-below 0`) that a reusable action should wrap. The lowest-effort, highest-fit packaging option is a **composite action** that downloads the matching release asset and invokes it.

## Detail

### Current release/build state

- `.goreleaser.yaml` builds `skill-auditor` for `linux`, `darwin`, `windows` × `amd64`, `arm64` (`CGO_ENABLED=0`), and archives as `tar.gz` (`zip` on Windows), with a `checksums.txt`. `release.disable: true` — goreleaser does not create the GitHub release itself.
- `.github/workflows/release.yml` runs `release-please-action@v4` on pushes to `main`. When it cuts a release, a second job runs `goreleaser-action@v6` (`release --clean`) against the new tag, then uploads `dist/*.tar.gz`, `dist/*.zip`, and `dist/checksums.txt` to that GitHub release via `gh release upload`.
- `release-please-config.json` already bumps `cmd/assets/tile.json`'s `version` field alongside the Go module version on every release PR — versioning is already unified across the CLI and the Tessl tile.
- `.github/workflows/skill-quality.yml` (the internal consumer of this tool) builds from source every run (`go build -o dist/skill-auditor .`) and then pipes the binary through:
  1. `./dist/skill-auditor validate artifacts`
  2. `./dist/skill-auditor duplication cmd/assets`
  3. `./dist/skill-auditor batch ./cmd/assets --fail-below B`
  4. `./dist/skill-auditor eval ./cmd/assets --fail-below 0` (structural gate; Tier 2 LLM-judge is advisory and separate)

  This is effectively hand-rolled "reusable action" logic, duplicated in-repo. Any other repo wanting the same quality gate today would have to copy these steps and vendor a Go build step.
- `LICENSE` exists at repo root (required for Marketplace publishing). No `action.yml`, `action.yaml`, or `Dockerfile` currently exists in the repo.

### Packaging options considered

| Option | Pros | Cons |
| --- | --- | --- |
| **Composite action** (`runs.using: composite`) downloading the release binary by OS/arch | No Go toolchain needed by consumers; reuses existing goreleaser artifacts directly; fast (no image pull); works on Linux/macOS/Windows runners | Requires a small shell step to resolve OS/arch → asset name and verify checksum |
| **Docker container action** | Hermetic, pinned exact binary and environment | Linux-only on hosted runners; image pull/build latency every job; leaves the darwin/windows goreleaser builds unused by CI |
| **JS/TS action** (`@actions/toolkit`, `tool-cache`) | Cross-platform tool caching across jobs; more idiomatic for polished Marketplace listings | Requires `ncc` bundling step and a second versioning/release surface to keep in sync with the Go binary; more moving parts than needed today |

## Recommended Action

Package as a **composite GitHub Action**:

- Add `action.yml` at repo root (`runs.using: composite`) that:
  1. Resolves `runner.os` / `runner.arch` to the matching goreleaser asset name (`skill-auditor_{os}_{arch}.tar.gz` / `.zip`).
  2. Downloads that asset from the tagged release (default to a pinned version input, e.g. `version: latest` or a specific tag) and verifies against `checksums.txt`.
  3. Extracts the binary onto `PATH` for the job.
  4. Exposes inputs mirroring `skill-quality.yml`'s steps — e.g. `path` (defaults to `cmd/assets`), `fail-below`, `commands` (subset of `validate`/`duplication`/`batch`/`eval` to run) — and runs them.
- Update `README.md` with a "Use as a GitHub Action" section showing `uses: pantheon-org/skill-quality-auditor@v1`.
- Once the action exists, consider migrating `skill-quality.yml` itself to consume it (dogfooding), replacing the from-source build with the packaged binary download — this would also cut CI time since it skips `go build`.

No implementation has started. Open questions before implementation:

- Should the action pin to a specific released version by default, or float on `latest`?
- Does `batch`/`eval` output need to be surfaced as Action `outputs` (e.g. `score`, `grade`) for downstream steps, or is exit-code gating sufficient?
- Should this repo's own `skill-quality.yml` be migrated to dogfood the new action in the same PR, or as a follow-up?
