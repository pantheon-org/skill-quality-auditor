---
title: "Plan: Native Eval Runner with LLM-as-Judge"
type: plan
status: active
date: 2026-07-01
related:
  - ../findings/skilleval-analysis-2026-06-30.md
  - ../findings/eval-gating-byok-2026-06-29.md
  - ../findings/tessl-eval-criteria-schema-2026-06-30.md
  - ../plans/migrate-off-tessl-eval-2026-06-29.md
  - ../../docs/ADR/adr-001-native-eval-runner.md
  - ../../docs/ADR/adr-007-two-tier-eval-gate.md
  - ../../docs/ADR/adr-018-ci-eval-guardrails.md
  - ../../docs/ADR/adr-024-ab-test-eval-mode.md
  - ../../docs/ADR/adr-025-provider-agnostic-llm-client.md
---
# Plan: Native Eval Runner with LLM-as-Judge

## Goal

Replace the Tessl CLI dependency in CI (`tessl review run cmd/assets/` via `tesslio/setup-tessl@v2`) with a native `skill-auditor eval` subcommand that runs the existing 6 eval scenarios against an LLM, grades outputs against checklist criteria, and writes results compatible with the D9 scorer. This removes the Tessl CLI from CI, the `TESSL_TOKEN` secret, and the `tesslio/setup-tessl` GitHub Action — while adding reusable LLM infrastructure for future features like A/B testing (SkillEval pattern) and adaptive criteria generation.

> **Note**: The actual CI workflow runs `tessl review run` (not `eval run`). Docs reference `tessl eval run`. Both are Tessl CLI commands to replace — the native runner supersedes all Tessl CLI invocations for evaluation.

## Background

The existing eval infrastructure at `cmd/assets/evals/` has 6 scenarios, each with a `task.md` prompt, `criteria.json` weighted checklist (summing to 100), and `capability.txt`. The D9 scorer (`scorer/d9_eval_validation.go`) reads these structurally — it never calls Tessl. The only Tessl runtime dependency in CI is `tessl review run cmd/assets/` via the `tesslio/setup-tessl@v2` action (docs additionally reference `tessl eval run` — both are Tessl CLI commands).

The existing draft plan (`migrate-off-tessl-eval`) analysed four options and recommended Option A (native Go eval runner). Two key constraints from the BYOK findings:

1. **Two-tier gate**: Structural D9 (deterministic, no key) is the required PR gate. LLM-judge (non-deterministic, needs a key) is advisory — runs on `main` pushes or labelled PRs, never required on fork PRs.
2. **BYOK design**: Configurable model/endpoint/key source, graceful degradation when no key is available, recorded-not-pinned reproducibility.

> **ADR context**: ADR-001 records the native runner decision, ADR-007 records the two-tier gate + BYOK architecture, ADR-018 records CI guardrails (N-sample median, label/nightly cadence, cost logging, read-only CI, single-turn precondition), and ADR-024 records the A/B test mode. This plan implements ADR-001/007/018; ADR-024 is future work enabled by the LLM infrastructure introduced here.
>
> **Stale referenced finding**: `tessl-eval-criteria-schema-2026-06-30.md` describes the CI workflow as `curl -fsSL https://get.tessl.io | sh` with `TESSL_VERSION` pinning. The actual workflow uses `tesslio/setup-tessl@v2` with no curl or version pin. The finding's schema-migration content remains accurate; only its CI-workflow description is stale.

## Phases

### Phase 1 — LLM client abstraction

**Package**: `internal/llmclient/`

A thin Go HTTP client for LLM APIs. The model-call boundary is an **abstraction with multiple provider implementations**, per ADR-007 #3 ("not a hardcoded Anthropic client"). v1 ships four providers: Anthropic (this repo's default), OpenAI, Gemini (native Google API), and an OpenAI-compatible passthrough that covers Ollama, vLLM, and internal gateways via `LLM_BASE_URL`.

| File | What it does |
|------|-------------|
| `internal/llmclient/client.go` | `Client` interface (`Chat(ctx, req) → Response`), `Provider` factory interface, `NewFromEnv()` selector. Reads `LLM_PROVIDER`, provider-specific key env vars, `LLM_BASE_URL`, `LLM_MODEL`. Returns nil when the selected provider has no key → callers degrade gracefully. |
| `internal/llmclient/anthropic.go` | `AnthropicClient` — default model `claude-sonnet-4-20250514`, key `ANTHROPIC_API_KEY`, Messages API. |
| `internal/llmclient/openai.go` | `OpenAIClient` — default model `gpt-4o`, key `OPENAI_API_KEY`, Chat Completions API. Honours `LLM_BASE_URL` so Ollama/vLLM/internal gateways slot in unchanged. |
| `internal/llmclient/gemini.go` | `GeminiClient` — default model `gemini-2.0-flash`, key `GEMINI_API_KEY` (falls back to `GOOGLE_API_KEY`), native Google `generateContent` API. |
| `internal/llmclient/types.go` | `Message` (role, content), `Response` (content, usage), `Config` (provider, model, baseURL, key, temperature) |
| `internal/llmclient/prompt.go` | Helper to build the judge prompt from a skill's content, task.md, and criteria.json; embeds and exposes the prompt SHA256 for `judge_prompt_version` |
| `internal/llmclient/client_test.go` | Mock client + per-provider unit tests (auth, request shape, error mapping) + integration test with recorded fixture |

**Interface**:

```go
type Client interface {
    Chat(ctx context.Context, req Request) (*Response, error)
}

type Request struct {
    Messages    []Message
    Model       string
    Temperature float64
    MaxTokens   int
}
```

