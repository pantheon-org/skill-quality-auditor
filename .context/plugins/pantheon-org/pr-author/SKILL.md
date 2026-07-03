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
# Discover template, fill from diff + commits, open PR
gh pr create --title "feat(scope): concise description" --body "$(cat filled-template.md)"
```

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

Search in this order (use `glob` or `ls` — no script needed):

1. `.github/PULL_REQUEST_TEMPLATE/` — directory with one or more `.md` files
2. `.github/pull_request_template.md` — single flat file
3. `docs/PULL_REQUEST_TEMPLATE.md`
4. `PULL_REQUEST_TEMPLATE.md` (repo root)
5. **Fallback:** use built-in sections (Summary, Type of change, Checklist, Related issues)

**Multi-template selection:** If `.github/PULL_REQUEST_TEMPLATE/` contains multiple files (e.g., `bug-fix.md`, `feature.md`, `release.md`), prefer the one whose filename best matches the branch prefix or change type:
- `feat/` branch → `feature.md` or similar
- `fix/` branch → `bug-fix.md` or similar
- `docs/` branch → `docs.md` or similar
- Fallback: `default.md` if present, otherwise alphabetically first file.

**HTML comments:** If the template contains `<!-- -->` guidance for human authors, **strip them** when filling — they are noise once the agent has populated the template.

### 2. Fill — generate the PR body from diff, commits, and linked issues

**Read the template** and fill section by section. For a flat template (`.github/pull_request_template.md`), read the file, fill it, then pass to `--body`. For directory templates (`.github/PULL_REQUEST_TEMPLATE/`), use `--template <filename>`.

**Correct `gh` invocation by template shape:**

- **Flat template:** `gh pr create --title "..." --body "$(cat filled-template.md)"`
- **Directory template:** `gh pr create --title "..." --template <filename>`

> **Anti-pattern:** NEVER use `--template` with a flat `.github/pull_request_template.md` — `gh` silently ignores it. Always read flat templates and pass them to `--body`.

**PR Title Generation:**

- Use conventional-commit style (`feat(scope): description` or `fix(scope): description`) if the repo uses it
- Keep under 72 characters (GitHub truncates in list views)
- Summarize the *change*, not the *task* (e.g., "Add user authentication middleware" not "Work on auth stuff")
- For single-commit PRs, default to the commit subject line (run the `commit-style` skill first if applicable)
- For multi-commit PRs, synthesize a higher-level summary that captures the overall intent

**Linked Issue Discovery:**

- Scan commit messages for `#NNN` references
- Check the branch name for issue references (`fix/123-...`, `feature/456-...`)
- Look at the PR template's "Related Issues" section for guidance on how the repo links issues
- If the repo uses GitHub's "Fixes #NNN" auto-close syntax, include it in the body

**Body Summarization (65536-character limit):**

GitHub silently truncates PR bodies beyond 65536 characters. When the diff is large:

- Use a bulleted list of files changed with one-line impact descriptions, not full diff hunks
- For >20 files changed, group by directory and give a high-level theme (e.g., "`scorer/` — D3 and D7 scorer updates")
- Omit full stack traces, build logs, or test output — link to CI artifacts instead
- If the diff summary alone exceeds 40000 chars, truncate with a "[... see full diff ...]" note and rely on the GitHub "Files changed" tab

**Merge strategy note:** If the repo convention or PR template calls for it, document the intended merge strategy (squash/rebase/merge). Most repos prefer squash for feature branches; only use rebase when individual commits are meaningful.

**Postcondition verification:**
- Confirm the PR was created with `gh pr view --json title,body,number`
- Verify the body is under 65536 characters (`wc -c` on the filled template)
- Check that no HTML comments remain in the submitted body

### 3. Update on push — refresh the description when new commits land

When the user says "I pushed more commits" or "update the PR":

1. Fetch the current PR body: `gh pr view --json body`
2. Regenerate the description from the updated diff and commit history
3. Update with: `gh pr edit --body "..."`

**Detecting force-push:**
- Use `gh api repos/{owner}/{repo}/pulls/{number}/events --jq '.[] | select(.event=="force-pushed")'`
- Fallback: compare `git rev-list --count` before/after

**Preserve reviewer comments:** When updating the body, only replace the PR description content — never overwrite PR comments. Comments are conversation threads; the body is the live document describing the change.

**Postcondition verification:**
- Confirm the updated body is under 65536 characters
- Verify no reviewer comments were modified (compare comment count before/after with `gh pr view --json comments`)

### 4. Respond to change requests — update body state, use comments for dialogue

