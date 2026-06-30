---
title: "ADR-007: Two-tier eval gate with BYOK architecture for skill evaluation"
status: proposed
date: 2026-06-30
context:
  - path: .context/findings/eval-gating-byok-2026-06-29.md
  - path: .context/plans/migrate-off-tessl-eval-2026-06-29.md
---

**Status:** Proposed
**Date:** 2026-06-30

## Context

The plan to replace Tessl runtime evaluation with a native Go eval runner (ADR-001) originally proposed a single hard gate (`--fail-below 80` on an LLM-judge call). Review found three unresolved constraints: fork PRs cannot access CI secrets, single-sample LLM-judge is non-deterministic, and consumers need to bring their own key (BYOK) without sending skill content to a hosted API.

## Decision

Adopt a **two-tier gate** model that separates deterministic structural checks from non-deterministic LLM-judge evaluation:

1. **Tier 1 — Structural D9 (required, every PR, everywhere):** Pure Go, deterministic, no network, no key. Runs via `go test ./scorer/...` and `skill-auditor evaluate --fail-below B`. Already works. This is the real merge gate.

2. **Tier 2 — LLM-judge eval (advisory, non-blocking):** Runs on same-repo pushes, label-triggered, or nightly against main. Posts per-scenario scores and trend delta as a PR comment or report. Uses N-sample median with a margin band, not a knife-edge threshold.

3. **BYOK is an explicit design goal:** The model-call boundary is an abstraction (not a hardcoded Anthropic client). Base URL, model ID, and key source are configurable. A consumer without a key gets the full structural D9 grade. The eval is fully optional.

4. **No runtime eval in pre-push:** The hk pre-push hook runs only the structural scorer. The LLM-judge runner is too slow, needs network and a key, and costs money per push.

5. **Graceful degradation:** `skill-auditor eval` detects key presence and degrades gracefully — no key means structural-only, said loudly.

## Consequences

- Fork PRs can pass the structural gate without secrets, resolving a hard CI constraint
- LLM-judge non-determinism is isolated to an advisory signal, not a blocking gate
- Consumer BYOK is a first-class design constraint, not an afterthought
- Structural gate remains identical locally and in CI (key or no key)
- CLI cost: the eval runner needs an abstraction layer for model providers, not a single hardcoded client
