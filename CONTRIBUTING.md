# Contributing

## Prerequisites

- Go 1.21+
- [Tessl CLI](https://tessl.io) (for skill tile work)

## Development workflow

Always work on a feature branch:

```bash
git checkout -b feat/my-change
```

### CLI

```bash
go test ./...
go vet ./...
go build -o dist/skill-auditor .
```

Run a quick smoke test against the fixture skills:

```bash
./dist/skill-auditor evaluate testdata/fixtures/skill-full
./dist/skill-auditor batch testdata/fixtures/skill-minimal testdata/fixtures/skill-full
```

### Tessl skill (`cmd/assets/`)

```bash
tessl eval run cmd/assets/
```

Evals must pass before bumping the tile version in `cmd/assets/tile.json`.

## Git Hooks

Pre-commit and pre-push hooks are managed via [hk](https://github.com/jdx/hk):

```bash
mise install   # installs go, node, golangci-lint, markdownlint-cli2, shellcheck, hk
               # the postinstall hook runs `hk install` automatically
               # (in an activated mise shell, the enter hook runs `mise install` for you)
```

```bash
# Install hk manually (if not using mise)
curl -fsSL https://hk.jdx.dev/install.sh | sh

# Activate hooks in the repo
hk install
```

Once installed, hooks trigger automatically:

- **pre-commit** — Go fmt/vet/lint, markdownlint, shellcheck, context frontmatter validation, ADR index freshness, and undocumented-decision checks.
- **pre-push** — full test suite, binary build, artifact validation, duplication detection, and batch audit (fails below B grade).

Useful hook commands:

```bash
hk run pre-commit   # run the pre-commit steps manually
hk check            # lint without fixing (same steps as pre-commit)
hk fix              # run fixers (e.g. markdownlint --fix) and restage
HK=0 git commit ... # bypass hooks for a single commit
```

For alternative hook managers (pre-commit, lefthook), see `cmd/assets/references/git-hooks-setup.md`. Markdown is linted with [markdownlint-cli2](https://github.com/DavidAnson/markdownlint-cli2); its rules and ignore globs live in `.markdownlint-cli2.jsonc`. Temporarily bypass all hooks with `git commit --no-verify` or `git push --no-verify`.

## Adding a new `init` target agent

1. Add the agent definition to `agents/registry.go`.
2. Add a test case to `agents/registry_test.go`.
3. Run `go test ./agents/...` to verify.

The `init` command auto-detects agents by checking whether the harness root directory
(first path component of `ProjectPath` / `GlobalPath`, e.g. `.claude/`) exists in the
install target. No changes to `cmd/init.go` are needed when adding a new agent — the
registry drives everything.

## Adding a new dimension scorer

1. Create `scorer/dN_<name>.go` with a `scoreDN(content, skillDir string) (int, []Diagnostic)` function.
2. Add a corresponding `dN_<name>_test.go`.
3. Register the scorer in `scorer/dimensions.go`.
4. Update the max-points table in `README.md`.

## Adding a new analysis detector

Pattern and semantic analysis lives in `analysis/`.

1. Add a new function to `analysis/patterns.go` (structural/rule-based) or `analysis/tfidf.go` (keyword-based).
2. Write tests in the corresponding `_test.go` file. Aim for ≥90% coverage.
3. Wire the new function into `cmd/analyze.go` and expose via a flag if appropriate.
4. The `analysis/` package must remain stdlib-only with no imports from sibling packages.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat(scorer): add D10 coherence dimension
fix(reporter): handle missing evals directory gracefully
docs: update scoring rubric reference
```

## Pull requests

- Keep PRs focused on a single concern.
- All tests must pass (`go test ./...`).
- Update `README.md` and any affected `cmd/assets/references/` docs.
