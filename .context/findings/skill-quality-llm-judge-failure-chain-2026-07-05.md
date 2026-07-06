---
title: "Finding: skill-quality.yml's LLM-judge failure chain — wrong diagnosis, missing key, then three rate-limited providers"
type: FINDING
status: ACTIVE
date: 2026-07-05
value: MEDIUM
related:
  - ../../.github/workflows/skill-quality.yml
  - ../../internal/llmclient/client.go
  - ../../docs/ADR/adr-025-provider-agnostic-llm-client.md
---
# Finding: skill-quality.yml's LLM-judge failure chain

> What started as "bump a retired model string" turned into: the model bump was correct but irrelevant (no Anthropic key exists in this org at all), the replacement provider (Gemini) hit a 5 req/min free-tier wall, its replacement (Cerebras) hit the same class of wall, and its replacement (Mistral) did best but still failed a third of the time — until `internal/llmclient` gained Retry-After-aware backoff, which is the fix that actually matters here, not the choice of provider.

## Summary

Investigating a corrected finding from an earlier PR-Agent session (`gemini-1.5-flash` was retired) raised the question of whether `skill-quality.yml`'s hardcoded `claude-sonnet-4-20250514` had the same problem. It had — Anthropic retired it 2026-06-15 — but fixing it didn't fix anything, because this org has no `ANTHROPIC_API_KEY` at all. Every step below was verified against the live API via `gh workflow run skill-quality.yml --ref <branch>` and a downloaded `eval-results.json` artifact, not assumed.

## Detail

**Step 1 — the initial (wrong) diagnosis.** `claude-sonnet-4-20250514` retired 2026-06-15 per Anthropic's model-deprecations page; recommended replacement `claude-sonnet-4-6`. Bumped `internal/llmclient/client.go`'s `DefaultModelAnthropic`. Live-tested via `workflow_dispatch` — still `"structural_only": true`, zero API errors logged.

**Step 2 — tracing the real cause.** `cmd/eval.go:201` sets `StructuralOnly: r.client == nil` — a deliberate graceful-degradation path (ADR-007 #5) triggered when the provider's API key is empty, not when a call fails. A retired-model error would have surfaced as an actual HTTP error. `gh secret list --repo` showed nothing, the same symptom `GEMINI_API_KEY` showed before it turned out to be an org secret. User confirmed: **no `ANTHROPIC_API_KEY` exists in this org at all** — only `GEMINI_API_KEY`, `MISTRAL_API_KEY`, `CEREBRAS_API_KEY`. `internal/llmclient`'s Anthropic support and default model were never wrong; only `skill-quality.yml`'s hardcoded `LLM_PROVIDER: anthropic` was.

**Step 3 — Gemini.** Switched `skill-quality.yml` to `LLM_PROVIDER: gemini`. Live test: genuinely connects (one scenario scored a real 100/100), but 5 of 6 scenarios hit `429 RESOURCE_EXHAUSTED`: *"Quota exceeded for metric: generativelanguage.googleapis.com/generate_content_free_tier_requests ... limit: 5, model: gemini-3.5-flash"*. Also discovered and fixed in passing: `DefaultModelGemini` (`gemini-2.0-flash`) was unverified against the live API; bumped to `gemini-3.5-flash`, which the same test confirmed working.

**Step 4 — adding Mistral and Cerebras as real options.** `internal/llmclient` had no native support for either (only Anthropic, OpenAI, Gemini, openai-compatible). Both expose OpenAI-wire-compatible chat completions APIs, so `mistral.go` and `cerebras.go` reuse `OpenAIClient` rather than duplicating HTTP/JSON logic — the same pattern already used for `openai-compatible`. Added `ProviderMistral`/`ProviderCerebras`, default models (`mistral-small-2603`, `gpt-oss-120b`), default base URLs (confirmed via each provider's own docs, not guessed), `MISTRAL_API_KEY`/`CEREBRAS_API_KEY` env wiring, and made `OpenAIClient`'s error messages provider-aware (`c.cfg.Provider` instead of a hardcoded `"openai"` string, since three more providers now share that code path). 5 new tests, all passing.

**Step 5 — Cerebras live test.** Connects successfully (real generated content confirmed), but **0 of 6** scenarios passed: `429 "Requests per minute limit exceeded"`.

**Step 6 — Mistral live test.** Connects successfully, **3 of 6** scenarios passed with real scores — the best of the three — but the other 3 hit `429 "Rate limit exceeded"`.

**Step 7 — the actual conclusion.** All three providers authenticate and respond correctly. All three fail under this eval runner's call volume (3 samples × 6 scenarios × actor+judge calls, up to 3 scenarios concurrently per `maxConcurrent: 3`). This is not a "pick a better provider" problem — it's a call-pacing problem. The existing retry logic (`Backoff`, 3 attempts, capped at 8s) is sized for transient 5xx errors, not for waiting out a per-minute quota window.

**Step 8 — the actual fix.** Added `RetryAfter()` (parses the standard HTTP `Retry-After` header — seconds or HTTP-date) and `RateLimitBackoff()` (uses it when the provider sends one; otherwise a 15s-base exponential fallback capped at 90s, long enough to have a real chance against a per-minute quota). Wired into every adapter's retry loop (`anthropic.go`, `openai.go`, `gemini.go`) via a shared `retryDelay()` helper that branches on status code (429 → `RateLimitBackoff`, else → the original `Backoff`). Bumped `MaxRetryAttempts` from 3 to 4 to give the longer waits room to matter. This benefits every provider this package supports, not just the one active in CI today.

**Step 9 — final config and confirmed fix.** `skill-quality.yml` set to `LLM_PROVIDER: mistral` (best empirical performer), now backed by the new backoff logic. Re-verified via `workflow_dispatch`: **all 6 scenarios passed with real scores** (`overall_pass: true`, genuinely — not the earlier silent `structural_only` false-positive). The backoff fix, not the provider choice, is what actually resolved this.

## Recommended Action

The corrected fix (native Mistral/Cerebras support + Retry-After-aware backoff) is a binding architectural change worth an ADR — see `docs/ADR/adr-042-llm-judge-provider-and-backoff.md`. No further action needed here beyond what that ADR and PR #178 already capture; this finding exists so the next person who sees `skill-quality.yml` say `LLM_PROVIDER: mistral` understands why it isn't Anthropic (no key) or Gemini (rate limits, tried first) and doesn't re-litigate the choice from scratch.
