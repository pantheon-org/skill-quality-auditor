# Scenario 01: Create a PR from a Branch

## User Prompt

"I've finished the feature on branch `feat/user-auth-middleware`. The branch is pushed. Create a PR for me."

## Input

- Branch: `feat/user-auth-middleware`
- Repo has `.github/pull_request_template.md` with sections: Summary, Type of change, Checklist, Related issues
- Recent commits:
  ```
  feat(auth): add JWT middleware for protected routes
  test(auth): add unit tests for JWT validation
  docs(auth): update README with auth setup instructions
  ```
- Diff summary: 3 files changed — `middleware/auth.go` (+120 lines), `middleware/auth_test.go` (+80 lines), `README.md` (+15 lines)
- No linked issues in commits or branch name

## Expected Behavior

1. Discover the template at `.github/pull_request_template.md` (flat file, second in precedence).
2. Read the template and strip any HTML comments.
3. Fill Summary with what the PR does (JWT middleware for protected routes) and why (secures API endpoints).
4. Check "New feature" in Type of change.
5. Fill Checklist with appropriate items (`go test ./...` passes, tests added, README updated, conventional commits used).
6. Leave Related issues empty or note "No linked issues" since none were found.
7. Generate a title under 72 chars: `feat(auth): add JWT middleware for protected routes`
8. Open the PR using `gh pr create --title "..." --body "$(cat filled-template.md)"` (not `--template`, since it's a flat file).

## Success Criteria

- Template discovered and read correctly.
- All sections filled; no placeholders or HTML comments remain.
- Title follows conventional-commit style and is under 72 characters.
- Correct `gh` CLI invocation for flat template (no `--template` flag).
- PR body is under 65536 characters.

## Failure Conditions

- Uses `--template` with a flat `.github/pull_request_template.md` (silently ignored by `gh`).
- Leaves unfilled placeholders (e.g., "<!-- What does this PR change? -->") in the body.
- Title exceeds 72 characters or does not follow conventional-commit style.
- PR body exceeds 65536 characters.
