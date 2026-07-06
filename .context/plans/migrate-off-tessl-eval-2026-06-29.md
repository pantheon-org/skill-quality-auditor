---
title: "Draft Plan: Migrate Skill Evaluation off Tessl"
type: PLAN
status: DRAFT
date: 2026-06-29
value: HIGH
effort: TBD
themes:
  - EVAL
related:
  - ../findings/eval-gating-byok-2026-06-29.md
---
# Draft plan: migrate skill evaluation off Tessl

Status: DRAFT for review
Date: 29-06-2026
Branch (proposed): `feat/native-eval-runner`
Author: investigation by AI agent, decisions pending human owner

**Effort:** TBD — Layer 1 (the native eval runner itself) already shipped via `native-eval-runner-2026-07-01.md` (done). Remaining scope depends on the 6 open owner decisions in Section 9, particularly whether distribution/packaging (Layer 3) is also dropped — re-size once those are answered.

> This is decision-support material. The final approach, scope, and any
> CI or secret changes must be decided and documented by a human maintainer
> before implementation.

## 1. Goal

Remove the hard runtime dependency on the Tessl CLI and the `TESSL_TOKEN`
secret for **evaluating** the skill, while keeping the runtime validation
signal that dimension D9 depends on. Distribution and packaging concerns
(`tile.json`, the Tessl registry) are treated as a separate, out-of-scope
follow-up and are listed in Section 8.

## 2. What "depends on Tessl" actually means here

The investigation found that Tessl coupling falls into four distinct layers.
Only the first is a runtime evaluation dependency.

### Layer 1 — Runtime eval harness (in scope)

The only place Tessl is invoked to evaluate the skill is CI:

`.github/workflows/skill-quality.yml`

```yaml
- name: Install Tessl CLI
  run: curl -fsSL https://get.tessl.io | sh

- name: Tessl eval run
  env:
    TESSL_TOKEN: ${{ secrets.TESSL_TOKEN }}
  run: tessl eval run cmd/assets/
```

This step runs the skill against the scenarios in `cmd/assets/evals/` using
Tessl's hosted harness and grades them. It needs network access, the Tessl
CLI installer, and a hosted token. This is the dependency to remove.

### Layer 2 — Eval scenario format (keep as is)

`cmd/assets/evals/` follows Tessl eval-harness conventions:

```text
cmd/assets/evals/
  instructions.json        # extracted instructions, classified by why_given
  summary.json             # { instructions_coverage: { coverage_percentage } }
  scenario-01..06/
    task.md                # user prompt + expected behaviour + input
    criteria.json          # checklist[] with max_score summing to 100
    capability.txt         # one-line capability statement
```

Important finding: the Go D9 scorer (`scorer/d9_eval_validation.go`) reads
this directory **structurally and entirely in Go**. It never shells out to
Tessl. It parses `instructions.json`, `summary.json`, each
`scenario-N/criteria.json` (checklist must sum to 100), and checks
`task.md` / `capability.txt` presence. It also computes mutation coverage,
an adversarial bonus, and an independent-authoring bonus.

Consequence: the scenario format is not a Tessl runtime dependency. It is a
local convention we own. We should keep the format unchanged so the D9
scorer keeps working without edits.

### Layer 3 — Distribution and packaging (out of scope for this plan)

- `tile.json`, `tessl.json` — tile manifest and vendored dependency lock.
- `.tessl/plugins/...` — vendored Tessl plugins (`eval-improve`, `eval-setup`).
- `mcp__tessl__*` MCP tools and the `tessl install` flow referenced in
  `docs/d4-specification-compliance.md`.

These concern how the skill is published and consumed, not how it is
evaluated. Removing them is a different decision (Section 8).

### Layer 4 — Documentation and naming (light cleanup)

- `cmd/assets/references/tessl-compliance-framework.md`
- `docs/d9-eval-validation.md` (mentions "tessl eval scenarios",
  "creating-eval-scenarios" skill, `tessl eval run` / `tessl eval view-status`)
- `cmd/assets/SKILL.md` description ("ensures tessl registry compliance")
- `README.md` ("Tessl tile")

These are wording, not behaviour. They should be reconciled once the runtime
approach is chosen.

## 3. Current recommendation (as of June 2026)

Recent guidance on evaluating agent skills converges on a few points:

