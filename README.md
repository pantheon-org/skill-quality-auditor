# skill-quality-auditor

[![CI](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml/badge.svg)](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml)
[![golangci-lint](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml/badge.svg?job=lint)](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pantheon-org/skill-quality-auditor)](https://github.com/pantheon-org/skill-quality-auditor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![skill-quality-auditor](docs/assets/logo.png)

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
| `validate` | Check file conventions (`artifacts`), review reports (`review`), and context frontmatter at a given path against the JSON schemas (`context <path>`) |
| `analyze <skill>` | Extract TF-IDF keywords and structural patterns |
| `eval <skill>` | Run `evals/` scenarios (structural gate by default; LLM-judged with a provider key) |
| `init` | Install the auditor skill into your agent environment |
| `update` | Self-update the binary (install.sh installs only) |
| `prune` | Remove old audit snapshots, keeping N per skill |
| `version` | Print the version and, for release builds, the release date |

### Common flags

| Flag | Available on |
| --- | --- |
| `-j / --json` | analyze, batch, duplication, trend, remediate, aggregate |
| `-m / --markdown` | evaluate, analyze, batch, duplication, trend |
| `-s / --store` | evaluate, analyze, batch, duplication, trend |
| `-r / --repo-root` | most commands |
| `-n / --dry-run` | aggregate, remediate, prune, init |
| `-c / --config <path>` | global — overrides the D1/D6/analysis-quality scoring pattern lists |
| `--no-user-config` | global — ignore any config file and score with the embedded/built-in patterns only |

`--json` and `--markdown` are mutually exclusive. Default format: JSON for evaluate/analyze/batch/remediate/aggregate, Markdown for duplication/trend. Run any command with `--help` for the full flag reference.

`-c/--config` and `--no-user-config` are persistent flags accepted by every scoring command; `skill-auditor eval` ignores both and always scores against the embedded config, so CI results stay reproducible. See [Configuring scoring patterns](docs/development/setup.md#configuring-scoring-patterns) for the full 5-tier precedence chain and the per-OS default config path.

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

See `cmd/assets/references/quality-thresholds-scoring.md` for the full rubric. Each dimension has a dedicated doc: [D1](docs/reference/d1-knowledge-delta.md) · [D2](docs/reference/d2-mindset-procedures.md) · [D3](docs/reference/d3-anti-pattern-coverage.md) · [D4](docs/reference/d4-specification-compliance.md) · [D5](docs/reference/d5-progressive-disclosure.md) · [D6](docs/reference/d6-freedom-calibration.md) · [D7](docs/reference/d7-pattern-recognition.md) · [D8](docs/reference/d8-practical-usability.md) · [D9](docs/reference/d9-eval-validation.md)

## CI Integration

```yaml
- name: Audit skills
  run: |
    skill-auditor batch skills/ --fail-below B --store
    skill-auditor duplication   # exits 2 on Critical pairs
    skill-auditor validate artifacts
    skill-auditor eval skills/my-skill --fail-below 0   # structural gate, no LLM key needed
```

`--fail-below` accepts any grade (A+ through F) for `batch`, or a percentage score for `eval`.
`duplication` exits with code 2 (not 1) on Critical (>35%) pairs so it can be distinguished from
a command error in pipeline logic. `eval` runs in structural-only mode (schema consistency, no
semantic grading) unless an LLM provider key is set in the environment — see
[docs/architecture/eval-runner.md](docs/architecture/eval-runner.md).

Full workflow example, mirroring this repo's own [`skill-quality.yml`](.github/workflows/skill-quality.yml):

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
      - name: Structural eval gate
        run: skill-auditor eval skills/my-skill --fail-below 0
```

## Repository layout

```text
.
├── main.go               CLI entrypoint
├── cmd/                  Cobra command implementations
├── cmd/assets/           Skill source — SKILL.md, tile.json, evals, schemas
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
