# D1: Knowledge Delta (20 points)

**Purpose:** Ensure the skill contains expert-only knowledge, not information the model already knows.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 18–20 | Pure expert knowledge, <5% redundancy |
| 15–17 | Mostly expert, 5–15% redundancy |
| 12–14 | 15–30% redundancy (acceptable) |
| 9–11 | 30–50% redundancy (needs improvement) |
| 0–8 | >50% redundancy (failing) |

**Core principle:** Skill = Expert Knowledge − What AI Assistants Already Know

## Three Knowledge Types

1. **Expert (KEEP):**
  - Domain-specific patterns the model doesn't know by default
  - Project-specific conventions
  - Lessons from production experience
  - Tool gotchas and non-obvious behaviour
  - Decision frameworks (when to use X vs Y)
  - Anti-patterns with WHY they fail

1. **Activation (BRIEF REMINDERS OK):**
  - When to use this skill
  - Trigger keywords for pattern matching
  - Brief context setting (2–3 sentences)

1. **Redundant (DELETE):**
  - Basic syntax the model already knows
  - Installation instructions from official docs
  - API documentation copied verbatim
  - Generic best practices
  - Obvious examples

## Red Flags for Low Knowledge Delta

- Teaching basic syntax (`if/else`, `function`, `class`)
- Copying official documentation (schema definitions, rule lists)
- Explaining fundamentals (what is REST, what is a database)
- Generic advice (write tests, use version control)
- Installation tutorials (`npm install`, `pip install`)

## Examples

**Low Knowledge Delta (12/20):**

```markdown
# TypeScript Basics

## Variables
Use `let` for mutable, `const` for immutable:
let count = 0
const name = "Alice"
```

Problem: the model already knows basic TypeScript syntax.

**High Knowledge Delta (19/20):**

```markdown
# TypeScript: Making Illegal States Unrepresentable

## The Pattern
Use discriminated unions to eliminate impossible states:

❌ BAD: Multiple optional fields create 16 possible states
type Request = { loading?: boolean; error?: string; data?: User }

✅ GOOD: Tagged union with 3 valid states only
type Request =
  | { status: 'loading' }
  | { status: 'error'; error: string }
  | { status: 'success'; data: User }
```

Expert pattern the model does not know by default.

## Academic References

- [Li et al., 2025 — Instruction Agent: Enhancing Agent with Expert Demonstration](https://arxiv.org/abs/2509.07098)
- [Deng et al., 2024 — From Novice to Expert: LLM Agent Policy Optimization via Step-wise Reinforcement Learning](https://arxiv.org/abs/2411.03817)
- [Yin et al., 2025 — Grounding Open-Domain Knowledge from LLMs to Real-World RL Tasks: A Survey](https://www.ijcai.org/proceedings/2025/1198.pdf)
