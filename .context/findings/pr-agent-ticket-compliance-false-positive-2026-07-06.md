---
title: "Finding: PR-Agent's ticket compliance analysis misfires on casual '#N' PR-description references"
type: FINDING
status: ACTIVE
date: 2026-07-06
value: LOW
themes:
  - PR-TOOLING
related:
  - ../plans/pr-agent-integration-2026-07-04.md
  - ../../.pr_agent.toml
  - ../../.github/workflows/pr-agent.yml
---
# Finding: PR-Agent's ticket compliance analysis misfires on casual "#N" PR-description references

> PR #187's PR-Agent `/review` comment flagged "🎫 Ticket compliance analysis ❌ — [186] Not compliant," listing requirements from PR #186 (an unrelated, already-merged docs-fix PR) as if #187 were supposed to satisfy them. Root cause: PR #187's description contained plain prose like "closing out #186" and "forked before PR #186" — narrative references, not tracked-work links. GitHub auto-links any bare `#186`, and PR-Agent's `require_ticket_analysis_review` feature (default `true`) treats any such auto-link as a "ticket" the current PR must be graded against, then fetches the linked PR/issue's own description and evaluates the current diff against it.

## Summary

Confirmed via `.pr_agent.toml`'s current config (no `require_ticket_analysis_review` setting present, so PR-Agent's documented default of `true` applies) and Qodo Merge's own documentation (`https://qodo-merge-docs.qodo.ai/tools/review/`, `https://docs.qodo.ai/qodo-documentation/qodo-merge/tools/tools-list/review`): setting `require_ticket_analysis_review = true` (or leaving it unset) makes `/review` add a "ticket compliance" section whenever the PR body/commits contain any GitHub/Jira/Linear reference, whether or not that reference was meant as a tracked requirement. This repo doesn't use GitHub Issues as its primary "ticket" mechanism — `.context/plans/` and `.context/findings/` already serve that role, established across every PR this integration has seen so far (#183–#188) — so any bare `#N` reference used for narrative continuity (a very natural thing to write, e.g. "closing out #186's follow-up work") is a false trigger for a feature designed around actual issue-tracker workflows.

## Detail

The specific false positive: PR #187's description said "4 real docs hit this while closing out #186" and "the branch had forked before PR #186," and a commit body said "was unreachable pre-rebase, since it forked before PR #186." None of these were "Fixes #186" / "Closes #186" closing-keyword syntax — just prose. GitHub's auto-linking doesn't distinguish intent, and PR-Agent's ticket-compliance feature inherits that ambiguity: it pulled #186's PR description (itself a checklist of doc-content fixes: "update eval-runner.md, overview.md, remediation-flow.md, skills-and-rules.md...") and graded #187's completely unrelated diff (a docs-drift tooling mechanism) against that checklist, correctly finding zero overlap and reporting "Not compliant" — a technically-accurate-but-meaningless result, since #187 was never trying to satisfy #186's requirements in the first place (those were already done, in #186, which had already merged).

This is exactly the kind of comment-quality/false-positive signal the PR-Agent integration plan's Phase 2 observation window (`.context/plans/pr-agent-integration-2026-07-04.md`, running through ~2026-07-19) exists to catch.

## Follow-up

Disabled `require_ticket_analysis_review` in `.pr_agent.toml` (see accompanying commit) rather than pursuing the alternative of introducing GitHub Issues as a required linked-ticket layer in this repo's workflow — that would duplicate the `.context/plans/`/`.context/findings/` mechanism already serving as this repo's "ticket" equivalent, for a feature whose main value (grading a PR against a tracked issue's requirements) doesn't apply when there's no such tracked issue to grade against. Logged in the PR-Agent plan's Phase 2 tracking as an observed false positive with a root cause and fix, not left as an unexplained one-off comment.
