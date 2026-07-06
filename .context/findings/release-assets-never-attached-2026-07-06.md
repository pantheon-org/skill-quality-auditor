---
title: "Finding: Release assets never attached — every install path broken (immutable releases)"
type: FINDING
status: DONE
date: 2026-07-06
value: HIGH
themes:
  - DISTRIBUTION
related:
  - ../../docs/ADR/adr-020-multi-method-distribution.md
  - ../../docs/ADR/adr-008-homebrew-distribution.md
  - ../../docs/ADR/adr-023-version-sync-policy.md
---
# Finding: Release assets never attached — every install path broken (immutable releases)

> Releases v0.20.0–v0.23.0 shipped with zero binary assets because the workflow tried to
> attach artifacts *after* release-please published the release, which GitHub immutable
> releases forbid — silently breaking every documented install path.

## Summary

The `Release` workflow builds cross-platform binaries with goreleaser, then attaches them
in a separate `gh release upload` step that runs **after** release-please has already
*published* the GitHub release. GitHub immutable releases reject post-publish asset upload
with `HTTP 422: Cannot upload assets to an immutable release`. Every recent release
(v0.20.0 through v0.23.0) therefore has **0 assets**.

## Detail

- **Confirmed in run `28801006094`** (release `0.23.0`): the goreleaser build succeeded
  end to end (`release succeeded after 1m20s`, all six OS/arch archives + `checksums.txt`
  produced), then the upload step failed with the 422 immutable-release error and the job
  exited 1.
- `.goreleaser.yaml` has `release: disable: true`, so goreleaser never uploads anything
  itself — the pipeline relied entirely on the post-publish `gh release upload` step.
- release-please published the release live (non-draft), so it was frozen the instant it
  existed, leaving no window to attach assets.

### Blast radius

Every install path documented in ADR-020 depends on the missing assets and was silently
non-functional:

- **install.sh** — resolves the latest tag, downloads the matching tarball, verifies
  `checksums.txt` (the README headline `curl … | sh` one-liner)
- **Homebrew tap** (ADR-008) — formula points at release tarballs
- **mise `ubi` backend** — auto-discovers GitHub release assets
- **`skill-auditor update`** — pulls the newer binary from releases

## Recommended Action

Fixed in the same PR via a **draft-then-publish** flow (recorded as ADR-053):

1. release-please creates the release as a **draft** (`draft: true` in
   `release-please-config.json`) — drafts stay mutable under immutability.
2. The goreleaser job checks out the release SHA and **pushes the git tag explicitly** — a
   draft release does not create the tag ref, and goreleaser needs the tag to derive
   `{{.Version}}`.
3. goreleaser builds; `gh release upload` attaches assets to the **draft**.
4. `gh release edit --draft=false --latest` publishes, freezing the release **with** its
   assets attached — satisfying immutable releases.

### Residual gap

Existing empty releases (v0.20.0–v0.23.0) **cannot** be backfilled — immutability blocks
adding assets to already-published releases. Users remain broken until the next release
cuts. Cutting a patch release immediately after merge is the fastest way to restore a
working `install.sh`.
