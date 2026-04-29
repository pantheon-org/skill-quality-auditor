---
name: code-quality-advisor
description: Evaluates code quality across a range of dimensions including readability, maintainability, and structural coherence. Provides actionable feedback to help developers improve their codebase through targeted suggestions and best practice guidance.
triggers:
  - code review requested
  - quality check needed
---

## Mindset

Apply this skill whenever you need an objective assessment of code quality. ALWAYS look beyond surface-level issues. NEVER ignore structural concerns in favor of style-only feedback.

## When to Use

- Reviewing code for quality improvements
- Assessing technical debt levels
- Onboarding new developers to quality standards

## Procedures

1. Analyze the code structure for cohesion and coupling
2. Check naming conventions and readability
3. Identify duplicated logic and refactoring opportunities
4. Score overall quality and list top improvements

## Anti-Patterns

BAD: Focusing only on formatting

```python
# BAD — formatting is surface-level; structural issues matter more
black my_module.py
```

GOOD: Evaluate logical structure, responsibilities, and abstractions first.

## References

- [Clean Code Principles](https://example.com/clean-code)
