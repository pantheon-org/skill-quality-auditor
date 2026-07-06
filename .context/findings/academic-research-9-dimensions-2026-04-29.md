---
title: "Academic Research Findings: 9-Dimension Quality Framework"
type: finding
status: active
date: 2026-04-29
value: low
---
# Academic Research Findings: 9-Dimension Quality Framework

**Date:** 2026-04-29  
**Source:** Google Scholar searches via `pantheon-ai/research@0.2.4` `google-scholar-search` skill  
**Purpose:** Literature review to identify improvements to the skill quality scoring framework (D1–D9).

> **Security note:** The `google-scholar-search` script scrapes Scholar pages directly (W011 advisory).
> Results below are candidates for triage — verify each before acting on them.

---

## Papers by Dimension

### D1 — Knowledge Delta (20 pts)

| Title | Authors | Link |
|---|---|---|
| Instruction Agent: Enhancing Agent with Expert Demonstration | Y Li, H Hultquist, J Wagle, K Koishida | https://arxiv.org/abs/2509.07098 |
| From Novice to Expert: LLM Agent Policy Optimization via Step-wise Reinforcement Learning | Z Deng, Z Dou, Y Zhu et al. | https://arxiv.org/abs/2411.03817 |
| Grounding Open-Domain Knowledge from LLMs to Real-World RL Tasks: A Survey | H Yin, H Qian, Y Shi et al. | https://www.ijcai.org/proceedings/2025/1198.pdf |
| KG-Agent: An Efficient Autonomous Agent Framework for Complex Reasoning over Knowledge Graph | J Jiang, K Zhou, WX Zhao et al. | https://aclanthology.org/2025.acl-long.468/ |

**Signal:** "Instruction Agent" shows that injecting expert demonstrations into agent instructions measurably improves task performance — directly validating that D1's *knowledge delta* concept has outcome-level impact. "From Novice to Expert" provides a step-wise reinforcement learning lens: skill quality should reflect progressive expertise, not flat knowledge dumps.

**Suggested improvement:** Add a **demonstration concreteness** sub-criterion — does the skill include at least one concrete worked example that an agent cannot derive from pretraining alone? This operationalises the knowledge delta from a content check to an outcome proxy.

---

### D2 — Mindset + Procedures (15 pts)

| Title | Authors | Link |
|---|---|---|
| Knowledge Activation: AI Skills as the Institutional Knowledge Primitive for Agentic Software Development | G Bakal | https://arxiv.org/abs/2603.14805 |
| Automating Skill Acquisition through Large-Scale Mining of Open-Source Agentic Repositories | S Bi, M Wu, H Hao et al. | https://arxiv.org/abs/2603.11808 |
| Real-Time Procedural Learning From Experience for AI Agents | D Bi, Y Hu, MN Nasir | https://arxiv.org/abs/2511.22074 |
| Procedural Knowledge Ontology (PKO) | VA Carriero, M Scrocca et al. | https://link.springer.com/chapter/10.1007/978-3-031-94578-6_19 |

**Signal:** "Knowledge Activation" (Bakal 2026) argues that skills are the *institutional knowledge primitive* — they encode the mindset and procedures that make organisational expertise reusable. The PKO paper establishes a formal ontology for procedural knowledge with components that map onto D2: *preconditions*, *steps*, *postconditions*, and *decision points*. "Real-Time Procedural Learning" shows agents drift when procedures lack external checkpoints.

**Suggested improvement:** Align D2 scoring against the PKO model — score whether the procedures include explicit **preconditions** (when to start), **decision points** (when to branch), and **postconditions** (how to verify completion). Currently D2 only checks for step-by-step presence, not structural completeness.

---

### D3 — Anti-Pattern Coverage (15 pts)

| Title | Authors | Link |
|---|---|---|
| Software Process Anti-Patterns Catalogue | P Brada, P Picha | https://dl.acm.org/doi/abs/10.1145/3361149.3361178 |
| Software Process Anti-Pattern Detection in Project Data | P Picha, P Brada | https://dl.acm.org/doi/abs/10.1145/3361149.3361169 |
| Data Quality Anti-Patterns for Software Analytics | A Bhatia, D Lin, GK Rajbahadur, B Adams et al. | https://arxiv.org/abs/2408.12560 |
| Code Quality Alarms: Techniques, Datasets, and Emerging Trends in Detecting Smells and Anti-Patterns | YVA Amarasinghe, P Asanka et al. | https://jdrra.sljol.info/articles/10.4038/jdrra.v3i2.93 |

**Signal:** Brada & Picha's anti-pattern catalogue establishes that anti-patterns must document: *context* (when the pattern emerges), *symptoms* (observable signals), *root cause*, and *refactored solution*. The data quality anti-patterns paper adds that anti-patterns without consequence statements are rarely acted on in practice — engineers need to know *what goes wrong*, not just *what is wrong*.

**Suggested improvement:** Extend the NEVER/WHY format to **NEVER / WHY / SYMPTOM / FIX** — the symptom (what the agent will observe if the anti-pattern occurs) and a brief refactored alternative. This aligns with the canonical anti-pattern catalogue structure and makes anti-patterns actionable rather than advisory.

