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
| 6 | Follow the finding-to-plan workflow | ALWAYS follow the 7-step read → plan → index → branch → commit sequence when drafting a plan from a `.context/findings/` entry |
| 7 | Always conduct session-end reflection | ALWAYS run a confidence audit and blind-spot check before concluding a session |
| 8 | Never use `/tmp` or `/temp` — use `.tmp` | NEVER write temporary files to `/tmp/` or `/temp/`; ALWAYS use the repo-local `.tmp/` |
| 9 | Never leak sensitive information | NEVER expose proprietary code, internal URLs, or other non-public data |
| 10 | Always consider subagent delegation when authoring skills | PREFER an independent subagent for self-reflective, meta-cognitive, or adversarial review steps |
| 11 | Never embed templates in markdown — use YAML template files | ALWAYS pair a template with a JSON Schema and a validation script under the skill's `assets/`/`scripts/` |
| 12 | Always check skill overlap before creating new skills | ALWAYS scan existing skills' triggers/descriptions for overlap before proposing a new one |
| 13 | Avoid Python/Node.js scripts in skills | PREFER bash/awk/sed over Python/Node.js for skill `scripts/` logic |
| 14 | No man left behind — triage every warning surfaced during work | NEVER leave a warning, issue, or error unaddressed just because it's unrelated to the current task; ALWAYS fix it or explicitly defer it with a documented reason |
| 15 | Regenerate, don't hand-merge, conflicts on auto-generated files | NEVER hand-resolve merge/rebase conflicts on generator-produced files (e.g. `.context/index.yaml`); ALWAYS take either side then re-run the generator script |
| 16 | Check for unrelated uncommitted work before `git reset --hard` | ALWAYS run `git status` immediately before discarding uncommitted changes; commit or stash anything unrelated first |

To add a new rule, load the [`rules-management`](#rules-management) skill — it will load existing rules, check for duplicates, and append with the correct format.

## Local skills

Local helper skills live under `.context/plugins/pantheon-org/<domain>/<skill>/` (managed via
`tessl.json`, `source: file:.context/plugins/...`), grouped into four domains:

### context-mgmt

- **context-file** — create `.context/` files (plans, findings, analysis, known-issues) with
  standard YAML frontmatter, appropriate sections, and correct placement in `plans/`,
  `findings/`, `analysis/`, or `known-issues/`.
- **context-index** — regenerate [`.context/index.yaml`](https://github.com/pantheon-org/skill-quality-auditor/blob/main/.context/index.yaml)
  from all `.context/**/*.md` frontmatter and validate that all files carry the required
  frontmatter block.

### governance

- **adr-capture** — capture Architecture Decision Records (ADRs) from `.context/` plans,
  findings, and analyses. Extracts binding decisions into `docs/ADR/` and maintains the
  machine-readable index. Also runs post-merge status sync: after a PR merges,
  `merge-status-sync.sh` detects plans/ADRs left `active`/`proposed` when the merge should
  have closed them out, auto-flipping single-phase plans and flagging the rest for a human.
- **rules-management** — manage the agent behavioural rules in `.agents/RULES.md` — load
  existing rules, check for duplicates, and append new rules in the standard format with
  directive and rationale.

### planning

- **plan-create** — create `.context/plans/*.md` files with standard frontmatter and
  phases/tasks/waves decomposition, inferring section conventions from existing plans.
- **design-debate** — stress-test an unwritten idea by spawning independent subagents in
  genuinely opposing roles (advocate, skeptic, migration/risk), grounded in real repo
  investigation, concluding in a synthesized verdict — used *before* a plan exists. The
  verdict decides how it's recorded: a finding if proceeding, a known-issue if not.
- **plan-review** — review a plan using three independent subagent reviewers (Technical,
  Strategic, Risk), consolidate their feedback, then resolve it: editorial fixes land
  directly, genuine tradeoffs go through a one-question-at-a-time interview and are
  recorded in the plan's `## Decisions` section.

### workshop

- **docs-check** — validate the GitHub Pages documentation site built by docmd — orphan
  detection, ADR index freshness, build verification, and LLM output audit.
- **guided-interview** — conduct a structured, one-question-at-a-time interview with
  concrete mutually-exclusive options plus a free-text path, adapting later questions to
  prior answers, ending in a user-confirmed recap.
- **pr-author** — create and maintain GitHub PRs with live descriptions — template
  discovery, intelligent filling, and lifecycle updates.
- **session-reflection** — conduct a two-question session-end reflection to catch blind
  spots and under-investigated areas before concluding. Verified-but-deferred gaps become
  `.context/known-issues/` entries rather than evaporating in chat.
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
