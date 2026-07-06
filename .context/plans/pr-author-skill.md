---
title: "Plan: PR Author — formatted PR creation and update skill"
type: plan
status: done
date: 2026-07-03
value: medium
---

# Plan: PR Author Skill

## Goal

Create a new helper skill (`.context/plugins/pantheon-org/`) that teaches agents to create and maintain GitHub PRs with live descriptions — discovering the repo's existing template, filling it intelligently, and keeping it in sync across the review cycle.

## Motivation

GitHub provides the template. It does not provide the filling or the maintenance. PR descriptions drift stale after the first push, reviewers lose context after change requests, and agents lack a consistent workflow for both *creating* and *updating* a formatted PR. This skill plugs that gap.

## Related

- [ADR-029](docs/ADR/adr-029-local-skills-as-tessl-plugins.md) — governs plugin structure (tile.json, tessl-package.json, SCRIPT_DIR convention)
- [Adding Helper Skills](.context/instructions/adding-helper-skills.md) — canonical process for creating file-based Tessl plugins
- [Ways of Working](.context/instructions/ways-of-working.md) — branch naming, commit conventions, and CI gates the skill must align with

## Scope

- **In scope:** template discovery, description generation from diff + commits, description updates on push, change-request response patterns, re-request review flow, merge strategy notes, draft-to-ready conversion
- **Out of scope:** branch naming conventions (separate concern), commit message formatting (use the `commit-style` skill), CI pipeline management, code review content itself (the agent already does this)

## Steps

### Step 1: Create the skill directory and all required files

Following [ADR-029](docs/ADR/adr-029-local-skills-as-tessl-plugins.md) and the [Adding Helper Skills](.context/instructions/adding-helper-skills.md) guide, create the full plugin structure:

```
.context/plugins/pantheon-org/pr-author/
├── SKILL.md
├── tile.json
├── tessl-package.json
├── evals/
│   ├── instructions.json
│   ├── scenario-01/
│   │   └── (scenario files — create PR from branch)
│   ├── scenario-02/
│   │   └── (scenario files — update after force-push)
│   └── scenario-03/
│       └── (scenario files — respond to change requests)
└── references/
    ├── template-paths.md
    └── pr-template-design.md
```

**Note:** The original plan included a `scripts/discover-pr-template.sh` for template discovery. This was removed during review — template discovery is a simple file-existence check (5 paths) that the agent can perform inline. A shell script adds maintenance burden without value, unlike the complex Python-in-bash cross-referencing in `adr-capture/scripts/`. The discovery logic belongs in the SKILL.md workflow itself.

#### tessl-package.json

```json
{
  "name": "pantheon-org/pr-author",
  "version": "0.1.0"
}
```

#### tile.json

```json
{
  "name": "pantheon-org/pr-author",
  "version": "0.1.0",
  "summary": "Create and maintain GitHub PRs with live descriptions — template discovery, intelligent filling, and lifecycle updates",
  "private": false,
  "skills": {
    "pr-author": {
      "path": "SKILL.md"
    }
  }
}
```

#### SKILL.md

**Frontmatter:**
- `name: pr-author`
- `description:` — comprehensive with triggers, including "PR", "pull request", "create PR", "update PR description", "draft PR", "open PR", "format PR", "PR template", "change request", "re-request review"

**Sections:**
- Prerequisites (git remote, `gh` CLI available, upstream configured)
- Quick Start (one-liner to open a PR with template filling)
- When to Use (creating a PR, updating during review, after change requests, converting draft to ready)
- When NOT to Use (draft PRs the user intends to fill manually, trivial single-commit PRs)
- Workflow:
  1. **Discover** — find the active PR template — search order: `.github/PULL_REQUEST_TEMPLATE/` (directory with files), `.github/pull_request_template.md` (flat snake_case), `docs/PULL_REQUEST_TEMPLATE.md`, `PULL_REQUEST_TEMPLATE.md` (root-level), or fall back to built-in sections. Use `glob` or `ls` — no script needed.
  2. **Fill** — from diff summary, commit messages, linked issues — section by section. Use `gh pr create --title "..." --body "..."` or `gh pr create --template <file>` when a template exists. For large diffs, keep the body under GitHub's 65536-character limit — summarise, don't dump.
  3. **Update on push** — when new commits land, refresh description (added/removed scope, new screenshots). Use `gh pr edit --body "..."` to update. Detect force-push via GitHub's PR timeline (`gh api repos/{owner}/{repo}/pulls/{number}/events` — look for `force-pushed` event type) or by comparing `git rev-list --count` before/after.
  4. **Respond to change requests** — update the PR body to reflect the *current state* of the change (what was added/removed/fixed since the last review). Use PR **comments** for conversational responses to reviewers (e.g., "Done", "Fixed in abc1234", "Disagree because..."). Never overwrite reviewer comments.
  5. **Re-request** — after addressing all feedback, use `gh pr review --request-changes` or `gh api` to re-request review, with a change summary comment.
  6. **Draft to ready** — when converting a draft PR, use `gh pr ready` and update the body to remove draft indicators and add any missing sections now that the change is finalised.
