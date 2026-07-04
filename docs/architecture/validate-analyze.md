# Validate & analyze

## Validate command

The `validate` command checks skill artifact conventions and review reports.

### Artifact validation

```text
validate artifacts [paths...]
  │
  └── walks each path (default: repo skills directory)
        ├── Schema files: .schema.json extension, valid JSON,
        │     $schema from json-schema.org
        ├── Template files: valid YAML, non-empty
        ├── Script files: correct shebangs
        │     (sh/bash, python3, bun, node)
        ├── SKILL.md checks:
        │     ├── ≤ 500 lines
        │     ├── frontmatter name matches directory
        │     └── no ../ refs outside code blocks
        └── Assets directory: only templates/, schemas/,
              requirements/, examples/
```

### Review report validation

```text
validate review <report-file>
  │
  └── validates against review-report.requirements.json
        ├── H1 title prefix
        ├── Required frontmatter keys
        ├── Metadata labels
        ├── H2 heading required groups and order
        ├── Dimension labels
        ├── Commands
        └── Recommended items (warnings unless --strict-recommended)
```

## Analyze command

The `analyze` command performs TF-IDF keyword extraction and rule-based pattern
detection on a single skill.

### Modes

| Mode | Flag | What it does |
|------|------|-------------|
| Semantic | `--semantic` | TF-IDF keyword extraction only |
| Patterns | `--patterns` | Rule-based pattern detection only |
| Pipeline | (none, default) | Both combined → `CombinedAnalysis` |

### TF-IDF extraction (`analysis/tfidf.go`)

- Builds a corpus from all skills in the skills directory
- Extracts top-N keywords with TF-IDF scores
- Corpus built via `duplication.Inventory`

### Pattern detection (`analysis/patterns.go`)

```text
DetectRequiredSections(skill)     → when to use, examples, triggers
DetectTriggerFrequency(skill)     → trigger word counts
DetectStructuralConformance(skill)→ canonical section checks
DetectAntiPatternSignals(skill)   → anti-pattern detection
```

`DetectAntiPatternSignals` sources its hedge/vague/passive word lists from
`internal/patternconfig` (the same maintainer-editable YAML config used by D1/D6 —
see [Architecture overview](overview.md) and ADR-028) rather than hardcoded Go slices.

### Output

Output type depends on flags:

- Default: `CombinedAnalysis` struct with keywords + rule matches
- `--json`: JSON output
- `--markdown`: human-readable markdown

Output path: `.context/analysis/pattern-report-<skill>-<date>.md`

## Source files

| File | Purpose |
|------|---------|
| `cmd/validate.go` | Validate command, artifact + review validators |
| `cmd/analyze.go` | Analyze command, TF-IDF + pattern dispatch |
| `analysis/patterns.go` | Rule-based pattern detectors |
| `analysis/tfidf.go` | TF-IDF keyword extraction |
| `reporter/combined_analysis.go` | CombinedAnalysis rendering |
