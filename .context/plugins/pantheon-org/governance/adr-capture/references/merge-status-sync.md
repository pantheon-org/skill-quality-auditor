# Merge Status Sync Reference

`scripts/merge-status-sync.sh` closes the gap where a PR ships the feature a plan or ADR describes, but the plan stays `ACTIVE`/`DRAFT` and the ADR stays `proposed` because nothing checks status against merge state. See `.context/plans/post-merge-status-sync-2026-07-04.md` for the full design rationale.

## Usage

```bash
scripts/merge-status-sync.sh --dry-run <pr-number>   # report drift, write nothing
scripts/merge-status-sync.sh <pr-number>              # apply safe auto-flips via a branch + PR
```

Run it from any branch — the script opens its own branch and PR when it has something to write, and always returns you to the branch you started on.

## Signal strength

For each `.context/plans/*.md` (`status: ACTIVE`/`DRAFT`) and `docs/ADR/adr-*.md` (`status: proposed`), the script cross-references the merged PR's touched files against:

1. The plan/ADR file's own path — **direct** signal.
2. Its `related:`/`context:` list, when the linked path is itself another `.context/` or `docs/ADR/` file — **frontmatter** signal.
3. Its `related:`/`context:` list, when the linked path is anything else (e.g. shared source or docs) — **file-touch** signal.

| Signal | Meaning | Auto-flip eligible? |
| --- | --- | --- |
| `direct` | The PR touched the plan/ADR file itself | Yes (plans only, single-phase) |
| `frontmatter` | The PR touched another governance artifact this one links to | Yes (plans only, single-phase) |
| `file-touch` | The PR touched a shared source/doc file this one merely references | No — always flagged |
| `none` | No overlap | Not reported |

## Auto-flip rules

- **Plans**: only single-phase plans (at most one `### Phase N` heading) with `direct`/`frontmatter` signal auto-flip `status: ACTIVE|DRAFT → DONE`. Multi-phase plans are always flagged, even under a strong signal — a PR that closes one phase of a multi-phase plan must not silently mark the whole plan done.
- **ADRs**: never auto-flip, regardless of signal strength. Acceptance is a deliberate decision, not a mechanical status field — flag it for a human to confirm and accept via the normal `adr-capture` workflow.
- **File-touch-only links**: always flagged, never auto-applied — a plan referencing a shared config file doesn't mean every PR touching that file implements the plan.

## What "flagged" means in practice

A flagged item is printed in the "Flagged for confirmation" section of the report and never written to. This includes cases where the heuristic itself is ambiguous — e.g. a plan whose `related:` list cites an ADR purely as background context can register a `frontmatter` signal against an unrelated PR that happens to touch that ADR. Because multi-phase plans and ADRs are always flagged rather than auto-applied, this kind of false positive is visible for a human to dismiss, never silently acted on.

## What auto-flip does

For eligible single-phase plans, an apply-mode run (no `--dry-run`):

1. Checks for an already-open sync PR (`chore/status-sync-pr-<n>`) — if one exists, it's a no-op.
2. Fetches `origin/main` and branches `chore/status-sync-pr-<n>` from it (never commits to the branch you're on, never pushes to `main`).
3. Flips `status: ACTIVE|DRAFT` to `status: DONE` in each eligible plan file.
4. Runs `context-index`'s `regenerate-context-index.sh`.
5. Commits, pushes, and opens a PR — the PR body includes the full flagged summary so ADRs and multi-phase plans still get surfaced for a human decision.
6. Returns to the branch you started on.

Re-running against a PR that's already fully synced (nothing left `ACTIVE`/`DRAFT`/`proposed`) reports "nothing to do" and exits 0.

## Squash-merge fallback

`gh pr view --json files` is the primary source of touched files. If it comes back empty or smaller than the PR's own commit count, the script falls back to parsing `git log --format=%B -n 1 <merge-sha>` for path-like tokens in the merge commit message — covering a squash-merge convention where the file list might not reflect individual commits.
