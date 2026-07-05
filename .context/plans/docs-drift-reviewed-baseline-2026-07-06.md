---
title: "Plan: Reviewed-baseline mechanism for check-docs-drift.sh"
type: plan
status: draft
date: 2026-07-06
effort: S
related:
  - ../findings/docs-drift-perpetual-noise-2026-07-06.md
  - ../findings/gh-pages-docs-drift-2026-07-05.md
  - ../../docs/ADR/adr-044-docs-drift-pr-gate.md
  - ../../scripts/check-docs-drift.sh
---

# Plan: Reviewed-baseline mechanism for check-docs-drift.sh

## Goal

Give `scripts/check-docs-drift.sh`'s cumulative (pre-push, advisory) mode a way to record "this doc was reviewed and confirmed still accurate as of commit X" separately from "this doc's content was last edited at commit X." Today the script has only one signal — the doc file's own git history — so a doc that's genuinely fine but hasn't been re-edited keeps reappearing with a growing, meaningless commit count on every future push, forever. See `.context/findings/docs-drift-perpetual-noise-2026-07-06.md` for the concrete case (4 docs, real today) this plan fixes.

## Scope

**In scope:**

- A new tracked sidecar file recording, per doc, the date of the most recent explicit review (independent of `check-docs-drift.sh`'s existing `MAPPINGS` array, which encodes doc-to-source relationships and changes far less often).
- A small helper script that stamps a review into that sidecar for one or more doc paths.
- A change to `check-docs-drift.sh`'s cumulative mode so its per-doc "since" baseline is the later of the doc's own last-commit date and its sidecar review date, if one exists.
- Applying the new helper to the 4 currently-flagged false positives (`README.md`, `docs/index.md`, `docs/architecture/duplication-flow.md`, `docs/reference/scoring-dimensions.md`) as part of this plan's implementation, so the finding's own triage lands in the tooling, not just a PR description.

**Out of scope:**

- Any change to gate mode (the `pull_request`-only CI check added in ADR-044). It only ever diffs the current PR against its base ref, never accumulates historical debt, and needs nothing from this plan.
- Any change to `MAPPINGS` itself — this plan doesn't add, remove, or re-scope which source globs map to which docs.
- Automating the review decision (e.g. an LLM judging "is this doc still accurate"). The review itself stays a human (or agent-assisted, human-confirmed) judgment call; this plan only makes recording that judgment durable.

## Phases

### Phase 1 — Sidecar file and helper script

1. Add `scripts/docs-drift-reviewed.tsv` (empty initially, or seeded — see Open Questions) with one line per reviewed doc: `doc_path<TAB>reviewed_date_ISO8601`.
2. Add `scripts/mark-docs-reviewed.sh <doc_path> [doc_path...]`: for each argument, look up `HEAD`'s commit date (`git log -1 --format=%cI`) and insert or replace that doc's line in the sidecar. Idempotent — running it twice on the same doc at the same `HEAD` produces no diff.
3. `shellcheck scripts/mark-docs-reviewed.sh` clean; `set -euo pipefail`; pure bash (no Python/Node, per `.agents/RULES.md` rule 13), matching `check-docs-drift.sh`'s existing style.

Exit criterion: running `scripts/mark-docs-reviewed.sh README.md` on a clean checkout adds exactly one correctly-formatted line to the sidecar.

### Phase 2 — Wire the baseline into cumulative mode

1. In `check-docs-drift.sh`'s default (no-arg) mode, before the `--after` check for each mapping entry: look up the doc's sidecar entry, if any, and compute `effective_since = max(doc_last_touch_date, reviewed_date)`.
2. Use `effective_since` in place of the current `doc_date` for the `git log --after` comparison. Gate mode (the `BASE_REF` branch) is untouched — it doesn't use `doc_date` at all today, and this plan doesn't change that.
3. Update the script's header comment to document the sidecar's role alongside the existing gate-mode/cumulative-mode comment block added in ADR-044.

Exit criterion: a doc with a sidecar entry newer than its most recent flagged source commit no longer appears in `check-docs-drift.sh`'s output; a doc with no sidecar entry behaves exactly as it does today (no regression for the 4 docs already fixed in `.context/findings/gh-pages-docs-drift-2026-07-05.md`'s follow-up work).

### Phase 3 — Apply to the 4 known false positives

1. Run `scripts/mark-docs-reviewed.sh README.md docs/index.md docs/architecture/duplication-flow.md docs/reference/scoring-dimensions.md`.
2. Run `./scripts/check-docs-drift.sh` with no args and confirm it reports zero stale docs.
3. Run `./scripts/check-docs-drift.sh origin/main` (gate mode) and confirm it still reports no new drift — proving Phase 2 didn't touch gate-mode behavior.

Exit criterion: cumulative mode reports 0 flagged docs; gate mode behavior is bit-for-bit unchanged from before this plan.

## Open Questions

- **Sidecar key: commit date or SHA?** The plan above uses an ISO8601 date (matching the existing `--after`-based comparison style already used elsewhere in the script), but a SHA would be more precise (immune to clock skew / rebases changing author dates) at the cost of an extra `git log` lookup per check. Leaning toward date for consistency with the rest of the script, but this is a real trade-off worth a second opinion.
- **Should `mark-docs-reviewed.sh` require the doc to be currently flagged before allowing a mark?** Guarding against marking an already-current doc (a no-op that adds sidecar noise) vs. keeping the script simple and trusting the human/agent invoking it. Leaning toward no guard (simplicity), but flagging for review.
- **Does the sidecar itself need a docs-drift mapping entry (i.e., should changes to `scripts/mark-docs-reviewed.sh` or the sidecar format ever prompt a doc update)?** Likely not — this is tooling-internal, not user-facing — but worth confirming it doesn't create a circular dependency on `check-docs-drift.sh`'s own `MAPPINGS`.

## Verification

```bash
# Phase 1
scripts/mark-docs-reviewed.sh README.md
git diff scripts/docs-drift-reviewed.tsv   # should show exactly one new/updated line
shellcheck scripts/mark-docs-reviewed.sh

# Phase 2 + 3
scripts/mark-docs-reviewed.sh README.md docs/index.md docs/architecture/duplication-flow.md docs/reference/scoring-dimensions.md
./scripts/check-docs-drift.sh              # expect: no stale docs reported
./scripts/check-docs-drift.sh origin/main  # expect: unchanged gate-mode behavior
```
