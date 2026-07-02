# Site Layout

The documentation site source lives under `docs/` and is built to `./site/` by `@docmd/core`.

## Source structure

```text
docs/
├── index.md                       # Homepage
├── ADR/                           # Architecture Decision Records
│   ├── index.yaml                 # Machine-readable ADR index
│   ├── adr-001-native-eval-runner.md
│   ├── adr-002-...
│   └── ...
├── architecture/                  # Code flow documentation
│   ├── overview.md
│   ├── evaluate-flow.md
│   ├── batch-flow.md
│   ├── duplication-flow.md
│   ├── aggregation-flow.md
│   ├── remediation-flow.md
│   ├── trend-flow.md
│   ├── eval-runner.md
│   ├── init-update-prune.md
│   └── validate-analyze.md
├── development/                   # Developer guides
│   ├── setup.md
│   ├── adding-a-scorer.md
│   ├── adding-an-agent.md
│   └── skills-and-rules.md
├── reference/                     # Scoring dimension references
│   ├── scoring-dimensions.md
│   ├── d1-knowledge-delta.md
│   ├── d2-mindset-procedures.md
│   ├── d3-anti-pattern-coverage.md
│   ├── d4-specification-compliance.md
│   ├── d5-progressive-disclosure.md
│   ├── d6-freedom-calibration.md
│   ├── d7-pattern-recognition.md
│   ├── d8-practical-usability.md
│   └── d9-eval-validation.md
└── assets/
    └── logo.png
```

## Navigation conventions

- `docs/index.md` is the homepage and the primary navigation hub
- All pages should be linked from `docs/index.md` or a subsection index
- ADR pages are self-organising — the discovery mechanism is `docs/ADR/index.yaml`
- Subsections (`architecture/`, `development/`, `reference/`) do not have their own index pages; all links are centralised in `docs/index.md`

## Output structure

The built site (`./site/`) mirrors the source with `.html` extensions:

```text
site/
├── index.html
├── 404.html
├── robots.txt
├── sitemap.xml
├── search-index.json
├── llms.txt               # LLM-friendly entry index
├── llms-full.txt          # Full LLM text dump
├── llms.json              # LLM JSON index
├── assets/                # CSS, JS, images
├── ADR/
├── architecture/
├── development/
└── reference/
```