`NewFromEnv()` reads:
- `LLM_PROVIDER` — selects provider: `anthropic` (default) \| `openai` \| `gemini` \| `openai-compatible`. `openai-compatible` is OpenAI's client with a *required* `LLM_BASE_URL` (for Ollama/vLLM/gateways where no canonical default exists).
- Provider-specific key: `ANTHROPIC_API_KEY` / `OPENAI_API_KEY` / `GEMINI_API_KEY` (Gemini falls back to `GOOGLE_API_KEY`).
- `LLM_BASE_URL` (optional, overrides the selected provider's default endpoint — enables local models and internal gateways for any provider).
- `LLM_MODEL` (optional, overrides the provider's default model id).

No key for the selected provider → `NewFromEnv()` returns nil — callers degrade gracefully. This preserves ADR-007 #5 (graceful degradation) and #3 (BYOK as a design goal, not a TODO): a consumer with only an OpenAI or Gemini key gets a working LLM-judge, and a consumer who will not send content to any hosted API still gets the full structural D9 grade.

**Implementation details** (not exhaustively specified above but required for reliability):
- **Retries**: exponential backoff on 429/5xx, max 3 retries, with jitter.
- **Timeouts**: default 60s per request, configurable via context.
- **Rate limiting**: scenarios run concurrently with a bounded worker pool (default 3 concurrent) to avoid hitting provider rate limits.
- **Temperature**: actor calls use the model default; judge calls are pinned to `temperature: 0` for determinism. Document this in code comments.
- **Streaming**: the interface returns a complete `*Response`; if the provider streams, the client reassembles chunks before returning. No streaming interface is exposed to the caller.

### Phase 2 — `eval` subcommand

**Package**: `cmd/eval.go`

New cobra command:

```bash
skill-auditor eval <path-or-key> [flags]
  --provider <id>         # override provider: anthropic|openai|gemini|openai-compatible (default: from LLM_PROVIDER env, else anthropic)
  --model <id>            # override actor model (default: per-provider, e.g. claude-sonnet-4-20250514 / gpt-4o / gemini-2.0-flash)
  --judge-model <id>      # override judge model separately (defaults to --model)
  --fail-below <pct>      # exit non-zero below threshold (default: 0, advisory only)
  --write-summary         # update evals/summary.json in place (local use only)
  --json                  # machine-readable output to stdout
  --samples N             # run each scenario N times, report median (ADR-018 #2)
  --margin PCT            # advisory band width for pass/fail (default 5)
  --cost-log              # log raw token usage (input/output) to stderr for cost derivation (ADR-018 #4)
```

No `--skip-actor` flag. Key detection drives degradation per ADR-007 #5: no key → structural-only mode, said loudly. Adding a second manual control axis risks diverging from the auto-detection path.

**Flow per scenario**:

1. **Load**: Read `task.md` (prompt), `criteria.json` (checklist + optional `max_output_tokens`), `capability.txt`. The runner does **not** read `instructions.json` or `summary.json` — those are static markers consumed only by the D9 scorer. The runner reports pass/fail and per-item scores; it does not compute or emit coverage.
2. **Actor run**: Send skill content + task prompt to model. Capture output text (bounded by `max_output_tokens`).
3. **Judge run**: Send captured output + criteria.json to judge model with pinned rubric prompt. Receive per-item scores.
4. **Score**: Sum item scores (already normalised to 100), record per-item breakdown. If `--samples > 1`, repeat steps 2–3 and report the median per scenario.
5. **Report**: Output per-scenario pass/fail and per-item scores to stdout/JSON. No modification of `summary.json` except via explicit `--write-summary` (local authoring only).

> **Precondition**: This flow is valid for single-turn reasoning scenarios only. If a future scenario requires invoking the `skill-auditor` binary or other tools (multi-turn), the actor-run degrades to grading the model's imitation of tool use — at which point Option C (Agent SDK) becomes the correct architecture.

**Structural-only mode** (no key available): Grade existing output if present, or run structural validation only (scenario dirs exist, criteria sum to 100). Same command locally and in CI — only divergence is whether a key is present. Driven by `NewFromEnv()` returning nil, not a manual flag. **Note**: this is a schema-consistency gate, not a semantic quality gate. It ensures eval artifacts are well-formed but does not verify that skill content produces good actor outputs. The semantic quality gate is the LLM-judge step, which is advisory-only.

**Output** (stdout, text mode):

```
Scenario 1/6: Skill quality assessment
  ✓ Nine-dimension scoring: 22/25 — All 9 dimensions scored with justification
  ✓ Knowledge Delta accuracy: 18/20 — Correctly assigned ≤10/20 for redundant content
  ✓ Redundancy detection: 12/15 — SQL basics flagged as low-delta
  ✓ Grade threshold: 15/15 — A-grade threshold correctly stated as ≥126/140
  ✓ Remediation plan quality: 13/15 — File-level changes with S/M/L sizing
  ✓ Specification and disclosure gaps: 8/10 — D4 and D5 weaknesses identified
  Total: 88/100 → PASS (advisory threshold: 80)
```

**Output** (`--json`):

> `provider`, `model`, and `judge_model` reflect the **selected provider** (via `LLM_PROVIDER`/`--provider`), not hardcoded values. The example below shows the Anthropic default; an OpenAI run reports `"provider": "openai"` with `gpt-4o`, etc.

```json
{
  "skill_path": "skills/my-skill",
  "provider": "anthropic",
  "model": "claude-sonnet-4-20250514",
  "judge_model": "claude-sonnet-4-20250514",
  "judge_prompt_version": "sha256:abc123...",
  "samples": 3,
  "timestamp": "2026-07-01T12:00:00Z",
  "scenarios": [
    {
      "id": "scenario-01",
      "actor_output_snippet": "...",
      "scores": [
        {"name": "Nine-dimension scoring", "score": 22, "max_score": 25, "justification": "..."},
        {"name": "Knowledge Delta accuracy", "score": 18, "max_score": 20, "justification": "..."},
        {"name": "Redundancy detection", "score": 12, "max_score": 15, "justification": "..."},
        {"name": "Grade threshold", "score": 15, "max_score": 15, "justification": "..."},
        {"name": "Remediation plan quality", "score": 13, "max_score": 15, "justification": "..."},
        {"name": "Specification and disclosure gaps", "score": 8, "max_score": 10, "justification": "..."}
      ],
      "total": 88,
      "pass": true
    }
  ],
  "overall_pass": true
}
```

**Judge prompt** (pinned in code, lives in `internal/llmclient/prompt.go` alongside the builder per Phase 1 table):

```
You are grading whether an AI agent correctly completed a task using a reference skill.
...
Output a JSON object with a "scores" array:
{"scores": [{"name": "<criterion name>", "score": N, "max_score": M, "justification": "..."}]}
```

### Phase 3 — CI workflow changes

**File**: `.github/workflows/skill-quality.yml`

Keep the existing Tessl review step alongside the native runner (parallel run, advisory only). Remove Tessl only after the native runner has been running in CI for a proven period.

> **Aligns with ADR-018** (CI gating guardrails): N-sample median + ±5% margin band, label/nightly cadence (not every push), per-run cost logging, read-only CI. ADR-018 is `proposed` — if this plan adopts a different position on any of its five decisions, explicitly supersede the relevant ADR-018 item here rather than silently diverging.

```yaml
# Structural gate — required, no key needed, every PR/main push (ADR-007 Tier 1)
- name: Structural eval gate
  run: ./dist/skill-auditor eval cmd/assets --fail-below 0

# LLM-judge eval — advisory, label-triggered or nightly (ADR-018 #4 cadence)
- name: LLM-judge eval (advisory)
  if: |
    (github.event_name == 'schedule') ||
    (github.event_name == 'workflow_dispatch') ||
    (github.event_name == 'pull_request' && !github.event.pull_request.head.repo.fork && contains(github.event.pull_request.labels.*.name, 'run-eval'))
  env:
    LLM_PROVIDER: anthropic
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: ./dist/skill-auditor eval cmd/assets --json --samples 3 --cost-log > eval-results.json
  continue-on-error: true
- name: Upload eval results
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: eval-results
    path: eval-results.json
- name: Post eval results comment
  if: always() && github.event_name == 'pull_request'
  uses: actions/github-script@v7
  with:
    script: |
      const fs = require('fs');
      let results;
      try {
        results = JSON.parse(fs.readFileSync('eval-results.json', 'utf8'));
      } catch (e) {
        core.info('No eval-results.json found, skipping comment');
        return;
      }
      const status = results.overall_pass ? '✅ PASS' : '⚠️ ATTENTION';
      const lines = [
        `### Native Eval Results — ${status}`,
        `Provider: \`${results.provider || 'n/a'}\` | Judge prompt version: \`${results.judge_prompt_version || 'n/a'}\``,
        `Model: \`${results.model || 'n/a'}\` | Samples: ${results.samples || 1}`,
        '',
        '| Scenario | Total | Status |',
        '|----------|-------|--------|',
        ...(results.scenarios || []).map(s =>
          `| ${s.id} | ${s.total}/100 | ${s.pass ? '✓' : '✗'} |`
        ),
        '',
        `_[Advisory only — does not block merge]_`
      ];
      await github.rest.issues.createComment({
        ...context.repo,
        issue_number: context.issue.number,
        body: lines.join('\n')
      });

# Tessl review — kept while native runner is proven, made advisory
- uses: tesslio/setup-tessl@v2
  if: always()
  with:
    token: ${{ secrets.TESSL_TOKEN }}
- name: Tessl review run
  if: always()
  run: tessl review run cmd/assets/ --workspace pantheon-ai --json --threshold 80
  continue-on-error: true
```

> **Provider parameterization (ADR-007 #3):** this repo's CI uses Anthropic, but the runner honours any provider. To switch, set `LLM_PROVIDER` and the matching secret (e.g. `LLM_PROVIDER: openai` + `OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}`). Fork-PR guard and graceful degradation are unchanged: a provider with no key in CI degrades to structural-only.

**New flags required by ADR-018 alignment** (add to Phase 2 command spec):
- `--samples N` — run each scenario N times and report the median score (ADR-018 #2). Default 1; CI uses 3.
- `--cost-log` — log raw token usage (input/output) to stderr for cost derivation (ADR-018 #4). Dollar conversion is left to the operator.
- `--margin PCT` — advisory band width for pass/fail reporting (default 5). A score within the band is reported as `BAND` rather than `PASS`/`FAIL`.

A `schedule` trigger must be added to the workflow `on:` block (e.g. `cron: '0 6 * * *'` for nightly). The `run-eval` label allows on-demand runs from PRs.

Once the native runner has been passing reliably (suggest: 2 weeks of green CI runs on the native eval step), remove the two Tessl steps and the `TESSL_TOKEN` secret. **Measuring "green"**: because both steps use `continue-on-error: true`, GitHub Actions will report success regardless of the runner's internal pass/fail result. During the proving period, treat the native runner as green when its JSON output contains `"overall_pass": true` and no scenario failures. A nightly or labeled PR run that reports `"overall_pass": false` is a signal to investigate, not a CI block. Optionally, add a step that parses the JSON artifact and posts a non-blocking PR check or comment so failures are visible without blocking merge. **This is now required** — see "Proving period — non-blocking PR comment" in the Resolved open questions section below for the full step YAML.

**Pre-push hook** (`hk.pkl`): Add `skill-eval-structural` step to the `prePush` section:
```pkl
["skill-eval-structural"] {
    glob = List("cmd/assets/**")
    depends = List("go-build")
    check = "./dist/skill-auditor eval cmd/assets --fail-below 0"
}
```
No key needed, fast (~1s), enforces structural consistency on every push.

### Phase 4 — Documentation

| File | Change |
|------|--------|
| `AGENTS.md` | Replace "Tessl eval changes require `tessl eval run cmd/assets/` to pass" with "Eval changes require `skill-auditor eval` to pass" |
| `CONTRIBUTING.md` | Replace Tessl eval section with native eval command guidance; remove Tessl CLI from prerequisites if Layer 3 also dropped |
| `README.md` | Update "Tessl tile" wording if Layer 3 (distribution) decision made |
| `docs/d9-eval-validation.md` | Update wording: "eval scenarios" not "tessl eval scenarios" |
| `docs/d4-specification-compliance.md` | Update "installed via `tessl install`" wording to agent-agnostic phrasing |
| `cmd/assets/references/framework-dimensions.md` | Update D9 section: "tessl eval scenarios" → "eval scenarios"; remove `tessl eval run` / `tessl eval view-status` command examples |
| `cmd/assets/SKILL.md` | Remove "ensures tessl registry compliance" from description |
| `cmd/assets/references/tessl-compliance-framework.md` | Keep as registry-submission reference (Layer 3 / distribution only; decouple from eval) |

## File-by-file change list

| File | Change | Effort |
|------|--------|--------|
| `internal/llmclient/client.go` (new) | `Client` interface + `Provider` factory + `NewFromEnv()` selector | M |
| `internal/llmclient/anthropic.go` (new) | `AnthropicClient` — Messages API, default model `claude-sonnet-4-20250514` | M |
| `internal/llmclient/openai.go` (new) | `OpenAIClient` — Chat Completions API, default `gpt-4o`; honours `LLM_BASE_URL` for Ollama/vLLM | M |
| `internal/llmclient/gemini.go` (new) | `GeminiClient` — native Google `generateContent` API, default `gemini-2.0-flash` | M |
| `internal/llmclient/types.go` (new) | Request/Response/Config types (incl. `Provider` field) | S |
| `internal/llmclient/prompt.go` (new) | Judge prompt builder + embedded prompt hash/version | M |
| `internal/llmclient/client_test.go` (new) | Mock client + per-provider unit tests + fixture test | M |
| `cmd/eval.go` (new) | Cobra command + scenario runner | L |
| `cmd/eval_test.go` (new) | Tests with mocked client | M |
| `scorer/d9_eval_validation.go` | Optional: absorb coverage computation from runner | S |
| `.github/workflows/skill-quality.yml` | Swap Tessl steps for native eval | M |
| `hk.pkl` | Add structural eval to pre-push | S |
| `AGENTS.md` | Update eval rule | S |
| `CONTRIBUTING.md` | Replace Tessl eval section | S |
| `docs/d9-eval-validation.md` | Wording update | S |
| `cmd/assets/SKILL.md` | Description update | S |
| `.env.example` (new) | Document `LLM_PROVIDER`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, `LLM_MODEL`, `LLM_BASE_URL` | S |

## Decisions

| # | Question | Decision |
|---|----------|----------|
| 1 | Judge output format | Single JSON array (`{"scores": [...]}`) — simpler to parse, no streaming benefit for 6 items |
| 2 | Actor output truncation | Add optional `max_output_tokens` field to `criteria.json` (default 4096). Actor call respects it via `max_tokens` so judge always sees full bounded output |
| 3 | Cost guard | Not implemented now. Document as future improvement: if usage grows, add `--max-cost` flag that estimates and aborts before API calls. Per-run cost ≈ $0.30–0.60 (12 calls: 6 scenarios × actor + judge, ~2K input / ~500 actor output / ~1K judge output tokens at Claude Sonnet 4 pricing). Runner logs raw token usage per run per ADR-018 #4; dollar conversion is left to the operator. Annual cost well under $500 at ~10–20 main pushes/day. |
| 4 | `--write-summary` safety | Local-authoring only. CI outputs to stdout/JSON; static `summary.json` unchanged. Structural D9 scorer unaffected |
| 5 | Tessl review run removal | Keep alongside native runner with `continue-on-error: true`. Remove after 2 weeks of green CI runs on the native eval. |
| 6 | Provider scope (ADR-007 #3 compliance) | v1 ships **all four** provider implementations: `anthropic`, `openai`, `gemini`, and `openai-compatible` (OpenAI client + required `LLM_BASE_URL` for Ollama/vLLM/gateways). The earlier "TODO: OpenAI client" stance is superseded — ADR-007 elevates provider-agnosticism to a decision, not a deferred comment. Gemini ships a native Google API client (not the OpenAI-compatibility shim) because its native API is the common path for `GEMINI_API_KEY` users. Selection is via `LLM_PROVIDER` env or `--provider` flag; default `anthropic`. Per-provider default models: `claude-sonnet-4-20250514` / `gpt-4o` / `gemini-2.0-flash`. Recorded as ADR-025. |

## Amendments (resolved during critical review)

### Prior concerns from `migrate-off-tessl-eval` Section 11

| # | Concern | Resolution |
|---|---------|------------|
| 1 | **Flaky CI gate**: single-sample LLM-judge at `--fail-below 80` is non-deterministic | ✅ **Structural gate is the only required gate** (`--fail-below 0`, deterministic, no key). The LLM-judge step (`--json`, no `--fail-below`) is advisory-only with `continue-on-error: true`. No PR blocks on a non-deterministic sample. Document in `--fail-below` flag help text that passing thresholds with LLM-judge are approximate and may vary between runs. |
| 2 | **Actor-run fidelity conditional** | ✅ Explicit precondition accepted: scenarios are single-turn reasoning tasks. If future scenarios require tool execution (calling `skill-auditor` binary), Option C (Agent SDK) becomes necessary. Add precondition note to Phase 2 documentation. |
| 3 | **Circular subject/actor/judge**: the skill teaches Claude to score skills, actor is Claude applying it, judge is Claude grading it | ⚠️ **Known limitation, not blocking.** Pinning model and prompt mitigates but does not eliminate the circularity. Document in D9 scorer output: "LLM-judge score reflects internal consistency, not ground-truth correctness." No design change needed. |
| 4 | **Cost unquantified**: 6 scenarios × actor + judge = 12 calls per run | ✅ **Advisory-only mitigates cost concern.** At ~2K tokens per scenario input, ~500 output tokens for actor, ~1K for judge, 12 calls per run at Claude Sonnet 4 pricing ≈ $0.30–$0.60/run. Only runs on `main` pushes (~10–20/day). Annual cost well under $500. Add rough estimate to Decision #3. |
| 5 | **Coverage computation redundancy**: D9 already derives mutation coverage | ✅ **Runner does not write `summary.json`.** CI uses `--json` (stdout only). The `--write-summary` flag is for local authoring only. D9 scorer continues to read the static `summary.json` unchanged. Remove "coverage" step from scenario flow — the runner reports pass/fail and per-item scores only. |

### BYOK gaps

| Gap | Status |
|-----|--------|
| Fork PR secret visibility | ✅ Addressed: LLM-judge runs on `main` pushes only, never on fork PRs |
| Subscription vs API key (Claude Max OAuth) | ❌ **Known gap, deferred — OAuth/subscription auth only.** This is about *auth method* (subscription/OAuth vs API key), not about *provider choice*. Multi-provider support is decided (ADR-007 #3) and shipped (Decision #6). A subscription/OAuth path (Option C's merit) is not implemented; document as `docs/BYOK-LIMITATIONS.md` if a consumer requests it. |
| Provider-agnostic interface | ✅ Addressed (ADR-007 #3, Decision #6): `Client` + `Provider` abstraction with four v1 implementations — `anthropic`, `openai`, `gemini`, `openai-compatible`. `LLM_PROVIDER` env / `--provider` flag selects; per-provider key env vars; `LLM_BASE_URL` overrides endpoint for any provider. Replaces the earlier "TODO: OpenAI client" stance. |

### Missing items added

| Item | Addition |
|------|----------|
| **Branch name** | `feat/native-eval-runner` (from prior plan) |
| **`max_output_tokens` scenario migration** | Decision #2 requires this field in all 6 `criteria.json` files. Add to change list below: `cmd/assets/evals/scenario-*/criteria.json` — add `max_output_tokens` field (default 4096). |
| **ADR record** | ADR-001 (`docs/ADR/adr-001-native-eval-runner.md`) records the native runner decision; ADR-007 records the two-tier gate + BYOK; ADR-018 records CI guardrails. **Do not create `adr-007-native-eval-runner.md` — that number is taken by `adr-007-two-tier-eval-gate.md`.** ADR-025 (`adr-025-provider-agnostic-llm-client.md`) is concretely planned to record the provider-agnostic decision (Decision #6) that tightens ADR-007 #3 from "TODO" to "shipped"; create it via `adr-capture` and add to `docs/ADR/index.yaml`. Any further binding decisions take ADR-026+. |
| **HK hook specificity** | Pre-push structural eval: add `skill-eval-structural` step to `hk.pkl` pre-push section that runs `./dist/skill-auditor eval cmd/assets --fail-below 0` (no key needed, fast). |
| **Actor output caching** | When `--samples > 1`, cache the actor output after the first run and re-sample only the judge. This reduces API cost by ~33% (actor is ~40% of per-scenario cost). Store cache in memory only (ephemeral per invocation). |
| **Judge output validation** | If the judge returns malformed JSON or a scores array that does not match the criteria.json items, retry once with a reminder to output valid JSON. If still malformed, mark the scenario as failed with a diagnostic noting the parse error. |
| **Rollback plan** | Before deleting `TESSL_TOKEN` from repository secrets, record its current value in the team's password manager or 1Password vault. If the native runner proves unreliable, restoring Tessl CI requires only re-adding the secret and reverting the workflow commit. Do not delete the secret until 30 days after Tessl steps are removed from the workflow (soft-delete grace period). |
| **`.env.example`** | Document `LLM_PROVIDER`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, `LLM_MODEL`, `LLM_BASE_URL` with comments explaining each. |

### File-by-file change additions

| File | Change | Effort |
|------|--------|--------|
| `cmd/assets/evals/scenario-*/criteria.json` | Add `max_output_tokens` field (default 4096) to each scenario | S (×6) |
| `docs/ADR/adr-025-provider-agnostic-llm-client.md` (new) | Records the provider-agnostic decision (ADR-007 #3 compliance): four v1 providers, `LLM_PROVIDER` selection, per-provider keys/defaults. Supersedes the plan's earlier "TODO: OpenAI client" stance. Add to `docs/ADR/index.yaml` via `adr-capture`. | S |
| `docs/BYOK-LIMITATIONS.md` (new, optional) | Document known BYOK limitation: subscription/OAuth auth not supported (API-key auth only across all providers). Provider lock-in is no longer a limitation post-ADR-025. | S |
| `.env.example` (new) | Document `LLM_PROVIDER`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, `LLM_MODEL`, `LLM_BASE_URL` | S |

## Verification

1. `go test ./...` passes (new + existing tests).
2. `./dist/skill-auditor eval testdata/fixtures/skill-full` runs structural-only mode without a key.
 3. `ANTHROPIC_API_KEY=sk-... ./dist/skill-auditor eval cmd/assets --json --samples 3 --cost-log` produces valid JSON output with per-scenario median scores, `judge_prompt_version` present, raw token usage on stderr, and the CI artifact upload step captures the JSON file.
4. `summary.json` is byte-compatible with existing D9 scorer — `go test ./scorer/...` passes.
5. After proving period: Tessl steps removed from CI, no `TESSL_TOKEN` reference, `grep -ril tessl .github/workflows/` empty. (During proving period: Tessl steps kept with `continue-on-error: true`, per Phase 3.)
6. `grep -ril tessl . --exclude-dir=.git --exclude-dir=.context | grep -v CHANGELOG.md` shows only intentional references. The full set of tessl references found in the repo at review time (must each be either updated or explicitly bucketed as intentional Layer-3 distribution):
   - `.github/workflows/skill-quality.yml` — removed in this plan (Phase 3)
   - `cmd/assets/SKILL.md` — description wording (Phase 4)
   - `cmd/assets/references/framework-dimensions.md` — D9 section references "tessl eval scenarios" (add to Phase 4 change list)
   - `docs/d9-eval-validation.md` — wording (Phase 4)
   - `docs/d4-specification-compliance.md` — "installed via `tessl install`" wording (add to Phase 4)
   - `CONTRIBUTING.md` — Tessl CLI prerequisite + `tessl eval run` (Phase 4)
   - `AGENTS.md` — eval rule (Phase 4)
   - `.mcpx.json`, `.markdownlint-cli2.jsonc`, `.gitignore`, `scripts/setup-skills.sh` — Layer 3 (distribution/MCP config), intentionally retained
   - `tessl.json` — Layer 3, intentionally retained
   - `CHANGELOG.md` — historical, excluded by grep filter
7. ADR consistency: `docs/ADR/index.yaml` lists no `adr-007-native-eval-runner.md`. If a new ADR is created during implementation, it takes the next free slot (ADR-025+) and is added to the index by the `adr-capture` skill.
 8. ADR-018 alignment: the `--samples`, `--margin`, and `--cost-log` flags are implemented and exercised by the CI step in verification step 3.
 9. `criteria.json` schema tolerance: adding `max_output_tokens` to all 6 scenario criteria files does not break `go test ./scorer/...` or any JSON Schema validation.
10. Judge output resilience: a malformed JSON response from the judge triggers one retry; if still malformed, the scenario is marked failed with a diagnostic.
11. Actor output caching: running `--samples 3` on a single scenario produces exactly 1 actor API call and 3 judge API calls (verified via mock client call counts in tests).
12. Fork PR guard: a PR opened from a fork with the `run-eval` label does not trigger the LLM-judge step (the `!github.event.pull_request.head.repo.fork` condition prevents it).
13. PR comment step: a labeled PR run posts a comment with per-scenario scores and `overall_pass` status. The comment is non-blocking (does not affect merge).
14. Provider abstraction (ADR-007 #3): `LLM_PROVIDER` env / `--provider` flag selects among `anthropic`, `openai`, `gemini`, `openai-compatible`; an unknown value errors with a diagnostic listing valid providers.
15. Per-provider runs: `LLM_PROVIDER=openai OPENAI_API_KEY=... ./dist/skill-auditor eval cmd/assets --json` and `LLM_PROVIDER=gemini GEMINI_API_KEY=... ./dist/skill-auditor eval cmd/assets --json` each produce valid JSON with `"provider"` reflecting the selection (verified via mock clients per provider in `client_test.go`).
16. Graceful degradation per provider: selecting a provider whose key is absent → `NewFromEnv()` returns nil → structural-only mode, said loudly. No partial auth (e.g. `LLM_PROVIDER=openai` with only `ANTHROPIC_API_KEY` set) silently falls back to Anthropic — it degrades to structural-only.
17. `LLM_BASE_URL` override: `LLM_PROVIDER=openai-compatible LLM_BASE_URL=http://localhost:11434/v1 ...` routes to a local endpoint (Ollama); verified via mock/asserted request URL in tests.
18. JSON output `provider` field is present and matches the selected provider for every run (steps 3, 15, 17).

