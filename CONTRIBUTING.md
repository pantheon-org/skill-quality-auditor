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
go build -o skill-auditor .
```

Run a quick smoke test against the fixture skills:

```bash
./skill-auditor evaluate testdata/fixtures/skill-full
./skill-auditor batch testdata/fixtures/skill-minimal testdata/fixtures/skill-full
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

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scorer): add D10 coherence dimension
fix(reporter): handle missing evals directory gracefully
docs: update scoring rubric reference
```

## Pull requests

- Keep PRs focused on a single concern.
- All tests must pass (`go test ./...`).
- Update `README.md` and any affected `skill/skill-quality-auditor/references/` docs.
