# Fit rubric and finding spine

Detailed reference for the `external-source-fit` skill. Load when rendering a verdict
or structuring the finding.

## The three verdict bands

### Good fit

All of:

- It fills a **real gap** in this project (a capability D1-D9, `validate`, `analyze`,
  the duplication engine, the native eval runner, or an existing helper skill does not cover).
- It is **generic enough** to adopt, or its generic kernel is cleanly separable from
  project-specific glue.
- The **vehicle is clear**: a deterministic offline computation maps to the Go CLI; an
  agentic prose-producing workflow maps to a helper skill.
- The **cost is proportionate**: porting effort or a new dependency is justified by the gap it closes.

Action: draft a `.context/plans/` plan to build it natively. Grade the finding `MEDIUM`+.

### Partial fit

The source as a whole does not belong, but it contains **one transferable idea**:

- A technique, data model, or heuristic worth extracting, even though the surrounding
  implementation is ill-suited (wrong language, wrong abstraction, project-specific).
- The idea is recorded, but built only if and when a concrete need appears.

Action: record the idea and the "build natively" shape. Grade by the idea's leverage,
usually `LOW` to `MEDIUM`.

### No fit

Any of:

- **Wrong abstraction**: it does not do what the project does (e.g. it enforces one
  project's architecture rather than scoring skill quality).
- **Project-specific**: mostly hardcoded literals meaningful only to its home repo.
- **Already covered**: its mechanism duplicates an existing capability, with no more-general residue.

Action: record the rejection so the question is not re-opened. Grade `LOW`. Still name any
salvageable idea, or state explicitly that there is none.

## Keep these axes separate

- **Novel vs needed.** Novelty is about the technique; need is about this project's gaps.
  A source can be highly novel and still `No fit` because the project has no such gap.
- **Overlap vs quality.** Check overlap with existing capability *before* judging how good
  the source is. A brilliant implementation of something already owned is still a duplication.
- **Idea vs implementation.** The transferable value is almost always the generic kernel,
  never the source's project-specific literals, glue, or language-specific plumbing.

## Overlap checklist (this project)

Run through each before concluding "novel":

- `scorer/d1_*.go` ... `d9_*.go` — does a dimension already measure this?
- `cmd/validate.go`, `cmd/validate_context.go` — is this artifact/frontmatter validation?
- `cmd/analyze.go` — is this static skill analysis?
- `duplication/` — is this similarity / overlap detection?
- `cmd/eval.go` + D9 — is this eval scenario running or LLM-judge machinery?
- `reporter/` — is this output formatting / persistence?
- `.context/plugins/**` — does a helper skill already provide this workflow?
- Existing findings in `.context/findings/` — has a similar source already been assessed?

## Finding write-up spine

Structure every fit finding this way (matches the house findings style):

```markdown
---
title: "Finding: <source> is a <fit band> (<one-line why>)"
type: FINDING
status: ACTIVE
date: YYYY-MM-DD
value: LOW | MEDIUM | HIGH
themes:
  - <primary>        # usually SKILL-QUALITY or EVAL
related:
  - ./<related-finding>.md
---

# Finding: ...

## What was investigated
Link the source. State the exact question asked.

## What it actually is
One or two sentences, mechanical, independent of the name/framing.

## Verdict: <Good fit | Partial fit | No fit>
The reasoning, ideally as a short comparison table (source vs this project).

## The salvageable idea
The generic kernel worth keeping and how it would be built natively.
State "nothing transferable" explicitly if that is the honest answer.

## Recommendation
Now: what to do (often "no code change; record the rejection").
Later: the trigger and shape if a need arises.
```

## Grading pointers

- Pure investigate-and-reject with no follow-on work: `value: LOW`.
- "We should build this small thing eventually": `value: LOW`-`MEDIUM`.
- "This closes a recurring gap several things depend on": `value: HIGH`, and draft a plan.
- Primary theme is usually `SKILL-QUALITY` (framework/behaviour) or `EVAL` (scoring/judge/runner).
  Grade against `.context/instructions/theme-vocabulary.md` and `value-rubric.md`.

## Hook gotcha

The `adr-undocumented` pre-commit hook flags any heading beginning `## Decision` as an
undocumented decision. A fit finding is analytical input, not a binding decision, so avoid
that heading text (use `## Recommendation` or `## Scope note`). If the finding genuinely
records a binding decision, run the `adr-capture` skill instead of renaming.