## Amendments (2026-07-01 critical review)

Critical review against the live codebase and ADR index found twelve issues during initial review; a subsequent critical review found an additional eight issues. Inline fixes applied to Phases 2–4, Decisions, Verification, and Missing Items; summary below.

### Blocking issues (resolved inline above)

| # | Issue | Resolution |
|---|-------|------------|
| 1 | **ADR-007 number collision + duplicate of ADR-001.** Amendments proposed `docs/ADR/adr-007-native-eval-runner.md`; that slot is taken by `adr-007-two-tier-eval-gate.md`, and `adr-001-native-eval-runner.md` already records the native runner decision. | "Missing items added" row corrected: do not create ADR-007; use ADR-025+ only if new binding decisions emerge. Frontmatter now links ADR-001/007/018/024. |
| 2 | **Contradicted ADR-018 (CI guardrails).** Plan ran single-sample LLM-judge on every `main` push with no cost logging; ADR-018 mandates N-sample median + ±5% band, label/nightly cadence, and per-run cost logging. | Phase 3 CI step rewritten: `--samples 3 --cost-log`, label/nightly/dispatch triggers instead of `push` to `main`. New flags `--samples`, `--margin`, `--cost-log` added to Phase 2 command spec. If the owner prefers the simpler single-sample/main-push design, explicitly supersede ADR-018 #2 and #4 rather than silently diverging. |
| 3 | **Plan didn't reference ADR-007/018/024.** Two-tier gate, BYOK, no-pre-push-LLM, and graceful degradation were re-derived as new resolutions despite being settled architecture. | Background now cites all four ADRs. "Amendments (resolved during critical review)" section retained but cross-references the ADRs that already record those decisions. |
| 4 | **Verification step 6 would fail.** `grep -ril tessl` finds unaddressed references in `framework-dimensions.md`, `d4-specification-compliance.md`, `.mcpx.json`, `.markdownlint-cli2.jsonc`, `.gitignore`, `scripts/setup-skills.sh`, `tessl.json`. | Verification step 6 now enumerates every tessl reference found at review time, bucketed as "updated by this plan" or "intentionally retained (Layer 3)". Phase 4 table expanded to include `framework-dimensions.md` and `d4-specification-compliance.md`. |
| 13 | **Coverage reported despite being out of scope.** The runner explicitly does not compute coverage, yet both text and JSON output examples included `Coverage: 92%` and `coverage_percentage`. | Coverage removed from both output examples. Runner reports pass/fail and per-item scores only. D9 scorer continues to derive coverage from `summary.json` independently. |
| 14 | **Proving period unmeasurable.** Both Tessl and native steps use `continue-on-error: true`, so "2 weeks of green CI runs" is meaningless — CI is always green. | Phase 3 updated: "green" means the native runner's JSON output shows `"overall_pass": true`. A nightly or labeled PR run that reports `"overall_pass": false` is a signal to investigate. Optional non-blocking PR check or comment recommended for visibility. |
| 15 | **Structural gate is too trivial.** The required PR gate (`--fail-below 0`) validates only directory existence and criteria sum-to-100. It is a schema lint, not a semantic eval gate. | Phase 2 structural-only mode updated with explicit note: "this is a schema-consistency gate, not a semantic quality gate." The semantic quality gate is the LLM-judge step, which is advisory-only. Documented downgrade from required semantic gate to required schema gate + advisory semantic gate. |

