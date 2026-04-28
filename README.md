# skill-quality-auditor

A 9-dimension scoring framework for auditing and improving AI skill quality. Combines structural validation with custom
scoring across Knowledge Delta, Mindset, Anti-Patterns, Specification Compliance, Progressive Disclosure, Freedom
Calibration, Pattern Recognition, Practical Usability, and Eval Validation.

## Repository layout

```text
skill/                  # Tessl tile (pantheon-ai/skill-quality-auditor)
skill-auditor/          # Go CLI binary
  cmd/                  # cobra commands: evaluate, batch, duplication, aggregate, remediate, trend
  scorer/               # D1–D9 dimension scorers
  duplication/          # word-level Jaccard similarity engine (used by duplication + aggregate)
  reporter/             # text/JSON formatters, audit store, duplication/aggregation/remediation reports
  testdata/             # fixture skills for unit tests
```

## Quick start

```bash
cd skill-auditor
go build -o skill-auditor .

# Evaluate a single skill
./skill-auditor evaluate skills/my-skill

# Evaluate with JSON output and persist result
./skill-auditor evaluate skills/my-skill --json --store

# Evaluate multiple skills; fail CI if any score below B
./skill-auditor batch skills/skill-a skills/skill-b --fail-below B

# Detect duplicate or overlapping skills
./skill-auditor duplication

# Generate an aggregation plan for a skill family
./skill-auditor aggregate --family bdd

# Generate a remediation plan from a stored audit
./skill-auditor remediate domain/my-skill

# Validate an existing remediation plan
./skill-auditor remediate domain/my-skill --validate

# Show score trends across stored audits
./skill-auditor trend
```

## CLI reference

### `evaluate`

```text
skill-auditor evaluate <skill> [flags]

Flags:
  --json        emit JSON output instead of human-readable text
  --store       persist result to .context/audits/
  --repo-root   repo root directory (auto-detected from .git / go.mod if omitted)
```

`<skill>` accepts a `domain/skill-name` key (resolved under `<repo-root>/skills/`), an absolute path to a directory
containing `SKILL.md`, or a direct path to `SKILL.md`.

### `batch`

```text
skill-auditor batch <skill1> [skill2 ...] [flags]

Flags:
  --json        emit JSON array output
  --store       persist each result to .context/audits/
  --fail-below  exit 1 if any skill scores below this grade (e.g. B+)
  --repo-root   repo root directory (auto-detected if omitted)
```

### `duplication`

```text
skill-auditor duplication [skills-dir] [flags]

Flags:
  --json        emit JSON array of pairs
  --skills-dir  skills directory (default: <repo-root>/skills)
  --repo-root   repo root directory (auto-detected if omitted)
```

Performs pairwise word-level Jaccard similarity across all SKILL.md files. Writes
`duplication-report-YYYY-MM-DD.md` to `.context/analysis/`. Exits with code 2 if any
Critical (>35%) pairs are found — suitable for use as a CI gate.

### `aggregate`

```text
skill-auditor aggregate --family <prefix> [skills-dir] [flags]

Flags:
  --family      skill family prefix to analyse (required, e.g. bdd, typescript)
  --dry-run     print plan to stdout without writing to disk
  --skills-dir  skills directory (default: <repo-root>/skills)
  --repo-root   repo root directory (auto-detected if omitted)
```

Identifies consolidation candidates within a family and produces a 6-step aggregation
plan at `.context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md`.

### `remediate`

```text
skill-auditor remediate <skill> [flags]

Flags:
  --target-score  desired total score (default: current + 20, max 140)
  --validate      validate an existing plan file instead of generating one
  --repo-root     repo root directory (auto-detected if omitted)
```

Reads the most recent stored audit for `<skill>` and generates a schema-compliant
YAML-frontmatter remediation plan at `.context/plans/<skill>-remediation-plan.md`.

With `--validate`, checks the existing plan against `remediation-plan.schema.json` and
reports any violations.

### `trend`

```text
skill-auditor trend [flags]

Flags:
  --json       emit JSON array output
  --repo-root  repo root directory (auto-detected if omitted)
```

Reads the two most recent stored audits per skill from `.context/audits/` and prints a
score-delta table with ↑ / ↓ / — indicators.

## Scoring dimensions

| ID | Dimension | Max pts |
| -- | --------- | ------- |
| D1 | Knowledge Delta | 20 |
| D2 | Mindset & Procedures | 15 |
| D3 | Anti-Pattern Coverage | 15 |
| D4 | Specification Compliance | 15 |
| D5 | Progressive Disclosure | 15 |
| D6 | Freedom Calibration | 15 |
| D7 | Pattern Recognition | 10 |
| D8 | Practical Usability | 15 |
| D9 | Eval Validation | 20 |

**Total: 140 pts.** Grades: **A+** (≥133) → **F** (<91). See
`skill/skill-quality-auditor/references/quality-thresholds-scoring.md` for the full rubric.

## Output locations

| Command | Output path |
|---------|-------------|
| `evaluate --store` / `batch --store` | `.context/audits/<skill>/<date>/` |
| `duplication` | `.context/analysis/duplication-report-YYYY-MM-DD.md` |
| `aggregate` | `.context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md` |
| `remediate` | `.context/plans/<skill>-remediation-plan.md` |

## Tessl skill

The `skill/` directory contains the published Tessl tile `pantheon-ai/skill-quality-auditor` (v0.1.5). Agents that install
this tile get structured guidance for running audits, generating remediation plans, detecting duplication, and enforcing
CI quality gates.

## Development

```bash
cd skill-auditor
go test ./...
go vet ./...
```
