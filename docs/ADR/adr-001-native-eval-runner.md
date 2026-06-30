---
title: "ADR-001: Native Go eval runner for skill evaluation"
status: proposed
date: 2026-06-30
context:
  - path: .context/plans/migrate-off-tessl-eval.md
  - path: .context/findings/eval-gating-byok-2026-06-29.md
---

**Status:** Proposed
**Date:** 2026-06-30

## Context

The skill evaluation pipeline has a hard runtime dependency on the Tessl CLI and
`TESSL_TOKEN` secret. This creates CI friction, requires a third-party hosted
service, and prevents local contributors from running evals without a Tessl
account. The eval scenario format and the D9 structural scorer are both
self-owned and unaffected.

## Decision

Replace the Tessl-based evaluation step with a native `skill-auditor eval`
command that:

1. Loads scenarios from the existing `cmd/assets/evals/` directory format
2. Runs each task prompt against a pinned Claude model with the skill in context
3. Grades output against `criteria.json` using an LLM-as-judge call
4. Writes `summary.json` in the existing schema for D9 consumption
5. Exits non-zero below a configurable `--fail-below` threshold for CI

This is Option A from the evaluation plan. The Tessl CLI and `TESSL_TOKEN`
are removed from CI. Distribution and packaging (`tile.json`, `tessl.json`)
remain unchanged for this phase.

## Consequences

- One fewer third-party runtime dependency in CI
- Local eval runs without a Tessl account (requires `ANTHROPIC_API_KEY`)
- CI swaps `TESSL_TOKEN` for `ANTHROPIC_API_KEY` (one secret for another, but a standard one)
- LLM judge calls introduce non-determinism — mitigation needed before CI gating
- Distribution side (tile.json, registry) remains on Tessl for now
