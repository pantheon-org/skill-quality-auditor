# skill-quality-auditor

[![CI](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml/badge.svg)](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A 9-dimension scoring framework for auditing and improving AI skill quality. Combines structural
validation with custom scoring across Knowledge Delta, Mindset, Anti-Patterns, Specification
Compliance, Progressive Disclosure, Freedom Calibration, Pattern Recognition, Practical Usability,
and Eval Validation.

- [Install](#install)
- [Command Usage](#command-usage)
- [Output Formats](#output-formats)
- [CI Integration](#ci-integration)
- [What it scores & why](#what-it-scores--why)
- [Repository layout](#repository-layout)
- [Development](#development)

---

## Install

### install.sh (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/pantheon-org/skill-quality-auditor/main/scripts/install.sh | sh
```

Override install directory or pin a version:

```bash
INSTALL_DIR=~/.local/bin curl -fsSL ... | sh
VERSION=v1.2.3 curl -fsSL ... | sh
```

### Homebrew

```bash
brew tap pantheon-org/tap
brew install skill-auditor
```

### mise

```bash
mise use ubi:pantheon-org/skill-quality-auditor
```

Or in `mise.toml`:

```toml
[tools]
"ubi:pantheon-org/skill-quality-auditor" = "latest"
```

### Go install

```bash
go install github.com/pantheon-org/skill-quality-auditor@latest
```

### Updating

| Method | Command |
| --- | --- |
| install.sh | `skill-auditor update` |
| Homebrew | `brew upgrade skill-auditor` |
| mise | `mise upgrade skill-auditor` |
| Go install | `go install github.com/pantheon-org/skill-quality-auditor@latest` |

> `skill-auditor update` also accepts `--check` (report without installing) and `--version-target vX.Y.Z`.

---

## Command Usage

| Stage | Command | What it answers |
| --- | --- | --- |
| Score a skill | [`evaluate`](#evaluate) | What is the overall quality grade and per-dimension breakdown? |
| Score many skills | [`batch`](#batch) | How do multiple skills compare, and does any fall below a CI threshold? |
| Find overlap | [`duplication`](#duplication) | Are any skills too similar to each other? |
| Plan consolidation | [`aggregate`](#aggregate) | How should a family of similar skills be merged? |
| Fix a skill | [`remediate`](#remediate) | What specific changes would raise this skill's score? |
| Track progress | [`trend`](#trend) | Are scores improving or regressing over time? |
| Validate format | [`validate`](#validate) | Do artifacts conform to conventions? Does a review report meet spec? |
| Check consistency | [`lint`](#lint) | Are frontmatter, shebangs, and structure correct across all skills? |
| Deep analysis | [`analyze`](#analyze) | What are the keyword signals and structural patterns in this skill? |
| Install skill | [`init`](#init) | How do I install this auditor skill into my agent environment? |
| Self-update | [`update`](#update) | Is a newer release available, and can I install it in place? |
| Housekeeping | [`prune`](#prune) | Which old audit snapshots can be removed? |

### `evaluate`

```text
skill-auditor evaluate <skill> [flags]

Flags:
  --json        emit JSON output instead of human-readable text
  --store       persist result to .context/audits/
  --repo-root   repo root directory (auto-detected from .git / go.mod if omitted)
```

`<skill>` accepts a `domain/skill-name` key (resolved under `<repo-root>/skills/`), a directory
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

Pairwise word-level Jaccard similarity across all `SKILL.md` files. Writes
`duplication-report-YYYY-MM-DD.md` to `.context/analysis/`. Exits with code 2 on any
Critical (>35%) pair — suitable as a CI gate.

### `aggregate`

```text
skill-auditor aggregate --family <prefix> [skills-dir] [flags]

Flags:
  --family      skill family prefix to analyse (required, e.g. bdd, typescript)
  --dry-run     print plan to stdout without writing to disk
  --skills-dir  skills directory (default: <repo-root>/skills)
  --repo-root   repo root directory (auto-detected if omitted)
```

Produces a 6-step consolidation plan at `.context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md`.

### `remediate`

```text
skill-auditor remediate <skill> [flags]

Flags:
  --target-score  desired total score (default: current + 20, max 140)
  --validate      validate an existing plan file instead of generating one
  --repo-root     repo root directory (auto-detected if omitted)
```

Reads the most recent stored audit for `<skill>` and generates a schema-compliant remediation
plan at `.context/plans/<skill>-remediation-plan.md`. Use `--validate` to check an existing plan
against `remediation-plan.schema.json`.

### `trend`

```text
skill-auditor trend [flags]

Flags:
  --json       emit JSON array output
  --repo-root  repo root directory (auto-detected if omitted)
```

Reads the two most recent stored audits per skill from `.context/audits/` and prints a score-delta
table with ↑ / ↓ / — indicators.

### `validate`

```text
skill-auditor validate artifacts [paths...] [flags]
skill-auditor validate review <file> [flags]

Flags (artifacts):
  --repo-root              repo root directory (auto-detected if omitted)

Flags (review):
  --strict-recommended     treat recommended fields as errors
  --repo-root              repo root directory (auto-detected if omitted)
```

`validate artifacts` checks `SKILL.md` line limits, frontmatter name match, asset subdirectory
conventions, script shebangs, and schema file validity. `validate review` checks a review report
against the embedded requirements spec. Exit code 1 on any error.

### `lint`

```text
skill-auditor lint [skills-dir] [flags]

Flags:
  --repo-root  repo root directory (auto-detected if omitted)
```

Checks each skill for a `SKILL.md`, a frontmatter block, and correct script shebangs. Prints
`MISSING_SKILL`, `NO_FRONTMATTER`, `BAD_SHEBANG` tags per issue. Exits with the issue count (0 = clean).

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

`--semantic` extracts TF-IDF top keywords scored against the full skill corpus. `--patterns` runs
rule-based detectors for required sections, trigger-word frequency, structural conformance, and
anti-pattern signals. Default `--pipeline` runs both and writes a combined report.

### `init`

```text
skill-auditor init [flags]

Flags:
  --agent   agent(s) to install into (default: auto-detect from installed environments)
  --global  install to global skill directory (~/<agent>/skills/)
  --method  installation method: symlink or copy (default: symlink)
```

Installs the embedded `skill-quality-auditor` SKILL.md and its `references/` directory into one or
more agent skill directories. Auto-detects supported environments (Claude Code, Cursor, etc.) when
`--agent` is omitted.

### `update`

```text
skill-auditor update [flags]

Flags:
  --check           report the latest version without installing
  --version-target  install a specific version (e.g. v1.2.3)
```

Fetches the latest release from GitHub and replaces the running binary in-place. Only applicable
when installed via `install.sh` — Homebrew and mise users should use their own update commands.

### `prune`

```text
skill-auditor prune [flags]

Flags:
  --keep       number of audit date-dirs to retain per skill (default 5)
  --repo-root  repo root directory (auto-detected if omitted)
```

Removes old date-stamped audit directories from `.context/audits/`, keeping the N most recent per
skill.

---

## Output Formats

All commands that produce structured data support `--json`. The text format is the default and is
optimised for terminal readability.

**JSON output** — pass `--json` to any command:

```bash
skill-auditor evaluate skills/my-skill --json
skill-auditor batch skills/skill-a skills/skill-b --json
skill-auditor trend --json
```

**Stored output** — pass `--store` to persist results for later use by `remediate` and `trend`:

| Command | Output path |
| --- | --- |
| `evaluate --store` / `batch --store` | `.context/audits/<skill>/<date>/` |
| `duplication` | `.context/analysis/duplication-report-YYYY-MM-DD.md` |
| `aggregate` | `.context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md` |
| `remediate` | `.context/plans/<skill>-remediation-plan.md` |
| `analyze --store` | `.context/analysis/pattern-report-<skill>-YYYY-MM-DD.md` |

---

## CI Integration

```yaml
- name: Audit skills
  run: |
    skill-auditor batch skills/ --fail-below B --store
    skill-auditor duplication   # exits 2 on Critical pairs
    skill-auditor validate artifacts
```

`--fail-below` accepts any grade: `A+`, `A`, `B+`, `B`, `C+`, `C`, `D`, `F`.

`duplication` exits with code 2 (not 1) on Critical pairs so it can be distinguished from a
command error in pipeline logic.

Full workflow example:

```yaml
jobs:
  skill-quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install skill-auditor
        run: curl -fsSL https://raw.githubusercontent.com/pantheon-org/skill-quality-auditor/main/scripts/install.sh | sh

      - name: Batch audit (fail below B)
        run: skill-auditor batch skills/ --fail-below B --store

      - name: Duplication check
        run: skill-auditor duplication
        continue-on-error: false   # exits 2 on Critical pairs

      - name: Artifact validation
        run: skill-auditor validate artifacts
```

---

## What it scores & why

| ID | Dimension | Max | What a low score signals |
| -- | --------- | --- | ------------------------ |
| D1 | Knowledge Delta | 20 | Content restates what the model already knows — no expert uplift |
| D2 | Mindset & Procedures | 15 | Missing mental models or step-by-step guidance the agent needs |
| D3 | Anti-Pattern Coverage | 15 | Common failure modes not called out — agent will repeat them |
| D4 | Specification Compliance | 15 | Frontmatter, structure, or naming deviates from the tile spec |
| D5 | Progressive Disclosure | 15 | Detail is front-loaded; references not used for depth |
| D6 | Freedom Calibration | 15 | Skill is either too prescriptive or too vague for the task |
| D7 | Pattern Recognition | 10 | No trigger conditions — agent won't know when to activate the skill |
| D8 | Practical Usability | 15 | Examples absent or unrealistic; hard to apply in practice |
| D9 | Eval Validation | 20 | No evals — quality claims are unverifiable |

**Total: 140 pts.** Grade bands:

| Grade | Score |
| ----- | ----- |
| A+ | ≥ 133 |
| A | ≥ 126 |
| B+ | ≥ 119 |
| B | ≥ 112 |
| C+ | ≥ 105 |
| C | ≥ 98 |
| D | ≥ 91 |
| F | < 91 |

See `cmd/assets/references/quality-thresholds-scoring.md` for the full rubric and per-dimension
scoring criteria.

---

## Repository layout

```text
go.mod / main.go      Go CLI root — build and run from here
cmd/                  cobra commands: evaluate, batch, duplication, aggregate,
                      remediate, trend, validate, lint, prune, analyze, init, update
cmd/assets/           Tessl tile — SKILL.md, tile.json, evals, references,
                      schemas, templates (single source of truth)
agents/               agent registry (supported environments for init)
scorer/               D1–D9 dimension scorers
analysis/             TF-IDF keyword extractor + rule-based pattern detectors
duplication/          word-level Jaccard similarity engine
reporter/             text/JSON formatters, audit store, report generators
scripts/              install.sh
testdata/             fixture skills for unit tests
```

---

## Development

```bash
go test ./...
go vet ./...
golangci-lint run ./...
shellcheck scripts/install.sh
```

Pre-commit and pre-push hooks are managed via [lefthook](https://github.com/evilmartians/lefthook):

```bash
mise install   # installs go, golangci-lint, markdownlint-cli2, shellcheck, lefthook
lefthook install
```

---

### Academic References

- **D6 Freedom Calibration:** [Zhang et al., 2025 — Reasoning over Boundaries: Enhancing Specification Alignment via Test-time Deliberation](https://arxiv.org/abs/2509.14760)
- **D6 Freedom Calibration:** [Sorensen, 2026 — Specification as the New Management](https://www.researchgate.net/publication/401626622)
- **D6 Freedom Calibration:** [Tao, 2025 — LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Multi-Model Collaboration](https://www.researchsquare.com/article/rs-9323974/latest)
