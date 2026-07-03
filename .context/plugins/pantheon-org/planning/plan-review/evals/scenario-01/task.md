# Scenario 01: Basic Plan Review

## User Prompt

"Please review the plan at .context/plans/evoskill-core-loop-port-2026-07-02.md"

## Expected Behavior

1. Identify the plan path and read the full file content.
2. Compose a self-contained plan brief (title, status, date, goal, steps, open questions, risks, context).
3. Check `opencode.json` for existing `subAgents` config.
4. Since no model routing is configured, ask the user which environment they are on and present model options.
5. After the user selects models, spawn 3 subagent reviewers in parallel: Technical (`general`), Strategic (`general`), Risk (`explore`), all receiving the same plan brief.
6. Collate the 3 reviews into a Consolidated Review Report following the structure in `assets/templates/review-report.yaml`.
7. Present the report to the user with model attribution.
8. Offer to investigate any flagged items.

## Success Criteria

- Agent reads the plan file before composing the brief.
- Agent checks `opencode.json` for `subAgents` before asking the user.
- Model selection is presented as structured options for the user to choose.
- All 3 reviewers are spawned in parallel (single message, 3 task calls).
- The final report includes all 8 required sections: Model Configuration, Structural Validation, Implementation Architecture, Scores, Critical Issues, Moderate Concerns, Strengths, Next Actions.
- Report is attributed per-reviewer, not presented as a single block.

## Failure Conditions

- Agent skips reading the plan and tries to review from memory.
- Agent spawns reviewers sequentially instead of in parallel.
- Agent does not ask about models and uses defaults.
- Agent presents raw subagent output without collation or attribution.
- Agent does not offer follow-up investigation.
