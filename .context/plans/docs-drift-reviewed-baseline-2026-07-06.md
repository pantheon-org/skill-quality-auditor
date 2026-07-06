---
title: "Plan: Reviewed-baseline mechanism for check-docs-drift.sh"
type: plan
status: done
date: 2026-07-06
effort: S
related:
  - ../findings/docs-drift-perpetual-noise-2026-07-06.md
  - ../findings/gh-pages-docs-drift-2026-07-05.md
  - ../../docs/ADR/adr-044-docs-drift-pr-gate.md
  - ../../docs/ADR/adr-045-docs-drift-reviewed-baseline.md
  - ../../scripts/check-docs-drift.sh
  - ../../hk.pkl
---

# Plan: Reviewed-baseline mechanism for check-docs-drift.sh

> Reviewed by 3 independent Claude Sonnet 5 subagents (Technical, Strategic, Risk) on 2026-07-06. All three perspectives' actionable findings are folded in below; resolved decisions replace what was originally left as Open Questions. Implemented and verified the same day — see Implementation Notes for two things that changed from the reviewed design during coding.

## Goal

Give `scripts/check-docs-drift.sh`'s cumulative (pre-push, advisory) mode — consumed by `hk.pkl`'s `pre-push` hook, its only caller — a way to record "this doc was reviewed and confirmed still accurate as of commit X" separately from "this doc's content was last edited at commit X." Today the script has only one signal — the doc file's own git history — so a doc that's genuinely fine but hasn't been re-edited keeps reappearing with a growing, meaningless commit count on every future push, forever. See `.context/findings/docs-drift-perpetual-noise-2026-07-06.md` for the concrete case (4 docs, real today) this plan fixes.

## Scope

**In scope:**

- A new tracked sidecar file recording, per doc, the date/commit of the most recent explicit review (separate from `check-docs-drift.sh`'s existing `MAPPINGS` array, which encodes doc-to-source relationships and changes far less often).
- A small helper script that stamps a review into that sidecar for one or more doc paths.
- A change to `check-docs-drift.sh`'s cumulative mode so its per-doc baseline is the later of the doc's own last-commit date and its sidecar review date, if one exists — computed via integer epoch comparison, not ISO8601 string comparison (see Decisions below).
- Applying the new helper to the 4 currently-flagged false positives (`README.md`, `docs/index.md`, `docs/architecture/duplication-flow.md`, `docs/reference/scoring-dimensions.md`) as part of implementation.

**Out of scope:**

- Any change to gate mode (the `pull_request`-only CI check added in ADR-044). It only ever diffs the current PR against its base ref via `git log --oneline "$BASE_REF"...HEAD -- <path>`, never accumulates historical debt, and needs nothing from this plan.
- Any change to `MAPPINGS` itself — this plan doesn't add, remove, or re-scope which source globs map to which docs.
- Automating the review decision (e.g. an LLM judging "is this doc still accurate"). The review itself stays a human (or agent-assisted, human-confirmed) judgment call; this plan only makes recording that judgment durable. `mark-docs-reviewed.sh` trusts its invoker — it is an advisory bookkeeping tool, not an access-controlled gate, matching the trust model of the rest of `check-docs-drift.sh` (which itself never blocks anything).
- A JSON-Schema-validated template for the sidecar file. The repo's "template + schema + validation script" convention (`.agents/RULES.md`) governs skill artifacts authored under `.context/plugins/`; a small JSONL file of internal tooling state is not that kind of artifact, and adding a schema for it would be disproportionate to an S-sized fix.
- Cleanup of orphaned sidecar rows when a doc is renamed or deleted. `check-docs-drift.sh` already skips any `MAPPINGS` entry whose `doc_file` doesn't exist (`[ -f "$doc_file" ] || continue`); an orphaned sidecar row for a deleted doc is simply never read again. Accepted as harmless dead data, not worth a cleanup step at this scope.

## Decisions (resolved during plan-review — see review findings below for the "why")

1. **Sidecar stores both an ISO8601 date and a pre-computed epoch integer, not just one**, as JSONL (one JSON object per line): `{"doc": "...", "reviewed_iso": "...", "reviewed_epoch": N}`. The ISO field is for human readability in `git diff`/PR review; the epoch field is what `check-docs-drift.sh` actually compares against. This avoids all ISO8601 string-comparison/timezone-offset risk the Technical and Risk reviews both flagged — integer comparison is exact and portable, unlike lexical comparison of differently-offset ISO strings. (Format changed from an originally-planned TSV to JSONL mid-implementation — see Implementation Notes.)
2. **`check-docs-drift.sh` fetches both `%cI` and `%ct` for a doc's last-commit in one `git log` call** (`--format='%cI|%ct'`), so the existing human-readable warning message is untouched while the epoch becomes available for comparison with no extra `git` invocation.
3. **The effective "since" cutoff for `git log --after` uses git's own `@<epoch>` date syntax** (verified live: `git log --after="@$epoch"` produces identical results to the existing `--after="$iso_date"` form) — so no external `date` binary parsing is introduced anywhere, sidestepping the GNU-vs-BSD `date` portability gap entirely (this script runs both in GitHub Actions/Ubuntu and on contributors' local macOS machines via the pre-push hook).
4. **`mark-docs-reviewed.sh` does not hard-block marking a doc that isn't currently flagged**, but it does print what it's recording (the commits between the doc's last-touch date and now for each mapped source glob, if any) so the action is visible and auditable rather than silent — addressing the Risk review's "an agent could self-mark everything reviewed to silence the check" concern without adding a rigid gate that would block the legitimate case of proactively confirming a doc the same session a source change lands.
5. **The sidecar needs no `MAPPINGS`/docs-drift coverage of its own.** It's tooling-internal state, not a user-facing doc — closing the third original Open Question.

## Implementation Notes (things that changed while coding, vs. the reviewed design)

- **TSV → JSONL.** The plan-reviewed design used a `doc_path<TAB>reviewed_date_iso<TAB>reviewed_epoch` TSV. Changed to JSONL (`{"doc": ..., "reviewed_iso": ..., "reviewed_epoch": ...}`) per explicit preference during implementation. `jq` (already an assumed-available dependency elsewhere in `scripts/` — see `plumber-*.sh`) replaces the `awk`/`sort`/`mv` TSV manipulation in both scripts. No behavioral difference from the reviewed design; same insert-or-replace, same idempotency guarantee, verified with the same test sequence.
- **No `declare -A` (bash associative arrays).** The reviewed design's Phase 2 step 1 called for loading the sidecar into an associative array. This broke immediately on this machine's `/usr/bin/env bash`, which resolves to macOS's stock `/bin/bash` 3.2.57 — a pre-bash-4 build with no associative arrays, despite a newer bash being installed elsewhere on the same machine (`/opt/homebrew/opt/bash`, not first in `PATH` for `env bash`'s resolution). Since this script runs on contributors' local Macs via the pre-push hook, not just CI, this is a real and likely-common failure mode, not a one-off. Fixed by replacing the associative-array preload with a per-doc `lookup_reviewed()` function that queries the JSONL sidecar directly via `jq` on each `MAPPINGS` iteration — no upfront loading, no bash-version dependency.
- **Verification gotcha, not a design change:** the plan's Verification section originally ran the regression test's `git reset --hard HEAD~1` while Phase 1–2's own code changes were still uncommitted working-tree edits (not yet part of any commit). `git reset --hard` discards uncommitted changes to tracked files, not just the target commit — it wiped the just-written `check-docs-drift.sh` changes along with the intended throwaway test commit. Re-applied from scratch (same diff) and re-verified before committing anything further. **Corrected verification order below: commit the real implementation before running any throwaway-commit-based regression test that ends in a hard reset.**

