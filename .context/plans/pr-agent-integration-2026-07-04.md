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
- `config.model: "gemini/gemini-3.5-flash"` (the current stable Gemini flash-tier model — `gemini-1.5-flash`, used in an earlier draft of this plan, has since been retired by Google) with `fallback_models` chained to Mistral (`mistral/mistral-small-2603`) then Cerebras (`cerebras/gpt-oss-120b`), using two additional org-provisioned keys (`MISTRAL_API_KEY`, `CEREBRAS_API_KEY`) discovered during Phase 1 smoke testing.
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

- Create `.github/workflows/pr-agent.yml`: trigger on `pull_request` (`opened`, `reopened`, `ready_for_review`) with `paths-ignore: ['cmd/assets/**']`, plus `issue_comment`; job-level `if` guards on non-bot sender AND non-fork PR (`github.event.pull_request.head.repo.fork != true` — `!= true`, not `== false`, because `issue_comment` events don't populate that path at all and must still run); `permissions: { issues: write, pull-requests: write }`; `uses: docker://pragent/pr-agent@sha256:ea2ea90f072fd97708755e59827a317272f66097a1ef349eca23d39160bb0baf`; step-level `continue-on-error: true` (advisory-only per Decision 1 — a Gemini outage/rate-limit must never redden this check); env `GITHUB_TOKEN`, `config.model: "gemini/gemini-3.5-flash"`, `config.fallback_models: '["mistral/mistral-small-2603", "cerebras/gpt-oss-120b"]'`, `GOOGLE_AI_STUDIO.GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}`, `MISTRAL.KEY: ${{ secrets.MISTRAL_API_KEY }}`, `CEREBRAS_API_KEY: ${{ secrets.CEREBRAS_API_KEY }}` (raw env var, not a dotted config key — PR-Agent has no special-case handling for Cerebras, unlike Mistral's `MISTRAL.KEY`), `github_action_config.auto_describe: "true"`, `github_action_config.auto_review: "true"`, `github_action_config.auto_improve: "false"` (must be explicit — see smoke test #2).
- Add a `.pr_agent.toml` scoping tone/focus to this repo (Go idioms, CLI UX, test coverage), plus a `pr_reviewer.extra_instructions` line telling the reviewer to disregard `cmd/assets/**` changes — PR-Agent has no native path-exclusion config (only PR title/branch/label/author/repo and file-extension/language-framework ignores), so a prompt-level instruction is the best available mitigation for the mixed-PR double-review gap (best-effort, not a hard guarantee).
- Smoke-test on a throwaway PR to confirm `/describe` and `/review` comments post correctly and the Plumber gate passes with the pinned digest.
- **Smoke test #1 (2026-07-05) caught a real bug**: closing/reopening PR #176 to fire the `reopened` trigger (the workflow only listens for `opened`/`reopened`/`ready_for_review`, not `synchronize`, so pushes since Phase 1 landed never actually ran it) showed the step completing with no comment posted. The container log showed `litellm.NotFoundError: GeminiException - models/gemini-1.5-flash is not found ... NOT_FOUND` on every retry — Google retired `gemini-1.5-flash`. The 404 (not a 401/403) confirmed the org-level `GEMINI_API_KEY` secret **is** reachable from this repo; only the model name was stale. Fixed by bumping to `gemini-3.5-flash` and adding Mistral/Cerebras fallbacks.
- **Smoke test #2 (2026-07-05) caught a second bug**: after the model fix, `/describe` and `/review` posted correctly — but so did a "PR Code Suggestions" (`/improve`) comment, despite Decision 2 explicitly deferring `/improve`. The container log showed `auto_improve=None` (the key was left unset, not `"false"`), and PR-Agent ran it anyway — leaving the flag unset does not mean disabled. Fixed by setting `github_action_config.auto_improve: "false"` explicitly.
- Exit criterion: a PR touching non-`cmd/assets/` files shows PR-Agent `/describe` and `/review` comments (and no `/improve` comment); a PR touching only `cmd/assets/**` shows none; Plumber CI passes; no `contents: write` requested anywhere in the new workflow. **Re-verify after the auto_improve fix** — re-run the reopen smoke test once this commit is pushed.

### Phase 2 — Observe

- **Owner: Thomas.**
- Run across real PRs for a fixed window: **2 weeks, no additional minimum-PR-count or numeric threshold** — whatever PR volume occurs in that window is the dataset for the Phase 3 decision (a deliberate choice to keep this lightweight rather than instrument thresholds up front; if the window turns out too quiet to judge anything, extend it rather than retrofit thresholds).
- Track: comment quality/signal, false positives, any overlap/noise on PRs that also touch `cmd/assets/**` (despite the Phase 1 prompt-instruction mitigation), and rate-limit throttling against Gemini's free-tier quota.
- Exit criterion: the 2-week window has elapsed; Thomas makes the Phase 3 call from whatever was observed.

### Phase 3 — Decide on next tier

- Based on Phase 2 observations: enable `/improve` and/or `/ask`, keep the current scope as-is, or retire the integration.
- If retained, write an ADR capturing the now-accepted decisions (mirroring `docs/ADR/adr-034-agent-scan-skill-security-scanning.md`'s pattern for `agent-scan`).
- Exit criterion: this plan's `status` flips to `done`; any resulting ADR reflects the final state.

## Risks

- **Free-tier rate limits** — Google AI Studio's free Gemini tier is rate-limited (requests/minute and/day); a burst of PR activity could throttle `/describe`/`/review` calls or delay comments. Mitigated by starting with only two lightweight tools enabled and tracking throttling in Phase 2; upgrading to a paid tier is the fallback if this repo's PR volume outgrows it.
- **Digest drift** — pinning by digest means picking up upstream fixes requires a manual re-pin; a forgotten digest can silently go stale. Mitigated by treating the digest as a recurring maintenance item, same as any other pinned action in this repo.
- **Docker-based action latency** — `uses: docker://...` steps pull an image on every run, unlike composite/JS actions; adds latency to every PR. Acceptable for an advisory step; revisit if it materially slows CI.
- **Review fatigue** — low-signal `/review` comments risk reviewers tuning the bot out entirely, the same alert-fatigue risk flagged in the `agent-scan` plan. Phase 2's observation window exists to catch this before Phase 3.
- **Mixed-path double review is not fully solved** — the Phase 1 `pr_reviewer.extra_instructions` line asking PR-Agent to disregard `cmd/assets/**` is a prompt-level instruction, not an enforced filter (PR-Agent has no native path-exclusion config). A PR mixing Go and skill-asset changes may still get commentary from both this bot and the `skill-quality.yml` judge on the asset portion. Accepted for the trial period; revisit only if Phase 2 shows it's actually noisy in practice.
- **No automated kill-switch** — if the key leaks, output turns abusive, or Gemini throttling degrades the experience badly, the only rollback is manually disabling the workflow from the Actions tab or reverting the merge commit. Accepted as sufficient for an advisory-only, low-traffic integration; matches how other advisory workflows in this repo (e.g. the Tessl review step) are handled.
- **Plumber does flag the docker step, but non-Critically — verified, not speculative.** A live Plumber run on the PR that introduced `pr-agent.yml` confirmed: the "container image pinning" control finds 0 container images (it doesn't classify `uses: docker://...` steps as images), but the "third-party actions must be pinned by commit SHA" control counts it as 1 of 25 action refs and marks it "Not Pinned By SHA" (a `sha256:` digest isn't a 40-char commit SHA). It's non-Critical, so `plumber-gate.sh` passes ("No Critical-severity Plumber findings. Gate passed."), but per this repo's ADR-038 process it will appear in the non-Critical rollup issue once merged to `main`. Decision 5's "no allowlist entry needed" call was correct (25/25 authorized without one), but its assumption that the docker ref is invisible to Plumber's pinning control was wrong — it's visible, just not blocking. Accepted as a known, tracked, non-Critical finding rather than something to fix.

## Verification

```bash
# After Phase 1 lands, confirm the workflow triggers and comments
gh pr view <pr-number> --json comments --jq '.comments[].body' | grep -iE "pr-agent|describe|review effort"

# Confirm a cmd/assets/-only PR does NOT trigger pr-agent.yml
gh run list --workflow=pr-agent.yml --json headBranch,event | jq .

# Confirm Plumber's actual disposition on the docker step (expect: non-Critical
# "Not Pinned By SHA" finding, gate still passes)
gh run view <plumber-run-id> --log | grep -i "Not Pinned By SHA\|Gate passed"
```

## Open Questions

Resolved via a 3-reviewer plan review (Technical/Strategic/Risk, all Claude Sonnet 5) plus a follow-up interview on 2026-07-05:

- **Phase 2 exit bar**: fixed 2-week window, no additional PR-count or numeric thresholds (deliberately kept lightweight).
- **Phase 2/3 owner**: Thomas.
- **Data egress to Google AI Studio**: self-certified as acceptable — this repo (skill-quality-auditor) is public, open-source Go tooling; diffs sent to Gemini are source code, docs, and CI config, never PLG participant or customer data.
- **`cmd/assets/**` mixed-PR double-review**: no native PR-Agent fix exists; mitigated with a prompt-level `extra_instructions` line (see Phase 1), accepted as best-effort rather than enforced.
- **Kill-switch**: manual disable via GitHub UI or commit revert is sufficient; no repo-variable toggle added.
- **Fork-PR trigger safety**: confirmed via `gh pr list` that this repo has had zero fork PRs historically (though `allow_forking: true`); the workflow already used the safe `pull_request` trigger (not `pull_request_target`), and a job-level fork guard was added in Phase 1 as defense-in-depth, mirroring `plumber.yml`'s existing pattern.
- **Plumber disposition on the docker step**: verified live against PR #176 (see Risks) rather than left as a speculative risk.
