# Scenario 02: Convention Inference

## User Prompt

"Draft a plan for migrating the test suite from testify to go-test. Make sure it follows our conventions."

## Expected Behavior

1. Scan existing plans: `grep -r '^## ' .context/plans/*.md | sed 's/.*## //' | sort | uniq -c | sort -rn`
2. Identify core sections (>= 40% frequency), common sections (20-39%), and rare sections (< 20%).
3. Check naming convention: timestamped (`topic-YYYY-MM-DD.md`) vs topic-only.
4. Draft the plan matching the inferred conventions.
5. Verify the draft uses the same section headings as the majority of existing plans.

## Success Criteria

- Agent runs the grep command to scan existing plan headings.
- Agent identifies which sections are core, common, and rare.
- Agent matches the naming convention of existing plans.
- The created plan uses sections that match the local convention, not a hardcoded template.

## Failure Conditions

- Agent hardcodes "Goal → Steps → Open Questions" without scanning first.
- Agent does not check naming conventions.
- Created plan uses sections not found in any existing plan.
- Agent ignores the frequency data and includes rare sections unnecessarily.