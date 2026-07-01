---
title: "Ways of Working"
type: instruction
status: active
date: 2026-07-01
---

# Ways of Working

## Branch workflow

1. Create a feature/fix branch from `main`:
   ```
   git checkout main && git pull && git checkout -b <type>/<short-description>
   ```
   Use conventional prefixes: `feat/`, `fix/`, `docs/`, `refactor/`, `chore/`.

2. Commit as you go — small, atomic commits with conventional messages:
   ```
   feat(scorer): add D9 mutation coverage
   fix(hook): regenerate index on plan changes
   docs: update README install section
   ```

3. If `main` has diverged, rebase instead of merging:
   ```
   git fetch origin && git rebase origin/main
   ```
   This keeps history linear. Resolve conflicts if they arise.

4. Run checks before pushing:
   ```
   hk check          # pre-commit checks (lint, validate, index freshness)
   go test ./...     # full test suite
   ```
   Pre-push includes a plan-status drift check that warns about plans marked `active` for more than 60 days. If you see these warnings, update the plan's frontmatter `status: active → done` if the work is complete.

5. Push and open a PR:
   ```
   git push -u origin <branch-name>
   ```
   Use `gh pr create` or push and open via GitHub.

## Keeping plans in sync

When you implement what a plan describes, update its frontmatter `status: active → done` in the same PR. The `context-index` hook will regenerate `.context/index.yaml` automatically.

## After merge

Delete the branch locally and remotely:
```
git checkout main && git pull && git branch -d <branch-name>
```
