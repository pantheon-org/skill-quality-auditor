---
title: "Ways of Working"
type: INSTRUCTION
status: ACTIVE
date: 2026-07-01
---

# Ways of Working

## Golden rule — never commit to main

**ALWAYS** work on a branch. Direct commits to `main` are forbidden. Every change — no matter how small — must go through a branch, a PR, and merge into `main`.

## Branch workflow

1. **Start from the latest `main`.** Fetch the latest remote state before branching:
   ```
   git checkout main && git pull && git checkout -b <type>/<short-description>
   ```
   Never branch from a stale local `main` — a `git pull` right before `checkout -b` is mandatory.

2. Use conventional prefixes: `feat/`, `fix/`, `docs/`, `refactor/`, `chore/`.

3. Commit as you go — small, atomic commits with conventional messages:
   ```
   feat(scorer): add D9 mutation coverage
   fix(hook): regenerate index on plan changes
   docs: update README install section
   ```

4. If `main` has diverged, rebase instead of merging:
   ```
   git fetch origin && git rebase origin/main
   ```
   This keeps history linear. Resolve conflicts if they arise.

5. Run checks before pushing:
   ```
   hk check          # pre-commit checks (lint, validate, index freshness)
   go test ./...     # full test suite
   ```
   Pre-push includes a plan-status drift check that warns about plans marked `ACTIVE` for more than 60 days. If you see these warnings, update the plan's frontmatter `status: ACTIVE → DONE` if the work is complete.

6. Push and open a PR:
   ```
   git push -u origin <branch-name>
   ```
   Use `gh pr create` or push and open via GitHub.

## Keeping plans in sync

When you implement what a plan describes, update its frontmatter `status: ACTIVE → DONE` in the same PR. The `context-index` hook will regenerate `.context/index.yaml` automatically.

Every plan (`type: PLAN`) with `status: DRAFT` or `ACTIVE` must also carry an `effort: S|M|L|TBD` frontmatter field — a T-shirt-sized total effort estimate, matching the `skill-auditor remediate` convention already used for skill remediation plans. `validate-context-frontmatter.sh` enforces this. Use `TBD` only when sizing is genuinely blocked on an unresolved item in the plan's Open Questions — don't pick a number just to pass validation. `effort` is set once at creation, like `date`; re-size only if scope materially changes. It surfaces in `.context/index.yaml` so plans can be triaged by effort without opening each file.

## Grading value

Every `PLAN`, `FINDING`, and `KNOWN_ISSUE` with `status: DRAFT` or `ACTIVE` must carry a `value: HIGH|MEDIUM|LOW` frontmatter field — the benefit-of-action grade, distinct from `effort` (cost-of-action) and `severity` (risk-of-inaction). `validate-context-frontmatter.sh` enforces this; `DONE` and `SUPERSEDED` entries are exempt. Grade against the rubric in [`value-rubric.md`](value-rubric.md) — leverage, consumers unblocked, reversibility — rather than by gut feel. It surfaces in `.context/index.yaml`.

**Re-grade on transitions.** Unlike `date`, `value` can go stale as context changes. Revisit it when a plan moves `DRAFT → ACTIVE`, or when scope materially changes — the same discipline as the `ACTIVE → DONE` sync above, applied to the value axis.

**Read protocol — "what's next".** To pick the highest-value item to do next, read `.context/index.yaml`, filter to `DRAFT`/`ACTIVE` `PLAN`/`FINDING`/`KNOWN_ISSUE`, sort by `value` descending, then `effort` ascending where present, and act on the top item without re-forming an independent judgement. `value` is an authoritative sort key, not an advisory label; relocating the judgement to read-time reopens the gap the field closes. The full protocol and rubric live in [`value-rubric.md`](value-rubric.md).

## After merge

1. Delete the branch locally (GitHub auto-deletes remote branches after PR merge):
   ```
   git checkout main && git pull && git branch -d <branch-name>
   ```

2. Check whether the merged PR closes out any linked plan or ADR that's still `ACTIVE`/`DRAFT`/`proposed` — run this from any branch, not from `main`, since the script opens its own branch and PR when it has something to write:
   ```
   .context/plugins/pantheon-org/governance/adr-capture/scripts/merge-status-sync.sh --dry-run <pr-number>
   ```
   Single-phase plans directly or frontmatter-linked to the PR auto-flip to `DONE` via a branch + PR when run without `--dry-run`. Multi-phase plans and ADRs are always flagged, never auto-applied — ADR acceptance stays a deliberate, separate decision (see `adr-capture`'s `references/merge-status-sync.md`). This replaces the manual "did I forget to flip the ADR" reminder.
