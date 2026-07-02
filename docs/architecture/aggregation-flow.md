# Aggregation planning

The `aggregate` command generates a consolidation plan for a family of related
skills (sharing a common prefix).

## Pipeline

```text
aggregate --family <prefix> [--dry-run]
  │
  ├── resolve repo root, skills directory
  │
  ├── duplication.Inventory(skillsDir) → all skills
  │
  ├── filter to family members (prefix match on skill base name)
  │
  ├── duplication.Detect(familyEntries) → pairwise similarity
  │
  ├── compute metrics:
  │     ├── skill count
  │     ├── total lines
  │     ├── average pairwise similarity
  │     └── pair count
  │
  ├── aggregation decision:
  │     ├── AGGREGATE — 2+ reasons or 1 reason + strong signal
  │     ├── CONSIDER — 1 reason
  │     └── MONITOR — no reasons
  │
  ├── render plan:
  │     ├── Markdown: full 6-step process with effort estimates
  │     └── JSON: structured plan data
  │
  └── if not --dry-run: write to .context/analysis/aggregation-plan-<family>-YYYY-MM-DD.md
```

## Aggregation decision logic

Reasons that trigger aggregation:

| Condition | Reason |
|-----------|--------|
| Any pair ≥ critical threshold (0.35) | "Critical similarity detected" |
| ≥ 3 skills AND avg similarity ≥ 0.20 | "Multiple skills with high similarity" |
| ≥ 3 skills | "Multiple skills may cause user confusion" |
| Total lines > 2000 | "Large documentation surface area" |

Decision rules:
- **AGGREGATE**: ≥ 2 reasons, or 1 reason + strong signal (avg similarity ≥ 0.25)
- **CONSIDER**: exactly 1 reason
- **MONITOR**: no reasons

## 6-step aggregation process

When the decision is AGGREGATE, a standard 6-step consolidation workflow is generated:

1. **Inventory** — catalogue all skill files, map overlapping concepts
2. **Draft merging** — produce a consolidated SKILL.md with unified triggers/examples
3. **Split by use case** — verify merged skill addresses distinct use cases
4. **Re-evaluate** — run `skill-auditor evaluate` on the consolidated skill
5. **Update references** — patch downstream references to point to consolidated skill
6. **Archive** — deprecate original skills, remove obsolete entries

Each step includes effort estimates (S/M/L), time ranges, and verification commands.

## Source files

| File | Purpose |
|------|---------|
| `cmd/aggregate.go` | Command entry point |
| `reporter/aggregation.go` | Decision logic, plan formatting, effort estimation |
