---
title: "ADR-054: All GitHub Actions pinned by commit SHA (first-party exemption removed)"
status: accepted
date: 2026-07-06
context:
  - path: .context/findings/actions-sha-pin-all-2026-07-06.md
  - path: .context/findings/plumber-cicd-security-2026-07-04.md
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

`.plumber.yaml`'s `actionsMustBePinnedByCommitSha` control shipped (ADR-036 / ADR-037)
with the tool's default `trustedOwners: [actions, github]`, which exempts first-party
actions from the pin-by-SHA requirement. The original finding
(`plumber-cicd-security-2026-07-04.md`) scoped SHA-pinning to the five third-party actions
only. The node24 migration then bumped `actions/*` to current majors but left them on
floating tags (`@v7`), which plumber accepted because of the exemption.

The tj-actions/changed-files compromise (CVE-2025-30066) is a mutable-ref supply-chain
attack: a retagged or compromised release runs with the caller workflow's secrets. That
vector applies to **any** mutable ref — a floating `@v7` on a first-party action is
silently repointable too. Leaving the exemption in place means the policy is followed by
convention but not enforced, so tag pins can drift back unnoticed.

## Decision

1. **Every** `uses:` reference in `.github/workflows/**` is pinned by a 40-char commit
   SHA — first-party `actions/*` included — with an exact-version trailing comment
   (`# v7.0.0`).
2. `.plumber.yaml` sets `actionsMustBePinnedByCommitSha.trustedOwners: []` so the control
   **enforces** SHA-pinning on all owners, not just third-party.
3. Dependabot (`.github/dependabot.yml`, github-actions ecosystem) keeps the SHA pins and
   their version comments fresh via `sha-and-version` behaviour.

This tightens, and supersedes in scope, the third-party-only SHA-pinning posture recorded
in ADR-036 (Decision 6) and the default exemption noted in
`plumber-cicd-security-2026-07-04.md`.

## Consequences

- **Easier:** a compromised/retagged release of a first-party action can no longer be
  silently pulled into CI; every action runs a reviewed, immutable commit.
- **Easier:** the policy is machine-enforced — `plumber analyze` fails if any ref
  (any owner) is not SHA-pinned, so drift is caught in CI rather than by convention.
- **Harder:** SHAs are opaque; readers rely on the `# vX.Y.Z` comment for the human-readable
  version. Dependabot must stay enabled to keep pins current, otherwise they ossify.
- `plumber analyze` scores 100/100 with `trustedOwners: []`, confirming no ref was missed.
