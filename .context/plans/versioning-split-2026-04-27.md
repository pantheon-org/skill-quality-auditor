---
title: "Plan: Fix release-please tile.json Path"
type: plan
status: done
date: 2026-04-27
---
# Plan: Fix release-please tile.json path (wrong path bug)

## Problem

`release-please-config.json` references `tile.json` under `extra-files` with the wrong path:

```json
"path": "skill-auditor/cmd/assets/tile.json"
```

The actual path in the repo is `cmd/assets/tile.json`. Release-please has been silently failing to
bump `tile.json` on every release, so it has drifted out of sync with the binary version.

## What release-please already provides (no changes needed)

The automation for automatic version bumping is already fully in place and correct:

```
PR merged to main (conventional commit: feat/fix/feat!)
  → release-please reads commit types → determines semver bump
  → opens/updates a release PR that bumps .release-please-manifest.json
  → release PR merged → git tag vX.Y.Z created
  → GoReleaser runs on tag
      → injects -X cmd.buildVersion={{.Version}} via ldflags  ✓
      → binary version = git tag                              ✓
  → release-please also bumps tile.json via extra-files       ✗ (broken path)
```

Binary and tile versions are **intentionally in sync** — this is a single repo where the CLI and
skill asset are co-released. Keeping them at the same version via release-please is the right
approach; no decoupling is needed.

The `cmd/root.go` fallback that reads `tile.json` at dev-build time is also correct and stays
unchanged.

## Root cause

The `extra-files` path was written as `skill-auditor/cmd/assets/tile.json` (likely copy-pasted
from a monorepo context) instead of `cmd/assets/tile.json`.

## Implementation steps

### Step 1 — Fix the path in `release-please-config.json`

```json
// Before (broken):
"path": "skill-auditor/cmd/assets/tile.json"

// After (correct):
"path": "cmd/assets/tile.json"
```

### Step 2 — Verify tile.json version matches the manifest

`.release-please-manifest.json` tracks `"." : "0.1.5"` and `tile.json` currently has
`"version": "0.1.5"` — they are in sync. No correction needed.

### Step 3 — Update `CLAUDE.md`

Add a note clarifying that `tile.json` version is managed automatically by release-please and
should **not** be bumped manually.

---

## Files changed

| File | Change |
|------|--------|
| `release-please-config.json` | fix `extra-files` path: `skill-auditor/cmd/assets/tile.json` → `cmd/assets/tile.json` |
| `CLAUDE.md` | note that tile.json version is auto-managed by release-please |

---

## What this does NOT change

- GoReleaser config — ldflags injection is already correct.
- `cmd/root.go` — tile.json fallback for dev builds is correct and stays.
- `.release-please-manifest.json` — already in sync at `0.1.5`.
- The `release.yml` / GoReleaser workflow — tag-driven, unchanged.
- The `update` command — compares against GitHub release tags, not tile.json.

---

## Out of scope

- Decoupling binary and tile versions into independent release tracks. Given this is a single
  co-released repo, in-sync versions are simpler and sufficient.
- A CI gate enforcing manual tile bumps — unnecessary since release-please handles it.
