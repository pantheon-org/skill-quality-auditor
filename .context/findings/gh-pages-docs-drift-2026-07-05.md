---
title: "Finding: GH Pages docs-drift investigation — no deploy pipeline bug found"
type: FINDING
status: ACTIVE
date: 2026-07-05
value: LOW
related:
  - ../../.github/workflows/docs.yml
  - ../../docmd.config.json
  - ../../scripts/check-docs-drift.sh
---
# Finding: GH Pages docs-drift investigation — no deploy pipeline bug found

> A report came in that the deployed GH Pages docs site hadn't updated with the latest repo changes and that a bug in the deploy pipeline was preventing it. Direct verification against the live site and the GitHub Pages deployment API shows the pipeline is working correctly; the actual gap is docs content drift, not a deployment defect.

## Summary

`.github/workflows/docs.yml` triggers on `push` to `main` for paths `docs/**` and `docmd.config.json`, builds with `npx @docmd/core build`, uploads via `actions/upload-pages-artifact@v3`, and deploys via `actions/deploy-pages@v4`. Checked against both the GitHub Actions run history and the live site's response headers, the pipeline has correctly deployed every `docs/**`-touching commit, including the most recent one. There is no stuck deployment, no silently-swallowed failure, and no CDN staleness. The perceived staleness is real, but it lives upstream of the pipeline: several `docs/**` pages describe features that have since changed in the code, and nobody updated the docs to match — the site is fully current with what's committed under `docs/`, but `docs/` itself has drifted from the source it documents.

## Detail

**Pipeline verification.** The most recent commit touching `docs/**` was `c5089fd` (PR #180, which adds ADR-043). Its `Deploy documentation` run succeeded at `2026-07-05T18:55:31Z`, and `gh api repos/.../deployments/<id>/statuses` confirms the resulting deployment reached `state: success`. Curling the live site directly (`https://pantheon-org.github.io/skill-quality-auditor/`) shows `last-modified: Sun, 05 Jul 2026 18:56:03 GMT`, matching the deploy time, and the site's `llms.txt` already lists ADR-043 — the newest content that should be there. That rules out a stuck deployment, a failed run being ignored, and CDN/edge caching serving something stale.

One transient failure did occur, on commit `275c2ca` ("Deployment failed, try again later" — a GitHub-side Pages hiccup surfaced by `actions/deploy-pages@v4`, not a workflow logic error). The very next push retried and succeeded automatically, so this self-healed and left no lasting gap.

At investigation time, the two most recent commits on `main` (`3126653` — the T-shirt effort sizing plan feature — and a follow-up plan-date edit) touch no files under `docs/**`, so `docs.yml` correctly never fired for them. That is the `paths:` trigger working as designed, not a bug: a commit that doesn't change `docs/**` or `docmd.config.json` has nothing new for the site to build.

**Root cause of the perceived staleness.** Several `docs/**` pages describe functionality that has since moved in the code — PR-Agent integration, `agent-scan`, T-shirt effort sizing, and others — without a corresponding docs edit landing alongside those changes. Because `docs.yml`'s trigger is (correctly) scoped to `docs/**`, a code change that never touches `docs/**` never causes a rebuild, and the stale prose sits there indefinitely with no mechanism forcing a refresh.

This repo already has a heuristic aimed at exactly this gap: `scripts/check-docs-drift.sh`, wired into `hk`'s pre-push hook. It holds a static map of each `docs/**` page to related source globs (e.g. `docs/architecture/remediation-flow.md` → `cmd/remediate.go;reporter/remediation*.go`) and flags a doc as possibly stale when source commits post-date the doc's last commit. It is informational only — the script always `exit 0`s — so it warns locally without blocking a push or failing CI, and staleness can accumulate silently past it.

At investigation time, running that heuristic flagged 8 stale docs: `README.md`, `docs/index.md`, `docs/architecture/overview.md`, `docs/architecture/duplication-flow.md`, `docs/architecture/remediation-flow.md`, `docs/architecture/eval-runner.md`, `docs/reference/scoring-dimensions.md`, and `docs/development/skills-and-rules.md`.

## Follow-up

No pipeline fix is needed — `docs.yml` and GitHub Pages deployment are functioning correctly and the live site is verified current with `docs/` source.

The actionable follow-up is closing the content gap: update the 8 docs pages flagged by `check-docs-drift.sh` so they reflect the features that have shipped since each was last touched.

Separately, worth a maintainer decision (not applied here): whether `check-docs-drift.sh` should become a blocking CI check rather than a pre-push-only advisory, so this class of staleness can't silently recur. This is a process trade-off, not an obvious fix — left open rather than auto-applied.
