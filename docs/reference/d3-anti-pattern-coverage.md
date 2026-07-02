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

```bibtex
@inproceedings{brada2019catalogue,
  title         = {Software Process Anti-Patterns Catalogue},
  author        = {Brada and Picha},
  year          = {2019},
  booktitle     = {Proceedings of the 2019 European Conference on Software Architecture (ECSA)},
  publisher     = {ACM},
  url           = {<https://dl.acm.org/doi/abs/10.1145/3361149.3361178}>
}

```

```bibtex
@inproceedings{picha2019detection,
  title         = {Software Process Anti-Pattern Detection in Project Data},
  author        = {Picha and Brada},
  year          = {2019},
  booktitle     = {Proceedings of the 2019 European Conference on Software Architecture (ECSA)},
  publisher     = {ACM},
  url           = {<https://dl.acm.org/doi/abs/10.1145/3361149.3361169}>
}

```

```bibtex
@article{bhatia2024dataquality,
  title         = {Data Quality Anti-Patterns for Software Analytics},
  author        = {Bhatia and Lin and Rajbahadur and Adams and others},
  year          = {2024},
  journal       = {arXiv preprint arXiv:2408.12560},
  eprint        = {2408.12560},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2408.12560}>
}

```

```bibtex
@article{amarasinghe2025codequality,
  title         = {Code Quality Alarms: A Review of Techniques, Datasets, and Emerging Trends in Detecting Smells and Anti-Patterns},
  author        = {Y. V. A. Amarasinghe and P. Asanka and others},
  year          = {2025},
  journal       = {Journal of Desk Research Reviews and Analysis},
  volume        = {3},
  number        = {2},
  url           = {<https://jdrra.sljol.info/articles/10.4038/jdrra.v3i2.93>}
}
```
