---
title: "ADR-030: pr-author skill is GitHub-only for v0.1.0; GitLab deferred to v0.2.0"
status: accepted
date: 2026-07-03
context:
  - path: .context/plans/pr-author-skill.md
  - path: .context/plugins/pantheon-org/pr-author/SKILL.md
---

**Status:** Accepted
**Date:** 2026-07-03

## Context

The `pr-author` helper skill was designed to teach agents how to create and maintain
GitHub PRs with live descriptions — template discovery, intelligent filling, and lifecycle
updates. During plan review, the question arose whether the initial version (v0.1.0)
should also support GitLab merge requests.

## Decision

1. **GitHub-only for v0.1.0** — The skill targets GitHub PRs exclusively in its first
   version. All `gh` CLI commands, API endpoints, and template paths are GitHub-specific.
2. **GitLab support deferred to v0.2.0** — If user demand arises, a future version will
   add GitLab merge request support with `glab` CLI or GitLab API equivalents.
3. **Document the scope boundary explicitly** — The SKILL.md "When NOT to Use" section
   states: "GitLab merge requests (GitHub-only for v0.1.0)".

## Consequences

- The skill is simpler and more focused for v0.1.0 — no conditional logic for platform
  differences (template paths, CLI commands, API shapes).
- Users on GitLab cannot use the skill until v0.2.0.
- The platform abstraction can be designed with real GitHub usage data before adding
  GitLab, rather than guessing at common patterns.
- When GitLab support is added, the template discovery paths will differ (`.gitlab/` vs
  `.github/`) and the CLI will switch from `gh` to `glab`.
