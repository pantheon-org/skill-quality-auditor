# Scenario 02: Update Description After Force-Push

## User Prompt

"I force-pushed to the PR branch. Can you update the PR description?"

## Input

- Existing PR #42, branch `feat/user-auth-middleware`
- Original description mentioned 3 files changed, JWT middleware, no external dependencies
- Force-push rewrote history: now 5 files changed, added OAuth2 provider integration, requires `GITHUB_CLIENT_ID` env var
- Reviewer comment exists: "Please add OAuth2 support before merging"

## Expected Behavior

1. Detect force-push via `gh api repos/{owner}/{repo}/pulls/42/events` looking for `force-pushed` event type.
2. Regenerate the description from the new diff and commit history.
3. Update the body to reflect the new scope:
   - 5 files changed (not 3)
   - OAuth2 provider integration added
   - New env var requirement documented
4. Use `gh pr edit --body "..."` to update only the description.
5. Leave the reviewer comment "Please add OAuth2 support before merging" untouched.

## Success Criteria

- Force-push detected via GitHub API (not reflog).
- Description updated to reflect the new 5-file scope and OAuth2 integration.
- Reviewer comment is preserved.
- Body is under 65536 characters.

## Failure Conditions

- Does not detect the force-push (assumes normal push).
- Overwrites or deletes the reviewer comment.
- Leaves stale claims (e.g., still says "3 files changed" or "no external dependencies").
- PR body exceeds 65536 characters.
