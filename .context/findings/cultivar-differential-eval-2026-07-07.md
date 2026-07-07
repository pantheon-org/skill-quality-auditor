---
title: "Finding: cultivar Differential Behavioural Eval"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: MEDIUM
themes:
  - EVAL
  - SKILL-QUALITY
related:
  - ./skillvitals-trigger-efficacy-2026-07-07.md
---

# Finding: cultivar Differential Behavioural Eval

## What is cultivar

[cultivar](https://github.com/pinecone-io/cultivar) is an MIT-licensed Python 3.11+ eval
framework (v0.2.0) from the Pinecone DevRel team. Its stated purpose: "Eval framework for
testing whether agent skills improve behavior across coding CLIs." You write natural-language
task specs, cultivar runs a coding agent against each task with and without the skill loaded,
an LLM grader scores the resulting run against your criteria, and you read the delta to decide
if the skill helps. It supports multiple agent CLIs (Claude, Copilot, Gemini), local or remote
isolated sandboxes (Modal), parallelism, and repeat runs.

## Current state of this project

**skill-quality-auditor** scores skills against a 9-dimension static-document framework plus
an LLM-as-judge native eval runner (`cmd/eval.go`) that runs authored scenarios **with the
skill loaded**. D9 validates that evals exist and pass with the skill; it never runs a
without-skill baseline, so it cannot separate "the skill is genuinely additive" from "the base
model already handled it". The judge machinery exists; what is missing is the *differential*
and *empirical* framing.

## Core methodology, differential A/B testing

For each task the same agent runs up to three variants:

- **with-skill:** skill mounted, prompt prefixed `Use the /<skill> skill.`
- **without-skill:** identical agent/model/task, skill absent (the baseline).
- **with-docs (optional):** no skill, but the task's reference docs are prepended raw to the
  prompt.

Two deltas are read: with-skill vs without-skill ("does the skill do anything?") and
**with-skill vs with-docs ("is my distilled skill better than dumping the raw docs into the
prompt?")**. Grading is an LLM judge (Claude Haiku by default) run locally so the API key
never enters the sandbox. Exact-match/regex grading is explicitly rejected because skill
criteria are qualitative.

## Concrete mechanisms worth adopting

### Task-case schema (`docs/task-yaml.md`, `evals/init.py`)

Per task: `id`, `intent` (the prompt), `category`; shell hooks `setup` / `teardown` /
`verify` where **`verify` stdout is fed to the grader** as state-based ground truth (for
example `pc index stats my-test-index`); `env` (preflight-checked); and a `ground_truth`
object with `criteria` (PASS/FAIL prose, the primary grader input), `commands`, `flexible`
(acceptable-variation notes), `outcome`, and `context_refs`.

### Grader contract (`evals/framework/grader.py`)

- Prompt assembled in fixed order: Skill Reference, Criteria, expected commands/flexible/
  outcome, Reference Material (cap 100 KB), Calibration examples, Agent conversation
  (truncated 50 KB), Verification output, Workdir files (extension-allowlisted, cap 40 KB),
  Instructions.
- Returns JSON only: `{pass, proposed_command, evidence, reasoning, suggestions}`. On
  fail/partial it populates `suggestions: [{cause, fix}]` (1 to 3 items), so the grader
  doubles as a remediation generator.

### Metrics (`evals/framework/reporting.py`)

Per variant: pass rate (`passed/total (%)`), avg duration, avg tokens (input incl.
cache-read + cache-creation, and output), avg cost USD, turn count. Results persisted as
`results/<timestamp>/` with per-run `.json`, `.md`, `.jsonl`, `.workdir/`, and `grades.json`.

## Novel ideas a naive evaluator would miss

1. **It grades behaviour change, not document quality.** A well-written skill that changes
   nothing scores as useless.
2. **The `with-docs` control is the sharpest idea.** Declaring `context_refs` auto-activates
   a third variant that dumps the raw source docs into the prompt. If the skill does not beat
   raw docs on pass rate or cost, the distillation added no value. This is a
   distillation-value test static scoring cannot see.
3. **Anti-contamination evidence rules.** The grader is forbidden from quoting the Skill
   Reference, Reference Material, or Calibration Examples as evidence; evidence must come from
   the actual run (conversation / verification output / generated files). Stops the judge
   marking PASS by reading the skill's own claims.
4. **Autofail guards that bypass the judge.** If the trace has no agent signal or a code-gen
   task produced an empty workdir, cultivar returns FAIL without calling the grader, and it
   distinguishes a genuine empty result from an infrastructure failure (auth, rate-limit,
   quota markers) to give an accurate cause.
5. **Truncation salvage.** If the grader JSON is cut off by max_tokens, a regex recovers the
   leading `"pass"` verdict rather than letting a naive parse invert it to a false FAIL.
6. **Baseline-honesty caveat, documented.** The without-skill baseline is not strictly
   identical across runners; cost comparisons are flagged as directionally useful, not
   like-for-like. Rare methodological candour.
7. **Cost and tokens are first-class quality axes**, measured per variant.

## Transferable ideas, mapped to this project

### Idea 1, Differential efficacy scoring (highest value)

Run each eval scenario with and without the SKILL.md in context and score the *delta*. No
current dimension measures whether the skill actually changes agent output; a skill can score
an A on all 9 dimensions and be inert. Implement as `--differential` in the native eval
runner: two Claude calls per scenario (system with vs without the body), judge both, report
`pass_rate_delta` and `quality_delta`. Candidate new scorer **D10 Behavioural Lift**, or a
metric feeding D9. Autofail the with-arm if it produced no differentiated output. This idea
also appears independently in the skillvitals finding; treat as shared scope.

### Idea 2, The `with-docs` distillation-value control

A third arm that injects the skill's raw source references instead of the skill. D1 Knowledge
Delta asks "does the skill add knowledge the model lacks?" but judges it from text; this
empirically tests whether the distillation beats pasting the raw docs. If they tie, the skill
is repackaging effort the agent could do itself. Extend the scenario schema with an optional
`reference_docs:` list (analogous to `context_refs`) and add the arm when present. Feed the
delta into D1 as an empirical corroborator or a **D1b Distillation Value** sub-metric.

### Idea 3, State-based verification hooks

Add `setup` / `verify` / `teardown` string fields to the scenario schema; the runner executes
them around the agent turn and injects `verify_output` into the judge prompt. Lets a scenario
assert on post-run world state (files written, a command's exit output) rather than only what
the agent said. Pairs naturally with a sandboxed runner.

### Idea 4, Calibration examples to anchor the judge (anti-drift)

Per-scenario labelled pass/fail example runs, filtered by scenario id, prepended to the judge
prompt. A reliability upgrade to the D9 judge: reduces variance and encodes "this specific
wrong behaviour = FAIL" without over-specifying prose criteria. Add an `examples/` convention
next to `evals/`. Cheap, high leverage.

### Idea 5, Judge evidence-provenance constraint (anti-contamination)

Instruct the D9 judge that evidence may be quoted only from the run output, never from the
skill text/references/examples, and post-check it: penalise a verdict whose quoted `evidence`
substring appears in the SKILL.md rather than the trace. Plus the autofail and JSON-salvage
guards above.

### Idea 6, Repeats and a consistency metric

Run each scenario N times, report pass rate `k/N` and variance. No dimension captures
consistency; a skill that helps 2/5 runs is worse than 5/5. Add `--repeat N`, aggregate to a
reliability score, surface as a D9 sub-metric. High-variance scenarios are flagged as
unreliable eval material, which also improves the eval-quality gate.

### Idea 7, Cost/token efficiency as a quality axis

Capture the usage block and cost per variant and compare with-skill vs baseline vs with-docs.
D5 and D8 gesture at economy from document structure only; this measures the actual run-time
footprint. A skill that passes but inflates cost vs raw docs is a real regression. Feed an
efficiency delta into D5 or D8 or report it alongside grades.

### Idea 8, Cross-agent portability testing

Run the same scenario across multiple agent runners behind a common interface. D4 checks spec
conformance statically; portability asks empirically whether the skill still helps on a
different agent. Relevant given the `agents/` registry and multi-environment `init`. Start
Claude-only; the runner-adapter abstraction is the valuable part to copy. Longer horizon.

### Idea 9, Judge-generated remediation suggestions (cause to fix)

On fail, have the judge return `suggestions: [{cause, fix}]` and route them into the existing
`remediate` command, so eval failures produce actionable plan steps automatically. Bridges D9
failures into remediation without new machinery.

## Technical considerations

| Concern | Assessment |
|---|---|
| Language boundary | Project is Go-only. cultivar is Python; port the technique (differential arms, grader guards, metrics), do not bridge. |
| What we already have | Claude-API judge runner (`cmd/eval.go`), scenario format under `cmd/assets/evals/`, `remediate` command, `reporter/` persistence. |
| What we would write | Second (and optional third) eval arm, delta metrics, scenario-schema fields (`reference_docs`, `setup`/`verify`/`teardown`, `repeat`), grader anti-contamination + autofail + salvage, usage/cost capture. |
| Complexity | Medium. Idea 1 is the anchor; Ideas 4 to 5 are cheap judge hardening; Ideas 2, 6, 7 build on the same differential plumbing; Idea 8 is longer horizon. |
| Relation to other findings | Idea 1 is the same behavioural axis as the skillvitals finding's counterfactual efficacy. Aligns with the superseded migrate-off-tessl-eval plan's Phase 2 (skills proving they change behaviour). |

## Recommendation

Build **Idea 1 (differential efficacy)** as the anchor and layer **Ideas 4 and 5 (calibration
examples, evidence-provenance guards)** as cheap hardening of the D9 judge we already run.
**Idea 2 (with-docs distillation control)** is the highest-signal follow-on because it is the
only empirical test of D1 Knowledge Delta. cultivar and this project are complementary: this
project is a deep static rubric auditor, cultivar is a differential behavioural framework; the
value is importing the differential framing into the native eval runner, not adopting the tool.
