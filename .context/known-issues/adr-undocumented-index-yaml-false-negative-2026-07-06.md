---
title: "Known issue: check-undocumented-decisions.sh silently skips any file mentioning 'index.yaml'"
type: known-issue
status: active
date: 2026-07-06
value: medium
severity: medium
related:
  - ../plugins/pantheon-org/governance/adr-capture/scripts/check-undocumented-decisions.sh
  - ../findings/pr-merge-validation-gap-2026-07-06.md
---

# Known issue: check-undocumented-decisions.sh silently skips any file mentioning 'index.yaml'

## What

`check-undocumented-decisions.sh` (the script behind the `adr-undocumented` pre-commit
check) has this line:

```python
# skip index.yaml reference files
content = md_file.read_text()
if "index.yaml" in content:
    continue
```

The comment's intent is presumably "don't flag files that are themselves *about*
`.context/index.yaml` or `docs/ADR/index.yaml` generation" — but the check is a bare
substring match against the file's entire body, not a check on the file's purpose or
path. Any `.context/**/*.md` file that mentions the string `index.yaml` *anywhere*, for
any reason, silently skips decision-indicator scanning entirely — even if it also
contains a genuine, undocumented decision.

## How this was found

While committing `.context/findings/pr-merge-validation-gap-2026-07-06.md` (on
`feat/pr-merge-skill-gap`... actually `docs/pr-merge-skill-gap`), the `adr-undocumented`
check passed cleanly despite the finding having a `## Recommended Action` heading — one
of `DECISION_KEYWORDS`. Investigation found the finding's Detail section mentions
`.context/index.yaml` and `docs/ADR/index.yaml` in an unrelated sentence about the
regenerate-don't-hand-merge conflict-resolution pattern (Rule 15), which tripped the
skip condition by coincidence. Two sibling findings drafted immediately after
(`session-reflection-procedural-repetition-blind-spot-2026-07-06.md`,
`rules-scoped-to-scripts-not-procedures-2026-07-06.md`) had genuinely equivalent
`## Recommended Action` content but did *not* mention `index.yaml`, and were correctly
flagged — confirming the skip, not the keyword match, was the difference.

## Why not fixed now

Fixing this properly means replacing the substring check with something that actually
identifies "this file is about index.yaml generation" (e.g. a path-based allowlist, or
requiring the mention to be in a code block referring to the file itself, or dropping
the skip condition entirely and re-auditing what it was meant to protect against — its
original purpose isn't documented anywhere, including in `adr-capture`'s own SKILL.md
or references). That requires archaeology into git blame / the original PR that added
the line, which is out of scope for the finding that surfaced this as a side discovery.

## Impact

Low urgency, not zero: a finding with a real, undocumented decision escapes the
pre-commit gate entirely if it happens to mention `index.yaml` for any unrelated
reason — a false negative in an enforcement mechanism whose whole job is catching
exactly that case. The two findings that surfaced this were manually caught and
corrected (heading renamed from `## Recommended Action` to `## Next Steps`, since
neither represented an actually-resolved decision) — but that was investigator
diligence, not something the check itself would have caught if the coincidence had
gone the other way.

## Revisit trigger

Revisit when `adr-capture` next gets a maintenance pass, or immediately if a real
undocumented decision is later found to have escaped review via this exact skip
condition (i.e. a merged `.context/` file with a genuine decision, mentioning
`index.yaml`, never captured as an ADR).
