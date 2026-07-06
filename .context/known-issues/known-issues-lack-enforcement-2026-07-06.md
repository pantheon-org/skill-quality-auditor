---
title: "Known Issue: known-issues has no enforcement, forcing function, or expiry"
type: known-issue
status: active
date: 2026-07-06
severity: high
related:
  - ../../docs/ADR/adr-046-known-issues-context-type.md
  - ../../.context/plugins/pantheon-org/planning/design-debate/SKILL.md
---
# Known Issue: known-issues has no enforcement, forcing function, or expiry

> Originally surfaced as the sharpest finding of a session-reflection sub-agent: the `known-issues` mechanism (ADR-046) has `severity`/`status` fields and sorts to the top of `.context/index.yaml` — that's it. No CI gate fails on a `severity: critical` entry sitting `active`. Run through `design-debate` (Advocate / Skeptic / Migration-Risk) to decide whether to add enforcement now. **Verdict: `do_not_proceed_for_now`.**

## Why this exists

The Advocate proposed reusing `scripts/check-plan-drift.sh`'s proven 60-day age-threshold pattern, wired into `ci.yml` (unlike `check-plan-drift.sh` itself, which has the identical "advisory, pre-push-only, never CI" gap). Three findings against proceeding now, none rebutted:

1. **Severity is self-assigned with no calibration track record** (Skeptic) — the only two real examples so far (the jq bug rated `critical`, the CI-visibility gap rated `high`) are arguably backwards in actual impact, and a gate tied to `severity: critical` creates a rating-gaming incentive that doesn't exist while the field stays advisory.
2. **Age-since-creation can't distinguish "actively being fixed" from "neglected"** (Migration-Risk) — `context-file`'s own convention is that `date` is creation date, never updated on edits, so a critical issue worked on for 10 days looks identical to one abandoned for 10 days. This is a design flaw in the proposed mechanism, not an implementation detail.
3. **A date-threshold CI gate is non-deterministic** (Migration-Risk) — a PR that passed yesterday could fail today with zero new commits, purely from wall-clock aging. Every other job in `ci.yml` is diff-triggered; this would be a real departure from that invariant, and the reused `check-plan-drift.sh` pattern doesn't even run cross-platform as-is (`date -j` is macOS-only; GitHub's `ubuntu-latest` runners need a GNU `date -d` branch that doesn't exist yet).
4. Also: the gate would be **trivially, undetectably silenceable** by bumping the `date` field forward — nothing validates date-immutability beyond the documented convention.
5. **No usage evidence yet** — the mechanism is hours old, three entries total, zero currently at any proposed trigger condition. The one precedent for "advisory failed" (`check-docs-drift.sh`'s cumulative mode) was a different mechanism, fixed the same day it was found — not evidence of multi-week neglect.

## Impact if unfixed

The mechanism built specifically to avoid "a list nobody reads" risks becoming exactly that over time: visible, not enforced, the same pattern already shown insufficient earlier in the same session for `check-docs-drift.sh` before that got a CI-visibility fix (which itself still isn't a hard gate — see `.context/known-issues/docs-drift-cumulative-not-enforced-2026-07-06.md`, the sibling gap this one echoes for a different mechanism).

## Revisit trigger

Revisit if either holds: (a) a genuinely critical known-issue is observed sitting neglected under real usage — impossible to have happened yet, since none has existed long enough — or (b) the design is reworked around a **last-reviewed timestamp instead of creation date**, mirroring the exact fix this same session already built for docs-drift (ADR-045's reviewed-baseline pattern: `max(created_epoch, reviewed_epoch)` instead of a single immutable date) — which would resolve the "active work vs. neglect" ambiguity Migration-Risk flagged, the same way it already resolved the analogous problem for docs.
