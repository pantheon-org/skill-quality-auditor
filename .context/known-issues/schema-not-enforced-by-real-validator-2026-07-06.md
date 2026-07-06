---
title: "Known Issue: context-frontmatter.schema.json is not enforced by a real JSON-schema validator"
type: known-issue
status: active
date: 2026-07-06
severity: low
value: low
related:
  - ../plugins/pantheon-org/context-mgmt/context-file/assets/schemas/context-frontmatter.schema.json
  - ../plugins/pantheon-org/context-mgmt/context-index/scripts/validate-context-frontmatter.sh
  - ../plans/context-prioritisation-signal-2026-07-06.md
---

`context-frontmatter.schema.json` is a JSON Schema (draft 2020-12) declaring
`additionalProperties: false`, per-field `enum`s, `required`, and `pattern`
constraints. But nothing in the repo runs a real JSON-schema validator against
it. The only enforcement is `validate-context-frontmatter.sh`, a shell script
that re-implements a *subset* of the schema's semantics in Python:

- It checks `required` presence, `enum` membership, and `pattern` matches by
  reading the schema's `properties` — so those three are enforced.
- It does **not** enforce `additionalProperties: false`. An unknown or
  mistyped frontmatter key (for example `valeu: high`) passes validation
  silently instead of being rejected.
- Per-type applicability (e.g. `value` and `effort` conceptually apply only to
  certain `type`s) is not modelled in the schema at all; it lives as hardcoded
  conditional logic in the validator.

## Why it matters

The schema reads as the authoritative contract, but the guarantees a reader
would infer from it (no stray keys, type-scoped fields) are only as strong as
the shell validator, which is weaker. A typo'd key is the most likely failure:
it would not fail-validate, so a field intended to carry a `value` grade could
silently be absent from the index.

## Discovered

While shipping Phase 1 of the prioritisation-signal plan. Adding the `value`
enum to the schema was picked up by the shell validator automatically (it
derives enum checks from `properties`), which confirmed the validator reads the
schema — but also surfaced that `additionalProperties` is never checked.

## Possible fix (not scheduled)

Add a real JSON-schema validation step (e.g. `python -m jsonschema` or a Go
`santhosh-tekuri/jsonschema` pass) to the pre-commit/CI gate that validates each
file's parsed frontmatter against the schema, either replacing or backing up the
shell validator. Deferred: low severity because the shell validator covers the
high-frequency failure modes (missing required, bad enum), and the typo-key
failure has not been observed in practice.
