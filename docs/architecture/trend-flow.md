# Trend tracking

The `trend` command tracks score changes over time by comparing stored audits.

## Pipeline

```text
trend [--store] [--json]
  │
  ├── read .context/audits/ directory
  │
  ├── groupAuditsBySkill(auditsRoot)
  │     └── walk audits/<skill>/<date>/audit.json
  │     └── group by skill key
  │
  ├── collectTrends(auditsRoot)
  │     └── for each skill with ≥ 2 audits:
  │           └── buildTrendEntry(skill, paths)
  │                 ├── load two most recent audits
  │                 ├── compute delta = newScore - oldScore
  │                 └── return TrendEntry
  │
  ├── sort alphabetically by skill name
  │
  ├── render:
  │     ├── Markdown table with ↑/↓/— arrows
  │     └── JSON array
  │
  └── if --store: write to .context/audits/trend-YYYY-MM-DD.md
```

## Trend entry structure

```go
type TrendEntry struct {
    Skill     string
    OldDate   string
    NewDate   string
    OldScore  int
    NewScore  int
    OldGrade  string
    NewGrade  string
    Delta     int
    Trend     string // "↑", "↓", or "—"
}
```

The `trendArrow(delta)` function returns:
- `↑` when delta > 0
- `↓` when delta < 0
- `—` when delta == 0

## Source files

| File | Purpose |
|------|---------|
| `cmd/trend.go` | Entry point, grouping, delta computation, rendering |
