---
name: skill-quality-auditor
description: "Evaluate, score, and remediate agent skill collections using a 9-dimension quality framework (Knowledge Delta, Mindset, Anti-Patterns, Specification Compliance, Progressive Disclosure, Freedom Calibration, Pattern Recognition, Practical Usability, Eval Validation). Performs duplication detection, generates remediation plans with T-shirt sizing, enforces CI quality gates, validates artifact conventions, tracks score trends, and ensures tessl registry compliance. Use when evaluating skill quality, auditing SKILL.md files, scoring agent skills, generating remediation plans, detecting duplicate skills, validating skill format, enforcing quality gates, optimizing for A-grade publication, comparing audit baselines, batch skill assessments, or checking tessl compliance. Triggers: 'check my skills', 'skill audit', 'improve my SKILL.md', 'quality check', 'A-grade scoring', 'quality gates', 'eval validation', 'audit all skills', 'remediation plan', 'skill judge', 'dimension scoring'."
---

# Skill Quality Auditor

Evaluate, maintain, and improve skill quality with 9-dimension framework scoring.

## Quick Start

```bash
# Single skill
skill-auditor evaluate <domain>/<skill-name> --json --store

# Batch with grade gate
skill-auditor batch <skill1> <skill2> --fail-below B --store
```

## When to Use

- Evaluate skills before merge or publication using 9-dimension scoring
- Generate remediation plans, detect duplication (High at 20%, Critical at 35%), or enforce CI quality gates
- Validate eval scenario coverage and artifact conventions

## When Not to Use

- Write the skill first — do not audit an unfinished draft
- Avoid using this as a substitute for peer review of logic or domain accuracy

## Prerequisites

- MUST have the skill directory with `SKILL.md` and at least one eval scenario.
- MUST run `--store` at least once before `remediate` or `trend` can produce output.

## Workflow

1. Run `skill-auditor evaluate <skill> --json --store`
2. Check artifacts and eval coverage using deterministic criteria
3. Generate a remediation plan with T-shirt sizing and score delta estimates
4. Re-run the auditor to verify improvement; if below target, focus on the lowest-scoring dimension

## Anti-Patterns

**NEVER** skip baseline comparison in recurring audits.
**WHY:** Score regressions go undetected without a prior stored audit.
**BAD:** `skill-auditor evaluate my-skill` with no prior `--store` run.
**GOOD:** `skill-auditor evaluate my-skill --store` after a prior stored audit exists.

**NEVER** ignore Knowledge Delta below 15/20.
**WHY:** Low D1 means the skill adds no value over LLM baseline knowledge.
**BAD:** Shipping a skill that restates generic framework documentation the LLM already knows.
**GOOD:** Ensuring every skill section contains constraints or thresholds absent from public docs.

**NEVER** apply subjective scoring.
**WHY:** Scores drift between evaluators and cannot be automated in CI pipelines.
**BAD:** Assigning D6 a score without checking hard:soft marker ratios.
**GOOD:** Running `skill-auditor evaluate` and using the numeric output as the canonical score.

**NEVER** create kitchen-sink skills covering unrelated tasks.
**WHY:** Broad scope kills D7 pattern recognition and prevents correct agent triggering.

**NEVER** use harness-specific paths in skill content.
**WHY:** Absolute paths break portability; use agent-neutral relative paths instead.

**NEVER** list references without "When to Use" conditions.
**WHY:** Unconditional loading bloats context and penalises D5 progressive disclosure.

See [Detailed Anti-Patterns](references/detailed-anti-patterns.md) for full failure modes, agent name references, and D4 heading rules.

## Examples

Remediation workflow:

```bash
./dist/skill-auditor evaluate cmd/assets --json --store
# Score increases after each remediation cycle
```

PR-scoped triage:

```bash
# Extract changed skills from the PR diff and batch-audit them
./dist/skill-auditor batch cmd/assets testdata/fixtures/skill-full --fail-below B --store
```

Audit all skills:

```bash
./dist/skill-auditor batch $(find skills -name "SKILL.md" | sed 's|skills/||;s|/SKILL.md||' | tr '\n' ' ')
```

## Troubleshooting

- A non-zero exit from `batch` means a skill scored below `--fail-below`; `duplication` exits 2 when a Critical (>=35%) pair is found.
- If a command exits below threshold, consider running `evaluate --store` to capture diagnostics; see [Scripts Workflow](references/scripts-audit-workflow.md) for per-command failure modes.

## Self-Audit

```bash
./dist/skill-auditor evaluate cmd/assets --json --store
# Expected: A grade (>= 126/140)
```

```bash
./dist/skill-auditor evaluate cmd/assets --json | jq '.grade'
# Confirms the grade: "A"

## References

### Framework

| Topic | Reference | When to Use |
| --- | --- | --- |
| Per-dimension criteria and bonus rules | [Dimensions](references/framework-dimensions.md) | Evaluating any dimension or understanding the rubric; skip if you only need the final grade |
| Score thresholds and grade bands | [Scoring Rubric](references/framework-scoring-rubric.md) | Calculating a total score or assigning a grade |
| A-grade checklist and red flags | [Quality Standards](references/framework-quality-standards.md) | Targeting A-grade or reviewing blockers |
| Trigger pattern density and keyword analysis | [Pattern Recognition](references/advanced-pattern-recognition.md) | Scoring D7 or improving description keywords |
| Canonical SKILL.md structure and References table standard | [SKILL Template](references/skill-template.md) | Authoring or refactoring a skill |

### Operations

| Topic | Reference | When to Use |
| --- | --- | --- |
| CI gate configuration and batch pass/fail logic | [Quality Thresholds](references/quality-thresholds-scoring.md) | Setting up CI quality gates |
| NEVER/WHY/BAD/GOOD failure modes per dimension | [Anti-Patterns](references/detailed-anti-patterns.md) | Explaining low scores or writing remediation guidance |
| T-shirt sizing and remediation roadmaps | [Remediation Planning](references/remediation-planning.md) | Writing a remediation plan for a C/D-grade skill |
| Deduplication workflow and aggregation guidance | [Duplication Detection](references/duplication-detection-algorithm.md) | Detecting skill overlap or planning aggregations |
| `skill-auditor evaluate/batch` usage and output formats | [Scripts Workflow](references/scripts-audit-workflow.md) | Running audits from the command line |
| Registry publication gates and tessl compliance checks | [Tessl Compliance](references/tessl-compliance-framework.md) | Preparing a skill for public registry submission |
