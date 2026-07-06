---
title: "Finding: agent rules formalize repeated scripts, not repeated procedures"
type: finding
status: active
date: 2026-07-06
value: low
related:
  - ../../.agents/RULES.md
  - ../findings/pr-merge-validation-gap-2026-07-06.md
---

# Finding: agent rules formalize repeated scripts, not repeated procedures

> `.agents/RULES.md` has a rule that triggers on a repeated ad hoc *script*, and a rule
> that gates skill *creation* on checking for overlap — but nothing that triggers on a
> repeated multi-step *agent workflow* that was never written down as a script at all.
> That gap is why the validate-and-merge pattern in
> `.context/findings/pr-merge-validation-gap-2026-07-06.md` went unformalized for the
> whole session.

## Summary

Second half of the same investigation as the `session-reflection` blind-spot finding:
even if reflection had asked the right question, no standing rule would have told the
agent what to do with the answer. `.agents/RULES.md` has no directive covering a
repeated multi-step procedure that isn't a script.

## Detail

The two closest existing rules, read precisely:

- **Rule 4 — "Formalise ad hoc scripts after repeated use."** Directive: `ALWAYS use
  .tmp/ for one-off scripts; formalise after 2nd use`. Scope is explicitly a *script
  file* — something written to disk and re-run. The validate-and-merge sequence was
  never a script; it was a sequence of `gh` CLI invocations and manual judgment calls
  typed fresh into the terminal each time. Rule 4 has no surface for this.
- **Rule 12 — "Always check skill overlap before creating new skills."** Directive:
  `ALWAYS scan existing skills' triggers/descriptions for overlap before proposing a
  new one`. This gates *creation*, once someone has already decided to propose a
  skill. It doesn't prompt *noticing* that a skill should be proposed in the first
  place.

Between them: a rule for "you've written the same script twice" and a rule for "before
you create a skill, check for duplicates" — but nothing for "you've now done the same
non-script, multi-step thing three, five, eight times; that's a skill-formalization
signal on its own." This is the same shape of gap Rule 4 already closed for scripts,
just never generalized past scripts.

Concretely, this session's repeated validate-and-merge pattern (poll checks to
terminal, classify advisory vs. real failures, resolve generated-file conflicts via
Rule 15, then merge) never tripped any rule, because it was never written to a file
Rule 4 could see, and nobody had proposed it as a skill yet for Rule 12 to gate.

## Next Steps

Draft a plan (via `rules-management`) for a new rule generalizing Rule 4's principle
from scripts to procedures: a manual, multi-step agent workflow repeated a small
threshold of times (e.g. 2–3) in a session is itself a signal to propose formalizing
it as a skill — independent of whether it was ever captured as a script. Check for
duplicates against Rules 4 and 12 first per `rules-management`'s own workflow; this is
a distinct trigger condition from both, so should not collapse into either by simply
broadening their wording.
