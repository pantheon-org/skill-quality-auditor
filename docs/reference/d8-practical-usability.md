# D8: Practical Usability (15 points)

**Purpose:** Ensure the skill is immediately useful with clear, outcome-linked examples.

**Scoring:**

| Points | Signal |
| ------ | ------ |
| 13–15 | Concrete + runnable + clear + outcome-linked |
| 10–12 | Most examples good, some outcome indicators |
| 7–9 | Examples present but missing outcome verification |
| 0–6 | Abstract or missing |

## Components

Implemented in `scorer/d8_practical_usability.go` (`scoreD8`). A base score of 5 points is
always awarded; the components below add up to 13 more, capped at 15.

### 1. Code Blocks (up to 6 points)

- More than 5 code blocks → +4; more than 2 → +2; any at all → +1
- At least one language-tagged code fence → +2

### 2. Runnable Command (4 points)

- Content contains an invocable command signal (`./`, `npm run`, `go run`, `make`, etc.)

### 3. Outcome Linkage (up to 3 points)

- Each fenced code block plus its immediately following prose is checked for a
  verifiable-outcome phrase (`# output:`, `expected:`, `should return`, `→`, `assert`, etc.)
- All segments linked → 3 points; at least half → 2 points; any → 1 point; none → 0
- Rationale: proxy metrics like code-block count correlate poorly with actual task
  completion — see Miller & Tang (arXiv:2505.08253) and the Mohammadi et al. survey below

## Academic References

```bibtex
@article{jiang2026buildrix,
  title         = {Buildrix: An Open Platform for Sharing and Benchmarking Agentic AI Skills in Building Engineering},
  author        = {Z. Jiang and B. Dong},
  year          = {2026},
  journal       = {arXiv preprint arXiv:2606.25139},
  eprint        = {2606.25139},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2606.25139>}
}

```

```bibtex
@inproceedings{stolee2025code,
  title         = {10 Years Later: Revisiting How Developers Search for Code},
  author        = {K. T. Stolee and T. Welp and C. Sadowski and S. Elbaum},
  year          = {2025},
  booktitle     = {Proceedings of the ACM on Software Engineering},
  volume        = {2},
  publisher     = {ACM},
  url           = {<https://dl.acm.org/doi/10.1145/3715774>}
}

```

```bibtex
@article{he2025bugs,
  title         = {A Survey of Bugs in AI-Generated Code},
  author        = {R. Gao and A. Tahir and P. Liang and T. Susnjak and others},
  year          = {2025},
  journal       = {arXiv preprint arXiv:2512.05239},
  eprint        = {2512.05239},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2512.05239>}
}

```

```bibtex
@article{abufarha2026prompt,
  title         = {Mitigating Prompt Dependency in Large Language Models: A Retrieval-Augmented Framework for Intelligent Code Assistance},
  author        = {S. Abufarha and A. A. Marouf and J. G. Rokne and R. Alhajj},
  year          = {2026},
  journal       = {Software},
  volume        = {5},
  number        = {1},
  publisher     = {MDPI},
  url           = {<https://www.mdpi.com/2674-113X/5/1/4>}
}
```

```bibtex
@article{mohammadi2025evalbench,
  title         = {Evaluation and Benchmarking of LLM Agents: A Survey},
  author        = {Mohammadi and others},
  year          = {2025},
  journal       = {ACM Computing Surveys},
  publisher     = {ACM},
  url           = {<https://dl.acm.org/doi/abs/10.1145/3711896.3736570}>
}

```

```bibtex
@article{miller2025metrics,
  title         = {Evaluating LLM Metrics Through Real-World Capabilities},
  author        = {Miller and Tang},
  year          = {2025},
  journal       = {arXiv preprint arXiv:2505.08253},
  eprint        = {2505.08253},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2505.08253}>
}

```

```bibtex
@article{yehudai2025survey,
  title         = {Survey on Evaluation of LLM-Based Agents},
  author        = {Yehudai and others},
  year          = {2025},
  journal       = {arXiv preprint arXiv:2503.16416},
  eprint        = {2503.16416},
  archivePrefix = {arXiv},
  url           = {<https://arxiv.org/abs/2503.16416}>
}
```