When the user says "addressed the review feedback" or similar:

1. Update the PR body to reflect the *current state* of the change (what was added/removed/fixed since the last review)
2. Add a PR **comment** for conversational responses (e.g., "Fixed in abc1234", "Done", "Disagree because...")
3. Never overwrite reviewer comments

> **Anti-pattern:** NEVER use the PR body for back-and-forth review conversation. The body describes the *change*; comments describe the *discussion*.

**Postcondition verification:**
- Confirm the PR body reflects the current state of all feedback items
- Verify conversational responses were posted as comments, not in the body

### 5. Re-request review — after addressing all feedback

When all feedback is addressed and the PR is ready for another round:

1. Add a summary comment listing what was changed since the last review
2. Re-request review via GitHub API:
   ```bash
   gh api repos/{owner}/{repo}/pulls/{number}/requested_reviewers \
     --method POST \
     --field reviewers[]=username1 \
     --field reviewers[]=username2
   ```
3. If the API is unavailable, instruct the user to click "Re-request review" in the GitHub UI

### 6. Draft to ready — convert when explicitly asked

When the user explicitly asks to make a draft PR ready for review:

1. Run `gh pr ready`
2. Update the body to remove draft indicators (e.g., "[WIP]", "Draft:") and add any missing sections now that the change is finalized
3. Do not convert without explicit user request

## Mindset

The PR description is a **live document**, not a creation-time artifact. It should always reflect the *current state* of the change. After every push, rebase, or feedback round, check whether the description is still accurate. A stale description is worse than no description — it actively misleads reviewers.

## Anti-Patterns

**NEVER leave a stale description after a force-push.**
**WHY:** Reviewers compare against the description first; force-push rewrites history so the existing description is guaranteed stale.
**BAD:** Pushing new commits without updating the PR body.
**GOOD:** After every force-push, regenerate the description from the current diff and update with `gh pr edit --body "..."`.

**NEVER overwrite reviewer comments when updating the description.**
**WHY:** Loses conversation context and breaks the review thread.
**BAD:** Using `gh pr edit --body` in a way that replaces the entire PR including comments.
**GOOD:** Update only the description field; leave all PR comments untouched.

**NEVER skip template discovery.**
**WHY:** The repo convention exists for a reason — it ensures consistent PR structure and reviewer expectations.
**BAD:** Generating a PR body from scratch without checking for `.github/pull_request_template.md`.
**GOOD:** Always run the 5-path discovery check before filling the PR.

**NEVER assume a single template format.**
**WHY:** Repos vary — flat `.md` file, directory with multiple files, `docs/` location, etc.
**BAD:** Assuming every repo uses `.github/pull_request_template.md`.
**GOOD:** Check directory first, then flat file, then `docs/`, then root, then fallback.

**NEVER use the PR body for back-and-forth review conversation.**
**WHY:** The body describes the *change*, not the *discussion* — use comments for dialogue, body for state.
**BAD:** Adding "Done" or "Fixed" notes in the PR description.
**GOOD:** Post conversational responses as PR comments; keep the body focused on the current change state.

**NEVER exceed GitHub's 65536-character PR body limit.**
**WHY:** The API silently truncates; reviewers miss content at the end.
**BAD:** Dumping full `git diff` output into the PR body.
**GOOD:** Summarise with a bulleted list of files changed and one-line impact descriptions; link to CI for full logs.

**NEVER use `--template` with a flat `.github/pull_request_template.md`.**
**WHY:** `gh` silently ignores `--template` for flat files; the PR opens with an empty body.
**BAD:** `gh pr create --template pull_request_template.md` when the file is at `.github/pull_request_template.md`.
**GOOD:** Read the flat template file and pass it to `--body`: `gh pr create --body "$(cat filled-template.md)"`.

**NEVER convert a draft PR to ready without explicit user request.**
**WHY:** Draft status communicates intent; overriding it without consent disrupts the user's workflow.
**BAD:** Running `gh pr ready` because "it looks complete."
**GOOD:** Only convert to ready when the user explicitly asks: "make this ready for review."

## References

| Topic | Reference | When to Use |
| --- | --- | --- |
| Template lookup locations and precedence | `references/template-paths.md` | Debugging discovery failures in a new repo |
| What makes a good PR template | `references/pr-template-design.md` | Designing or evaluating a repo's PR template |
| Repo-specific PR template | `.github/pull_request_template.md` | The active template in the current repo |
| Commit message formatting | `commit-style` skill | When the agent needs to format commits for the PR |
| Native eval runner | `skill-quality-auditor` skill | When validating skill assets |
