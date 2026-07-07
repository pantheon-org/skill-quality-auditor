---
title: "Plan: add a DEFERRED lifecycle status for real-but-parked context items"
type: PLAN
status: DONE
date: 2026-07-07
related:
  - ../../docs/ADR/adr-060-deferred-lifecycle-status.md
  - ../instructions/value-rubric.md
  - ../instructions/ways-of-working.md
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
---

# Plan: add a DEFERRED lifecycle status

Status: DONE — implemented and verified on branch `feat/deferred-status`
Date: 07-07-2026

## Context

The status enum was `DRAFT | ACTIVE | DONE | SUPERSEDED`. Items that are real and
not done but cannot be actioned yet — date-gated (e.g. the tessl-eval-decommission
Bucket A, gated to ~15-07-2026) or externally blocked (e.g. the plan-review
execution-location lens, blocked on an eval-harness limitation) — had to be filed
as `ACTIVE`. That polluted the "what's next" read protocol: blocked or gated work
surfaced as top candidates it could not, in fact, be picked up.

## Decision

Add `DEFERRED` as a fifth status: a real item that **cannot be actioned yet**
(date-gated or externally blocked), distinct from `ACTIVE` (pick-up-next) and
`DRAFT` (not yet reviewed).

Design choices (the first two confirmed with the maintainer; the rest added after
a critical review of ADR-060 — see
[`deferred-status-critical-review-2026-07-07`](../findings/deferred-status-critical-review-2026-07-07.md)):

1. **Read-protocol placement — strict second tier (not exclusion).** The protocol
   splits candidates into tier 1 (`DRAFT`/`ACTIVE`) and tier 2 (`DEFERRED`), always
   exhausting tier 1 first. Within a tier the existing sort applies (value desc,
   effort asc, `themes[0]`). A `DEFERRED` item never outranks a tier-1 item
   regardless of value. *Alternative considered and rejected:* excluding `DEFERRED`
   from the pick entirely and listing it separately — rejected because "lower
   priority than active, still visible in the same queue" was the stated intent.
2. **`value`/`themes`/`effort` stay REQUIRED on `DEFERRED`** (same as `DRAFT`/`ACTIVE`),
   so a parked item re-ranks cleanly the moment it is reactivated. Only
   `DONE`/`SUPERSEDED` remain exempt.
3. **Semantics narrowed to "cannot", not "won't".** `DEFERRED` excludes merely
   low-priority work, which stays `ACTIVE` with `value: LOW`. This removes an overlap
   with the value axis flagged in the critical review.
4. **Optional `deferred_until: YYYY-MM-DD` field** for date-gated items — only valid
   with `status: DEFERRED` (validator-enforced). The read protocol surfaces an item
   whose `deferred_until` has passed as reactivation-eligible, so date-gated work
   does not rot in tier 2.

## Changes

- Schemas: `context-frontmatter.schema.json`, `remediation-plan.schema.json` — enum
  gains `DEFERRED`; the frontmatter schema description documents the tier semantics
  and adds the optional `deferred_until` property (date-patterned).
- Validator: `validate-context-frontmatter.sh` — the three requiredness branches
  (effort, value, themes) now gate on `("DRAFT", "ACTIVE", "DEFERRED")`; a new guard
  rejects `deferred_until` on any status other than `DEFERRED` (its date format is
  auto-validated from the schema pattern).
- Migration: the plan-review execution-location lens is re-statused `ACTIVE → DEFERRED`
  (externally blocked, no `deferred_until`). The tessl-eval-decommission Bucket A
  (date-gated to 2026-07-15) is re-statused on its own branch to avoid a cross-branch
  status conflict.
- Read protocol: `AGENTS.md` (and its `CLAUDE.md` symlink), `value-rubric.md`
  (canonical two-tier steps), `ways-of-working.md` (protocol + re-grade rule extended
  to `ACTIVE ↔ DEFERRED` transitions), `team-onboarding.md`.
- Author docs: context-file `SKILL.md`, `yaml-frontmatter-guide.md`,
  `context-file-template.yaml`.

The `.tessl` mirror is gitignored and regenerated from source in CI, so no mirror
edit is committed. The Go CLI does not gate on the status enum (it only emits
`DRAFT` at remediation-plan generation), so no Go change was needed.

## Verification

- Both schemas parse as JSON.
- Validator self-test: a `type: PLAN` `status: DEFERRED` file fails without
  value/themes/effort and passes with them.
- `go test ./...`, `hk check` (context-frontmatter, context-index, adr-undocumented,
  markdownlint) all green.
