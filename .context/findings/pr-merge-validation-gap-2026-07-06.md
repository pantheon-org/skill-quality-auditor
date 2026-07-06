---
title: "Finding: no skill covers validating and merging a PR"
type: finding
status: active
date: 2026-07-06
value: low
related:
  - ../plugins/pantheon-org/workshop/pr-author/SKILL.md
  - ../plugins/pantheon-org/governance/adr-capture/SKILL.md
  - ../instructions/ways-of-working.md
---

# Finding: no skill covers validating and merging a PR

> This repo has skills for PR *description* lifecycle and post-merge status sync, but
> nothing covers the step in between: confirming a PR is actually safe to merge, then
> merging it. That step has been done ad hoc, by hand, for every PR this session.

## Summary

Investigated whether an existing skill covers "validate a PR's CI status and merge
it" before formalizing the ad hoc pattern used repeatedly this session. None does.

## Detail

Two skills touch PR lifecycle, and both explicitly stop short of this gap:

- **`pr-author`** (`.context/plugins/pantheon-org/workshop/pr-author/SKILL.md`) —
  scoped entirely to PR *description* lifecycle: create, update on push, respond to
  change requests, re-request review, draft-to-ready. Its "Merge strategy note" only
  says what to *document* in the body ("prefer squash for feature branches"), not what
  to *run*. Nothing in its Workflow, Anti-Patterns, or References sections executes a
  merge.
- **`adr-capture`'s `merge-status-sync.sh`** — runs strictly *after* a PR has already
  merged, to catch plans/ADRs left `active`/`draft`/`proposed` when the merge should
  have closed them out. It has no pre-merge role.
- **`ways-of-working.md`** documents the branch → PR → merge → "after merge" sequence
  as prose, but the merge step itself is one line ("Use `gh pr create` or push and open
  via GitHub") with no guidance on confirming readiness first.

`grep -rl "gh pr merge" .context/plugins/pantheon-org/` returns nothing — no skill
anywhere in this repo invokes a merge.

Meanwhile, this exact validate-and-merge step happened by hand at least eight times
this session (PRs #187 through #199), always the same shape:

1. Poll `gh pr checks <n> --json name,bucket` (or a `Monitor`-based poll loop) until
   every check is terminal, not just started.
2. If a rebase hit a merge conflict on an auto-generated file (`.context/index.yaml`,
   `docs/ADR/index.yaml`), resolve via `git checkout --ours <file>` → re-run the
   regenerate script → re-verify → force-push-with-lease, per Rule 15 — never
   hand-resolve conflict markers.
3. Only once every check is green (or a known-advisory `continue-on-error` step is the
   sole non-green entry), run `gh pr merge`.
4. Confirmed at least once this session that `gh pr merge` succeeding while checks were
   still "pending" is a real risk worth flagging explicitly, since this repo has no
   required-status-checks branch protection forcing the wait.

Nobody wrote this down as a procedure — it was re-derived from memory each time, which
is exactly the pattern Rule 4 ("Formalise ad hoc scripts after repeated use") exists to
catch, extended here to a repeated *skill-shaped* pattern rather than a script.

## Recommended Action

Draft a plan for a new skill — tentatively `pr-merge`, workshop domain, alongside
`pr-author` — that formalizes:

- Waiting for every check to reach a terminal state (not just started), including
  distinguishing a real failure from an expected `continue-on-error` advisory step.
- The regenerate-don't-hand-merge conflict-resolution sequence for auto-generated
  files (Rule 15), since a rebase-induced conflict is the most common reason a PR
  needs attention between "opened" and "mergeable."
- A default merge strategy (squash, matching `pr-author`'s documented preference) and
  branch deletion, matching `ways-of-working.md`'s "after merge" step.
- Explicit confirmation before merging is not optional — this is a shared-state,
  hard-to-reverse action per this project's own risk-taking guidance, even though it
  happens after CI is green.

No overlap with `pr-author` (description-only) or `adr-capture`'s post-merge sync
(after the fact) — this fills the gap between them.