---

### D4 — Specification Compliance (15 pts)

| Title | Authors | Link |
|---|---|---|
| Test-Driven AI Agent Definition (TDAD): Compiling Tool-Using Agents from Behavioral Specifications | T Rehan | https://arxiv.org/abs/2603.08806 |
| Agentic AI for Behaviour-Driven Development Testing Using Large Language Models | C Paduraru, M Zavelca, A Stefanescu | https://www.researchgate.net/... |
| LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Multi-Model Collaboration | R Tao | https://www.researchsquare.com/article/rs-9323974/latest |
| Automated Structural Testing of LLM-Based Agents: Methods, Framework, and Case Studies | J Kohl, O Kruse, Y Mostafa, A Luckow et al. | https://ieeexplore.ieee.org/abstract/document/11401679/ |

**Signal:** TDAD (Rehan 2026) is the strongest result: treats agent specs as *compiled artifacts*, validates via visible/hidden test splits and semantic mutation testing. Achieves 97% regression safety. The BDD-testing paper reinforces that natural-language specs (like SKILL.md) are more reliably validated via executable Given/When/Then scenarios than structural inspection alone.

**Suggested improvement:** Add **mutation resistance** as a D4 criterion: a specification is compliant only if removing or inverting one instruction would produce a detectably different agent. This turns D4 from a format check into a behavioural signal. TDAD's open benchmark (https://github.com/f-labs-io/tdad-paper-code) provides a ready reference implementation.

---

### D5 — Progressive Disclosure (15 pts)

| Title | Authors | Link |
|---|---|---|
| Progressive Disclosure: Designing for Effective Transparency | A Springer, S Whittaker | https://arxiv.org/abs/1811.02164 |
| The Role of Cognitive Load in Shaping Web Usability Requirements | A Timileyin | https://papers.ssrn.com/sol3/papers.cfm?abstract_id=5247018 |
| Designing Effective Training Dataset Explanations: The Impact of Information Depth and Progressive Disclosure | AI Anik, A Bunt | (ACM CHI proceedings) |
| AI-Enhanced Modular Information Architecture for Cognitive-Efficient User Experiences | F Pastrakis, M Konstantakis, G Caridakis | https://www.mdpi.com/2078-2489/17/1/92 |

**Signal:** Springer & Whittaker (arXiv:1811.02164) is the canonical progressive disclosure paper for AI systems — defines transparency layers and when to expose each. Anik & Bunt's dataset explanation study shows empirically that progressive disclosure *only* reduces cognitive load when disclosure conditions are explicit (users must know when to go deeper, not just that deeper content exists). Passive availability is insufficient.

**Suggested improvement:** Add a **disclosure trigger precision** sub-criterion: lazy-load conditions must specify *when NOT to load* a reference, not just when to load it. The framework already shows this in examples; making it scoreable would penalise vague "load for scoring" conditions that provide passive availability without active guidance.

---

### D6 — Freedom Calibration (15 pts)

| Title | Authors | Link |
|---|---|---|
| Reasoning over Boundaries: Enhancing Specification Alignment via Test-time Deliberation | H Zhang, Y Li, X Hu et al. | https://arxiv.org/abs/2509.14760 |
| Specification as the New Management | S Sorensen | https://www.researchgate.net/publication/401626622 |
| LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Collaboration | R Tao | https://www.researchsquare.com/article/rs-9323974/latest |

**Signal:** "Reasoning over Boundaries" (Zhang et al.) shows that specification alignment degrades when constraints are applied uniformly — agents need to distinguish *hard boundaries* (never cross) from *soft preferences* (default but overridable). "Specification as the New Management" argues that the key enterprise competency is *pre-delegation architecture*: knowing which decisions to pre-specify vs. which to leave to agent judgement. LLM-Skill Orchestration demonstrates that rule-augmented agents outperform unconstrained ones by large margins on structured tasks, but shows diminishing returns on open-ended tasks.

**Suggested improvement:** Add a **constraint typology** criterion: does the skill explicitly distinguish *hard constraints* (MUST/NEVER) from *soft defaults* (PREFER/AVOID)? Agents that receive only hard constraints over-refuse on edge cases; agents with no hard constraints drift. Scoring the presence of both types catches both failure modes.

---

### D7 — Pattern Recognition (10 pts)

| Title | Authors | Link |
|---|---|---|
| AgentRouter: A Knowledge-Graph-Guided LLM Router for Collaborative Multi-Agent QA | Z Zhang, K Shi, Z Yuan et al. | https://arxiv.org/abs/2510.05445 |
| Efficient and Interpretable Multi-Agent LLM Routing via Ant Colony Optimization | X Wang, C Zhang, J Zhang et al. | https://arxiv.org/abs/2603.12933 |
| AgentPoison: Red-Teaming LLM Agents via Poisoning Memory or Knowledge Bases | Z Chen, Z Xiang, C Xiao et al. | https://proceedings.neurips.cc/paper_files/paper/2024/hash/eb113910e9c3f6242541c1652e30dfd6-Abstract-Conference.html |

