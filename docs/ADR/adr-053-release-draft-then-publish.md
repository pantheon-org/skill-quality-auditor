---
title: "ADR-053: Release assets use a draft-then-publish flow (immutable releases)"
status: accepted
date: 2026-07-06
context:
  - path: .context/findings/release-assets-never-attached-2026-07-06.md
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

The `Release` workflow built cross-platform binaries with goreleaser but attached them in a
separate `gh release upload` step that ran **after** release-please had already *published*
the GitHub release. GitHub immutable releases reject post-publish asset upload with
`HTTP 422: Cannot upload assets to an immutable release`, so releases v0.20.0–v0.23.0
shipped with **zero assets**.

This silently broke every install path documented in ADR-020 (install.sh, Homebrew tap per
ADR-008, mise `ubi` backend, and `skill-auditor update`), plus the README headline
`curl … | sh` one-liner. See the finding for the confirming run (`28801006094`).

Immutability freezes a release the instant it is *published*. Assets can therefore only be
attached while a release is still a draft (drafts stay mutable). This constraint applies
regardless of which tool performs the upload, so "let goreleaser own the release" offers no
escape — a non-draft `goreleaser release` hits the same 422.

## Decision

Attach release assets via a **draft-then-publish** flow:

1. release-please creates the release as a **draft** (`draft: true` in
   `release-please-config.json`).
2. The goreleaser job checks out the release **SHA** and **pushes the git tag explicitly**.
   A draft release does not create the underlying git tag ref, and goreleaser needs the tag
   to derive `{{.Version}}`.
3. goreleaser builds (it keeps `release: disable: true` and remains a pure builder).
4. `gh release upload` attaches the archives + `checksums.txt` to the **draft**.
5. `gh release edit --draft=false --latest` publishes, freezing the release **with** its
   assets attached — satisfying immutable releases.

This operationalises ADR-020, ADR-008, and ADR-023, whose install paths were non-functional
until this fix.

## Consequences

- New releases carry their binary assets, restoring all documented install paths.
- The explicit **tag-push step** and the **`draft: true`** flag are load-bearing. They MUST
  NOT be "simplified" away: removing either reintroduces the HTTP 422 immutable-release
  failure and empty releases. This ADR is the record of *why* the odd tag-push exists.
- Publishing is now a distinct step, so a build failure leaves a draft (no half-published
  release) that can be retried or discarded.
- Existing empty releases (v0.20.0–v0.23.0) **cannot** be backfilled — immutability blocks
  it. Users remain broken until the next release cuts; a patch release immediately after
  merge is the fastest recovery.