- Mindset (the description is a live document, not a creation-time artifact)
- Anti-Patterns:
  - NEVER leave a stale description after a force-push (WHY: reviewers compare against description first; force-push rewrites history so the existing description is guaranteed stale)
  - NEVER overwrite reviewer comments when updating the description (WHY: loses conversation context)
  - NEVER skip template discovery (WHY: the repo convention exists for a reason)
  - NEVER assume a single template format (WHY: repos vary — flat `.md` file, directory with multiple files, `docs/` location, etc.)
  - NEVER use the PR body for back-and-forth review conversation (WHY: the body describes the *change*, not the *discussion* — use comments for dialogue, body for state)
  - NEVER exceed GitHub's 65536-character PR body limit (WHY: the API silently truncates; summarise large diffs instead of dumping full logs)
- References:
  | Topic | Reference | When to Use |
  | --- | --- | --- |
  | Template lookup locations and precedence | `references/template-paths.md` | Debugging discovery failures in a new repo |
  | What makes a good PR template | `references/pr-template-design.md` | Designing or evaluating a repo's PR template |
  | Repo-specific PR template | `.github/pull_request_template.md` | The active template in the current repo |
  | Commit message formatting | `commit-style` skill | When the agent needs to format commits for the PR |

#### evals/

Create three eval scenarios that exercise the core workflows. Follow the pattern from existing plugins: `instructions.json` with scenario metadata (`why_given`, `type`, `content`), per-scenario directories with `task.md`, `criteria.json`, and `capability.txt`.

| Scenario | Workflow exercised | Key criteria |
| --- | --- | --- |
| `scenario-01` | Create PR from branch — discover template, fill sections, open via `gh pr create` | Template found, all sections filled, no stale placeholders |
| `scenario-02` | Update description after force-push — detect force-push, refresh body | Stale content removed, new scope reflected, no reviewer comments overwritten |
| `scenario-03` | Respond to change requests — update body state, add review comment | Body reflects current state, comment used for dialogue, re-request triggered |

#### references/template-paths.md

Document all known PR template paths, the precedence order, and fallback behaviour:

1. `.github/PULL_REQUEST_TEMPLATE/` directory (one or more named templates)
2. `.github/pull_request_template.md` (single flat file — used by this repo)
3. `docs/PULL_REQUEST_TEMPLATE.md`
4. `PULL_REQUEST_TEMPLATE.md` (repo root)
5. Fallback: built-in sections (Summary, Type of change, Checklist, Related issues)

#### references/pr-template-design.md

Document what a good PR template looks like, so the skill can recommend or evaluate templates. This repo's template (`.github/pull_request_template.md`) serves as the reference implementation. Key principles:

**Structure:**
- **Summary** — One-paragraph "what and why", focused and scannable
- **Type of change** — Checklist with conventional-commit prefixes (`feat/`, `fix/`, `refactor/`, `docs/`, `chore/`)
- **Checklist** — Grouped by concern (code quality, skill/tile changes, documentation, commit conventions)
- **Merge strategy** — Explicit preference (squash/rebase/merge) with guidance on when to use each
- **Related issues** — Auto-close syntax (`Closes #NNN`), links to related PRs/ADRs/context files

