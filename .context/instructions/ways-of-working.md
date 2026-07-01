---
title: "Ways of Working"
type: instruction
status: active
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
   Pre-push includes a plan-status drift check that warns about plans marked `active` for more than 60 days. If you see these warnings, update the plan's frontmatter `status: active → done` if the work is complete.

6. Push and open a PR:
   ```
   git push -u origin <branch-name>
   ```
   Use `gh pr create` or push and open via GitHub.

## Keeping plans in sync

When you implement what a plan describes, update its frontmatter `status: active → done` in the same PR. The `context-index` hook will regenerate `.context/index.yaml` automatically.

## After merge

Delete the branch locally (GitHub auto-deletes remote branches after PR merge):
```
git checkout main && git pull && git branch -d <branch-name>
```
