# Scenario 02: Pre-Configured Model Routing

## User Prompt

"Review the plan at .context/plans/yaml-content-validation-config-2026-07-03.md"

## Setup

The user's `opencode.json` already contains:

```jsonc
{
  "subAgents": {
    "general": { "model": "deepseek-v4-flash" },
    "explore": { "model": "minimax-m3" }
  }
}
```

## Expected Behavior

1. Read the plan file and compose a plan brief.
2. Check `opencode.json` and detect that `subAgents` is configured with two distinct models.
3. Note the existing config and skip the model selection question entirely.
4. Proceed directly to spawning 3 reviewers.
5. Include the model assignments in the final report.

## Success Criteria

- Agent reads `opencode.json` and detects the pre-configured routing.
- Agent does NOT ask the user about model preferences.
- The final report includes the model assignments from `opencode.json`.
- All 3 reviewers are spawned in parallel.

## Failure Conditions

- Agent asks the user about models despite existing configuration.
- Agent ignores the `subAgents` config and uses defaults.
- Agent does not include model assignments in the report.