### Significant issues (resolved inline above)

| # | Issue | Resolution |
|---|-------|------------|
| 5 | **Cost estimate inconsistency.** Decision #3 said "~$1–3 per `main` push"; Amendment #4 derived "$0.30–0.60/run" from token counts. | Decision #3 replaced with the derived $0.30–0.60 figure and its token-count basis. |
| 6 | **Judge prompt filename mismatch.** Phase 1 table listed `prompt.go`; Phase 2 said `judge_prompt.go`. | Phase 2 corrected to reference `internal/llmclient/prompt.go` per the Phase 1 table. |
| 7 | **`instructions.json` handling undefined.** Phase 2 flow didn't mention it; Amendment #5 removed the coverage step without stating the runner's relationship to `instructions.json`. | Phase 2 step 1 now states explicitly: the runner does not read `instructions.json` or `summary.json`; those are static markers consumed only by the D9 scorer. |
| 16 | **Judge prompt versioning undefined.** The judge prompt is "pinned in code" but has no version. Score trends become incomparable if the prompt is refined later. | `judge_prompt_version` field added to JSON output example (value: SHA256 of prompt content). Prompt builder in `internal/llmclient/prompt.go` must embed and expose this hash. Score trends are valid only within a prompt version. |
| 17 | **`internal/llmclient/` under-specified.** No mention of retries, timeouts, rate limiting, concurrency, or judge temperature. | Added implementation details block to Phase 1: retries (3× exponential backoff), timeouts (60s), bounded concurrency (3 workers), judge temperature pinned to 0, streaming reassembled internally. |
| 18 | **Cost logging is hand-wavy.** `--cost-log` was described as logging "estimated API cost" but hardcoding prices in the binary goes stale quickly. | Decision #3 and Phase 2 flag help updated: `--cost-log` logs raw token usage (input/output) to stderr. Dollar conversion is left to the operator. |
| 19 | **Fork PR secret visibility risk.** The `if:` condition on the LLM-judge step triggers on `pull_request` + label, which includes fork PRs if a maintainer adds the label. `ANTHROPIC_API_KEY` would be empty; runner degrades silently to structural mode. | Phase 3 CI workflow updated with `!github.event.pull_request.head.repo.fork` guard in the `if:` condition. |
| 20 | **Missing operational elements.** No actor output caching (wastes API calls on `--samples > 1`), no judge output validation strategy for malformed JSON, no rollback plan if Tessl must be restored, `.env.example` described but not specified. | "Missing items added" table expanded: actor output caching (cache actor output, re-sample only judge), judge output validation (retry once, then fail scenario with diagnostic), rollback plan (soft-delete `TESSL_TOKEN` for 30 days), `.env.example` content described. |

