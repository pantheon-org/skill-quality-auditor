---
title: "Draft Plan: Integrate PR-Agent as an Advisory PR Review Bot"
type: plan
status: draft
date: 2026-07-04
related:
  - ../findings/pr-agent-integration-2026-07-04.md
  - ../../.github/workflows/skill-quality.yml
  - ../../.plumber.yaml
  - ../instructions/ways-of-working.md
---

## Goal

Add `The-PR-Agent/pr-agent` as an advisory GitHub Action that posts `/describe` and `/review` comments on pull requests, reusing the existing `ANTHROPIC_API_KEY` secret and this repo's Plumber pinning/permissions posture, without gating merge. `/improve` and `/ask` are deferred until a trial period shows the lighter-weight tools are worth the added noise. See `.context/findings/pr-agent-integration-2026-07-04.md` for the feasibility research this plan is based on.

## Scope

**In scope:**

- A new workflow file, `.github/workflows/pr-agent.yml`, separate from `skill-quality.yml`.
- `uses: docker://pragent/pr-agent@sha256:<pinned-digest>` — digest-pinned, not the floating `the-pr-agent/pr-agent@main` ref shown in the project's own quick-start — with `gh attestation verify` run once to confirm the digest before use.
- `the-pr-agent/pr-agent` (or the Docker Hub image owner) added to `.plumber.yaml`'s `trustedGithubActions`.
- `ANTHROPIC_API_KEY` reused via `config.model: "anthropic/claude-<model>"` + `ANTHROPIC.KEY: ${{ secrets.ANTHROPIC_API_KEY }}` — no new secret.
- Only `github_action_config.auto_describe: "true"` and `github_action_config.auto_review: "true"` enabled.
- `permissions: { issues: write, pull-requests: write }` only — no `contents: write`.
- A `.pr_agent.toml` (or inline `extra_instructions`) tuned to this repo (Go idioms, CLI UX, test coverage) rather than defaults.

**Out of scope (deferred):**

- `/improve` (inline code-suggestion comments) — deferred pending a trial-period read on signal-to-noise, per the finding's recommendation for a small, already-disciplined Go codebase.
- `/ask` — reactive, comment-triggered tool; adds no new infra decisions, so it can be turned on later without revisiting this plan.
- A dedicated PR-Agent LLM budget/secret separate from `ANTHROPIC_API_KEY` — only worth doing if Phase 2 usage data shows cost or quota contention.
- GitHub App / webhook deployment — the Action-based mode is consistent with this repo's existing CI patterns (Plumber, agent-scan, skill-quality) and needs no separate app installation.
- `/update_changelog` — not requested; avoids needing `contents: write`.
- Any workflow-file edits in this PR — this plan documents intent for review; implementation is a separate PR once the plan (and any resulting ADR) is accepted.

## Decisions (proposed — pending review)

1. **Advisory-only, no blocking gate.** PR-Agent's comments are informational; no CI status check is made required to merge. Justification: matches the proving-period pattern already established for the Tessl review step and the `agent-scan` plan; LLM review comments are inherently subjective and shouldn't block merge on their own.
2. **Start with `/describe` + `/review` only.** `/improve` and `/ask` are deferred to a later phase. Justification: per the finding, noisy inline suggestion comments risk adding review friction rather than reducing it on this codebase.
3. **Reuse `ANTHROPIC_API_KEY`, don't provision a new secret.** Justification: cheaper, one fewer credential to manage; revisit only if Phase 2 shows cost/quota contention with the nightly LLM-judge in `skill-quality.yml`.
4. **Pin via Docker digest with attestation verification, not a floating branch ref.** Justification: this repo's Plumber policy (`actionsMustBePinnedByCommitSha`) would flag `the-pr-agent/pr-agent@main`; the project's own docs offer the digest-pinned form as the supported alternative.
5. **Add `the-pr-agent/pr-agent` to `.plumber.yaml`'s `trustedGithubActions`.** Justification: third-party, non-same-org action — same precedent as adding `getplumber/plumber` itself.
6. **Permissions scoped to `issues: write` + `pull-requests: write` only.** Justification: `/update_changelog` isn't enabled in this rollout, so `contents: write` isn't needed — matches this repo's least-privilege posture.
7. **Own workflow file, not folded into `skill-quality.yml`.** Justification: distinct trigger semantics (all PRs vs. path-scoped skill assets) and a distinct concern (general code review vs. skill-content scoring) — mirrors the existing one-workflow-per-concern pattern (`ci.yml`, `plumber.yml`, `skill-quality.yml`).
8. **Exact mechanism for excluding/de-duplicating commentary on `cmd/assets/**` diffs is deferred to Phase 1 implementation**, not decided up front — see Open Questions.

## Phases

