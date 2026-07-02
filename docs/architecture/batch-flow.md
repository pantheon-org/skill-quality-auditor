# Batch flow

The `batch` command evaluates multiple skills in a single invocation, reporting
a sorted leaderboard.

## Pipeline

```text
batch <skill1> <skill2> ... [--store] [--fail-below B] [--json|--markdown]
  │
  ├── resolve output format, repo root, store flag
  │
  ├── for each arg:
  │     ├── resolveSkillPath(arg, repoRoot)
  │     ├── scorer.Score(ctx, path)
  │     │     └── same pipeline as the evaluate command
  │     └── if store: reporter.Store(...)
  │
  ├── sort entries by Total score descending
  ├── render output
  │     ├── --json:  JSON array of results
  │     └── default: Markdown table with averages
  │
  └── if --fail-below <grade>:
        check all results ≥ threshold grade
        exit 1 if any result is below
```

## Batch entry structure

```go
type batchEntry struct {
    arg    string
    result *scorer.Result
    err    error
}
```

## Fail-below gating

The `--fail-below` flag enforces a minimum grade floor for batch runs:

```bash
./dist/skill-auditor batch skill-a skill-b --fail-below B
```

Uses `scorer.GradeRank` map to compare grades numerically
(A+=8, A=7, B+=6, B=5, etc.). Any result below the threshold causes exit 1.

## Source files

| File | Purpose |
|------|---------|
| `cmd/batch.go` | Command entry point, entry loop, error handling |
