# Eval runner

The `eval` command runs structured evaluation scenarios against an LLM provider
to test whether an AI agent can correctly use the skill to produce task output.

## Overview

The eval runner has two modes:

1. **Structural gate** (no API key) — validates scenario structure, deterministic
2. **LLM judge** (with API key) — runs actor + judge, multi-sample, advisory only

Per ADR-007 and ADR-018, the structural gate is Tier 1 (required CI gate).
The LLM-judge step is Tier 2 (advisory, never blocks merge).

## Pipeline

```text
eval <evalsDir> [--samples N] [--cost-log FILE] [--json]
  │
  ├── loadScenarios(evalsDir)
  │     └── discover scenario-N/ directories
  │     └── each requires:
  │           ├── task.md           — the task prompt
  │           ├── criteria.json     — weighted checklist (must sum to 100)
  │           └── capability.txt    — capability description
  │
  ├── newRunner(skillPath, skillContent, envConfig)
  │     └── llmclient.NewFromEnv(provider) → Client or nil
  │     └── nil client → structural-only mode
  │
  └── run(ctx, scenarios)
        │
        ├── create bounded worker pool (maxConcurrent = 3)
        │
        ├── for each scenario (goroutine):
        │     ├── structural mode (no client):
        │     │     └── runStructural(s)
        │     │           ├── criteria.json sums to 100?
        │     │           ├── task.md exists?
        │     │           └── capability.txt exists?
        │     │
        │     └── LLM mode (has client):
        │           └── runLLM(ctx, s, judgeTemp)
        │                 ├── 1. actor: ActorMessages(skill, task)
        │                 │           └── client.Chat() → actor output
        │                 │
        │                 ├── 2. judge: JudgeMessages(skill, task, output, criteria)
        │                 │           └── client.Chat() × N samples
        │                 │           └── ParseJudgeResponse → criteria scores
        │                 │
        │                 ├── 3. per-item median scores across samples
        │                 ├── 4. aggregate total → PASS / MARGIN / FAIL
        │                 └── 5. log tokens if --cost-log
        │
        ├── sort scenarios by ID
        ├── append structural diagnostics as synthetic scenario if needed
        └── determine overall pass (all scenarios pass)
```

## LLM client architecture

The LLM client (`internal/llmclient/`) is a provider-agnostic abstraction:

```text
Client interface
  ├── Chat(ctx, Request) → (*Response, error)
  └── Config() → Config

Providers (registered in client.go):
  ├── anthropic → internal/llmclient/anthropic.go
  ├── openai    → internal/llmclient/openai.go
  ├── gemini    → internal/llmclient/gemini.go
  ├── mistral   → internal/llmclient/mistral.go (OpenAI-wire-compatible, reuses OpenAIClient)
  ├── cerebras  → internal/llmclient/cerebras.go (OpenAI-wire-compatible, reuses OpenAIClient)
  └── openai-compatible → (same client, different base URL)
```

### Environment-based configuration

```go
NewFromEnv(providerOverride) → (Client, error)
```

1. Selects provider: `--provider` flag > `LLM_PROVIDER` env > "anthropic" default
2. Reads provider-specific env vars for API key, base URL, model
3. Returns `(nil, nil)` when no key configured → triggers structural-only mode

Retry policy: `IsRetryable(429, 5xx)`. Transient 5xx errors use `Backoff()`
(500ms base, 8s max, random jitter). 429 rate-limit responses use
`RateLimitBackoff()` instead, which honors the provider's `Retry-After` header
when present, falling back to a 15s-base/90s-capped schedule otherwise — sized
for per-minute quota windows rather than transient failures. `MaxRetryAttempts`
is 4 (across all adapters), giving the longer 429 backoff room to matter.

### Prompt architecture

`internal/llmclient/prompt.go`:

- **`ActorMessages(skillContent, taskPrompt)`** — builds system + user message pair
  that frames the model as an agent with access to the skill
- **`JudgeMessages(skillContent, taskPrompt, actorOutput, criteriaJSON)`** — builds
  the judge message with skill context, task, agent output, and criteria checklist
- **`ParseJudgeResponse(raw)`** — extracts JSON from judge output (handles extra prose)
- **`PromptVersion()`** — SHA256 of `JudgePrompt` for audit trail

## Scenario structure

Each eval scenario lives in a numbered directory:

```text
evals/
  scenario-1/
    task.md           # The task prompt for the actor
    criteria.json     # Weighted checklist:
                      # { "criteria": [{"item": "...", "weight": 20}, ...] }
                      # weights must sum to 100
    capability.txt    # Capability being tested
```

## Source files

| File | Purpose |
|------|---------|
| `cmd/eval.go` | Command entry, runner, scenario loading |
| `internal/llmclient/client.go` | Client factory, retry, providers registry |
| `internal/llmclient/types.go` | Message, Request, Response, Config types |
| `internal/llmclient/prompt.go` | JudgePrompt, ActorMessages, JudgeMessages |
| `internal/llmclient/anthropic.go` | Anthropic provider implementation |
| `internal/llmclient/openai.go` | OpenAI (+ compatible) provider implementation |
| `internal/llmclient/gemini.go` | Gemini provider implementation |
| `internal/llmclient/mistral.go` | Mistral provider implementation (OpenAI-wire-compatible) |
| `internal/llmclient/cerebras.go` | Cerebras provider implementation (OpenAI-wire-compatible) |
