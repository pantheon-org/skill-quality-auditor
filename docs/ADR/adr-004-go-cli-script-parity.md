---
title: "ADR-004: Implement shell script functionality in Go CLI"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/go-cli-script-parity-2026-04-27.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

Several maintenance and workflow operations were implemented as shell scripts
rather than as Go subcommands. This creates a fragmented user experience where
some operations are `skill-auditor <cmd>` and others require running `.sh`
files. The scripts also lack cross-platform compatibility.

## Decision

Port all essential shell script functionality into the Go CLI as new
subcommands. Each script becomes a self-documenting `skill-auditor <cmd>` with
proper flag parsing, error handling, and cross-platform support.

## Consequences

- Consistent CLI interface — all operations are `skill-auditor <cmd>`
- Cross-platform compatibility (Windows support)
- Better error handling and testability
- Shell scripts remain available as backwards-compatible wrappers during migration
