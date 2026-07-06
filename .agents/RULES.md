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

---

### Rule: Always consider subagent delegation when authoring skills

**Directive:** When creating or revising a SKILL.md, ALWAYS assess whether any phase of the workflow benefits from subagent delegation. If the skill requires self-reflection, meta-cognitive auditing, or adversarial review, PREFER spawning an independent subagent rather than having the main agent perform the introspection itself. Document the subagent recommendation as a note in the relevant workflow section or as an "Advanced" subsection.

**Rationale:** LLMs are poor judges of their own outputs — self-reflection by the main agent suffers from the same blind spots it is trying to catch. An independent subagent brings a fresh perspective, catches normalised errors, and can be routed to a cheaper model. Skills like `session-reflection` already demonstrate this pattern with measurable quality improvements.

---

### Rule: Never embed templates in markdown — use YAML template files

**Directive:** NEVER embed artifact templates (YAML frontmatter, file structures, code scaffolds) directly in skill markdown files. ALWAYS create a separate YAML file under the skill's `assets/templates/` directory that describes the structure of the artifact. Each template MUST have a corresponding JSON Schema under `assets/schemas/` and a validation script under `scripts/` that validates artifacts against the schema.

**Rationale:** Templates embedded in markdown are invisible to machine validation, rot silently when their schema changes, and cannot be consumed programmatically. A YAML template + JSON Schema + validation script "trio" makes the contract explicit, testable, and discoverable by other tools and agents.

---

### Rule: Always check skill overlap before creating new skills

**Directive:** BEFORE creating or proposing a new skill, ALWAYS scan existing skills in the repository to identify overlap in purpose, triggers, or workflow. Use `grep -r 'triggers\|description' .context/plugins/**/SKILL.md` and review the Tessl manifest (`tessl.json`) to map current capabilities. If overlap is found, propose a path forward: merge, specialise, deprecate, or proceed as-is with documented rationale.

**Rationale:** Skills with overlapping triggers fragment agent attention and confuse discovery. A 30-second overlap check prevents duplicate work and surfaces consolidation opportunities before they become maintenance debt.

---

### Rule: Avoid Python/Node.js scripts in skills

**Directive:** NEVER implement a skill's `scripts/` logic in Python or Node.js when a pure POSIX shell (bash/awk/sed) implementation is feasible. PREFER bash/awk/sed for skill scripts, even where it requires more code than an equivalent Python one-liner.

**Rationale:** Skill scripts under `.context/plugins/` run on whatever machine invokes the skill; bash/awk/sed are available on virtually every Unix-like system by default, while Python 3 and Node.js are not guaranteed to be installed or on `PATH`. A skill that silently depends on an uninstalled interpreter fails opaquely for portability reasons that a shell-only implementation avoids entirely.

---

### Rule: No man left behind — triage every warning surfaced during work

**Directive:** NEVER leave a warning, issue, or error unaddressed just because it is unrelated to the current task, including ones surfaced incidentally by a script, linter, build, or tool run. ALWAYS triage it in the same session: either fix it, or explicitly defer it with a documented reason (e.g. a `.context/findings/` entry or a note to the user) — never let it pass by silently.

**Rationale:** Incidental warnings are often the cheapest signal of latent problems available — ignoring them because they're "out of scope" lets known issues accumulate invisibly until they resurface as harder-to-diagnose failures. Triage-or-document costs little and keeps the repository's issue surface honest.

---

### Rule: Regenerate, don't hand-merge, conflicts on auto-generated files

**Directive:** NEVER hand-resolve merge/rebase conflict markers on a file that's produced by a generator script (e.g. `.context/index.yaml`, `docs/ADR/index.yaml`). ALWAYS take either side (`git checkout --ours` or `--theirs`) to clear the conflict, then re-run that file's generator script (e.g. `regenerate-context-index.sh`, `regenerate-adr-index.sh`) so the committed content is freshly derived from the current source-of-truth files, then stage and continue.

**Rationale:** These files are deterministically derived by scanning current frontmatter/content across the repo, not authored by hand. A manually merged set of conflict markers can satisfy git's conflict resolution while still being internally inconsistent with what the generator would actually produce (duplicate entries, stale counts, wrong ordering) — regenerating is strictly correct and no slower than merging by hand.

---

### Rule: Check for unrelated uncommitted work before `git reset --hard`

**Directive:** ALWAYS run `git status` immediately before any `git reset --hard` (or other command that discards uncommitted working-tree changes, e.g. `git checkout -- .`, `git clean -f`). If uncommitted changes exist beyond what you intend to discard, commit or stash them first — `git reset --hard` wipes ALL uncommitted changes to tracked files, not just the target commit's diff.

**Rationale:** Learned live: a regression-test throwaway commit was discarded via `git reset --hard HEAD~1` while real, working, tested implementation changes were still sitting as uncommitted working-tree edits — the hard reset silently destroyed that real work along with the intended throwaway commit. Caught only because the next verification step's output looked wrong, and the diff had to be manually reconstructed from memory rather than recovered. A `git status` check immediately before the reset costs nothing and would have caught this before it happened.

