---
title: "ADR-025: Provider-agnostic LLM client for the native eval runner"
status: proposed
date: 2026-07-01
context:
  - path: .context/plans/native-eval-runner-2026-07-01.md
  - path: .context/findings/eval-gating-byok-2026-06-29.md
  - path: docs/ADR/adr-007-two-tier-eval-gate.md
---

**Status:** Proposed
**Date:** 2026-07-01

## Context

ADR-007 #3 decides that "the model-call boundary is an abstraction (not a hardcoded
Anthropic client)" and that "base URL, model ID, and key source are configurable", with
the explicit consequence that "the eval runner needs an abstraction layer for model
providers, not a single hardcoded client". The BYOK findings
(`eval-gating-byok-2026-06-29.md` §3) reinforce this: "the eval boundary should be an
interface, not a hardcoded Anthropic client ... at minimum support an OpenAI-compatible
base URL so a local model (Ollama, vLLM) or internal gateway can be slotted in."

The native eval runner plan (`native-eval-runner-2026-07-01.md`) as originally written
was **non-compliant** with ADR-007 #3: it shipped only an `AnthropicClient`, read only
`ANTHROPIC_API_KEY`, hardcoded `claude-sonnet-4-20250514` as the default model, and
demoted provider-agnosticism to a "TODO: OpenAI client" code comment in the BYOK gaps
table. A consumer with only an OpenAI or Gemini key, or a data-governance constraint
mandating a specific provider or a local model, could not use the LLM-judge at all. This
matters because the product is a CLI that consumers point at their own skills — BYOK is a
design constraint, not a nice-to-have.

## Decision

Ship a **provider-agnostic LLM client** in v1 of the native eval runner with **four**
provider implementations:

1. **`anthropic`** — Messages API, default model `claude-sonnet-4-20250514`, key
   `ANTHROPIC_API_KEY`. This repo's default.
2. **`openai`** — Chat Completions API, default model `gpt-4o`, key `OPENAI_API_KEY`.
   Honours `LLM_BASE_URL` so Ollama, vLLM, and internal gateways slot in unchanged.
3. **`gemini`** — native Google `generateContent` API, default model `gemini-2.0-flash`,
   key `GEMINI_API_KEY` (falls back to `GOOGLE_API_KEY`). A native client is shipped
   (not the OpenAI-compatibility shim) because the native API is the common path for
   `GEMINI_API_KEY` users.
4. **`openai-compatible`** — the OpenAI client with a *required* `LLM_BASE_URL`, for
   local models and gateways where no canonical default endpoint exists.

Selection is via the `LLM_PROVIDER` environment variable (default `anthropic`) or the
`--provider` CLI flag. `NewFromEnv()` reads the provider-specific key env var and returns
nil when the selected provider has no key, preserving ADR-007 #5 (graceful degradation):
no key → structural-only mode, said loudly. A consumer who will not send content to any
hosted API still gets the full structural D9 grade.

This supersedes the plan's earlier "TODO: OpenAI client" stance, which demoted a settled
ADR-007 decision to a deferred code comment. Provider-agnosticism is a delivered
capability in v1, not a future task.

## Consequences

- A consumer with any of an Anthropic, OpenAI, or Gemini API key, or a local
  OpenAI-compatible endpoint, gets a working LLM-judge — ADR-007 #3 is satisfied on
  delivery.
- The model-call boundary is a `Client` + `Provider` abstraction, so adding a future
  provider is a new file implementing the interface, not a refactor of a hardcoded
  client.
- JSON output gains a `provider` field so runs are reproducible and auditable across
  providers; score trends are comparable only within a `(provider, model,
  judge_prompt_version)` triple.
- CI for this repo keeps `ANTHROPIC_API_KEY` (this repo's choice) but the workflow is
  parameterised via `LLM_PROVIDER` + the matching secret — switching providers is a
  config change, not a code change.
- Graceful degradation is keyed to the *selected* provider's key: selecting
  `LLM_PROVIDER=openai` with only `ANTHROPIC_API_KEY` set degrades to structural-only
  (no silent fallback to Anthropic), keeping the auth model predictable.
- This ADR does **not** address subscription/OAuth auth (e.g. Claude Max subscription).
  That remains a known, deferred gap documented in the plan's BYOK gaps table; it is an
  auth-method concern, not a provider-choice concern.
- v1 ships four implementations, adding ~2 medium files (`openai.go`, `gemini.go`) plus
  per-provider tests on top of the single-client design — a deliberate scope increase to
  honour ADR-007 #3 at delivery rather than after it.
