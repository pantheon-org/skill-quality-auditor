---
title: "Remediation Plan — design-debate"
type: audit
status: done
date: 2026-07-06
---

# Remediation Plan — design-debate

**Current Grade:** C (102/140)

## Priority Actions

### Practical Usability (5/15) — 10 pts available

Add more fenced code blocks (aim for >5 pairs). Include `./` or `bun run` commands. Use language-tagged fences (```bash, ```typescript).

### Freedom Calibration (6/15) — 9 pts available

Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).

### Mindset + Procedures (10/15) — 5 pts available

⚠️ no postcondition signals detected — add external verification steps (e.g. run tests, confirm output)

Add a `## Mindset` or `## Philosophy` section. Use numbered procedure lists. Add `## When to Use` and `## When NOT to Use` sections.

### Progressive Disclosure (10/15) — 5 pts available

Add a `references/` directory with focused deep-dive `.md` files. Keep `SKILL.md` under 150 lines to maximise the score.

### Specification Compliance (11/15) — 4 pts available

⚠️ agent-specific reference found: claude code

⚠️ .context/ or .agents/ reference outside code blocks: .context/

Expand the `description` frontmatter to >100 characters. Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.

### Knowledge Delta (17/20) — 3 pts available

Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern. Remove beginner-oriented patterns (npm install, getting started, hello world).

### Eval Validation (18/20) — 2 pts available

⚠️ adversarial bonus: +1 pts — not applied to score

⚠️ independent authoring bonus: +1 pts — not applied to score

Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.

