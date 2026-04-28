---
category: advanced
priority: LOW
source: historical quality tracking
---

# Trends Analysis

Historical tracking and trend analysis for skill quality metrics over time.

## Overview

**Purpose**: Track skill collection health over time  
**Value**: Identify regressions early, measure improvement efforts  
**Frequency**: Weekly data points, monthly trend reports

## Data Collection

### Weekly Snapshot

```bash
# Generate weekly snapshot via trend command
skill-auditor trend --json
```

**Captured Data:**

- Total skill count
- Grade distribution
- Average score
- Duplication percentage
- File size statistics

### Snapshot Format

```json
{
  "date": "2026-02-20",
  "metrics": {
    "total_skills": 50,
    "grade_distribution": {
      "A": 36,
      "B": 12,
      "C": 2
    },
    "average_score": 98.2,
    "duplication_avg": 12.3,
    "avg_file_size": 180,
    "max_file_size": 450
  },
  "changes": {
    "skills_added": 2,
    "skills_deprecated": 1,
    "skills_improved": 5,
    "skills_regressed": 0
  }
}
```

## Trend Analysis

### Quality Score Trend

```text
Average Score (12 weeks)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
120 ┤
110 ┤                    ╭────╮
100 ┤              ╭─────╯    ╰────╮
 90 ┤        ╭─────╯                ╰────
 80 ┤  ╭─────╯
    └──┴──────┴──────┴──────┴──────┴──────
      Jan 6  Jan 20  Feb 3   Feb 17  Mar 3

Analysis: Steady improvement (+12 points over 12 weeks)
Driver: BDD family consolidation completed Jan 20
```

### Grade Distribution Trend

```text
A-Grade Rate (12 weeks)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
100% ┤                              ╭────
 90% ┤                        ╭─────╯
 80% ┤                  ╭─────╯
 70% ┤            ╭─────╯
 60% ┤      ╭─────╯
     └──────┴──────┴──────┴──────┴──────
       Jan 6  Jan 20  Feb 3   Feb 17

Analysis: A-grade rate increased 60% → 90%
Milestone: All B-grade skills improved to A by Feb 3
```

### Duplication Trend

```text
Average Duplication (12 weeks)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
35% ┤╮
30% ┤╰╮
25% ┤ ╰╮
20% ┤  ╰─╮
15% ┤    ╰──╮
10% ┤       ╰───╮
 5% ┤           ╰──────
    └──────────────────
      Jan 6  Feb 3  Mar 3

Analysis: Duplication reduced 35% → 5%
Action: Completed 3 aggregations during period
```

## Trend Reports

### Monthly Trend Report

```markdown
# Skill Quality Trend Report - February 2026

## Executive Summary

Skill collection quality improved significantly this month:
- A-grade rate: 60% → 90% (+30%)
- Average score: 86 → 102 (+16 points)
- Duplication: 25% → 8% (-17%)

## Key Improvements

### BDD Family Consolidation
- Consolidated 6 skills → 1 aggregation
- Size reduction: 2,032 lines → 64 lines (96.8%)
- Score improvement: 90 → 98 (+8 points)

### TypeScript Family Consolidation
- Consolidated 5 skills → 1 aggregation
- Size reduction: 3,372 lines → 87 lines (97.4%)
- Score improvement: 88 → 96 (+8 points)

## Regressions

None this month.

## Recommendations

1. Continue with Bun family consolidation (next month)
2. Address remaining 2 C-grade skills
3. Maintain weekly monitoring

## Next Month Goals

- A-grade rate: 90% → 95%
- Duplication: 8% → <5%
- Complete 2 more aggregations
```

### Regression Detection

**Alert Conditions:**

| Condition | Threshold | Action |
| --------- | --------- | ------ |
| Score drop | >5 points | Investigate immediately |
| Grade drop | Any | Review within 24 hours |
| Duplication increase | >3% | Schedule remediation |
| New C-grade | Any | Priority fix |

**Regression Report:**

```markdown
# Regression Alert - 2026-02-15

## Issue
Skill `api-testing` regressed from A (102) to B (94)

## Analysis
- D1 (Knowledge Delta): 18 → 14 (-4)
- Cause: Added tutorial content in recent update
- Commit: abc123 "Add getting started guide"

## Recommendation
Move tutorial content to references/, restore expert focus

## Impact
- Users may receive less focused guidance
- Collection A-rate dropped 72% → 70%
```

## Historical Data Storage

### Directory Structure

```text
.context/
├── analysis/
│   ├── snapshots/
│   │   ├── 2026-01-01.json
│   │   ├── 2026-01-08.json
│   │   └── ...
│   ├── trends/
│   │   ├── 2026-01-monthly.md
│   │   └── 2026-02-monthly.md
│   └── current/
│       └── metrics.json
└── history/
    └── skill-changes/
        ├── 2026-01/
        └── 2026-02/
```

### Data Retention

| Data Type | Retention | Reason |
| --------- | --------- | ------ |
| Weekly snapshots | 1 year | Trend analysis |
| Monthly reports | 3 years | Historical record |
| Regression alerts | 1 year | Pattern analysis |
| Full audits | 6 months | Storage efficiency |

## Comparative Analysis

### Period Comparison

```bash
# Compare trends across stored audit snapshots
skill-auditor trend
```

**Output:**

```markdown
## Period Comparison: January vs February 2026

| Metric | January | February | Change |
|--------|---------|----------|--------|
| Total Skills | 55 | 50 | -5 (consolidated) |
| A-Grade Rate | 60% | 90% | +30% |
| Avg Score | 86 | 102 | +16 |
| Duplication | 25% | 8% | -17% |
| Avg File Size | 320 | 180 | -44% |

## Key Events
- Jan 15: BDD family consolidated (6 → 1)
- Jan 28: TypeScript family consolidated (5 → 1)
- Feb 5: 3 skills deprecated
```

### Year-over-Year

```markdown
## Year-over-Year Comparison

| Metric | Feb 2025 | Feb 2026 | Change |
|--------|----------|----------|--------|
| Skills | 72 | 50 | -30% |
| A-Grade | 45% | 90% | +45% |
| Avg Score | 78 | 102 | +24 |
| Duplication | 42% | 8% | -34% |

## Major Improvements
1. Implemented aggregation pattern
2. Established quality gates
3. Automated auditing
4. Reduced technical debt
```

## Predictive Analysis

### Trend Projection

```text
Based on current improvement rate:
- A-grade rate: 90% → 95% (projected: March 15)
- Duplication: 8% → 5% (projected: March 1)
- Avg score: 102 → 108 (projected: April 1)
```

### Resource Planning

```markdown
## Upcoming Consolidations

| Family | Skills | Est. Effort | ROI |
|--------|--------|-------------|-----|
| bun-* | 6 | 4 hours | High |
| biome-* | 2 | 2 hours | Medium |
| markdown-* | 4 | 3 hours | High |

Total estimated effort: 9 hours
Expected improvement: +8 avg score, -5% duplication
```

## See Also

- `reporting-analysis.md` - Report interpretation
- `reporting-dashboards.md` - Visualization
- `advanced-custom-metrics.md` - Custom tracking
