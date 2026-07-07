---
title: "Finding: consensus-rnd degradation.py is not a fit (drift-guard, not a scorer)"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: LOW
themes:
  - SKILL-QUALITY
related:
  - ./skill-validator-vs-native-eval-2026-07-01.md
  - ./tessl-mirror-drift-protection-2026-07-06.md
---

# Finding: consensus-rnd degradation.py is not a fit (drift-guard, not a scorer)

## What was investigated

The file [`skills/consensus-loop/scripts/codex_refactor_loop/checks/degradation.py`](https://github.com/ChronoAIProject/consensus-rnd/blob/dev/skills/consensus-loop/scripts/codex_refactor_loop/checks/degradation.py)
from `ChronoAIProject/consensus-rnd` (607 lines, Python) was reviewed to answer a single
question: would it fit the Go `skill-quality-auditor` codebase, and is it worth importing or
porting? The prompt implied it might be a reusable skill-quality check.

## What the file actually is

It is **not a skill quality scorer.** It is a bespoke architecture-conformance gate hardwired
to one specific skill (`consensus-loop`) in its own source repo. A `SkillDriftChecker` runs
~12 static checks that assert the source repo still conforms to a prior architectural
decision (issue #66 / #419 consensus: "this skill must expose a static check only, never a
standalone degradation runtime"). Concretely it verifies:

- required files exist and forbidden runtime files are absent (`degradation_watchdog.py`,
  `skill_degradation_daemon.py`, and any stray `*degradation*` script)
- `SKILL.md`, CI, and release workflows contain a long list of **hardcoded literal marker
  strings** (`REQUIRED_SKILL_MARKERS`, `REQUIRED_CI_MARKERS`, `REQUIRED_RELEASE_MARKERS`, ...)
- forbidden "expansion surface" regexes are absent (`WorkUnitReplacement`, `ControllerEvent`,
  `auto-fix`, `plugin registry`, ...)
- the checker itself is read-only (it greps its own source for `subprocess.`, `write_text`,
  `unlink(`, ...)
- a bounded temporary host-fixture smoke test passes

It emits frozen `Finding` dataclasses as text or JSON with severities, and no-ops gracefully
(`not-source-repo`) when run outside its home repo. In one sentence: it is a **regression
guard that fails the build if a future edit reintroduces a forbidden surface.**

## Verdict: not a good fit

Do not import or port it. Reasons:

1. **Wrong abstraction.** It scores nothing and produces no grade or dimension. There is no
   D1-D9 analogue because rubric grading is not what it does. It is a self-referential drift
   guard for one skill's internal decisions.
2. **Not generic.** Roughly 95% of the file is literal strings meaningful only to
   `consensus-rnd` (issue #66/#419, `RELEASE_AUTO_ENABLE=false`, clean-room, host-fixture,
   `HOST_GITHUB_RELEASE_REQUIRED_CHECKS`). Strip those and the residue is "does this file
   contain markers X and lack markers Y" — a thin marker-presence linter.
3. **Already covered.** That thin kernel is what [`scorer/d4_specification.go`](../../scorer/d4_specification.go)
   already does generically (it pattern-matches harness paths like `.claude/`, agent
   references, and description quality) and what [`cmd/validate.go`](../../cmd/validate.go)
   does for artifact conventions. Importing it would duplicate D4 + `validate` with a less
   general implementation.
4. **Language and architecture mismatch.** Python vs the Go embedded-assets CLI. Porting is
   only worth it if the *idea* is worth porting, and here it mostly is not.

## The one salvageable idea

The genuinely novel concept is the **per-skill custom regression guard**: a skill records an
architectural decision once, and a config-driven rule set enforces "must contain / must not
contain" on every future edit. The auditor is deliberately generic and has no per-skill
custom-assertion mechanism. This is roughly what `.aislop/rules.yaml` gives aislop for source
code, and it overlaps loosely with `tessl-mirror-drift-protection` (guarding against silent
drift of an authored surface).

If PLG skills ever want project-specific gates, the correct shape is a **config-driven
`required_markers` / `forbidden_markers` rule file consumed by the `validate` command**,
designed natively and generically — not a port of degradation.py's hardcoded literals. This
is deferrable and speculative until a concrete skill needs it, which is why this finding is
graded LOW.

## Recommendation

- **Now:** no code change. Record that the file was evaluated and rejected so the question is
  not re-opened.
- **If a need arises:** draft a small plan for a generic custom-lint-rules surface on
  `validate` (config-driven marker allow/deny lists per skill). Do not resurrect degradation.py.

## Scope note

This finding is analytical input for a maintainer's design decision. It does not itself change
any scorer, gate, or published artifact; any follow-on feature should be planned and reviewed
before implementation.

## Fit assessment (structured record)

<!-- fit-assessment -->
```yaml
schema_version: 1
source:
  name: ChronoAIProject/consensus-rnd (degradation.py)
  url: https://github.com/ChronoAIProject/consensus-rnd/blob/dev/skills/consensus-loop/scripts/codex_refactor_loop/checks/degradation.py
  license: unstated
  language: Python
characterisation: >-
  Not a skill quality scorer. A bespoke architecture-conformance gate hardwired
  to one skill (consensus-loop) in its own repo: ~12 static checks asserting the
  source still conforms to a prior decision (required/forbidden files and
  hardcoded marker strings). Input: its home repo tree. Output: pass/fail
  Findings. For: that repo's CI as a regression guard.
overlap:
  d1_d9_scorers:
    level: full
    note: The thin marker-presence kernel is already covered generically by D4.
  validate_analyze:
    level: full
    note: validate already checks artifact conventions.
  duplication:
    level: none
    note: No similarity detection.
  eval_runner:
    level: none
    note: Scores nothing; not eval machinery.
  helper_skills:
    level: none
    note: No agent-workflow equivalent.
verdict: No fit
vehicle_if_adopted: go-cli
salvageable:
  present: true
  description: >-
    The per-skill custom regression guard concept: a config-driven
    required_markers / forbidden_markers rule file consumed by the validate
    command, designed generically. Deferrable and speculative until a concrete
    skill needs it; never port the hardcoded literals.
recommendation:
  action: record-and-hold
  detail: >-
    No code change now; record that it was evaluated and rejected. If a need
    arises, draft a small plan for a generic custom-lint surface on validate.
value: LOW
```
