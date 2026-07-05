---
title: "Finding: Shared bot identity caused automated PR comments to clobber each other"
type: finding
status: done
date: 2026-07-05
related:
  - ../../docs/ADR/adr-039-plumber-pr-comment.md
  - ../../docs/ADR/index.yaml
---
# Finding: Shared bot identity caused automated PR comments to clobber each other

> `gh pr comment --edit-last` edits the last comment by the *authenticated identity*, not the last comment by *the script that invoked it*. Every automated comment in this repo posts as the same `github-actions[bot]` identity, so once a second workflow started commenting on PRs, `--edit-last` began silently overwriting whichever comment had been posted most recently — regardless of which tool wrote it.

## Summary

While adding a PR comment step to `skill-quality.yml` (surfacing structural audit results that previously only appeared in Actions logs), a user asked why automated PR comments didn't identify which tool posted them. Investigating that led to discovering a more serious bug: `plumber.yml`'s existing comment step (`gh pr comment --edit-last --create-if-none`, added in ADR-039) was silently destroying the new `skill-quality.yml` comment whenever Plumber ran afterward on the same PR, because both workflows' comments are posted under the identical `github-actions[bot]` identity and `--edit-last` has no way to distinguish "my last comment" from "any bot's last comment."

Observed live on PR #180: the `skill-quality.yml` structural-audit comment vanished entirely after a subsequent Plumber run overwrote it.

## Timeline

1. Added a `github.rest.issues.createComment`/`listComments`/`updateComment` step to `skill-quality.yml`, using an explicit HTML-comment marker (`<!-- skill-quality-structural-comment -->`) to find-and-update its own comment across reruns — the correct pattern, chosen because a marker was already necessary to distinguish it from the pre-existing advisory eval comment step in the same workflow.
2. First live run hit `403 Resource not accessible by integration`. The response's `x-accepted-github-permissions` header revealed that commenting on a PR via the Issues API requires **both** `issues: write` and `pull-requests: write`, not `issues: write` alone — confirmed by `plumber.yml`, which already declared both. Fixed by adding `pull-requests: write` to `skill-quality.yml`'s permissions.
3. After that fix, both workflows' comment steps succeeded independently — but the PR then showed the Plumber-format comment where the skill-quality comment should have been, and vice versa depending on run order. Comparing timestamps (`created_at` vs `updated_at`) across all comments on the PR confirmed the pattern: whichever workflow ran *last* had overwritten whichever comment happened to be *most recent at the time*, sometimes Plumber's own prior comment, sometimes skill-quality.yml's.
4. Root cause: `gh pr comment --edit-last` (added in ADR-039, reasoned there as needing "no custom marker-search logic... `gh` already tracks 'the authenticated identity's last comment on this PR' natively") is scoped to the **authenticated identity**, not to the invoking script. That reasoning held as long as `plumber.yml` was the only workflow commenting on PRs; it silently broke the moment a second workflow (`skill-quality.yml`) started doing the same thing under the same bot identity.
5. Fixed `plumber-pr-comment.sh` to use the same marker-based lookup already used by `skill-quality.yml`: search `GET /issues/{pr}/comments` for a body containing `<!-- plumber-pr-comment -->`, `PATCH` that specific comment ID if found, otherwise create one.
6. That fix introduced a second, independent bug: `gh api ... -f body="@$TMP_BODY"` posted the **literal string** `@/path/to/file` as the comment body instead of the file's contents. `gh api --help` documents the `@file` "read value from file" convention only for `-F`/`--field` (typed), not `-f`/`--raw-field` (raw string) — `-f` takes its value completely literally. Confirmed by deliberately triggering the failure, reading `gh api --help` in full, and testing `-F` directly against the live (already-corrupted) comment to confirm it fixed the content before committing.
7. Manually repaired the corrupted comment via `gh api ... -X PATCH -F body=@file` before pushing the fix, so the live PR wasn't left in a broken state.

## Root Causes

1. **Identity-scoped "last comment" lookup is unsafe once more than one workflow shares a bot identity.** `--edit-last` has no concept of "my comments" vs "any comment by this identity" — it will happily edit a comment it did not create.
2. **`gh api`'s `@file` file-reading convention is `-F`-only.** `-f`/`--raw-field` is a raw string parameter with no special handling of a leading `@` — an easy mistake since both flags accept `key=value` syntax and only differ in this one respect.

## Fix

- `plumber-pr-comment.sh`: replaced `gh pr comment --edit-last --create-if-none` with an explicit `<!-- plumber-pr-comment -->` marker search (`gh api .../comments --paginate | jq 'select(...) | last.id'`), then `PATCH`es that specific comment ID via `gh api -X PATCH -F body="@$TMP_BODY"` if found, or creates a fresh comment if not.
- `skill-quality.yml`: added `pull-requests: write` to permissions; both comment-posting steps (`Post eval results comment`, `Post structural audit comment`) now include a bold header naming the source workflow and a footer link to the generating run, for the same reason a human reviewer asked about in the first place — comment provenance should be legible without reading a hidden HTML marker.
- `plumber-pr-comment.sh` and `skill-quality.yml`'s structural-audit comment now both self-identify with a `##`-level header (`🛡️ Plumber`, `🔍 Skill Quality Gate`) plus a `_[View this run](...)_` footer link.

## Recommended Action

Captured as ADR-043 (supersedes ADR-039's comment-update mechanism): every automated PR-commenting script in this repo must use marker-based comment lookup-and-update, never `gh pr comment --edit-last`, because bot identity is shared across every workflow in the repo and will remain so for any future comment-posting addition.
