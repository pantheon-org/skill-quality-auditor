---
name: pr-author
description: >
  Create and maintain GitHub PRs with live descriptions — template discovery,
  intelligent filling, and lifecycle updates. Triggers on: PR, pull request,
  create PR, update PR description, draft PR, open PR, format PR, PR template,
  change request, re-request review, review feedback, force-push PR, PR ready
  for review.
---

# PR Author

Create and maintain GitHub PRs with live descriptions that stay in sync across the review cycle.

## Prerequisites

1. **Git remote configured** — `git remote -v` shows an `origin` with GitHub URL.
2. **`gh` CLI installed and authenticated** — run `gh auth status`. If not authenticated, prompt the user to run `gh auth login` before proceeding. Do not fail silently on auth errors.
3. **Upstream branch pushed** — the local branch exists on the remote (`git push -u origin <branch>`) before creating the PR.

## Quick Start

```bash
# Discover template, fill, and open PR
gh pr create --title "feat(scope): description" --body "$(cat filled-template.md)"
```
# Creates: a formatted PR with all template sections filled and no placeholders.

## When to Use

- Creating a new PR from a pushed branch
- Updating a PR description after new commits or a force-push
- Converting a draft PR to ready-for-review (only when the user explicitly asks)
- Responding to change requests by updating the PR body to reflect the *current state* of the change
- Re-requesting review after addressing all feedback

## When NOT to Use

- **Draft PRs where the user explicitly intends to write the description themselves** — do not override the user's manual intent. Only fill when the user asks you to create or update the PR.
- Trivial single-commit PRs that need no explanation beyond the commit message
- GitLab merge requests (GitHub-only for v0.1.0)

## Workflow

### 1. Discover — find the active PR template

Search in this order: `.github/PULL_REQUEST_TEMPLATE/` (directory), `.github/pull_request_template.md` (flat), `docs/PULL_REQUEST_TEMPLATE.md`, `PULL_REQUEST_TEMPLATE.md` (root), or fallback to built-in sections (Summary, Type of change, Checklist, Related issues).

**Multi-template selection:** If `.github/PULL_REQUEST_TEMPLATE/` contains multiple files, match by branch prefix: `feat/` → feature template, `fix/` → bug-fix template, `docs/` → docs template. Fallback to `default.md` or alphabetically first.

**HTML comments:** If the template contains `<!-- -->` guidance, **strip them** — they are noise once the agent populates the template.

### 2. Fill — generate the PR body from diff, commits, and linked issues

**Read the template** and fill section by section. For a flat template (`.github/pull_request_template.md`), read the file, fill it, then pass to `--body`. For directory templates (`.github/PULL_REQUEST_TEMPLATE/`), use `--template <filename>`.

**Correct `gh` invocation by template shape:**
- Flat template:
  ```bash
  gh pr create --title "..." --body "$(cat filled-template.md)"
  ```
  # assert: PR opens with flat template content passed to --body.
- Directory template:
  ```bash
  gh pr create --title "..." --template <filename>
  ```
  # assert: PR opens with directory template selected by name.
- **Anti-pattern:** NEVER use `--template` with a flat file — `gh` silently ignores it.

**PR Title Generation:**
- PREFER conventional-commit style (`feat(scope): description`) under 72 characters
- Summarize the *change* (e.g., "Add user auth middleware"), not the *task* ("Work on auth stuff")
- For single-commit PRs, default to the commit subject; for multi-commit, synthesize a higher-level summary

**Linked Issue Discovery:**
- Scan commits for `#NNN`, check branch name for issue refs (`fix/123-...`), look for auto-close syntax ("Fixes #NNN")
- Include the template's "Related Issues" section guidance

**Body Summarization (65536-character limit):**
- Use a bulleted list of files with one-line impact descriptions, not full diff hunks
- For >20 files, group by directory with a high-level theme
- Omit stack traces, build logs, or test output — link to CI artifacts
- If >40000 chars, truncate with "[... see full diff ...]" and rely on GitHub's "Files changed" tab

**Merge strategy note:** BY DEFAULT prefer squash for feature branches. TYPICALLY only use rebase when individual commits are meaningful. Document the intended strategy if the template calls for it.

**Postcondition verification:**
- Confirm the PR was created (`gh pr view --json title,body,number`)
- Verify body is under 65536 chars (`wc -c`) and no HTML comments remain

### 3. Update on push

1. Fetch current body (`gh pr view --json body`), regenerate from updated diff, update (`gh pr edit --body "..."`)
2. Detect force-push via `gh api repos/{owner}/{repo}/pulls/{number}/events --jq '.[] | select(.event=="force-pushed")'` or compare `git rev-list --count`
3. Only replace PR description — never overwrite comments (UNLESS the user explicitly asks)

