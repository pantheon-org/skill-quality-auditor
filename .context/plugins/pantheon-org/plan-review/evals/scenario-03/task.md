# Scenario 03: Plan with Structural Issues

## User Prompt

"Please audit the plan at .context/plans/onboarding-improvements-2026-04-27.md for structural issues"

## Expected Behavior

1. Read the plan file.
2. Run the frontmatter validation script: context-index/scripts/validate-context-frontmatter.sh
3. Scan H2 headings across all other plans to infer the local convention.
4. Check the plan's implementation architecture (phases, tasks, waves).
5. Include all structural findings in the plan brief so the 3 reviewers can reference them.
6. Spawn the 3 reviewers with the structural issues included in the brief.

## Success Criteria

- Agent runs the validation script on the plan file.
- Agent scans H2 headings across existing plans to infer the convention.
- Agent identifies any missing core sections (Goal, Steps, Open Questions).
- Agent identifies any frontmatter violations.
- Structural issues are included in the plan brief shared with all 3 reviewers.
- The final report's Structural Validation section documents all findings.

## Failure Conditions

- Agent skips the validation script and checks frontmatter manually.
- Agent hardcodes a template instead of inferring from existing plans.
- Agent does not include structural issues in the plan brief.
- Agent ignores the implementation architecture entirely.
