## Summary

<!-- What does this PR change and why? Keep it focused — one concern per PR. -->

## Type of change

- [ ] Bug fix (`fix/`)
- [ ] New feature (`feat/`)
- [ ] Refactor / cleanup (`refactor/`)
- [ ] Documentation / assets only (`docs/`)
- [ ] CI / tooling / dependencies (`chore/`)

## Checklist

### Code quality

- [ ] `hk check` passes (lint, format, validation)
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] New or updated tests cover the change (aim ≥90% for new analysis detectors)

### Skill / tile changes (if applicable)

- [ ] `cmd/assets/` files updated (SKILL.md, tile.json, evals, references)
- [ ] `./dist/skill-auditor eval ./cmd/assets --fail-below 0` passes (structural gate)
- [ ] `./dist/skill-auditor evaluate <changed-skill> --store` passes (self-audit)
- [ ] `tessl status` shows no warnings or out-of-sync plugins
- [ ] `tessl install` run if `tessl.json` was modified

### Documentation

- [ ] README updated if behaviour changed
- [ ] `cmd/assets/references/` updated if rubric or scoring changed
- [ ] ADR created or updated if the PR contains a binding decision
- [ ] Plan frontmatter updated (`status: active → done`) if this PR implements an existing plan

### Commit conventions

- [ ] Conventional commit messages used (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`)
- [ ] Scope included where applicable (`feat(scorer):`, `fix(reporter):`)

## Merge strategy

- [ ] Squash and merge
- [ ] Rebase and merge
- [ ] Merge commit

> **Note:** This repo prefers **squash and merge** for feature branches to keep history linear. Use rebase and merge only for long-lived branches where individual commits are meaningful.

## Related issues

Closes #

<!-- Link related PRs, ADRs, or context files: -->
<!-- Related: .context/plans/... -->
<!-- Related: docs/ADR/adr-... -->