## Phases

### Phase 1 — Sidecar file and helper script

1. Add `scripts/docs-drift-reviewed.jsonl` (starts empty). One JSON object per line: `{"doc": "...", "reviewed_iso": "...", "reviewed_epoch": N}`.
2. Add `scripts/mark-docs-reviewed.sh <doc_path> [doc_path...]`:
   - For each argument: verify the doc path exists in `check-docs-drift.sh`'s `MAPPINGS` (warn, don't fail, if not — likely a typo since an unmapped entry is silently never read).
   - Fetch `HEAD`'s commit date via `git log -1 --format='%cI'` / `--format='%ct'`.
   - Print the commits (if any) between the doc's current last-touch date and now for each mapped source glob, so the review action is auditable in the terminal even though it isn't hard-blocked (Decision 4).
   - Replace-or-insert: `jq -c --arg d "$doc" 'select(.doc != $d)'` to filter the existing sidecar into a temp file, append the new JSON line via `jq -nc`, re-sort by `doc` for diff-friendliness, `mv` over the sidecar.
   - Idempotent — running it twice against the same `HEAD` for the same doc produces no further diff.
3. `shellcheck scripts/mark-docs-reviewed.sh` clean; `set -euo pipefail`; pure bash — consistent with `check-docs-drift.sh`'s existing style (not because of `.agents/RULES.md`'s Python/Node rule, which is scoped to skill scripts under `.context/plugins/`, not top-level `scripts/` — corrected from the original draft's imprecise citation).

Exit criterion: running `scripts/mark-docs-reviewed.sh README.md` on a clean checkout adds exactly one correctly-formatted JSON line to the sidecar; running it again immediately produces no further diff. **Verified.**

### Phase 2 — Wire the baseline into cumulative mode

