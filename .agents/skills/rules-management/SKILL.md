---
name: rules-management
description: "Manage agent behavioural rules in `.agents/RULES.md`. Use when the user says 'new rule', 'add a rule', 'record that', 'make a rule that', or asks to codify an agent directive. Load `.agents/RULES.md` to check existing rules before adding a new one. DO NOT use for ephemeral notes, one-off instructions, or instructions that belong in AGENTS.md itself. Triggers: 'new rule', 'add a rule', 'record that', 'make a rule', 'codify this', 'rule about', 'create a directive'."
---

# Rules Management

Manage project-level agent behavioural rules in `.agents/RULES.md`.

## Quick Start

```bash
# Read current rules
cat .agents/RULES.md
```

## Where Rules Live

All agent rules reside in a single file: `.agents/RULES.md`. This is the authoritative source that all agents read before acting in this repository.

## When to Use

Create a new rule whenever the user says something like:
- "New rule: ..."
- "Add a rule that ..."
- "Record that ..."
- "Make a rule about ..."
- "Codify this as a rule"

The user's instruction becomes the rule body. You MUST:
1. Read the existing `.agents/RULES.md` first to avoid duplicates
2. Append the new rule following the schema below
3. Inform the user you've added it

## Schema

Every rule entry MUST include:

```markdown
### Rule: <short imperative title>

**Directive:** <clear actionable instruction — prefer ALWAYS/NEVER phrasing>

**Rationale:** <why this rule exists — one or two sentences>
```

## When NOT to Use

- Ephemeral notes or one-off instructions — use a context file or tell the user directly
- Instructions that belong in AGENTS.md (repo map, workflow, tool conventions)
- Personal preferences that don't affect agent behaviour

## Mindset

- Rules are binding for all agents in this repository — be precise and unambiguous
- Prefer ALWAYS/NEVER phrasing for directives (e.g., "NEVER recommend a dependency without checking the registry")
- Always include a rationale so future readers understand why the rule exists
- Read existing rules before adding — no duplicates
- Rules should be few and high-signal — avoid documenting the obvious
