# Contributing

## Prerequisites

- Go 1.21+
- [Tessl CLI](https://tessl.io) (for skill tile work)

## Development workflow

Always work on a feature branch:

```bash
git checkout -b feat/my-change
```

### CLI (`skill-auditor/`)

```bash
cd skill-auditor
go test ./...
go vet ./...
go build -o bin/skill-auditor .
```

Run a quick smoke test against the fixture skills:

```bash
./bin/skill-auditor evaluate testdata/fixtures/skill-full
./bin/skill-auditor batch testdata/fixtures/skill-minimal testdata/fixtures/skill-full
```

### Tessl skill (`skill/`)

```bash
cd skill/skill-quality-auditor
tessl eval run .
```

Evals must pass before bumping the tile version in `tile.json`.

## Adding a new dimension scorer

1. Create `skill-auditor/scorer/dN_<name>.go` with a `scoreDN(content, skillDir string) (int, []Diagnostic)` function.
2. Add a corresponding `dN_<name>_test.go`.
3. Register the scorer in `scorer/dimensions.go`.
4. Update the max-points table in `README.md`.

## Adding a new analysis detector

Pattern and semantic analysis lives in `skill-auditor/analysis/`.

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
- Update `README.md` and any affected `skill/skill-quality-auditor/references/` docs.
