---
title: "Known Issue: check-undocumented-decisions.sh false-positives on prose that quotes its own markers"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-07
severity: LOW
value: LOW
themes:
  - GOVERNANCE
related:
  - ../plans/adr-immutability-wording-2026-07-07.md
  - ../findings/adr-immutability-wording-discrepancy-2026-07-07.md
---

# Known Issue: undocumented-decision detector matches its markers inside prose

`check-undocumented-decisions.sh` flags a `.context` file as containing an
undocumented decision by grepping for its heading markers. The match is a plain
substring test — it ignores backticks, code spans, and surrounding context. So any
`.context` doc that merely *documents or quotes* those markers (for example, a plan or
finding explaining how the gate itself works) trips the gate and fails `hk`, even
though it contains no actual decision.

> Verified: drafting `adr-immutability-wording-2026-07-07.md` failed the pre-commit
> `adr-undocumented` step purely because a Risks bullet quoted the marker strings while
> explaining the gate. Rewording the bullet to describe the markers without reproducing
> them verbatim cleared it.

## Impact if unfixed

Authors documenting the governance tooling must remember not to write the literal
marker strings, or they get a confusing gate failure unrelated to any real decision.
It is a nuisance, not a correctness risk — the workaround (paraphrase the markers) is
trivial once known, which is why this is LOW severity.

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

Make the detector ignore fenced code blocks and inline code spans before matching, or
require the marker at the start of a line as a real heading (anchor the regex to
`^#{2,3} ` for the heading forms) rather than matching anywhere in the line. Either
narrows the match to genuine headings and stops the prose false-positive.
