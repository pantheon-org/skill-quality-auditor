---
title: "Finding: Arxiv Self-Evolution Research Survey"
type: FINDING
status: ACTIVE
date: 2026-07-03
value: LOW
related:
  - ../findings/evoskill-integration-2026-07-02.md
  - ../plans/evoskill-core-loop-port-2026-07-02.md
  - ../../docs/ADR/adr-026-evoskill-core-loop.md
---

# Finding: Arxiv Self-Evolution Research Survey

Surveyed arxiv for papers on automated skill improvement loops for LLM agents.
Six papers triaged in depth. Summary below.

---

## 1. SkillForge (2604.08618, SIGIR 2026) — Most actionable architecture

**URL:** https://arxiv.org/abs/2604.08618

**Pipeline:** Domain-Contextualized Skill Creator → Agent Execution → Failure Analyzer → Skill Diagnostician → Skill Optimizer. A three-stage automated loop: (1) Failure Analyzer classifies each bad case across 4 dimensions (Knowledge, Tool, Clarification, Style), (2) Skill Diagnostician aggregates failures by category and maps them to specific sections of SKILL.md, (3) Skill Optimizer rewrites the skill via a Virtual File System (CRUD on SKILL.md and references/).

**Key results:** +9-12pp Strict Consistency Rate across 3 iterations from any starting point (human-authored, domain-generated, or generic). Automated evolution surpassed manually curated expert knowledge. Demonstrated on 1,883 real cloud-support tickets across 5 domains.

**What to steal:**
- Multi-dimensional failure analysis (4 axes) is more structured than EvoSkill's single failure signal
- Aggregation phase: cluster individual failures into systemic patterns before diagnosis
- Skill Diagnostician maps failures to specific SKILL.md sections (not just "add more content")
- Gap-vs-bug diagnosis: different remediation paths for knowledge gaps vs implementation errors

---

## 2. OpenSkill (2606.06741) — Best no-regression mechanism

**URL:** https://arxiv.org/abs/2606.06741

**Pipeline:** Open-world knowledge acquisition → skill drafting → leakage-free evolution against self-built virtual tests → zero-shot target deployment. The key innovation is a **Virtual Verifier** that generates pytest suites from independently verifiable facts (API docs, known dataset properties, documented output formats), never touching ground-truth tests.

**Key results:**
- Best automated pass rate on SkillsBench (+8.9/+8.8 over strongest closed-world baseline on Opus 4.6/GPT 5.2)
- Virtual verifier covers 88.9% of ground-truth test intents despite never seeing them
- Skills transfer across models (Opus 4.6 → Haiku 4.5, Qwen3 Coder, DeepSeek V3, Mistral Large 3) with no adaptation needed
- Cross-entropy peaks at 3 iterations then degrades (overfitting to virtual feedback)

**What to steal:**
- Self-built verifier for no-regression guarantees — solves a key open question in our draft plan
- Gap-vs-bug diagnosis with targeted re-retrieval when knowledge is missing
- Iteration budget of J=3 is empirically optimal across multiple papers
- Model-agnostic skill transfer proves skill-level optimization is the right abstraction

---

## 3. Ctx2Skill (2604.27660) — Best adversarial robustness

**URL:** https://arxiv.org/abs/2604.27660

**Pipeline:** Multi-agent self-play loop: Challenger generates tasks+rubrics, Reasoner solves them guided by evolving skill set, Judge provides binary feedback. Both agents co-evolve separate skill sets through dedicated Proposer+Generator pairs. A Cross-time Replay mechanism selects the most generalizable skill set across iterations.

**Key results:**
- +5.4pp on GPT-4.1 (11.1%→16.5%), +4.7pp on GPT-5.1 (21.1%→25.8%), +3.2pp on GPT-5.2 (18.2%→21.4%)
- Cross-time Replay beats every fixed-iteration selection (including best single iteration)
- Co-evolution prevents adversarial collapse better than single-sided updates

**What to steal:**
- Cross-time Replay mechanism — critical for our loop. Selects skill set maximizing ρ_hard × ρ_easy across all iterations
- Co-evolving both sides (Challenger tightens probes as Reasoner improves) prevents over-specialization
- Multiplicative hard×easy scoring rejects sets that regress on easier cases
- Strictly adversarial: no side sees the other's skill set

