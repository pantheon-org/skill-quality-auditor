---
title: "Plan: Post-Merge ADR/Plan Status Sync"
type: plan
status: done
date: 2026-07-04
related:
  - ../../docs/ADR/adr-032-user-configurable-scoring-patterns.md
  - ../instructions/ways-of-working.md
  - ../plugins/pantheon-org/governance/adr-capture/SKILL.md
  - ../plugins/pantheon-org/context-mgmt/context-index/SKILL.md
---

## Goal

Close the gap where a PR ships the feature an ADR or plan describes, but the ADR stays `proposed` and the plan stays `active`/`draft` because nothing checks status against merge state. ADR-032 ("User-configurable scoring pattern overrides") is the motivating case: PR #118 merged it three commits ago, but ADR-032 was still `proposed` until a manual audit caught it during an unrelated "what's left on our plans" question. Success looks like: after any PR merges, its linked plans and ADRs are checked against merge state, safe status flips happen automatically, judgment-call flips are surfaced for confirmation, and `.context/index.yaml` / `docs/ADR/index.yaml` stay current without a human remembering to run the regen scripts.

## Scope

**In scope:**
- Detecting merged PRs and resolving which `.context/plans/*.md` and `docs/ADR/*.md` files they relate to (via frontmatter `related:`/`context:` links, and/or files touched by the PR).
- Auto-flipping plan `status: active → done`, via a branch + PR, when the PR fully implements a single-phase plan (see Decisions); multi-phase plans are flagged for confirmation instead.
- Flagging (not auto-flipping) ADR `status: proposed → accepted` — this is a judgment call, same as the ADR-028/ADR-032 precedent where acceptance happened in a deliberate follow-up commit, not automatically.
- Running the existing `regenerate-context-index.sh` and `regenerate-adr-index.sh` scripts as part of any auto-flip PR.
- Running as a manual/on-demand skill in the first cut (triggered by the user or invoked at PR-merge time via `gh`), not a CI-blocking gate.

**Out of scope (deferred):**
- A GitHub Actions workflow that runs this automatically on every merge to `main` — the plan below produces something runnable on demand first; wiring it into CI is a follow-up once the detection logic is trusted.
- Auto-accepting ADRs without confirmation, even when confidence is high — acceptance is a documented decision, not a mechanical status field.
- Retroactively auditing every historical plan/ADR pair — this plan only wires up detection going forward; a one-off backfill (like the ADR-032 fix already committed) is a separate, smaller task if a full sweep is wanted.

## Decisions

