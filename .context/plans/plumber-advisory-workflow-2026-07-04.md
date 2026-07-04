---
title: "Plan: Advisory-only Plumber CI/CD security workflow"
type: plan
status: draft
date: 2026-07-04
related:
  - ../findings/plumber-cicd-security-2026-07-04.md
---
# Plan: Advisory-only Plumber CI/CD security workflow

## Goal

Add a non-blocking `.github/workflows/plumber.yml` that runs the official `getplumber/plumber` GitHub Action on every pull request and on push to `main`, surfacing CI/CD security findings without gating merges. This is step 1 of the phased rollout recommended in `plumber-cicd-security-2026-07-04.md` — get real signal into PRs first, fix the pre-existing pinning/permissions debt second, promote to a blocking gate third.

## Scope

in_scope:

- New workflow file `.github/workflows/plumber.yml`, triggered on `pull_request` (not `pull_request_target` — fork PRs must get a restricted, read-only token) and `push: branches: [main]`.
- `score-push: false` — no public score badge, no public repo-name disclosure.
- Advisory framing: the Plumber step runs with `continue-on-error: true`, mirroring the Tessl review proving-period step already in `skill-quality.yml`.
- A minimal `permissions:` block on the new workflow: `contents: read` only (see Amendment 2 below — `security-events: write` is dropped for this phase).
- `timeout-minutes` on the job and a `concurrency` group keyed on ref, to bound CI-minute cost if the Action hangs or a branch gets duplicate runs.
- A named owner and review cadence for reading the advisory findings, plus a dated follow-up (issue or plan stub) that decides whether to proceed to the deferred pinning/permissions/blocking phases — without this, the advisory period has no forcing function to ever end.

out_of_scope (deferred to later phases of the parent finding):

- Pinning the five existing third-party actions (`golangci-lint-action`, `mise-action`, `release-please-action`, `goreleaser-action`, `setup-tessl`) by commit SHA.
- Adding `permissions:` blocks to `ci.yml` and `skill-quality.yml`.
- Extending `.plumber.yaml`'s `trustedGithubActions` allowlist.
- Committing or trimming the existing untracked `.plumber.yaml` default config.
- Enabling `score-push` / the public score badge.
- Promoting this check to a required/blocking gate or setting a `--threshold`.

## Phases

### Phase 1 — Author the workflow

Exit criterion: `.github/workflows/plumber.yml` exists on a feature branch, is syntactically valid, and is advisory-only (cannot fail the job).

Tasks:

- Create `.github/workflows/plumber.yml` on a new branch (`feat/plumber-advisory-workflow`), triggered on `pull_request` (confirmed not `pull_request_target`) and `push` to `main`. (wave: single)
- Set workflow-level `permissions: { contents: read }` — no `security-events: write` in this phase (Amendment 2: dropped because Phase 1 ships no SARIF-upload step, so the permission has no consumer). (wave: single)
- Add the `getplumber/plumber` step pinned to a commit SHA of the latest tagged release, not a floating version tag — this repo's own `.plumber.yaml` enforces SHA-pinning on every other third-party action, so exempting the checker itself would be inconsistent (Amendment 1). If no release has a resolvable SHA yet, pin to the tag and add an inline comment noting the exception and why. (wave: single)
- Mark the Plumber step `continue-on-error: true` so a low score or findings never fail the job. (wave: single)
- Add `timeout-minutes: 10` on the job and a `concurrency: { group: plumber-${{ github.ref }}, cancel-in-progress: true }` block, so a hung Action run or duplicate pushes cannot silently consume CI minutes indefinitely. (wave: single)
- Confirm whether `getplumber/plumber` reads `.plumber.yaml` from the checked-out working tree by default. Since that file is currently untracked, either commit a trimmed version of it in this phase, or explicitly note in the workflow (comment) that Phase 1 runs against Plumber's built-in defaults until it's committed — don't leave this ambiguous. (wave: single)

Dependencies: none.

### Phase 2 — Land and verify

Exit criterion: the workflow has run successfully (green, non-blocking) on both a real PR and a push to `main`, and its findings are visible in the run output.

Tasks:

