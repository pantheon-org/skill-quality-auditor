---
title: "ADR-057: CalVer YYYY.0M.PATCH considered and rejected (stay on SemVer)"
status: rejected
date: 2026-07-06
context:
  - path: docs/ADR/adr-020-multi-method-distribution.md
  - path: docs/ADR/adr-023-version-sync-policy.md
  - path: docs/ADR/adr-008-homebrew-distribution.md
---

**Status:** Rejected
**Date:** 2026-07-06

## Context

A proposal was raised to replace SemVer (`0.24.0`, driven by release-please from Conventional
Commits) with a calendar version `YYYY.0M.PATCH` — four-digit year, zero-padded two-digit
month, and a PATCH resetting to 0 each month (`v2026.07.0`, `v2026.07.1`, `v2026.08.0`). The
goal was a version that shows *when* a release shipped at a glance.

The proposition was stress-tested with a structured design debate (Advocate / Skeptic /
Migration-Risk), grounded in the repo's actual version consumers. This ADR records the
outcome: **rejected, stay on SemVer.**

## Decision

Do **not** adopt CalVer. Keep SemVer via release-please. If a date signal is wanted, surface
the build date in `skill-auditor version` output and release notes instead (a separate,
zero-breakage change).

## Why — what the debate established

**The local install chain is format-agnostic (this argument held up).** Verified against the
repo: `scripts/install.sh` resolves the release via GitHub's `releases/latest` API (release.yml
marks the release `--latest`, ADR-053); `cmd/update.go` compares versions by pure string
equality with no semver library in `go.mod`; goreleaser uses `{{.Version}}` only as an ldflags
string with `release: disable: true` and no `{{.Major}}` helpers; the Homebrew `brews:` block
is commented out. None of these would break on a CalVer string.

**But `go install` breaks — decisively, and under both padded and non-padded variants.**
`go install github.com/pantheon-org/skill-quality-auditor@latest` is a documented headline
install path (README, docs/index.md, ADR-020), and the module path has no `/vN` suffix.
Go's module rules mean:

- `v2026.07.0` — leading zero is invalid SemVer; Go rejects it outright.
- `v2026.7.0` — valid SemVer but major version 2026; Go requires the module path to end in
  `/v2026`. It does not, so `@latest` silently keeps resolving the last `0.x` tag and never
  sees any CalVer release. A silent freeze, worse than a hard error.

The non-padded "valid SemVer" variant therefore does **not** rescue the one consumer that
actually parses SemVer. The only escapes are dropping `go install` as a supported path or
adding a `/v2026` module suffix (which rewrites every import path — a non-starter).

**Ongoing and irreversible costs (Migration-Risk).** release-please has no native CalVer and
no monthly-reset; adopting it means replacing working automation with a bespoke action
(month rollover, intra-month tie-breaking, non-SemVer `tile.json` sync) sized L/borderline XL,
maintained forever. Two hard gates remain unverified: the ubi/mise *pinned/upgrade* path and
whether the Tessl registry validates `tile.json $.version` as SemVer. And it is a one-way door
— `2026.x` sorts above `0.x` forever, so the first real CalVer tag is irreversible.

**The zero-padded variant specifically was the worst option:** it converts "unusual but
parseable" into "rejected by every strict SemVer parser" for pure column alignment.

**Benefit vs cost.** For a pre-1.0 CLI with no public API contract, "when did it ship" is
already available from tag dates, the changelog, and release metadata. The payoff does not
cover a permanent, partly-irreversible cost.

## Consequences

- SemVer and release-please remain unchanged. ADR-020/023/008/053 keep their SemVer
  assumptions intact.
- The date-at-a-glance need is redirected to a cheaper change: inject `buildDate` via
  goreleaser `{{ .CommitDate }}` ldflags and print it from the `version` command.

## Revisit trigger

Reopen CalVer only if **both** hold: (a) `go install` is formally dropped from the supported
install paths in README, docs, and ADR-020; **and** (b) the ubi/mise-pinned and
Tessl-registry gates are verified green against a throwaway `-rc` pre-release CalVer tag. If
reopened, use the non-padded `YYYY.M.PATCH` variant — it is strictly less bad than zero-padded.
