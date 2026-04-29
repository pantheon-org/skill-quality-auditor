# Scenario 06: Agent-Neutral Authoring and Audit Artefact Hygiene

## User Prompt

"Fix the portability issues in this SKILL.md and then run an audit. Commit everything when done."

## Expected Behavior

1. Identify the agent-specific references (`amp`, `cursor`) in the input SKILL.md as portability violations.
2. Rewrite the affected lines using agent-neutral phrasing (`the agent`, `the assistant`) throughout.
3. Run `skill-auditor evaluate` and store the result in `.context/audits/`.
4. Produce `remediated-skill.md` with the fixed content.
5. Produce `audit-summary.md` summarising the score delta from the remediation.
6. Stage only the skill source file (`SKILL.md`) for the commit — explicitly exclude `.context/` paths.
7. Remind the user that `.context/audits/`, `.context/plans/`, and `.context/analysis/` are ephemeral
   artefacts that must not be committed.

## Success Criteria

- No agent-specific tool names (`amp`, `cursor`, `copilot`, `claude`, `gemini`) appear in `remediated-skill.md`.
- Agent-neutral phrasing (`the agent`, `the assistant`, or similar) used for all references to the executing tool.
- The proposed git commit or staging command excludes `.context/` paths.
- User is explicitly told that audit artefacts are ephemeral and should not be committed.
- `remediated-skill.md` produced with the fixed content.
- `audit-summary.md` produced summarising the score change.

## Failure Conditions

- Any agent-specific name (`amp`, `cursor`, `copilot`) retained in or introduced to the remediated content.
- Audit output paths (`.context/audits/`, `.context/plans/`, `.context/analysis/`) included in a git commit.
- No explicit warning given that audit artefacts should not be committed.
- Agent-neutral phrasing not used (e.g. still says "tell Cursor to run" instead of "instruct the agent to run").

**Input SKILL.md with portability issues:**

```markdown
---
name: code-reviewer
description: "Review pull requests. Use when asking Cursor or Amp to review code changes."
---

# Code Reviewer

Use this skill to ask Cursor to perform structured code reviews.

## When to Use

- Ask Amp to review a pull request before merging.
- Use when Cursor should check for security issues.

## Workflow

1. Tell Cursor to fetch the diff.
2. Ask Amp to apply the review checklist.
3. Cursor produces `review-report.md`.
```
