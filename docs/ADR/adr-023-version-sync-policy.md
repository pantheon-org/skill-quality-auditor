---
title: "ADR-023: Keep CLI binary and tile.json versions in sync via release-please"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/versioning-split.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

`release-please-config.json` referenced `tile.json` at the wrong path (`skill-auditor/cmd/assets/tile.json` instead of `cmd/assets/tile.json`), causing release-please to silently fail at bumping the tile.json version on every release. The tile.json version drifted out of sync with the binary version. A question arose: should the tile.json version be decoupled from the CLI binary version?

## Decision

1. **Fix the path** in `release-please-config.json` to `cmd/assets/tile.json`
2. **Keep CLI binary and tile.json versions in sync** — intentional decision NOT to decouple them. The tile is distributed from this repo, co-located with the binary source. A version mismatch would confuse consumers who install the binary and expect the tile to match.
3. **Do not manually bump tile.json version** — release-please auto-bumps it alongside the binary version on every release PR. Manual bumps cause merge conflicts with release-please.

## Consequences

- Single version number for both binary and tile — consumers get matching versions
- Release-please manages both bumps atomically
- No manual version management needed for tile.json
- Decoupling would require a separate release process for the tile (not warranted for a co-located asset)
