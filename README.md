# skill-quality-auditor

A 9-dimension scoring framework for auditing and improving AI skill quality. Combines structural validation with custom
scoring across Knowledge Delta, Mindset, Anti-Patterns, Specification Compliance, Progressive Disclosure, Freedom
Calibration, Pattern Recognition, Practical Usability, and Eval Validation.

## Repository layout

```text
skill-auditor/              # Go CLI binary
  cmd/                      # cobra commands: evaluate, batch, duplication, aggregate, remediate, trend, validate, lint, prune, analyze, init
  cmd/assets/               # Tessl tile — SKILL.md, tile.json, evals, references, schemas, templates (single source of truth)
  agents/                   # agent registry (supported agent environments for `init`)
  scorer/                   # D1–D9 dimension scorers
  analysis/                 # TF-IDF keyword extractor + rule-based pattern detectors (used by analyze)
  duplication/              # word-level Jaccard similarity engine (used by duplication + aggregate)
  reporter/                 # text/JSON formatters, audit store, duplication/aggregation/remediation/analysis reports
  testdata/                 # fixture skills for unit tests
```

## Quick start

```bash
cd skill-auditor
go build -o bin/skill-auditor .

# Evaluate a single skill
./bin/skill-auditor evaluate skills/my-skill

# Evaluate with JSON output and persist result
./bin/skill-auditor evaluate skills/my-skill --json --store

# Evaluate multiple skills; fail CI if any score below B
./bin/skill-auditor batch skills/skill-a skills/skill-b --fail-below B

# Detect duplicate or overlapping skills
./bin/skill-auditor duplication

# Generate an aggregation plan for a skill family
./bin/skill-auditor aggregate --family bdd

# Generate a remediation plan from a stored audit
./bin/skill-auditor remediate domain/my-skill

# Validate an existing remediation plan
./bin/skill-auditor remediate domain/my-skill --validate

# Show score trends across stored audits
./bin/skill-auditor trend

# Validate skill artifact conventions
./bin/skill-auditor validate artifacts

# Check skill consistency (frontmatter, shebangs)
./bin/skill-auditor lint

# Prune old stored audits, keep last 5 per skill
./bin/skill-auditor prune

# Full semantic + pattern analysis pipeline
./bin/skill-auditor analyze domain/my-skill

# Install the skill into local agent environments
./bin/skill-auditor init
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

### `validate`

```text
skill-auditor validate artifacts [paths...] [flags]
skill-auditor validate review <file> [flags]

Flags (artifacts):
  --repo-root  repo root directory (auto-detected if omitted)

Flags (review):
  --strict-recommended  treat recommended fields as errors
  --repo-root           repo root directory (auto-detected if omitted)
```

`validate artifacts` checks `SKILL.md` line limits, frontmatter name match, asset subdirectory
conventions, script shebangs, and schema file validity. `validate review` checks a review report
file against the embedded requirements spec for required/recommended sections, headings, and labels.
Exit code 1 on any error.

### `lint`

```text
skill-auditor lint [skills-dir] [flags]

Flags:
  --repo-root  repo root directory (auto-detected if omitted)
```

Checks each skill directory for a `SKILL.md`, a frontmatter block, and correct script shebangs.
Prints `MISSING_SKILL`, `NO_FRONTMATTER`, `BAD_SHEBANG` tags per issue. Exits with the issue count
(0 = clean).

### `prune`

```text
skill-auditor prune [flags]

Flags:
  --keep       number of audit date-dirs to retain per skill (default 5)
  --repo-root  repo root directory (auto-detected if omitted)
```

Removes old date-stamped audit directories from `.context/audits/`, keeping the N most recent per
skill. Preserves `latest` symlinks.

### `init`

```text
skill-auditor init [flags]

Flags:
  --agent   agent(s) to install into (default: auto-detect from installed environments)
  --global  install to global skill directory (~/<agent>/skills/)
  --method  installation method: symlink or copy (default: symlink)
```

Installs the embedded `skill-quality-auditor` SKILL.md (and its `references/` directory) into one or
more agent skill directories. Auto-detects supported environments (Claude Code, Cursor, etc.) when
`--agent` is omitted.

### `analyze`

```text
skill-auditor analyze <skill> [flags]

Flags:
  --semantic   run TF-IDF keyword extraction only
  --patterns   run rule-based pattern detection only
  --pipeline   run full pipeline — semantic + patterns + combined report (default)
  --json       emit JSON output
  --store      write report to .context/analysis/
  --limit int  max keywords to include (default 20)
  --repo-root  repo root directory (auto-detected if omitted)
```

Performs semantic and structural analysis of a skill without requiring external NLP or ML tooling.
`--semantic` extracts TF-IDF top keywords scored against the full skill corpus. `--patterns` runs
rule-based detectors for required sections, trigger-word frequency, structural conformance, and
anti-pattern signals. The default `--pipeline` mode runs both and writes a combined report to
`.context/analysis/pattern-report-<skill>-YYYY-MM-DD.md`.

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
| --- | --- |
| `evaluate --store` / `batch --store` | `.context/audits/<skill>/<date>/` |
| `duplication` | `.context/analysis/duplication-report-YYYY-MM-DD.md` |
| `aggregate` | `.context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md` |
| `remediate` | `.context/plans/<skill>-remediation-plan.md` |
| `analyze --store` | `.context/analysis/pattern-report-<skill>-YYYY-MM-DD.md` |

## Tessl skill

The Tessl tile `pantheon-ai/skill-quality-auditor` (v0.1.5) is published from this repo. All tile assets — `SKILL.md`,
`tile.json`, `evals/`, `references/`, `schemas/`, and `templates/` — live under `skill-auditor/cmd/assets/`. Agents that
install this tile get structured guidance for running audits, generating remediation plans, detecting duplication, and
enforcing CI quality gates.

## Development

```bash
cd skill-auditor
go test ./...
go vet ./...
```
