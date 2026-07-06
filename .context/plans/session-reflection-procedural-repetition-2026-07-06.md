---
title: "Plan: add a procedural-repetition check to session-reflection"
type: plan
status: draft
date: 2026-07-06
value: medium
effort: S
related:
  - ../findings/session-reflection-procedural-repetition-blind-spot-2026-07-06.md
  - ../plugins/pantheon-org/workshop/session-reflection/SKILL.md
---

# Plan: add a procedural-repetition check to session-reflection

## Goal

Close the gap in `.context/findings/session-reflection-procedural-repetition-blind-spot-2026-07-06.md`:
`session-reflection`'s confidence audit and blind-spot check both orient outward (content
correctness, shared understanding with the user) and never inward, at the agent's own
repeated manual procedures. Add a check that does.

## Scope

**In scope:**

- Amend `.context/plugins/pantheon-org/workshop/session-reflection/SKILL.md` so the
  reflection scans the session summary for a manual, multi-step sequence the agent
  repeated 2+ times, and — if found — surfaces it explicitly rather than letting it
  pass silently the way the validate-and-merge pattern did for eight PRs this session.
- Decide (Open Questions, for `plan-review`) whether this is a **third question**
  alongside confidence-audit and blind-spot, or a **fold-in** to one of the existing
  two — most likely the blind-spot check, since "what am I missing" already has the
  right shape, just the wrong direction of attention.
- Update the sub-agent-spawn prompt template (the skill already offloads reflection to
  an `explore` sub-agent for deep sessions) so the delegated prompt also asks for
  repeated procedures, not just confidence items and blind spots.
- Update the skill's eval scenarios (if any exist) or add one demonstrating the new
  check catching a repeated pattern.

**Out of scope:**

- Building any automated repetition *detector* (e.g. parsing tool-call logs for
  duplicate sequences). This plan is about adding the right question, not building
  tooling to answer it — matching how the existing two questions are answered by the
  agent's own judgment, not a script.
- The actual `pr-merge` skill this pattern would produce — that's
  `.context/plans/pr-merge-skill-2026-07-06.md`, a separate, already-drafted plan.

## Phases

### Phase 1 — Draft the amendment

1. Decide third-question vs. fold-in (Open Questions).
2. Update the Workflow section: add the new question (or amend the blind-spot
   question's guidance) with the same rigor as the existing two — a "why this works"
   rationale, example items, and guidance on precision (specific procedure name +
   repeat count, not "I might be repeating myself").
3. Update the Advanced: Sub-agent spawn pattern section's delegated prompt template to
   include the new question.
4. Update Anti-Patterns: add a "never treat a repeated success as evidence there's
   nothing to formalize" entry, mirroring the actual failure mode from this session.

Exit criterion: `SKILL.md` reads coherently end to end; the new question doesn't
duplicate the wording or scope of the other two.

### Phase 2 — Validate

1. `./dist/skill-auditor evaluate .context/plugins/pantheon-org/workshop/session-reflection --store`
   — confirm no score regression, matching the precedent set for `design-debate` and
   `plan-review`'s amendments this session (both landed with identical before/after
   scores).
2. `./dist/skill-auditor validate artifacts .context/plugins/pantheon-org/workshop/session-reflection`.

Exit criterion: grade does not regress; artifact validation passes.

### Phase 3 — Land

1. Update `docs/development/skills-and-rules.md`'s `session-reflection` one-liner if
   the new check changes what the skill does at a glance (matching the fix already
   applied for `plan-review` in PR #198, and for `design-debate` in PR #196).
2. Open a PR; confirm `scripts/check-docs-drift.sh origin/main` shows no new drift.
3. Merge.

Exit criterion: PR merged, no docs-drift gate failure.

## Open Questions

- **Third question, or fold into the blind-spot check?** A third question adds
  reflection overhead to every session-end; folding in keeps the two-question shape
  but risks the new angle getting lost under the existing blind-spot framing. Leaning
  fold-in (the blind-spot question is already "what am I missing" — repeated
  procedures are a specific instance of that), but not decided here.
- **Should the new check name a threshold** (e.g. "repeated 2+ times") or leave it to
  judgment like the other two questions do? A hard number is easier to act on
  consistently; a soft judgment call matches the skill's existing style.

## Verification

```bash
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/session-reflection-procedural-repetition-2026-07-06.md
./dist/skill-auditor evaluate .context/plugins/pantheon-org/workshop/session-reflection --store
./dist/skill-auditor validate artifacts .context/plugins/pantheon-org/workshop/session-reflection
scripts/check-docs-drift.sh origin/main
```
