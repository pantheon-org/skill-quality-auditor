---
title: "Finding: check-docs-drift.sh's cumulative mode never runs in CI, only at local pre-push"
type: finding
status: active
date: 2026-07-06
value: medium
related:
  - ../plans/docs-drift-reviewed-baseline-2026-07-06.md
  - ../../docs/ADR/adr-044-docs-drift-pr-gate.md
  - ../../docs/ADR/adr-045-docs-drift-reviewed-baseline.md
  - ../../.github/workflows/ci.yml
  - ../../hk.pkl
---
# Finding: check-docs-drift.sh's cumulative mode never runs in CI, only at local pre-push

> While answering a question about how the reviewed-baseline mechanism (ADR-045) actually surfaces to reviewers, it became clear that `check-docs-drift.sh`'s cumulative mode — the mode the reviewed-baseline sidecar exists to serve — only ever runs via `hk.pkl`'s local `pre-push` hook. It never runs in CI. Anyone who pushes with `--no-verify`, merges via GitHub's web UI, or lands a commit via the API bypasses it entirely, reviewed-baseline or not. This should have been surfaced during `.context/plans/docs-drift-reviewed-baseline-2026-07-06.md`'s plan-review, not discovered afterward.

## Summary

`.github/workflows/ci.yml`'s `docs-drift` job (added in ADR-044) runs `scripts/check-docs-drift.sh "origin/${{ github.base_ref }}"` — gate mode only. Gate mode's code path never reads `doc_date`, the `MAPPINGS`-driven cumulative loop, or the reviewed-baseline sidecar at all; it only diffs the current PR's own commit range. `hk.pkl`'s `["docs-drift"]` step runs `scripts/check-docs-drift.sh` with no arguments — cumulative mode — but `hk`'s `pre-push` hooks are client-side git hooks, installed locally via `hk install` (`mise.toml`'s `postinstall` hook) and invoked only when a contributor runs `git push` on their own machine. Nothing server-side ever invokes it.

Practical consequence: the entire reviewed-baseline mechanism just built (ADR-045) — the fix for "a doc reviewed as accurate keeps re-flagging forever" — only benefits contributors who push through a checkout with `hk`'s hooks installed and don't bypass them. A PR opened via GitHub's UI, a squash-merge, or any push with `--no-verify` never triggers cumulative mode at all, so nobody reviewing that PR ever sees a cumulative-mode warning, correct or false-positive.

## Detail

This was in front of the amending agent the whole time — the plan's own Goal section states cumulative mode is "consumed by `hk.pkl`'s `pre-push` hook, its only caller" — but that fact was recorded, not interrogated. None of the three plan-review prompts (Technical: feasibility/gaps; Strategic: goal alignment/completeness; Risk: blind spots/failure modes) explicitly asked "where does this mechanism actually execute, and who never sees it run" — a question that sits at the intersection of Strategic's "completeness" and Risk's "blind spots" but wasn't prompted for directly in either brief. The three reviews it did receive were thorough on the mechanism's internals (date-comparison correctness, idempotency, trust model) and caught real bugs (the ISO8601-string-comparison risk, the Phase-3-unreachable-pre-rebase issue) — but a mechanism-internals review doesn't automatically surface a scope/coverage question about where the mechanism is invoked at all.

Process learning, stated plainly: documenting a fact about a system's invocation path ("X is the only caller of Y") is not the same as tracing what that fact implies for coverage or bypassability. The plan-review skill's standard reviewer prompts don't currently ask any reviewer to verify "given how/where this triggers, does its actual reach match the stated goal" — that's a gap in the review's own question set, not a one-off oversight by a specific reviewer pass.

## Follow-up

Immediate, same-day fix (implemented in the same branch/PR as ADR-045, since that work hasn't merged yet): add cumulative mode as a second step in `ci.yml`'s existing `docs-drift` job. Cumulative mode already always `exit 0`s (informational-only by design), so this cannot newly fail or block anything — it only makes its output visible in every PR's Actions log, closing the "silently bypassed" gap without turning cumulative mode into a second blocking gate (which was never proposed or reviewed, and isn't warranted here).

Broader, not decided here: whether the `plan-review` skill's Technical/Strategic/Risk prompts should be updated to explicitly ask "trace where/how this mechanism is actually invoked, and who bypasses it" as a standing question, so this class of gap is caught during review rather than after. Left as an open question for whoever owns that skill's prompts — not applied unilaterally in this finding.
