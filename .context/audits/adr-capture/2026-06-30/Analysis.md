---
title: "Skill Audit — adr-capture"
type: audit
status: done
date: 2026-06-30
---

# Skill Audit — adr-capture

**Grade:** F (85/140)

## Dimension Scores

| Dimension | Score | Max |
|---|---|---|
| Knowledge Delta | 17 | 20 |
| Mindset + Procedures | 3 | 15 |
| Anti-Pattern Quality | 15 | 15 |
| Specification Compliance | 10 | 15 |
| Progressive Disclosure | 10 | 15 |
| Freedom Calibration | 5 | 15 |
| Pattern Recognition | 10 | 10 |
| Practical Usability | 15 | 15 |
| Eval Validation | 0 | 20 |

## Diagnostics

### Warnings

- **D2** no precondition signals detected — add explicit entry conditions (e.g. ## Prerequisites)
- **D2** no postcondition signals detected — add external verification steps (e.g. run tests, confirm output)
- **D2** no decision-point signals detected — add branching logic for error cases (e.g. ## Troubleshooting)
- **D4** harness-specific path found: .agents/
- **D4** ../ reference outside code blocks (self-containment violation)
- **D4** .context/ or .agents/ reference outside code blocks: .context/
- **D7** description lacks negative and workflow anchors — skill may over-trigger on adjacent topics
- **D9** evals/ directory missing entirely
- **D5** no references/ directory (progressive disclosure missing)
