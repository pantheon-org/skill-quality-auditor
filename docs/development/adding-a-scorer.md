# Adding a dimension scorer

This guide walks through adding a new dimension (e.g., D10) to the framework.

## Steps

### 1. Create the scorer file

Create `scorer/d10_<name>.go` with the function signature:

```go
package scorer

func scoreD10(content, skillDir string, bridge *validatorBridge) (int, []Diagnostic) {
    // Your scoring logic here
}
```

The function receives:
- `content` — the raw SKILL.md text
- `skillDir` — the directory containing SKILL.md (for reading auxiliary files)
- `bridge` — the validator bridge for cached skill-validator results

Return:
- `int` — the score (0 to max, defined in `AllDimensions`)
- `[]Diagnostic` — error/warning/hint diagnostics

### 2. Register the dimension

Add the new dimension to `AllDimensions` in `scorer/dimensions.go`:

```go
var AllDimensions = []Dimension{
    // ... existing D1–D9
    {Code: "D10", Key: "yourDimensionKey", Label: "Your Dimension Name", Max: 20},
}
```

### 3. Wire it into the scoring registry

In `scorer/scorer.go`, add an entry to the `registry` slice inside `ScoreFromContent`:

```go
registry := []dimensionEntry{
    // ... existing entries
    {AllDimensions[9], func(c, dir string, b *validatorBridge) (int, []Diagnostic) {
        return scoreD10(c, dir, b)
    }},
}
```

The `dimensionFn` type erases signature differences via closure adaptation.

### 4. Update the max-points table

Update the max-points table in `README.md` to include the new dimension.

### 5. Write tests

Create `scorer/d10_<name>_test.go`:

```go
package scorer

import "testing"

func TestScoreD10(t *testing.T) {
    tests := []struct {
        name    string
        content string
        want    int
    }{
        {"basic case", "...", 15},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, diags := scoreD10(tt.content, "", nil)
            if got != tt.want {
                t.Errorf("got %d, want %d; diags: %v", got, tt.want, diags)
            }
        })
    }
}
```

### 6. Update documentation

Add the new dimension to `docs/reference/scoring-dimensions.md`.

### 7. Run tests

```bash
go test ./scorer/...
```

## Scorer conventions

- Use constants from `scorer/thresholds.go` for rubric cut-points
- Use helper functions from `scorer/dimensions.go`:
  - `countPattern(re, content)` — regex match count
  - `countLines(content)` — line count
  - `parseFrontmatter(content)` — YAML frontmatter
  - `codeBlockCount(content)` — code fence count
- Use `errDiag(dim, msg)`, `warnDiag(dim, msg)`, `hintDiag(dim, msg)` for diagnostics
- Return 0–max integer, never negative or exceeding max
- **Pattern/signal word lists should be externalised, not hardcoded.** Per ADR-028, D1 and
  D6 read their beginner/expert signal words and "when not to use" phrases from
  `cmd/assets/assets/config/scoring-patterns.yaml` via `internal/patternconfig`, rather than
  from Go string slices. If your new dimension needs a similar word/phrase list, add a
  section to that YAML file (validated against `scoring-patterns.schema.json`) instead of
  inlining it in the scorer — this keeps pattern tuning out of Go release cycles.