### Minor issues (resolved inline above)

| # | Issue | Resolution |
|---|-------|------------|
| 8 | **Stray markdown fence** at line 177 opened an unclosed code block. | Fence removed. |
| 9 | **Decision #5 truncated** — no closing punctuation. | Sentence completed. |
| 10 | **`--skip-actor` flag vs key-detection.** ADR-007 #5 specifies graceful degradation by key detection; a manual flag creates a divergent control axis. | `--skip-actor` removed from command spec. Structural-only mode is driven by `NewFromEnv()` returning nil. |
| 11 | **Stale referenced finding.** `tessl-eval-criteria-schema-2026-06-30.md` mis-describes the CI workflow as curl-pipe-shell with `TESSL_VERSION`; actual workflow uses `tesslio/setup-tessl@v2`. | Background now notes the finding's schema content is accurate but its CI-workflow description is stale. |
| 12 | **`--workspace pantheon-ai` semantics dropped.** Tessl step uses `--workspace` (registry identity); plan maps `--threshold` to `--fail-below` but ignores workspace. | Noted: `--workspace` is a Layer-3 distribution concept. If Layer 3 is later removed, the workspace identity goes with it. No native-runner equivalent needed for eval; the flag is only relevant to the Tessl review step being retained during the proving period. |
| 23 | **`cmd/assets/references/tessl-compliance-framework.md` clarity.** Kept as registry reference but not explicitly marked as Layer-3-only. | Phase 4 table updated: `tessl-compliance-framework.md` is explicitly bucketed as "Layer 3 / Distribution only" to avoid confusion. |

