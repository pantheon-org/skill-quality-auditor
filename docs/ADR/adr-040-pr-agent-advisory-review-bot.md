---
title: "ADR-040: PR-Agent as an advisory PR review bot, describe+review only"
status: superseded
date: 2026-07-04
superseded_by: "adr-041"
context:
  - path: ".context/findings/pr-agent-integration-2026-07-04.md"
  - path: ".context/plans/pr-agent-integration-2026-07-04.md"
---

**Status:** Superseded by ADR-041
**Date:** 2026-07-04

## Context

`.context/findings/pr-agent-integration-2026-07-04.md` confirmed `The-PR-Agent/pr-agent` is the actively maintained, Apache-2.0, community-owned continuation of Codium/Qodo's original PR-Agent (not Qodo's own separate paid "Qodo 2.0" product). It adds LLM-backed `/describe`, `/review`, `/improve`, and `/ask` comments to pull requests, deployable as a GitHub Action, and supports Anthropic as an LLM provider — meaning this repo's existing `ANTHROPIC_API_KEY` secret (already used by the nightly LLM-judge in `skill-quality.yml`) can be reused. The finding also identified friction points specific to this repo: its own quick-start uses a floating `@main` ref that this repo's Plumber policy (`actionsMustBePinnedByCommitSha`) would flag, its example workflow requests broader `contents: write` permission than needed, and running it unscoped would produce a second, redundant LLM opinion on `cmd/assets/**` diffs already covered by the skill-quality LLM-judge. `.context/plans/pr-agent-integration-2026-07-04.md` lays out a phased, advisory-first rollout addressing each of these. This ADR captures the binding decisions from that plan.

## Decision

1. **Integrate as an advisory-only bot — no required CI status check, no blocking gate.** PR-Agent's comments are informational. This mirrors the proving-period pattern already used for the Tessl review step and `agent-scan`.
2. **Enable only `/describe` and `/review` initially.** `/improve` (inline code-suggestion comments) and `/ask` are deferred pending a trial-period read on signal-to-noise, given the risk of low-signal suggestion comments adding review friction on this codebase.
3. **Reuse `ANTHROPIC_API_KEY` rather than provisioning a new secret**, via `config.model: "anthropic/claude-<model>"` + `ANTHROPIC.KEY: ${{ secrets.ANTHROPIC_API_KEY }}`. Revisit only if usage/cost data during the observation window shows contention with the nightly LLM-judge.
4. **Pin the action via Docker digest with attestation verification (`docker://pragent/pr-agent@sha256:<digest>`, `gh attestation verify`), not the floating `the-pr-agent/pr-agent@main` ref shown in the project's own quick-start.** Required to pass this repo's `actionsMustBePinnedByCommitSha` Plumber control.
5. **Add `the-pr-agent/pr-agent` to `.plumber.yaml`'s `trustedGithubActions`.** Same precedent as adding `getplumber/plumber` itself — a third-party, non-same-org action needs an explicit authorized-sources entry.
6. **Scope permissions to `issues: write` + `pull-requests: write` only — no `contents: write`.** `/update_changelog` (the only tool needing `contents: write`) is not enabled in this rollout.
7. **Lives in its own workflow file, `.github/workflows/pr-agent.yml`, not folded into `skill-quality.yml`.** Distinct trigger semantics (all PRs vs. path-scoped skill assets) and a distinct concern (general code review vs. skill-content scoring) — mirrors this repo's existing one-workflow-per-concern pattern.
8. **The exact mechanism for excluding or de-duplicating commentary on `cmd/assets/**` diffs is resolved during Phase 1 implementation, not fixed by this ADR.** The plan's Open Questions leave open whether that's a workflow-level path filter, a `.pr_agent.toml` ignore setting, or accepting the overlap pending Phase 2 observation.
9. **Gating (or dropping) the integration is revisited after a fixed observation window** (proposed: 2 weeks of real PRs, matching the language already used for the Tessl and `agent-scan` proving periods), based on comment quality, false-positive rate, and cost — not left advisory indefinitely by default, and not expanded to `/improve`/`/ask` without that observation period.

## Consequences

- **Easier:** every PR gets an automated, PR-visible summary (`/describe`) and structured review comment (`/review`) without requiring reviewers to write these by hand — using this repo's existing Anthropic credential, at no new secret-provisioning cost.
- **Easier:** because the step never blocks merge during the advisory period, an unproven third-party bot can't stall shipping — consistent with how Tessl review and `agent-scan` were both handled.
- **Harder:** a new third-party action needs Plumber allowlisting and digest-pinning maintenance (re-pinning on upstream updates) that didn't exist before.
- **Harder:** PR-Agent now shares `ANTHROPIC_API_KEY` usage with the nightly LLM-judge; a spike in PR volume could raise costs or contend for quota with that existing consumer — tracked during the Phase 2 observation window, not assumed to be a non-issue.
- **Harder:** until the Phase 1 exclusion mechanism is chosen and implemented, PRs touching `cmd/assets/**` may see two separate LLM opinions on the same diff (the skill-quality judge and PR-Agent's `/review`) — acceptable short-term given Phase 2 exists specifically to judge whether this is actually noisy in practice, but this ADR's status should move to `accepted` (or be superseded) once Phase 3's decision is made, not left `proposed` indefinitely.
