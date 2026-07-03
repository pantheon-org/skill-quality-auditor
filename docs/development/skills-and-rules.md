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

Local helper skills live under `.context/plugins/pantheon-org/<domain>/<skill>/` (managed via
`tessl.json`, `source: file:.context/plugins/...`), grouped into four domains:

### context-mgmt

- **context-file** — create `.context/` files (plans, findings, analysis) with standard YAML
  frontmatter, appropriate sections, and correct placement in `plans/`, `findings/`, or `analysis/`.
- **context-index** — regenerate [`.context/index.yaml`](https://github.com/pantheon-org/skill-quality-auditor/blob/main/.context/index.yaml)
  from all `.context/**/*.md` frontmatter and validate that all files carry the required
  frontmatter block.

### governance

- **adr-capture** — capture Architecture Decision Records (ADRs) from `.context/` plans,
  findings, and analyses. Extracts binding decisions into `docs/ADR/` and maintains the
  machine-readable index.
- **rules-management** — manage the agent behavioural rules in `.agents/RULES.md` — load
  existing rules, check for duplicates, and append new rules in the standard format with
  directive and rationale.

### planning

- **plan-create** — create `.context/plans/*.md` files with standard frontmatter and
  phases/tasks/waves decomposition, inferring section conventions from existing plans.
- **plan-review** — review a plan using three independent subagent reviewers (Technical,
  Strategic, Risk) and consolidate their feedback.

### workshop

- **docs-check** — validate the GitHub Pages documentation site built by docmd — orphan
  detection, ADR index freshness, build verification, and LLM output audit.
- **pr-author** — create and maintain GitHub PRs with live descriptions — template
  discovery, intelligent filling, and lifecycle updates.
- **session-reflection** — conduct a two-question session-end reflection to catch blind
  spots and under-investigated areas before concluding.
- **socratic-method** — refine vague, complex, or high-stakes prompts through Socratic
  dialogue before committing to an implementation.

## Registry plugins

Skills installed from the Tessl registry (not authored in this repo) live under
`.tessl/plugins/<org>/<skill>/`:

- **pantheon-ai/markdown-authoring** — author high-quality Markdown with deterministic
  structure, lint compliance, and CI integration.
- **pantheon-ai/skill-quality-auditor** — the registry-distributed copy of this repo's own
  tile (evaluate, score, and remediate skill collections against the 9-dimension framework).
- **pantheon-ai/software-design-principles** — apply SOLID principles, detect design
  anti-patterns, and evaluate architectural trade-offs.
- **pantheon-ai/research/google-scholar-search** — search Google Scholar for academic
  papers and author profiles.
- **tessl-labs/eval-setup** — generate eval scenarios from repo commits, configure
  multi-agent runs, execute baseline and with-context evals, and compare results.
- **tessl-labs/eval-improve** — analyse eval results, diagnose failures, apply targeted
  fixes, and re-run to verify improvements.

## How skills are installed

The `init` command installs the canonical skill from `cmd/assets/SKILL.md` into agent harness directories. The agent registry in `agents/registry.go` maps each supported agent to its skill install path:

```text
.context/plugins/  → local skills (tessl-managed, file: source)
.claude/skills/    → claude-code
```

See [adding an agent](adding-an-agent.md) for the full agent list and how to add a new one.
