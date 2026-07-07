---
title: "ADR-059: Markdown placement by category, authored docs under docs/, minimal root"
status: proposed
date: 2026-07-07
context:
  - path: .agents/RULES.md
---

**Status:** Proposed
**Date:** 2026-07-07

## Context

The repository root accumulates Markdown over time. Some of it belongs there by
strong external convention, but ad hoc documentation added at the root has no home
rule, so it drifts there by default. Authoring `ONBOARDING.md` at the root surfaced
the gap: nothing said where a new document should live, and nothing flagged the stray
file.

An early draft of this decision used a flat allowlist of seven root files justified as
"convention." That framing was inconsistent: the seven files are not at the root for
the same reason. `CHANGELOG.md` is a generated artifact, not authored documentation.
`README.md` is a platform entry point. `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, and
`SECURITY.md` are GitHub community-health files that GitHub also reads from `.github/`.
`AGENTS.md` and `CLAUDE.md` are agent entry points. A flat list hides these
distinctions and invites bikeshedding over which names are "allowed."

The docs site (`docmd.config.json`) builds from `src: ./docs`, so any Markdown moved
under `docs/` is published to the public GitHub Pages site and is subject to the
`docs-check` orphan rule (it must be linked from the docs navigation). Placement is
therefore also a publish decision, not only a tidiness decision.

The rule is scoped to **Markdown only**. Non-Markdown tool and build configuration
(`go.mod`, `main.go`, `.goreleaser.yaml`, `mise.toml`, `hk.pkl`, `tessl.json`,
`docmd.config.json`, `release-please-config.json`, `renovate.json`, and similar) is
out of scope and stays wherever its tool resolves it, typically the root.

## Decision

Place Markdown by **category**, not by a flat allowlist:

1. **Authored human documentation** lives under `docs/` (published to the docs site).
2. **Generated artifacts** live wherever their generator writes them, not under
   `docs/`. `CHANGELOG.md` (release-please) stays at the root because that is where
   its generator is configured to write; it is output, not an authored document.
3. **Entry files that a platform or agent resolves from a fixed path** live at that
   path:
   - `README.md` at the root (GitHub repo landing page).
   - `AGENTS.md` and `CLAUDE.md` (symlink to `AGENTS.md`) at the root (agent tooling
     resolves them from the repository root).
   - GitHub community-health files (`CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`,
     `SECURITY.md`) under `.github/`, GitHub's first-class location for them,
     alongside the existing `CODEOWNERS` and `pull_request_template.md`.

Applying this, the community-health trio moves from the root to `.github/`, leaving
the root Markdown set as: `README.md`, `CHANGELOG.md` (generated), `AGENTS.md`,
`CLAUDE.md`.

Moving a document under `docs/` publishes it to the GitHub Pages site. A document
intended to be internal-only is not made public merely to satisfy category 1; if it
must stay unpublished it lives outside `docs/` (for example under `.context/`), and
that choice is recorded where the document is added.

## Consequences

- **Easier:** every Markdown file has a home derived from what it *is* (authored,
  generated, or entry point), so a new document's location is a lookup, not a debate.
- **Easier:** the root Markdown set is now minimal and each remaining file is
  justified by a concrete resolver (GitHub landing, release-please, agent tooling), so
  a stray root Markdown file is a clear, mechanically checkable violation.
- **Cost paid:** the trio move updated the one live inbound link (`README.md` to
  `CONTRIBUTING.md`). Historical references in `.context/plans/*` were left unchanged
  as institutional records; they are excluded from the lint and docs site.
- **Constraint:** placing a doc under `docs/` publishes it and requires a navigation
  link, so authors decide public-versus-internal at authoring time.
- **Applied:** the internal onboarding guide (`ONBOARDING.md`) was an internal-only
  authored doc, so it moved to `.context/instructions/team-onboarding.md` rather than
  `docs/`, demonstrating the internal-only escape hatch above.
- **Out of scope by design:** this ADR does not relocate any non-Markdown config and
  does not add enforcement tooling. A pre-commit guard for stray root Markdown is a
  follow-up item.
- **Companion rule:** the convention is mirrored as an agent rule in
  `.agents/RULES.md` so agents apply it when creating documentation.
