# PR Template Lookup Paths

This document lists all known PR template locations and the precedence order the `pr-author` skill uses when discovering the active template.

## Precedence Order

1. **`.github/PULL_REQUEST_TEMPLATE/` directory**
   - One or more `.md` files (e.g., `bug-fix.md`, `feature.md`, `release.md`)
   - Used by GitHub when opening PRs via the web UI with a template selector
   - If multiple files exist, the skill selects based on branch prefix matching (see SKILL.md Workflow Step 1)

2. **`.github/pull_request_template.md`** (flat file, snake_case)
   - The most common single-template location
   - Auto-detected by `gh pr create` in interactive mode
   - **Important:** `gh pr create --template` does NOT work with flat files — the skill reads the file and passes it to `--body`

3. **`docs/PULL_REQUEST_TEMPLATE.md`**
   - Alternative location for projects that keep templates in `docs/`

4. **`PULL_REQUEST_TEMPLATE.md`** (repo root)
   - Legacy / minimal project location

5. **Fallback: built-in sections**
   - If no template is found, the skill generates a body with:
     - Summary
     - Type of change
     - Checklist
     - Related issues

## Template Shapes and `gh` CLI Behavior

| Shape | Path Pattern | `gh` Flag | Works? |
| --- | --- | --- | --- |
| Directory | `.github/PULL_REQUEST_TEMPLATE/*.md` | `--template <filename>` | Yes |
| Flat | `.github/pull_request_template.md` | `--template` | No (silently ignored) |
| Flat | `.github/pull_request_template.md` | `--body "$(cat file)"` | Yes |

## Notes

- HTML comments (`<!-- -->`) in templates are guidance for human authors. The skill strips them when filling.
- The fallback sections are intentionally minimal — they provide structure without imposing repo-specific conventions.
- Template discovery is a file-existence check (5 paths). No external script is needed.
