# D7: Pattern Recognition (10 points)

**Purpose:** Ensure the skill activates when needed via description keywords and trigger conditions.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 9–10 | Rich keywords, comprehensive triggers |
| 7–8 | Good keywords, could expand |
| 5–6 | Basic keywords |
| 0–4 | Missing or poor |

## Requirements

- Description must include domain keywords
- Trigger scenarios in the description or a "When to Apply" section
- Example: "Use when writing BDD tests, feature files, Gherkin scenarios…"

The best description = exhaustive trigger list + concrete examples.

## Discriminativeness (diagnostic signal)

A high-quality description reduces false positives by anchoring the skill to specific contexts:

1. **Negative anchor** — explicitly states when NOT to activate
   (e.g., `Does not apply`, `SKIP when`, `Not for`, `Exclude`, `DO NOT trigger`, `not intended for`)

2. **Workflow anchor** — trigger tied to a concrete artifact or action
   (e.g., references `file`, `PR`, `commit`, `test`, `config`, `pipeline`, `migration`)

| Anchors present | Diagnostic |
| --------------- | ---------- |
| Both | INFO — positive signal |
| Neither | WARN — may over-trigger on adjacent topics |
| One | No diagnostic |

This is a diagnostic signal only — it does not affect the numeric score in the current iteration.

## Academic References

```bibtex
@article{zhang2025agentrouter,
  title         = {AgentRouter: A Knowledge-Graph-Guided LLM Router for Collaborative Multi-Agent Question Answering},
  author        = {Zhang and others},
  year          = {2025},
  journal       = {arXiv preprint arXiv:2510.05445},
  eprint        = {2510.05445},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2510.05445}>
}

```

```bibtex
@article{wang2026aco,
  title         = {Efficient and Interpretable Multi-Agent LLM Routing via Ant Colony Optimization},
  author        = {Wang and others},
  year          = {2026},
  journal       = {arXiv preprint arXiv:2603.12933},
  eprint        = {2603.12933},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2603.12933}>
}

```

```bibtex
@article{yehudai2025survey,
  title         = {Survey on Evaluation of LLM-Based Agents},
  author        = {A. Yehudai and L. Eden and A. Li and G. Uziel and Y. Zhao and R. Bar-Haim and A. Cohan and M. Shmueli-Scheuer},
  year          = {2025},
  journal       = {arXiv preprint arXiv:2503.16416},
  eprint        = {2503.16416},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2503.16416>}
}

```

```bibtex
@inproceedings{chen2024agentpoison,
  title         = {AgentPoison: Red-teaming LLM Agents via Poisoning Memory or Knowledge Bases},
  author        = {Z. Chen and Z. Xiang and C. Xiao and D. Song and B. Li},
  year          = {2024},
  booktitle     = {Advances in Neural Information Processing Systems (NeurIPS 2024)},
  url           = {<https://proceedings.neurips.cc/paper_files/paper/2024/hash/eb113910e9c3f6242541c1652e30dfd6-Abstract-Conference.html>}
}
```
