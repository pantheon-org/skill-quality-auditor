---
title: "Finding: automate post-merge plan/ADR status sync via a GitHub Action"
type: FINDING
status: ACTIVE
date: 2026-07-04
value: MEDIUM
themes:
  - GOVERNANCE
  - PR-TOOLING
related:
  - ../plans/post-merge-status-sync-2026-07-04.md
  - ../plugins/pantheon-org/governance/adr-capture/scripts/merge-status-sync.sh
  - ../instructions/ways-of-working.md
  - ../../.context/index.yaml
---
# Finding: automate post-merge plan/ADR status sync via a GitHub Action

> `merge-status-sync.sh` already does the work — plan/ADR status sync and index regeneration after a PR merges — but only when a human remembers to invoke it. This finding surfaces wiring it to a merge-triggered GitHub Action, which the `post-merge-status-sync-2026-07-04` plan already identified and explicitly deferred.

## Summary

Idea surfaced while wrapping up the Plumber CI/CD integration (plans `plumber-advisory-workflow-2026-07-04`, ADR-036 through ADR-039): after each of the four PRs in that effort merged, the plan's own frontmatter `status` and `.context/index.yaml` needed a manual follow-up commit to stay accurate. That's exactly the gap `post-merge-status-sync-2026-07-04` (status: `done`) already built a fix for — `merge-status-sync.sh --dry-run <pr-number>` detects drift, and running it without `--dry-run` opens its own branch + PR with safe status flips and a regenerated index.

Checking that plan's own frontmatter confirms this isn't a new idea: its "Out of scope (deferred)" section explicitly lists *"A GitHub Actions workflow that runs this automatically on every merge to `main` — the plan below produces something runnable on demand first; wiring it into CI is a follow-up once the detection logic is trusted."* This finding is that follow-up thread, picked back up now that the script has had real usage (the Plumber work is at least the second multi-PR effort to lean on it).

## Detail

**What exists today:** `merge-status-sync.sh <pr-number>` (and its `--dry-run` variant), invoked manually per `ways-of-working.md`'s "After merge" section, step 2. It:
- Resolves which plans/ADRs a merged PR relates to (frontmatter `related:`/`context:` links first, file-touch overlap second, reusing `check-plan-drift.sh`'s path resolution).
- Auto-flips single-phase plan `status: active → done` via its own branch + PR — never a direct commit to `main`.
- Always *flags* (never auto-applies) ADR `status: proposed → accepted` and multi-phase plan completions, since those are documented judgment calls, not mechanical field flips (the original plan's Decision 1 and 2).
- Regenerates `.context/index.yaml` / `docs/ADR/index.yaml` as part of any auto-flip PR.

**What's missing:** nothing triggers it. A human has to remember "did I forget to flip the ADR" after every merge — the exact manual-reminder problem the plan was written to replace, just moved one level up (from "remember to update status" to "remember to run the sync script").

**Proposed automation shape** (not yet implemented, not yet decided):
- Trigger: `pull_request: types: [closed]` filtered to `github.event.pull_request.merged == true`, or a `workflow_run` following a successful CI run on `main` — the former is more direct (fires exactly once per merge, has the PR number for free) and is probably the better fit.
- Action: run `merge-status-sync.sh <pr-number>` (no `--dry-run`) against the just-merged PR.
- Output handling must not change: the script already opens its own branch + PR for anything it writes rather than committing to `main` directly (Decision 3 of the original plan) — the automation should call the *existing* script unmodified, not reimplement its output path as a direct commit.

## Recommended Action

Worth doing, but scope carefully around what the original plan already decided should stay a human judgment call:

1. **Good fit for automation as-is:** the mechanical parts — single-phase plan flips and index regeneration — since the script already gates these behind "safe to auto-apply" logic and routes them through a review-able PR rather than a direct write.
2. **Must not be automated further than the script already allows:** ADR `proposed → accepted` and multi-phase plan completions must keep surfacing as flagged, human-reviewed items, not something a scheduled/triggered run silently resolves. The risk with full automation isn't the mechanism (a GH Action calling an idempotent, already-PR-gated script is low-risk) — it's scope creep in a future edit that has the Action start resolving the flagged items too, since "it's already automated" is an easy rationalization once the workflow exists.
3. Before wiring the trigger: confirm `merge-status-sync.sh`'s `gh` calls work correctly with the default `GITHUB_TOKEN` in Actions context (it currently assumes an authenticated `gh` CLI session, which in local/manual use is the invoking human's own credentials — permissions and rate limits may differ under Actions).

This finding does not create an ADR — no binding decision has been made, only an observation that the previously-deferred follow-up is now worth reconsidering.
