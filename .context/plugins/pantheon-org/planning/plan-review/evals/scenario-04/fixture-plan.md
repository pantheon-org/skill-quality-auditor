---
title: "Plan: block broken documentation links before they ship"
type: PLAN
status: DRAFT
date: 2026-05-12
effort: S
value: MEDIUM
themes:
  - DOCS
---

# Plan: block broken documentation links before they ship

## Goal

No broken internal links ever reach the published documentation site. Every link
in `docs/**/*.md` that points at another repo file must resolve to an existing
target at the moment code is shipped.

## Scope

**In scope:** a `scripts/check-doc-links.sh` link checker; wiring it so it runs
automatically as part of the developer's normal flow.

**Out of scope:** external (http) link checking; rewriting existing broken links
(a separate cleanup).

## Phases

### Phase 1 — Build the checker

- Task 1.1: Write `scripts/check-doc-links.sh` — parse every `[text](path)` in
  `docs/**/*.md`, resolve `path` relative to the file, exit non-zero on any miss.
- Task 1.2: Add a fixture doc with a known-broken link and confirm the script
  exits non-zero on it.

### Phase 2 — Wire it in

- Task 2.1: Register `check-doc-links.sh` in `hk.pkl`'s `pre-push` hook, so it
  runs on every `git push`. The pre-push hook is the script's only caller.
- Task 2.2: Document in `CONTRIBUTING.md` that pushes are blocked on broken links.

## Verification

- The checker exits non-zero on the fixture broken link.
- A `git push` carrying a broken link is rejected locally.

## Open Questions

- Should the fixture doc live under `docs/` or `testdata/`?
