---
name: rules-management
description: "Manage agent behavioural rules in `.agents/RULES.md`. Use when the user says 'new rule', 'add a rule', 'record that', 'make a rule that', or asks to codify an agent directive. Load `.agents/RULES.md` to check existing rules before adding a new one. DO NOT use for ephemeral notes, one-off instructions, or instructions that belong in AGENTS.md itself. Triggers: 'new rule', 'add a rule', 'record that', 'make a rule', 'codify this', 'rule about', 'create a directive'."
---

# Rules Management

Manage project-level agent behavioural rules.

## Prerequisites

- The `.agents/RULES.md` file must exist in the repository root. If it is missing, create it with a `# Agent Rules` heading first.
- Only invoke this skill when the user explicitly requests a new rule — do not infer rule creation from unrelated conversation.
- Depends on the user providing a clear behavioural directive; if the instruction is vague, ask for clarification before writing.

## Quick Start

```bash
# Read current rules
cat .agents/RULES.md
```
→ Produces the full current ruleset for duplicate checking.

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
1. Read the existing `.agents/RULES.md` first — check every `### Rule:` heading to avoid duplicates
2. If a rule with the same directive already exists, inform the user and do NOT create a duplicate
3. Append the new rule following the format below
4. Inform the user you've added it
5. Confirm by reading the updated file

## Rule Format

Every rule entry MUST include:

```markdown
### Rule: <short imperative title>

**Directive:** <clear actionable instruction — prefer ALWAYS/NEVER phrasing>

**Rationale:** <why this rule exists — one or two sentences>
```

After appending, run `cat .agents/RULES.md` to verify the entry appears correctly and the file still has valid structure.

## When NOT to Use

- Ephemeral notes or one-off instructions — use a context file or tell the user directly
- Instructions that belong in AGENTS.md (repo map, workflow, tool conventions)
- Personal preferences that don't affect agent behaviour
- "Remember this" or "save this for later" — route to vault-capture instead

## Anti-Patterns

### NEVER: Add a rule without reading existing rules first

**WHY:** Blind appending creates duplicate or conflicting directives, violating the single-source-of-truth contract and confusing future agents.

```markdown
BAD — skips read step, appends blindly:
    readFile(".agents/RULES.md", append=true)
```

```markdown
GOOD — reads first, appends only after confirming no overlap:
    rules = readFile(".agents/RULES.md")
    if hasDuplicate(rules, userDirective):
        informUser("This rule already exists")
    else:
        append(rules, newEntry)
```

### NEVER: Accept a vague user instruction as the rule body

**WHY:** A rule like "use good logging" is not actionable. Every rule must have a precise ALWAYS/NEVER directive and a rationale.

```markdown
BAD — copied verbatim, no structure:
    - Don't use bad logging
```

```markdown
GOOD — reformatted with the rule format:
    ### Rule: Always use structured logging
    **Directive:** ALWAYS use structured logging (JSON) for production services.
    **Rationale:** Structured logs enable aggregation, search, and alerting.
```

### NEVER: Create a rule for memory or ephemeral notes

**WHY:** Rules are behavioural directives binding all agents. Ephemeral notes belong in a context file or vault, not in `.agents/RULES.md`.

**CONSEQUENCE:** RULES.md grows unbounded with noise, reducing signal-to-noise ratio and causing agents to skip reading it.

## Mindset

- Rules are binding for all agents in this repository — be precise and unambiguous
- Prefer ALWAYS/NEVER phrasing for directives (e.g., "NEVER recommend a dependency without checking the registry")
- Always include a rationale so future readers understand why the rule exists
- Read existing rules before adding — no duplicates
- Rules should be few and high-signal — avoid documenting the obvious
- If the user asks for a rule that overlaps with an existing one, suggest amending rather than duplicating
- PREFER merging related rules into a single entry TYPICALLY over scattering them across multiple entries
