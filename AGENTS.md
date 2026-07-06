# Agent guidance

This file is the authoritative entry point for AI agents working in this repository.
Read it before exploring any other files. `CLAUDE.md` is a symlink to this file — there
is one source of truth, not two files that could drift apart.

## What this repo does

`skill-quality-auditor` is a Go CLI (`skill-auditor`) and a Tessl tile that scores AI skills against a 9-dimension quality
framework. It produces letter grades, per-dimension diagnostics, and remediation guidance.

The Tessl tile is distributed from this repo: `tile.json` lives at `cmd/assets/tile.json`,
co-located with `SKILL.md` and `evals/`. All skill assets (SKILL.md, tile.json, references, evals, schemas, templates)
live exclusively under `cmd/assets/` — there is no separate `skill/` directory.

## Repo map

| Path | What it is |
| ---- | ---------- |
| `go.mod` / `main.go` | Go CLI root — build and run from repo root |
| `agents/` | Agent registry — supported environments for the `init` command |
| `docs/` | Per-dimension documentation — scoring criteria, examples, and academic references |
| `docs/ADR/` | Architecture Decision Records — indexed at [`docs/ADR/index.yaml`](docs/ADR/index.yaml) |
| `scorer/` | D1–D9 scorers; each file is one dimension |
| `duplication/` | Word-level Jaccard similarity engine (inventory, pairwise detect) |
| `reporter/` | Formats results as text or JSON; persists to `.context/audits/`; duplication, aggregation, and remediation plan formatters |
| `cmd/` | `evaluate`, `batch`, `duplication`, `aggregate`, `remediate`, `trend`, `validate`, `analyze`, `prune`, `init`, `update` cobra commands; `init` installs into CWD by default, `~` with `--global` |
| `cmd/assets/` | Embedded SKILL.md, tile.json, references, evals, schemas, templates, requirements — single source of truth |
| `internal/` | Shared utilities (tokenizer) |
| `.context/plugins/` | Tessl helper skill plugins (socratic-method, etc.) — available to AI agents; not part of the Go CLI |
| `testdata/` | Fixture skills for unit tests — do not modify without updating tests |

## Ways of working

Before making any change, read `.context/instructions/ways-of-working.md` for the branch workflow, commit conventions, and plan-status sync rules. The short version:

1. **Create a branch from `main` first** — use `feat/`, `fix/`, `chore/` prefixes.
2. Commit atomically with conventional messages.
3. Rebase on `main` if it diverges.
4. Run `hk check && go test ./...` before pushing.
5. **Update plan frontmatter** (`ACTIVE → DONE`) when you implement what a plan describes.

## How to evaluate a skill

```bash
go build -o dist/skill-auditor .
./dist/skill-auditor evaluate <path-or-key> [--store]
./dist/skill-auditor batch <skill1> <skill2> [--fail-below B]
```

`<path-or-key>` is either a `domain/skill-name` key (resolved under `<repo-root>/skills/`), a directory containing
`SKILL.md`, or a direct path to `SKILL.md`.

## How to detect duplication and plan aggregation

```bash
# Detect duplicate/overlapping skills (exits 2 on Critical pairs)
./dist/skill-auditor duplication

# Generate aggregation plan for a skill family
./dist/skill-auditor aggregate --family <prefix>        # writes .context/analysis/
./dist/skill-auditor aggregate -f <prefix>              # writes .context/analysis/
./dist/skill-auditor aggregate --family <prefix> --dry-run  # stdout only
./dist/skill-auditor aggregate -f <prefix> -n           # stdout only (--dry-run)
```

## How to generate and validate remediation plans

```bash
# Requires a prior --store run for the skill
./dist/skill-auditor remediate <skill> [--target-score N]  # writes .context/plans/
./dist/skill-auditor remediate <skill> [-t N]              # writes .context/plans/
./dist/skill-auditor remediate <skill> -n                  # dry-run: stdout only
./dist/skill-auditor remediate <skill> --validate          # validate existing plan
./dist/skill-auditor remediate <skill> -v                  # validate existing plan
```

## How to track score trends

```bash
./dist/skill-auditor trend        # table with ↑/↓/— per skill
./dist/skill-auditor trend --json # machine-readable
```

## Scoring dimensions (D1–D9)

| ID | Dimension | Max |
| -- | --------- | --- |
| D1 | Knowledge Delta | 20 |
| D2 | Mindset & Procedures | 15 |
| D3 | Anti-Pattern Coverage | 15 |
| D4 | Specification Compliance | 15 |
| D5 | Progressive Disclosure | 15 |
| D6 | Freedom Calibration | 15 |
| D7 | Pattern Recognition | 10 |
| D8 | Practical Usability | 15 |
| D9 | Eval Validation | 20 |

Total: **140 pts.** Grade bands and CI thresholds: `cmd/assets/references/quality-thresholds-scoring.md`.

## Key rules