- Open the PR from the Phase 1 branch (never commit straight to `main`, per this repo's working conventions).
- Confirm the workflow triggers on the PR event and completes without failing the job, even though the repo's known pinning/permissions gaps will produce real findings.
- Confirm the new check has not become an implicitly required status check via an existing branch-protection wildcard rule (e.g. "require all status checks to pass") — `continue-on-error` only makes the *step* tolerant; a broad branch-protection rule can still make the *job* required, silently turning "advisory" into "blocking" (Amendment 4).
- Read the Plumber output on that run, distinguish "ran clean, no findings" from "the Action itself failed to run" (the latter is masked by `continue-on-error` and would otherwise look identical), and record anything unexpected as a note in the PR description, including any network calls the Action makes even with `score-push: false`.
- Assign a named owner and review cadence for reading advisory findings going forward, and open a tracked follow-up (issue or plan stub) with a dated decision point for whether to proceed to the deferred pinning/permissions/blocking phases (Amendment 5) — mirroring the Tessl precedent's defined 2-week end state rather than leaving the advisory period open-ended.
- Merge once the rest of CI (`ci.yml`, `skill-quality.yml`) is green.

Dependencies: Phase 1.

## Risks

- The Action reference needs a concrete, resolvable release SHA at implementation time; if `getplumber/plumber` has no tagged release with a stable SHA, this plan's own pinning requirement (Amendment 1) cannot be fully satisfied and the exception must be documented, not silently dropped.
- `continue-on-error: true` makes the step non-blocking, but it also masks the difference between "Plumber ran and found nothing" and "Plumber crashed before producing output" — both look identical in the Actions UI. Phase 2 must check which one happened, not just that the job is green.
- If nobody reads the run output during the advisory period, the integration adds CI minutes without adding safety. Without a dated graduation-or-removal trigger (Amendment 5), this risk compounds indefinitely — unlike the Tessl proving-period precedent this plan mirrors, which has a defined 2-week end state.
- `getplumber/plumber`'s own supply-chain trustworthiness (maintainer history, release cadence, prior incidents) has not been independently vetted in this investigation. Given this repo belongs to a regulated organisation (PLG), adding an unvetted third-party Action to every PR's CI run may warrant a lightweight Security/Compliance check-in before merge, separate from the later decision to promote this check to blocking — this is a routing question, not one this plan resolves on its own.
- A broad branch-protection rule (e.g. "require all status checks") could make this new job an implicitly required check despite `continue-on-error`, converting "advisory" into "blocking" without anyone deciding that. See the new Phase 2 verification task.

## Verification

- Validate the new YAML as an explicit Phase 1 task, not a post-hoc check: run `actionlint .github/workflows/plumber.yml` if available; otherwise the PR's own Actions run is the first real validation (note this is a weaker pre-merge guarantee and accept it as such rather than treating it as equivalent).
- Confirm the workflow appears in the Actions tab and completes with a green check on the PR, regardless of Plumber's findings.
- Confirm the workflow's `permissions:` block is exactly `{ contents: read }` — no broader scope was introduced.
- Confirm `score-push` did not publish anything to `score.getplumber.io` (check the job log for the score-push step's outcome).
- Confirm the trigger is `pull_request`, not `pull_request_target`, so fork PRs run with a restricted token.
- Confirm the new check does not appear as a required status check under the repo's branch-protection settings unless that was an explicit, separate decision.

## Open Questions

- Should this repo commit a trimmed `.plumber.yaml` as part of this phase (currently ~1000 lines, mostly inapplicable GitLab controls), or is it acceptable for Phase 1 to run Plumber against its built-in defaults for now? This determines whether the config work belongs in this plan or stays deferred.
- Does adding an unvetted third-party Action to CI warrant a lightweight Security/Compliance check-in given this org's regulatory posture, and if so, who owns making that call before Phase 2 merges?
- Who is the named owner for reading advisory findings on an ongoing basis, and what date should the follow-up plan (pinning, permissions, threshold, promotion-or-removal decision) be scheduled? Phase 2 now requires opening a tracked follow-up as an exit condition, but the owner and date are still unassigned at plan-approval time.

## Critical Review Findings (2026-07-04)

Reviewed by 3 independent Claude Sonnet 5 subagents (Technical, Strategic, Risk) via the `plan-review` skill, using an identical self-contained brief. All three converged independently on the same core issue: the plan's original Open Questions were being deferred into implementation rather than resolved before it, and the advisory period had no defined end state — unlike the Tessl proving-period precedent it mirrors. The Risk reviewer additionally flagged that this repo belongs to a regulated organisation, which the Technical and Strategic reviews did not weigh.

Amendments applied above, numbered for traceability:

1. **Action pinning resolved, not deferred** — pin `getplumber/plumber` by commit SHA in Phase 1 itself, consistent with the pinning hygiene this repo already expects of every other third-party action (Technical, Strategic).
2. **`security-events: write` dropped from this phase** — Phase 1 ships no SARIF-upload step, so the permission had no consumer; requesting it "on spec" contradicted the plan's own minimal-permissions framing (all three reviewers).
3. **Fork-PR and cost controls made explicit** — confirmed trigger is `pull_request` not `pull_request_target`; added `timeout-minutes` and a `concurrency` group so a hung or duplicated run cannot consume CI minutes unbounded (Risk, Technical).
4. **Branch-protection check added to Phase 2** — `continue-on-error` only protects the step; a broad branch-protection rule can still mark the job required, silently converting advisory into blocking (Risk, Technical).
5. **Ownership and a graduation/removal trigger made a Phase 2 exit condition** — the single finding all three reviewers raised independently: without a named owner, a review cadence, and a dated follow-up decision, the advisory integration risks running forever with no one reading it, unlike the Tessl precedent's defined 2-week end state (Technical, Strategic, Risk).
6. **`.plumber.yaml` discovery clarified as a Phase 1 task** — the file is currently untracked; if the Action can't see it in CI, Phase 1 silently runs against built-in defaults instead of the intended config (Technical).
7. **Regulated-org routing surfaced as an Open Question, not resolved by this plan** — the Risk reviewer noted that adding an unvetted third-party Action to CI may warrant Security/Compliance input given this org's posture; this plan does not make that call, it only ensures the question is asked before Phase 2 merges.

Not adopted: the Strategic reviewer's suggestion to formalise open questions 1–2 as a separate "Phase 0" was folded directly into Phase 1's task list instead, since both were resolvable now rather than requiring separate sequencing.
