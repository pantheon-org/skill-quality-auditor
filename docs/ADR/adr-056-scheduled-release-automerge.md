---
title: "ADR-056: Release-please PRs are auto-merged on a schedule via a GitHub App token"
status: accepted
date: 2026-07-06
context:
  - path: docs/ADR/adr-053-release-draft-then-publish.md
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

release-please opens a `chore(main): release X.Y.Z` PR that accumulates changelog and
version-bump changes from merged commits. Cutting a release required a human to manually
merge that PR. That manual merge is load-bearing: the `Release` workflow runs on push to
`main`, and the merge push is what re-triggers it so goreleaser builds and publishes the
release (draft-then-publish flow, ADR-053).

The manual step was unwanted friction. Two constraints shaped the fix:

1. **Cadence.** Releases should batch — changes accumulate in the release PR and ship on a
   predictable schedule, not on every merge to `main`.
2. **Token anti-recursion.** A merge performed with the default `GITHUB_TOKEN` does **not**
   trigger new workflow runs. If auto-merge used `GITHUB_TOKEN`, the release PR would land
   but the `Release` workflow would never fire, producing a tag with no binaries — the exact
   empty-release failure fixed in PR #219 / ADR-053.

## Decision

Auto-merge the open release-please PR on a schedule, authenticating with a **GitHub App
installation token** so the merge push triggers the `Release` workflow.

- New workflow `.github/workflows/release-automerge.yml` runs weekly (`cron: 0 9 * * 1`,
  Mondays 09:00 UTC) and on `workflow_dispatch`.
- It mints a short-lived App token with `actions/create-github-app-token` (SHA-pinned per
  ADR-054), finds the open `chore(main): release` PR, and squash-merges it. Non-mergeable
  states (DIRTY/BLOCKED/BEHIND) are skipped, not forced.
- release-please in `Release` continues to create the PR with `GITHUB_TOKEN`; only the
  **merge** needs the App identity, because only the pusher's identity governs whether the
  resulting push triggers workflows.

A **GitHub App token** was chosen over a Personal Access Token: it is short-lived, scoped to
the installation, and carries no long-lived broadly-scoped credential — consistent with the
supply-chain posture established in ADR-054 and ADR-055.

We **reuse the existing org-wide GitHub App** rather than creating a new one. The org already
runs a general automation App (`pantheon-ai-bot`, installed org-wide with `contents: write`
and `pull_requests: write`), surfaced to workflows as the Actions variable `GH_APP_ID` and
the secret `GH_APP_PRIVATE_KEY`. This is the same pair `pantheon-org/claude-code-personalities`
uses for its release automation, so the convention is established.

## Required one-time setup

Most of this is already in place org-wide. Confirm, do not recreate:

1. **Org credentials resolve for this repo.** The workflow reads `vars.GH_APP_ID` and
   `secrets.GH_APP_PRIVATE_KEY`. Ensure the org variable/secret grant access to
   `skill-quality-auditor` (org secret/variable repository-access policy). Needs org admin
   only if this repo is not yet in scope.
2. **Ruleset bypass (only if needed).** The `main` ruleset requires 0 approving reviews, so
   the App merge is expected to pass unblocked (confirmed: the open release PR reports
   `mergeStateStatus: CLEAN`). If GitHub's code-owner rule (CODEOWNERS `* @thoroc`) ever
   blocks the App merge, add the App as a **bypass actor** on the `main` ruleset.

No per-repo App creation, private key generation, or new secrets are required.

## Consequences

- Releases ship on a predictable weekly cadence with no manual merge. `workflow_dispatch`
  allows an on-demand release when needed.
- The App token is the load-bearing detail: switching the merge back to `GITHUB_TOKEN` would
  silently stop triggering the `Release` workflow and reintroduce empty releases. This ADR
  records *why* the App token exists.
- The release PR is not gated on CI (it never was — its contents are generated from commits
  already validated on `main`). The workflow refuses to merge non-clean PRs as a safeguard.
- Rotating or losing the App private key disables auto-merge; the fallback is the original
  manual merge, which still works unchanged.
