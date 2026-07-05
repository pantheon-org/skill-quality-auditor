---
title: "ADR-045: check-docs-drift.sh gains a reviewed-baseline sidecar"
status: accepted
date: 2026-07-06
context:
  - path: ".context/findings/docs-drift-perpetual-noise-2026-07-06.md"
  - path: ".context/plans/docs-drift-reviewed-baseline-2026-07-06.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

`.context/findings/docs-drift-perpetual-noise-2026-07-06.md` identified that `scripts/check-docs-drift.sh`'s cumulative (pre-push, advisory) mode has only one signal: a doc's own last-commit date. It cannot distinguish "reviewed and confirmed still accurate" from "never looked at" — so a doc that's genuinely fine but hasn't been re-edited keeps reappearing, with a growing and increasingly meaningless commit count, on every future push forever. This is a real, live problem: 4 docs (`README.md`, `docs/index.md`, `docs/architecture/duplication-flow.md`, `docs/reference/scoring-dimensions.md`) were reviewed and confirmed accurate while closing out `.context/findings/gh-pages-docs-drift-2026-07-05.md`'s follow-up work, and will keep re-flagging under the prior behavior. Left unaddressed, this risks alert fatigue — a team that learns to skim past known-false-positive warnings stops reading new ones too.

`.context/plans/docs-drift-reviewed-baseline-2026-07-06.md` was drafted to fix this and went through a 3-reviewer plan-review (Technical, Strategic, Risk — all Claude Sonnet 5). The review surfaced two decisions worth recording here rather than leaving as implementation detail: how the "reviewed" baseline is actually compared against source-commit dates, and what trust model governs marking a doc reviewed.

## Decision

1. **A new tracked sidecar, `scripts/docs-drift-reviewed.tsv`, records per-doc review baselines** — `doc_path<TAB>reviewed_date_iso<TAB>reviewed_epoch` — separate from `check-docs-drift.sh`'s existing `MAPPINGS` array (which encodes doc-to-source-glob relationships and changes far less often than review timestamps do). A new helper, `scripts/mark-docs-reviewed.sh <doc_path>...`, writes to it.
2. **The cumulative check's effective baseline per doc is `max(doc_last_touch_epoch, reviewed_epoch)`, computed as integer comparison — never ISO8601 string comparison.** The plan-review's Technical and Risk passes both independently flagged that lexical comparison of ISO8601 timestamps is timezone-offset-fragile (silently wrong if a commit's offset differs from the sidecar entry's). The sidecar stores a pre-computed epoch (from `git log -1 --format=%ct` at write time) specifically so no date-string parsing is ever needed at check time. Live-verified: `git log --after="@$epoch"` produces identical results to the existing `--after="$iso_date"` form, so the fix required no new date-parsing dependency (avoiding the GNU-vs-BSD `date` portability gap between this script's CI runner and contributors' local macOS machines).
3. **`mark-docs-reviewed.sh` does not require a doc to be currently flagged before allowing a mark, and does not gate on any authorization check.** It is a bookkeeping tool for an advisory, non-blocking check, matching the trust model of `check-docs-drift.sh` itself (which never fails a push). Marking prints what commits are being confirmed-reviewed for visibility/audit-trail purposes, but this is not an access control — it can't be, without contradicting the check's own advisory nature. The plan-review's Risk pass raised "an agent could self-mark everything to silence the check"; accepted as an inherent property of an advisory tool rather than something to lock down, since a hard block would also break the legitimate case of proactively confirming a doc reviewed in the same session a source change lands.
4. **Gate mode (the `pull_request`-only CI check from ADR-044) is untouched.** It diffs the current PR against its base ref and never accumulates historical debt, so it was never subject to the problem this ADR fixes.
5. **No JSON-Schema-validated template for the sidecar.** This repo's template+schema+validation-script convention governs skill artifacts under `.context/plugins/`; a 2-column TSV of internal tooling state isn't that kind of artifact, and a schema for it would be disproportionate to the fix's scope.

## Consequences

- **Easier:** a doc reviewed and confirmed accurate stops reappearing in pre-push output, without requiring a hollow edit to its prose just to reset a timestamp.
- **Easier:** the epoch-based comparison eliminates an entire class of latent timezone bugs before they could ship — verified live rather than assumed safe.
- **Harder:** the sidecar can drift from reality if a reviewer marks a doc reviewed against a stale local checkout (behind `origin/main`) — mitigated only by usage guidance, not enforced, consistent with this being advisory tooling rather than a security control.
- **Binding for future work:** any change to `check-docs-drift.sh`'s cumulative-mode comparison logic must preserve the epoch-based (not string-based) comparison — reverting to ISO8601 string comparison would reintroduce the exact bug this ADR fixes.
