# Git Hooks Setup

This project uses [hk](https://github.com/jdx/hk) as its hook manager (`hk.pkl` in the repo root). This reference covers setup for hk and alternative managers.

## hk (recommended)

```bash
# Install hk (requires Rust or installer)
curl -fsSL https://hk.jdx.dev/install.sh | sh

# Or via mise
mise use -g hk

# Activate hooks in the repo
hk install
```

### Available Hooks

| Hook | Trigger | What it checks |
|------|---------|---------------|
| `go-fmt` | pre-commit | Go files are gofmt-clean |
| `go-vet` | pre-commit | `go vet ./...` passes |
| `golangci-lint` | pre-commit | `golangci-lint run ./...` passes |
| `markdownlint` | pre-commit | Markdown files comply with `.markdownlint.json` |
| `shellcheck` | pre-commit | `scripts/**/*.sh` files pass shellcheck |
| `context-frontmatter` | pre-commit | All `.context/*.md` files have valid YAML frontmatter |
| `context-index` | pre-commit | `.context/index.yaml` exists (auto-fix regenerates it) |
| `adr-frontmatter` | pre-commit | `docs/ADR/adr-*.md` files have valid frontmatter |
| `adr-index` | pre-commit | `docs/ADR/index.yaml` exists (auto-fix regenerates it) |
| `adr-undocumented` | pre-commit | No `.context/` files contain decisions without ADR coverage |
| `go-test` | pre-push | Full test suite passes |
| `go-build` | pre-push | Binary builds successfully |
| `skill-validate` | pre-push | Artifact conventions are valid |
| `skill-duplication` | pre-push | No critical duplication (≥35%) detected |
| `skill-batch` | pre-push | `cmd/assets` scores ≥ B |

### Skipping Hooks

Temporarily bypass hooks (e.g. for WIP commits):

```bash
git commit --no-verify
git push --no-verify
```

## pre-commit Framework

If you use [pre-commit](https://pre-commit.com), adapt the hk checks:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: context-frontmatter
        name: Context Frontmatter Validation
        entry: .agents/skills/context-index/scripts/validate-context-frontmatter.sh
        language: script
        files: ^\.context/.*\.md$
        exclude: ^\.context/(audits|plans)/
      - id: adr-frontmatter
        name: ADR Frontmatter Validation
        entry: .agents/skills/adr-capture/scripts/validate-adr-frontmatter.sh
        language: script
        files: ^docs/ADR/adr-.*\.md$
```

## Lefthook

If you use [lefthook](https://github.com/evilmartians/lefthook), adapt the checks:

```yaml
# lefthook.yml
pre-commit:
  commands:
    context-frontmatter:
      glob: ".context/**/*.md"
      exclude: ".context/(audits|plans)/**"
      run: .agents/skills/context-index/scripts/validate-context-frontmatter.sh {staged_files}
    adr-frontmatter:
      glob: "docs/ADR/adr-*.md"
      run: .agents/skills/adr-capture/scripts/validate-adr-frontmatter.sh {staged_files}
    adr-index:
      glob: "docs/ADR/adr-*.md"
      run: .agents/skills/adr-capture/scripts/regenerate-adr-index.sh
    context-index:
      glob: ".context/**/*.md"
      run: .agents/skills/context-index/scripts/regenerate-context-index.sh && test -f .context/index.yaml
```

## CI Integration

These same checks run in CI via `hk check` and `hk fix`. See [Quality Thresholds](quality-thresholds-scoring.md) for CI gate configuration.
