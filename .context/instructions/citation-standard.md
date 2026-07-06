---
title: "Citation Standard"
type: INSTRUCTION
status: ACTIVE
date: 2026-06-30
---

All academic references in dimension docs MUST use BibTeX format inside individually fenced code blocks.

## Format

Each BibTeX entry is wrapped in its own ````bibtex` ... ```` fence:

````markdown
```bibtex
@article{lastnameYYYYkeyword,
  title         = {Full Title of Paper},
  author        = {Author and Others},
  year          = {2025},
  journal       = {arXiv preprint arXiv:XXXX.XXXXX},
  eprint        = {XXXX.XXXXX},
  archivePrefix = {arXiv},
  url           = {https://arxiv.org/abs/XXXX.XXXXX}
}
```
````

## Rules

| Rule | Example |
|------|---------|
| One fence per entry | ````bibtex ... ```` around each `@article{}` or `@inproceedings{}` |
| Citation key | `{lastnameYYYYkeyword}` — lowercase, author surname + year + short keyword |
| arXiv papers | Include `eprint`, `archivePrefix`, `journal = {arXiv preprint...}` |
| Conference papers | Use `@inproceedings` with `booktitle`, omit `eprint`/`archivePrefix` |
| Journal papers | Use `@article` with `journal` name, omit `eprint` if not a preprint |
| URLs | Always include `url` field with DOI or direct link |
| Missing years | Omit `year` field entirely rather than guessing |
| Section heading | Exactly `## Academic References` — the last section in the file |

## Section placement

The `## Academic References` heading MUST be the last section in the dimension doc, after all scoring criteria and examples. References are ordered by relevance to the dimension, not alphabetically.
