# Model Routing for Plan Review

Achieving 3 different models for the 3 reviewers depends on your opencode
configuration. This reference documents the available configuration options.

## opencode.json subAgents config

OpenCode supports routing different `subagent_type` values to different models
via `subAgents` in `opencode.json`:

```jsonc
{
  "subAgents": {
    "general": {
      "model": "claude-sonnet-5",
      "systemPrompt": "You are a thorough, detail-oriented reviewer."
    },
    "explore": {
      "model": "deepseek-v4-flash",
      "systemPrompt": "You are a skeptical, adversarial reviewer focused on finding flaws."
    }
  }
}
```

## Recommended model pairings

| Reviewer | subagent_type | Recommended model | Rationale |
|----------|--------------|-------------------|-----------|
| Technical | `general` | Claude Sonnet 5 or GPT 5.1 Codex Mini | Strong reasoning for implementation feasibility |
| Strategic | `general` | Claude Sonnet 5 or GPT 5.1 Codex Mini | Both Technical and Strategic need reasoning depth — use the same strong model |
| Risk | `explore` | DeepSeek V4 Flash or GPT 5 Nano | A different model brings different blind spots; cheaper models are fine for adversarial review |

## When only 2 subagent types are available

OpenCode provides `general` and `explore` subagent types (plus `atlassian`
which is manual-only). If you can only configure 2 distinct models, use:

- **Strongest model** → `general` (covers both Technical and Strategic — run them
  as separate `general` calls on the same model; the different prompts still
  produce divergent analysis)
- **Second model** → `explore` (Risk reviewer gets a fresh perspective)

## When all subagents share one model

The reviewer prompts and question sets still provide useful diversity:
- Technical asks about feasibility, gaps, consistency
- Strategic asks about alignment, scope, priority
- Risk asks about blind spots, failure modes, edge cases

Even on the same model, these produce meaningfully different output. The
consolidation catches contradictions between them.
