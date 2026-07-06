---
title: "Finding: Tessl Eval Criteria JSON Schema — 2026-06-30"
type: FINDING
status: ACTIVE
date: 2026-06-30
value: LOW
---

# Finding: Tessl Eval Criteria JSON Schema — 2026-06-30

Date: 2026-06-30
Status: DECISION-SUPPORT, not actioned

> The tessl CLI now requires `criteria.json` to use the `weighted_checklist` schema with `type`, `context`, `name`, and `description` fields. All 6 scenarios were migrated from the old `checklist` format.

## Summary

A CI failure on PR #78 revealed that the tessl CLI (version installed by `curl -fsSL https://get.tessl.io | sh`) validates `criteria.json` against a schema that requires:

1. `"type": "weighted_checklist"` at the top level
2. A `context` field (string) describing the evaluation scope
3. Each checklist item must have `name` and `description` (string) instead of `id` and `criterion`

The old format (`{"checklist": [{"id": "c1", "criterion": "...", "max_score": N}]}`) was silently accepted by earlier CLI versions but now causes a hard validation failure.

## Detail

### Old format (broken)

```json
{
  "checklist": [
    {"id": "c1", "criterion": "Do the thing", "max_score": 25}
  ]
}
```

### New format (compatible)

```json
{
  "type": "weighted_checklist",
  "context": "Describes the overall evaluation scenario purpose.",
  "checklist": [
    {"name": "Short criterion name", "description": "Detailed description of what is being evaluated", "max_score": 25}
  ]
}
```

### Changes per item

| Old field | New field | Reason |
|-----------|-----------|--------|
| `id` (e.g. `"c1"`) | `name` (human-readable string) | IDs were meaningless ordinals; names are self-documenting |
| `criterion` (string) | `description` (string) | Renamed for clarity |
| `max_score` (number) | `max_score` (number) | Unchanged |
| *(missing)* | `type` at top level | Required enum: `"weighted_checklist"` |
| *(missing)* | `context` at top level | Required string explaining evaluation scope |

### Error signature

The CLI validates all scenarios eagerly and reports the first failure. The error message was:

```
✘ Invalid criteria.json (in cmd/assets/evals/scenario-06):
  context: Invalid input: expected string, received undefined
  type: Invalid input: expected "weighted_checklist"
  checklist.0.name: Invalid input: expected string, received undefined
  checklist.0.description: Invalid input: expected string, received undefined
```

### Affected files

All 6 scenarios under `cmd/assets/evals/scenario-*/criteria.json` were migrated in commit `a15cbd2`.

### Prevention

When adding new scenarios, ensure `criteria.json` uses the `weighted_checklist` format above. The tessl eval-setup skill auto-generates scenarios with the correct format — only manually created scenarios risk using the old schema.

### Tessl CLI version pinning

The CI quality-gate workflow pins tessl CLI to v0.88.2 (`skill-quality.yml:44`). This was required because v0.89.0 introduced a mandatory project link — `tessl eval run` fails with:

```
✘ No existing project safely matches this directory.
  Run tessl project create ... if this should create a new project,
  or run tessl project repair if this directory should link...
```

The project link is stored outside the repo (in `~/.tessl/`), so it can't be committed. Creating a project on every CI run is not viable. The version is pinned via:

```yaml
- name: Install Tessl CLI
  env:
    TESSL_VERSION: "0.88.2"
  run: curl -fsSL https://get.tessl.io | sh
```

When upgrading beyond v0.89.0, either:
1. Create a tessl project once and commit the link metadata (requires tessl to support in-repo config)
2. Or reassess whether CI should create an ephemeral project per run
