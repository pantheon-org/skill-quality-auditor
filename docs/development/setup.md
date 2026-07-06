# Development setup

## Prerequisites

- Go 1.25.5+
- [hk](https://github.com/jdx/hk) (hook manager)
- [mise](https://mise.jdx.dev) (recommended for tool version management)
- Node.js 18+ (for docmd documentation preview)

## Quick start

```bash
# Clone the repo
git clone https://github.com/pantheon-org/skill-quality-auditor.git
cd skill-quality-auditor

# Install tools (via mise)
mise install

# Install git hooks
hk install

# Build the CLI
go build -o dist/skill-auditor .

# Run a smoke test
./dist/skill-auditor evaluate testdata/fixtures/skill-full
```

## Git hooks

Pre-commit and pre-push hooks are managed via hk:

```bash
hk run pre-commit   # run the pre-commit steps manually
hk check            # lint without fixing
hk fix              # run fixers and restage
HK=0 git commit ... # bypass hooks for a single commit
```

### Pre-commit checks

- Go fmt, vet, lint (golangci-lint)
- markdownlint (markdownlint-cli2)
- shellcheck on shell scripts
- Context frontmatter validation
- ADR index freshness check (`regenerate-adr-index.sh --check` regenerates and diffs
  `docs/ADR/index.yaml`; a stale index fails, not just a missing one)
- Undocumented decision detection (binding `## Decision` headings not covered by an ADR)

### Pre-push checks

- Full test suite (`go test ./...`)
- Context frontmatter JSON-schema validation (`skill-auditor validate context`) — enforces the schemas' `additionalProperties:false`, catching typo'd/unknown keys the pre-commit shell check cannot; runs alongside it during a proving period
- Plan-drift check (`scripts/check-plan-drift.sh`) — flags active plans whose related files changed after the plan was written
- Docs-drift check (`scripts/check-docs-drift.sh`) — flags docs whose mapped source paths changed after the doc was last updated; both drift checks are informational only (exit 0)
- Binary build
- Artifact validation
- Duplication detection (exits 2 on Critical pairs)
- Batch audit (fails below B grade)
- Structural eval gate (`eval ./cmd/assets --fail-below 0`) — schema/scenario consistency, no LLM key needed

## Common workflows

### Run tests

```bash
go test ./...
```

### Test a specific package

```bash
go test ./scorer/...
go test ./reporter/...
```

### Run a specific test

```bash
go test -run TestGrade ./scorer/
```

### Lint

```bash
go vet ./...
golangci-lint run
```

### Preview documentation

```bash
npx @docmd/core dev    # starts dev server at localhost:3000
```

### Build documentation

```bash
npx @docmd/core build  # outputs static site to ./site/
```

## Configuring scoring patterns

The D1 (knowledge delta), D6 (freedom calibration), and analysis-quality word/phrase lists live in `scoring-patterns.yaml`, embedded in the binary at `cmd/assets/assets/config/scoring-patterns.yaml`. Every scoring command (`evaluate`, `batch`, `analyze`, `duplication`) resolves the active pattern config through a 5-tier precedence chain, highest wins:

1. **`-c/--config <path>`** — an explicit path, persistent across all subcommands. Missing or invalid is a **hard error**.
2. **`./scoring-patterns.yaml`** — an opportunistic file in the current working directory. Missing is silently skipped; malformed warns to stderr and falls through. Never auto-created.
3. **The default per-OS config directory**, auto-generated on first run if nothing else resolves:
   - Linux: `$XDG_CONFIG_HOME/skill-quality-auditor/scoring-patterns.yaml` (or `~/.config/skill-quality-auditor/scoring-patterns.yaml` if `$XDG_CONFIG_HOME` is unset)
   - macOS: `~/Library/Application Support/skill-quality-auditor/scoring-patterns.yaml`
   - Windows: `%AppData%\skill-quality-auditor\scoring-patterns.yaml`
4. **The embedded config** shipped with the binary.
5. **Hardcoded Go defaults**, used only if even the embedded config fails to parse.

Pass `--no-user-config` to skip tiers 1–3 entirely and score with the embedded/hardcoded patterns only (also suppresses auto-generation) — useful on CI runners where a stray config file shouldn't silently change scores.

`skill-auditor eval` is exempt from this chain: it always scores against the embedded config, so `evals/summary.json` and the CI structural eval gate stay reproducible across machines regardless of any local override.

A user config must define every pattern group — partial overrides of a single group are not supported; the file replaces the whole set.

## Project layout

```text
.
├── main.go               # Entry point
├── cmd/                  # Cobra CLI commands + embedded assets
├── scorer/               # D1–D9 scoring engine
├── reporter/             # Formatting, persistence, plans
├── duplication/          # Pairwise similarity detection
├── agents/               # Agent registry
├── analysis/             # TF-IDF + pattern detection
├── internal/
│   ├── llmclient/        # Provider-agnostic LLM client
│   └── tokenize/         # Text normalization
├── docs/                 # Documentation (this site)
│   ├── architecture/     # Code flow documentation
│   ├── reference/        # Dimension reference
│   └── development/      # Development guides
├── docs/ADR/             # Architecture Decision Records
├── cmd/assets/           # Embedded skill assets
│   ├── references/       # Scoring rubrics, anti-patterns, thresholds
│   ├── evals/            # Evaluation scenarios
│   ├── schemas/          # JSON schemas
│   ├── templates/        # Templates
│   └── requirements/     # Requirements
└── testdata/             # Fixture skills for tests
```

## Release workflow

Releases are automated via release-please:

1. Conventional commits trigger release PRs
2. CI builds cross-platform binaries
3. Homebrew tap is updated automatically
4. Tile version is synced with binary version (ADR-023)