1. **Plan flips auto-apply only when the plan has a single open phase/task-group; multi-phase plans are flagged, not flipped.** `ways-of-working.md` already treats `active → done` on a plan as mechanical for the common case ("update its frontmatter status: active → done in the same PR"). But this repo's own plans are frequently multi-phase and shipped across several PRs (this plan is itself an example). Auto-flipping `status: done` on a PR that only closes one phase of a multi-phase plan would be a silent data-integrity regression — worse than the manual gap this plan replaces. So: single-phase (or already-fully-matched multi-phase) plans auto-flip; anything else is flagged for confirmation, same as ADRs.
2. **ADR flips always require confirmation.** ADR-028 and ADR-031 show acceptance happening as a deliberate, separate decision, sometimes with wording changes beyond the status field. Automating that would make "accepted" mean less.
3. **Auto-flips are committed via a branch + PR, never directly to `main`.** `ways-of-working.md`'s top rule is "never commit to main directly." `merge-status-sync.sh` therefore creates a branch (e.g. `chore/status-sync-pr-<n>`), commits the frontmatter flip(s) and regenerated index, and opens a PR — mirroring exactly what a human following `ways-of-working.md` would do by hand. It never pushes to `main`.
4. **Merge detection uses `gh pr view --json mergedAt,files` against the PR the user names**, not a background poller. A cron-based approach would need write access to run outside a session and a place to persist "last-checked PR number" state — both bigger asks than the problem currently justifies. Start as an on-demand skill invocation ("check for status drift on PR #118" or "check all recently merged PRs").
5. **Linking heuristic: frontmatter first, file-touch second — reusing `scripts/check-plan-drift.sh`'s existing path-resolution logic.** A plan/ADR's `related:`/`context:` field is the authoritative link. `check-plan-drift.sh` already solves resolving those relative paths (`realpath -m "$PLAN_DIR/$raw_path"`) against repo-root-relative paths — the new script calls the same resolution logic rather than re-deriving it, and the new script's output is explicitly cross-referenced against `check-plan-drift.sh`'s existing pre-push output so the two drift mechanisms (commit-date heuristic vs. PR-merge-event heuristic) don't silently disagree. Where a merged PR's file list overlaps with a plan/ADR's resolved `related:` paths, that's a strong signal; where no plan/ADR links the PR's files at all, the skill reports "no linked plan/ADR found" rather than guessing.
6. **Single script, `--dry-run`/`-n` flag** — not a separate `check-*`/`apply-*` pair. This matches the existing convention in `skill-auditor aggregate`/`remediate` (`--dry-run`/`-n` prints what would change; omitting it applies). `merge-status-sync.sh --dry-run <n>` reports drift and proposed actions without writing; `merge-status-sync.sh <n>` applies safe flips (Decision 1/3) and always prints the ADR-flag summary for confirmation.
7. **The script is idempotent.** Re-running `merge-status-sync.sh <n>` against a PR that's already fully synced is a no-op (exit 0, "nothing to do"). Failures from `gh` (PR not found, not yet merged, unauthenticated, rate-limited) are reported and exit non-zero; a partial failure between flipping a plan's frontmatter and regenerating `.context/index.yaml` is safe to recover from by re-running the script, since index regeneration always re-derives from current file state regardless of whether a flip happened this invocation.
8. **Lives as an extension to `adr-capture`, not a new top-level skill.** `adr-capture` already owns `regenerate-adr-index.sh` and `check-undocumented-decisions.sh` — this is the same family of "keep the ADR index honest" work, just triggered by merge state instead of by new-decision detection. The plan-status half (`.context/index.yaml` regeneration) calls into `context-index`'s existing `regenerate-context-index.sh` rather than duplicating it. A new script (`merge-status-sync.sh`) plus a new workflow section in `adr-capture/SKILL.md` is smaller than standing up a separate skill with its own eval suite.
9. **Phase 3's eventual CI integration posts a PR comment, not a follow-up issue.** A comment lands where a reviewer's attention already is (on the just-merged PR) and requires no separate triage/close lifecycle. This is a decision for when Phase 3 is actually built, not a change to Phase 3's current "evaluate, don't build" scope.
10. **The file-touch heuristic also parses squash-merge commit messages, not just `gh pr view --json files`.** `gh pr view --json files` is confirmed sufficient for this repo's own merge convention (standard merge commits), but a squash-merged PR elsewhere could summarise multiple original commits' files into a message body rather than the API's file list reflecting them individually. The detection script treats `--json files` as the primary source and falls back to parsing `git log --format=%B -n 1 <merge-sha>` for individual commit references only when `files` comes back empty or suspiciously small relative to the PR's commit count — this keeps the common path simple while covering the edge case defensively.

## Phases

### Phase 1 — Detection script

- Write `merge-status-sync.sh --dry-run <n>` under `adr-capture/scripts/`: given a PR number, resolve `gh pr view <n> --json mergedAt,files`, resolve `related:`/`context:` paths using `check-plan-drift.sh`'s existing `realpath -m` logic, cross-reference against `.context/index.yaml` and `docs/ADR/index.yaml`, and print any linked plan/ADR whose status is still `active`/`draft` (plans) or `proposed` (ADRs), plus whether a flagged plan is single-phase (auto-flip candidate) or multi-phase (always flagged, per Decision 1).
- Implement the squash-merge fallback from Decision 10: when `files` is empty or its count looks inconsistent with the PR's commit count, parse `git log --format=%B -n 1 <merge-sha>` for individual commit references before giving up on file-touch matching.
- Unit-test the cross-reference logic against **three** fixtures: (a) PR #118 checked out at its pre-acceptance commit (before `d06d7b6`), so the test actually simulates the "PR merged, ADR/plan still open" state ADR-032 was caught in, rather than testing against today's already-fixed repo; (b) a synthetic PR whose file list overlaps only via file-touch (not frontmatter) with an unrelated plan/ADR's `related:` paths, to prove the file-touch-only case is flagged, never auto-applied; (c) a synthetic squash-merged PR with an empty/undersized `files` response, to prove the commit-message fallback from Decision 10 fires and still finds the match.
- Cross-check output against `check-plan-drift.sh`'s existing pre-push report for the same repo state to confirm the two heuristics don't produce contradictory signals.
- Exit criterion: fixture (a) reports the ADR-032 drift that actually existed at merge time; fixture (b) reports a flag, not an auto-flip; fixture (c) reports the match via the fallback path; running against PR #118 today (fully synced) reports "nothing to do."

