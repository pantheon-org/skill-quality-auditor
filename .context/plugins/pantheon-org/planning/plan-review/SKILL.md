---
name: plan-review
description: >
  Review .context/plans/*.md files using 3 independent subagent reviewers:
  Technical (feasibility, gaps, soundness), Strategic (scope, alignment, priority),
  and Risk (blind spots, edge cases, failure modes). The skill FIRST asks the user
  which models to assign to each reviewer — presenting available options based on
  the user's environment (OpenCode Zen, Go, Claude Code, or BYOK). Each reviewer
  receives the same plan with a role-specific lens and returns structured feedback.
  The main agent consolidates all three perspectives. Triggers: 'review this plan',
  'audit plan', 'plan review', 'check my plan', 'what's wrong with this plan',
  'review all plans', 'plan quality check'. Do NOT use for plans not in .context/plans/,
  one-off notes, or external documents.
---

# Plan Review — 3-Agent Multi-Perspective Audit

Review `.context/plans/` files through 3 independent lenses: **Technical**,
**Strategic**, and **Risk**. Each reviewer is a separate subagent with a unique
prompt and perspective. The main agent collates results into a consolidated report.

This catches what a single reviewer normalises — each subagent brings a fresh
set of assumptions and focus areas.

## Prerequisites

- A `.context/plans/*.md` file exists to review (or a glob selecting multiple)
- The plan has standard YAML frontmatter (`title`, `type`, `status`, `date`)
- Your environment supports spawning `general` and `explore` subagent types

## Quick Start

```bash
# No CLI commands — this is an agent workflow skill.
# Load the skill, then say: "review the plan at .context/plans/<name>.md"
```

## When to Use

- A `.context/plans/*.md` file needs an independent multi-perspective review before implementation
- A draft plan needs validation before marking it `ACTIVE`
- Multiple plans exist in the same domain and need prioritisation
- A stale plan needs a freshness check against current project state
- The user explicitly asks for a plan audit

## When NOT to Use

- For one-off notes, scratch files, or non-plan documents — use `session-reflection` instead
- For plans outside `.context/plans/` — the reviewer prompts assume `.context/` frontmatter conventions
- When the plan is trivially small (1 paragraph, no steps) — the overhead of 3 subagents is not justified
- When the user explicitly asks for a quick opinion, not a full audit

## Mindset

- Three independent reviewers catch what a single reviewer normalises. The value is in the divergence between their findings, not in consensus.
- Model diversity is the strongest lever for perspective diversity. Different models trained on different data notice different failure modes.
- A plan review is a service to the plan author, not a judgement. Findings should be actionable, not critical.
- Structural validation is a prerequisite, not the goal. The frontmatter check exists to keep the plan visible in the index, but the real value is the implementation architecture and risk analysis from the 3 reviewers.
- If a review reveals a critical issue, the right outcome is to improve the plan, not to reject it.

## Workflow

### 1. Identify the plan to review

The user specifies a plan path, or you suggest one from `.context/plans/`.
Read the plan in full. Note its frontmatter (`status`, `date`, `related`),
goal, steps, open questions, and risks.

### 2. Compose the plan brief

From the plan file, create a structured brief that all 3 reviewers will receive.
The brief must be self-contained — each reviewer should not need to read the
original plan file. Include:

- **Title, status, date** — from frontmatter
- **Goal** — the plan's stated objective (verbatim if concise, summarised if long)
- **Steps** — numbered list of implementation steps
- **Dependencies / related files** — from `related:` frontmatter
- **Open questions** — listed verbatim
- **Known risks** — listed verbatim from the plan's risk section (if any)
- **Implementation status** — what's already been done vs what's still pending
- **Context** — any relevant background from related findings or docs

### 3. Validate plan structure against the standard trio

Every `.context/` plan should follow a **three-part standard**: a YAML template
defining required fields, a JSON schema for machine validation, and a validation
script. Check the plan against all three.

#### The trio

| Component | Location | What it defines |
|-----------|----------|-----------------|
| **YAML template** | `context-file` SKILL.md (frontmatter section) | Required fields: `title`, `type`, `status`, `date`, optional `related` |
| **JSON schema** | `context-file/assets/schemas/context-frontmatter.schema.json` | Field types, allowed values (`enum`), patterns, required list |
| **Validation script** | `context-index/scripts/validate-context-frontmatter.sh` | CLI script that reads the schema and checks one or more files against it |

#### What to check

Run the validation script on the plan file:

```bash
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/<plan-file>.md
```

This checks:
- Frontmatter exists and opens/closes with `---`
- All required fields present (`title`, `type`, `status`, `date`)
- `type` is one of the allowed values (`PLAN`, `FINDING`, `ANALYSIS`, `INSTRUCTION`, `AUDIT`)
- `status` is one of `DRAFT`, `ACTIVE`, `DONE`, `SUPERSEDED`
- `date` matches `YYYY-MM-DD` pattern
- No extra fields beyond those defined in the schema

#### If validation passes

Include `**Structure: Valid**` in the plan brief. Proceed to model selection.

#### If validation fails

Capture each validation error and include it in the plan brief under a
`## Structural Issues` section so all 3 reviewers can reference them. Sample:

```
## Structural Issues
- Missing required field: 'status'
- 'date' does not match pattern YYYY-MM-DD, got '01-07-2026'
```

#### Plan body structure — infer the local standard

Beyond frontmatter, check the plan's body sections (`##` headings) against what
other plans in this repo actually use. Do NOT hardcode a template — scan the
existing plans and infer what's conventional.

##### How to infer

Run this to collect all H2 headings from every plan:

```bash
grep -r '^## ' .context/plans/*.md | sed 's/.*## //' | sort | uniq -c | sort -rn
```

This gives a frequency table like:

```
 7 ## Goal
 7 ## Steps
 5 ## Open Questions
 4 ## Scope
 4 ## Related
 4 ## Decisions
 3 ## Verification
 3 ## Risks
 3 ## Success Criteria
 3 ## Background
```

From this, identify the **core sections** (present in >= 40% of plans), the
**common sections** (present in 20-39%), and the **rare sections** (<20%).

##### What to report

Include in the plan brief under `## Structural Issues`:

```
### Body Structure (inferred from <N> existing plans)
Core sections (≥40% of plans):
  - Goal: present/missing
  - Steps/Phases: present/missing  
  - Open Questions/Decisions: present/missing

Common sections (20-39%):
  - Scope/Out of scope: present/missing
  - Related findings: present/missing
  - Verification: present/missing
  - Risks/sequencing: present/missing

Note: Sections like Background, File-by-file change list, Academic References,
and Critical Review appear in <20% of plans and are optional.
```

Missing core sections are **advisory warnings** — they don't invalidate the plan
(plans can be valid with different structures) but the reviewer should note them.
Missing common sections are **neutral observations**.

Rare or unique sections are also interesting — they may indicate scope creep or
a plan that doesn't fit the local conventions. Note them as "unusual sections"
but do not flag them as problems.

##### Why infer instead of hardcode

Hardcoding a template (Goal → Steps → Open Questions) penalises plans that
legitimately need a different structure. Inferring from actual plans reflects
what this repo's conventions actually are — and those conventions evolve as
new plans are added. A plan that omits `## Goal` but has a `## Problem` and
`## Target State` is following the spirit of the convention, just not the
letter.

#### Implementation architecture — Phases, Tasks, and Waves

A well-structured plan divides the work into **Phases**, **Tasks**, and
**Waves** — and the review should check for this structure.

| Term | Meaning | Example |
|------|---------|---------|
| **Phase** | Sequential stage — Phase N must finish before Phase N+1 starts | Phase 1: Foundation, Phase 2: Integration |
| **Task** | Smallest unit of work within a phase — independently assignable | Task 1.1: Create schema, Task 1.2: Write migration |
| **Wave** | Group of tasks within a phase that can run in **parallel** | Wave A (can run with Wave B): frontend + backend |

Common bad patterns to flag:

- **Flat step list** — 15 numbered steps with no grouping, no parallelisation.
  Every step waits for the previous one, even when they don't depend on each other.
- **Missing dependency annotation** — steps are listed sequentially but some
  could run in parallel. The plan doesn't say which.
- **Oversized phases** — "Phase 1: Everything" with 20+ tasks and no sub-structure.
  A phase should be small enough that it can be reviewed and shipped independently.
- **Missing task boundaries** — steps describe vague outcomes ("improve things")
  instead of concrete, completable units ("add --format flag to evaluate cmd").

##### What to check

Look at the plan's `## Steps` / `## Implementation` / `## Phases` section:

1. **Are steps grouped into phases?** The plan should have 2-5 sequential phases
   (numbered or labelled). If there are >10 flat steps with no grouping, flag it.

2. **Are tasks the right size?** Each task should be completable in a single
   session (hours, not weeks). Tasks with vague descriptions ("finalise",
   "improve", "handle remaining") are a red flag.

3. **Are parallel waves identified?** Within a phase, are there tasks explicitly
   marked as parallelisable? Look for language like "can run in parallel",
   "independent of", "Wave A / Wave B". If all tasks are strictly sequential
   without justification, flag it.

4. **Are dependencies explicit?** Does the plan say why Phase 1 must come before
   Phase 2? Are cross-phase dependencies called out? If dependencies are
   implicit, that's a risk.

5. **Are there concrete exit criteria per phase?** Each phase should define what
   "done" means — a deliverable, a passing test, a merged PR. Phases that end
   with "iterate" or "review" without a concrete gate are risky.

##### What to report

Include in the plan brief under `## Structural Issues`:

```
### Implementation Architecture
Phases: <N> phases identified — <list phase names>
   - Phase dependencies: explicit/implicit/missing
   - Phase exit criteria: present for <N>/<N> phases

Task granularity: good / too vague / too large
   - <count> of <total> tasks have actionable descriptions

Parallelisation:
   - Waves identified: yes/no
   - Tasks flagged as parallel: <N> of <N>
   - Parallelism opportunity: <N> tasks could run concurrently but aren't grouped


#### Why this matters

A plan with structurally invalid frontmatter will be silently excluded from
`.context/index.yaml` — making it invisible to other agents. Catching this
during review prevents the plan from becoming orphaned context.

### 4. Check opencode.json for existing model config

Before asking the user anything, check whether `opencode.json` already has
`subAgents` configured:

```bash
cat opencode.json | python3 -c "import sys,json; c=json.load(sys.stdin); print('configured' if c.get('subAgents') else 'not-configured')"
```

If `subAgents` is configured with at least two distinct models, note the
mapping and proceed to compose the plan brief (skip model selection — the
user already configured it). Include the model assignments in the final report.

If **not configured**, proceed to step 5 and advise the user.

### 5. Ask the user which models to use

Do NOT proceed without asking. The agent should proactively advise based on
what it detects. First check if this is Claude Code or OpenCode by looking
for environment cues (e.g., `CLAUDE_CODE` env var, or `ANTHROPIC_API_KEY`
is set). If you cannot determine the environment, ask:

> "Are you using **OpenCode** or **Claude Code**? This determines which models are available for the subagent reviewers."

Use the appropriate guidance below.

#### If OpenCode (no subAgents configured)

Advise the user with a capability-vs-cost breakdown. Present options using the
`question` tool. The goal is to assign one model to **Technical + Strategic**
(both use `general` subagent type) and a **different model** to **Risk**
(`explore` subagent type). Different models = different blind spots.

**OpenCode Zen (pay-as-you-go) — capability vs cost:**

| Tier | Pair | Tech/Strategic | Risk | Cost/run | When |
|------|------|---------------|------|----------|------|
| **Best value** (Recommended) | DS Flash + GPT 5 Nano | DeepSeek V4 Flash ($0.14/M in) | GPT 5 Nano ($0.05/M in) | ~$0.03 | Default for most reviews |
| **Balanced** | DS Flash + Claude Haiku | DeepSeek V4 Flash ($0.14/M in) | Claude Haiku 4.5 ($1.00/M in) | ~$0.08 | Want an Anthropic perspective on risk |
| **Max depth** | DS Pro + GPT 5 Nano | DeepSeek V4 Pro ($1.74/M in) | GPT 5 Nano ($0.05/M in) | ~$0.12 | Critical, high-stakes plan review |
| **Budget** | GPT 5 Nano (all) | GPT 5 Nano ($0.05/M in) | GPT 5 Nano ($0.05/M in) | ~$0.01 | Draft plan, quick sanity check |

**OpenCode Go (flat-rate subscription):**

All models cost the same (no per-token charge). Prioritise diversity:

```jsonc
{
  "subAgents": {
    "general": { "model": "deepseek-v4-flash" },   // Technical/Strategic
    "explore": { "model": "minimax-m3" }            // Risk — different family
  }
}
```

Recommended pairs:

| Pair | Tech/Strategic | Risk | Why |
|------|---------------|------|-----|
| **DS Flash + MiniMax** (Recommended) | DeepSeek V4 Flash (1M ctx) | MiniMax M3 (512K ctx) | Different model families, good diversity |
| **DS Flash + GLM-5.2** | DeepSeek V4 Flash (1M ctx) | GLM-5.2 (1M ctx) | Risk gets same giant context for long plans |
| **DS Flash + Kimi Code** | DeepSeek V4 Flash (1M ctx) | Kimi K2.7 Code (262K ctx) | Code-oriented risk lens |

Example `question` tool call (structure defined in [assets/templates/model-selection.yaml](assets/templates/model-selection.yaml),
validated by [assets/schemas/model-selection.schema.json](assets/schemas/model-selection.schema.json)):

```
header: "Model selection for plan reviewers"
question: "I need to assign models for 3 subagent reviewers..."
options:
  - "Best value — DS Flash + GPT 5 Nano" (Recommended) — $0.03/run
  - "Balanced — DS Flash + Claude Haiku" — $0.08/run
  - "Max depth — DS Pro + GPT 5 Nano" — $0.12/run
  - "Budget — GPT 5 Nano for all" — $0.01/run
  - "Custom — let me specify per reviewer"
```

If the user picks "Custom", ask them for each reviewer individually and
suggest sensible defaults based on their chosen model.

#### If Claude Code

Advise the user on the Haiku vs Sonnet vs Opus tradeoff. All three are
Anthropic models available in Claude Code:

| Model | Input/MTok | Output/MTok | Context | Reasoning | Best for |
|-------|-----------|------------|---------|-----------|----------|
| **Claude Haiku 4.5** | $1.00 | $5.00 | 200K | Adequate | Risk reviewer — cheap, fast, still catches blind spots |
| **Claude Sonnet 5** | $3.00 | $15.00 | 1M | Strong | Technical/Strategic — sweet spot for depth |
| **Claude Opus 4.8** | $5.00 | $25.00 | 1M | Maximum | Both reviewers — only if plan is mission-critical |

Recommended pairs:

| Pair | Tech/Strategic | Risk | Cost/run | When |
|------|---------------|------|----------|------|
| **Sonnet + Haiku** (Recommended) | Claude Sonnet 5 | Claude Haiku 4.5 | ~$0.10 | Best capability/cost balance |
| **Sonnet + Sonnet** | Claude Sonnet 5 | Claude Sonnet 5 | ~$0.15 | Same model everywhere (less diverse) |
| **Opus + Haiku** | Claude Opus 4.8 | Claude Haiku 4.5 | ~$0.18 | Max depth on technical analysis, cheap risk |
| **Opus + Sonnet** | Claude Opus 4.8 | Claude Sonnet 5 | ~$0.25 | Both reviewers get strong reasoning |

Example `question` tool call (structure in [assets/templates/model-selection.yaml](assets/templates/model-selection.yaml)):

```
header: "Model selection for plan reviewers"
question: "I'll use 3 subagent reviewers..."
options:
  - "Sonnet + Haiku" (Recommended)
  - "Opus + Haiku"
  - "Opus + Sonnet"
  - "Sonnet for all"
  - "Custom — let me specify per reviewer"
```

#### If Bring Your Own Key (BYOK)

Ask the user which models their provider supports. If unsure, recommend:
- Technical/Strategic: the strongest reasoning model available
- Risk: a different model (ideally from a different provider family)

#### If the user says "not sure" or "surprise me"

Recommend the best-value diversity pair:
- **OpenCode Zen:** DeepSeek V4 Flash (Technical/Strategic) + GPT 5 Nano (Risk) — ~$0.03/run
- **OpenCode Go:** DeepSeek V4 Flash (Technical/Strategic) + MiniMax M3 (Risk) — flat-rate
- **Claude Code:** Claude Sonnet 5 (Technical/Strategic) + Claude Haiku 4.5 (Risk) — ~$0.10/run
- **BYOK:** Strongest model (Technical/Strategic) + cheapest different model (Risk)

### 6. Map model choices to subagent configuration

If the user wants to configure `opencode.json`, show them the snippet:

```jsonc
{
  "subAgents": {
    "general": {
      "model": "<chosen Technical/Strategic model>",
      "systemPrompt": "You are a thorough, detail-oriented reviewer."
    },
    "explore": {
      "model": "<chosen Risk model>",
      "systemPrompt": "You are a skeptical, adversarial reviewer focused on finding flaws."
    }
  }
}
```

Ask if they want to apply this now or proceed with current routing (may use
the same model for all types). If they proceed without configuration, run all
3 but note in the output that model diversity was limited — the reviewer
prompts alone provide the perspective separation.

### 7. Spawn 3 reviewers (in parallel)

Spawn all 3 subagents at once using the `task` tool. Each receives the
**same plan brief** prefixed with their role-specific instructions.

| Reviewer | subagent_type | Lens |
|----------|--------------|------|
| Technical | `general` | Feasibility, implementation gaps, technical soundness |
| Strategic | `general` | Goal alignment, scope correctness, priority, sequencing |
| Risk | `explore` | Blind spots, edge cases, failure modes, dependencies |

#### Reviewer 1: Technical

Use this prompt template with the plan brief inserted:

```json
{
  "subagent_type": "general",
  "description": "Technical review of plan",
  "prompt": "You are a TECHNICAL REVIEWER evaluating a plan. Your focus: feasibility, implementation gaps, technical correctness, and completeness.\n\nAnswer these questions with specific, actionable items:\n\n1. FEASIBILITY: Are the implementation steps technically sound? Identify any steps that are ambiguous, under-specified, or impossible as written.\n\n2. GAPS: What specific details are missing? (e.g., missing error handling, no test strategy, no rollback plan, missing configuration). Be precise — reference step numbers.\n\n3. CONSISTENCY: Do the steps internally agree? Any contradictions in approach, ordering, or assumptions?\n\n4. EFFORT ESTIMATE: Is the effort realistic for the scope? Identify any steps that look under- or over-scoped.\n\n5. DEPENDENCIES: Are all prerequisites and dependencies called out? Any implicit ones that should be explicit?\n\nPLAN BRIEF:\n<PASTE PLAN BRIEF HERE>"
}
```

#### Reviewer 2: Strategic

```json
{
  "subagent_type": "general",
  "description": "Strategic review of plan",
  "prompt": "You are a STRATEGIC REVIEWER evaluating a plan. Your focus: goal alignment, scope, priority, and completeness relative to objectives.\n\nAnswer these questions with specific, actionable items:\n\n1. GOAL ALIGNMENT: Does every step clearly serve the stated goal? Identify any steps that are off-track, scope-creep, or not connected to the goal.\n\n2. SCOPE: Is the scope right? Too narrow (misses important outcomes) or too broad (trying to solve everything)?\n\n3. PRIORITY: Are the steps in the right order? Identify sequencing issues — steps that should come earlier or later, or missing prerequisites.\n\n4. COMPLETENESS: Are there obvious missing steps or phases? What would a reasonable person expect that isn't here?\n\n5. SUCCESS CRITERIA: Does the plan define what 'done' looks like? If not, what's missing?\n\nPLAN BRIEF:\n<PASTE PLAN BRIEF HERE>"
}
```

#### Reviewer 3: Risk

```json
{
  "subagent_type": "explore",
  "description": "Risk review of plan",
  "prompt": "You are a RISK REVIEWER evaluating a plan. Your focus: blind spots, edge cases, failure modes, and unstated assumptions.\n\nAnswer these questions with specific, actionable items:\n\n1. BLIND SPOTS: What is the plan not considering? Look for implicit assumptions that could be wrong, stakeholders not mentioned, or system boundaries not considered.\n\n2. FAILURE MODES: For each step, what's the most likely failure? What's the worst-case outcome? Be specific — reference step numbers.\n\n3. EDGE CASES: What scenarios would break this plan? Consider: incomplete data, external dependencies failing, concurrent changes, user error.\n\n4. RECOVERY: If a step fails, is there a rollback or recovery path? Identify steps where failure = irreversible or costly.\n\n5. RESILIENCE: How brittle is this plan? Could small changes in assumptions invalidate large parts of it?\n\nPLAN BRIEF:\n<PASTE PLAN BRIEF HERE>"
}
```

### 8. Receive and collate results

Each subagent returns a structured review. Collect all 3 and compile a
**Consolidated Review Report** following the structure defined in
[assets/templates/review-report.yaml](assets/templates/review-report.yaml).
The template is validated against [assets/schemas/review-report.schema.json](assets/schemas/review-report.schema.json)
— run `scripts/validate-review-report.sh` on the generated report to confirm
schema compliance.

Required sections: Model Configuration, Structural Validation, Implementation
Architecture, Scores, Critical Issues, Moderate Concerns, Strengths, Next Actions.

### 9. Present to the user

Present the consolidated report with a preamble that includes the models used:

> "I ran this plan through 3 independent reviewers using **<model A>** (Technical/Strategic) and **<model B>** (Risk). Here's the consolidated report:"

If the user wants to dig deeper on any finding, offer to spawn a follow-up
investigation subagent with the relevant context.

### 10. Classify findings and resolve — this step is not optional

Step 9's presentation is a progress update, not sign-off. Before the review counts as
finished, sort every item in `critical_issues` and `moderate_concerns` into one of two
buckets:

- **Editorial** — a contradiction, missing detail, wrong ownership, or unspecified
  mechanism with one clearly correct fix given the rest of the plan and this repo's
  existing conventions. No real tradeoff exists; a competent engineer looking at the
  same facts would land on the same fix. ALWAYS apply these directly to the plan file —
  don't make the user approve something that isn't actually a choice.
- **Decision** — a genuine tradeoff with two or more valid answers (hard-fail vs.
  advisory rollout, gate-on-condition vs. run-unconditionally, spike-first vs.
  attempt-and-recover). No amount of re-reading the repo resolves these; only the plan
  owner's judgment does.

For decision-classified items, NEVER leave them sitting in `## Open Questions` for the
user to notice later — run a short interview immediately: one question at a time, each
with concrete mutually-exclusive options plus room for free text, matching
`guided-interview`'s pattern. End with a one-message recap of every answer and get
explicit confirmation before writing anything into the plan.

Fold both outcomes back into the plan file, not just into chat:

- Editorial fixes land directly in the relevant section (Scope, Phases, Verification).
- Decisions land in a `## Decisions` section, each with the chosen option and why the
  alternative didn't win. If a decision carries a review-later condition (e.g. "flip to
  hard-fail after N clean runs"), state it as a concrete, checkable revisit trigger —
  never "revisit later" — the same rule `design-debate` enforces on its own verdicts.
- A `## Decisions` heading is itself an ADR-capture trigger in this repo
  (`check-undocumented-decisions.sh` flags any `## Decision` heading and fails
  `hk`'s pre-commit check). Check whether an existing ADR already covers the plan's
  topic; if not, run `adr-capture` before treating the amendment as done — this is a
  hard gate, not a suggestion, since the pre-commit hook will block the commit either
  way.

### 11. (Optional) Persist a standalone finding

If the review surfaced something beyond this one plan — a pattern likely to recur in
future reviews, a gap in a related skill — capture it as a finding via `context-file`,
separate from the decisions already folded into the plan in Step 10.

## When NOT to Use

- For one-off notes, scratch files, or non-plan documents — use `session-reflection` instead
- For plans outside `.context/plans/` — the reviewer prompts assume `.context/` frontmatter conventions
- When the plan is trivially small (1 paragraph, no steps) — the overhead of 3 subagents is not justified
- When the user explicitly asks for a quick opinion, not a full audit

## Anti-Patterns

**NEVER** — Skip the plan brief and just pass the file path

**SYMPTOM:** Reviewer produces shallow, generic analysis because they lack the full plan context. Each reviewer may interpret the plan differently, leading to inconsistent feedback.

**CONSEQUENCE:** The review misses critical issues because no reviewer had the complete picture. The final report contradicts itself across reviewer sections.

**WHY:** Subagents cannot read files unless they have the `general` type with file access. The plan brief ensures every reviewer works from the same complete information regardless of subagent capability.

**BAD:** Sending "review .context/plans/foo.md" as the prompt.
**GOOD:** Extracting the plan content into a self-contained brief first.

**NEVER** — Spawn reviewers sequentially

**SYMPTOM:** The review takes 3× longer than necessary. The user waits while each subagent initialises and runs in turn.

**CONSEQUENCE:** Context builds up between sequential calls, increasing cost and delay. The user may interrupt before all reviews complete.

**WHY:** They are fully independent — sequential spawning wastes time and context. Always use parallel tool calls.

**BAD:** Spawn Technical, wait for result, spawn Strategic, wait, spawn Risk.
**GOOD:** Spawn all 3 in one message with 3 parallel `task` tool calls.

**NEVER** — Proceed without asking about models first

**SYMPTOM:** The review uses the same model for all 3 reviewers, producing three reports with the same blind spots. The user never knew they had a choice.

**CONSEQUENCE:** The cost of model diversity awareness was zero, but the cost of missing a critical blind spot is unbounded. The user loses trust when they discover routing was available but unused.

**WHY:** Model diversity is the main lever for catching different blind spots.
Asking first gives the user control over cost vs. depth and lets them
configure routing if needed. Skipping this step silently defaults to a single
model for all reviewers.

**BAD:** "Reviewing the plan now with 3 reviewers..."
**GOOD:** "Before I start, which models should I use for the reviewers?"

**NEVER** — Present raw subagent output without attribution

**SYMPTOM:** The user receives three blocks of undifferentiated text and must guess which reviewer produced which finding. Conflicting opinions cannot be traced to their source.

**CONSEQUENCE:** Actionable findings get mixed with noise because the user cannot weigh each reviewer's credibility. The strategic reviewer's insight carries the same weight as the risk reviewer's false alarm.

**WHY:** The user needs to know which perspective each finding comes from. Raw output without labels loses the multi-reviewer value.

**BAD:** Paste three blocks of text with no headers.
**GOOD:** Organise by reviewer with clear headings and attribution.

**NEVER** — Modify the plan brief differently per reviewer

**SYMPTOM:** Two reviewers disagree on a key question, but the disagreement is caused by different input, not different analysis. The conflict is an artefact, not a signal.

**CONSEQUENCE:** False conflicts waste time in resolution meetings. Real consensus is hidden because each reviewer had different assumptions.

**WHY:** The whole point is that different reviewers reach different conclusions from the same information. Giving each a different brief invalidates the comparison.

**BAD:** Giving the Strategic reviewer extra context about project history.
**GOOD:** Identical brief to all 3; the differences in output come from their lenses.

**NEVER** — Use the same subagent type for all 3 reviewers

**SYMPTOM:** All 3 reviewers converge on the same set of concerns. The "multi-perspective" review effectively becomes a single-perspective review with 3× the cost.

**CONSEQUENCE:** The review provides no more value than a single pass, yet costs 3× in both time and tokens. The user pays for diversity but receives uniformity.

**WHY:** Model diversity is the main lever for catching different blind spots. If all 3 use the same model via the same subagent type, you lose the independent-perspective benefit — they'll converge on similar blind spots.

**BAD:** All 3 use `general`.
**GOOD:** Mix `general` and `explore` (and more if your opencode.json routes them differently).

**NEVER** — Leave a decision-type finding sitting in Open Questions after the review

**SYMPTOM:** The consolidated report lists a real tradeoff (e.g. hard-fail vs. advisory
rollout) as an Open Question, the user reads the report, and the conversation moves on.
Nobody circles back.

**CONSEQUENCE:** The plan ships with the tradeoff unresolved. Implementation starts
against whichever reading of the plan happened to be top of mind, not a decision anyone
actually made — the exact failure mode this skill's Step 10 exists to prevent.

**WHY:** A report is a snapshot, not a commitment. Presenting findings (Step 9) and
resolving them (Step 10) are different steps; skipping straight from one to "the review
is done" drops every decision-type finding on the floor.

**BAD:** Present the consolidated report, thank the user, stop.
**GOOD:** Present the report, then interview the user on every decision-type finding
before treating the plan as amended (Step 10).

## Advanced: Model Routing Configuration

The review's power comes from each subagent having a **different cognitive bias**
— which is best achieved by routing them to **different underlying models** via
`opencode.json`.

### How to configure

In your `opencode.json`, map subagent types to specific models:

```jsonc
{
  "subAgents": {
    "general": {
      "model": "claude-sonnet-5",
      "systemPrompt": "You are a thorough, detail-oriented reviewer."
    },
    "explore": {
      "model": "deepseek-v4-flash",
      "systemPrompt": "You are a skeptical, adversarial reviewer focused on finding flaws."
    }
  }
}
```

### If model routing is not configured

If your opencode setup does not support per-type model routing:
- The reviewer prompts (different lenses, different question sets) still provide
  some diversity of analysis — they ask different questions of the same plan.
- The `explore` subagent type (Risk reviewer) is a different agent implementation
  even on the same model, which may produce different outputs due to different
  tool access and system prompts.
- Make sure to note this limitation when presenting the report.

## Verification

After presenting the consolidated review report, run these checks before
signing off:

1. **Schema compliance** — run the validation script to verify the report output:
   ```bash
   scripts/validate-review-report.sh <path-to-report>
   ```
   This checks the report conforms to the required structure and all sections
   are present. Re-run if validation fails.

2. **Model attribution check** — confirm the report states which models were
   used for Technical/Strategic and for Risk, and whether routing was
   pre-configured or user-selected.

3. **Structural validation captured** — confirm the plan's frontmatter validation
   results and body structure inference are included in the report. The structural
   validation must pass before the review is considered complete.

4. **Actionability review** — verify the report ends with specific Recommended
   Next Actions that reference specific reviewer findings. Each action must be
   concrete and directly address a finding.

5. **Investigation offered** — ask the user if they want to dig deeper on any
   finding. If they accept, run the investigation before concluding.

6. **Every critical/moderate finding is classified and resolved** — confirm each item
   from `critical_issues` and `moderate_concerns` is either applied as an editorial fix
   in the plan file, or answered through the Step 10 interview and recorded in a
   `## Decisions` section. Run `grep -c '^- ' <finding-list>` against the plan's diff if
   unsure whether anything was dropped.

If any of these checks fail, correct the issue before presenting the report as
final.

## Error Handling

| Situation | Response |
|-----------|----------|
| Report validation fails | Check the report against `assets/schemas/review-report.schema.json` for missing or malformed sections; otherwise regenerate |
| A reviewer does not return results | Re-spawn that reviewer individually with the same brief; if a reviewer fails twice, fall back to the other two |
| User has no preference on models | Recommend the best-value pair for their detected environment; stop if they confirm |
| Plan is trivially small | Skip if fewer than 2 H2 sections — do a single pass instead |
| Review reveals a critical issue | Stop if critical and present immediately; ask if the user wants to fix before continuing |

## When a Reviewer Fails

If a subagent reviewer returns an error or times out, do not block the entire
review. Spawn a replacement using the same subagent type and brief. If the
replacement also fails, omit that perspective and proceed with the remaining
reviews. Note the failure in the final report so the user knows a perspective
is missing.

## Templates

This skill follows the repository convention of YAML template + JSON Schema + validation script
for all structured artifacts. See the project rules for the convention.

| Artifact | Template | Schema | Validation Script |
|----------|----------|--------|-------------------|
| Consolidated Review Report | [assets/templates/review-report.yaml](assets/templates/review-report.yaml) | [assets/schemas/review-report.schema.json](assets/schemas/review-report.schema.json) | [scripts/validate-review-report.sh](scripts/validate-review-report.sh) |
| Model Selection Question | [assets/templates/model-selection.yaml](assets/templates/model-selection.yaml) | [assets/schemas/model-selection.schema.json](assets/schemas/model-selection.schema.json) | [scripts/validate-model-selection.sh](scripts/validate-model-selection.sh) |

## References

- [Model Routing Reference](references/model-routing.md) — detailed opencode.json
  configuration examples for model diversity
- `context-file` skill — to persist findings from the review, and to write the plan's
  `## Decisions` section (Step 10)
- `guided-interview` skill — the one-question-at-a-time pattern Step 10 follows when
  resolving decision-type findings
- `design-debate` skill — source of the "always name a concrete revisit trigger" rule
  Step 10 applies to time-boxed decisions
- `session-reflection` skill — for single-agent session-end reflection (complementary use)
- `adr-capture` skill — required, not optional, once a plan gains a `## Decisions`
  section (Step 10) — this repo's pre-commit hook enforces it
