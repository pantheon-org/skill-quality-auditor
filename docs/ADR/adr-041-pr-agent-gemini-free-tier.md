---
title: "ADR-041: PR-Agent advisory review bot — Gemini free tier, no Plumber allowlist entry"
status: proposed
date: 2026-07-04
context:
  - path: "docs/ADR/adr-040-pr-agent-advisory-review-bot.md"
  - path: ".context/findings/pr-agent-integration-2026-07-04.md"
  - path: ".context/plans/pr-agent-integration-2026-07-04.md"
---

**Status:** Proposed
**Date:** 2026-07-04

## Context

ADR-040 proposed integrating `The-PR-Agent/pr-agent` reusing this repo's `ANTHROPIC_API_KEY` secret. During implementation it emerged that no Anthropic (or OpenAI, or AWS Bedrock) credential is actually available for this integration — all three are paid providers. Of PR-Agent's supported LLM providers, only Google AI Studio's Gemini API offers a genuinely free tier (rate-limited, no billing account required), so the provider choice changes. Two other implementation details were also resolved that ADR-040 had left open: the exact Docker digest to pin (confirmed via Docker Hub's tags API against the `0.38.0-github_action` tag), and the mechanism for excluding `cmd/assets/**`-only PRs from double LLM commentary (a `paths-ignore` filter on the workflow trigger, resolved as sufficient without needing `.plumber.yaml` changes for the docker-based `uses:` step). ADR-040 is immutable per this repo's ADR-capture rules, so it is marked superseded rather than edited; this ADR restates the decisions that carry over unchanged and replaces the ones that don't.

## Decision

1. **Integrate as an advisory-only bot — no required CI status check, no blocking gate.** *(Unchanged from ADR-040 #1.)*
2. **Enable only `/describe` and `/review` initially**; `/improve` and `/ask` deferred. *(Unchanged from ADR-040 #2.)*
3. **Use Gemini via Google AI Studio's free tier, not Anthropic.** New secret `GEMINI_API_KEY`, `config.model: "gemini/gemini-1.5-flash"` with matching `fallback_models`. *(Replaces ADR-040 #3.)* Justification: no paid LLM credential exists for this integration; Gemini's free tier is the only supported provider that doesn't require one. Free-tier rate limits are an accepted tradeoff, tracked during the Phase 2 observation window in the plan.
4. **Pin the action via the confirmed Docker digest** — `docker://pragent/pr-agent@sha256:ea2ea90f072fd97708755e59827a317272f66097a1ef349eca23d39160bb0baf` (Docker Hub tag `0.38.0-github_action`, resolved 2026-07-04), not the floating `the-pr-agent/pr-agent@main` ref. *(Refines ADR-040 #4 — same intent, digest now concrete instead of a placeholder.)*
5. **Do not add an entry to `.plumber.yaml`'s `trustedGithubActions`.** *(Reverses ADR-040 #5.)* Justification: that allowlist and the `actionsMustBePinnedByCommitSha` control it backs match `uses: owner/repo@ref` GitHub Action references. A `uses: docker://...@sha256:...` step is a Docker image reference, not a GitHub-hosted action reference, so on inspection it falls outside what those controls parse — unlike `getplumber/plumber`, which is genuinely invoked as `uses: getplumber/plumber@<sha>`. If a live Plumber run flags the docker step anyway, the allowlist entry gets added then, not pre-emptively.
6. **Scope permissions to `issues: write` + `pull-requests: write` only — no `contents: write`.** *(Unchanged from ADR-040 #6.)*
7. **Lives in its own workflow file, `.github/workflows/pr-agent.yml`.** *(Unchanged from ADR-040 #7.)*
8. **Exclude `cmd/assets/**`-only PRs via `paths-ignore` on the workflow trigger.** *(Resolves ADR-040 #8, which had deferred this.)* Justification: GitHub Actions skips a `paths-ignore`-filtered workflow only when every changed file matches the ignore pattern, so mixed PRs still get reviewed while skill-only PRs — already covered by the `skill-quality.yml` LLM-judge — don't get a redundant second opinion.
9. **Gating (or dropping) the integration is revisited after a fixed observation window** (2 weeks of real PRs). *(Unchanged from ADR-040 #9, cost language replaced with rate-limit language.)*

## Consequences

- **Easier:** every PR gets an automated summary (`/describe`) and structured review comment (`/review`) at zero marginal LLM cost, using a free-tier credential instead of drawing on the paid Anthropic quota already committed to the nightly LLM-judge.
- **Easier:** because the step never blocks merge, an unproven third-party bot and an unproven free-tier quota can't stall shipping.
- **Easier:** skipping the Plumber allowlist change removes one moving part from Phase 1 — fewer files touched, one less thing that could be wrong.
- **Harder:** Gemini's free tier is rate-limited; a burst of PR activity could throttle or delay comments, and there's no existing usage pattern in this repo to estimate headroom against. Tracked explicitly in the plan's Phase 2.
- **Harder:** Decision 5's reasoning about Plumber's control scope is inferred from reading `.plumber.yaml`'s documented control descriptions, not verified against a live Plumber run on this exact workflow file — flagged as a plan Risk, not asserted as certain.
- **Harder:** this repo now has two Anthropic-adjacent facts to keep straight for future readers — the nightly LLM-judge uses `ANTHROPIC_API_KEY`, PR-Agent uses `GEMINI_API_KEY` — a minor but real source of confusion if a future change assumes both use the same provider.
