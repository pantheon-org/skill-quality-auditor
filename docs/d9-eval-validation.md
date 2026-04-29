# D9: Eval Validation (20 points)

**Purpose:** Verify the skill has been validated at runtime through tessl eval scenarios, proving agents actually follow its instructions.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 17–20 | Complete evals with ≥80% instruction coverage, ≥3 valid scenarios |
| 13–16 | Evals present with partial coverage or incomplete scenarios |
| 7–12 | Evals directory exists but missing key files |
| 1–6 | Minimal eval structure, no coverage data |
| 0 | No evals directory |

**Core principle:** Static quality (D1–D8) is necessary but not sufficient. Runtime validation proves the skill actually changes agent behaviour.

## Components

### 1. Eval Directory Structure (4 points)

- `evals/` directory exists with proper layout
- Follows tessl eval harness conventions

### 2. Instruction Inventory (3 points)

- `instructions.json` present and non-empty
- Every instruction extracted from `SKILL.md`
- Each instruction classified by `why_given`: `reminder`, `new knowledge`, `preference`

### 3. Coverage Statistics (6 points)

- `summary.json` with `instructions_coverage` data (3 points)
- Coverage percentage ≥ 80% (3 points)

### 4. Valid Scenarios (4 points)

- ≥ 3 scenarios with complete structure (`task.md` + `criteria.json` + `capability.txt`)
- Each `criteria.json` sums to exactly 100

### 5. Criteria Quality (3 points)

- 10+ checklist items per scenario
- Binary yes/no criteria traceable to specific instructions
- No instruction leakage in `task.md`

## Relationship to D1 and D3

When `instructions.json` exists, its data enriches other dimensions:

- **D1 (Knowledge Delta):** The `why_given` distribution (`new knowledge` + `preference` vs `reminder`) provides a more accurate expert content ratio than heuristics alone.
- **D3 (Anti-Pattern Quality):** Instructions containing NEVER/ALWAYS/anti-pattern keywords are cross-referenced with scenario coverage for a stronger signal.

## Creating Evals

Use the `creating-eval-scenarios` skill to generate evaluation scenarios:

```bash
# Ensure the skill is packaged as a tessl tile first
tessl eval run <tile-path>
tessl eval view-status <status_id> --json
```

## Examples

**High Eval Validation (19/20):**

```text
skill-name/evals/
  instructions.json      # 28 instructions extracted
  summary.json           # 100% coverage, 5 scenarios
  summary_infeasible.json
  scenario-0/            # task.md + criteria.json (sum=100) + capability.txt
  scenario-1/
  scenario-2/
  scenario-3/
  scenario-4/
```

**Low Eval Validation (4/20):**

```text
skill-name/evals/
  instructions.json      # present but only 5 instructions
  # no summary.json, no scenarios
```

**Zero Eval Validation (0/20):**

```text
skill-name/
  SKILL.md               # no evals/ directory at all
```

## Academic References

- [Rehan, 2026 — Test-Driven AI Agent Definition (TDAD): Compiling Tool-Using Agents from Behavioral Specifications](https://arxiv.org/abs/2603.08806)
- [Alami, 2026 — Cognitive Camouflage: Specification Gaming in LLM-Generated Code Evades Holistic Evaluation but Not Adversarial Execution](https://papers.ssrn.com/sol3/papers.cfm?abstract_id=6512960)
- [Wang, Chen, Deng, Lin, Harman et al. — A Comprehensive Study on Large Language Models for Mutation Testing](https://dl.acm.org/doi/abs/10.1145/3805038)
- [Pan, Hu, Xia, Yang — Re-Evaluating Code LLM Benchmarks Under Semantic Mutation](https://arxiv.org/abs/2506.17369)
- [Bouafif, Hamdaqa, Zulkoski — PrimG: Efficient LLM-Driven Test Generation Using Mutant Prioritization](https://dl.acm.org/doi/abs/10.1145/3756681.3756991)
