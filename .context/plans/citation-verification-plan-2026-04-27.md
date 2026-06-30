---
title: "Citation Verification Plan"
type: plan
status: active
date: 2026-04-27
---
# Citation Verification Plan

**Goal:** Verify that every academic citation in `docs/d*.md` (a) exists, (b) has correct metadata, and (c) actually supports the claim made in the dimension doc.

**Total citations:** 28 (including duplicates across D1–D9)

---

## Per-citation checks

| Check | Question |
| ----- | -------- |
| Existence | Does the paper resolve at the cited URL? |
| Attribution | Are title, authors, and year correct? |
| Claim fidelity | Does the paper's content support how we use it? |

---

## Phases

### Phase 1 — Existence + metadata (Google Scholar)

Use the `tessl__google-scholar-search` skill to query each paper by title.

- Confirms: real paper, correct authors/year, real venue
- Flag: no match, wrong metadata, URL dead

### Phase 2 — Claim fidelity (in-session, notebooklm-mcp-secure)

All citations appear under an `## Academic References` section in each dimension doc with no inline claim text — the link between paper and dimension rationale is implicit. We verify it in-session via `@pan-sec/notebooklm-mcp` ([Pantheon-Security/notebooklm-mcp-secure](https://github.com/Pantheon-Security/notebooklm-mcp-secure)), a security-hardened fork of the upstream that adds `create_notebook` (eliminating the manual browser step) and 17 security layers.

**Prerequisite — install and authenticate the MCP (one-time)**

Add the server to `.mcpx.json` via `mcpx add`, which syncs it to all configured providers (claude-code, gemini-cli, openai-codex, opencode):

```bash
mcpx add notebooklm
# When prompted:
#   transport: stdio
#   command:   npx
#   args:      @pan-sec/notebooklm-mcp@latest
#   env:       NLMCP_AUTH_ENABLED=true
#              NLMCP_AUTH_TOKEN=<output of: openssl rand -base64 32>
```

Then authenticate once with your Google account:

```bash
setup_auth(show_browser=true)
```

**Step 1 — Create one notebook per dimension**

```
create_notebook(name: "D1 — Knowledge Delta")
create_notebook(name: "D2 — Mindset & Procedures")
# ... repeat for D3–D9
```

No manual browser steps required.

**Step 2 — Add the dimension doc as a source**

```
select_notebook(name: "D1 — Knowledge Delta")
add_source(type: "text", content: "<full text of docs/d1-knowledge-delta.md>", title: "Dimension doc")
```

**Step 3 — Add each paper as a URL source**

For each paper that passed Phase 1:

```
add_source(type: "url", url: "<paper URL>", title: "<short paper title>")
```

NotebookLM crawls the URL and ingests the full content. Works for arxiv, ACM DL, SSRN, MDPI, and ResearchSquare. For paywalled ACM full-text, the abstract page is sufficient.

**Step 4 — Query claim fidelity**

Once all sources are loaded, run two questions per dimension:

```
ask_question(question: "Based on the dimension doc, what are the core claims or design decisions that need academic backing? For each paper source, does it provide direct evidence for any of those claims? Quote the specific finding. If the paper does not address a claim at all, say so explicitly.")
```

```
ask_question(question: "Are there any claims or scoring thresholds in the dimension doc that none of the paper sources support? List them.")
```

NotebookLM returns grounded answers with citations pinned to the specific source, making it straightforward to see which papers are load-bearing vs decorative.

**What to record per paper:**

| Result | Meaning |
| ------ | ------- |
| Answer cites the paper with a direct finding | Citation valid — no action |
| Answer cites the paper tangentially | Flag as weak — note which claim remains unsupported |
| Answer does not cite the paper at all | Citation is decorative — remove or replace |
| Answer contradicts the paper | Citation is invalid — rewrite the rationale |

### Phase 3 — Remediation

| Failure mode | Action |
| ------------ | ------ |
| Paper doesn't exist | Remove or replace citation |
| Wrong metadata | Fix authors / year / title in `docs/dN-*.md` |
| Claim unsupported | Rewrite surrounding rationale or find a better paper |

---

## Execution order

Start with highest-risk papers first:

1. **2026-dated papers** (6 papers — post-cutoff, highest hallucination risk)
2. **arxiv / ACM / NeurIPS / ACL papers** (easiest URL verification)
3. **SSRN / ResearchGate / ResearchSquare preprints** (highest miss rate)

---

## Full citation inventory

### 2026 papers (verify first)

| Dimension | Citation | URL |
| --------- | -------- | --- |
| D9 | Rehan, 2026 — Test-Driven AI Agent Definition (TDAD) | https://arxiv.org/abs/2603.08806 |
| D9 | Alami, 2026 — Cognitive Camouflage: Specification Gaming | https://papers.ssrn.com/sol3/papers.cfm?abstract_id=6512960 |
| D1 | Bi, Wu, Hao et al., 2026 — Automating Skill Acquisition | https://arxiv.org/abs/2603.11808 |
| D1 | Bakal, 2026 — Knowledge Activation: AI Skills as the Institutional Knowledge Primitive | https://arxiv.org/abs/2603.14805 |
| D7 | Wang et al., 2026 — Efficient and Interpretable Multi-Agent LLM Routing | https://arxiv.org/abs/2603.12933 |
| D6 | Sorensen, 2026 — Specification as the New Management | https://www.researchgate.net/publication/401626622 |

### arxiv / ACM / NeurIPS papers

| Dimension | Citation | URL |
| --------- | -------- | --- |
| D1 | Li et al., 2025 — Instruction Agent: Enhancing Agent with Expert Demonstration | https://arxiv.org/abs/2509.07098 |
| D1 | Bi, Hu, Nasir, 2025 — Real-Time Procedural Learning From Experience | https://arxiv.org/abs/2511.22074 |
| D2 | Deng et al., 2024 — From Novice to Expert: LLM Agent Policy Optimization | https://arxiv.org/abs/2411.03817 |
| D2 | Xu et al., 2024 — TheAgentCompany: Benchmarking LLM Agents | https://arxiv.org/abs/2412.14161 |
| D3 | Bhatia, Lin, Rajbahadur, Adams et al., 2024 — Data Quality Anti-Patterns | https://arxiv.org/abs/2408.12560 |
| D3 | Chen et al., NeurIPS 2024 — AgentPoison: Red-teaming LLM Agents | https://proceedings.neurips.cc/paper_files/paper/2024/hash/eb113910e9c3f6242541c1652e30dfd6-Abstract-Conference.html |
| D4 | Zhang et al., 2025 — Reasoning over Boundaries: Specification Alignment | https://arxiv.org/abs/2509.14760 |
| D5 | Springer & Whittaker, 2018 — Progressive Disclosure: Designing for Effective Transparency | https://arxiv.org/abs/1811.02164 |
| D5 | Anik & Bunt, 2021 — Designing Effective Training Dataset Explanations | https://dl.acm.org/doi/10.1145/3411764.3445382 |
| D7 | Zhang et al., 2025 — AgentRouter: Knowledge-Graph-Guided LLM Router | https://arxiv.org/abs/2510.05445 |
| D7 | Yehudai et al., 2025 — Survey on Evaluation of LLM-Based Agents | https://arxiv.org/abs/2503.16416 |
| D7 | Miller & Tang, 2025 — Evaluating LLM Metrics Through Real-World Capabilities | https://arxiv.org/abs/2505.08253 |
| D8 | Mohammadi et al., 2025 — Evaluation and Benchmarking of LLM Agents: A Survey | https://dl.acm.org/doi/abs/10.1145/3711896.3736570 |
| D9 | Wang, Chen, Deng, Lin, Harman et al. — LLMs for Mutation Testing | https://dl.acm.org/doi/abs/10.1145/3805038 |
| D9 | Pan, Hu, Xia, Yang — Re-Evaluating Code LLM Benchmarks Under Semantic Mutation | https://arxiv.org/abs/2506.17369 |
| D9 | Bouafif, Hamdaqa, Zulkoski — PrimG: LLM-Driven Test Generation | https://dl.acm.org/doi/abs/10.1145/3756681.3756991 |
| D4 | J Kohl, O Kruse, Y Mostafa, A Luckow et al. — Automated Structural Testing | *(URL needed — partial citation in doc)* |

### SSRN / ResearchGate / ResearchSquare preprints

| Dimension | Citation | URL |
| --------- | -------- | --- |
| D5 | Timileyin, 2025 — Cognitive Load in Shaping Web Usability Requirements | https://papers.ssrn.com/sol3/papers.cfm?abstract_id=5247018 |
| D5 | Pastrakis, Konstantakis, Caridakis, 2025 — AI-Enhanced Modular Information Architecture | https://www.mdpi.com/2078-2489/17/1/92 |
| D6 | R Tao, 2025 — LLM-Skill Orchestration: 202/202 Subtask Completion | https://www.researchsquare.com/article/rs-9323974/latest |
| D8 | Yin et al., 2025 — Grounding Open-Domain Knowledge from LLMs to Real-World RL Tasks | https://www.ijcai.org/proceedings/2025/1198.pdf |

### ACM (non-arxiv)

| Dimension | Citation | URL |
| --------- | -------- | --- |
| D3 | Amarasinghe, Asanka et al., 2024 — Code Quality Alarms | https://jdrra.sljol.info/articles/10.4038/jdrra.v3i2.93 |
| D3 | Brada & Picha, 2019 — Software Process Anti-Patterns Catalogue | https://dl.acm.org/doi/abs/10.1145/3361149.3361178 |
| D3 | Picha & Brada, 2019 — Software Process Anti-Pattern Detection | https://dl.acm.org/doi/abs/10.1145/3361149.3361169 |

---

## Notes

- D9 has a duplicate: Rehan 2026 / T Rehan 2026 — same paper cited twice. Deduplicate after verification.
- D6 has a duplicate: R Tao 2025 / Tao 2025 — same ResearchSquare preprint. Deduplicate after verification.
- D4 Kohl et al. citation is incomplete (no URL in doc) — needs manual search.
