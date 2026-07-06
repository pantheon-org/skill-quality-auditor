---
title: "ADR-052: governance known-issue debt is triaged by value, not batch-fixed"
status: accepted
date: 2026-07-06
context:
  - path: ".context/plans/governance-tooling-hardening-2026-07-06.md"
  - path: ".context/plans/governance-enum-and-schema-hardening-2026-07-06.md"
  - path: ".context/known-issues/known-issues-lack-enforcement-2026-07-06.md"
  - path: "docs/ADR/adr-046-known-issues-context-type.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

Six `GOVERNANCE`-themed known-issues had accumulated about the integrity of the
`.context/` and ADR tooling: the adr-index gate checked existence not freshness;
`check-undocumented-decisions.sh` silently skipped any file mentioning
`index.yaml`; generated remediation plans omitted the now-required `value` field;
the frontmatter schema's `additionalProperties: false` was not enforced by a real
validator; `effort` used single-letter values against the UPPER_CASE-word
convention (ADR-050); and known-issues themselves had no enforcement gate. A
verification pass confirmed all six were still live against current source. A
3-reviewer `plan-review` then examined a single consolidated remediation plan.

## Decision

1. **Fix by value, not in one batch.** The six issues are split by their own
   `value` grades rather than bundled. The three quick, genuinely-valuable gaps
   (adr-index freshness, undocumented-decisions false-negative, remediation-plan
   `value` — all individually `value: MEDIUM`/`S`-effort) go in
   `governance-tooling-hardening-2026-07-06.md`. The two `value: LOW`,
   high-blast-radius gaps (effort full-word enum migration; a real JSON-schema
   validator) go in a separate deferred plan,
   `governance-enum-and-schema-hardening-2026-07-06.md`.

2. **Do not gate valuable work behind low-value work.** The consolidated plan
   originally sequenced the LOW-value effort-enum migration as a hard prerequisite
   to the valuable remediation-plan `value` fix. Two independent reviewers
   (Strategic, Risk) flagged that a blended `value: MEDIUM` hid a LOW majority-
   effort tail and that the coupling blocked a live fix. The split removes the
   coupling and keeps each plan sorting honestly under the read protocol.

3. **Known-issues enforcement stays rejected (reaffirms the deferral).** The
   `known-issues-lack-enforcement` gap is a settled `do_not_proceed_for_now`
   design-debate verdict — self-assigned severity, creation-date cannot detect
   neglect, a wall-clock expiry gate is non-deterministic, and any such gate is
   trivially silenceable. It is not actioned by either plan and stays `ACTIVE` as
   a tracked deferral. This complements ADR-046 (which established the known-issue
   context type) by recording that the type deliberately carries no enforcement.

4. **Generated remediation plans get a static `value: MEDIUM`, not a derived
   grade.** Deriving `value` from the score gap would invent a second heuristic
   for no proven benefit; a static default is the minimal fix that makes generated
   plans satisfy the validator. Revisit only if the grades prove systematically
   wrong.

5. **The deferred plan sequences the enum migration before the schema validator,
   and does not enforce `additionalProperties: false` against generated
   remediation plans until they have a dedicated schema or a path-glob exemption.**
   The effort enum must be final before a strict validator locks it in, and a
   strict validator would otherwise reject generated remediation plans on their
   extra keys.

## Consequences

- **Easier:** the read protocol stays honest — the MEDIUM quick-wins plan is not
  weighed down by a LOW deferred tail, and the valuable remediation-plan fix ships
  without waiting on an optional migration.
- **Prevents re-litigation:** a future agent will not re-attempt known-issues
  enforcement (Decision 3), re-bundle the enum/schema work into the quick-wins
  plan (Decisions 1-2), or add a `gapValue()` heuristic without cause (Decision 4).
- **Ongoing:** the deferred plan carries genuinely open design questions (effort
  vocabulary, schema-validator runner, remediation-plan schema scope) to be
  resolved in its own `plan-review` when it reaches the top of the queue.
- **Observed irony worth noting:** the quick-wins plan is currently exempted from
  the `check-undocumented-decisions` scan by the very `index.yaml`-substring
  false-negative (G3) it exists to fix — this ADR gives it real coverage
  regardless, and G3's fix will bring it under the scan legitimately.
