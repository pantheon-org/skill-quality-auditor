# Remediation Plan — context-file

**Current Grade:** D (91/140)

## Priority Actions

### Eval Validation (0/20) — 20 pts available

⚠️ evals/ directory missing entirely

Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.

### Freedom Calibration (6/15) — 9 pts available

Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).

### Mindset + Procedures (8/15) — 7 pts available

⚠️ no postcondition signals detected — add external verification steps (e.g. run tests, confirm output)

⚠️ no decision-point signals detected — add branching logic for error cases (e.g. ## Troubleshooting)

Add a `## Mindset` or `## Philosophy` section. Use numbered procedure lists. Add `## When to Use` and `## When NOT to Use` sections.

### Specification Compliance (10/15) — 5 pts available

⚠️ harness-specific path found: .agents/

⚠️ absolute skill path outside code blocks: skills/context-index/regenerate-context-index

⚠️ .context/ or .agents/ reference outside code blocks: .context/

Expand the `description` frontmatter to >100 characters. Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.

### Knowledge Delta (17/20) — 3 pts available

Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern. Remove beginner-oriented patterns (npm install, getting started, hello world).

### Progressive Disclosure (12/15) — 3 pts available

⚠️ no references/ directory (progressive disclosure missing)

Add a `references/` directory with focused deep-dive `.md` files. Keep `SKILL.md` under 150 lines to maximise the score.

### Practical Usability (13/15) — 2 pts available

Add more fenced code blocks (aim for >5 pairs). Include `./` or `bun run` commands. Use language-tagged fences (```bash, ```typescript).