**Signal:** AgentRouter shows that knowledge-graph-guided routing outperforms keyword matching for skill selection — semantic context matters, not just surface triggers. AgentPoison demonstrates that triggers are an attack surface: adversarially crafted inputs can hijack skill routing. The routing literature consistently shows that triggers must be *discriminative* (fire on target contexts, not fire on related-but-wrong contexts).

**Suggested improvement:** Add a **discriminativeness** criterion: triggers should be testable for false positives — does the trigger avoid firing on at least 2 plausibly-related but incorrect contexts? Currently D7 only scores keyword presence and count. Discriminativeness scoring would catch triggers that are too broad (fire on everything) or too narrow (never fire).

---

### D8 — Practical Usability (15 pts)

| Title | Authors | Link |
|---|---|---|
| TheAgentCompany: Benchmarking LLM Agents on Consequential Real-World Tasks | FF Xu, Y Song, B Li et al. | https://arxiv.org/abs/2412.14161 |
| Evaluation and Benchmarking of LLM Agents: A Survey | M Mohammadi, Y Li, J Lo, W Yip | https://dl.acm.org/doi/abs/10.1145/3711896.3736570 |
| Evaluating LLM Metrics Through Real-World Capabilities | JK Miller, W Tang | https://arxiv.org/abs/2505.08253 |
| Survey on Evaluation of LLM-Based Agents | A Yehudai, L Eden, A Li et al. | https://arxiv.org/abs/2503.16416 |

**Signal:** TheAgentCompany (Xu et al. 2024) is the benchmark of record for real-world agent task evaluation — tasks span software engineering, data analysis, and admin work, scored on *task completion* not surface quality. The survey papers consistently identify that usability benchmarks fail when they measure proxy metrics (example quality, response length) rather than outcome completion. "Evaluating LLM Metrics Through Real-World Capabilities" shows that common proxy metrics correlate poorly with actual task success.

**Suggested improvement:** Add an **outcome linkage** criterion: each example in the skill should specify the *verifiable artifact* that proves success (a file, a passing test, a diff, a log entry). Current D8 rewards concrete examples; outcome linkage requires the example to be falsifiable — the reader should be able to determine success or failure from a concrete observable.

---

### D9 — Eval Validation (20 pts)

| Title | Authors | Link |
|---|---|---|
| A Comprehensive Study on Large Language Models for Mutation Testing | B Wang, M Chen, M Deng, Y Lin, M Harman et al. | https://dl.acm.org/doi/abs/10.1145/3805038 |
| Cognitive Camouflage: Specification Gaming in LLM-Generated Code Evades Holistic Evaluation but Not Adversarial Execution | D Alami | https://papers.ssrn.com/sol3/papers.cfm?abstract_id=6512960 |
| Re-Evaluating Code LLM Benchmarks Under Semantic Mutation | Z Pan, X Hu, X Xia, X Yang | https://arxiv.org/abs/2506.17369 |
| PrimG: Efficient LLM-Driven Test Generation Using Mutant Prioritization | MS Bouafif, M Hamdaqa, E Zulkoski | https://dl.acm.org/doi/abs/10.1145/3756681.3756991 |
| Test-Driven AI Agent Definition (TDAD) | T Rehan | https://arxiv.org/abs/2603.08806 |

**Signal:** "Cognitive Camouflage" (Alami 2026) is the most directly relevant — documents how LLM-generated specs game holistic evaluations but fail adversarial execution. The mutation testing literature (Wang et al., Pan et al.) establishes that benchmarks only have validity if a semantically-wrong variant *fails* them. PrimG shows that not all mutations are equal; prioritising mutations that target instruction-level changes is more efficient than random perturbation.

**Suggested improvement (highest priority):**
1. **Mutation score requirement**: at least one eval scenario must fail when a single instruction in SKILL.md is removed or inverted. Currently D9 only checks eval *existence*.
2. **Independent authoring signal**: evals authored after the skill (or by a different person) are stronger validation than evals written alongside the skill — add this as a bonus criterion.
3. **Adversarial execution**: at least one scenario should test an edge case or boundary condition, not just the happy path.

---

## Priority Summary

| Priority | Dimension | Improvement | Effort |
|---|---|---|---|
| 1 | **D9** | Mutation score requirement + adversarial scenario | High |
| 2 | **D4** | Mutation resistance criterion (behavioural compliance) | Medium |
| 3 | **D3** | Extend to NEVER/WHY/SYMPTOM/FIX (canonical anti-pattern structure) | Low |
| 4 | **D2** | PKO-aligned scoring: preconditions + decision points + postconditions | Low |
| 5 | **D6** | Constraint typology: hard constraints vs. soft defaults | Low |
| 6 | **D1** | Demonstration concreteness sub-criterion | Low |
| 7 | **D7** | Discriminativeness criterion for triggers | Low |
| 8 | **D8** | Outcome linkage for examples | Low |
| 9 | **D5** | Score "when NOT to load" trigger precision | Low |
