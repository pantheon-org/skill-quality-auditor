---
title: "Finding: .context carries no value signal, so \"highest value\" is re-derived by ad-hoc judgement every time"
type: FINDING
status: ACTIVE
date: 2026-07-06
value: HIGH
related:
  - ../index.yaml
  - ../../.agents/skills/context-index/scripts/regenerate-context-index.sh
  - ../../.agents/skills/context-file/assets/schemas/context-frontmatter.schema.json
  - ../instructions/ways-of-working.md
---

# Finding: .context carries no value signal, so "highest value" is re-derived by ad-hoc judgement every time

> Asked "what's the highest-value draft plan or finding to address?", an agent cannot
> answer from `.context/index.yaml` or from any file's frontmatter. There is no field
> that encodes value, impact, or priority. The answer is re-derived from out-of-band
> reasoning (recency, session memory of recurring pain, `related` graphs, judgement)
> every single time, which is non-reproducible and unauditable. The inability to read
> the ranking off the index is itself the signal that the schema is missing the field.

## Summary

The user asked for the highest-value item among 7 draft plans and 36 active entries.
The agent produced a ranking, but not from any stored signal: it weighed recency, a
recurring manual pain point recalled from session memory, and severity borrowed from a
different entry type. When challenged, the honest position is that `.context/` gives no
clear signal for "highest value" — and closing that gap is itself the highest-value
add, because "what should I do next?" is the single most frequently repeated decision
made against `.context/`, and it currently has zero data backing.

## Detail

### What the index actually promotes

`grep -oE '^    [a-z_-]+:' .context/index.yaml | sort | uniq -c` over 100 entries:

| Field | Count | Meaning |
| ----- | ----- | ------- |
| `title` | 100 | label |
| `status` | 100 | lifecycle (active / done / draft) |
| `date` | 100 | creation date |
| `related` | 49 | cross-links |
| `effort` | 13 | **cost** signal, plans only |
| `severity` | 5 | **risk** signal, known-issues only |

### What is absent

- **No `value`, `impact`, `priority`, `roi`, or `score` field at top level** on any
  plan, finding, or known-issue. A repo-wide grep for those keys in frontmatter returns
  only `effort` (13 hits) — nothing that expresses benefit.
- `priority` *does* exist, but only buried inside some remediation-plan bodies
  (`executive_summary.priority` and per-phase `priority:` under `remediation_phases`).
  It is never lifted into top-level frontmatter and never reaches the index, so it is
  invisible to any "what's next?" query.

### The two signals that do exist are cost and risk, not value, and both are siloed

- `effort` is a **cost** axis, present on only 13 of ~36 plans, and at least one value
  is the placeholder `TBD` (`migrate-off-tessl-eval`). Cost alone cannot rank value:
  a cheap low-benefit chore outranks nothing.
- `severity` is a **risk** axis, present on only the 5 known-issues. It cannot rank
  draft *plans* at all, because plans do not carry it.
- There is no axis shared across all three `.context/` types, so the types cannot be
  compared against one another on a common scale. "Fix the high-severity known-issue"
  vs "ship the M-effort pr-merge skill" is an apples-to-oranges comparison the schema
  offers no way to resolve.

### Why the ranking today is unreliable

To answer "highest value" an agent must: open plan bodies, recall session memory for
recurring pain (e.g. the validate-and-merge sequence run 8+ times), walk `related`
graphs, and apply judgement. None of that is stored, so:

- **Non-reproducible** — a fresh session with no memory reaches a different ranking.
- **Unauditable** — the reasoning is not written down anywhere a human can check.
- **Expensive** — every "what's next?" question re-pays the full reasoning cost.

### Why this is the highest-value fix

Prioritisation is the most repeated decision made against `.context/` — every session
that opens the index asks some form of "what should I do next?". A stored value signal
turns that from bespoke reasoning into a query, and the leverage compounds across every
future session and every agent. No other single `.context/` change touches as many
future decisions.

## Next Steps

Draft a plan (and, since it changes the frontmatter contract, an ADR) to add a
prioritisation signal to `.context/`. Options to weigh in that plan, not decide here:

1. A top-level `value` or `impact` enum (e.g. high / medium / low) on plans and
   findings, promoted into `index.yaml` by the context-index generator.
2. A derived priority = value vs `effort` (a lightweight ROI), which reuses the
   `effort` field already present on plans.
3. A single shared prioritisation axis across plan / finding / known-issue so the three
   types become mutually comparable, rather than the current split of `effort` (plans)
   and `severity` (known-issues).

Whichever is chosen, sub-issues to fold in: `effort` coverage is partial (13/36 plans)
and contains a `TBD`, and `priority` is stranded inside remediation-plan bodies where
the index never sees it. Note that creating this finding is likely to trip the
`check-undocumented-decisions` gate on push; the schema decision belongs in the plan and
ADR, so the gate is expected and correct here.
