# Context Index Regeneration Reference

This reference covers the technical details and edge cases for `regenerate-context-index.sh`.

## How It Works

1. Recursively finds all `.md` files under `.context/`
2. Parses YAML frontmatter from each file (content between `---` delimiters)
3. Extracts: `title`, `type`, `status`, `date`, `related`
4. Skips files with missing or malformed frontmatter (prints to stderr)
5. Sorts entries by path
6. Writes `.context/index.yaml`

## Output Format

```yaml
- title: "Plan: Add Structured Logging"
  type: plan
  status: draft
  date: 2026-06-30
- title: "Finding: Database Migration Strategy"
  type: finding
  status: active
  date: 2026-06-30
  related:
    - "../plans/add-structured-logging.md"
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success — index written, zero warnings |
| 0 (with stderr) | Success — index written, but some files excluded |
| Non-zero | Error — index not written |

## Common Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| Files appear in `.context/` but not in index | Missing or malformed frontmatter | Run check-context-frontmatter.sh to identify, then fix frontmatter |
| Index has duplicate entries | Multiple files with same path after rename | Regenerate — duplicates are resolved by path sorting |
| Index empty after regeneration | No `.md` files with valid frontmatter in `.context/` | Verify files exist and have proper `---\ntitle:\ntype:\nstatus:\ndate:\n---` blocks |

## Verification

```bash
# Check which files are missing frontmatter
scripts/check-context-frontmatter.sh .context/**/*.md

# Regenerate index
scripts/regenerate-context-index.sh

# Verify index contains expected entries
grep "^  status:" .context/index.yaml | sort | uniq -c
```

## Best Practices

- Regenerate after every `.context/` file create, rename, or delete
- Always fix frontmatter warnings before treating the index as complete
- Stage `.context/index.yaml` alongside any `.context/` file changes
- Do not hand-edit the index — it is a generated cache

