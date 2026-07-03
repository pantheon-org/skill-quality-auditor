# Scenario 03: Respond to Change Requests

## User Prompt

"I've addressed the review feedback on PR #42. Update the PR and let the reviewers know it's ready for another look."

## Input

- PR #42, branch `feat/user-auth-middleware`
- Reviewer feedback:
  - "Please add rate limiting to the JWT middleware"
  - "The tests don't cover the expired token case"
  - "Document the env var requirements in the README"
- Changes made since review:
  - Added rate limiting in `middleware/auth.go`
  - Added expired token test in `middleware/auth_test.go`
  - Documented `JWT_SECRET` and `GITHUB_CLIENT_ID` in `README.md`

## Expected Behavior

1. Update the PR body to reflect the current state:
   - Mention rate limiting is now included
   - Note expired token test coverage added
   - Document env vars in the change summary
2. Add a PR **comment** (not in the body) addressing each feedback item:
   - "Rate limiting added in commit abc1234"
   - "Expired token test added in commit def5678"
   - "Env var docs updated in README — see commit ghi9012"
3. Re-request review via GitHub API:
   ```bash
   gh api repos/{owner}/{repo}/pulls/42/requested_reviewers \
     --method POST \
     --field reviewers[]=reviewer-username
   ```
4. If the API fails, instruct the user to click "Re-request review" in the GitHub UI.

## Success Criteria

- PR body updated to reflect all three feedback items as resolved.
- Conversational responses posted as PR comments, not in the body.
- Review re-requested via API or UI guidance provided.
- No existing reviewer comments modified or deleted.

## Failure Conditions

- Puts conversational responses ("Done", "Fixed") in the PR body instead of comments.
- Does not re-request review.
- Overwrites existing reviewer comments.
- PR body exceeds 65536 characters.
