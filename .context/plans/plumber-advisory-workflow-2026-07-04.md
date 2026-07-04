---
title: "Plan: Plumber CI/CD security workflow — fail on Critical, track the rest as issues"
type: plan
status: draft
date: 2026-07-04
related:
  - ../findings/plumber-cicd-security-2026-07-04.md
  - ../../docs/ADR/adr-036-plumber-cicd-security-advisory-workflow.md
  - ../../docs/ADR/adr-037-plumber-critical-fail-issue-tracking.md
---
# Plan: Plumber CI/CD security workflow — fail on Critical, track the rest as issues

## Goal

Add `.github/workflows/plumber.yml` that runs the official `getplumber/plumber` GitHub Action on every pull request and push to `main`, hard-failing the job when any Critical-severity finding is present and filing a tracked GitHub issue for every High, Medium, or Low finding instead of blocking on them. This supersedes the advisory-only design recorded in `adr-036`, per explicit directive: (1) `.plumber.yaml` will need configuring for this repo regardless of timing, so that work is folded into this phase rather than deferred; (2) this repo is not in PLG's regulated scope, so no compliance check-in gates this decision; (3) the integration should fail the pipeline outright on the severity that matters and turn everything else into tracked, non-blocking work rather than advisory output nobody reads.

## Scope

in_scope:

- New workflow file `.github/workflows/plumber.yml`, triggered on `pull_request` (not `pull_request_target` — fork PRs must get a restricted, read-only token) and `push: branches: [main]`.
- `score-push: false` — no public score badge, no public repo-name disclosure. Unchanged; no directive to revisit this.
- Trim `.plumber.yaml` to the `github:` controls this repo actually needs (drop the `gitlab:` block entirely — this repo has no GitLab CI) and commit it as part of this phase, so the Action runs against the intended config from day one instead of built-in defaults.
- `permissions: { contents: read, issues: write }` — `issues: write` is new, required for the issue-filing step. No `security-events: write` — this design has no SARIF-upload step.
- The `getplumber/plumber` step pinned to a commit SHA of the latest tagged release (not a floating tag or `@latest`), consistent with the pinning hygiene this repo's own `.plumber.yaml` enforces on every other third-party action.
- `timeout-minutes` on the job and a `concurrency` group keyed on ref, to bound CI-minute cost.
- Structured output via `plumber analyze --output results.json` — not relying on Plumber's own exit code, which reflects an overall A-E score against `--threshold`, not per-finding severity.
- A gate step that parses `results.json` and fails the job (non-zero exit, no `continue-on-error`) if any **Critical**-severity finding is present.
- An issue-filing step, run via `if: always()` so it still runs even when the gate step fails, that files a deduplicated GitHub issue for every **High, Medium, or Low** finding not already tracked.

out_of_scope:

- Pinning the five pre-existing third-party actions (`golangci-lint-action`, `mise-action`, `release-please-action`, `goreleaser-action`, `setup-tessl`) by commit SHA — stays a separate follow-up, unless Phase 1's dry-run (below) shows one of them currently produces a Critical finding, in which case fixing it must be pulled forward so this PR doesn't block itself.
- Adding `permissions:` blocks to `ci.yml` and `skill-quality.yml`.
- Enabling `score-push` / the public score badge.
- SARIF upload to GitHub Code Scanning (no `security-events: write`, no consumer step in this design).

## Phases

### Phase 1 — Author the workflow and configuration

Exit criterion: `.github/workflows/plumber.yml` and a trimmed, committed `.plumber.yaml` exist on a feature branch; the workflow is syntactically valid; a dry run shows what it will actually gate on before it ever runs against a real PR.

Tasks:

