---
title: "Finding: Cross-Reference Status Drift Across Plans, Findings, and ADRs"
type: FINDING
status: ACTIVE
date: 2026-07-04
value: LOW
themes:
  - GOVERNANCE
related:
  - ../plans/post-merge-status-sync-2026-07-04.md
  - ../findings/automate-post-merge-status-sync-2026-07-04.md
  - ../plugins/pantheon-org/governance/adr-capture/scripts/merge-status-sync.sh
  - ../instructions/ways-of-working.md
  - ../../.context/index.yaml
  - ../../docs/ADR/index.yaml
---
# Finding: Cross-Reference Status Drift Across Plans, Findings, and ADRs

> A full audit of `.context/index.yaml` and `docs/ADR/index.yaml` against shipped code and workflows found 8 items with stale status — not because a PR forgot to flip its own plan, but because a *later* PR's plan superseded an *older* plan/finding/ADR and nobody went back to close the loop on the older one.

## Summary

`automate-post-merge-status-sync-2026-07-04.md` (merged, PR #173) proposes wiring `merge-status-sync.sh` to a merge-triggered GitHub Action. That script resolves candidates from the *merging PR's own* frontmatter/file-touch links and auto-flips single-phase plans it can trace directly to that PR. It is a correct, well-scoped fix for one drift pattern. This finding documents a second, distinct pattern that pattern-1 tooling structurally cannot catch, evidenced by 8 concrete items found this session.

## Detail — two distinct drift patterns

**Pattern 1 (already has a fix in flight):** a PR merges and its own linked plan/ADR isn't flipped. `merge-status-sync.sh` solves this today (manually) and is proposed for GH Action automation.

**Pattern 2 (this finding, not yet addressed by any existing tooling):** item A (a draft plan, active finding, or proposed ADR) gets superseded by a *different*, later plan B that actually implements the decision and is marked done — but A is never touched, because nothing about merging B's PR points back at A. `merge-status-sync.sh` resolves candidates from the merging PR's own files/links; it has no mechanism to notice that PR B's done plan cited PR A's now-stale finding as a predecessor it fulfilled.

## Concrete evidence found this session

1. `.context/plans/migrate-off-tessl-eval-2026-06-29.md` — status: `draft`, should be `done`. Its recommendation (native Go eval runner, Option A) was implemented by `.context/plans/native-eval-runner-2026-07-01.md` (status: `done`), which lists this plan in its own `related:`.
2. `.context/findings/eval-gating-byok-2026-06-29.md` — status: `active`, should be `done`. Its ask (two-tier gate: structural required everywhere incl. forks, LLM-judge advisory/non-blocking; provider-agnostic BYOK client) is fully implemented: `.github/workflows/skill-quality.yml` has a "Tier 1: structural eval gate (required every PR/main push)" step calling `skill-auditor eval ... --fail-below 0` and a `continue-on-error: true` advisory step calling `skill-auditor eval --json --samples 3 --cost-log` gated on `ANTHROPIC_API_KEY`; `internal/llmclient/` provides the provider-agnostic client.
3. `.context/findings/tessl-eval-criteria-schema-2026-06-30.md` — status: `active`, should be `done`. Its subject (pinning the `tessl eval run` CLI to v0.88.2 in `skill-quality.yml`) is moot — `tessl eval run` no longer appears in `skill-quality.yml`; only `tessl review run` (a separate, unrelated advisory step) remains.
4. `.context/findings/plumber-cicd-security-2026-07-04.md` — status: `active`, should be `done`. All 4 of its own "Recommended Action" items are verified complete: `.github/workflows/plumber.yml` exists and is committed, `.plumber.yaml` is tracked in git, the 5 third-party actions are pinned by commit SHA and permissions blocks added (PR #171, merged), and fail-on-Critical gating shipped (plan `plumber-advisory-workflow-2026-07-04.md`, status `done`; ADR-037/038/039 accepted).
5. `.context/findings/scoring-pattern-config-review-2026-07-03.md` — status: `active`, should be `done`. Its own "Recommended Action" ("flip ADR-028's frontmatter status: proposed → accepted") is already done — `docs/ADR/adr-028-scoring-pattern-config.md` is status `accepted`, and `docs/ADR/adr-031-analysis-quality-scope.md` (the sibling decision this finding also motivated) is status `accepted` too.
6. `docs/ADR/adr-001-native-eval-runner.md` — status: `proposed`, should be `accepted`. Same implementation evidence as item 1/2.
7. `docs/ADR/adr-007-two-tier-eval-gate.md` — status: `proposed`, should be `accepted`. Same implementation evidence as item 2.
8. `docs/ADR/adr-005-freedom-calibration-remediation.md` — status: `proposed`, should be `accepted`. Its decision ("restructure D6 scorer to weight actions by situational context") is implemented as `scoreConstraintTypology` in `scorer/d6_freedom_calibration.go`, and its context plan (`.context/plans/dimension-improvements/d6-freedom-calibration-2026-04-29.md`) is status `done`.

Two near-misses were investigated and confirmed **not** to be drift, recorded here so a future auditor doesn't re-litigate them:

- `.context/findings/skilleval-analysis-2026-06-30.md` — its recommended `ab-test` command genuinely isn't built yet; correctly still `active`.
- `.context/findings/yaml-content-validation-config-2026-07-03.md` — proposes a content-safety `validation/` package that is a broader, distinct scope from what the identically-dated, `done` plan of the same name actually delivered; correctly still `active` despite the name collision.

## How this was detected (worth automating)

Two mechanical, scriptable checks caught most of this without reading file content:

- **(a)** For every ADR with `status: proposed`, resolve its `context:` links against `.context/index.yaml` — if any linked plan/finding has `status: done`, the ADR is a drift candidate.
- **(b)** For every plan with `status: done`, walk its `related:` list — any `active`/`draft` plan or finding appearing there is a drift candidate (a done plan citing an active one as a predecessor it fulfilled is a strong "this got superseded" signal).

Both are pure YAML cross-referencing (no LLM judgment needed) and would have caught 7 of the 8 items above without manual code verification. Item 3 (`tessl-eval-criteria-schema`) needed a content check (grep for a since-removed CI step) that only makes sense as a semantic/LLM-assisted second pass, not a pure status cross-reference.

## Recommended Action

Do not auto-apply any of these flips — same principle `automate-post-merge-status-sync-2026-07-04.md` established: ADR acceptance and cross-item supersession are human judgment calls, not mechanical field writes.

1. Fix the 8 items above by hand, in one PR, now that they're verified.
2. Build a periodic (not per-PR) drift-audit automation — see the companion plan for the design — that runs checks (a) and (b) above on a schedule (e.g. weekly cron GitHub Action), and opens or updates a single tracking issue listing drift candidates for human review. It must never auto-write status changes itself, only surface candidates, mirroring the auto-flip/flag split `merge-status-sync.sh` already established for the per-PR case.

This finding does not create an ADR — no binding decision has been made, only an observation plus a recommendation to build tooling; the design decisions belong in the companion plan.
