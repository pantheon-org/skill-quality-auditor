---
title: "ADR-042: skill-quality.yml's LLM-judge uses Mistral; internal/llmclient gains Retry-After-aware backoff"
status: accepted
date: 2026-07-05
context:
  - path: ".context/findings/skill-quality-llm-judge-failure-chain-2026-07-05.md"
---

**Status:** Accepted
**Date:** 2026-07-05

## Context

`.context/findings/skill-quality-llm-judge-failure-chain-2026-07-05.md` documents a chain of corrections: a retired-Anthropic-model diagnosis was confirmed but proved irrelevant (this org has no `ANTHROPIC_API_KEY` at all), Gemini was tried next and genuinely connects but hits its free tier's 5 req/min cap under this eval runner's call volume, Cerebras hits the same class of per-minute rate limit, and Mistral does best (3/6 scenarios) but still fails partially — until `internal/llmclient` gained retry logic that honors the standard `Retry-After` header, at which point all 6 scenarios passed cleanly. Every step was verified live via `workflow_dispatch` runs and downloaded `eval-results.json` artifacts, not assumed.

## Decision

1. **`skill-quality.yml`'s LLM-judge uses `LLM_PROVIDER: mistral`, not Anthropic, Gemini, or Cerebras.** Justification: no Anthropic key exists in this org; of the three genuinely available keys (Gemini, Mistral, Cerebras), Mistral tolerated this eval runner's burst call pattern best in a controlled comparison, and — combined with Decision 3 below — now clears all 6 scenarios reliably.
2. **`internal/llmclient` gains native `ProviderMistral` and `ProviderCerebras` support**, reusing `OpenAIClient`'s wire format since both APIs are OpenAI-compatible (new `mistral.go`/`cerebras.go`, `MISTRAL_API_KEY`/`CEREBRAS_API_KEY` env wiring, default models/base URLs). This is available to any caller of the package, not just `skill-quality.yml` — Cerebras remains a documented, tested fallback option even though Mistral is the current default.
3. **`internal/llmclient`'s retry logic now honors the standard HTTP `Retry-After` header on 429 responses** (`RetryAfter()`, `RateLimitBackoff()`), falling back to a 15s-base/90s-capped exponential schedule when a provider doesn't send one — a deliberately longer schedule than `Backoff()`'s 8s cap, which is sized for transient 5xx errors, not per-minute quota windows. This is the change that actually fixed the eval runner, not the provider swap; it applies uniformly to every provider adapter (Anthropic, OpenAI, Gemini, Mistral, Cerebras, openai-compatible).
4. **`MaxRetryAttempts` increased from 3 to 4** across all adapters, giving the longer 429 backoff room to matter (3 wait opportunities instead of 2).
5. **`DefaultModelGemini` bumped from `gemini-2.0-flash` (unverified) to `gemini-3.5-flash` (confirmed working live)** and **`DefaultModelAnthropic` bumped from the retired `claude-sonnet-4-20250514` to `claude-sonnet-4-6`**, in passing — both correct regardless of which provider `skill-quality.yml` ends up using, since `internal/llmclient` is a general-purpose package used by anyone with any of these keys.

## Consequences

- **Easier:** `skill-quality.yml`'s advisory LLM-judge now produces real, non-degraded scores (confirmed: all 6 scenarios passing with genuine content-based grades) instead of silently reporting a false `overall_pass: true` with zero actual evaluation, as it had been doing for at least three weeks.
- **Easier:** the Retry-After fix benefits every current and future caller of `internal/llmclient`, not just this one workflow — any provider's free-tier rate limiting is now handled correctly by default.
- **Easier:** Cerebras is available as a documented, tested fallback if Mistral's free tier ever becomes insufficient, without needing new Go code — just an env var change.
- **Harder:** retries can now legitimately take significantly longer (up to ~90s per wait, several waits per call) when a provider is genuinely rate-limiting — acceptable for an advisory, non-blocking CI step with no tight timeout, but worth knowing if this pattern is reused somewhere latency-sensitive.
- **Harder:** this repo now depends on three different LLM provider accounts (Gemini, Mistral, Cerebras) plus Anthropic support in code with no matching key — slightly more moving parts to track than a single-provider setup, mitigated by all of it being centralized in one small package.
