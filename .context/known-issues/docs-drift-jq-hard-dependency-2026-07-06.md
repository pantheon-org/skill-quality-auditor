---
title: "Known Issue: check-docs-drift.sh hard-fails pre-push if jq is missing"
type: known-issue
status: done
date: 2026-07-06
severity: critical
related:
  - ../plans/docs-drift-reviewed-baseline-2026-07-06.md
  - ../../docs/ADR/adr-045-docs-drift-reviewed-baseline.md
  - ../../scripts/check-docs-drift.sh
  - ../../scripts/mark-docs-reviewed.sh
---
# Known Issue: check-docs-drift.sh hard-fails pre-push if jq is missing

> Verified live (not speculative): with `jq` removed from `PATH`, `scripts/check-docs-drift.sh`'s cumulative mode exits **127** with a bare "command not found" — it does not degrade gracefully. Since this script runs on every contributor's local `pre-push` (via `hk.pkl`), any contributor without `jq` on `PATH` gets their ability to push **broken**, not just a missing feature, the first time they push after ADR-045's reviewed-baseline sidecar work merges.

## Why this exists

ADR-045 (`.context/plans/docs-drift-reviewed-baseline-2026-07-06.md`) added a `lookup_reviewed()` function to `check-docs-drift.sh` that shells out to `jq` to query the new `scripts/docs-drift-reviewed.jsonl` sidecar, on every `MAPPINGS` iteration, every time cumulative mode runs. `jq` was already an unstated dependency of three other scripts in this repo (`plumber-*.sh`), so it was judged safe to reuse without adding it to `mise.toml` — but those three scripts only ever run in CI (where `ubuntu-latest` ships `jq`), never at local pre-push. `check-docs-drift.sh`'s cumulative mode is the first `jq`-dependent script in this repo that runs on a contributor's own machine on every push, and `set -euo pipefail` means a missing `jq` aborts the entire script rather than being caught and reported cleanly.

Caught during a session-reflection sub-agent review that tested the failure live, rather than assuming `jq`'s presence on 3 prior scripts implied safety for a 4th, differently-triggered one.

## Impact if unfixed

A contributor whose machine lacks `jq` (not on `PATH` via Homebrew, system package, or personal dotfiles) pushes normally, `hk`'s `pre-push` hook runs `check-docs-drift.sh`, and the push fails with a cryptic `jq: command not found` / exit 127 — no guidance that `jq` is the missing piece, no indication this is unrelated to their actual change. Blocks pushing entirely until they either install `jq` or discover `--no-verify`.

## Fix applied

Chose option (a) from the original suggested-fix list: `check-docs-drift.sh` now checks `command -v jq` once at startup; if absent, it prints one clear warning ("jq not found on PATH — skipping reviewed-baseline lookups...") and `lookup_reviewed()` short-circuits, degrading to the pre-ADR-045 doc-edit-only comparison rather than the script itself. `mark-docs-reviewed.sh` (whose entire purpose requires `jq`) instead fails fast with a clear, actionable error message naming the missing dependency and how to install it, rather than an opaque `jq: command not found`.

Verified live, reproducing the exact failure mode this issue reported: built a curated `PATH` with symlinks to every tool the script needs except `jq`, confirmed `command -v jq` genuinely fails in that environment, then ran both scripts under it. `check-docs-drift.sh` now exits 0 with the warning and correctly falls back to flagging all docs by edit-date only (including previously-reviewed ones, since the reviewed-baseline lookup is unavailable without `jq`) instead of exit 127. `mark-docs-reviewed.sh` now exits 1 with the clear dependency message instead of an unguarded jq invocation failing partway through.
