---
title: "Finding: Plan Frontmatter Status Drift"
type: finding
status: done
date: 2026-07-01
related:
  - ../instructions/ways-of-working.md
---

# Finding: Plan Frontmatter Status Drift

## Problem

8 of 9 dimension improvement plans (D1–D7, D9) had been fully implemented in code but their frontmatter still showed `status: active`. The `.context/index.yaml` faithfully reflected the stale frontmatter, making it appear that far more work remained than was actually the case.

## Root Causes

1. **No process** required updating plan status when implementation was done. Plans were written, coded, and the status field was never revisited.

2. **Hook excluded plans** — the `context-index` hook in `hk.pkl` excluded `.context/plans/**` from its glob, so even if someone updated a plan's status, the index would not regenerate on that change. The `check` was simply `test -f .context/index.yaml` which always passed.

3. **No freshness check** — the index was never validated against the actual files. A stale index was indistinguishable from a fresh one.

## Actions Taken

- Removed `.context/plans/**` from the `context-index` and `context-frontmatter` hook excludes in `hk.pkl`.
- Added `--check` flag to `regenerate-context-index.sh` that compares generated output against the current file and exits 1 if stale.
- Changed the hook `check` from `test -f .context/index.yaml` to `regenerate-context-index.sh --check`.
- Added `"instruction"` to the frontmatter schema enum.
- Updated 8 plan statuses from `active` to `done`.

## Recommendation

Update a plan's `status` to `done` in the same PR that implements it. The hook now enforces index freshness at commit time.
