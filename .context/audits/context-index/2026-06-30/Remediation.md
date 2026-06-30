# Remediation Plan — context-index

**Current Grade:** F (88/140)

## Priority Actions

### Eval Validation (0/20) — 20 pts available

⚠️ evals/ directory missing entirely

Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.

### Freedom Calibration (6/15) — 9 pts available

Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).

### Mindset + Procedures (10/15) — 5 pts available

⚠️ no decision-point signals detected — add branching logic for error cases (e.g. ## Troubleshooting)

Add a `## Mindset` or `## Philosophy` section. Use numbered procedure lists. Add `## When to Use` and `## When NOT to Use` sections.

### Specification Compliance (10/15) — 5 pts available

⚠️ harness-specific path found: .agents/

⚠️ absolute skill path outside code blocks: skills/context-index/check-context-frontmatter

⚠️ .context/ or .agents/ reference outside code blocks: .context/

Expand the `description` frontmatter to >100 characters. Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.

### Anti-Pattern Quality (11/15) — 4 pts available

Add NEVER statements paired with `WHY:` explanations. Include BAD/GOOD contrast examples.

### Knowledge Delta (17/20) — 3 pts available

Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern. Remove beginner-oriented patterns (npm install, getting started, hello world).

### Progressive Disclosure (12/15) — 3 pts available

⚠️ no references/ directory (progressive disclosure missing)

Add a `references/` directory with focused deep-dive `.md` files. Keep `SKILL.md` under 150 lines to maximise the score.

### Practical Usability (12/15) — 3 pts available

Add more fenced code blocks (aim for >5 pairs). Include `./` or `bun run` commands. Use language-tagged fences (```bash, ```typescript).

