---
title: "Finding: don't track the .tessl/plugins mirror — use a CI-only ephemeral diff instead"
type: FINDING
status: ACTIVE
date: 2026-07-06
value: LOW
related:
  - index-yaml-split-review-2026-07-06.md
  - ../../.context/plugins/pantheon-org/planning/design-debate/SKILL.md
---
# Finding: don't track the .tessl/plugins mirror — use a CI-only ephemeral diff instead

> Run via `design-debate` (Advocate / Skeptic / Migration-Risk). Decision: "Should `.tessl/plugins/pantheon-org/**` (the vendored mirror of this repo's own authored helper skills) be tracked in git instead of gitignored, so CI can catch drift between source and the installed mirror?" **Verdict: `proceed_with_modification`** — protect against drift, but via a CI-only ephemeral diff, not by tracking the mirror.

## Summary

Raised by the `.context/index.yaml` split debate's Migration-Risk finding: `.tessl/plugins/**` is entirely gitignored (`.gitignore` line 11), so any drift between an authored skill and its installed mirror is structurally invisible to git-based CI, repo-wide, not just for the one skill (`context-index`) originally found diverged.

Grounding facts: `.tessl/plugins/pantheon-org/**` (our own 12 authored skills) is 436KB/74 files — small. `.tessl/plugins/pantheon-ai/**` and `tessl-labs/**` (third-party registry content) total ~174MB — explicitly out of scope, nobody proposed tracking that. A fresh drift check across all 12 `pantheon-org` skills found **zero** meaningful content differences today — the one apparent diff (`adr-capture`) was just the expected `evals/` exclusion plus a stray `.aislop` artifact file, not real drift.

## Why proceed_with_modification, not proceed

The Advocate's case (review-time visibility via a tracked diff, precedent from `.context/audits/` already being tracked generated content, resilience against a Tessl CI step going advisory) was real but didn't survive Migration-Risk's finding: **tracking the mirror is inert on its own** — nothing in `hk.pkl` or `.github/workflows/` verifies the tracked copy matches a fresh install today, so tracking alone adds zero protection without *also* writing a new CI step. Once a new CI step is required either way, the Skeptic's case wins: an ephemeral `tessl install` + diff in CI gives identical protection with none of tracking's costs — a fragile `.gitignore` negation (Migration-Risk verified a naive `!.tessl/plugins/pantheon-org/` does *nothing*; the correct layered pattern is non-obvious and easy to widen by mistake later), a doubled 74-file diff on every skill-content PR that reviewers can't meaningfully re-review anyway, and permanent inert weight for content 100% derivable from already-tracked source.

## Fix (not yet implemented)

Add a CI step (new job or step in an existing workflow) that runs `tessl install` into a scratch location and diffs the result against `.context/plugins/pantheon-org/**`'s expected install output, failing the build on divergence. Scope the diff to `pantheon-org/**` only — never touch or evaluate the third-party `pantheon-ai/**`/`tessl-labs/**` content. No `.gitignore` change needed. Not implemented in this finding; a natural next step for `plan-create` if picked up.

## Recorded per design-debate's persistence rule

Verdict is `proceed_with_modification` (something is being acted on, even if not yet implemented) → recorded as a finding, not a known-issue.