---

## 4. Continual Harness (2605.09998) — Best reset-free online learning

**URL:** https://arxiv.org/abs/2605.09998

**Pipeline:** Two-loop architecture — inner loop acts in environment, outer loop (Refiner) edits all 4 harness components (prompt, sub-agents, skills, memory) mid-episode from a sliding trajectory window. No environment resets. Optionally extends to model-harness co-learning where open-source weights update jointly with harness state.

**Key results:**
- First AI system to complete multiple Pokémon RPGs (Blue, Yellow Legacy hard mode, Crystal)
- Recovers majority of gap to hand-engineered expert harness from a minimalist baseline
- Path-cost deficit of evolved navigation skills falls from ~50% to single digits within a single episode
- Online co-learning loop drives sustained in-game milestone progress in open-source models

**What to steal:**
- Reset-free inline refinement — no train/eval split needed
- CRUD edits to all harness components (not just skills)
- Process-reward co-learning loop for model-harness joint improvement
- Capability floor detection: weak models regress with the harness

---

## 5. EvoAgent (2604.20133) — Best skill lifecycle management

**URL:** https://arxiv.org/abs/2604.20133

**Pipeline:** Online execution loop with three-stage skill matching (keyword→embedding→LLM) + offline evolution loop driven by user feedback. Skills have evolutionary metadata (usage_count, success_rate) and maturity levels (Budding→Growing→Mature→Proficient). Three-layer memory system (SOUL.md/USER.md/MEMORY.md).

**Key results:**
- +28% overall average score on GPT5.2 in real-world foreign trade scenarios (5-dim LLM-as-Judge)
- Model transfer experiments show agent architecture matters as much as model capability
- Skills persist and improve across sessions without labeled data

**What to steal:**
- Skill maturity levels for pruning decisions — natural way to remove unused skills
- Three-stage matching cascades cheap→expensive
- Three-layer memory (SOUL/USER/MEMORY) for cross-session knowledge accumulation
- MDP formulation of the evolution problem formalizes the loop

---

## 6. EvoSkill (2603.02766) — Already in ADR-026, additional findings

**URL:** https://arxiv.org/abs/2603.02766

**Pipeline:** Executor → Proposer → Skill-Builder. Pareto frontier of top-k agent programs. Each iteration: round-robin parent selection from frontier, detect failures, propose skill edit, build candidate, evaluate on held-out, frontier update.

**Key results (new beyond ADR-026):**
- Round-robin parent selection from frontier ensures candidate diversity
- Category-aware stratified sampling clusters failures before proposing fixes
- Skill-merge post-hoc: combining unique skills from independent runs beats any single run
- Zero-shot transfer: SealQA skill → BrowseComp +5.3%
- 10% training data sweet spot — diminishing returns beyond that

**What to steal:**
- Round-robin parent selection ensures diverse candidates
- Skill-merge post-hoc: combine unique skills across runs
- Category-aware failure clustering

---

## Synthesis: Recommended adoptions for `--evolve` design

| Component | Best source | Why |
|-----------|-------------|-----|
| Failure analysis | SkillForge (4-dim) | More structured than EvoSkill's single failure signal |
| Diagnostician | SkillForge | Maps failures to specific SKILL.md sections |
| No-regression guard | OpenSkill + Ctx2Skill | Virtual verifier + Cross-time Replay |
| Loop controller | Continual Harness | Reset-free, inline refinement |
| Skill lifecycle | EvoAgent | Maturity levels for pruning |
| Proposer diversity | EvoSkill | Round-robin from Pareto frontier |
| Iteration budget | Multiple papers | J=3 empirically optimal |

The existing draft plan (`evoskill-core-loop-port-2026-07-02.md`) should be revised to incorporate SkillForge's Diagnostician→Optimizer pipeline (replace EvoSkill's simpler Proposer→Builder) and OpenSkill's virtual verifier for no-regression guarantees plus Ctx2Skill's Cross-time Replay for skill selection.
