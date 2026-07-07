---
name: external-source-fit
description: "Assess whether an external GitHub repo, file, or project fits the skill-quality-auditor project: what it actually is (not what its name implies), whether it overlaps existing capability (D1-D9 scorers, validate, duplication, native eval runner, helper skills), whether it belongs as a Go CLI change or a helper skill, what is worth learning even from a rejected source, and whether the project needs it. Ends by writing a house-standard FINDING and flagging any binding decision for ADR capture. Use when someone links an external repo or file and asks whether it fits, whether we can learn from it, or whether we need it. Do NOT use for reviewing internal code (use review-changes), scoring a SKILL.md (use skill-quality-auditor), or web research (use deep-research). Triggers: 'does this repo fit', 'evaluate this project', 'assess this external repo', 'would this fit our project', 'can we learn from', 'is this worth adopting', 'fit assessment', 'should we import', 'should we port this'."
---

# External Source Fit Assessment

Given a link to an external GitHub repo, file, or project, decide whether it fits this project and write the answer up as a finding. The recurring trap is answering from the name or the asker's framing ("it's a skill-eval tool, so it must be relevant") instead of from what the source actually does. Most sources are a partial or no fit; the value is in naming the one transferable idea precisely and in not re-opening the question later.

This skill is **project-aware**: every verdict is rendered through the skill-quality-auditor lens, not in the abstract.

## Prerequisites

- Read access to the source. Public repos: `gh repo clone` (shallow) or fetch raw files with WebFetch / `curl` into the scratchpad. Private repos: use `gh` with the caller's auth. Never paste secrets from the source into the conversation.
- The `context-file` skill, to write the finding with valid frontmatter (`value` + `themes` are required for FINDING type while DRAFT/ACTIVE).
- The `adr-capture` skill, only if the assessment produces a binding decision (e.g. "we will not adopt approach X").
- Optional grounding: `./dist/skill-auditor` and the dimension references under `cmd/assets/references/` to check overlap against D1-D9 precisely.
- Read this project's `CLAUDE.md` first if unfamiliar: the repo has a Go CLI (`scorer/`, `cmd/`, `duplication/`, `reporter/`) and separate helper skills under `.context/plugins/`.

## Quick Start

```bash
# 1. Acquire only what you need (a few key files beats cloning everything)
gh repo clone <owner>/<repo> "$SCRATCH/<repo>" -- --depth=1
# or, for a single file:
curl -sL "https://raw.githubusercontent.com/<owner>/<repo>/<ref>/<path>" -o "$SCRATCH/<file>"

# 2. Read the source, characterise it, map it against this project, render a verdict.
# 3. Write the finding via the context-file skill, then regenerate the index.
```

## Workflow

### 1. Acquire the source (cheaply)

Pull only what answers the question. For a single file, fetch that file. For a repo, read the README, the entry points, and the one or two modules that embody the claim. Reading the whole repo is rarely necessary and burns context. Note the **license** and whether the source is generic or hardwired to its own project.

### 2. Characterise what it actually is

State, in one or two sentences, what the source does mechanically, independent of its name or the asker's framing. This is the step that catches the common error: a file whose name and location imply one purpose (say, a quality scorer) can turn out to do something entirely different (say, enforce one project's own architecture). Ask: what is the input, what is the output, and who is it for? If the source is mostly hardcoded literals meaningful only to its home project, say so now.

### 3. Map it against this project

Check overlap against existing capability before judging novelty. The relevant surfaces:

| Existing capability | Ask |
| --- | --- |
| D1-D9 scorers (`scorer/`) | Does this measure something a dimension already covers? |
| `validate` / `analyze` (`cmd/`) | Is this artifact/structure validation we already do? |
| duplication engine (`duplication/`) | Is this similarity detection we already have? |
| native eval runner (`cmd/eval.go`, D9) | Is this eval/judge machinery we already own? |
| helper skills (`.context/plugins/`) | Does an agent-workflow skill already do this? |

Then decide the **vehicle if adopted**: a deterministic, generic, offline computation belongs in the Go CLI; an agentic, prose-producing, judgement-heavy workflow belongs in a helper skill. Language and architecture mismatch (e.g. Python vs the Go embedded-assets CLI) is a cost, not a blocker, but it only matters if the idea is worth porting.

### 4. Render the verdict

Use the fit rubric (three bands: **Good fit / Partial fit / No fit**) in [references/fit-rubric.md](references/fit-rubric.md). Keep novelty and need separate: a technique can be genuinely novel and still not needed, and a source can be a poor fit overall yet contain one idea worth keeping.

