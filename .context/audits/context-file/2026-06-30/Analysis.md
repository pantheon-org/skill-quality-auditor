---
title: "Skill Audit — context-file"
type: audit
status: done
date: 2026-06-30
---

# Skill Audit — context-file

**Grade:** D (91/140)

## Dimension Scores

| Dimension | Score | Max |
|---|---|---|
| Knowledge Delta | 17 | 20 |
| Mindset + Procedures | 8 | 15 |
| Anti-Pattern Quality | 15 | 15 |
| Specification Compliance | 10 | 15 |
| Progressive Disclosure | 12 | 15 |
| Freedom Calibration | 6 | 15 |
| Pattern Recognition | 10 | 10 |
| Practical Usability | 13 | 15 |
| Eval Validation | 0 | 20 |

## Diagnostics

### Warnings

- **D2** no postcondition signals detected — add external verification steps (e.g. run tests, confirm output)
- **D2** no decision-point signals detected — add branching logic for error cases (e.g. ## Troubleshooting)
- **D4** harness-specific path found: .agents/
- **D4** absolute skill path outside code blocks: skills/context-index/regenerate-context-index
- **D4** .context/ or .agents/ reference outside code blocks: .context/
- **D9** evals/ directory missing entirely
- **D5** no references/ directory (progressive disclosure missing)
