# Agent Rules

This file is the single source of truth for agent behavioural rules in this repository.
Rules are authored as entries with a title, directive, and rationale.

## Adding a rule

Use the `rules-management` skill to add new rules. Each rule MUST follow this schema:

```markdown
### Rule: <short imperative title>

**Directive:** <clear actionable instruction — prefer ALWAYS/NEVER phrasing>

**Rationale:** <why this rule exists — one or two sentences>
```

---

### Rule: Never advise outdated dependencies

**Directive:** NEVER recommend a dependency version without first verifying it is the latest stable release. Check the package's official registry or repository before suggesting it.

**Rationale:** Outdated dependency advice silently accumulates technical debt and security risk. The agent must not be the source of stale recommendations.

---

### Rule: Always validate newly created SKILL files

**Directive:** ALWAYS run both evals and audit on any newly created or modified SKILL.md before considering the task done. Run `./dist/skill-auditor eval ./cmd/assets --fail-below 0` for structural validation and `./dist/skill-auditor evaluate <path>` for scoring.

**Rationale:** A SKILL.md that fails structural validation or scores poorly defeats the purpose of the quality framework. Validating early catches issues when they're cheapest to fix.

