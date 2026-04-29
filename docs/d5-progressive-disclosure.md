# D5: Progressive Disclosure (15 points)

**Purpose:** Structure content for on-demand loading, not front-loading everything.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 13–15 | Navigation hub + `references/` + categories + lazy-load guidance |
| 10–12 | Some organisation, could improve |
| 7–9 | Everything front-loaded, >300 lines |
| 0–6 | No structure, >500 lines |

## Components

### 1. Navigation Hub Approach (5 points)

- `SKILL.md` is <100 lines
- Overview + when-to-use + reference guide — NOT full content
- Example: `supabase-postgres-best-practices` (65 lines)

### 2. References Directory (4 points)

- Detailed content in `references/*.md`
- Each reference 100–500 lines, focused on ONE topic

### 3. Category Organisation (3 points)

- Files organised by prefix (`principles-`, `patterns-`, etc.)
- Priority labels (CRITICAL, HIGH, MEDIUM, LOW)

### 4. Lazy Loading Guidance (3 points) — REQUIRED

- References table includes a concrete "When to Use" column that tells agents exactly which task triggers loading each reference
- `AGENTS.md` explicitly instructs agents to load only the minimum references needed for the current task
- Each reference entry states a specific, actionable condition rather than a generic description
- **WHY:** Without explicit lazy-load guidance, agents default to loading all references eagerly, wasting context on irrelevant content
- **IMPACT:** Skills without lazy-load guidance consume 3–10× more context than necessary

## References Section Standard

Every `SKILL.md` with references MUST end with a `## References` section using a 3-column Markdown table:

```markdown
## References

| Topic | Reference | When to Use |
| --- | --- | --- |
| Security patterns, caching, and trigger configuration | [Best Practices](references/best-practices.md) | Every time you generate a workflow |
| Pinned action versions and input/output specs | [Common Actions](references/common-actions.md) | When using any public action |
| Official workflow syntax and expression reference | [GitHub Actions Docs](https://docs.github.com/en/actions) | For syntax lookup |
```

| Rule | Requirement |
| ---- | ----------- |
| Heading | Exactly `## References` — no variants (`## Resources`, `## See Also`, etc.) |
| Position | Last H2 section in the file |
| Format | Markdown table with `Topic \| Reference \| When to Use` columns — no bullet lists, no bare URLs |
| Reference column | Every cell MUST be a markdown link `[text](url)` |
| Topic column | One-line description of what the referenced file or resource covers |
| When to Use column | Concrete scenario that tells the agent when to load the reference |
| Sub-sections | Optional H3 headings are allowed to group rows by theme |
| Omission | Allowed only when the skill has nothing to reference (no penalty) |

## Lazy Loading Anti-Patterns

❌ **NEVER list references without "When to Use" conditions** — forces agents to load everything or guess.

❌ **NEVER use vague "When to Use" entries** (`"For scoring"` is not actionable — it does not say when NOT to load).

✅ **Explicit lazy-load conditions:**

```markdown
| Topic | Reference | When to Use |
| --- | --- | --- |
| Per-dimension criteria and bonus rules | [Dimensions](references/dimensions.md) | Evaluating any individual dimension or understanding the rubric |
| Score thresholds and grade bands | [Scoring Rubric](references/scoring.md) | Calculating a total score or assigning a grade — skip if only auditing structure |
```

## Examples

**Excellent Progressive Disclosure (15/15):**

```text
bdd-testing/
├── SKILL.md (64 lines — navigation hub with actionable "When to Use" per reference)
├── AGENTS.md (explicit: "load only references needed for current task")
└── references/
    ├── principles-three-amigos.md (CRITICAL, 250 lines)
    ├── gherkin-syntax.md (HIGH, 180 lines)
    └── practices-tags.md (MEDIUM, 120 lines)
```

**Poor Progressive Disclosure (6/15):**

```text
bdd-testing/
└── SKILL.md (1,800 lines — everything front-loaded)
```

**Missing Lazy Loading (10/15 — loses 3 points):**

```text
bdd-testing/
├── SKILL.md (80 lines — good hub, but no "When to Use" column in references table)
├── AGENTS.md (says "load all references before starting")
└── references/
    ├── principles-three-amigos.md
    └── gherkin-syntax.md
```

## Academic References

- [Springer & Whittaker, 2018 — Progressive Disclosure: Designing for Effective Transparency](https://arxiv.org/abs/1811.02164)
- [Anik & Bunt, 2021 — Designing Effective Training Dataset Explanations: The Impact of Information Depth and Progressive Disclosure](https://dl.acm.org/doi/10.1145/3411764.3445382)
- [Timileyin, 2025 — The Role of Cognitive Load in Shaping Web Usability Requirements](https://papers.ssrn.com/sol3/papers.cfm?abstract_id=5247018)
- [Pastrakis, Konstantakis, Caridakis, 2025 — AI-Enhanced Modular Information Architecture for Cognitive-Efficient User Experiences](https://www.mdpi.com/2078-2489/17/1/92)
