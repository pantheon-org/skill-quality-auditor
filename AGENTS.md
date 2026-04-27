# Agent guidance

This file is the authoritative entry point for AI agents working in this repository.
Read it before exploring any other files.

## What this repo does

`skill-quality-auditor` is a Go CLI (`skill-auditor`) and a Tessl tile that scores AI skills against a 9-dimension quality
framework. It produces letter grades, per-dimension diagnostics, and remediation guidance.

## Repo map

| Path | What it is |
| ---- | ---------- |
| `skill-auditor/` | Go CLI — build and run this to audit skills |
| `skill-auditor/scorer/` | D1–D9 scorers; each file is one dimension |
| `skill-auditor/reporter/` | Formats results as text or JSON; persists to `.context/audits/` |
| `skill-auditor/cmd/` | `evaluate` (single skill) and `batch` (multiple skills) cobra commands |
| `skill-auditor/testdata/` | Fixture skills for unit tests — do not modify without updating tests |
| `skill/skill-quality-auditor/` | Tessl tile; contains `SKILL.md`, `AGENTS.md`, evals, references, scripts |
| `skill/skill-quality-auditor/references/` | Authoritative rubrics, scoring criteria, anti-pattern catalogues |

## How to evaluate a skill

```bash
cd skill-auditor
go build -o skill-auditor .
./skill-auditor evaluate <path-or-key> [--json] [--store]
./skill-auditor batch <skill1> <skill2> [--fail-below B]
```

`<path-or-key>` is either a `domain/skill-name` key (resolved under `<repo-root>/skills/`), a directory containing
`SKILL.md`, or a direct path to `SKILL.md`.

## Scoring dimensions (D1–D9)

| ID | Dimension | Max |
| -- | --------- | --- |
| D1 | Knowledge Delta | 20 |
| D2 | Mindset & Procedures | 20 |
| D3 | Anti-Pattern Coverage | 20 |
| D4 | Specification Compliance | 10 |
| D5 | Progressive Disclosure | 10 |
| D6 | Freedom Calibration | 10 |
| D7 | Pattern Recognition | 5 |
| D8 | Practical Usability | 5 |
| D9 | Eval Validation | 10 |

Total: 110 pts. Grade bands and CI thresholds: `skill/skill-quality-auditor/references/quality-thresholds-scoring.md`.

## Key rules

- **Never commit to `main` directly.** Branch → PR → merge.
- **Run `go test ./...` before reporting any Go change as done.**
- **Tessl eval changes require `tessl eval run skill/skill-quality-auditor/` to pass.**
- For deep rubric questions, load `skill/skill-quality-auditor/references/framework-dimensions.md` first.
- For anti-pattern analysis, load `skill/skill-quality-auditor/references/detailed-anti-patterns.md`.
- Audit outputs land in `.context/audits/` — never commit that directory.

## Suggested task workflows

### Improve a dimension scorer

1. Read `skill-auditor/scorer/dN_<name>.go` and its test file.
2. Consult the matching rubric section in `references/framework-dimensions.md`.
3. Edit the scorer, add/update tests, run `go test ./scorer/...`.

### Add a new skill to evaluate

1. Place `SKILL.md` under `skills/<domain>/<name>/`.
2. Run `./skill-auditor evaluate <domain>/<name> --store`.
3. Review diagnostics and iterate.

### Update the Tessl tile

1. Edit files under `skill/skill-quality-auditor/`.
2. Run `tessl eval run skill/skill-quality-auditor/`.
3. Bump `version` in `skill/skill-quality-auditor/tile.json`.
