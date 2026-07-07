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

Add `DEFERRED` as a fifth status: a real item intentionally parked (date-gated,
externally blocked, or deprioritised), distinct from `ACTIVE` (pick-up-next) and
`DRAFT` (not yet reviewed).

Two design choices were confirmed with the maintainer:

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

## Changes

- Schemas: `context-frontmatter.schema.json`, `remediation-plan.schema.json` — enum
  gains `DEFERRED`; the frontmatter schema description documents the tier semantics.
- Validator: `validate-context-frontmatter.sh` — the three requiredness branches
  (effort, value, themes) now gate on `("DRAFT", "ACTIVE", "DEFERRED")`.
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
