# Agent guidance

This file is the authoritative entry point for AI agents working in this repository.
Read it before exploring any other files.

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
| `scorer/` | D1–D9 scorers; each file is one dimension |
| `duplication/` | Word-level Jaccard similarity engine (inventory, pairwise detect) |
| `reporter/` | Formats results as text or JSON; persists to `.context/audits/`; duplication, aggregation, and remediation plan formatters |
| `cmd/` | `evaluate`, `batch`, `duplication`, `aggregate`, `remediate`, `trend`, `validate`, `analyze`, `prune`, `init`, `update` cobra commands |
| `cmd/assets/` | Embedded SKILL.md, tile.json, references, evals, schemas, templates, requirements — single source of truth |
| `internal/` | Shared utilities (tokenizer) |
| `testdata/` | Fixture skills for unit tests — do not modify without updating tests |

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
- **Tessl eval changes require `tessl eval run cmd/assets/` to pass.**
- For deep rubric questions, load `cmd/assets/references/framework-dimensions.md` first.
- For anti-pattern analysis, load `cmd/assets/references/detailed-anti-patterns.md`.
- **Edit assets directly under `cmd/assets/`** — there is no separate `skill/` directory to mirror.
- Audit outputs: `.context/audits/`, `.context/analysis/`, `.context/plans/` — never commit those directories.

## Output locations

| Command | Output path |
| --- | --- |
| `evaluate --store` / `batch --store` | `.context/audits/<skill>/<date>/` |
| `duplication` | `.context/analysis/duplication-report-YYYY-MM-DD.md` |
| `aggregate` | `.context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md` |
| `remediate` | `.context/plans/<skill>-remediation-plan.md` |

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
2. Run `tessl eval run cmd/assets/`.
3. Do **not** bump `version` in `cmd/assets/tile.json` manually — release-please auto-bumps it
   alongside the binary version on every release PR.
