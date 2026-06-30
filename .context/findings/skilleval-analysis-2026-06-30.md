---
title: "Finding: SkillEval Analysis — 2026-06-30"
type: finding
status: active
date: 2026-06-30
---

# Finding: SkillEval Analysis — 2026-06-30

Date: 2026-06-30
Status: DECISION-SUPPORT, not actioned

> Investigation of [justinwetch/SkillEval](https://github.com/justinwetch/SkillEval) — a visual A/B testing workbench for AI Agent Skills — found three concrete features that could improve skill-quality-auditor.

## Summary

SkillEval is a React GUI (v1.2.1) that lets users upload two Agent Skill packages, auto-generate evaluation criteria and test prompts via an AI judge, run both skills against the prompts simultaneously, then judge and compare results side-by-side. It supports Anthropic, OpenAI, Gemini, and xAI models.

The projects are complementary rather than overlapping: ours is a CLI-driven static analysis framework with fixed 9-dimension scoring (D1–D9), while SkillEval is a visual A/B testing tool using AI-as-judge with dynamic criteria.

## Detail

### 1. A/B Testing Mode (High value)

SkillEval's core differentiator. It runs two skills against identical prompts, then an LLM judge scores each output per-criterion and declares a winner. Our project currently evaluates one skill at a time against fixed D1–D9 dimensions.

**Could adopt as:** a new `ab-test` command that takes two skill paths, generates shared test prompts, runs both through an LLM, and produces a comparative report. This would give skill authors a data-driven way to prove improvements (e.g., "my v2 skill is 23% better than v1").

### 2. AI-Generated Adaptive Criteria (Medium value)

`generateConfig.js` (app/src/utils/generateConfig.js) uses an LLM (default Claude Opus 4.8) to analyze both skills and produce 4–6 custom evaluation criteria with 1–5 rubrics, along with test prompts of varying difficulty (20% easy, 60% medium, 20% hard). It supports partial regeneration (criteria-only, prompts-only, or output-type-only) and caches configs by skill content hash.

**Could adopt as:** a `--judge` flag on `evaluate` that adds an LLM-generated criteria overlay alongside our fixed D1–D9 scoring, or as a `generate-prompts` command that creates challenge prompts from a skill's description.

### 3. Visual Evaluation via Screenshots (Medium value)

`judgeEval.js` and `screenshot-server.js` use Puppeteer/Chrome to capture rendered screenshots of HTML/CSS output for visual skills. The judge prompt embeds these as base64 images alongside source code, enabling evaluation of visual quality (typography, spacing, color, layout). Our project is text-only.

**Could adopt as:** an optional Puppeteer backend for skills that produce HTML/CSS output, capturing screenshots and including them in reports. Particularly relevant if the project ever evaluates frontend/design skills.

### Architectural notes

| Area | SkillEval | skill-quality-auditor |
|------|-----------|----------------------|
| Approach | LLM-as-judge, dynamic criteria | Rule-based scoring, fixed D1–D9 |
| Interface | React GUI | CLI |
| Output | A/B winner + per-criterion scores | Per-dimension scores + letter grade |
| Models | Anthropic, OpenAI, Gemini, xAI | Model-agnostic (no API calls) |
| Skills | `.zip` packages, `.md`, `.txt` | `SKILL.md` on filesystem |


### Additional minor items

- **Config caching**: SkillEval hashes skill content with SHA-256 and caches generated configs in memory (`cache.js`). Our project could adopt a similar strategy for the `--store` audit output (check hash, serve cached if unchanged).
- **Multi-provider API abstraction**: `api.js` provides a clean `callModel()` dispatcher across providers. If we ever add an LLM judge mode, this pattern is worth reusing.
- **Skill ZIP parsing**: `skillPackage.js` has thorough ZIP handling (file type detection, role inference, content size budgeting, image support) that may be useful if skills are ever delivered as ZIPs.
- **Structured judge prompt format**: `buildJudgePrompt.js` constructs a prompt with per-criterion rubrics and a required JSON output block. The format (`winner`, `scoreA`, `scoreB`, `breakdown` map) is a good schema for any LLM-as-judge feature.

### Dimensional fit

| Feature | Fit | Rationale |
|---------|-----|-----------|
| **A/B test command** | Cross-cutting (new command, like `batch`) | Compares two skills on identical prompts — not a single dimension, best as its own `ab-test` command alongside `evaluate`/`batch` |
| **AI-generated adaptive criteria** | D9 (Eval Validation) or D8 (Practical Usability) | Auto-generates test prompts and per-criterion rubrics to verify the skill does what it claims. D9 covers evaluation methodology; D8 covers "does the skill work in practice" |
| **Visual evaluation (screenshots)** | D8 (Practical Usability) extension | We don't evaluate rendered output today. For skills producing HTML/CSS, visual quality is an aspect of practical usability. Could extend D8's rubric or add a sub-score |
| **Config caching by SHA-256** | Cross-cutting (infrastructure) | Operational improvement — no dimension change needed |
| **Multi-provider API abstraction** | N/A (future LLM-judge feature only) | Only relevant if we adopt an AI judge mode |

## Recommended Action

1. **Implement `ab-test` command** — most impactful, directly requested by skill authors wanting to prove improvements. New command, no dimension change needed.
2. **Consider `--judge` flag** on `evaluate` — LLM-scored adaptive criteria overlay. Would slot into D9 or D8 as an optional scoring pass.
3. **Consider visual evaluation** only if the project expands into frontend/design skill evaluation. Would extend D8.