### Phase 2 — Auto-flip single-phase plans, flag everything else

- Extend the script to drop `--dry-run` and apply: auto-write `status: done` only for plans matching Decision 1's single-phase criterion; for multi-phase plans and all ADRs, print a confirmation summary and make no writes.
- On an applicable auto-flip, create a branch (`chore/status-sync-pr-<n>`), commit the frontmatter change(s) plus `regenerate-context-index.sh`'s output, and open a PR — per Decision 3, never push to `main` directly.
- Idempotency: re-running against an already-synced PR is a no-op (exit 0); `gh` failures (not found, not merged, unauthenticated, rate-limited) are reported and exit non-zero without partial writes.
- Document the new workflow step in `adr-capture/SKILL.md` (`## Workflow`) and add a `## Scripts` entry, following the existing doc pattern; update `adr-capture`'s eval suite (`evals/scenario-0{1,2,3}`, `summary.json`) so it reflects the new script and workflow step.
- Exit criterion: running against a merged PR that fully closes a single-phase plan opens a PR flipping that plan to `done` with a regenerated index; running against a PR that only partially closes a multi-phase plan, or whose linked ADR is still `proposed`, produces a flagged report and opens no PR; re-running either case a second time is a no-op.

### Phase 3 — Wire into the merge workflow

- Add a step to `ways-of-working.md`'s "After merge" section (currently empty per source-plan review) pointing at `merge-status-sync.sh` as the replacement for the manual reminder — run from any branch, not from `main`, since the script itself opens its own branch/PR when it needs to write.
- Evaluate (don't yet build) whether this becomes a required step in a GitHub Actions post-merge workflow. Per Decision 9, if and when that workflow is built, it posts a PR comment on the merged PR listing any drifted ADRs/plans — not a follow-up issue.
- Exit criterion: `ways-of-working.md` reflects the new tool-assisted step; the manual "did I forget to flip the ADR" failure mode has a documented, repeatable command to run instead of relying on memory.

## Risks

- **False positives on the file-touch heuristic** — a PR that happens to touch a file a plan links to (e.g. a shared config file referenced by multiple plans) could be misattributed. Mitigated by treating file-touch overlap as a secondary signal behind explicit frontmatter links (Decision 5), and by only ever flagging (never auto-flipping) when the signal is file-touch-only.
- **Multi-phase plans misread as fully done** — mitigated by Decision 1's single-phase-only auto-flip guard; anything else is flagged for confirmation rather than written.
- **Scope creep toward a full CI gate** — it would be easy to over-build this into a blocking merge-queue check before the detection logic has been exercised on enough real PRs to trust it. Phase 3 explicitly defers that decision.
- **ADR acceptance, and multi-phase plan confirmation, still require a human to act on the flagged report** — this plan reduces the chance of forgetting by making the check cheap to run and by opening a ready-to-review PR for the safe cases, but doesn't eliminate the "nobody ran the script" failure mode for the flagged cases. A CI reminder (not a gate) may be worth revisiting once Phase 3's evaluation lands.
- **Unprotected `main`** — this repo's `main` branch currently has no branch protection configured, so nothing external would stop a future bug from pushing there directly. Decision 3's "always branch + PR" behavior is the only safeguard; it should not be weakened even though nothing would technically block a direct push today.

## Verification

```bash
# Phase 1: detection only, no writes
.context/plugins/pantheon-org/governance/adr-capture/scripts/merge-status-sync.sh --dry-run 118

# Phase 2: apply safe flips (opens a branch + PR; never pushes to main) + regenerate indices
.context/plugins/pantheon-org/governance/adr-capture/scripts/merge-status-sync.sh 118

# Re-run to confirm idempotency (should report "nothing to do")
.context/plugins/pantheon-org/governance/adr-capture/scripts/merge-status-sync.sh 118

# Confirm indices are internally consistent afterward
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/*.md
.context/plugins/pantheon-org/governance/adr-capture/scripts/validate-adr-frontmatter.sh docs/ADR/adr-*.md
```
