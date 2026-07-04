---
title: "ADR-033: Post-merge ADR/plan status sync via merge-status-sync.sh"
status: accepted
date: 2026-07-04
context:
  - path: ".context/plans/post-merge-status-sync-2026-07-04.md"
  - path: "docs/ADR/adr-032-user-configurable-scoring-patterns.md"
---

**Status:** Accepted
**Date:** 2026-07-04

## Context

A PR can ship the feature an ADR or plan describes while the ADR stays `proposed` and the plan stays `active`/`draft`, because nothing checks status against merge state. ADR-032 was the motivating case: PR #118 merged it, but ADR-032 stayed `proposed` until a manual audit caught it during an unrelated "what's left on our plans" question. `.context/plans/post-merge-status-sync-2026-07-04.md` designed a script to close this gap; this ADR captures its binding decisions.

## Decision

1. **Plan flips auto-apply only when the plan has a single phase; multi-phase plans are always flagged, never auto-flipped.** This repo's own plans are frequently multi-phase and shipped across several PRs, so auto-flipping `status: done` on a PR that only closes one phase would be a silent data-integrity regression.
2. **ADR flips always require human confirmation, regardless of signal strength.** Acceptance is a deliberate decision, sometimes with wording changes beyond the status field (see ADR-028, ADR-031) — automating it would make "accepted" mean less.
3. **Auto-flips are committed via a branch + PR, never directly to `main`**, mirroring `ways-of-working.md`'s golden rule exactly as a human would follow it by hand.
4. **Merge detection uses `gh pr view --json mergedAt,files,commits,mergeCommit` against a PR number the user names**, not a background poller — a cron-based approach needs write access and persisted "last-checked" state, both bigger asks than the problem justifies today.
5. **Linking heuristic: the PR's own file list is cross-referenced against each candidate plan/ADR's own path (`direct` signal) and its `related:`/`context:` list (`frontmatter` signal when the linked path is itself another `.context/`/`docs/ADR/` file, `file-touch` signal otherwise).** File-touch-only links are always flagged, never auto-applied, since a plan referencing a shared config file doesn't mean every PR touching that file implements the plan.
6. **Single script, `--dry-run`/`-n` flag** — matches the existing `skill-auditor aggregate`/`remediate` convention rather than a separate `check-*`/`apply-*` pair.
7. **The script is idempotent by re-deriving from live file state** — re-running against an already-synced PR reports "nothing to do" with no persisted state needed; an already-open sync PR short-circuits a second apply-mode run.
8. **Lives as an extension to `adr-capture`** (`scripts/merge-status-sync.sh`, `references/merge-status-sync.md`), not a new top-level skill — it's the same family of "keep the ADR index honest" work as `regenerate-adr-index.sh`/`check-undocumented-decisions.sh`, just triggered by merge state instead of new-decision detection.
9. **CI integration (posting a PR comment listing drifted ADRs/plans) is evaluated, not built** — Phase 3 wires a manual step into `ways-of-working.md`'s "After merge" section; a required GitHub Actions gate is deferred until the detection logic has been exercised on enough real PRs to trust it.
10. **Implemented in pure POSIX shell (bash/awk/sed), not Python** — `.agents/RULES.md`'s "Avoid Python/Node.js scripts in skills" rule, adopted during this same implementation, applies here: skill scripts must not assume an interpreter beyond what ships on a stock Unix-like system. This also surfaced two real portability bugs fixed along the way: bash 3.2 (macOS's stock `/bin/bash`) loses `BASH_REMATCH` capture groups for a `[[ =~ ]]` pattern used inside a function loaded via `source`, and macOS/BSD `realpath` has no `-m` flag at all (GNU-only) — both replaced with portable `case`-glob matching and a hand-rolled path normalizer.

## Consequences

- **Easier:** the "did I forget to flip the ADR" failure mode has a cheap, repeatable command instead of relying on memory — `merge-status-sync.sh --dry-run <n>`.
- **Easier:** single-phase plan closures are fully automated end-to-end (branch, commit, push, PR) with no risk of a stray commit to `main`.
- **Harder:** ADR acceptance and multi-phase plan confirmation still require a human to act on the flagged report — this reduces the chance of forgetting but doesn't eliminate the "nobody ran the script" failure mode. Revisit once Phase 3's CI evaluation lands.
