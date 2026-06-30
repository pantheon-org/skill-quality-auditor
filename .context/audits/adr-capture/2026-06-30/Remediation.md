# Remediation Plan — adr-capture

**Current Grade:** F (85/140)

## Priority Actions

### Eval Validation (0/20) — 20 pts available

⚠️ evals/ directory missing entirely

Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.

### Mindset + Procedures (3/15) — 12 pts available

⚠️ no precondition signals detected — add explicit entry conditions (e.g. ## Prerequisites)

⚠️ no postcondition signals detected — add external verification steps (e.g. run tests, confirm output)

⚠️ no decision-point signals detected — add branching logic for error cases (e.g. ## Troubleshooting)

Add a `## Mindset` or `## Philosophy` section. Use numbered procedure lists. Add `## When to Use` and `## When NOT to Use` sections.

### Freedom Calibration (5/15) — 10 pts available

Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).

### Specification Compliance (10/15) — 5 pts available

⚠️ harness-specific path found: .agents/

⚠️ ../ reference outside code blocks (self-containment violation)

⚠️ .context/ or .agents/ reference outside code blocks: .context/

Expand the `description` frontmatter to >100 characters. Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.

### Progressive Disclosure (10/15) — 5 pts available

⚠️ no references/ directory (progressive disclosure missing)

Add a `references/` directory with focused deep-dive `.md` files. Keep `SKILL.md` under 150 lines to maximise the score.

### Knowledge Delta (17/20) — 3 pts available

Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern. Remove beginner-oriented patterns (npm install, getting started, hello world).

