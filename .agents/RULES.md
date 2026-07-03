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

---

### Rule: Allow skills to evolve with experience

**Directive:** NEVER treat a skill as final after its initial creation. ALWAYS revisit and improve skills when experience reveals better patterns, edge cases, or anti-patterns. Update the SKILL.md, references, evals, and scripts as knowledge grows.

**Rationale:** Skills are living documentation. Freezing them at creation time means they ossify and lose relevance. Iteration based on real usage is what keeps them high-signal.

---

### Rule: Formalise ad hoc scripts after repeated use

**Directive:** ALWAYS use `.tmp/` for one-off or experimental scripts. When a script has been used more than twice from `.tmp/`, ALWAYS formalise it — move it to an appropriate permanent location (`scripts/`, `tools/`, or the relevant package), add documentation, and update references so agents can discover it. NEVER let `.tmp/` scripts become de facto permanent tools.

**Rationale:** Scripts that are run repeatedly from `.tmp/` become invisible to other agents and accumulate silently. Formalising after the second use ensures discoverability, maintainability, and prevents `.tmp/` from becoming a dumping ground.

---

### Rule: Never write temporary scripts outside .tmp

**Directive:** NEVER write ad hoc, experimental, or one-off scripts anywhere other than `.tmp/`. When you need a script for a quick task, create it under `.tmp/`. If the script proves useful and is used more than twice, formalise it to a permanent location.

**Rationale:** Scattering temporary scripts across the repository creates cleanup debt, confuses other agents, and pollutes permanent directories with throwaway code. A single `.tmp/` convention makes it obvious what is ephemeral and what is permanent.

---

### Rule: Follow the finding-to-plan workflow

**Directive:** ALWAYS use this sequence when creating a draft plan from a `.context/findings/` file: (1) read the findings file thoroughly, (2) check related plans and source code referenced in the findings, (3) create the plan in `.context/plans/` following the `context-file` skill template with `status: draft`, (4) regenerate `.context/index.yaml`, (5) branch off `main` using `feat/` prefix, (6) stage plan + related findings + index.yaml, (7) commit with a conventional message (`feat(scope): ...`). If the finding contains a `## Recommendation` or other decision indicator, assess whether an ADR is needed.

**Rationale:** This sequence ensures the plan is grounded in existing research, properly indexed, and committed on a clean branch — avoiding stale context, broken index entries, and undocumented decisions.

---

### Rule: Always conduct session-end reflection

**Directive:** BEFORE concluding any session, ALWAYS initiate a two-question reflection. Say something like "Before we wrap up, let me reflect on what I'm unsure about and what I might be missing." Then:

1. **Confidence audit** — "What am I least confident about right now?" List 3–7 specific items that were under-investigated, assumed, or skipped. For each, state why confidence is low (e.g., shallow search, unverified assumption, skipped edge case).
2. **Blind-spot check** (Sam Altman) — "What's the biggest thing I'm missing about this situation? What don't I realize?" Identify potential assumptions, unexamined alternatives, or overlooked evidence from the user's perspective.

After both questions, offer to investigate any item the user flags. If they accept, do deep root-cause investigation — search for contradicting evidence, trace assumptions, update conclusions — before concluding.

**Rationale:** ~1 in 4 sessions surfaces a critical gap that would silently invalidate delivered work. Catching it before sign-off is the lowest-cost intervention point. The `session-reflection` skill has detailed guidance.

---

### Rule: Never use /tmp or /temp — use .tmp

**Directive:** NEVER write temporary files to `/tmp/` or `/temp/`. ALWAYS use `.tmp/` (the repo-local `.tmp` directory) for all temporary, experimental, or one-off artifacts.

**Rationale:** System `/tmp/` is not shared across agent sessions and is cleared on reboot, making temp files invisible to other agents and unrecoverable after a restart. The repo-local `.tmp/` is gitignored, discoverable by agents working in this repo, and persists across sessions for the lifetime of the worktree.

---

### Rule: Never leak sensitive information

**Directive:** NEVER include, expose, or reference sensitive information about any company, individual, or internal process — including proprietary code, internal URLs, unreleased features, employee details, organisational structure, internal tooling, process documentation, or any non-public business logic. When in doubt, treat information as sensitive.

**Rationale:** Leaked sensitive information creates legal, security, and reputational risk for the company and individuals involved. The agent must default to nondisclosure and sanitise any output that could expose non-public data.

