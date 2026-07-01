# skill-quality-auditor

[![CI](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml/badge.svg)](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pantheon-org/skill-quality-auditor)](https://goreportcard.com/report/github.com/pantheon-org/skill-quality-auditor)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pantheon-org/skill-quality-auditor)](https://github.com/pantheon-org/skill-quality-auditor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

AI agent skills promise expert guidance — but how do you know they're any good? `skill-auditor` scores SKILL.md files against a 9-dimension quality framework and produces concrete diagnostics to make them better. Think of it as a linter with opinions: it catches structural issues, gaps in guidance, missing evals, and anti-patterns before your users do.

## Quick Start

```bash
# Install (one line)
curl -fsSL https://raw.githubusercontent.com/pantheon-org/skill-quality-auditor/main/scripts/install.sh | sh

# Evaluate a skill
skill-auditor evaluate path/to/skill
```

```text
Skill: skill-full
Grade: B+ (123/140)

Dimensions:
  Knowledge Delta              20/20
  Mindset + Procedures          7/15
  Anti-Pattern Quality          9/15
  Specification Compliance     16/15
  Progressive Disclosure       15/15
  Freedom Calibration          13/15
  Pattern Recognition          10/10
  Practical Usability          15/15
  Eval Validation              18/20

Warnings:
  [D2]  no precondition signals detected
  [D2]  no postcondition signals detected
  [D7]  no negative anchors in description — skill may over-trigger
```

Each warning links to a dimension doc with scoring criteria and targeted remediation advice.

## Install

| Method | Command |
| --- | --- |
| **install.sh** (Linux / macOS) | [one-liner in Quick Start](#quick-start) |
| **mise** | `mise use github:pantheon-org/skill-quality-auditor` |
| **Go** | `go install github.com/pantheon-org/skill-quality-auditor@latest` |

Set `INSTALL_DIR=~/.local/bin` or `VERSION=v1.2.3` to customise the install.sh download.

Once installed, run `skill-auditor update` (or `mise upgrade skill-auditor`) to get the latest release.

**Prerequisites:** Pre-built binaries need no runtime. To build from source you need Go 1.25+.

## Usage

`skill-auditor <command> [<args>] [flags]`

| Command | What it does |
| --- | --- |
| `evaluate <skill>` | Score one skill |
| `batch <skill1> [skill2 ...]` | Score several skills (`--fail-below` for CI gating) |
| `duplication [dir]` | Detect overlapping skills via Jaccard similarity |
| `aggregate --family <prefix>` | Plan consolidation for a skill family |
| `remediate <skill>` | Generate step-by-step fix recommendations |
| `trend` | Compare score deltas across stored audits |
| `validate` | Check file conventions (`artifacts` subcommand) and review reports (`review`) |
| `analyze <skill>` | Extract TF-IDF keywords and structural patterns |
| `init` | Install the auditor skill into your agent environment |
| `update` | Self-update the binary (install.sh installs only) |
| `prune` | Remove old audit snapshots, keeping N per skill |

### Common flags

| Flag | Available on |
| --- | --- |
| `-j / --json` | analyze, batch, duplication, trend, remediate, aggregate |
| `-m / --markdown` | evaluate, analyze, batch, duplication, trend |
| `-s / --store` | evaluate, analyze, batch, duplication, trend |
| `-r / --repo-root` | most commands |
| `-n / --dry-run` | aggregate, remediate, prune, init |

`--json` and `--markdown` are mutually exclusive. Default format: JSON for evaluate/analyze/batch/remediate/aggregate, Markdown for duplication/trend. Run any command with `--help` for the full flag reference.

```bash
skill-auditor evaluate path/to/skill              # JSON (default)
skill-auditor evaluate path/to/skill -m            # Markdown
skill-auditor batch skills/skill-a skills/skill-b --json
skill-auditor duplication --store                  # persist for trend
```

## Scoring Dimensions

| ID | Dimension | Max | What a low score signals |
| -- | --------- | --- | ------------------------ |
| D1 | Knowledge Delta | 20 | Content restates what the model already knows — no expert uplift |
| D2 | Mindset & Procedures | 15 | Missing mental models or step-by-step guidance the agent needs |
| D3 | Anti-Pattern Coverage | 15 | Common failure modes not called out — agent will repeat them |
| D4 | Specification Compliance | 15 | Frontmatter, structure, or naming deviates from the tile spec |
| D5 | Progressive Disclosure | 15 | Detail is front-loaded; references not used for depth |
| D6 | Freedom Calibration | 15 | Skill is either too prescriptive or too vague for the task |
| D7 | Pattern Recognition | 10 | No trigger conditions — agent won't know when to activate |
| D8 | Practical Usability | 15 | Examples absent or unrealistic; hard to apply in practice |
| D9 | Eval Validation | 20 | No evals — quality claims are unverifiable |

**Total: 140 pts.** Grade bands: A+ ≥133, A ≥126, B+ ≥119, B ≥112, C+ ≥105, C ≥98, D ≥91, F <91.

See `cmd/assets/references/quality-thresholds-scoring.md` for the full rubric. Each dimension has a dedicated doc: [D1](docs/d1-knowledge-delta.md) · [D2](docs/d2-mindset-procedures.md) · [D3](docs/d3-anti-pattern-coverage.md) · [D4](docs/d4-specification-compliance.md) · [D5](docs/d5-progressive-disclosure.md) · [D6](docs/d6-freedom-calibration.md) · [D7](docs/d7-pattern-recognition.md) · [D8](docs/d8-practical-usability.md) · [D9](docs/d9-eval-validation.md)

## CI Integration

```yaml
- name: Audit skills
  run: |
    skill-auditor batch skills/ --fail-below B --store
    skill-auditor duplication   # exits 2 on Critical pairs
    skill-auditor validate artifacts
```

`--fail-below` accepts any grade (A+ through F). `duplication` exits with code 2 (not 1) on Critical (>35%) pairs so it can be distinguished from a command error in pipeline logic.

Full workflow example:

```yaml
jobs:
  skill-quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install
        run: curl -fsSL https://raw.githubusercontent.com/pantheon-org/skill-quality-auditor/main/scripts/install.sh | sh
      - name: Batch audit
        run: skill-auditor batch skills/ --fail-below B --store
      - name: Duplication check
        run: skill-auditor duplication
      - name: Artifact validation
        run: skill-auditor validate artifacts
```

## Repository layout

```text
.
├── main.go               CLI entrypoint
├── cmd/                  Cobra command implementations
├── cmd/assets/           Tessl tile — SKILL.md, tile.json, evals, schemas
├── agents/               Agent registry (supported init targets)
├── scorer/               D1–D9 dimension scorers
├── analysis/             TF-IDF extraction + pattern detectors
├── duplication/          Jaccard similarity engine
├── reporter/             Output formatters and audit persistence
├── docs/                 Per-dimension scoring docs and ADRs
├── internal/             Shared utilities (tokenizer)
├── scripts/              install.sh
└── testdata/             Fixture skills for unit tests
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow, commit conventions, and pre-commit hook setup. All tests must pass (`go test ./...`) before pushing.

## License

MIT — see [LICENSE](LICENSE) for the full text.
