# docmd Configuration

The documentation site is built with `@docmd/core`. Configuration lives in `docmd.config.json` at the repo root.

## Current config

```json
{
  "title": "skill-quality-auditor",
  "url": "https://pantheon-org.github.io/skill-quality-auditor",
  "src": "./docs",
  "out": "./site",
  "theme": {
    "primaryColor": "#CB3837",
    "favicon": "https://raw.githubusercontent.com/pantheon-org/skill-quality-auditor/main/assets/favicon.ico"
  },
  "git": { "enabled": true },
  "llms": { "enabled": true },
  "seo": { "enabled": true, "description": "A Go CLI that scores AI skill definitions against a 9-dimension quality framework" }
}
```

## Key fields

| Field | Value | Purpose |
| --- | --- | --- |
| `src` | `./docs` | Source directory for markdown content |
| `out` | `./site` | Output directory for the built static site |
| `git.enabled` | `true` | Adds source links to every page |
| `llms.enabled` | `true` | Generates `llms.txt`, `llms-full.txt`, `llms.json` |
| `seo.enabled` | `true` | Generates sitemap, robots.txt, search index |

## Commands

| Command | Purpose |
| --- | --- |
| `npx @docmd/core build` | Build the static site to `./site/` |
| `npx @docmd/core dev` | Start dev server with hot reload at `localhost:3000` |

## Deployment

The site deploys automatically via GitHub Actions (`.github/workflows/docs.yml`) on push to `main` touching `docs/**` or `docmd.config.json`.
