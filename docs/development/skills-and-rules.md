# Skills and rules

This repository maintains agent behavioural rules and local skills that guide AI agents working in this codebase. They live under `.agents/` and are installed via the `init` command from the canonical source at `cmd/assets/`.

## Agent rules

Located in [`.agents/RULES.md`](https://github.com/pantheon-org/skill-quality-auditor/blob/main/.agents/RULES.md) — the single source of truth for agent behavioural directives.

| # | Rule | Directive |
|---|------|-----------|
| 1 | Never advise outdated dependencies | ALWAYS verify latest stable release before recommending a dependency |
| 2 | Always validate newly created SKILL files | ALWAYS run evals and audit on any new or modified `SKILL.md` |
| 3 | Allow skills to evolve with experience | NEVER treat a skill as final; ALWAYS revisit and improve |
| 4 | Formalise ad hoc scripts after repeated use | ALWAYS use `.tmp/` for one-off scripts; formalise after 2nd use |
| 5 | Never write temporary scripts outside `.tmp` | NEVER write ad hoc scripts outside `.tmp/` |

To add a new rule, load the [`rules-management`](#rules-management) skill — it will load existing rules, check for duplicates, and append with the correct format.

## Local skills

Local helper skills are installed under `.context/plugins/pantheon-org/` (managed via `tessl.json`). Registry plugins live under `.tessl/plugins/`. The following skills are available:

### adr-capture

Capture Architecture Decision Records (ADRs) from `.context/` plans, findings, and analyses. Extracts binding decisions into `docs/ADR/` and maintains the machine-readable index.

### context-file

Create `.context/` files (plans, findings, analysis) with standard YAML frontmatter, appropriate sections, and correct placement in `plans/`, `findings/`, or `analysis/`.

### context-index

Regenerate [`.context/index.yaml`](https://github.com/pantheon-org/skill-quality-auditor/blob/main/.context/index.yaml) from all `.context/**/*.md` frontmatter and validate that all files carry the required frontmatter block.

### docs-check

Validate the GitHub Pages documentation site built by docmd — orphan detection, ADR index freshness, build verification, and LLM output audit.

### rules-management

Manage the agent behavioural rules in `.agents/RULES.md` — load existing rules, check for duplicates, and append new rules in the standard format with directive and rationale.

### tessl__eval-improve

Analyse eval results, diagnose failures, apply targeted fixes, and re-run to verify improvements. Used when debugging evaluation scores or iterating on skill content.

### tessl__eval-setup

Generate eval scenarios from repo commits, configure multi-agent runs, execute baseline and with-context evals, and compare results.

### tessl__google-scholar-search

Search Google Scholar for academic papers and author profiles. Returns titles, authors, abstracts, and links.

### tessl__markdown-authoring

Author high-quality Markdown documentation with deterministic structure, lint compliance, and CI integration.

### tessl__skill-quality-auditor

Evaluate, score, and remediate agent skill collections using the 9-dimension quality framework. Performs duplication detection, generates remediation plans, enforces CI quality gates, and tracks score trends.

### tessl__software-design-principles

Apply SOLID principles, detect design anti-patterns, and evaluate architectural trade-offs for code reviews, design decisions, and refactoring.

## How skills are installed

The `init` command installs the canonical skill from `cmd/assets/SKILL.md` into agent harness directories. The agent registry in `agents/registry.go` maps each supported agent to its skill install path:

```text
.context/plugins/  → local skills (tessl-managed, file: source)
.claude/skills/    → claude-code
```

See [adding an agent](adding-an-agent.md) for the full agent list and how to add a new one.
