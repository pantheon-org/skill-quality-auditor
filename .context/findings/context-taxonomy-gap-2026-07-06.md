---
title: "Finding: .context has no thematic axis, so within-tier items are interchangeable and \"which area\" is un-queryable"
type: FINDING
status: DONE
date: 2026-07-06
value: MEDIUM
themes:
  - GOVERNANCE
related:
  - ../index.yaml
  - ../instructions/value-rubric.md
  - ../plans/context-prioritisation-signal-2026-07-06.md
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
---

# Finding: .context has no thematic axis, so within-tier items are interchangeable and "which area" is un-queryable

> With the `value` signal shipped, the read protocol can now answer "what is the
> single highest-value item?". But it cannot answer "which area carries the most
> open debt?", "show me all the eval-migration work", or break a tie between two
> items of equal `value` and `effort`. Nothing in the frontmatter or the index
> encodes *what part of the system an item is about*. That subject axis is
> re-derived by reading titles and walking `related` graphs by hand, every time.

## Summary

The frontmatter contract carries `title`, `type`, `status`, `date`, `related`,
`effort`, `severity`, and `value`. Every one of those is either structural
(`type`, `status`, `date`), a cross-link (`related`), or a magnitude
(`effort`, `severity`, `value`). None of them says what an entry is *about*.

The `value` signal answers "how much good does this do?" and orders the queue by
magnitude. But it deliberately stops there: within a single `value`+`effort`
bucket, items are interchangeable, and the protocol says to act on the top item
"without re-forming an independent judgement". In practice a person choosing among
three equal-tier items still wants to pick by area of the system, and the schema
gives them nothing to pick on. There is a latent taxonomy in the data (titles and
`related` links cluster the open work into recognisable themes), but it is not
stored, not enumerated, and not surfaced in the index.

## Detail

### The three magnitude axes measure size, never subject

| Axis | Question | Field |
| ---- | -------- | ----- |
| Benefit-of-action | How much good does doing this do? | `value` |
| Cost-of-action | How much work is it? | `effort` (plans only) |
| Risk-of-inaction | How bad is leaving it undone? | `severity` (known-issues only) |

There is no fourth axis for *domain / theme / area*. So the index can rank the
whole backlog on one scale but cannot slice it: "everything about eval migration",
"all the docs-drift entries", "the governance cluster" are all queries the index
cannot serve.

### A latent taxonomy already exists, un-stored

The open DRAFT/ACTIVE action-candidates cluster into recognisable themes purely
from their titles and `related` links:

- **Eval migration** — `migrate-off-tessl-eval`, `evoskill-core-loop-port`,
  `eval-gating-byok`, `skill-quality-llm-judge-failure-chain`
- **PR tooling / CI automation** — `pr-agent-integration`, `pr-merge-skill`,
  `agent-scan-integration`, `plumber-cicd-security`, `github-action-packaging`
- **Docs-drift** — the three `docs-drift-*` entries, tightly interrelated
- **Governance / context-system integrity** — `cross-reference-drift-audit`,
  `automate-post-merge-status-sync`, the four ADR/known-issue-enforcement entries,
  `value-historical-corpus-learnings`
- **Agent rules / skill quality** — `rule-repeated-procedures`,
  `session-reflection-procedural-repetition`, `scoring-pattern-config-review`
- **Distribution** — `homebrew-tap-pending`

This clustering is real and useful, but it lives only in an agent's head each
session. It is re-derived, non-reproducible, and invisible to any query.

### Why it matters

- **Tie-breaking.** The read protocol resolves `value` then `effort`. Below that
  it is silent. A stored theme lets a person or a future "what's next" skill break
  a same-tier tie by "the area I am already in" instead of arbitrarily.
- **Batch closure.** Interrelated clusters (the three docs-drift entries) are
  cheaper to close together than piecemeal. Without a theme axis you cannot see
  the cluster to plan the batch.
- **Debt visibility.** "Which theme has the most open debt?" is a planning
  question the index cannot answer. Governance is currently the largest open
  cluster, but only a manual count reveals that.

### Why this is MEDIUM, not HIGH

Unlike the `value` gap, this does not close a decision that is made every session
with zero backing; `value` already carries the primary "what's next" call. A theme
axis is a refinement: it improves grouping and tie-breaking and feeds the future
"what's next" skill, but nothing is blocked on it and the queue still functions
without it. Clear standalone benefit, limited leverage.

## Next Steps

Draft a plan (and, since it changes the frontmatter contract, an ADR) to add a
thematic axis to `.context/`. The central decision to weigh in that plan, not here,
is the shape of the vocabulary:

1. A single `theme` enum (one theme per entry) with a small controlled vocabulary
   ratified up front, mirroring how `value`/`severity` are enums.
2. Multi-valued `tags` (an entry can belong to several themes), which fits items
   that genuinely span areas but complicates sort/tie-break semantics.
3. Deriving the grouping from the existing `related` graph instead of a new field,
   avoiding a schema change at the cost of accuracy and a manual link discipline.

The migration path is already proven by the `value` signal: schema field,
optional-first validator, index emission, authoring-skill prompts, a single
serialised backfill of active/draft entries, and an ADR. Note that creating this
finding may trip the `check-undocumented-decisions` gate on push; the vocabulary
decision belongs in the plan and ADR, so the gate is expected and correct here.
