# D4: Specification Compliance (15 points)

**Purpose:** Ensure proper frontmatter, single-task focus, activation keywords, and cross-harness portability.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 13–15 | Perfect spec compliance |
| 10–12 | Minor issues |
| 7–9 | Missing key elements |
| 0–6 | Non-compliant |

## Components

### 1. Task Focus Declaration (4 points) — CRITICAL

- Skill indicates ONE type of task it helps complete
- Description clearly scopes to a single purpose
- No ambiguity about what the skill does
- Example: "Write BDD tests" (good) vs "Testing and development" (bad — two tasks)

### 2. Description Field Quality (6 points)

- **Primary agents:** Exactly 3 words
- **Other agents:** Comprehensive with trigger examples
- Must include activation keywords
- Determines whether the skill activates at all

### 3. Cross-Harness Portability (3 points) — CRITICAL

- **No harness-specific paths (1 point):** Avoid `.opencode/`, `.claude/`, `.cursor/`, `.aider/`, `.continue/`
- **No agent-specific references (1 point):** Do not mention "Claude Code", "Cursor Agent", "GitHub Copilot", etc. in instructions
- **Relative path usage (1 point):** Reference files relative to the skill directory (`scripts/`, `references/`, `templates/`)
- **WHY:** Skills must work across 40+ agentic harnesses without modification
- **IMPACT:** Harness-specific paths break skill discovery when synced to other agents

### 4. Self-Containment (penalties: up to −12 points) — CRITICAL

*SKILL.md penalties (checked outside fenced code blocks):*

- **No parent-escaping paths (−2 points):** Must not use `../` references outside code fences
- **No absolute repo paths (−1 point):** Must not reference `skills/X/Y/Z` or other hardcoded repo paths outside code fences
- **No repo-root directory references (−1 point):** Must not reference `.context/`, `.agents/`, or other repo-root directories outside code fences

*scripts/ penalties (−1 per file with violation, cap −2 per category):*

- No absolute repo paths in scripts (capped at −2 total)
- No repo-root directory references in scripts (capped at −2 total)

*references/ penalties (−1 per file with violation, cap −2 per category):*

- No absolute repo paths in references (capped at −2 total)
- No repo-root directory references in references (capped at −2 total)

**WHY:** Skills must be fully self-contained. When installed via `tessl install`, they land in arbitrary directories — any reference to files outside the skill's own directory tree will break.

### 5. Script Language Portability (bonus: +1 point)

- Skills with `scripts/` containing Python (`.py`), TypeScript (`.ts`), or JavaScript (`.js`) files earn a portability bonus
- Shell scripts (`.sh`) remain the accepted default and receive no penalty
- Accepted shebangs: `#!/usr/bin/env python3`, `#!/usr/bin/env bun`, `#!/usr/bin/env node`

### 6. Proper Frontmatter (1 point)

- `name` and `description` fields present
- Correct YAML syntax

### 7. Activation Keywords (1 point)

- Domain terms that trigger the skill
- Example: "BDD, Gherkin, Given-When-Then, Cucumber"

### 8. References Section Format (bonus: +1 point)

- Heading is exactly `## References`
- Last H2 in SKILL.md
- Content is a Markdown table with `Topic \| Reference \| When to Use` columns
- Every `Reference` cell is a markdown link
- See [D5](d5-progressive-disclosure.md) for the References Section Standard

## Examples

**Excellent Specification Compliance (15/15):**

```yaml
---
name: bdd-testing
description: Behavior-Driven Development with Given-When-Then scenarios, Cucumber.js,
  Three Amigos collaboration, Example Mapping, living documentation, and acceptance
  criteria. Use when writing BDD tests, feature files, or planning discovery workshops.
---
```

*Comprehensive description, portable paths, no agent-specific mentions.*

**Poor Specification Compliance (7/15):**

```yaml
---
name: bdd-testing
description: BDD testing patterns
---
```

Instructions reference `.opencode/scripts/run-tests.sh` and `.claude/docs/file.md`.

*Problems: weak description, harness-specific paths, agent-specific references.*

## Academic References

- [T Rehan, 2026 — Test-Driven AI Agent Definition (TDAD): Compiling Tool-Using Agents from Behavioral Specifications](https://arxiv.org/abs/2603.08806)
- [C Paduraru, M Zavelca, A Stefanescu — Agentic AI for Behaviour-Driven Development Testing Using Large Language Models](https://www.researchgate.net/publication/390835646)
- [R Tao, 2025 — LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Multi-Model Collaboration](https://www.researchsquare.com/article/rs-9323974/latest)
- [J Kohl, O Kruse, Y Mostafa, A Luckow et al. — Automated Structural Testing of LLM-Based Agents](https://ieeexplore.ieee.org/abstract/document/11401679/)
