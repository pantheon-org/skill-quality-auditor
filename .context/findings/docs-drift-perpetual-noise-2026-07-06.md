---
title: "Finding: check-docs-drift.sh's cumulative mode can never clear a reviewed-but-unedited doc"
type: FINDING
status: ACTIVE
date: 2026-07-06
value: MEDIUM
related:
  - ../../scripts/check-docs-drift.sh
  - ../../docs/ADR/adr-044-docs-drift-pr-gate.md
  - gh-pages-docs-drift-2026-07-05.md
---
# Finding: check-docs-drift.sh's cumulative mode can never clear a reviewed-but-unedited doc

> While closing out the docs content gaps found in `.context/findings/gh-pages-docs-drift-2026-07-05.md`, 4 of the 8 flagged pages turned out to be false positives — already accurate, no edit warranted. Leaving them un-edited is correct, but it exposes a real limitation: `check-docs-drift.sh`'s cumulative (pre-push) mode has no way to record "reviewed, confirmed still current" separately from "doc content edited." Those 4 pages will keep surfacing, with a growing commit count, on every future push indefinitely — a source of compounding noise that risks training the team to ignore the check altogether.

## Summary

`scripts/check-docs-drift.sh`'s default mode flags a doc as possibly stale when any commit after the doc's own last-touch date modifies one of its mapped source globs. That's the only signal it has: the doc file's own git history. There is no way to tell the script "I looked at this, the flagged commits don't actually require a doc change" without editing the doc's prose — which would mean inserting a change purely to reset a timestamp, not because the content needed it.

This surfaced concretely on 2026-07-06: after fixing the 4 real gaps from the GH Pages docs-drift investigation (`docs/architecture/eval-runner.md`, `docs/architecture/overview.md`, `docs/architecture/remediation-flow.md`, `docs/development/skills-and-rules.md` — all now current), the remaining 4 (`README.md`, `docs/index.md`, `docs/architecture/duplication-flow.md`, `docs/reference/scoring-dimensions.md`) were reviewed and found already accurate:

- `README.md` / `docs/index.md` — the flagged commits (Mistral/Cerebras provider support, scoring-pattern-config externalisation) either already had a prior doc update covering them, or touch details these pages operate at too high an altitude to enumerate (individual LLM providers, in this case — that lives in `eval-runner.md`).
- `docs/architecture/duplication-flow.md` — the flagged commit *fixed the Go implementation to match what the doc already correctly said* ("exit code 2 on Critical pairs"), not the other way around.
- `docs/reference/scoring-dimensions.md` — the flagged commit refactored `removeCodeBlocks`'s internals (now delegates to `patternconfig.StripCodeBlocks`); the doc never described that function's implementation in the first place.

None of these four need a content change. But because `check-docs-drift.sh` only clears a flag when the doc file's own last-commit timestamp advances, all four will keep appearing — with an ever-growing commit count — on every subsequent `git push` that touches their mapped source globs, forever, regardless of how many times someone reviews and confirms them fine.

## Detail

This is a structural gap, not a one-off annoyance, for two reasons:

1. **Unbounded growth with no resolution path.** A doc flagged today at "1 commit since last touch" could read "40 commits since last touch" a year from now, having been correctly reviewed as current a dozen times in between. The warning gives no indication of that review history — every reader has to re-derive "is this one of the known-fine ones?" from scratch, or trust a comment buried in a PR description from months ago.
2. **Alert fatigue defeats the check's purpose.** Once a handful of files are known to be perpetual false positives, the natural human response is to skim past the whole `docs-drift` block in pre-push output rather than read each line — which is exactly the failure mode `.agents/RULES.md`'s "No man left behind" rule (added the same week, ironically one of the fixes in this batch) warns against: incidental warnings that get silently normalised into background noise.

The PR-level gate added in ADR-044 (`.github/workflows/ci.yml`'s `docs-drift` job) is not affected by this — it only diffs the current PR against its base ref, so pre-existing debt never fails a PR that doesn't touch the affected source. This is purely a problem with the pre-push, cumulative, all-history advisory mode.

## Follow-up

The missing primitive is a way to record "reviewed as of commit/date X" per doc, independent of the doc's own edit history, that the cumulative check can use as its effective baseline instead of (or in addition to) the doc's last-touch date. Sketch of a fix — not decided here, left for the plan this finding feeds:

- A small tracked sidecar file separate from `check-docs-drift.sh`'s `MAPPINGS` array (which changes rarely, on doc/source restructuring) — something like `scripts/docs-drift-reviewed.tsv` (which changes whenever a review happens) mapping `doc_path` → last-reviewed commit date.
- A tiny helper script (`scripts/mark-docs-reviewed.sh <doc_path>...`) that stamps the current `HEAD` commit's date into that sidecar, so marking a doc reviewed is a one-line, auditable, git-tracked action distinct from editing the doc's prose.
- `check-docs-drift.sh`'s cumulative mode uses `max(doc_last_touch_date, reviewed_date)` as the effective baseline per doc, instead of only `doc_last_touch_date`.
- Applying the new mechanism to the 4 currently-flagged false positives as part of implementing the fix, so this finding's own triage is reflected in the tooling rather than left as tribal knowledge in a PR description.
