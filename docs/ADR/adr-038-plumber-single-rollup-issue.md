---
title: "ADR-038: Plumber non-Critical findings — single rollup issue, not one issue per finding"
status: accepted
date: 2026-07-04
context:
  - path: "docs/ADR/adr-037-plumber-critical-fail-issue-tracking.md"
  - path: ".context/findings/plumber-cicd-security-2026-07-04.md"
  - path: ".context/plans/plumber-advisory-workflow-2026-07-04.md"
---

**Status:** Accepted
**Date:** 2026-07-04

## Context

ADR-037 point 3 specified: "Every High/Medium/Low finding is filed as a deduplicated GitHub issue." The first implementation (`.github/workflows/plumber.yml` + `scripts/plumber-file-issues.sh`, merged in #124) did exactly that — one issue per individual finding, deduplicated by a fingerprint hash of the finding's control, code, and full issue payload.

Verifying that implementation against a real PR run plus the push-to-main run it triggered on merge exposed a bug: the fingerprint hashed Plumber's `url` field verbatim, and that field is a live GitHub blob link (`.../blob/<commit-sha>/path#Lline`) embedding the *current* commit SHA. Since the SHA changes on every commit, the same underlying finding produced a different fingerprint on every run and was never recognized as already tracked. The PR run and the push-to-main run each filed a full set of findings, producing 29 duplicate issues for 15 unique findings. All 29 were closed as duplicates once diagnosed.

While the immediate cause was fixable (normalize the commit-ref segment out of the URL before hashing — done in #154), the user separately directed a change to the underlying design, independent of that bug: **"We should only have only ONE issue per run and NO duplicates."** One issue per finding was never the right shape even with correct dedup — a repo with a dozen open High/Medium findings would carry a dozen separate GitHub issues for what is really one ongoing "Plumber compliance backlog" concern. This reverses ADR-037 point 3's specific mechanism, so per this repo's ADR-immutability rule, ADR-037 is superseded rather than edited.

## Decision

1. **Exactly one persistent issue tracks all High/Medium/Low findings**, not one issue per finding. Located by a stable marker HTML comment in the issue body (`<!-- plumber-compliance-backlog -->`), not by title, since a title can be edited by anyone.
2. **Every run regenerates the issue's full body** from the current `results.json` — a per-severity table of every open finding (code, job/branch location, source link). The body is not a history; it reflects only the current state.
3. **Lifecycle, driven by whether the marker issue exists and its state:**
   - No existing issue + findings present → create it.
   - Existing open issue + findings present → edit its body in place.
   - Existing closed issue + findings present → reopen it and edit its body (findings regressed after being resolved).
   - Existing open issue + zero findings → comment that everything is resolved, then close it.
   - No existing issue + zero findings → no-op.
   - This guarantees at most one open tracking issue at any time; "no duplicates" holds by construction, not by a fingerprint that can drift.
4. **The per-finding dedup fingerprint mechanism from the first implementation is removed entirely** — it is no longer needed once there is only ever one issue to find or not find.
5. **Critical findings are still never listed in this issue** — they block the pipeline via `plumber-gate.sh` (ADR-037 point 2, unchanged) and are fixed before merge, not tracked as backlog.
6. **The rollup issue is only maintained on `push` to `main`, not on `pull_request`.** Live-verifying this PR (#154) against the real repo surfaced the reason: `pull_request` runs scan the PR branch's speculative merged state, not `main`. With one shared issue, two concurrent/unrelated PRs would each overwrite it with their own branch's findings, and a PR that fixes a finding could have it reappear if another PR's run lands after. The Critical gate (`plumber-gate.sh`) still runs on every PR and push — only the non-blocking backlog tracking is push-to-main-only. This also removes the need for the fork-PR skip step from the first implementation, since issue-filing no longer runs on `pull_request` at all.

## Consequences

- **Easier:** the issue tracker gets exactly one signal for Plumber's non-Critical backlog instead of a growing pile of one-off issues; anyone can find current compliance debt by looking at one open issue instead of searching/filtering.
- **Easier:** no more fingerprint design to get right — "does the marker exist and what state is it in" is simpler and more robust than "does this specific finding's hash match a previous one."
- **Harder:** the issue body loses per-finding history — closing and reopening the same tracking issue means there's no persistent record of when an individual finding first appeared or how long it sat open, beyond what's visible in the issue's edit/comment timeline.
- **Harder:** if someone manually edits the rollup issue's body (adding notes, triage state), the next run's regeneration overwrites it. The body explicitly says "do not edit it by hand," but nothing enforces that.
