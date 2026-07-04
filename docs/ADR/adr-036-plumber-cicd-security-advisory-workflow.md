---
title: "ADR-036: Advisory-only Plumber CI/CD security workflow"
status: superseded
date: 2026-07-04
superseded_by: "adr-037"
context:
  - path: ".context/findings/plumber-cicd-security-2026-07-04.md"
  - path: ".context/plans/plumber-advisory-workflow-2026-07-04.md"
---

**Status:** Proposed
**Date:** 2026-07-04

## Context

This repo has never run a CI/CD-pipeline-specific security scan — existing gates (`golangci-lint`, `go vet`, `gofmt`, `markdownlint`, `shellcheck`, and the skill-quality gates) check code and skill-asset quality, not the `.github/workflows/*.yml` files themselves. `.context/findings/plumber-cicd-security-2026-07-04.md` confirmed `getplumber/plumber` is a real, actively maintained open-source tool for exactly this gap: it audits committed workflow YAML for OWASP CICD-SEC issues (unpinned actions, missing least-privilege permissions, dangerous triggers, unverified script execution) and scores the result. An untracked `.plumber.yaml` — the tool's own generated default config — was already present in this repo's working tree, unconnected to any CI job.

The finding also established that a first real run would immediately surface pre-existing debt: none of this repo's five third-party actions (`golangci-lint-action`, `mise-action`, `release-please-action`, `goreleaser-action`, `setup-tessl`) are pinned by commit SHA, and two of four workflows (`ci.yml`, `skill-quality.yml`) have no `permissions:` block. `.context/plans/plumber-advisory-workflow-2026-07-04.md` lays out a phased, advisory-first rollout for step 1 (add the check, non-blocking) and was independently reviewed by three subagents (Technical, Strategic, Risk) before being amended. This ADR captures the binding decisions from that reviewed plan.

This is a distinct layer from `docs/ADR/adr-034-agent-scan-skill-security-scanning.md`: agent-scan inspects skill *content* (`cmd/assets/`); Plumber inspects *pipeline configuration* (`.github/workflows/`). The two are complementary, not overlapping.

## Decision

1. **Integrate as an advisory-only step, not a merge-blocking gate, for an initial observation window.** `continue-on-error: true` on the Plumber step; `score-push: false` (no public score badge or repo-name disclosure). This mirrors the existing Tessl review step's proving-period pattern already in `skill-quality.yml`.
2. **A new, dedicated `.github/workflows/plumber.yml`, not folded into an existing job.** Unlike agent-scan (scoped to `cmd/assets/**`), Plumber's checks apply to every workflow file in the repo, so it runs on its own trigger (`pull_request`, `push` to `main`) rather than being appended to `skill-quality.yml`'s `cmd/assets/**`-gated job.
3. **The `getplumber/plumber` action reference itself is pinned by commit SHA, not a floating version tag**, for consistency with the exact pinning control this tool enforces on every other action in the repo. Exempting the checker from its own rule was flagged as inconsistent by all three plan reviewers.
4. **Permissions are `contents: read` only for this phase.** `security-events: write` is deliberately deferred — Phase 1 ships no SARIF-upload step, so granting it now would be unused scope with no consumer.
5. **`timeout-minutes` and a `concurrency` group are set on the job** to bound CI-minute cost if the Action hangs or a branch produces duplicate runs.
6. **Pre-existing pinning/permissions debt is explicitly not fixed as part of this phase.** The Action will surface real findings against `ci.yml`, `skill-quality.yml`, and the five unpinned actions immediately — this is accepted, expected signal, not a blocker to landing the advisory workflow. Fixing that debt (pinning, permissions blocks, `.plumber.yaml` trust-list extension) is deferred to a follow-up phase of the parent finding.
7. **A named owner, a review cadence, and a dated graduation-or-removal decision are required exit conditions before Phase 2 merges** — not left open-ended. All three plan reviewers independently identified this as the plan's most significant gap: without it, the advisory period has no forcing function to end, unlike the Tessl precedent's defined 2-week window.
8. **Two items remain open, recorded here but not resolved by this ADR:** whether to commit a trimmed `.plumber.yaml` in this phase or defer it, and whether adding an unvetted third-party Action to CI warrants a lightweight Security/Compliance check-in given this org's regulated posture. Per this repo's `adr-034` precedent for unresolved human sign-off items, this ADR records that both questions must be answered before Phase 2 merges — it does not answer them itself.

## Consequences

- **Easier:** this repo's CI/CD pipeline configuration gets automated, PR-visible security coverage for the first time — a gap that existed across all four existing workflows with no prior tooling.
- **Easier:** because the check is advisory and non-blocking, it can ship immediately without first clearing the pre-existing pinning/permissions debt it will report on.
- **Harder:** a new, not-yet-vetted third-party Action is added to CI. SHA-pinning it (Decision 3) mitigates the mutable-reference risk but does not itself establish the action's trustworthiness — that remains an open question this ADR defers to a human/org decision (Decision 8).
- **Harder:** without Decision 7's exit condition, the advisory workflow could otherwise run indefinitely, consuming CI minutes while never being acted on — this ADR treats naming an owner and a dated review as mandatory, not optional, before the workflow ships.
