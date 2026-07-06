---
title: "Theme Vocabulary and Tagging Rules"
type: INSTRUCTION
status: ACTIVE
date: 2026-07-06
related:
  - ../plans/context-theme-taxonomy-2026-07-06.md
  - ../findings/context-taxonomy-gap-2026-07-06.md
  - ../instructions/value-rubric.md
---

The `themes` frontmatter field records **what area of the system** a `.context/`
action-candidate is about. It is the subject axis, orthogonal to the three
magnitude axes (`value`, `effort`, `severity`): those say how big or urgent an
item is, `themes` says what it touches.

`themes` applies to the three action-candidate types: `PLAN`, `FINDING`, and
`KNOWN_ISSUE`. It does not apply to `ANALYSIS`, `INSTRUCTION`, or `AUDIT`, which
are reference material rather than things to do next.

## Shape

`themes` is a **multi-valued, ordered list** — an entry can genuinely belong to
several areas, so it is not a single enum. Every member is drawn from the
controlled vocabulary below; free-form text is not permitted, so the axis stays
queryable.

```yaml
themes:
  - EVAL
  - GOVERNANCE
```

The list is **ordered, and the first entry (`themes[0]`) is the primary theme.**
The primary answers "what is this mainly about?" and is the only member that
participates in the read-protocol tie-break (see below). The remaining members
are for filtering and cluster views, never for ordering. Authors write the list
primary-first.

## Controlled vocabulary

Six themes, kept deliberately coarse. A too-fine vocabulary is as useless as none.

| Theme | Covers |
| ----- | ------ |
| `EVAL` | Skill evaluation and scoring: the native eval runner, LLM-judge, eval gating, migration off Tessl eval, self-evolution loops. |
| `PR-TOOLING` | Pull-request and CI automation: PR-Agent, agent-scan, Plumber, the pr-author/pr-merge workflow, GitHub Action packaging. |
| `DOCS` | Documentation site and drift: docmd build, GH Pages, check-docs-drift, reviewed baselines. |
| `GOVERNANCE` | The `.context/` system itself and its integrity: the index, frontmatter contract, ADR capture, status-sync, prioritisation and theme signals, cross-reference drift. |
| `SKILL-QUALITY` | The scoring framework and agent behaviour: D1-D9 scorers, scoring-pattern config, agent rules, session-reflection, skill remediation. |
| `DISTRIBUTION` | Getting the tool to users: Homebrew tap, release packaging, binary distribution. |

## Split-on-evidence rule

The vocabulary ships coarse and is refined only on observed need, not
speculatively. If, after backfill, a single theme carries a disproportionate
share of entries (rough guide: more than ~30% of active/draft action-candidates,
with `GOVERNANCE` the likely first candidate), split it into finer themes. Any
split is recorded as an amendment to the ADR that ratified this vocabulary, not
an ad-hoc edit. This mirrors how the `value` signal deferred a numeric scale
until within-bucket ties proved to block the sort: ship the simple thing, refine
on evidence.

## Choosing the primary theme — worked examples

The primary is the area the item most changes, not merely one it touches.

- **`plans/context-theme-taxonomy-2026-07-06.md` → `[GOVERNANCE]`.** It changes
  the `.context/` frontmatter contract and index; that is squarely governance.
  No second theme is needed.
- **`plans/migrate-off-tessl-eval-2026-06-29.md` → `[EVAL]`.** Purely about the
  evaluation pipeline. Single theme.
- **`findings/cross-reference-status-drift-2026-07-04.md` → `[GOVERNANCE]`.**
  Even though it is realised through PR and merge flows, what it is *about* is
  the integrity of the `.context/` cross-reference graph, so governance leads.
- **`plans/agent-scan-integration-2026-07-04.md` → `[PR-TOOLING, SKILL-QUALITY]`.**
  Primarily a CI/PR security check (`PR-TOOLING`), but it scans skill assets, so
  `SKILL-QUALITY` is a genuine secondary — the primary is still the CI mechanism.

## Read protocol interaction

`themes[0]` is the final tie-breaker in the "what's next" sort, below `value`
then `effort`:

1. Filter to `DRAFT`/`ACTIVE` `PLAN`/`FINDING`/`KNOWN_ISSUE`.
2. Sort by `value` descending (`HIGH` > `MEDIUM` > `LOW`).
3. Then `effort` ascending (`S` < `M` < `L` < `TBD`) where present.
4. Then, only to break a remaining tie, prefer the item whose `themes[0]` matches
   the area already in focus. Theme expresses preference-of-area, not priority,
   which is why it sits below both magnitude axes.

`themes` is also a filter/slice dimension in its own right: "show me all `EVAL`
work" or "which theme carries the most open debt" are queries the index answers
by reading the field directly, independent of the sort.