1. Add a `lookup_reviewed()` function to `check-docs-drift.sh`'s cumulative-mode branch that queries the JSONL sidecar (if present) for a given `doc_path` via `jq`, per-lookup rather than preloaded into an associative array (see Implementation Notes — bash 3.2 has none).
2. For each `MAPPINGS` entry, fetch the doc's `%cI|%ct` in one call (Decision 2). Compute `effective_epoch = max(doc_epoch, reviewed_epoch)` via integer comparison if a sidecar entry exists, else `effective_epoch = doc_epoch`. Skip and warn (don't crash) on a malformed/non-numeric sidecar entry.
3. Use `git log --oneline --after="@$effective_epoch" -- "$g"` in place of the current `--after="$doc_date"` (Decision 3). Keep displaying the human-readable date in the warning message — if the sidecar's reviewed date is the effective baseline, show that instead of the doc's edit date, so the message truthfully reflects what's actually suppressing (or not suppressing) the warning.
4. Gate mode (the `BASE_REF` branch) is untouched — it doesn't use `doc_date` or the sidecar at all today, and this plan doesn't change that.
5. Update the script's header comment to document the sidecar's role alongside the existing gate-mode/cumulative-mode comment block added in ADR-044.

Exit criterion: a doc with a sidecar entry whose epoch is newer than its most recent flagged source commit no longer appears in `check-docs-drift.sh`'s output; a doc with no sidecar entry behaves exactly as it does today (no regression for the 4 docs already fixed via PR #186's content changes). **Verified.**

### Phase 3 — Apply, verify, and regression-test

1. Run `scripts/mark-docs-reviewed.sh README.md docs/index.md docs/architecture/duplication-flow.md docs/reference/scoring-dimensions.md`.
2. Run `./scripts/check-docs-drift.sh` with no args and confirm it reports **zero** stale docs (achieved: this branch is rebased onto `main` post-PR-#186, which already closed the other 4 originally-flagged docs). **Verified.**
3. Run `./scripts/check-docs-drift.sh origin/main` (gate mode) and confirm it still reports no new drift — proving Phase 2 didn't touch gate-mode behavior. **Verified.**
4. **Regression test:** commit the real implementation first (per Implementation Notes' gotcha), then make a throwaway commit touching one of a reviewed doc's mapped source globs (e.g. `cmd/root.go`), confirm `check-docs-drift.sh` **does** re-flag that doc with its warning showing the "(reviewed)" baseline was surpassed — proving the reviewed-mark only suppresses up to the reviewed commit, not permanently — then discard the throwaway commit via `git reset --hard HEAD~1` (safe now that real work is committed). **Verified**, including confirming discard didn't lose anything.
5. Run the repo's standard gates: `hk check && go test ./...`.

Exit criterion: cumulative mode reports 0 flagged docs; gate mode behavior is bit-for-bit unchanged; the reviewed-then-re-edited regression test correctly re-flags; `hk check && go test ./...` pass clean.

## Known, accepted limitations (not fixed by this plan)

- **Stale local checkout.** If `mark-docs-reviewed.sh` runs against a local `HEAD` that's behind `origin/main`, the recorded review silently doesn't account for commits that landed on `main` in the meantime. Mitigated only by usage guidance (fetch/rebase before marking), not enforced — consistent with this being an advisory tool, not a security control.
- **Concurrent sidecar edits.** Two branches marking different docs reviewed in parallel will conflict on the same flat JSONL file if their changes land on adjacent lines. Resolved via normal git merge-conflict handling; no special tooling planned for this at S scope.
- **Orphaned rows on doc rename/delete.** See Scope's Out-of-scope list — accepted as harmless.

## Verification

```bash
# Phase 1
scripts/mark-docs-reviewed.sh README.md
git diff scripts/docs-drift-reviewed.jsonl   # exactly one new/updated JSON line
scripts/mark-docs-reviewed.sh README.md
git diff scripts/docs-drift-reviewed.jsonl   # no further diff (idempotency)
shellcheck scripts/mark-docs-reviewed.sh scripts/check-docs-drift.sh

# Phase 2 + 3
scripts/mark-docs-reviewed.sh README.md docs/index.md docs/architecture/duplication-flow.md docs/reference/scoring-dimensions.md
./scripts/check-docs-drift.sh              # expect: no stale docs reported
./scripts/check-docs-drift.sh origin/main  # expect: unchanged gate-mode behavior

# COMMIT real implementation before the regression test below — git reset
# --hard discards ALL uncommitted tracked-file changes, not just the
# throwaway commit (a real mistake made and caught during this plan's own
# implementation).
git add scripts/check-docs-drift.sh scripts/mark-docs-reviewed.sh scripts/docs-drift-reviewed.jsonl
git commit -m "..."

# Regression test: reviewed mark must not permanently silence
echo "" >> cmd/root.go && git add cmd/root.go && git commit -q -m "TEST: throwaway" --no-verify
./scripts/check-docs-drift.sh              # expect: reviewed doc(s) re-flagged, showing "(reviewed)" baseline
git reset --hard HEAD~1                    # safe now — only discards the throwaway commit

# Repo gates
hk check && go test ./...
```