### 5. Extract the salvageable idea, separately

Even a "no fit" verdict should name what, if anything, is worth learning, and how it would be built **natively** rather than imported. Resist porting hardcoded literals; extract the generic kernel. If there is genuinely nothing transferable, say that explicitly rather than manufacturing a takeaway.

### 6. Write it up as a finding

Use the `context-file` skill to create a FINDING under `.context/findings/` following the reusable spine in the rubric reference: *What was investigated → What it actually is → Verdict → The salvageable idea → Recommendation*. Grade `value` against `.context/instructions/value-rubric.md` and `themes` against `.context/instructions/theme-vocabulary.md` (a pure investigation-and-reject is usually `LOW`; a "we should build this" is `MEDIUM`+). Regenerate the context index afterwards.

If the finding contains a binding decision, invoke `adr-capture`. Otherwise avoid heading text that starts with `## Decision`, which the `adr-undocumented` pre-commit hook reads as an undocumented decision.

### 7. Verify and report

Before reporting, verify the write-up holds together:

- Run `check-context-frontmatter.sh` on the new finding and confirm the context index regenerates with no stderr warnings (the finding must appear in `index.yaml`).
- Confirm the verdict is exactly one of the three bands, and that the finding either names a concrete salvageable idea or states explicitly that there is none.
- Confirm no heading begins with `## Decision` unless an ADR was captured (otherwise the pre-commit hook fails).

Then report to the user: lead with the one-word verdict and the reason, then the salvageable idea and recommendation. Link the finding file. Do not bury the answer under a file tour.

## Fit verdict bands (summary)

| Band | Meaning | Typical action |
| --- | --- | --- |
| **Good fit** | Fills a real gap, generic enough to adopt, low overlap with existing capability | Draft a plan to build it (natively) |
| **Partial fit** | One transferable idea inside an otherwise ill-fitting source | Record the idea; build only the kernel if/when needed |
| **No fit** | Wrong abstraction, project-specific, or already covered | Record the rejection so it is not re-opened |

Full criteria and the finding spine: [references/fit-rubric.md](references/fit-rubric.md).

## Anti-patterns

**NEVER** judge the source from its name or the asker's framing.
**WHY:** Names mislead. A check named for skill "degradation" can be an architecture drift guard, not a scorer.
**BAD:** "It's in a skills folder and scores things, so it's a quality scorer we could reuse."
**GOOD:** Read the code, state what its input and output actually are, then judge.

**NEVER** propose importing or porting a source before checking overlap with existing capability.
**WHY:** Most "novel" ideas are already covered by D1-D9, `validate`, or the duplication engine, so the import would duplicate with a less general implementation.
**BAD:** Porting a marker-presence linter that D4 already subsumes.
**GOOD:** Map against the capability table first; adopt only the residue that is genuinely new.

**NEVER** port a source's hardcoded, project-specific literals.
**WHY:** They are meaningful only to the source's home project; the transferable value is the generic kernel.
**BAD:** Copying another repo's issue numbers, marker strings, and env-var names into ours.
**GOOD:** Extract the generic mechanism and redesign it config-driven and generic.

**NEVER** conflate "novel" with "needed".
**WHY:** A technique can be clever and still solve a problem this project does not have.
**BAD:** Grading a finding HIGH because the technique is interesting.
**GOOD:** Grade against the value rubric on leverage and gap-closure, not on novelty.

## Illustrative verdicts

Generic shapes each band tends to take (not tied to any one source):

- **No fit:** a source that reads like a quality scorer but is really a project-specific drift guard, mostly hardcoded literals, whose generic kernel is already covered by an existing scorer or the `validate` command. Record the rejection; extract at most one config-driven idea, built natively.
- **Partial fit:** an eval framework whose overall design does not match this project, but which contributes one measurement technique worth adopting.
- **Good fit:** a source that cleanly fills a gap no dimension covers, generic enough to adopt with proportionate effort.

Prior completed assessments are kept as findings in the findings directory. Read the two or three most recent before writing a new one, to match the house tone and depth rather than copying any single example.

## References

| Topic | Reference | When to use |
| --- | --- | --- |
| Verdict bands, decision criteria, and the finding write-up spine | [references/fit-rubric.md](references/fit-rubric.md) | Rendering the verdict and structuring the finding |
| This project's capability surfaces (scorers, commands, engines) | Repo root `CLAUDE.md` | Step 3, mapping overlap |
| Value grading and theme tagging for the finding | `.context/instructions/value-rubric.md`, `.context/instructions/theme-vocabulary.md` | Step 6, grading the finding |