**Quality signals:**
- Uses HTML comments (`<!-- -->`) for invisible guidance to human authors — these should be **stripped** by the agent when filling (see Open Question resolution #4)
- Checklist items are concrete and verifiable (not "code looks good" but "`go test ./...` passes")
- Groups related checks (e.g., all skill/tile checks together) so agents can skip irrelevant groups
- Includes repo-specific conventions (e.g., `./dist/skill-auditor eval`, `tessl status`, ADR requirements)
- Notes character limits (GitHub body: 65536 chars; title: 72 chars for list views)

**Anti-patterns to avoid in template design:**
- Vague checkboxes ("tests pass" → "`go test ./...` passes")
- Missing merge strategy guidance (creates ambiguity at merge time)
- No distinction between code changes and asset changes (different checks apply)
- Overly long default sections that encourage dumping rather than summarising

### Step 2: Register in tessl.json

Add to the `dependencies` block in `tessl.json` as a file-source dependency, maintaining alphabetical order (after `pantheon-org/docs-check`, before `pantheon-org/rules-management`). Note: file-source entries do **not** include a `"version"` field (matching the existing convention in `tessl.json`, despite `adding-helper-skills.md` suggesting otherwise):

```json
"pantheon-org/pr-author": {
  "source": "file:.context/plugins/pantheon-org/pr-author"
}
```

### Step 3: Run tessl install

```bash
tessl install
```

Verify the plugin was registered and synced:

```bash
tessl status
```

Confirm that `.claude/skills/tessl__pr-author/` (or equivalent Tessl-managed symlink in `.github/skills/`) was created.

### Step 4: Self-audit

Run the skill auditor with `--store` to persist results:

```bash
./dist/skill-auditor evaluate .context/plugins/pantheon-org/pr-author/SKILL.md --store
```

Iterate on diagnostics until the report is clean. Verify with `tessl status` that the plugin manifests are valid.

### Step 5: Branch, implement, commit

**Branch first, then implement.** All file creation and registration happens on the branch, never on `main`.

1. Pull latest `main` and branch: `git checkout main && git pull && git checkout -b feat/pr-author-skill`
2. Create all files from Step 1, register from Step 2, install from Step 3, audit from Step 4
3. Commit atomically: `git add .context/plugins/pantheon-org/pr-author/ tessl.json` then `git commit -m "feat(skills): add pr-author helper skill"`
4. Run `hk check && go test ./...` before pushing
5. Plan frontmatter: update `status: draft` → `status: active` when work begins, `active` → `done` when PR merges

## Resolved Decisions

| Question | Decision |
| --- | --- |
| Template paths: inline vs reference | **Reference** — `references/template-paths.md` |
| Discovery script | **Removed during review** — template discovery is a 5-path file-existence check, not warranting a shell script. The adr-capture scripts justify their existence with embedded Python cross-referencing logic; template discovery does not. |
| Tessl registry overlap check | **Checked.** `jbvc/create-pr` exists but is Sentry-specific, ignores templates, has no update lifecycle. No overlap with our approach. |
| `.github/skills/` convention | **Not needed** — Tessl install creates symlinks in both `.claude/skills/` and `.github/skills/` automatically (per ADR-029 #2). No manual entry required. |
| ADR-029 compliance | **Confirmed** — tile.json, tessl-package.json, and `file:` source registration all follow ADR-029 #1. |
| Evals required? | **Yes — 3 scenarios** (create, force-push update, change-request response). Matches existing plugin convention (socratic-method: 3, adr-capture: 3). |
| Force-push detection | **GitHub API** — use `gh api` PR timeline events (`force-pushed` event type) or `git rev-list --count` comparison. Reflog is local-only and unreliable for shared branches. |
| Branch naming for implementation | **`feat/pr-author-skill`** — per Ways of Working `feat/<short-description>` convention. Removed unnecessary `skills/` nesting. |
| tessl.json version field | **Omitted** — file-source entries in tessl.json do not include `"version"`, matching all 7 existing `pantheon-org/*` entries. The `adding-helper-skills.md` instruction doc incorrectly suggests including it; the plan follows actual convention. |
| PR body character limit | **65536 chars** — GitHub silently truncates beyond this. The skill must instruct agents to summarise large diffs rather than dump full logs. |
| Draft-to-ready workflow | **Included** — `gh pr ready` + body update to remove draft indicators. Common flow that was missing from the original plan. |

## Critical Review Findings

The following issues were identified during critical review and addressed in the amendments above:

### 1. Anti-pattern contradiction (fixed)

The original plan had a direct contradiction:
- **Anti-pattern:** "NEVER modify the PR body to respond to review feedback"
- **Workflow step 4:** "update description to reflect resolved vs outstanding feedback"

**Resolution:** The anti-pattern was poorly worded. The intent is: don't use the PR body as a *conversation thread* (back-and-forth with reviewers). The body should always reflect the *current state of the change*. Review dialogue belongs in PR comments. Reworded to: "NEVER use the PR body for back-and-forth review conversation."

### 2. Step ordering was misleading (fixed)

Steps 1–4 were described as sequential actions, then Step 5 said "branch first, then do Steps 1–4." This presentation implied Steps 1–4 happen on `main`. Restructured Step 5 to lead with "Branch first, then implement" and made the ordering explicit.

### 3. Only 1 eval scenario planned (fixed → 3)

Existing plugins consistently have 3 eval scenarios (socratic-method: 3, adr-capture: 3). One scenario is insufficient to exercise the three distinct workflows (create, update, respond). Expanded to 3 scenarios with a table mapping each to its workflow and key criteria.

### 4. No concrete `gh` CLI commands (fixed)

The original workflow was purely conceptual. An agent skill needs actionable commands. Added `gh pr create`, `gh pr edit --body`, `gh pr ready`, `gh pr review`, and `gh api` for force-push detection throughout the workflow steps.

### 5. Discovery script was over-engineered (removed)

`scripts/discover-pr-template.sh` would wrap a 5-path file-existence check. Unlike adr-capture's scripts (which embed Python for ADR cross-referencing), template discovery has no logic complexity. The discovery steps belong inline in the SKILL.md workflow. Removed from the directory structure and references table.

### 6. Branch name had unnecessary nesting (simplified)

`feat/skills/pr-author-helper-skill` → `feat/pr-author-skill`. The Ways of Working convention is `feat/<short-description>`, not `feat/<category>/<description>`. No other branches in the repo use a `skills/` subdirectory prefix.

### 7. Force-push detection was vague (improved)

"detect force-push via reflog or commit count drop" was hand-wavy. Reflog is local-only and unreliable for shared branches. Replaced with GitHub API approach (`gh api` PR timeline events looking for `force-pushed` event type) with `git rev-list --count` as a fallback.

### 8. Missing draft-to-ready workflow (added)

Converting a draft PR to ready-for-review is a common flow that was entirely absent. Added as workflow step 6 with `gh pr ready` and body update guidance.

### 9. Missing PR body character limit (added)

GitHub's 65536-character PR body limit is silently enforced (truncation, not error). Added as an anti-pattern and noted in the Fill workflow step.

### 10. Missing `commit-style` skill cross-reference (added)

The repo has a `commit-style` skill available. Added to the References table so the agent knows to use it when formatting commits for the PR.

### 11. `adding-helper-skills.md` version field discrepancy (documented)

The instruction doc says to include `"version": "0.1.0"` in tessl.json file-source entries, but all 7 existing entries omit it. The plan follows actual convention and documents the discrepancy in the Resolved Decisions table. Consider updating `adding-helper-skills.md` to match.

### 12. Multi-repo usage not addressed (noted)

The skill is described as repo-agnostic in the workflow, but the References table hardcodes "this repo's" template path. The current approach is acceptable — the reference points to the *current repo's* template as an example, and the discovery workflow handles other repos automatically. No change needed, but worth verifying during implementation.

---

## Additional Critical Review Findings (Post-Amendment)

The following issues were identified in a second-pass critical review of the amended plan. They must be addressed before implementation begins:

### 13. Re-request review command is incorrect (new — must fix)

The workflow step 5 states: "use `gh pr review --request-changes` or `gh api` to re-request review." This is wrong. `gh pr review --request-changes` is how a *reviewer* requests changes *on* a PR — it does not re-request review *from* reviewers. There is no native `gh` CLI command to re-request review from specific people. The correct approach is the GitHub API: `gh api repos/{owner}/{repo}/pulls/{number}/requested_reviewers` with a POST body containing the reviewer logins. The SKILL.md must be corrected to use the API method explicitly, or simply instruct the agent to click "Re-request review" in the GitHub UI when using `gh` is not possible.

**Fix:** Replace `gh pr review --request-changes` in workflow step 5 and the references table with the correct `gh api` endpoint and method.

### 14. `gh pr create --template <file>` is misleading for flat templates (new — must fix)

The Fill workflow says: "Use `gh pr create --title "..." --body "..."` or `gh pr create --template <file>` when a template exists." GitHub CLI's `--template` flag only works for templates inside `.github/PULL_REQUEST_TEMPLATE/` (you pass the *filename* without path). For a flat `.github/pull_request_template.md` (the most common case, and the one used by this repo), `--template` does not work — `gh` ignores it. The agent skill needs to distinguish:
- **Directory templates** (`.github/PULL_REQUEST_TEMPLATE/*.md`): use `--template <filename>`
- **Flat template** (`.github/pull_request_template.md`): read the file content and pass it to `--body` (or let `gh` auto-detect it, which it does when no `--body`/`--template` is provided and the user is in a terminal)

Since agents are non-interactive, the skill should instruct: read the template file, fill it, then use `--body "$(cat filled-template.md)"`.

**Fix:** Update the Fill workflow step to clarify the two template shapes and the correct `gh` invocation for each. Add an anti-pattern: "NEVER use `--template` with a flat `.github/pull_request_template.md` — it is silently ignored."

### 15. Missing `evals/summary.json` and `capability.txt` in eval structure (new — must fix)

The plan's eval directory structure mentions `instructions.json`, `task.md`, and `criteria.json`, but omits two files present in every existing plugin's evals:
- **`evals/summary.json`** — contains `instructions_coverage.coverage_percentage` (always `100` in existing plugins)
- **`scenario-N/capability.txt`** — a one-line capability identifier (e.g., `adr-capture`, `socratic-method`)

Without these, the eval structure does not match the established convention, and Tessl eval indexing may fail.

**Fix:** Add `summary.json` and `capability.txt` to the directory tree and the eval creation guidance. The capability.txt should contain `pr-author` for all three scenarios.

### 16. No guidance on PR title generation (new — must fix)

The workflow mentions `--title "..."` but gives no instruction on *how* to generate the title. A bad title is as damaging as a bad body. The skill should include:
- Use conventional-commit style (`feat(scope): description` or `fix(scope): description`) if the repo uses it
- Keep under 72 characters (GitHub truncates in list views)
- Summarize the *change*, not the *task* (e.g., "Add user authentication middleware" not "Work on auth stuff")
- For single-commit PRs, default to the commit subject line (after running the `commit-style` skill if applicable)
- For multi-commit PRs, synthesize a higher-level summary

**Fix:** Add a "PR Title Generation" subsection under the Fill workflow step, or as a standalone mini-section in SKILL.md.

### 17. Missing multi-template selection logic (new — must fix)

If `.github/PULL_REQUEST_TEMPLATE/` contains multiple files (e.g., `bug-fix.md`, `feature.md`, `release.md`), the agent needs guidance on which to choose. The current discovery step only says "directory with files" but gives no selection heuristic.

**Fix:** Add to the Discover workflow step: if multiple templates exist, prefer the one whose filename best matches the branch prefix or change type (`feat/` → feature template, `fix/` → bug-fix template), falling back to the alphabetically first or a default `default.md` if present.

### 18. `go test ./...` in Step 5 is cargo-culted and irrelevant (new — must fix)

Step 5 says: "Run `hk check && go test ./...` before pushing." This PR contains only Tessl plugin assets (SKILL.md, tile.json, evals, references) — there is no Go code. `go test ./...` will pass trivially but wastes CI time and confuses the purpose of the check. The `hk check` step may also not apply (`.context/plugins/` is excluded from pre-commit hooks per ADR-029 #4).

**Fix:** Replace with: "Run `./dist/skill-auditor evaluate .context/plugins/pantheon-org/pr-author` for scoring and `./dist/skill-auditor eval ./cmd/assets --fail-below 0` if any `cmd/assets/` files were modified. Verify no `tessl status` warnings before pushing."

### 19. Missing `gh auth status` prerequisite (new — must fix)

The Prerequisites section says "`gh` CLI available" but does not mention authentication. An unauthenticated `gh` CLI will fail on every command with a confusing error. The skill should instruct the agent to verify auth first, and if not authenticated, guide the user to run `gh auth login` rather than failing silently.

**Fix:** Add to Prerequisites: "Run `gh auth status` — if not authenticated, prompt the user to run `gh auth login` before proceeding."

### 20. Missing `.context/index.yaml` regeneration step (new — must fix)

Per Ways of Working and the `context-index` skill, creating a new plan should trigger index regeneration. The plan creates a new `.context/plans/` file but never mentions updating `.context/index.yaml`. While the pre-commit hook regenerates it automatically, the plan should explicitly include this step so the agent doesn't miss it if hooks are bypassed.

**Fix:** Add a step after file creation: "Regenerate `.context/index.yaml` by running the `context-index` skill or `hk run pre-commit` (which includes the context-index hook)."

### 21. Out-of-scope draft PR criterion creates tension (new — must fix)

"When NOT to Use" says: "draft PRs the user intends to fill manually" — but "When to Use" says: "converting draft to ready." This creates ambiguity: if the skill should not be used for draft PRs at all, why does it have a draft-to-ready workflow? The intent is that the skill should not be used when the user opens a draft PR *and explicitly says* "I'll fill this in later myself." But when the user says "make this draft ready for review," the skill is absolutely in scope.

**Fix:** Reword the out-of-scope entry to: "Draft PRs where the user explicitly intends to write the description themselves — do not override the user's manual intent." And add a note in the draft-to-ready workflow: "Only convert to ready when the user explicitly asks."

### 22. Missing linked-issue discovery guidance (new — should fix)

The Fill step mentions "linked issues" as an input but provides no method for discovering them. Agents need concrete instructions:
- Scan commit messages for `#NNN` references
- Check the branch name for issue references (`fix/123-...`, `feature/456-...`)
- Look at the PR template's "Related Issues" section for guidance on how the repo links issues
- If the repo uses GitHub's "Fixes #NNN" auto-close syntax, include it in the body

**Fix:** Add a "Linked Issue Discovery" subsection under Fill, or inline bullet points.

### 23. Body summarization guidance is vague (new — should fix)

The anti-pattern says "summarise, don't dump" for the 65536-char limit, but doesn't say *how*. Agents will struggle with "summarise" as an instruction. Concrete guidance:
- Use a bulleted list of files changed with one-line impact descriptions, not full diff hunks
- For >20 files changed, group by directory and give a high-level theme
- Omit full stack traces, build logs, or test output — link to CI artifacts instead
- If the diff summary alone exceeds 40000 chars, truncate with a "[... see full diff ...]" note and rely on the GitHub "Files changed" tab

**Fix:** Add these specifics to the Fill workflow step and the 65536-char anti-pattern.

### 24. Merge strategy notes are in scope but never addressed (new — noted)

Scope says "merge strategy notes" are in scope, but no workflow step or anti-pattern covers merge strategies (squash, rebase, merge). This is acceptable to defer — the primary concern is PR creation and maintenance, not merge execution — but it should be explicitly noted as a future enhancement or removed from scope to avoid scope creep.

**Fix:** Either remove "merge strategy notes" from the scope list, or add a lightweight workflow step: "Document the intended merge strategy (squash/rebase/merge) if the repo convention or PR template calls for it."

---

## Open Questions for Implementation

| Question | Current Stance | When to Resolve |
| --- | --- | --- |
| Should `adding-helper-skills.md` be updated to remove the incorrect `"version"` field from file-source examples? | Yes — file a follow-up chore | Before or during implementation PR |
| Should the skill support GitLab PRs (merge requests) as well as GitHub? | No — GitHub-only for v0.1.0 | If user demand arises, scope v0.2.0 |
| Should the skill include a `scripts/` directory for anything (e.g., a `fill-pr-template.py` helper)? | No — keep it script-less like rules-management | Revisit if template-filling logic grows beyond markdown substitution |
| How should the agent handle PR templates that contain HTML comments (`<!-- -->` used as invisible instructions)? | Strip or preserve? — preserve them as they guide human reviewers | Verify during SKILL.md drafting |