- **LLM-as-judge with checklist grading** is the accepted method for the
  subjective parts of skill evaluation (task completion, reasoning quality).
  Checklist evaluation breaks the judgement into per-item binary or scored
  criteria, which is exactly what our `criteria.json` already encodes.
- **Vanilla vs skills-augmented comparison** (the SkillsBench / Harbor
  pattern) is the recommended way to prove a skill actually changes
  behaviour: run each task with and without the skill and measure the
  delta. This matches the stated purpose of D9 ("runtime validation proves
  the skill actually changes agent behaviour").
- **Deterministic verification where possible**, LLM-judge only where not.
  Pin the judge model and prompt for reproducibility.
- Common third-party harnesses named are Harbor, Promptfoo, and Braintrust.
- A pinned judge model and a fixed grading prompt are needed for stable,
  CI-friendly scores.

Takeaway: we do not need a hosted service to get a credible eval. We already
own the scenario format and the checklist criteria. We need a runner that
executes each scenario against a Claude model, grades the output against the
checklist with an LLM judge, and writes results back into the existing
`summary.json` schema.

## 4. Options considered

### Option A — Native Go eval runner using the Claude API (recommended)

Add a `skill-auditor eval` command that:

1. Loads scenarios from a skill's `evals/` directory (existing format).
2. For each scenario, runs the task prompt against a Claude model with the
   skill content supplied as context (the "skills-augmented" run).
3. Grades the produced output against `criteria.json` using a second Claude
   call as an LLM judge, scoring each checklist item up to its `max_score`
   (the items already sum to 100).
4. Writes `summary.json` with `instructions_coverage.coverage_percentage`
   so the existing D9 scorer consumes it unchanged.
5. Exits non-zero below a configurable pass threshold for CI.

Pros: self-contained, no external service, reuses the format and the D9
scorer, single Go binary, fits the repo. Pinned model and prompt give
reproducibility. Cons: we own the harness and the judge prompt; needs an
Anthropic API key secret (`ANTHROPIC_API_KEY`) in CI, swapping one secret
for a more standard one.

### Option B — Adopt Promptfoo or another third-party harness

Drive the existing scenarios through Promptfoo or similar via a YAML config.

Pros: maintained externally, multi-grader support out of the box. Cons:
swaps the Tessl dependency for a Node toolchain plus a new config surface;
the scenario format would need a shim; less aligned with a Go-first repo.

### Option C — Claude Agent SDK harness scripts

Same idea as A but implemented as SDK-driven scripts rather than a Go
subcommand.

Pros: closer to a real agent loop, supports tool use if scenarios need it.
Cons: adds a second runtime (SDK language) alongside Go; heavier to maintain
than a Go subcommand for what are currently single-turn scenarios.

### Option D — Drop runtime eval, keep only the structural D9 scorer

Remove the CI eval step entirely and rely on the Go D9 scorer's structural
checks.

Pros: zero cost, immediate. Cons: loses the runtime validation that is the
entire point of D9. The skill would grade its own D9 on structure while
claiming runtime validation it no longer performs. Not recommended beyond a
short interim if needed.

### Recommendation

Option A, delivered in two phases. Phase 1 gives parity with what Tessl did
(run + checklist grade + coverage). Phase 2 adds the vanilla-vs-skill
comparison from the SkillsBench pattern to strengthen the D1 / D9 signal.

## 5. Proposed design (Option A)

### New command

```bash
skill-auditor eval <path-or-key> [flags]
  --model <id>          # judge + actor model, pinned default
  --judge-model <id>    # override judge model (default: same as --model)
  --fail-below <pct>    # CI gate, default e.g. 80
  --write-summary       # update evals/summary.json in place
  --json                # machine-readable results
  --compare             # phase 2: run vanilla vs skill-augmented
```

### Flow per scenario

1. Read `task.md`, extract the user prompt and input block.
2. Actor run: call the model with the skill content plus the task. Capture
   the output artefacts the scenario expects (for our own scenarios this is
   an audit report and a remediation plan in text form).
3. Judge run: send the output plus `criteria.json` to a pinned judge model
   with a fixed rubric prompt; require a per-item score and a one-line
   justification per checklist item.
4. Sum item scores (already normalised to 100), record pass/fail against
   `--fail-below`.

### Coverage and summary

- Compute instruction coverage from `instructions.json` against scenario
  criteria (the same overlap idea the Go scorer already uses for mutation
  coverage), and write `summary.json.instructions_coverage.coverage_percentage`.
- Keep the schema byte-compatible with what the D9 scorer reads so no D9
  edits are required.

### Reproducibility

- Pin the default model id and the judge prompt in code.
- Set temperature low / deterministic where the API allows.
- Record the model id and prompt version in the JSON output for auditability.

### Secrets and config

- Read `ANTHROPIC_API_KEY` from the environment.
- No token if `--compare` and the actor run are skipped (structure-only
  mode), so local contributors without a key can still run the structural
  checks.

## 6. CI changes

Replace the two Tessl steps in `.github/workflows/skill-quality.yml`:

```yaml
# remove
- name: Install Tessl CLI
  run: curl -fsSL https://get.tessl.io | sh
- name: Tessl eval run
  env:
    TESSL_TOKEN: ${{ secrets.TESSL_TOKEN }}
  run: tessl eval run cmd/assets/

# add
- name: Skill eval (LLM-judge)
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: ./dist/skill-auditor eval cmd/assets --fail-below 80 --write-summary
```

Open question for the owner: whether the eval step should be required on
every PR (cost and flakiness of model calls) or gated to a label / nightly
schedule, with the structural D9 check kept on every PR. See Section 9.

## 7. File-by-file change list

| File | Change | Effort |
| ---- | ------ | ------ |
| `cmd/eval.go` (new) | New `eval` cobra command and runner | L |
| `internal/anthropic/` (new) | Thin API client (actor + judge calls) | M |
| `cmd/eval_test.go` (new) | Tests with a mocked client, fixture scenarios | M |
| `.github/workflows/skill-quality.yml` | Swap Tessl steps for `eval` step | S |
| `docs/d9-eval-validation.md` | Replace `tessl eval run` guidance with new command; keep format description | S |
| `cmd/assets/SKILL.md` | Reword description away from "tessl registry compliance" if Layer 3 also removed; otherwise leave | S |
| `README.md` | Update "Tessl tile" wording per Layer 3 decision | S |
| `cmd/assets/references/tessl-compliance-framework.md` | Decide: keep as registry-submission guidance, or fold the agent-agnostic / portability checks into a renamed reference | M |
| `CLAUDE.md` / `AGENTS.md` | Replace "Tessl eval changes require `tessl eval run`" rule with the new command | S |

Phase 2 additions (vanilla-vs-skill comparison):

| File | Change | Effort |
| ---- | ------ | ------ |
| `cmd/eval.go` | `--compare` mode, delta reporting | M |
| `scorer/d9_eval_validation.go` | Optional: consume a recorded efficacy delta as a bonus signal | M |

## 8. Out of scope (separate decision)

Distribution / packaging removal is not required to stop evaluating with
Tessl and should be decided separately:

- `tile.json`, `tessl.json`, `.tessl/plugins/...`
- The `tessl install` consumption path and `mcp__tessl__*` tooling
- Whether the skill stays published on the Tessl registry at all

If the skill remains on the registry, the Layer 4 wording should keep an
honest description of registry compliance while making clear that runtime
evaluation is now native.

## Related findings

- [eval-gating-byok-2026-06-29.md](../findings/eval-gating-byok-2026-06-29.md)
  — pushback on CI gating, local runs, and bring-your-own-key. To be folded in
  later (suggests a new Section 12 and a two-tier gate model).

## 9. Open decisions for the owner

1. Confirm scope: evaluation only (Layers 1, 2, 4 wording), or also drop
   distribution (Layer 3)?
2. Option A vs B vs C. Recommendation is A.
3. CI cadence for the model-driven eval: every PR, or label / nightly with
   structural D9 kept on every PR?
4. Pinned judge model id and a default `--fail-below` threshold.
5. New secret `ANTHROPIC_API_KEY` provisioning and ownership.
6. Phase 2 scope: ship vanilla-vs-skill comparison now or later?

## 10. Verification (once implemented)

- `go test ./...` passes, including new `cmd/eval` tests.
- `./dist/skill-auditor eval cmd/assets --fail-below 80` runs the six
  scenarios and reports per-scenario and total scores.
- `summary.json` is regenerated with a coverage percentage and the existing
  D9 scorer reads it without changes (`go test ./scorer/...`).
- CI green with the Tessl steps removed and `TESSL_TOKEN` no longer
  referenced.
- `grep -ril tessl` shows only intentional, documented references.

## 11. Critical review (29-06-2026)

Decision-support note: the findings below were produced by an AI agent
reviewing this plan against the current codebase. A human maintainer must
decide which to action before implementation.

### Claims verified against the code

- Layer 1 is accurate: the two Tessl steps in
  `.github/workflows/skill-quality.yml` match the quoted YAML exactly, and
  `TESSL_TOKEN` is referenced only there (plus this plan).
- The central claim holds: `scorer/d9_eval_validation.go` reads `evals/`
  entirely in Go and never shells out to Tessl. Its one `os/exec` call is
  `git log --follow` for the independent-authoring bonus, not Tessl.
- `summary.json` is exactly `instructions_coverage.coverage_percentage`, and
  the D9 field is typed `any` with loose parsing, so the schema-compatibility
  requirement is easier to meet than the plan implies.
- `criteria.json` checklists sum to 100 as described.

### Substantive concerns (resolve before approval)

1. **Flaky CI gate.** A hard `--fail-below 80` gate on a single LLM-judge
   sample will be non-deterministic; the Claude API is not reproducible even
   at temperature 0. The design lists reproducibility as a Pro but provides no
   mechanism. Needs N-sample median per scenario, a margin band rather than a
   knife-edge threshold, or equivalent, before it is CI-safe.
2. **Actor-run fidelity is conditional.** The current scenarios (for example
   `scenario-01/task.md`) are single-turn reasoning tasks, so Option A's
   "skill content in context, capture text output" is well matched. State this
   as an explicit precondition: if any future scenario requires invoking the
   `skill-auditor` binary or other tools, Option A degrades to grading the
   model's imitation of the tool and Option C becomes necessary. This, not
   maintenance weight, is the real reason A beats C.
3. **Circular subject / actor / judge.** The skill teaches Claude to score
   skills, the actor is Claude applying it, and the judge is Claude grading
   against criteria the skill shaped. Pinning model and prompt does not remove
   the conflict. Name this as a known limitation of the D9 number.
4. **`--write-summary` in CI mutates a tracked file.** The Section 6 snippet
   overwrites the checked-in `summary.json` during the run, which is either
   discarded or causes drift. The CI gate should be read-only (assert and
   exit); reserve `--write-summary` for local authoring.
5. **Coverage computation looks redundant.** D9 already derives mutation
   coverage from SKILL.md imperatives vs `criteria.json` and separately reads
   `summary.json.coverage_percentage`. Clarify what the runner-computed
   `coverage_percentage` certifies, or leave `summary.json` as the static
   marker it currently is.

### Smaller issues

- **Change list incomplete against its own verification.** Section 10 expects
  `grep -ril tessl` to show only intentional references, but tessl references
  also exist in `CONTRIBUTING.md`, `docs/d4-specification-compliance.md`,
  `.mcp.json`, `.mcpx.json`, `.vscode/mcp.json`, and `opencode.json`. Bucket
  the MCP config files explicitly (Layer 3) and mark `CHANGELOG.md` as
  deliberately left (history), or the final grep will read as incomplete work.
- **Cost unquantified.** Six scenarios times actor plus judge calls, doubled
  in Phase 2, on a workflow that triggers on `**/*.go` (most PRs). Provide a
  rough per-run cost figure to inform the cadence decision (Open decision 3).
- **Security side effect.** Removing `curl -fsSL https://get.tessl.io | sh`
  eliminates a curl-pipe-to-shell step in CI; list this as a benefit. Adding
  `ANTHROPIC_API_KEY` introduces an outbound path that sends skill content to
  the API in CI (low concern, own content, but note it).
- **Branch coordination.** Proposed branch `feat/native-eval-runner` overlaps
  with other in-flight migration plans (hk/markdownlint, go-cli-script-parity-2026-04-27).
  Add a sequencing note for the CI workflow edits.

### Overall

Approvable in principle. Before implementation the owner should resolve, at
minimum: the flaky-gate mitigation, read-only CI vs `--write-summary`, the
coverage-percentage semantics, and CI cadence with a cost estimate; and add an
explicit statement that Option A is valid only while scenarios remain
single-turn.
