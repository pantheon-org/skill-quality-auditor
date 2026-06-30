---
title: "ADR-002: Port SkillLens Mode/ModeSet data structures into analysis package"
status: proposed
date: 2026-06-30
context:
  - path: .context/findings/skilllens-integration-2026-06-30.md
---

**Status:** Proposed
**Date:** 2026-06-30

## Context

The SkillLens research framework from Microsoft defines reusable data
structures for classifying agent behaviour patterns: `Mode` (success/failure
pattern with evidence) and `ModeSet` (collection of modes with source
trajectories). The skill-quality-auditor's analysis pipeline currently lacks
a structured way to represent and pass around these patterns.

## Decision

Port the `Mode` and `ModeSet` types from SkillLens into a new
`analysis/modes.go` file:

```go
type ModeType string
const (
    SuccessMode ModeType = "success"
    FailureMode ModeType = "failure"
)

type Mode struct {
    Type                ModeType
    Pattern             string
    Description         string
    Evidence            string
    SourceTrajectoryIDs []string
}

type ModeSet struct {
    SuccessModes        []Mode
    FailureModes        []Mode
    SourceTrajectoryIDs []string
    Summary             string
}
```

Also extract SkillLens's failure mode categories (error patterns,
anti-patterns, pitfalls) into the D3 anti-pattern scorer as additional
detection patterns.

## Consequences

- Low-effort, high-value port with immediate integration points
- Duplication analysis pipeline can output `ModeSet` for richer reports
- D3 scorer gains empirically grounded categories from 5 benchmarks
- Meta-skill findings inform rubric documentation in `framework-dimensions.md`
