---
title: "ADR-044: check-docs-drift.sh gates PRs on newly introduced drift"
status: accepted
date: 2026-07-06
context:
  - path: ".context/findings/gh-pages-docs-drift-2026-07-05.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

`.context/findings/gh-pages-docs-drift-2026-07-05.md` investigated a report that the deployed GH Pages docs site wasn't reflecting recent repo changes. The deploy pipeline itself (`docs.yml` + `actions/deploy-pages`) turned out to be working correctly — the actual gap was content drift: `scripts/check-docs-drift.sh` already maps each `docs/**` page to related source globs and flags a doc as possibly stale when source commits post-date the doc's last commit, but it runs only as a pre-push hook and always `exit 0`s. That means it prints a warning to the pusher's own scrolling terminal, never fails the push, never runs in CI, and never surfaces on the PR itself — so a reviewer on GitHub's web UI has no way to see it, and staleness accumulates silently. The finding left open, as a maintainer call rather than something to auto-apply, whether to promote this check into a blocking CI gate.

At investigation time the cumulative check already had 8 flagged stale docs (`README.md`, `docs/index.md`, and six others). Any gate reusing that same all-history logic verbatim would fail every PR immediately on pre-existing debt unrelated to that PR's own changes, which would make the gate impossible to land without first clearing the backlog.

## Decision

1. **`scripts/check-docs-drift.sh` gains a second mode, selected by passing a base ref as `$1`.** With no argument, it keeps its existing behavior unchanged: the pre-push hook's cumulative, all-history, always-`exit 0` advisory check. With a base ref, it switches to gate mode.
2. **Gate mode only fails on drift introduced by the current diff.** For each mapped doc, it checks whether the doc itself was touched between the base ref and `HEAD`; if so, it's considered current regardless of source changes. If not, it checks whether any of that doc's mapped source globs were touched in the same range — if so, the gate fails (exit 1) and names the doc and the offending commits. This deliberately does not reuse the pre-push hook's "any commit since the doc's last touch, however long ago" comparison, so pre-existing drift never blocks a PR that doesn't touch the affected source.
3. **Wired into `.github/workflows/ci.yml` as a new `docs-drift` job, scoped to `pull_request` only** (not `push: [main]` — by the time a change lands on `main` it's already merged, too late to block). The job checks out with `fetch-depth: 0` (required for the `base...HEAD` commit-range diff) and runs `scripts/check-docs-drift.sh "origin/${{ github.base_ref }}"`.
4. **The 8 pre-existing stale docs are not addressed by this change and are not force-fixed as a prerequisite.** They remain visible via the unchanged pre-push advisory check; closing that backlog is separate follow-up work, not a blocker for landing the gate.

## Consequences

- **Easier:** new source changes that should come with a doc update now fail CI visibly on the PR itself, instead of relying on a warning that only the pushing developer's terminal ever showed.
- **Easier:** the gate can land immediately without a separate cleanup pass, because it only judges the PR's own diff, not the repo's accumulated history.
- **Harder:** a PR that legitimately doesn't need a doc update (e.g. a source change that doesn't change any user-facing behavior described in the mapped doc) will still fail the gate and needs a docs touch (even a trivial one) to pass, or an explanation in the PR description — there's no override flag by design, matching this repo's other advisory-vs-blocking precedents of favoring an explicit human override over a machine-parsed skip mechanism.
- **Binding for future work:** any new `docs/**` page or new source area that should be kept in sync needs an entry added to `check-docs-drift.sh`'s `MAPPINGS` array to be covered by either mode; unmapped source has no drift detection at all.
- **Deferred:** whether/when to clear the 8 pre-existing stale docs remains open, tracked informally via the unchanged cumulative pre-push check rather than a dedicated plan.