### Phase 0 — Prerequisites (human action, blocks Phase 1)

- Confirm `ANTHROPIC_API_KEY` reuse is acceptable on cost/quota grounds, or that a separate key is preferred (Decision 3).
- Confirm the `/improve`/`/ask` deferral (Decision 2) — a product-taste call, not an agent decision.
- Resolve and record the Docker Hub digest to pin (`pragent/pr-agent@sha256:...`), verified via `gh attestation verify`.
- Exit criterion: digest confirmed and recorded; secret-reuse and scope decisions recorded (e.g. as a PR comment or ADR).

### Phase 1 — Wire in the advisory workflow

- Add `the-pr-agent/pr-agent` to `.plumber.yaml`'s `trustedGithubActions`.
- Create `.github/workflows/pr-agent.yml`: trigger on `pull_request` (`opened`, `reopened`, `ready_for_review`) + `issue_comment`; `permissions: { issues: write, pull-requests: write }`; `uses: docker://pragent/pr-agent@sha256:<pinned-digest>`; env `GITHUB_TOKEN`, `config.model`, `ANTHROPIC.KEY: ${{ secrets.ANTHROPIC_API_KEY }}`, `github_action_config.auto_describe: "true"`, `github_action_config.auto_review: "true"` (`auto_improve` left unset).
- Add a `.pr_agent.toml` scoping tone/focus to this repo and resolving the `cmd/assets/**` exclusion mechanism from the Open Questions below.
- Smoke-test on a throwaway PR to confirm `/describe` and `/review` comments post correctly and the Plumber gate passes with the pinned digest.
- Exit criterion: a PR touching non-`cmd/assets/` files shows PR-Agent `/describe` and `/review` comments; Plumber CI passes; no `contents: write` requested anywhere in the new workflow.

### Phase 2 — Observe

- Run across real PRs for a fixed window (proposed: 2 weeks, mirroring the proving-period language already used for Tessl review and `agent-scan`).
- Track: comment quality/signal, false positives, any overlap/noise on PRs that also touch `cmd/assets/**`, and usage/cost against the shared `ANTHROPIC_API_KEY` budget.
- Exit criterion: enough observed runs to decide on enabling `/improve`/`/ask`, adjusting scope, or dropping the integration.

### Phase 3 — Decide on next tier

- Based on Phase 2 observations: enable `/improve` and/or `/ask`, keep the current scope as-is, or retire the integration.
- If retained, write an ADR capturing the now-accepted decisions (mirroring `docs/ADR/adr-034-agent-scan-skill-security-scanning.md`'s pattern for `agent-scan`).
- Exit criterion: this plan's `status` flips to `done`; any resulting ADR reflects the final state.

## Risks

- **Duplicate/noisy commentary on `cmd/assets/**` PRs** — two LLM opinions (the `skill-quality.yml` judge and PR-Agent's `/review`) on the same diff. Mitigated by resolving an exclusion mechanism in Phase 1 before merge.
- **Shared secret contention/cost** — PR-Agent draws on the same `ANTHROPIC_API_KEY` quota as the nightly LLM-judge; a spike in PR volume could raise costs or hit rate limits for both consumers. Mitigated by Phase 2 cost tracking; revisit secret-splitting if needed.
- **Digest drift** — pinning by digest means picking up upstream fixes requires a manual re-pin; a forgotten digest can silently go stale. Mitigated by treating the digest as a recurring maintenance item, same as any other pinned action in this repo.
- **Docker-based action latency** — `uses: docker://...` steps pull an image on every run, unlike composite/JS actions; adds latency to every PR. Acceptable for an advisory step; revisit if it materially slows CI.
- **Review fatigue** — low-signal `/review` comments risk reviewers tuning the bot out entirely, the same alert-fatigue risk flagged in the `agent-scan` plan. Phase 2's observation window exists to catch this before Phase 3.

## Verification

```bash
# Confirm the pinned digest before wiring in (Phase 0)
gh attestation verify "oci://index.docker.io/pragent/pr-agent@sha256:<digest>" --repo The-PR-Agent/pr-agent

# After Phase 1 lands, confirm the workflow triggers and comments
gh pr view <pr-number> --json comments --jq '.comments[].body' | grep -iE "pr-agent|describe|review effort"
```

## Open Questions

- What mechanism excludes or de-duplicates commentary on `cmd/assets/**` diffs — a workflow-level path filter that skips the job when a PR touches only `cmd/assets/**`, a `.pr_agent.toml` ignore setting, or accepting the overlap and letting Phase 2 judge whether it's actually noisy in practice?
- Should this plan get its own ADR now (mirroring ADR-034 for `agent-scan`), or only once Phase 0's human decisions are actually made? Several decisions above are proposed, not yet accepted — flagged for reviewer judgment rather than assumed.
