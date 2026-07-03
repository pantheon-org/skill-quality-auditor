# skill-quality-auditor

[![CI](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml/badge.svg)](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml)
[![golangci-lint](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml/badge.svg?job=lint)](https://github.com/pantheon-org/skill-quality-auditor/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pantheon-org/skill-quality-auditor)](https://github.com/pantheon-org/skill-quality-auditor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![skill-quality-auditor](assets/logo.png)

AI agent skills promise expert guidance — but how do you know they're any good? `skill-auditor` scores SKILL.md files against a 9-dimension quality framework and produces concrete diagnostics to make them better. Think of it as a linter with opinions: it catches structural issues, gaps in guidance, missing evals, and anti-patterns before your users do.

## Quick start

```bash
go build -o dist/skill-auditor .
./dist/skill-auditor evaluate <path-or-key>
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

## Install

| Method | Command |
| --- | --- |
| **install.sh** (Linux / macOS) | `curl -fsSL <https://raw.githubusercontent.com/pantheon-org/skill-quality-auditor/main/scripts/install.sh> | sh` |
| **mise** | `mise use github:pantheon-org/skill-quality-auditor` |
| **Go** | `go install github.com/pantheon-org/skill-quality-auditor@latest` |

## Commands

| Command | Purpose |
| ------- | ------- |
| `evaluate` | Score a single skill (D1–D9, grade A+ to F) |
| `batch` | Score multiple skills, sorted by grade |
| `duplication` | Detect overlapping skills (Jaccard similarity) |
| `aggregate` | Generate consolidation plan for a skill family |
| `remediate` | Generate or validate a remediation plan |
| `trend` | Track score changes over time |
| `eval` | Run LLM-based eval scenarios against a skill |
| `analyze` | TF-IDF and pattern analysis for a single skill |
| `validate` | Check skill artifact conventions |
| `init` | Install the skill into agent harness directories |
| `update` | Self-update the binary from GitHub releases |
| `prune` | Remove old audit directories, keep N per skill |

## Architecture documentation

- **[Architecture overview](architecture/overview.md)** — high-level package layout and data flow
- **[Evaluate flow](architecture/evaluate-flow.md)** — the core scoring pipeline
- **[Batch flow](architecture/batch-flow.md)** — multi-skill evaluation
- **[Duplication detection](architecture/duplication-flow.md)** — Jaccard-based overlap analysis
- **[Aggregation planning](architecture/aggregation-flow.md)** — skill family consolidation
- **[Remediation flow](architecture/remediation-flow.md)** — generating and validating plans
- **[Trend tracking](architecture/trend-flow.md)** — score history and deltas
- **[Eval runner](architecture/eval-runner.md)** — LLM-based scenario evaluation
- **[Init, update, prune](architecture/init-update-prune.md)** — lifecycle commands
- **[Validate & analyze](architecture/validate-analyze.md)** — artifact validation and static analysis
- **[Scoring dimensions](reference/scoring-dimensions.md)** — D1–D9 reference
  - [D1: Knowledge Delta](reference/d1-knowledge-delta.md)
  - [D2: Mindset & Procedures](reference/d2-mindset-procedures.md)
  - [D3: Anti-Pattern Coverage](reference/d3-anti-pattern-coverage.md)
  - [D4: Specification Compliance](reference/d4-specification-compliance.md)
  - [D5: Progressive Disclosure](reference/d5-progressive-disclosure.md)
  - [D6: Freedom Calibration](reference/d6-freedom-calibration.md)
  - [D7: Pattern Recognition](reference/d7-pattern-recognition.md)
  - [D8: Practical Usability](reference/d8-practical-usability.md)
  - [D9: Eval Validation](reference/d9-eval-validation.md)
- **[Development setup](development/setup.md)** — prerequisites and workflow
- **[Adding a scorer](development/adding-a-scorer.md)** — how to extend the framework
- **[Adding an agent](development/adding-an-agent.md)** — how to support a new harness
- **[Skills and rules](development/skills-and-rules.md)** — local agent rules and skills

## Key packages

| Package | Path | Responsibility |
| ------- | ---- | -------------- |
| `cmd` | `cmd/` | Cobra CLI commands, asset embedding |
| `scorer` | `scorer/` | D1–D9 scoring engine |
| `reporter` | `reporter/` | Formatting, persistence, plans |
| `duplication` | `duplication/` | Inventory, pairwise Jaccard detection |
| `analysis` | `analysis/` | TF-IDF keywords, rule-based pattern detection |
| `agents` | `agents/` | Agent registry for the `init` command |
| `internal/llmclient` | `internal/llmclient/` | Provider-agnostic LLM client |
| `internal/patternconfig` | `internal/patternconfig/` | Loads externalised D1/D6/analysis-quality pattern words from YAML |
| `internal/tokenize` | `internal/tokenize/` | Text normalization and tokenization |