### Provider-agnosticism gap (2026-07-01 BYOK compliance review)

A further review pass against ADR-007 found the plan **non-compliant** with decision #3 ("the model-call boundary is an abstraction, not a hardcoded Anthropic client") and the BYOK findings (`eval-gating-byok-2026-06-29.md` §3: "the eval boundary should be an interface, not a hardcoded Anthropic client ... at minimum support an OpenAI-compatible base URL"). The plan as written shipped only an `AnthropicClient`, read only `ANTHROPIC_API_KEY`, hardcoded `claude-sonnet-4-20250514`, and demoted provider-agnosticism to a "TODO: OpenAI client" code comment — directly contradicting a settled ADR decision.

| # | Issue | Resolution |
|---|-------|------------|
| 24 | **Plan contradicted ADR-007 #3.** Only `AnthropicClient` implemented; no OpenAI/Gemini; `NewFromEnv()` read only `ANTHROPIC_API_KEY`; default model hardcoded; BYOK gaps table marked provider-agnosticism "⚠️ Partially addressed". | Phase 1 rewritten: `Client` + `Provider` abstraction with four v1 implementations (`anthropic`, `openai`, `gemini`, `openai-compatible`). `NewFromEnv()` reads `LLM_PROVIDER` + per-provider keys + `LLM_BASE_URL` + `LLM_MODEL`. Phase 2 adds `--provider` flag + `provider` JSON field. Decision #6 records scope; ADR-025 records the binding decision; BYOK gaps table flipped to ✅; Resolved Q#4 clarified as OAuth-only; Verification steps 14–18 added. |