- **Never commit to `main` directly.** Branch → PR → merge.
- **Run `go test ./...` before reporting any Go change as done.**
- **Eval changes require `./dist/skill-auditor eval ./cmd/assets` to pass** (native eval runner; the structural gate runs every PR and every pre-push via `hk run pre-push`). The Tessl review step in CI is advisory during the proving period and will be removed once the native runner has 2 weeks of green CI runs.
- For deep rubric questions, load `cmd/assets/references/framework-dimensions.md` first.
- For anti-pattern analysis, load `cmd/assets/references/detailed-anti-patterns.md`.
- **Edit assets directly under `cmd/assets/`** — there is no separate `skill/` directory to mirror.
- All `.context/` files (plans, findings, analyses, index) are tracked in git — they are part of the project's institutional knowledge.

## Output locations

| Command | Output path |
| --- | --- |
| `evaluate --store` / `batch --store` | `.context/audits/<skill>/<date>/` |
| `duplication` | `.context/analysis/duplication-report-YYYY-MM-DD.md` |
| `aggregate` | `.context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md` |
| `remediate` | `.context/plans/<skill>-remediation-plan-<date>.md` |

## Context Index

Actionable plans, findings, and analyses live under `.context/`. The machine-readable index is at [`.context/index.yaml`](.context/index.yaml) — each entry carries `title`, `type`, `status`, `date`, optional `related` links, and (on `PLAN`/`FINDING`/`KNOWN_ISSUE`) `effort`, `severity`, `value`, and `themes` (an ordered subject-area list from a controlled vocabulary).

Read the index before starting a new task to surface active work items, pending decisions, and historical findings relevant to your change.

To pick the **highest-value item to do next**, use the read protocol: filter to `DRAFT`/`ACTIVE` `PLAN`/`FINDING`/`KNOWN_ISSUE`, sort by `value` (`HIGH` > `MEDIUM` > `LOW`) descending, then `effort` ascending, then `themes[0]` (prefer the area in focus) to break remaining ties, and act on the top item without re-judging. `value` is graded against [`.context/instructions/value-rubric.md`](.context/instructions/value-rubric.md) and `themes` against [`.context/instructions/theme-vocabulary.md`](.context/instructions/theme-vocabulary.md); both are required for those types while DRAFT/ACTIVE (`DONE`/`SUPERSEDED` exempt). See [`ways-of-working.md`](.context/instructions/ways-of-working.md) for grading and re-grade rules.

## Agent Rules

Behavioural rules for agents live in `.agents/RULES.md` — read it before any task. To add a new rule, load the `rules-management` skill.

## Architecture Decision Records (ADRs)

Architectural and process decisions extracted from `.context/` analyses, findings, and plans are recorded as ADRs under `docs/ADR/`. The machine-readable index is at [`docs/ADR/index.yaml`](docs/ADR/index.yaml) — each entry carries `adr`, `title`, `status`, `date`, `context` (source `.context/` files), and optional `superseded_by`.

Read the ADR index before making design decisions to avoid revisiting settled questions. When creating a `.context/` file that makes a binding decision, also create an ADR using the `adr-capture` skill.

## Suggested task workflows

### Improve a dimension scorer

1. Read `scorer/dN_<name>.go` and its test file.
2. Consult the matching rubric section in `cmd/assets/references/framework-dimensions.md`.
3. Edit the scorer, add/update tests, run `go test ./scorer/...`.

### Add a new skill to evaluate

1. Place `SKILL.md` under `skills/<domain>/<name>/`.
2. Run `./dist/skill-auditor evaluate <domain>/<name> --store`.
3. Review diagnostics and iterate.

### Detect and remediate duplication

1. Run `./dist/skill-auditor duplication` to find overlapping skills.
2. Run `./dist/skill-auditor aggregate --family <prefix>` to plan consolidation.
3. Follow the 6-step process in the generated plan.

### Generate a remediation plan

1. Run `./dist/skill-auditor evaluate <skill> --store` to capture an audit.
2. Run `./dist/skill-auditor remediate <skill>` to generate the plan.
3. Run `./dist/skill-auditor remediate <skill> --validate` to verify it is schema-compliant.
4. Implement the phased steps and re-evaluate.

### Update the Tessl tile

1. Edit files under `cmd/assets/`.
2. Run `./dist/skill-auditor eval ./cmd/assets` (native eval runner; structural gate, no key needed).
   For the LLM-judge advisory, run with `ANTHROPIC_API_KEY=... ./dist/skill-auditor eval ./cmd/assets --json --samples 3 --cost-log`.
3. Do **not** bump `version` in `cmd/assets/tile.json` manually — release-please auto-bumps it
   alongside the binary version on every release PR.

### Add a helper skill (non-Go Tessl plugin)

Helper skills live under `.context/plugins/` and provide agent workflows independent of the Go CLI.

1. Create the plugin files under `.context/plugins/<workspace>/<skill>/` (SKILL.md, tile.json, tessl-package.json).
2. Register in `tessl.json` with `"source": "file:.context/plugins/..."`.
3. Run `tessl install` to sync.
4. See `.context/instructions/adding-helper-skills.md` for full details.
