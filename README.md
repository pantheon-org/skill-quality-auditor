# skill-quality-auditor

A 9-dimension scoring framework for auditing and improving AI skill quality. Combines structural validation with custom scoring across Knowledge Delta, Mindset, Anti-Patterns, Specification Compliance, Progressive Disclosure, Freedom Calibration, Pattern Recognition, Practical Usability, and Eval Validation.

## Repository layout

```
skill/                  # Tessl tile (pantheon-ai/skill-quality-auditor)
skill-auditor/          # Go CLI binary
  cmd/                  # cobra commands: evaluate, batch, version
  scorer/               # D1–D9 dimension scorers
  reporter/             # text/JSON formatter and audit store
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
```

## CLI reference

### `evaluate`

```
skill-auditor evaluate <skill> [flags]

Flags:
  --json        emit JSON output instead of human-readable text
  --store       persist result to .context/audits/
  --repo-root   repo root directory (auto-detected from .git / go.mod if omitted)
```

`<skill>` accepts a `domain/skill-name` key (resolved under `<repo-root>/skills/`), an absolute path to a directory containing `SKILL.md`, or a direct path to `SKILL.md`.

### `batch`

```
skill-auditor batch <skill1> [skill2 ...] [flags]

Flags:
  --json        emit JSON array output
  --store       persist each result to .context/audits/
  --fail-below  exit 1 if any skill scores below this grade (e.g. B+)
  --repo-root   repo root directory (auto-detected if omitted)
```

## Scoring dimensions

| ID | Dimension | Max pts |
|----|-----------|---------|
| D1 | Knowledge Delta | 20 |
| D2 | Mindset & Procedures | 20 |
| D3 | Anti-Pattern Coverage | 20 |
| D4 | Specification Compliance | 10 |
| D5 | Progressive Disclosure | 10 |
| D6 | Freedom Calibration | 10 |
| D7 | Pattern Recognition | 5 |
| D8 | Practical Usability | 5 |
| D9 | Eval Validation | 10 |

Grades: **A** (≥90) → **F** (<50). See `skill/skill-quality-auditor/references/scoring-rubric.md` for the full rubric.

## Tessl skill

The `skill/` directory contains the published Tessl tile `pantheon-ai/skill-quality-auditor` (v0.1.5). Agents that install this tile get structured guidance for running audits, generating remediation plans, detecting duplication, and enforcing CI quality gates.

## Development

```bash
cd skill-auditor
go test ./...
go vet ./...
```

Results are stored under `.context/audits/<domain>/<skill-name>/<timestamp>.json` when `--store` is used.
