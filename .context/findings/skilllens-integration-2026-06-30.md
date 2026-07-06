---
title: "Finding: SkillLens Integration Assessment"
type: FINDING
status: ACTIVE
date: 2026-06-30
value: LOW
themes:
  - EVAL
---

# Finding: SkillLens Integration Assessment

Date: 2026-06-30
Status: DECISION-SUPPORT, not actioned

> SkillLens (Microsoft, MIT) is a complementary research framework for studying model-generated agent skills; not a drop-in for this project, but its empirical methodology and meta-skill findings could inform scoring rubric improvements and a future D10 scorer.

## Summary

[SkillLens](https://github.com/microsoft/SkillLens) (arXiv:2605.23899) is a Python research framework from Microsoft that studies the full lifecycle of model-generated agent skills: **experience generation → skill extraction → skill consumption**. It runs target models on five benchmarks (SWE-bench, ALFWorld, SpreadsheetBench, BFCL v4, SEAL-0), extracts "modes of behavior" from trajectories via sequential or parallel methods, then re-injects those skills to measure benchmark performance deltas. Key finding: skills help on average but exhibit non-trivial negative transfer, and extractor strength does not predict consumer strength.

## Detail

### What SkillLens does

| Stage | Subcommand | Output |
|-------|-----------|--------|
| 1. Raw experience generation | `skilllens infer` | Raw agent trajectories |
| 2. Schema normalization | `skilllens convert` | Unified `Trajectory` JSON |
| 3. Skill extraction | `skilllens extract` | `skill_set.json` (sequential or parallel) |
| 4. Skill consumption | `skilllens infer --skill-set` | Benchmark results with/without skills |

Metrics: *Extraction Efficacy* (does the skill capture useful behavior?) and *Target Evolvability* (does the target model improve when consuming it?).

### Key research findings

- Model-generated skills are beneficial on average but produce non-trivial negative transfer
- Extractor and consumer capabilities are independent — a model can be a strong extractor but weak consumer, and vice versa
- Skill utility is independent of model scale or baseline task strength
- The paper distills these into a "meta-skill" that guides extraction toward features tied to actual utility, consistently reducing negative transfer

### Comparison to skill-quality-auditor

| Dimension | SkillLens | skill-quality-auditor |
|-----------|-----------|----------------------|
| Evaluation method | Empirical (benchmark pass rates) | Analytic (rubric scoring) |
| Skill source | LLM-generated from trajectories | Human-written SKILL.md |
| Evaluation scope | Did the skill improve the model? | Is the skill document well-structured? |
| Tech stack | Python, LLM APIs, per-benchmark sandboxes | Go CLI, static analysis |
| Maturity | Early (10 commits, no releases) | Mature (full CLI, tests, CI) |

## Internal Architecture

SkillLens has 5 layers:

### 1. Data Model Layer (`skilllens/schema/`)

| Type | Purpose | Key fields |
|------|---------|------------|
| `Trajectory` | Unified agent run representation | `id`, `steps[]`, `reward`, `outcome`, `benchmark` |
| `Step` | Single action within trajectory | `role`, `content`, `tool_calls`, `observation` |
| `Mode` | Success/failure pattern | `type` (success/failure), `pattern`, `description`, `evidence` |
| `ModeSet` | Collection of modes from one map/merge | `success_modes[]`, `failure_modes[]`, `summary` |
| `Skill` | Agent Skills standard format | `name`, `description`, `body`, `references[]`, `scripts[]` |
| `SkillSet` | Extraction output bundle | `skills[]`, `extractor_model`, `extraction_config` |

### 2. Extraction Layer (`skilllens/extraction/`)

Two methods implementing `ExtractionMethod(ABC)`:

- **SequentialExtraction**: ReAct loop. For each trajectory, LLM gets a fresh conversation but the SkillStore persists across trajectories. LLM uses 7 tools (list_skills, view_skill, read_skill_file, add_skill, update_skill, delete_skill, finish_extraction) to manage skills. One LLM call per trajectory with up to `max_tool_rounds` tool call rounds.

- **ParallelExtraction** (Map-Reduce, the paper's primary method):
  1. **Map phase** (parallel): Each trajectory → independent LLM call → `ModeSet` JSON. Resolved → success modes, unresolved → failure modes. Configurable `max_modes_per_trajectory`, `max_concurrency`.
  2. **Intermediate reduce** (hierarchical): Groups of ModeSets → LLM merge → unified ModeSet. Repeats until ≤ `merge_group_size` sets remain.
  3. **Final reduce**: Remaining ModeSets → LLM with tool-calling → SkillStore operations (add/update/delete skills).

### 3. SkillStore (`skilllens/extraction/skill_store.py`)

In-memory skill manager exposed as OpenAI function-calling tools. Enforces:
- `max_skills` (count limit)
- `max_skill_chars` per skill (description + body + references + scripts)
- `max_total_chars` across all skills
- Progressive disclosure: `view_skill` returns body + file names but NOT file contents; `read_skill_file` returns specific file content on demand
- Slug-based naming with collision handling
- Full operation history for audit/logging

### 4. Prompt Layer (`skilllens/prompts/`)

Carefully engineered prompts for each phase:
- **Map prompt**: Instructs LLM to extract high-level, transferable patterns; avoid task-specific details; generalize aggressively. Separate prompts for success vs failure mode extraction. Includes `meta_skill_guidance` injection point for extracted research findings.
- **Intermediate reduce prompt**: Instructs LLM to deduplicate, merge, and generalize. "Raise the abstraction level to cover more scenarios."
- **Final reduce prompt**: Instructs LLM to integrate success + failure modes into cohesive skills with recommended approaches, pitfalls to avoid, and decision criteria.
- **Sequential prompt**: Full ReAct system prompt with 7 tool descriptions and skill quality requirements.

### 5. Client Layer (`skilllens/client/openai_client.py`)

Unified LLM client supporting Azure OpenAI, OpenAI, vLLM, and Gemini. Uses Responses API (`/v1/responses`) primarily with Chat Completions fallback. Features:
- Exponential-backoff retry (5 attempts, 2-120s wait)
- Reasoning model detection (omits temperature for o1/o3/gpt-5)
- `chat_with_tools()` — ReAct loop driver: sends messages, processes function_call output items, calls tool_handler, appends results, repeats until finish_tool or max_rounds

## Portability Assessment

Six components ranked by effort and value for porting into skill-quality-auditor:

### 1. Mode/ModeSet data structures (Low effort, High value)

Port `skilllens/schema/modes.py` to Go — success/failure pattern types with evidence, source trajectory IDs, and summaries.

```go
// New file: analysis/modes.go
type ModeType string
const (
    SuccessMode ModeType = "success"
    FailureMode ModeType = "failure"
)

type Mode struct {
    Type                 ModeType
    Pattern              string   // e.g. "incremental-validation"
    Description          string   // 2-4 sentence description
    Evidence             string   // Brief evidence from trajectory
    SourceTrajectoryIDs  []string
}

type ModeSet struct {
    SuccessModes         []Mode
    FailureModes         []Mode
    SourceTrajectoryIDs  []string
    Summary              string
}
```

*Integration points*: Use `ModeSet` as output format for the duplication analysis pipeline. Use failure `Mode` categories to classify duplication risks. Store in `.context/analysis/` reports.

### 2. Failure mode categories → D3 anti-pattern scorer (Low effort, High value)

SkillLens's map prompts define a taxonomy of failure modes:
- **Error patterns**: Categories of mistakes (data format assumptions, wrong operation ordering)
- **Anti-patterns**: Approaches to avoid (in-place mutation without backup, ignoring edge cases)
- **Pitfalls and traps**: Non-obvious failure causes (syntax differences between interfaces, silent type coercion)

*Integration*: Add these categories as detection patterns in `scorer/d3_anti_pattern_coverage.go`. The D3 scorer currently checks for anti-pattern coverage in SKILL.md — these categories from SkillLens are empirically grounded in actual agent failures across 5 benchmarks.

### 3. Meta-skill guidance → rubric docs (Low effort, Medium value)

The paper's "meta-skill" distills properties correlated with actual skill utility:
- **Generality**: Applies to broad class of tasks
- **Information density**: Concrete, actionable, non-obvious guidance
- **Self-containedness**: Followable without trajectory context
- **Balanced coverage**: Integrates both success strategies AND failure modes

*Integration*: Add these properties to `cmd/assets/references/framework-dimensions.md` as grounding for D1 (Knowledge Delta), D2 (Mindset & Procedures), and D3 (Anti-Pattern Coverage) criteria.

### 4. Parallel extraction → new `skillmine` command (High effort, High value)

Port the Map-Reduce algorithm as `cmd/skillmine.go`:
1. Takes agent trajectories (JSON/JSONL) as input
2. Implements parallel map phase (goroutines + LLM API calls per trajectory)
3. Implements hierarchical reduce (iterative merge groups)
4. Implements SkillStore-style final synthesis
5. Outputs SKILL.md files compatible with `skill-auditor evaluate`

*Architecture*: New `skillmine/` package parallel to `scorer/`. LLM client abstraction (`internal/llmclient/`). Shares `analysis/modes.go` types. Generated skills can be immediately piped through the existing scorer for quality validation.

*Dependencies*: Needs an LLM API client for Go. Could use openai-go or a thin HTTP wrapper. The SkillStore's 7-tool ReAct pattern is the most complex part to port.

### 5. D10 Empirical Validator scorer (Medium effort, High value)

New dimension: `scorer/d10_empirical.go`. Runs a skill through SkillLens-style empirical validation:
1. Select relevant benchmark (matching the skill's domain)
2. Run target model on benchmark WITHOUT the skill (baseline)
3. Run same model WITH skill injected into system prompt
4. Score = performance delta normalized by baseline

*Score mapping*:
- ≥+10% improvement → 20/20
- +5-10% → 15/20
- +1-5% → 10/20
- ±1% → 5/20 (no effect)
- Negative → 0/20 (harmful)

*Integration*: Lightweight benchmark harness in `testdata/benchmarks/`. Uses existing trajectory fixtures. Does NOT need the full per-benchmark sandbox infra — token-level pass/fail on fixed test pools is sufficient for relative scoring.

### 6. SkillStore tool pattern → remediation engine UX (Medium effort, Medium value)

The SkillStore's progressive-disclosure pattern can improve `cmd/remediate.go`:
- `list_skills` → `view_skill` → `read_skill_file` pattern maps to `--list`, `--view`, `--read` flags
- Character budget enforcement can be reused for remediation plan size limits
- The dispatch/handler pattern can structure the remediation CLI's interactive mode

## Recommended Action

1. **Do now (low effort)**: Port Mode/ModeSet data structures to `analysis/modes.go`. Extract failure mode categories into D3 scorer. Add meta-skill findings to `cmd/assets/references/`.

2. **This quarter (medium effort)**: Build D10 empirical validator with lightweight benchmark harness using existing `testdata/` fixtures.

3. **Future (high effort)**: Port the full parallel extraction pipeline as `cmd/skillmine` once an LLM Go client is established and benchmark infra is available.

4. **Monitor**: Watch SkillLens repo for maturation — star count, releases, expanded benchmarks. Reassess when it reaches ≥500 stars or a tagged release.
