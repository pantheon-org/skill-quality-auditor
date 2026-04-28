# Scenario 03: Skill Improvement Planning

## User Prompt

"Create a remediation plan for ci-cd/github-actions-generator to bring it to A-grade."

## Expected Behavior

1. Acknowledge the current score (94/140, C+) and weakest dimensions: D3: 6/15, D5: 7/15, D9: 4/20.
2. Produce `remediation-plan.md` opening with an executive summary: current score, target score
   (>=126/140), current grade (C+), target grade (A), priority, and top focus areas.
3. Include a critical issues table identifying D3, D5, and D9 by dimension number with severity
   and estimated impact.
4. Organise improvement steps into phases (e.g., Phase 1: Anti-Patterns, Phase 2: Progressive
   Disclosure) with a per-phase score target.
5. Name exact files to create or modify (e.g., `SKILL.md`, `references/anti-patterns.md`) and include content examples.
6. Produce `success-criteria.md` defining per-dimension score targets (e.g., D3 >=12/15, D9 >=16/20).
7. Produce `implementation-steps.sh` with `skill-auditor evaluate` commands to verify each phase's score impact.
8. Ensure the projected score delta, if all phases are completed, reaches >=126/140.

## Success Criteria

- Plan opens with current score, target score, current/target grade, priority, and top focus areas.
- Critical issues table references D3, D5, D9 by dimension number with severity and impact.
- Steps grouped into named phases with per-phase targets.
- Exact files named with content examples (not just file-type suggestions).
- `success-criteria.md` defines per-dimension numerical targets.
- Each phase has an S/M/L effort estimate and approximate time in hours.
- `implementation-steps.sh` includes `skill-auditor evaluate` commands to verify each phase.
- Projected score delta would bring the skill to >=126/140 if all phases completed.

## Failure Conditions

- Remediation plan is a flat list of suggestions without phase structure.
- No effort sizing (S/M/L) on any step.
- Exact files not named; only general advice given.
- `implementation-steps.sh` missing or contains no verification commands.
- `success-criteria.md` absent or contains vague targets.
- Projected outcome does not reach A-grade threshold.

**Context:**

Audit results for three skills:

- `ci-cd/github-actions-generator`: 94/140 (C+) — D3: 6/15, D5: 7/15, D9: 4/20
- `documentation/markdown-authoring`: 101/140 (C) — D1: 12/20, D3: 5/15, D7: 6/10
- `testing/bdd-testing`: 88/140 (C) — D2: 8/15, D3: 4/15, D5: 6/15

All three are blocked from publishing (threshold: >=126/140 for A-grade).
