---
title: "Known Issue: check-docs-drift.sh hard-fails pre-push if jq is missing"
type: known-issue
status: active
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

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

Add a `command -v jq >/dev/null` guard near the top of `check-docs-drift.sh`'s cumulative-mode branch (and `mark-docs-reviewed.sh`, which has the same dependency). On missing `jq`, print a clear one-line explanation and either: (a) skip the reviewed-baseline lookup and fall back to the pre-ADR-045 doc-edit-only comparison (degrades the feature, not the script), or (b) skip cumulative mode's drift check entirely with a warning (matches the script's existing "informational only" character — never block a push over a missing optional tool). Option (a) is closer to graceful degradation; option (b) is simpler. Neither has been decided yet.
