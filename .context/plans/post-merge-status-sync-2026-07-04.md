---
title: "Plan: Post-Merge ADR/Plan Status Sync"
type: plan
status: draft
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
- Auto-flipping plan `status: active → done` when the PR implements what the plan describes (this mirrors what `ways-of-working.md` already asks a human to do manually in the same PR).
- Flagging (not auto-flipping) ADR `status: proposed → accepted` — this is a judgment call, same as the ADR-028/ADR-032 precedent where acceptance happened in a deliberate follow-up commit, not automatically.
- Running the existing `regenerate-context-index.sh` and `regenerate-adr-index.sh` scripts as the final step whenever a flip happens.
- Running as a manual/on-demand skill in the first cut (triggered by the user or invoked at PR-merge time via `gh`), not a CI-blocking gate.

**Out of scope (deferred):**
- A GitHub Actions workflow that runs this automatically on every merge to `main` — the plan below produces something runnable on demand first; wiring it into CI is a follow-up once the detection logic is trusted.
- Auto-accepting ADRs without confirmation, even when confidence is high — acceptance is a documented decision, not a mechanical status field.
- Retroactively auditing every historical plan/ADR pair — this plan only wires up detection going forward; a one-off backfill (like the ADR-032 fix already committed) is a separate, smaller task if a full sweep is wanted.

## Decisions

1. **Plan flips auto-apply; ADR flips require confirmation.** `ways-of-working.md` already treats `active → done` on a plan as mechanical ("update its frontmatter status: active → done in the same PR"). ADR acceptance is different in kind — ADR-028 and ADR-031 show acceptance happening as a deliberate, separate decision, sometimes with wording changes beyond the status field. Automating that would make "accepted" mean less.
2. **Merge detection uses `gh pr view --json mergedAt,files` against the PR the user names**, not a background poller. A cron-based approach would need write access to run outside a session and a place to persist "last-checked PR number" state — both bigger asks than the problem currently justifies. Start as an on-demand skill invocation ("check for status drift on PR #118" or "check all recently merged PRs").
3. **Linking heuristic: frontmatter first, file-touch second.** A plan/ADR's `related:`/`context:` field is the authoritative link. Where a merged PR's file list overlaps with a plan/ADR's `related:` paths, that's a strong signal; where no plan/ADR links the PR's files at all, the skill reports "no linked plan/ADR found" rather than guessing.
4. **Lives as an extension to `adr-capture`, not a new top-level skill.** `adr-capture` already owns `regenerate-adr-index.sh` and `check-undocumented-decisions.sh` — this is the same family of "keep the ADR index honest" work, just triggered by merge state instead of by new-decision detection. A new script (`check-merge-status-drift.sh`) plus a new workflow section in `adr-capture/SKILL.md` is smaller than standing up a separate skill with its own eval suite.

## Phases

### Phase 1 — Detection script

- Write `check-merge-status-drift.sh` under `adr-capture/scripts/`: given a PR number, resolve `gh pr view <n> --json mergedAt,files`, cross-reference against `related:`/`context:` fields in `.context/index.yaml` and `docs/ADR/index.yaml`, and print any linked plan/ADR whose status is still `active`/`draft` (plans) or `proposed` (ADRs).
- Unit-test the cross-reference logic against the ADR-032 case as a fixture (PR #118's file list vs. ADR-032's `context:` block) to confirm it would have caught the drift.
- Exit criterion: running the script against PR #118 today reports zero drift (since ADR-032 is now `accepted`), and running it against a synthetic "PR merged, ADR still proposed" fixture reports the mismatch.

### Phase 2 — Auto-flip plans, flag ADRs

- Extend the script (or add a sibling `apply-merge-status-sync.sh`) to auto-write `status: done` on any plan flagged in Phase 1, then run `regenerate-context-index.sh`.
- For flagged ADRs, print a confirmation prompt/summary instead of writing — matching the "flag, don't auto-flip" decision above — and leave `regenerate-adr-index.sh` to run only after a human accepts.
- Document the new workflow step in `adr-capture/SKILL.md` (`## Workflow`) and add a `## Scripts` entry for both new scripts, following the existing doc pattern in that file.
- Exit criterion: running the combined flow against a merged PR that finished a plan flips that plan to `done` and regenerates the index unattended; a PR whose ADR is still `proposed` produces a flagged report and makes no file changes until confirmed.

### Phase 3 — Wire into the merge workflow

- Add a step to `ways-of-working.md`'s "After merge" section pointing at the new script as the replacement for the manual reminder.
- Evaluate (don't yet build) whether this becomes a required step in a GitHub Actions post-merge workflow — capture the tradeoffs as an Open Question below rather than deciding here.
- Exit criterion: `ways-of-working.md` reflects the new tool-assisted step; the manual "did I forget to flip the ADR" failure mode has a documented, repeatable command to run instead of relying on memory.

## Risks

- **False positives on the file-touch heuristic** — a PR that happens to touch a file a plan links to (e.g. a shared config file referenced by multiple plans) could be misattributed. Mitigated by treating file-touch overlap as a secondary signal behind explicit frontmatter links (Decision 3), and by only ever flagging (never auto-flipping) when the signal is file-touch-only.
- **Scope creep toward a full CI gate** — it would be easy to over-build this into a blocking merge-queue check before the detection logic has been exercised on enough real PRs to trust it. Phase 3 explicitly defers that decision.
- **ADR acceptance still requires a human to actually run the confirmation step** — this plan reduces the chance of forgetting by making the check cheap to run, but doesn't eliminate the "nobody ran the script" failure mode. A CI reminder (not a gate) may be worth revisiting once Phase 3's evaluation lands.

## Verification

```bash
# Phase 1: detection only, no writes
.context/plugins/pantheon-org/governance/adr-capture/scripts/check-merge-status-drift.sh 118

# Phase 2: apply safe flips + regenerate indices
.context/plugins/pantheon-org/governance/adr-capture/scripts/apply-merge-status-sync.sh 118

# Confirm indices are internally consistent afterward
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/*.md
.context/plugins/pantheon-org/governance/adr-capture/scripts/validate-adr-frontmatter.sh docs/ADR/adr-*.md
```

## Open Questions

- Should Phase 3's GitHub Actions integration post a PR comment listing drifted ADRs/plans, or open a follow-up issue? Left for whoever picks up Phase 3.
- Does the file-touch heuristic need to look at squash-merge commit messages too (for PRs merged without preserving individual commits), or is the `gh pr view --json files` list sufficient in all repos this might spread to? Untested against a squash-merged PR.
- Is a single script per concern (`check-*` / `apply-*`) the right split, or should this be one script with a `--dry-run` flag, matching the `--dry-run`/`-n` convention already used by `skill-auditor aggregate` and `remediate`? Worth matching existing CLI conventions in Phase 1 rather than inventing a new pattern.
