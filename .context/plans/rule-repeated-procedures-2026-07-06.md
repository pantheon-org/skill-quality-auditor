---
title: "Plan: add a rule for repeated multi-step procedures, not just scripts"
type: PLAN
status: DRAFT
date: 2026-07-06
value: MEDIUM
effort: S
related:
  - ../findings/rules-scoped-to-scripts-not-procedures-2026-07-06.md
  - ../../.agents/RULES.md
---

# Plan: add a rule for repeated multi-step procedures, not just scripts

## Goal

Close the gap in `.context/findings/rules-scoped-to-scripts-not-procedures-2026-07-06.md`:
Rule 4 formalizes a repeated *script*; Rule 12 gates skill *creation* once one is
already proposed. Neither rule prompts *noticing* that a repeated, non-script,
multi-step agent workflow (like this session's validate-and-merge pattern) is itself a
formalization signal. Add a rule that does, via the `rules-management` skill.

## Scope

**In scope:**

- Use `rules-management` to check `.agents/RULES.md` for duplicates against Rules 4
  and 12 first (its own required first step), confirming this is a distinct trigger
  condition, not an overlap.
- Append a new rule: a manual, multi-step procedure repeated a small threshold of
  times in a session (2–3, to match Rule 4's "formalise after 2nd use" precedent) is a
  signal to propose formalizing it as a skill — independent of whether it was ever
  written down as a script.
- Mirror the new rule into `docs/development/skills-and-rules.md`'s rule table, per
  the existing convention (every rule in `.agents/RULES.md` has a matching row there —
  confirmed for Rules 15 and 16, added earlier this session).

**Out of scope:**

- Amending Rule 4 or Rule 12's existing wording — the finding recommends a new rule,
  not broadening either existing one, since collapsing "repeated script" and "repeated
  procedure" into one directive would blur two genuinely different trigger conditions
  (one is "does a file exist that's been touched twice," the other is "has a sequence
  of actions with no file happened repeatedly").
- Building tooling to detect repetition automatically — same reasoning as the
  companion `session-reflection` plan: this is a rule an agent applies via judgment,
  not a script that scans logs.

## Phases

### Phase 1 — Draft and append the rule

1. Load `rules-management`; read `.agents/RULES.md` in full; confirm no existing rule
   covers this exact trigger (expected outcome, per the finding's analysis of Rules 4
   and 12 — but re-verify live, not from memory, per `rules-management`'s own
   anti-pattern against blind appending).
2. Draft the rule in the required format: short imperative title, `Directive:` in
   ALWAYS/NEVER phrasing, `Rationale:`.
3. Append to `.agents/RULES.md`.

Exit criterion: `cat .agents/RULES.md` shows the new rule correctly formatted, no
duplicate directive.

### Phase 2 — Mirror and validate

1. Add the matching row to `docs/development/skills-and-rules.md`'s rule table.
2. Run `scripts/check-docs-drift.sh origin/main` to confirm the mirrored doc update
   satisfies the gate (this exact class of gap — a rule added without its
   `skills-and-rules.md` counterpart — is what the `docs-drift` mapping is designed to
   catch).

Exit criterion: rule table and `.agents/RULES.md` agree; docs-drift gate passes.

### Phase 3 — Land

1. Open a PR; confirm `hk check` passes (adr-undocumented, context-index if
   applicable, markdownlint).
2. Merge.

Exit criterion: PR merged, rule active for future sessions.

## Open Questions

- **Exact threshold wording** — "2-3 times" mirrors Rule 4's "after 2nd use," but a
  procedure is fuzzier to count than a script (is a slightly-varied sequence still
  "the same procedure"?). Leaving the threshold as guidance rather than a hard number
  may fit better; `plan-review` or the implementer should weigh in.
- **Does this rule apply retroactively to this session's own validate-and-merge
  pattern**, i.e. should landing this plan also trigger starting
  `.context/plans/pr-merge-skill-2026-07-06.md`'s implementation as the first real
  application of the new rule? Not decided here — sequencing call for whoever
  implements this plan.

## Verification

```bash
.context/plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh .context/plans/rule-repeated-procedures-2026-07-06.md
cat .agents/RULES.md
scripts/check-docs-drift.sh origin/main
hk check
```
