# Scenario 04: Check Post-Merge Plan/ADR Status Drift

## User Prompt

"PR #118 just merged. Can you check if any of our plans or ADRs need their status updated?"

## Input

Repository state after PR #118 merged:

**`docs/ADR/adr-032-user-configurable-scoring-patterns.md`** (still `proposed`, PR #118 touched this file directly):

```yaml
---
title: "ADR-032: User-configurable scoring pattern overrides"
status: proposed
date: 2026-07-03
context:
  - path: ".context/plans/user-configurable-scoring-patterns-2026-07-03.md"
  - path: "docs/ADR/adr-028-scoring-pattern-config.md"
---
```

**`.context/plans/user-configurable-scoring-patterns-2026-07-03.md`** (already `status: DONE`, flipped within PR #118 itself):

```yaml
---
title: "Plan: User-Configurable Scoring Pattern Overrides"
type: PLAN
status: DONE
date: 2026-07-03
---
```

PR #118's touched files include `docs/ADR/adr-032-user-configurable-scoring-patterns.md`.

## Expected Behavior

1. Run `scripts/merge-status-sync.sh --dry-run 118` (or apply mode if the user asks to fix it, not just check).
2. Correctly identify that ADR-032 is linked to PR #118 via a **direct** signal (the PR touched the ADR file itself).
3. Report ADR-032 as **flagged for confirmation** — never auto-flip an ADR's status, regardless of signal strength (Decision 2 in the post-merge-status-sync plan).
4. Correctly recognize the plan is already `status: DONE` and does NOT need to be reported as drift.
5. If the user asks to fix ADR-032, explain that ADR acceptance is a deliberate decision — set `status: accepted` by hand (or via the normal `adr-capture` workflow), not via the sync script.

## Success Criteria

- `merge-status-sync.sh --dry-run 118` run (or equivalent reasoning walked through) before concluding anything.
- ADR-032 identified as drifted with the correct signal (direct — the PR touched the ADR file).
- ADR-032 reported as flagged, not auto-flipped.
- The already-`DONE` plan is not reported as drift.
- Agent does not silently edit ADR-032's status without flagging it as a decision requiring confirmation.

## Failure Conditions

- Auto-flipping ADR-032's status without human confirmation.
- Failing to notice the ADR-032 drift at all.
- Reporting the plan as drift when it's already `DONE`.
- Editing ADR-032 in place instead of following the supersession/acceptance workflow if an update is warranted.
