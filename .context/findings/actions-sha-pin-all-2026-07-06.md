---
title: "Finding: All GitHub Actions pinned by commit SHA (first-party exemption removed)"
type: FINDING
status: DONE
date: 2026-07-06
value: HIGH
themes:
  - GOVERNANCE
related:
  - ../../docs/ADR/adr-036-plumber-cicd-security-advisory-workflow.md
  - ../../docs/ADR/adr-037-plumber-critical-fail-issue-tracking.md
  - plumber-cicd-security-2026-07-04.md
---
# Finding: All GitHub Actions pinned by commit SHA (first-party exemption removed)

> The node24 action sweep initially left first-party `actions/*` on floating major
> tags (`@v7`), consistent with `.plumber.yaml`'s default `trustedOwners: [actions, github]`
> exemption. Repo policy is now hardened to require **every** action — first-party
> included — to be pinned by commit SHA, and `.plumber.yaml` enforces it.

## Summary

`.plumber.yaml`'s `actionsMustBePinnedByCommitSha` control shipped with the default
`trustedOwners: [actions, github]`, which exempts first-party actions from the pin-by-SHA
requirement (see the original finding `plumber-cicd-security-2026-07-04.md`, which scoped
SHA-pinning to the five third-party actions only). The node24 migration bumped `actions/*`
to current majors but kept them tag-pinned, so plumber stayed green without them being
SHA-pinned.

## Detail

- The tj-actions/changed-files compromise (CVE-2025-30066) is a mutable-ref attack: a
  retagged or compromised release runs with the caller workflow's secrets. This vector
  applies to **any** mutable ref, not only third-party ones — a floating `@v7` is
  silently repointable.
- Leaving `trustedOwners: [actions, github]` means the pin-by-SHA policy is not *enforced*
  on first-party actions, so tag pins could drift back in unnoticed.

## Recommended Action (applied this PR — ADR-054)

1. Pin every first-party `actions/*` reference by 40-char commit SHA with an exact-version
   comment (`# v7.0.0`), matching the existing third-party convention.
2. Set `.plumber.yaml` `actionsMustBePinnedByCommitSha.trustedOwners: []` so the control
   enforces SHA-pinning on all owners.
3. Dependabot (`.github/dependabot.yml`, github-actions ecosystem) keeps the SHA pins and
   their version comments fresh.

`plumber analyze` scores 100/100 with the tightened config, confirming no ref was missed.
