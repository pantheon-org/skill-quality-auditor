# D6: Freedom Calibration (15 points)

**Purpose:** Balance prescription (rigid rules) vs flexibility (guidelines) for the skill type.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 13–15 | Appropriate calibration for skill type |
| 10–12 | Slightly too rigid or loose |
| 7–9 | Mismatched calibration |
| 0–6 | Completely wrong |

## Pattern configuration

The "when not to use" phrase list used to detect scoping signals is not hardcoded in
`scorer/d6_freedom_calibration.go` — it lives in `scoring-patterns.yaml` under
`patterns.d6_freedom_calibration`, loaded via `internal/patternconfig`. Maintainers editing
the shipped defaults tune `cmd/assets/assets/config/scoring-patterns.yaml` directly (see
ADR-028). Anyone running a pre-built binary can override the same list without
recompiling — see [Configuring scoring patterns](../development/setup.md#configuring-scoring-patterns)
for the full mechanism (`-c/--config`, a project-local `./scoring-patterns.yaml`, or a
per-OS user config directory). `skill-auditor eval` always scores against the embedded
defaults regardless of any override, so CI eval results stay reproducible.

## Calibration Levels

### Rigid (Mindset skills)

Strong rules, must follow.

- Example: proof-of-work — "NEVER trust agent reports without verification"
- Use for: critical foundations, security, correctness

### Balanced (Process skills)

Clear steps with contextual flexibility.

- Example: TDD — "Red → Green → Refactor (adapt to context)"
- Use for: workflows, methodologies

### Flexible (Tool skills)

Options and trade-offs presented.

- Example: typescript-type-system — "Choose based on use case"
- Use for: technical tools, patterns

## Examples

**Well-Calibrated (14/15):**

```markdown
# Proof of Work (Mindset skill)

## Zero-Tolerance Rules
NEVER trust agent completion reports without verification.
ALWAYS show command output as proof.
ZERO exceptions to verification protocol.
```

Appropriately rigid for a critical verification mindset skill.

**Miscalibrated (7/15):**

```markdown
# TypeScript Basics (Tool skill)

## Rules
ALWAYS use const for all variables.
NEVER use let or var under any circumstances.
```

Too rigid — `let` has valid use cases in a tool skill.

## Academic References

```bibtex
@inproceedings{staufer2026agentindex,
  title         = {The 2025 AI Agent Index: Documenting Technical and Safety Features of Deployed Agentic AI Systems},
  author        = {L. Staufer and K. Feng and K. Wei and L. Bailey and Y. Duan and M. Yang and A. P. Ozisik and S. Casper and N. Kolt},
  year          = {2026},
  booktitle     = {Proceedings of the ACM Conference on Fairness, Accountability, and Transparency (ACM FAccT 2026)},
  publisher     = {ACM},
  url           = {<https://arxiv.org/abs/2602.17753>}
}

```

```bibtex
@inproceedings{dibia2025mas,
  title         = {Designing Multi-Agent Systems: Principles, Patterns and Implementation for AI Agents},
  author        = {V. Dibia},
  year          = {2025},
  publisher     = {O'Reilly Media},
  url           = {<https://www.oreilly.com/library/view/designing-multi-agent-systems/9781098194945/>}
}

```

```bibtex
@article{bednarbrandt2026autonomy,
  title         = {From Autonomy to Agency: A 10-Level Framework for AI's Evolution and Organisational Readiness},
  author        = {M. Bednar-Brandt},
  year          = {2026},
  journal       = {SSRN},
  url           = {<https://papers.ssrn.com/sol3/papers.cfm?abstract_id=6226382>}
}
```

```bibtex
@article{sorensen2026specification,
  title         = {Specification as the New Management},
  author        = {Sorensen},
  year          = {2026},
  journal       = {ResearchGate},
  url           = {<https://www.researchgate.net/publication/401626622}>
}

```

```bibtex
@article{tao2025orchestration,
  title         = {LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Multi-Model Collaboration},
  author        = {R. Tao},
  year          = {2025},
  journal       = {Research Square},
  url           = {<https://www.researchsquare.com/article/rs-9323974/latest}>
}
```
