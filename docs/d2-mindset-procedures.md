# D2: Mindset & Procedures (15 points)

**Purpose:** Provide philosophical framing and step-by-step workflows.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 13–15 | Clear mindset + detailed procedures + when/when-not guidance |
| 10–12 | Has most elements, minor gaps |
| 7–9 | Missing a key element |
| 0–6 | Generic or absent |

## Components

1. **Clear Mindset/Philosophy (5 points)**
  - Core principle or philosophy
  - Why this approach over alternatives
  - Example: "Trust but verify" (proof-of-work), "Composition over inheritance" (structural-design)

1. **Step-by-Step Procedures (5 points)**
  - Numbered workflow
  - Clear entry/exit points
  - Validation steps
  - Example: TDD cycle (Red → Green → Refactor)

1. **When/When-Not Guidance (5 points)**
  - Clear activation criteria
  - Explicit non-applicable scenarios
  - Example: "Use for backend APIs, NOT for UI styling"

## Example

**Strong Mindset + Procedures (15/15):**

```markdown
# Test-Driven Development

## Mindset
Write tests BEFORE implementation. The test defines the contract; implementation fulfils it.

## Workflow
1. Red: Write failing test (verify it fails)
2. Green: Minimum code to pass
3. Refactor: Improve without breaking tests

## When to Apply
✅ New functions, features, bug fixes (reproduce first)
❌ UI styling, configuration, documentation

## When NOT to Apply
- Throwaway prototypes
- Generated code
- Trivial getters/setters
```

## Sub-scorer Breakdown

The D2 scorer runs three sub-scorers independently:

- **Preconditions** — "Before you start" or "Requirements" sections that set up entry criteria
- **Postconditions** — "After completion" or "Definition of Done" sections that define success
- **Decision points** — conditional guidance ("if X then Y, otherwise Z") embedded in the workflow

All three should be present for a full score.

## Academic References

- [Bakal, 2026 — Knowledge Activation: AI Skills as the Institutional Knowledge Primitive for Agentic Software Development](https://arxiv.org/abs/2603.14805)
- [Carriero, Scrocca et al. — Procedural Knowledge Ontology (PKO)](https://link.springer.com/chapter/10.1007/978-3-031-94578-6_19)
- [Bi, Hu, Nasir, 2025 — Real-Time Procedural Learning From Experience for AI Agents](https://arxiv.org/abs/2511.22074)
- [Bi, Wu, Hao et al., 2026 — Automating Skill Acquisition through Large-Scale Mining of Agentic Repositories](https://arxiv.org/abs/2603.11808)