- Trim `.plumber.yaml` to the `github:` controls block only (drop `gitlab:` entirely) and commit it. (wave: single)
- Create `.github/workflows/plumber.yml`, triggered on `pull_request` (confirmed not `pull_request_target`) and `push` to `main`. (wave: single)
- Set `permissions: { contents: read, issues: write }`. (wave: single)
- Add the `getplumber/plumber` step pinned to a commit SHA of the latest tagged release; if none is resolvable, pin to the tag and add an inline comment documenting the exception. (wave: single)
- Add `timeout-minutes: 10` and a `concurrency: { group: plumber-${{ github.ref }}, cancel-in-progress: true }` block. (wave: single)
- Run `plumber analyze --output results.json` with `continue-on-error: true` on this step specifically — its own exit code is score-threshold-based, not severity-based, and must never be what decides pass/fail here. (wave: single)
- Add a gate step (no `continue-on-error`) that parses `results.json`, counts findings with `severity: Critical` (verify the exact field name and value casing against a real run's output before finalizing — not yet confirmed against live JSON), and exits non-zero if the count is greater than zero. (wave: single)
- Add an issue-filing step (`if: always()`) that reads every High/Medium/Low finding from `results.json` and creates a GitHub issue for each one not already open, deduplicated by a stable key (e.g. Plumber's issue code plus the affected file/job path) — check for an existing open issue with that key (e.g. via `gh issue list --search`) before creating a new one. (wave: single)
- **Before opening the PR**, dry-run `plumber analyze` against current `main` (locally, or via `workflow_dispatch` on the feature branch) and inspect the JSON output for any Critical findings among the repo's known pre-existing gaps (five unpinned actions, two workflows missing `permissions:` blocks). If any are Critical-severity, this PR cannot merge under its own new gate until they're fixed — pull that fix forward into this phase rather than deferring it. (wave: single, but sequenced before Phase 2 — do this before opening the PR, not after)

Dependencies: none.

### Phase 2 — Land and verify

Exit criterion: the workflow runs on a real PR, fails only when a Critical finding is present, files (and does not duplicate) issues for every other finding, and the PR merges once that behaviour is confirmed.

Tasks:

- Open the PR from the Phase 1 branch (never commit straight to `main`, per this repo's working conventions).
- Confirm the gate step's pass/fail behaviour matches the Phase 1 dry run — no surprises between the dry run and the real PR run.
- Confirm the issue-filing step creates exactly one issue per unique finding (re-run the workflow, e.g. by pushing an empty commit, and confirm no duplicate issues appear for findings already filed).
- Confirm the new check has not become an implicitly required status check via an unrelated branch-protection wildcard rule beyond the intended "fail on Critical" behaviour.
- Confirm the trigger is `pull_request`, not `pull_request_target`. Note that fork PRs will run with a restricted token and likely cannot file issues even from the non-blocking step — confirm this degrades cleanly (a clear log message) rather than as a confusing API error, since it only affects the issue-filing step, not the Critical gate.
- Merge once the rest of CI (`ci.yml`, `skill-quality.yml`) is green and the above are confirmed.

Dependencies: Phase 1.

## Risks

- **This PR could block itself.** The new gate fails on any Critical finding, and this repo has known pre-existing gaps (unpinned actions, missing permissions blocks) whose severity hasn't been confirmed against Plumber's actual taxonomy. Phase 1's dry-run task exists specifically to catch this before it becomes a live blocker — treat that task as mandatory, not optional.
- **Severity field assumptions are unverified.** This plan assumes `results.json` exposes a `severity` field with a `Critical` value distinguishable from `High`/`Medium`/`Low`, based on the aggregate categories shown on Plumber's public Radar page. The exact JSON schema has not been inspected directly — confirm against a real `results.json` before finalizing the gate-parsing script, and adjust field names/casing as needed.
- **Issue-filing dedup is custom-built, not a Plumber feature.** Plumber has no native "file a GitHub issue" output — this requires a hand-written dedup step. A dedup key that's too loose spams the repo with duplicates on every run; one that's too strict silently stops filing new issues after the first match.
- **`issues: write` is a new, broader permission** than anything else this workflow needs — scope it to the job or step that uses it rather than the whole workflow if the Action step itself doesn't need it.
- The `getplumber/plumber` action reference needs a concrete, resolvable release SHA at implementation time; if none exists, the pinning requirement can't be fully satisfied and the exception must be documented, not silently dropped.
- Fork PRs run with a restricted `GITHUB_TOKEN` and likely cannot file issues even from the non-blocking step — confirm this fails clearly rather than silently.

## Verification

- Validate the new YAML with `actionlint .github/workflows/plumber.yml` if available.
- Confirm `.plumber.yaml` is committed, trimmed to the `github:` block only, and is the config the Action actually reads (not a built-in default).
- Confirm the workflow's `permissions:` block is exactly `{ contents: read, issues: write }` (or narrower, per the risk above) — no broader scope introduced.
- Confirm `score-push` did not publish anything to `score.getplumber.io`.
- Confirm the gate step fails the job if and only if a Critical finding is present.
- Confirm the issue-filing step does not create duplicate issues across repeated runs.

## Open Questions

- What is the exact field name and value set Plumber's JSON output uses for severity (confirm `severity: Critical|High|Medium|Low` or equivalent) — needed before the gate-parsing script can be written precisely. **Resolved during Phase 1 implementation — see "Implementation Notes" below.**
- Does Plumber support a native baseline/suppression mechanism for pre-existing findings (so day-one debt doesn't need fixing before the gate can be enabled), or is Phase 1's "fix Critical findings first" dry-run the only option? Worth checking `docs/scoring.md` / `plumber explain` before assuming no such feature exists. **Resolved: no such feature exists** (confirmed against `getplumber/plumber`'s `docs/scoring.md` and `README.md` — no `baseline`/`suppress`/`ignore` mechanism). The Phase 1 dry-run was the only option, and it found zero pre-existing Critical findings, so this repo never needed one.
- Who triages the GitHub issues this workflow files for High/Medium/Low findings, and on what cadence? The fail-on-Critical gate is self-enforcing, but the issue backlog it creates needs an owner or it becomes as unread as the advisory design it replaces. **Still open** — a human/team decision, not resolved by this implementation.

## Implementation Notes (2026-07-04, Phase 1)

The plan's central open question — Plumber's JSON severity schema — turned out to be more involved than "confirm the field name." Read directly from `getplumber/plumber`'s source (`internal/engine/opa/engine.go`, `control/scoring.go`, `control/types.go`) rather than relying on stale example files in that repo (`output-example.json`, `reports/*.json` are older-format and do not reflect current output):

- The top-level `findings[]` array (which does carry a per-finding `severity` field, lowercase `critical|high|medium|low`, via a custom `MarshalJSON`) is **not populated for GitHub-side controls** in the current release (v0.3.86) — a scaffold comment in the source confirms GitHub findings aren't wired into it yet. A real run against this repo confirmed `findings: []` while `plumberScore.counts` showed real numbers.
- Actual per-instance issue data lives scattered across ~20 different `<control>Result.issues[]` arrays (`actionPinningResult`, `permissionsResult`, `branchProtectionResult`, etc.), each with a different shape — no shared `severity` field per issue, no shared location field name (`jobName` vs `job` vs `branchName`).
- Per-code severity **is** available, in `plumberScore.codeLosses[]` (`{code, severity, count, ...}`), aggregated across whatever codes fired in that run. `plumberScore.counts.critical` gives the simple boolean-ish gate check directly, with no parsing of individual controls needed.
- `scripts/plumber-gate.sh` therefore gates on `plumberScore.counts.critical`, not a `findings[]` scan.
- `scripts/plumber-file-issues.sh` walks every `*Result` key generically (`to_entries[] | select(.key | endswith("Result"))`), joins each `issues[]` entry back to its severity via the `codeLosses` code map, and dedupes on `sha256(resultKey|code|canonicalized-issue-json)` rather than a hand-picked field, since no field is universal across controls.
- The official `getplumber/plumber` GitHub Action (`action.yml`) already wraps the CLI in a composite action with its own `soft-fail` / `threshold` inputs and an internal "Enforce threshold" step. Passing `threshold: '0'` and `soft-fail: true` makes that internal step a no-op for compliance purposes, so the plan's original `continue-on-error: true` on a raw `plumber analyze` step wasn't necessary — the Action step now only fails the job on a genuine Plumber runtime/config error, which should legitimately fail the job rather than be swallowed.
- The Phase 1 mandatory dry-run (`plumber analyze --output results.json --threshold 0` against local `main`) found **zero Critical findings**: 11 High (5× `ISSUE-701` unpinned actions, 5× `ISSUE-713` unauthorized sources on the same five actions, 1× `ISSUE-505` branch protection) and 4 Medium (`ISSUE-801` missing `permissions:` blocks). This PR does not block itself under its own new gate.
- Adding `getplumber/plumber` itself to `.plumber.yaml`'s `trustedGithubActions` allowlist was pulled into Phase 1 (not deferred with the other four pre-existing unpinned actions): without it, the new workflow immediately flags its own Action as an unauthorized source, which would refile the same issue forever. This is a one-line, obviously-correct fix caused directly by this change, distinct from the pre-existing five-action pinning debt this plan explicitly defers.

## Critical Review Findings (2026-07-04)

Reviewed by 3 independent Claude Sonnet 5 subagents (Technical, Strategic, Risk) via the `plan-review` skill, against the original advisory-only design. All three converged independently on the same core issue: that design's Open Questions were being deferred into implementation rather than resolved before it, and its advisory period had no defined end state. The Risk reviewer additionally flagged this repo's (then-assumed) regulated-org posture, which the Technical and Strategic reviews did not weigh.

Amendments applied to the advisory-only design at the time (see "Amendments (2026-07-04, second pass)" below for what has since changed):

1. **Action pinning resolved, not deferred** — pin `getplumber/plumber` by commit SHA in Phase 1 itself (Technical, Strategic). *Retained in the current design.*
2. **`security-events: write` dropped** — no SARIF-upload consumer (all three reviewers). *Retained.*
3. **Fork-PR and cost controls made explicit** — `pull_request` not `pull_request_target`; `timeout-minutes`; `concurrency` group (Risk, Technical). *Retained.*
4. **Branch-protection check added to Phase 2** (Risk, Technical). *Retained, reframed since the gate is now intentionally blocking for Critical findings by design.*
5. **Ownership and a graduation/removal trigger made a Phase 2 exit condition** (Technical, Strategic, Risk). *Superseded — the Critical gate is now self-enforcing and permanent by design; the ownership question moves to the issue backlog instead (see Open Questions).*
6. **`.plumber.yaml` discovery clarified** (Technical). *Superseded — resolved outright: trim and commit it in Phase 1.*
7. **Regulated-org routing surfaced as an Open Question** (Risk). *Resolved — this repo is confirmed out of PLG's regulated scope; no compliance check-in required.*

## Amendments (2026-07-04, second pass — user directives)

After the critical review above, the user resolved all three of its remaining Open Questions directly, and one of those answers changed the design's core mechanism:

1. **`.plumber.yaml` timing:** "We'll need to configure plumber so either way is fine" — resolved by folding the trim-and-commit work into Phase 1 rather than deferring it, since the timing was explicitly a non-issue.
2. **Regulated-org routing:** "This isn't a PLG repo, so we are going to use Plumber" — resolved. The Security/Compliance check-in flagged by the Risk reviewer and recorded as an open item in `adr-036` does not apply to this repository; no further routing needed.
3. **Gating design:** "We need to fail the pipeline on failure, and anything lower needs to be an issue" — clarified via follow-up: **Critical-severity findings fail the pipeline; High, Medium, and Low findings are filed as tracked GitHub issues instead of blocking.** This replaces the advisory-only, `continue-on-error` design outright — a reversal of the original central decision (non-blocking → blocking-for-Critical), not an extension of it. Captured as a new decision, `adr-037`, which supersedes `adr-036` rather than amending it (ADRs are immutable once created).

This second pass introduces implementation surface the first review never saw: JSON-output parsing for severity, a custom issue-filing/dedup step, and a new `issues: write` permission. Flagged here for visibility, not as a blocker — the user's directives are explicit, and what remains is implementation detail (exact JSON field names, dedup key design) rather than an open design question.
