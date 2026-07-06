---
title: "Remediation Plan — plan-review"
type: AUDIT
status: DONE
date: 2026-07-03
---

# Remediation Plan — plan-review

**Current Grade:** B+ (124/140)

## Priority Actions

### Freedom Calibration (10/15) — 5 pts available

Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).

### Mindset + Procedures (12/15) — 3 pts available

Add a `## Mindset` or `## Philosophy` section. Use numbered procedure lists. Add `## When to Use` and `## When NOT to Use` sections.

### Progressive Disclosure (12/15) — 3 pts available

Add a `references/` directory with focused deep-dive `.md` files. Keep `SKILL.md` under 150 lines to maximise the score.

### Specification Compliance (13/15) — 2 pts available

⚠️ agent-specific reference found: claude code

⚠️ .context/ or .agents/ reference outside code blocks: .context/

Expand the `description` frontmatter to >100 characters. Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.

### Eval Validation (18/20) — 2 pts available

⚠️ adversarial bonus: +1 pts — not applied to score

⚠️ independent authoring bonus: +1 pts — not applied to score

Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.

### Knowledge Delta (19/20) — 1 pt available

Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern. Remove beginner-oriented patterns (npm install, getting started, hello world).

