# D6: Freedom Calibration (15 points)

**Purpose:** Balance prescription (rigid rules) vs flexibility (guidelines) for the skill type.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 13–15 | Appropriate calibration for skill type |
| 10–12 | Slightly too rigid or loose |
| 7–9 | Mismatched calibration |
| 0–6 | Completely wrong |

## Calibration Levels

### Rigid (Mindset skills)

Strong rules, must follow.

- Example: proof-of-work — "NEVER trust agent reports without verification"
- Use for: critical foundations, security, correctness

### Balanced (Process skills)

Clear steps with contextual flexibility.

- Example: TDD — "Red → Green → Refactor (adapt to context)"
- Use for: workflows, methodologies

### Flexible (Tool skills)

Options and trade-offs presented.

- Example: typescript-type-system — "Choose based on use case"
- Use for: technical tools, patterns

## Examples

**Well-Calibrated (14/15):**

```markdown
# Proof of Work (Mindset skill)

## Zero-Tolerance Rules
NEVER trust agent completion reports without verification.
ALWAYS show command output as proof.
ZERO exceptions to verification protocol.
```

Appropriately rigid for a critical verification mindset skill.

**Miscalibrated (7/15):**

```markdown
# TypeScript Basics (Tool skill)

## Rules
ALWAYS use const for all variables.
NEVER use let or var under any circumstances.
```

Too rigid — `let` has valid use cases in a tool skill.

## Academic References

```bibtex
@article{zhang2025reasoning,
  title         = {Reasoning over Boundaries: Enhancing Specification Alignment via Test-time Deliberation},
  author        = {Zhang and others},
  year          = {2025},
  journal       = {arXiv preprint arXiv:2509.14760},
  eprint        = {2509.14760},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2509.14760}>
}

```

```bibtex
@article{sorensen2026specification,
  title         = {Specification as the New Management},
  author        = {Sorensen},
  year          = {2026},
  journal       = {ResearchGate},
  url           = {<https://www.researchgate.net/publication/401626622}>
}

```

```bibtex
@article{tao2025orchestration,
  title         = {LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Multi-Model Collaboration},
  author        = {R. Tao},
  year          = {2025},
  journal       = {Research Square},
  url           = {<https://www.researchsquare.com/article/rs-9323974/latest}>
}
```
