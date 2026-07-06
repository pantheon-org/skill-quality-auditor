---
title: "ADR-046: known-issue becomes a first-class .context/ file type, driven by session-reflection"
status: accepted
date: 2026-07-06
context:
  - path: ".context/known-issues/docs-drift-jq-hard-dependency-2026-07-06.md"
  - path: ".context/known-issues/docs-drift-cumulative-not-enforced-2026-07-06.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

A `session-reflection` run (an independent sub-agent reviewing a session's work) surfaced several verified, concrete gaps — most notably that `check-docs-drift.sh`'s new `jq`-based sidecar lookup hard-fails a contributor's `pre-push` hook with an unhelpful exit 127 if `jq` isn't on `PATH`, confirmed live by actually removing `jq` from `PATH` and re-running the script. Items like this, surfaced during reflection but not fixed in the same session, had no durable home: the `session-reflection` skill's existing "Follow up" step said only "if a finding warrants preservation, create a new finding entry" — findings are for research/investigation output, not naturally for "this is a real bug, deliberately not fixed right now, don't forget it." Without a dedicated place, verified gaps like this tend to evaporate into chat scrollback once a session ends.

## Decision

1. **`known-issue` is added as a sixth `.context/` file type**, alongside `plan`, `finding`, `analysis`, `instruction`, `audit` — living under `.context/known-issues/`, one file per issue (matching the existing one-file-per-topic convention for findings, not a single running checklist). Added to `context-frontmatter.schema.json`'s `type` enum and `regenerate-context-index.sh`'s type-grouping tables.
2. **A `severity` field (`critical | high | medium | low`) is required for `type: known-issue`**, enforced by `validate-context-frontmatter.sh` the same way `effort` is required for draft/active plans. The `Known Issues` index section is sorted by severity (critical first) so the highest-urgency items surface without opening every file.
3. **The `Known Issues` section is placed first in `.context/index.yaml`'s output**, ahead of Plans — this list is meant to be the first thing a reader scans, not buried after other sections.
4. **`status` reuses the existing `draft | active | done | superseded` enum** rather than inventing new values — `active` means still open, `done` means fixed. No new status vocabulary was needed.
5. **The `session-reflection` skill's "Follow up" step is the primary, designated source of `known-issue` entries.** Its `SKILL.md` now explicitly instructs: a reflection item that's a verified, concrete gap NOT being fixed in the current session must become a `.context/known-issues/<topic>-YYYY-MM-DD.md` entry via `context-file`, rather than only being discussed and left in chat. Not creating one for something about to be fixed immediately — fix it instead; `known-issue` is for consciously deferred work only.
6. **Two known-issue entries are seeded from the reflection that motivated this decision**, both stemming from the reviewed-baseline docs-drift work (ADR-045, not yet merged at time of writing): the `jq` hard-dependency bug (critical) and cumulative-mode's CI-visible-but-not-enforced gap (high). Lower-signal reflection items (an untested-in-Actions assumption, a manual-only regression test) were deliberately not promoted to known-issue status — the list is for genuinely critical/high-severity, verified gaps, not every minor observation, to keep its signal meaningful.

## Consequences

- **Easier:** a verified-but-deferred gap now has one obvious, git-tracked, indexed home instead of relying on someone remembering a chat message or re-reading old PR descriptions.
- **Easier:** `.context/index.yaml`'s severity-sorted `Known Issues` section gives an at-a-glance "what's the most urgent debt right now" view without opening every file, the same way `effort` does for plans.
- **Harder:** a sixth file type is one more thing to remember when authoring or reviewing `.context/` files — mitigated by reusing the existing frontmatter schema/status vocabulary rather than inventing new mechanics, so the marginal complexity is one new type name and one new field.
- **Binding for future work:** any future edit to the `session-reflection` skill's "Follow up" step must preserve the known-issue creation instruction — removing it would silently regress reflection findings back to evaporating in chat, the exact problem this ADR fixes.