**Postcondition:** Confirm under 65536 chars and comment count preserved:
```bash
wc -c <(gh pr view --json body) && gh pr view --json comments | jq 'length'
```
# output: character count and comment count displayed.

### 4. Respond to change requests

1. AVOID updating before feedback is addressed — only reflect the *current state* once changes are made (what was added/removed/fixed since last review)
2. Post conversational responses as PR **comments** (e.g., "Fixed in abc1234", "Disagree because...")
3. Never overwrite reviewer comments — body describes the *change*, comments describe the *discussion*

**Postcondition:** Confirm body reflects current state and responses are in comments, not the body

### 5. Re-request review

1. Add a summary comment listing changes since the last review
2. Use the API to re-request review:
   ```bash
   gh api repos/{owner}/{repo}/pulls/{number}/requested_reviewers \
     --method POST --field reviewers[]=username1 --field reviewers[]=username2
   ```
   # result: review re-requested from the specified reviewers.
3. If API is unavailable, tell the user to click "Re-request review" in the GitHub UI

### 6. Draft to ready

Only when the user explicitly asks — run `gh pr ready`, update body to remove draft indicators ("[WIP]", "Draft:"), and add missing sections. Never convert without explicit request.

## Mindset

The PR description is a **live document**, not a creation-time artifact. It should always reflect the *current state* of the change. After every push, rebase, or feedback round, check whether the description is still accurate. A stale description is worse than no description — it actively misleads reviewers.

## Anti-Patterns

**NEVER** leave a stale description after a force-push.
**SYMPTOM:** After a force-push, the PR body describes commits that no longer exist in the branch.
**CONSEQUENCE:** Reviewers compare against stale claims instead of the actual diff, eroding trust in the description.
**WHY:** Reviewers compare against the description first; force-push rewrites history so the existing description is guaranteed stale.
**BAD:** Pushing new commits without updating the PR body.
**GOOD:** After every force-push, regenerate the description from the current diff and update with `gh pr edit --body "..."`.

**NEVER** overwrite reviewer comments when updating the description.
**SYMPTOM:** The PR's comment thread is empty or has gaps after a body update.
**CONSEQUENCE:** Review context is lost; reviewers must re-raise points already discussed.
**WHY:** Loses conversation context and breaks the review thread.
**BAD:** Using `gh pr edit --body` in a way that replaces the entire PR including comments.
**GOOD:** Update only the description field; leave all PR comments untouched.

**NEVER** skip template discovery.
**WHY:** The repo convention exists for a reason — it ensures consistent PR structure and reviewer expectations.
**BAD:** Generating a PR body from scratch without checking for `.github/pull_request_template.md`.
**GOOD:** Always run the 5-path discovery check before filling the PR.

**NEVER** assume a single template format.
**WHY:** Repos vary — flat `.md` file, directory with multiple files, `docs/` location, etc.
**BAD:** Assuming every repo uses `.github/pull_request_template.md`.
**GOOD:** Check directory first, then flat file, then `docs/`, then root, then fallback.

**NEVER** use the PR body for back-and-forth review conversation.
**WHY:** The body describes the *change*, not the *discussion* — use comments for dialogue, body for state.
**BAD:** Adding "Done" or "Fixed" notes in the PR description.
**GOOD:** Post conversational responses as PR comments; keep the body focused on the current change state.

**NEVER** exceed GitHub's 65536-character PR body limit.
**WHY:** The API silently truncates; reviewers miss content at the end.
**BAD:** Dumping full `git diff` output into the PR body.
**GOOD:** Summarise with a bulleted list of files changed and one-line impact descriptions; link to CI for full logs.

**NEVER** use `--template` with a flat `.github/pull_request_template.md`.
**WHY:** `gh` silently ignores `--template` for flat files; the PR opens with an empty body.
**BAD:** `gh pr create --template pull_request_template.md` when the file is at `.github/pull_request_template.md`.
**GOOD:** Read the flat template file and pass it to `--body`: `gh pr create --body "$(cat filled-template.md)"`.

**NEVER** convert a draft PR to ready without explicit user request.
**WHY:** Draft status communicates intent; overriding it without consent disrupts the user's workflow.
**BAD:** Running `gh pr ready` because "it looks complete."
**GOOD:** Only convert to ready when the user explicitly asks: "make this ready for review."

## References

| Topic | Reference | When to Use |
| --- | --- | --- |
| Template lookup locations and precedence | `references/template-paths.md` | Debugging discovery failures in a new repo |
| What makes a good PR template | `references/pr-template-design.md` | Designing or evaluating a repo's PR template |
| Repo-specific PR template | `.github/pull_request_template.md` | The active template in the current repo |
| skip if no template exists | built-in sections | skip if discovery finds no template file |
| Commit message formatting | `commit-style` skill | When the agent needs to format commits for the PR |
| Native eval runner | `skill-quality-auditor` skill | When validating skill assets |
