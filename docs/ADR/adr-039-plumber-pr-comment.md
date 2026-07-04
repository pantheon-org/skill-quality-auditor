---
title: "ADR-039: Plumber PR comment — single comment per PR, edited in place"
status: accepted
date: 2026-07-04
context:
  - path: "docs/ADR/adr-037-plumber-critical-fail-issue-tracking.md"
  - path: "docs/ADR/adr-038-plumber-single-rollup-issue.md"
---

**Status:** Accepted
**Date:** 2026-07-04

## Context

A question about `plumber.yml` surfaced that a failing Critical gate produces no comment on the PR — only a failed check and `::error::` annotations, neither of which is as visible as a PR conversation comment. Confirmed against both `plumber.yml` (no comment step existed) and the upstream `getplumber/plumber` Action (`action.yml` writes a job summary to `$GITHUB_STEP_SUMMARY`, not a PR comment). This repo's own `skill-quality.yml` already has a comment step for its eval results, so the gap was specific to `plumber.yml`.

`skill-quality.yml`'s comment step (`github.rest.issues.createComment`, unconditional on every triggering run) was the obvious template, but copying it verbatim would repeat the mistake ADR-038 just corrected: a fresh comment on every run means every push to a long-lived PR piles up another comment in the conversation, exactly like the per-finding issues that had to be cleaned up in #154.

## Decision

1. **`plumber.yml` gets a comment step, `pull_request`-only** (`if: always() && github.event_name == 'pull_request' && ...`), reporting the Critical gate's pass/fail state and a per-severity breakdown table (code, job/branch location, source link) for every finding on the branch — the same level of detail `plumber-file-issues.sh` writes into the rollup issue, capped at 25 rows per severity with an "...and N more" note pointing at the run's `plumber-compliance` artifact. A bare count ("11 High, 4 Medium") tells a PR author nothing actionable; the table does. Also links to the persistent rollup issue (ADR-038) for the High/Medium/Low subset, noting explicitly that the rollup issue reflects `main`, not this PR's branch, since the two can diverge until merge.
2. **The comment is edited in place across reruns**, via `gh pr comment --edit-last --create-if-none` — no custom marker-search logic needed, unlike the rollup issue: `gh` already tracks "the authenticated identity's last comment on this PR" natively, and edits it if present or creates one if not. Applies the same "one artifact, regenerated in place" principle as ADR-038, using a simpler native mechanism available for PR comments but not for issues.
3. **`permissions:` gains `pull-requests: write`**, alongside the existing `contents: read` and `issues: write` — `gh pr comment` operates on the PullRequest object, a distinct scope from issue comments.
4. **Fork PRs are skipped explicitly** with a `::notice::`, mirroring the same read-only-`GITHUB_TOKEN` constraint documented for issue-filing in the original design (ADR-037).

## Consequences

- **Easier:** PR authors see the gate result and backlog counts directly in the conversation, without opening the Actions tab — closing the exact gap that prompted this ADR.
- **Easier:** unlike `skill-quality.yml`'s comment step, this never accumulates duplicate comments across reruns of the same PR.
- **Harder:** `pull-requests: write` is a third distinct write scope on this workflow (alongside `contents: read` and `issues: write`), a slightly larger permission surface than the original design.
- **Note:** `skill-quality.yml`'s own comment step still creates a fresh comment on every run and was not changed by this ADR — retrofitting it to the same `--edit-last --create-if-none` pattern is a reasonable follow-up but out of scope here.
