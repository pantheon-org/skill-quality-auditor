# D3: Anti-Pattern Coverage (15 points)

**Purpose:** Teach what NOT to do, with clear explanations of WHY.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 13–15 | NEVER lists + concrete examples + consequences |
| 10–12 | Has most elements |
| 7–9 | Generic warnings |
| 0–6 | Missing or weak |

## Components

1. **NEVER Lists with WHY (5 points)**
  - Explicit "NEVER do X because Y" statements
  - Use strong language — not just "avoid"
  - Example: "NEVER trust agent completion reports without verification"

1. **Concrete Examples (5 points)**
  - Show bad code, not just descriptions
  - Side-by-side ❌ BAD / ✅ GOOD comparisons
  - Real-world scenarios

1. **Consequences Explained (5 points)**
  - What breaks when the anti-pattern is used
  - Impact: security, performance, maintainability
  - Example: "Leads to SQL injection attacks"

## Example

**Strong Anti-Patterns (14/15):**

```markdown
## Anti-Patterns

❌ **NEVER use string interpolation for SQL**
WHY: Opens SQL injection vulnerabilities

// BAD — vulnerable to injection
db.query(`SELECT * FROM users WHERE id = ${userId}`)

// GOOD — safe with prepared statements
db.query('SELECT * FROM users WHERE id = ?', [userId])

**Consequence:** Attacker can inject `1 OR 1=1` to dump the entire table.

❌ **NEVER skip test failure verification**
WHY: False positives waste hours debugging phantom issues

**Consequence:** Test passes even with bugs, leading to production failures.
```

## Academic References

- [Brada & Picha, 2019 — Software Process Anti-Patterns Catalogue](https://dl.acm.org/doi/abs/10.1145/3361149.3361178)
- [Picha & Brada, 2019 — Software Process Anti-Pattern Detection in Project Data](https://dl.acm.org/doi/abs/10.1145/3361149.3361169)
- [Bhatia, Lin, Rajbahadur, Adams et al., 2024 — Data Quality Anti-Patterns for Software Analytics](https://arxiv.org/abs/2408.12560)
- [Amarasinghe, Asanka et al., 2024 — Code Quality Alarms: Techniques, Datasets, and Emerging Trends in Detecting Smells and Anti-Patterns](https://jdrra.sljol.info/articles/10.4038/jdrra.v3i2.93)
