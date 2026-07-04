---
title: "Draft Plan: Integrate PR-Agent as an Advisory PR Review Bot"
type: plan
status: active
date: 2026-07-04
related:
  - ../findings/pr-agent-integration-2026-07-04.md
  - ../../.github/workflows/skill-quality.yml
  - ../../.github/workflows/pr-agent.yml
  - ../../.pr_agent.toml
  - ../../docs/ADR/adr-041-pr-agent-gemini-free-tier.md
---

## Goal

Add `The-PR-Agent/pr-agent` as an advisory GitHub Action that posts `/describe` and `/review` comments on pull requests, using Google AI Studio's free-tier Gemini API (no Anthropic/OpenAI key available) and this repo's Plumber pinning/permissions posture, without gating merge. `/improve` and `/ask` are deferred until a trial period shows the lighter-weight tools are worth the added noise. See `.context/findings/pr-agent-integration-2026-07-04.md` for the feasibility research this plan is based on.

## Scope

**In scope:**

- A new workflow file, `.github/workflows/pr-agent.yml`, separate from `skill-quality.yml`.
- `uses: docker://pragent/pr-agent@sha256:ea2ea90f072fd97708755e59827a317272f66097a1ef349eca23d39160bb0baf` (the `0.38.0-github_action` multi-arch index digest, confirmed against Docker Hub's tags API 2026-07-04) — digest-pinned, not the floating `the-pr-agent/pr-agent@main` ref shown in the project's own quick-start.
- `GEMINI_API_KEY` — a new repo secret, provisioned by a human from [Google AI Studio](https://aistudio.google.com/) (free tier, rate-limited, no billing required). No Anthropic or OpenAI key is available for this integration.
- `config.model: "gemini/gemini-1.5-flash"` with matching `fallback_models`, per PR-Agent's own documented Gemini quick-start.
- Only `github_action_config.auto_describe: "true"` and `github_action_config.auto_review: "true"` enabled.
- `permissions: { issues: write, pull-requests: write }` only — no `contents: write`.
- A `.pr_agent.toml` (or inline `extra_instructions`) tuned to this repo (Go idioms, CLI UX, test coverage) rather than defaults.
- `paths-ignore: ['cmd/assets/**']` on the workflow's `pull_request` trigger, so PRs touching only skill assets don't get a second, redundant LLM opinion from the existing `skill-quality.yml` judge (resolves the Open Question below).

**Out of scope (deferred):**

- `/improve` (inline code-suggestion comments) — deferred pending a trial-period read on signal-to-noise, per the finding's recommendation for a small, already-disciplined Go codebase.
- `/ask` — reactive, comment-triggered tool; adds no new infra decisions, so it can be turned on later without revisiting this plan.
- Anthropic/OpenAI/Bedrock as the LLM provider — no budget/key available for this integration; Gemini's free tier is the only option that doesn't require a paid credential.
- GitHub App / webhook deployment — the Action-based mode is consistent with this repo's existing CI patterns (Plumber, agent-scan, skill-quality) and needs no separate app installation.
- `/update_changelog` — not requested; avoids needing `contents: write`.
- Any workflow-file edits in this PR — this plan documents intent for review; implementation is a separate PR once the plan (and any resulting ADR) is accepted.

## Decisions (proposed — pending review)

1. **Advisory-only, no blocking gate.** PR-Agent's comments are informational; no CI status check is made required to merge. Justification: matches the proving-period pattern already established for the Tessl review step and the `agent-scan` plan; LLM review comments are inherently subjective and shouldn't block merge on their own.
2. **Start with `/describe` + `/review` only.** `/improve` and `/ask` are deferred to a later phase. Justification: per the finding, noisy inline suggestion comments risk adding review friction rather than reducing it on this codebase.
3. **Use Gemini (Google AI Studio free tier) via a new `GEMINI_API_KEY` secret — not Anthropic, OpenAI, or Bedrock.** Justification: no paid LLM credential is available for this integration; Gemini is the only PR-Agent-supported provider with a genuinely free tier (rate-limited, no billing account required). This supersedes the `ANTHROPIC_API_KEY`-reuse decision originally captured in ADR-040 — see that ADR's supersession record.
4. **Pin via Docker digest, not a floating branch ref.** The pinned digest is `sha256:ea2ea90f072fd97708755e59827a317272f66097a1ef349eca23d39160bb0baf` (Docker Hub tag `0.38.0-github_action`, confirmed 2026-07-04). Justification: this repo's Plumber policy (`actionsMustBePinnedByCommitSha`) would flag `the-pr-agent/pr-agent@main`; the project's own docs offer the digest-pinned form as the supported alternative.
5. **Do not add an entry to `.plumber.yaml`'s `trustedGithubActions`.** Justification: that list (and the `actionsMustBePinnedByCommitSha` / `githubActionMustComeFromAuthorizedSources` controls it backs) matches `uses: owner/repo@ref` GitHub Action references. A `uses: docker://...@sha256:...` step is a Docker image reference, not a GitHub-hosted action reference, so it falls outside what those controls parse — unlike `getplumber/plumber`, which genuinely is invoked as `uses: getplumber/plumber@<sha>`. If a live Plumber run flags the docker step anyway, add the entry then rather than pre-emptively.
6. **Permissions scoped to `issues: write` + `pull-requests: write` only.** Justification: `/update_changelog` isn't enabled in this rollout, so `contents: write` isn't needed — matches this repo's least-privilege posture.
7. **Own workflow file, not folded into `skill-quality.yml`.** Justification: distinct trigger semantics (all PRs vs. path-scoped skill assets) and a distinct concern (general code review vs. skill-content scoring) — mirrors the existing one-workflow-per-concern pattern (`ci.yml`, `plumber.yml`, `skill-quality.yml`).
8. **Exclude `cmd/assets/**`-only PRs via `paths-ignore` on the workflow trigger.** Justification: GitHub Actions skips a `paths-ignore`-filtered workflow only when every changed file matches the ignore pattern, so mixed PRs (Go + skill assets) still get reviewed while skill-only PRs — already covered by the `skill-quality.yml` LLM-judge — don't get a second, redundant opinion. Resolves the Open Question from the original draft.

## Phases

### Phase 0 — Prerequisites (human action, blocks Phase 1)

- Sign up for [Google AI Studio](https://aistudio.google.com/), generate a Gemini API key, add it as the `GEMINI_API_KEY` repository secret. **This is the only remaining blocker** — no Anthropic/OpenAI key is available, and Gemini's free tier needs no billing account.
- Confirm the `/improve`/`/ask` deferral (Decision 2) — a product-taste call, not an agent decision.
- Exit criterion: `GEMINI_API_KEY` exists in repo secrets.

### Phase 1 — Wire in the advisory workflow

- Create `.github/workflows/pr-agent.yml`: trigger on `pull_request` (`opened`, `reopened`, `ready_for_review`) with `paths-ignore: ['cmd/assets/**']`, plus `issue_comment`; `permissions: { issues: write, pull-requests: write }`; `uses: docker://pragent/pr-agent@sha256:ea2ea90f072fd97708755e59827a317272f66097a1ef349eca23d39160bb0baf`; env `GITHUB_TOKEN`, `config.model: "gemini/gemini-1.5-flash"`, `config.fallback_models: '["gemini/gemini-1.5-flash"]'`, `GOOGLE_AI_STUDIO.GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}`, `github_action_config.auto_describe: "true"`, `github_action_config.auto_review: "true"` (`auto_improve` left unset).
- Add a `.pr_agent.toml` scoping tone/focus to this repo (Go idioms, CLI UX, test coverage).
- Smoke-test on a throwaway PR to confirm `/describe` and `/review` comments post correctly and the Plumber gate passes with the pinned digest.
- Exit criterion: a PR touching non-`cmd/assets/` files shows PR-Agent `/describe` and `/review` comments; a PR touching only `cmd/assets/**` shows none; Plumber CI passes; no `contents: write` requested anywhere in the new workflow.

### Phase 2 — Observe

- Run across real PRs for a fixed window (proposed: 2 weeks, mirroring the proving-period language already used for Tessl review and `agent-scan`).
- Track: comment quality/signal, false positives, any overlap/noise on PRs that also touch `cmd/assets/**`, and rate-limit throttling against Gemini's free-tier quota.
- Exit criterion: enough observed runs to decide on enabling `/improve`/`/ask`, adjusting scope, upgrading to a paid tier, or dropping the integration.

### Phase 3 — Decide on next tier

- Based on Phase 2 observations: enable `/improve` and/or `/ask`, keep the current scope as-is, or retire the integration.
- If retained, write an ADR capturing the now-accepted decisions (mirroring `docs/ADR/adr-034-agent-scan-skill-security-scanning.md`'s pattern for `agent-scan`).
- Exit criterion: this plan's `status` flips to `done`; any resulting ADR reflects the final state.

## Risks

- **Free-tier rate limits** — Google AI Studio's free Gemini tier is rate-limited (requests/minute and/day); a burst of PR activity could throttle `/describe`/`/review` calls or delay comments. Mitigated by starting with only two lightweight tools enabled and tracking throttling in Phase 2; upgrading to a paid tier is the fallback if this repo's PR volume outgrows it.
- **Digest drift** — pinning by digest means picking up upstream fixes requires a manual re-pin; a forgotten digest can silently go stale. Mitigated by treating the digest as a recurring maintenance item, same as any other pinned action in this repo.
- **Docker-based action latency** — `uses: docker://...` steps pull an image on every run, unlike composite/JS actions; adds latency to every PR. Acceptable for an advisory step; revisit if it materially slows CI.
- **Review fatigue** — low-signal `/review` comments risk reviewers tuning the bot out entirely, the same alert-fatigue risk flagged in the `agent-scan` plan. Phase 2's observation window exists to catch this before Phase 3.
- **Plumber may still flag the docker step** — Decision 5's reasoning (docker refs fall outside the GitHub Action pinning/authorized-sources controls) is based on reading the control descriptions, not a live Plumber run against this exact workflow. If Plumber's `plumber.yml` run does flag it, add `the-pr-agent/pr-agent` to `trustedGithubActions` at that point.

## Verification

```bash
# After Phase 1 lands, confirm the workflow triggers and comments
gh pr view <pr-number> --json comments --jq '.comments[].body' | grep -iE "pr-agent|describe|review effort"

# Confirm a cmd/assets/-only PR does NOT trigger pr-agent.yml
gh run list --workflow=pr-agent.yml --json headBranch,event | jq .
```

## Open Questions

- None remaining from the original draft — provider (Gemini/free-tier), digest, and `cmd/assets/**` exclusion mechanism are now resolved above. The only open item is Phase 0's human action: provisioning `GEMINI_API_KEY`.
