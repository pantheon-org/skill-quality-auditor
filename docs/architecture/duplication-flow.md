# Duplication detection

The `duplication` command identifies overlapping skills via pairwise
Jaccard similarity on word tokens and structural headers.

## Pipeline

```text
duplication [--store]
  │
  ├── resolve skills directory (<repoRoot>/skills by default)
  │
  ├── duplication.Inventory(skillsDir)
  │     └── filepath.WalkDir → find all SKILL.md files
  │     └── returns []SkillEntry{Key, Path, Content}
  │
  ├── duplication.Detect(entries)
  │     └── O(n²) pairwise loop (capped at 500 entries)
  │     └── applies thresholds:
  │           Critical ≥ 0.35
  │           High     ≥ 0.20
  │     └── sorts descending by similarity
  │
  ├── reporter.DuplicationReport(pairs, entries, date)
  │     └── Markdown: summary stats, grouped by family, top pairs
  │     └── or JSON
  │
  ├── if --store: writes to .context/analysis/duplication-report-YYYY-MM-DD.md
  │
  └── exit code 2 if any Critical pairs found
```

## Similarity algorithm

The composite similarity score has two components:

```text
Similarity = 0.7 × WordJaccard + 0.3 × HeaderJaccard
```

### Word-level Jaccard (70%)

1. `tokenize.Normalize(text)` — strips markdown formatting (`#`, `*`, backtick),
   lowercases, removes stopwords & short tokens, trims punctuation
2. `tokenize.Set(text)` — unique normalized tokens
3. `Jaccard(a, b) = |intersection| / |union|`

### Structural header Jaccard (30%)

1. Regex extract of `#` to `######` header text
2. `Jaccard(headerSet(a), headerSet(b))`

## Thresholds

| Threshold | Value | Severity | Exit code |
|-----------|-------|----------|-----------|
| Critical  | ≥ 0.35 | "Critical" | 2 |
| High      | ≥ 0.20 | "High"     | 0 |
| Low       | < 0.20 | ignored    | 0 |

Maximum 500 entries (configurable via `MaxDetectEntries`).

## Source files

| File | Purpose |
|------|---------|
| `duplication/inventory.go` | Directory walk, SkillEntry type |
| `duplication/detect.go` | Pairwise comparison, thresholds |
| `duplication/similarity.go` | Jaccard, TokenSet, SectionHeaders |
| `cmd/duplication.go` | Command entry point |
| `reporter/duplication.go` | DuplicationReport formatting |
| `internal/tokenize/tokenize.go` | Text normalization |
