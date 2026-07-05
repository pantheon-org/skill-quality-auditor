---
title: "ADR-043: Marker-based lookup for all automated PR comment upserts"
status: accepted
date: 2026-07-05
context:
  - path: "docs/ADR/adr-039-plumber-pr-comment.md"
  - path: ".context/findings/pr-comment-clobbering-2026-07-05.md"
---

**Status:** Accepted
**Date:** 2026-07-05

## Context

ADR-039 chose `gh pr comment --edit-last --create-if-none` for `plumber.yml`'s PR comment, reasoning that "no custom marker-search logic [is] needed... `gh` already tracks 'the authenticated identity's last comment on this PR' natively." That reasoning held only as long as `plumber.yml` was the sole workflow commenting on PRs in this repo.

Adding a PR comment step to `skill-quality.yml` (surfacing structural audit results that previously only appeared in Actions logs) broke that assumption. Both workflows' comments post under the identical `github-actions[bot]` identity — the identity GitHub Actions grants to every workflow's default `GITHUB_TOKEN`, with no per-workflow distinction. `--edit-last` edits the last comment by that identity, not the last comment made *by the invoking script*, so once a second commenting workflow existed, each run's `--edit-last` began silently overwriting whichever comment — Plumber's or Skill Quality Gate's — happened to be most recent at that moment. This was observed live on PR #180: the Skill Quality Gate's audit comment vanished after a subsequent Plumber run. Full investigation in `.context/findings/pr-comment-clobbering-2026-07-05.md`.

## Decision

1. **Every automated PR-commenting script or step in this repo must locate its own comment by an explicit, unique HTML marker** (e.g. `<!-- plumber-pr-comment -->`, `<!-- skill-quality-structural-comment -->`) embedded in the comment body, never by relying on "last comment by the authenticated identity."
2. **The update path is: list comments on the PR/issue, find the one containing the marker, `PATCH` that specific comment ID.** If none is found, create a new comment. In `bash`, this is `gh api .../issues/{n}/comments --paginate | jq 'select(...) | last.id'` then `gh api .../issues/comments/{id} -X PATCH -F body="@file"` (`-F`, not `-f` — see the Gotcha below). In `actions/github-script`, this is `github.rest.issues.listComments` → find by marker → `updateComment` or `createComment`.
3. **`plumber-pr-comment.sh` is retrofitted to this pattern**, replacing `--edit-last --create-if-none`. `skill-quality.yml`'s structural-audit comment step already used this pattern from the start (chosen independently, for the unrelated reason that the workflow already has a second, marker-free comment step for advisory eval results and needed to tell the two apart).
4. **Every comment additionally self-identifies with a visible `##`-level header naming the source** (e.g. `🛡️ Plumber`, `🔍 Skill Quality Gate — Structural Audit`) and a footer link to the generating workflow run. The marker solves *correctness* (don't edit the wrong comment); the header solves *legibility* (a human reviewer shouldn't need to view raw markdown to see which tool commented) — both were raised as the same underlying "we don't know what commented" concern.

### Gotcha: `gh api`'s `@file` convention is `-F`-only

While implementing this, `-f body="@$TMP_BODY"` posted the literal string `@/path/to/file` as the comment body instead of its contents. `gh api --help` documents the "if the value starts with `@`, read it from that file" behavior only for `-F`/`--field` (typed); `-f`/`--raw-field` takes its value completely literally. Both flags accept identical `key=value` syntax, making this an easy mistake — anyone adding a new marker-based comment script in this repo should use `-F`, not `-f`, when sourcing a body from a file.

## Consequences

- **Easier:** any number of workflows can post PR comments under the shared `github-actions[bot]` identity without risk of one clobbering another's comment.
- **Easier:** a reviewer can identify which automation posted a given comment without opening raw markdown to find a hidden marker.
- **Harder:** marginally more code per comment step than `--edit-last` (a list-and-find step before the update), though this repo already had that logic in `skill-quality.yml`'s structural-audit step, so the pattern was proven before being applied to `plumber.yml`.
- **Binding for future work:** any new automated PR-commenting feature added to this repo must follow this pattern from the start, not `--edit-last`. `--edit-last` remains safe only for a repo where exactly one workflow ever comments on PRs — not the case here, and unlikely to become simpler as more automation is added.