### Resolved open questions (2026-07-01 owner review)

| # | Question | Decision |
|---|----------|----------|
| 1 | **ADR-018 alignment**: keep `--samples`/`--margin`/`--cost-log` or supersede ADR-018 #2/#4 for simpler single-sample design? | **Keep ADR-018 alignment.** Implement `--samples`, `--margin`, `--cost-log` as specified. CI uses `--samples 3` on label/nightly/dispatch triggers. ADR-018 should be moved to `accepted` status as part of this plan's implementation. |
| 2 | **Layer 3 (distribution)**: should this plan also decide the fate of the Tessl tile? | **Keep Layer 3 — eval only.** This plan removes Tessl from CI eval gating only. The Tessl tile (tile.json, tessl.json, .mcpx.json, scripts/setup-skills.sh) remains for distribution. README keeps "Tessl tile" wording. A future plan can decide whether to keep or drop the tile. |
| 3 | **Cost guard (`--max-cost`)**: defer or implement now? | **Defer.** Not implemented in this plan. `--cost-log` provides raw token usage visibility (ADR-018 #4). Document `--max-cost` as a future improvement if usage grows. |
| 4 | **BYOK OAuth gap**: create `docs/BYOK-LIMITATIONS.md` proactively or defer? | **Defer until requested — OAuth/subscription auth only.** This gap is about *auth method* (subscription/OAuth vs API key), **not** provider choice. Multi-provider support is decided by ADR-007 #3 and shipped in v1 (Decision #6 / ADR-025). Only create `docs/BYOK-LIMITATIONS.md` if a consumer requests OAuth/subscription support. |
| 5 | **Proving period visibility**: add non-blocking PR comment step or rely on artifact + manual check? | **Add non-blocking PR comment step.** See Phase 3 "Proving period — non-blocking PR comment" section below. |
| 6 | **Provider scope for v1**: ship all providers or defer OpenAI/Gemini? | **Ship all four (anthropic, openai, gemini, openai-compatible) in v1.** ADR-007 #3 elevates provider-agnosticism to a decision; the earlier "TODO: OpenAI client" stance demoted it to a deferred comment and was non-compliant. Gemini ships a native Google API client (its common path for `GEMINI_API_KEY` users), not the OpenAI-compatibility shim. Recorded as ADR-025. See Decision #6. |

### Proving period — non-blocking PR comment

During the proving period, both Tessl and native steps use `continue-on-error: true`, so CI is always green. A non-blocking PR comment step parses the eval-results.json artifact and posts per-scenario scores and `overall_pass` status as a PR comment. This makes native eval failures visible without downloading the artifact.

Add after the "Upload eval results" step in Phase 3:

```yaml
- name: Post eval results comment
  if: always() && github.event_name == 'pull_request'
  uses: actions/github-script@v7
  with:
    script: |
      const fs = require('fs');
      let results;
      try {
        results = JSON.parse(fs.readFileSync('eval-results.json', 'utf8'));
      } catch (e) {
        core.info('No eval-results.json found, skipping comment');
        return;
      }
      const status = results.overall_pass ? '✅ PASS' : '⚠️ ATTENTION';
      const lines = [
        `### Native Eval Results — ${status}`,
        `Judge prompt version: \`${results.judge_prompt_version || 'n/a'}\``,
        `Model: \`${results.model || 'n/a'}\` | Samples: ${results.samples || 1}`,
        '',
        '| Scenario | Total | Status |',
        '|----------|-------|--------|',
        ...(results.scenarios || []).map(s =>
          `| ${s.id} | ${s.total}/100 | ${s.pass ? '✓' : '✗'} |`
        ),
        '',
        `_[Advisory only — does not block merge]_`
      ];
      await github.rest.issues.createComment({
        ...context.repo,
        issue_number: context.issue.number,
        body: lines.join('\n')
      });
```

**Comment dedup**: the script creates a new comment on each run. If this becomes noisy during the proving period, add a `find-and-update` strategy using a hidden HTML marker comment. Deferred until noise is observed.

**After proving period**: remove the PR comment step alongside Tessl removal, or keep it as a permanent advisory signal (owner's choice at that time).
