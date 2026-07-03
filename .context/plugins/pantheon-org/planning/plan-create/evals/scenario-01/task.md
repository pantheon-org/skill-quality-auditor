# Scenario 01: Basic Plan Creation

## User Prompt

"I need a plan to add a --format json flag to the evaluate command. Can you draft it?"

## Expected Behavior

1. Ask the user about the goal, scope, phases, tasks, dependencies, and risks.
2. Scan existing plans with `grep -r '^## ' .context/plans/*.md` to infer local conventions.
3. Draft the plan file with YAML frontmatter (title, type: plan, status: draft, date).
4. Structure the body with `## Goal`, `## Phases` (with tasks and wave annotations), and `## Open Questions`.
5. Run `validate-context-frontmatter.sh` on the created file.
6. Offer to run the `plan-review` skill on the new plan.

## Success Criteria

- Agent asks about goal, scope, and phases before drafting.
- Agent scans existing plans to infer conventions.
- Created plan has valid YAML frontmatter with all required fields.
- Plan body has `## Goal` and `## Phases` sections.
- Phases contain concrete tasks with wave annotations where parallelisable.
- Agent runs validation script on the created file.
- Agent offers to run plan-review.

## Failure Conditions

- Agent creates the plan without asking about phases or scope.
- Agent does not scan existing plans for conventions.
- Plan file has no frontmatter or invalid frontmatter.
- Plan body has no `## Phases` section.
- Agent does not run validation.
- Agent does not offer plan-review.