# Scenario 01: Skill Quality Assessment

## User Prompt

"Audit this SKILL.md and produce a scored quality report."

## Expected Behavior

1. Apply the 9-dimension framework (D1–D9) to the provided `skills/sql-query-builder/SKILL.md` content.
2. Score each dimension with a numerical value (out of its max) and a brief justification.
3. Flag the installation guide, SQL syntax examples, and generic best-practices bullets as redundant
   content that provides no knowledge delta for an expert audience.
4. Identify the weak description field (`"Help with SQL queries."`) as a D4 specification compliance failure.
5. Note the absence of a `references/` directory and content frontloading as a D5 progressive disclosure gap.
6. State the A-grade threshold (>=126/140) as the target.
7. Produce `audit-report.md` with a per-dimension score table summing to a total out of 140.
8. Produce `remediation-plan.md` listing specific file-level changes (what to add, remove, or rewrite)
   with S/M/L effort sizing for each item.

## Success Criteria

- All 9 dimensions scored with a numerical value and brief per-dimension justification.
- D1 Knowledge Delta assigned <=10/20, reflecting the high ratio of redundant content.
- SQL basics, installation instructions, and generic best-practices identified as redundant (not worth expert attention).
- A-grade threshold stated as >=126/140 (or equivalent percentage).
- Remediation plan contains specific file-level changes with S/M/L effort sizing.
- Weak description field (`"Help with SQL queries."`) flagged as a D4 compliance failure.
- Absence of `references/` directory or content frontloading identified as a D5 weakness.

## Failure Conditions

- Dimensions scored without per-dimension justification.
- D1 assigned a high score (>12/20) despite predominantly redundant content.
- Installation steps and generic best-practices not flagged as low-delta content.
- Remediation plan consists of vague suggestions without file-level specifics or effort estimates.
- A-grade threshold not referenced.
- D4 and D5 weaknesses not identified.

**Input:**

~~~markdown
---
name: sql-query-builder
description: "Help with SQL queries."
---

# SQL Query Builder

SQL (Structured Query Language) is used to interact with relational databases.

## Installation

Install the database driver:
```bash
pip install psycopg2
```

## Basic Usage

SELECT statement:
```sql
SELECT * FROM users;
```

INSERT statement:
```sql
INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com');
```

## Best Practices

- Use indexes for better performance
- Avoid SELECT * in production
- Use parameterized queries to prevent SQL injection
~~~
