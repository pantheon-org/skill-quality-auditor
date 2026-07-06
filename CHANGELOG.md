# Changelog

## [0.24.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.23.0...v0.24.0) (2026-07-06)


### Features

* **context:** implement themes taxonomy (4 phases, ADR-051) ([#208](https://github.com/pantheon-org/skill-quality-auditor/issues/208)) ([3b5f9bb](https://github.com/pantheon-org/skill-quality-auditor/commit/3b5f9bbf655108e85ade5f44a0fff66028e004e9))
* **validate:** real JSON-schema validator for .context frontmatter (G5) ([#215](https://github.com/pantheon-org/skill-quality-auditor/issues/215)) ([6f490c7](https://github.com/pantheon-org/skill-quality-auditor/commit/6f490c72697e3dfcbf2b76e7498f1b0a5d88c13b))


### Bug Fixes

* **governance:** adr-index freshness gate (G2) + undocumented-decisions false-negative (G3) ([#211](https://github.com/pantheon-org/skill-quality-auditor/issues/211)) ([397e0b3](https://github.com/pantheon-org/skill-quality-auditor/commit/397e0b3c4608021c8414830f970c8610b1b5b5fd))
* **release:** attach artifacts via draft-then-publish flow (immutable releases) ([#219](https://github.com/pantheon-org/skill-quality-auditor/issues/219)) ([147ab54](https://github.com/pantheon-org/skill-quality-auditor/commit/147ab54af8f379b7264d080d6b4c3685e0c98f98))
* **reporter:** generated remediation plans carry value + themes (G4) ([#212](https://github.com/pantheon-org/skill-quality-auditor/issues/212)) ([c44acab](https://github.com/pantheon-org/skill-quality-auditor/commit/c44acab00c6c727e6247c875af4014501bab1395))
* **validate:** accept a path for 'validate context' instead of hardcoding .context ([#216](https://github.com/pantheon-org/skill-quality-auditor/issues/216)) ([643f6ab](https://github.com/pantheon-org/skill-quality-auditor/commit/643f6ab6bc2be85e41dc5a497a3fd8143d6fdf2c))

## [0.23.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.22.0...v0.23.0) (2026-07-06)


### Features

* CI check for .tessl/plugins/pantheon-org mirror drift ([#199](https://github.com/pantheon-org/skill-quality-auditor/issues/199)) ([333e9c0](https://github.com/pantheon-org/skill-quality-auditor/commit/333e9c00b25f3ff662e6dacf7e8153a9e6bc1246))
* **ci:** gate PRs on newly introduced docs drift ([#185](https://github.com/pantheon-org/skill-quality-auditor/issues/185)) ([6784b2f](https://github.com/pantheon-org/skill-quality-auditor/commit/6784b2f06f2971ec0994d94bfc7754bde089fd9b))
* **context:** known-issue as a first-class context file type, driven by session-reflection ([#188](https://github.com/pantheon-org/skill-quality-auditor/issues/188)) ([a19612f](https://github.com/pantheon-org/skill-quality-auditor/commit/a19612f2dd0f4d3c78735af1f1b2176a39a75fc8))
* **context:** value prioritisation signal — full plan (Phases 1-5) ([#204](https://github.com/pantheon-org/skill-quality-auditor/issues/204)) ([71fc98c](https://github.com/pantheon-org/skill-quality-auditor/commit/71fc98c2a78f001ec24cf9a885acd6911b62b706))
* **planning:** require T-shirt effort sizing on draft/active plans ([#182](https://github.com/pantheon-org/skill-quality-auditor/issues/182)) ([3126653](https://github.com/pantheon-org/skill-quality-auditor/commit/31266530905f5082cad84c574592302eb4984e77))
* **scripts:** reviewed-baseline mechanism for check-docs-drift.sh ([#187](https://github.com/pantheon-org/skill-quality-auditor/issues/187)) ([be35f5b](https://github.com/pantheon-org/skill-quality-auditor/commit/be35f5b1f7e85649e83336e88038b75e6625c471))
* **skills:** add design-debate helper skill ([#192](https://github.com/pantheon-org/skill-quality-auditor/issues/192)) ([26c1656](https://github.com/pantheon-org/skill-quality-auditor/commit/26c165685742a13db1cbea807ff0a20376898f45))
* **skills:** add guided-interview helper skill ([#180](https://github.com/pantheon-org/skill-quality-auditor/issues/180)) ([c5089fd](https://github.com/pantheon-org/skill-quality-auditor/commit/c5089fde756985f9ec0dd3f9c483b7bd28f4d2ef))
* **skills:** plan-review auto-triggers interview for decision findings ([#198](https://github.com/pantheon-org/skill-quality-auditor/issues/198)) ([345ab50](https://github.com/pantheon-org/skill-quality-auditor/commit/345ab50323f1cbfca2d48a7b6afee585d3d9a058))


### Bug Fixes

* **ci:** disable PR-Agent's ticket-compliance analysis, log false positive ([#189](https://github.com/pantheon-org/skill-quality-auditor/issues/189)) ([da9f6ed](https://github.com/pantheon-org/skill-quality-auditor/commit/da9f6ed8b038391f1b917b9dc8be89d65022f1be))
* **scripts:** degrade gracefully instead of hard-failing when jq is missing ([#193](https://github.com/pantheon-org/skill-quality-auditor/issues/193)) ([4f3db4d](https://github.com/pantheon-org/skill-quality-auditor/commit/4f3db4d8fbd6804faeee8893fa663af7b19d6d53))
* **skills:** design-debate persistence rule — verdict decides the artifact ([#196](https://github.com/pantheon-org/skill-quality-auditor/issues/196)) ([9af849d](https://github.com/pantheon-org/skill-quality-auditor/commit/9af849d7aac40d640b56bef0149a51d5f8e67e37))

## [0.22.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.21.1...v0.22.0) (2026-07-05)


### Features

* **ci:** add PR-Agent advisory review bot on Gemini free tier ([#176](https://github.com/pantheon-org/skill-quality-auditor/issues/176)) ([421a6c5](https://github.com/pantheon-org/skill-quality-auditor/commit/421a6c5c2d77ba2a27e6c6797980505e0c193d6e))


### Bug Fixes

* **ci:** scope Skill Quality Gate to skill content, not all Go changes ([#179](https://github.com/pantheon-org/skill-quality-auditor/issues/179)) ([804a002](https://github.com/pantheon-org/skill-quality-auditor/commit/804a002955402a47a913b81325f5418c30b80d88))
* **llmclient:** fix skill-quality.yml's LLM-judge (Mistral + Retry-After backoff) ([#178](https://github.com/pantheon-org/skill-quality-auditor/issues/178)) ([e9f40f0](https://github.com/pantheon-org/skill-quality-auditor/commit/e9f40f097e7d7fdba4048cb3e42977b95924eb1b))

## [0.21.1](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.21.0...v0.21.1) (2026-07-04)


### Bug Fixes

* **ci:** pin third-party actions by SHA, declare workflow permissions ([#171](https://github.com/pantheon-org/skill-quality-auditor/issues/171)) ([6c9fda4](https://github.com/pantheon-org/skill-quality-auditor/commit/6c9fda4921452d0960bf314a4c7ee3b98b0b84a9))

## [0.21.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.20.0...v0.21.0) (2026-07-04)


### Features

* **ci:** Plumber CI/CD security gate — fail on Critical, track the rest as issues ([#124](https://github.com/pantheon-org/skill-quality-auditor/issues/124)) ([3817570](https://github.com/pantheon-org/skill-quality-auditor/commit/3817570125cd19e466095ea3825cdb13277a1ccb))
* **ci:** single rollup issue for Plumber findings (fixes duplicate-issue bug) ([#154](https://github.com/pantheon-org/skill-quality-auditor/issues/154)) ([bfabdc3](https://github.com/pantheon-org/skill-quality-auditor/commit/bfabdc3c4092a1744e44261ce0c65fc80c9dce90))
* **governance:** post-merge ADR/plan status sync ([#119](https://github.com/pantheon-org/skill-quality-auditor/issues/119)) ([bf1ee3b](https://github.com/pantheon-org/skill-quality-auditor/commit/bf1ee3b9a2b775204e7c7cee90585f25fa1f42e7))

## [0.20.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.19.0...v0.20.0) (2026-07-04)


### Features

* **scorer:** externalise D1/D6/analysis scoring patterns to YAML config ([#114](https://github.com/pantheon-org/skill-quality-auditor/issues/114)) ([cb9c654](https://github.com/pantheon-org/skill-quality-auditor/commit/cb9c6543289dbce60fbb375be66e24358441eb4d))
* user-configurable scoring pattern overrides ([#118](https://github.com/pantheon-org/skill-quality-auditor/issues/118)) ([673bef9](https://github.com/pantheon-org/skill-quality-auditor/commit/673bef95adb3e3e9b7045d1518f170d98d4eea7d))

## [0.19.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.18.0...v0.19.0) (2026-07-03)


### Features

* **scripts:** add related-path heuristic to plan-drift check ([#112](https://github.com/pantheon-org/skill-quality-auditor/issues/112)) ([cfb26eb](https://github.com/pantheon-org/skill-quality-auditor/commit/cfb26eb12aa592d2a49e105033176c8a22846adc))
* **skills:** add plan-review and plan-create skills, restructure plugins into domains ([#113](https://github.com/pantheon-org/skill-quality-auditor/issues/113)) ([8c4fb1d](https://github.com/pantheon-org/skill-quality-auditor/commit/8c4fb1d5b04589305308ce12bd571a1001c381e5))
* **skills:** add pr-author helper skill (A grade) ([#111](https://github.com/pantheon-org/skill-quality-auditor/issues/111)) ([6cdea37](https://github.com/pantheon-org/skill-quality-auditor/commit/6cdea373df6d087ee24fca2609dff581bc16bb72))
* **skills:** migrate local skills to .context/plugins/ and improve socratic-method to A-grade ([#108](https://github.com/pantheon-org/skill-quality-auditor/issues/108)) ([3e3b4c8](https://github.com/pantheon-org/skill-quality-auditor/commit/3e3b4c8b12039cfcca0082cb7f442e8e333f3fa6))

## [0.18.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.17.0...v0.18.0) (2026-07-03)


### Features

* add session-end reflection mechanism (rule + skill + evals) ([#103](https://github.com/pantheon-org/skill-quality-auditor/issues/103)) ([c4e95d3](https://github.com/pantheon-org/skill-quality-auditor/commit/c4e95d32e5eeded4f9000a1a9b9d72b9304d713d))
* **context:** add ai-native-eval assessment and .tmp gitignore rule ([#106](https://github.com/pantheon-org/skill-quality-auditor/issues/106)) ([7499c9b](https://github.com/pantheon-org/skill-quality-auditor/commit/7499c9b884574c2e6162819dea0f0e3ed2c3827c))
* **context:** add arxiv self-evolution research survey findings ([#105](https://github.com/pantheon-org/skill-quality-auditor/issues/105)) ([c83b8f2](https://github.com/pantheon-org/skill-quality-auditor/commit/c83b8f20df211c87af4c12f9e4c2cd4ef5c16931))


### Bug Fixes

* validate and index .context/audits/ files as first-class context ([#107](https://github.com/pantheon-org/skill-quality-auditor/issues/107)) ([b315827](https://github.com/pantheon-org/skill-quality-auditor/commit/b3158274e73d8d7990176ba8491f67ddb36942b3))

## [0.17.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.16.2...v0.17.0) (2026-07-02)


### Features

* add docs-check skill and document local skills and rules on GH Pages ([#102](https://github.com/pantheon-org/skill-quality-auditor/issues/102)) ([cfe6be0](https://github.com/pantheon-org/skill-quality-auditor/commit/cfe6be0efedaf4d874e5397782ce354cc30a38ad))
* add rules-management infrastructure ([#98](https://github.com/pantheon-org/skill-quality-auditor/issues/98)) ([de4fcd4](https://github.com/pantheon-org/skill-quality-auditor/commit/de4fcd46481efe3afd966326f2a025ec6125d4a5))
* **rules-management:** add evals, audit improvements, and reference docs ([#101](https://github.com/pantheon-org/skill-quality-auditor/issues/101)) ([d2d979f](https://github.com/pantheon-org/skill-quality-auditor/commit/d2d979fa5e8012d86aee0e6ad480580be59bf977))


### Bug Fixes

* disable goreleaser release management for release immutability compatibility ([#100](https://github.com/pantheon-org/skill-quality-auditor/issues/100)) ([c79a9c4](https://github.com/pantheon-org/skill-quality-auditor/commit/c79a9c40fcf5cdc1200849fb76a031d7b58ebdb0))

## [0.16.2](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.16.1...v0.16.2) (2026-07-02)


### Bug Fixes

* **release:** comment out homebrew brews stanza blocking binary uploads ([#95](https://github.com/pantheon-org/skill-quality-auditor/issues/95)) ([6d28914](https://github.com/pantheon-org/skill-quality-auditor/commit/6d289149e83b449a0b92ddc6dee04400f5492733))

## [0.16.1](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.16.0...v0.16.1) (2026-07-01)


### Bug Fixes

* **release:** merge release-please and goreleaser into single workflow, add windows builds ([#93](https://github.com/pantheon-org/skill-quality-auditor/issues/93)) ([2c0c45d](https://github.com/pantheon-org/skill-quality-auditor/commit/2c0c45d2b9073deab50efe4a10d4718160bc6e92))

## [0.16.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.15.0...v0.16.0) (2026-07-01)


### Features

* **scorer:** implement D8 outcome-linkage sub-criterion ([#90](https://github.com/pantheon-org/skill-quality-auditor/issues/90)) ([0b6b8f2](https://github.com/pantheon-org/skill-quality-auditor/commit/0b6b8f28b9c036f2c9afc91c79d04b83273e5253))
* **skills:** finalise adr-capture, context-file, and context-index remediation phases ([#86](https://github.com/pantheon-org/skill-quality-auditor/issues/86)) ([da3059e](https://github.com/pantheon-org/skill-quality-auditor/commit/da3059e5307aaf833e01a5b4549dac34f4d26345))


### Bug Fixes

* **remediate:** include date in remediation plan filenames ([#85](https://github.com/pantheon-org/skill-quality-auditor/issues/85)) ([c756250](https://github.com/pantheon-org/skill-quality-auditor/commit/c756250bbf4b9e4953ec290e374195b40e4f6b0b))
* sync plan statuses with implementation, fix hook stale-index detection, add ways-of-working ([#83](https://github.com/pantheon-org/skill-quality-auditor/issues/83)) ([c1e70ba](https://github.com/pantheon-org/skill-quality-auditor/commit/c1e70ba12fd16588dc5bf979140d9372d224b63d))

## [0.15.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.14.0...v0.15.0) (2026-07-01)


### Features

* **context:** group index.yaml entries by type and fix remediation plan frontmatter ([#80](https://github.com/pantheon-org/skill-quality-auditor/issues/80)) ([d08aeb5](https://github.com/pantheon-org/skill-quality-auditor/commit/d08aeb50609a598b79468b218a86e631a50100a7))
* **docs:** verify all 28 academic citations, fix metadata, and add attribution refs ([#79](https://github.com/pantheon-org/skill-quality-auditor/issues/79)) ([92f2603](https://github.com/pantheon-org/skill-quality-auditor/commit/92f26036c1c7667517f2a306b4c6d267bc891352))

## [0.14.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.13.1...v0.14.0) (2026-06-30)


### Features

* D4 loose-scripts check, artifact trio bonus, 3 skill remediations, docs overhaul ([#78](https://github.com/pantheon-org/skill-quality-auditor/issues/78)) ([3a54a67](https://github.com/pantheon-org/skill-quality-auditor/commit/3a54a670e2b729fcc0db657f1d5526db193cadca))


### Bug Fixes

* **assets:** replace hardcoded maintainer path in skill-taxonomy reference ([#74](https://github.com/pantheon-org/skill-quality-auditor/issues/74)) ([a727fea](https://github.com/pantheon-org/skill-quality-auditor/commit/a727feaf47d4e7087611e71994cd26400931e66c))

## [0.13.1](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.13.0...v0.13.1) (2026-04-29)


### Bug Fixes

* **init:** tilde paths, deduplicate shared targets, list assets in dry-run ([#72](https://github.com/pantheon-org/skill-quality-auditor/issues/72)) ([521e55e](https://github.com/pantheon-org/skill-quality-auditor/commit/521e55ed900c764518884fb486bb91bcbf3126ba))

## [0.13.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.12.2...v0.13.0) (2026-04-29)


### Features

* **init:** harness detection, interactive mode, full asset copy, CWD default ([#70](https://github.com/pantheon-org/skill-quality-auditor/issues/70)) ([7dba25d](https://github.com/pantheon-org/skill-quality-auditor/commit/7dba25d2a89e5b3a513c2e9b3c90619a48b2244c))

## [0.12.2](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.12.1...v0.12.2) (2026-04-29)


### Bug Fixes

* **evals:** add scenario-06 to cover agent-neutral authoring and audit artefact hygiene ([#68](https://github.com/pantheon-org/skill-quality-auditor/issues/68)) ([e5ac62e](https://github.com/pantheon-org/skill-quality-auditor/commit/e5ac62e75eb94a828941b7df9079c7ff7b1cd858))

## [0.12.1](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.12.0...v0.12.1) (2026-04-29)


### Bug Fixes

* **skill:** self-audit improvements to reach A grade (127/140) ([#66](https://github.com/pantheon-org/skill-quality-auditor/issues/66)) ([05967e2](https://github.com/pantheon-org/skill-quality-auditor/commit/05967e24150b7c5e91f90e819f5e4367fc589ba8))

## [0.12.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.11.0...v0.12.0) (2026-04-29)


### Features

* **cmd:** extract resolveOutputFormat helper and update README flag reference ([#64](https://github.com/pantheon-org/skill-quality-auditor/issues/64)) ([c03cbeb](https://github.com/pantheon-org/skill-quality-auditor/commit/c03cbeb4a94d89533b7e461921c8ecce205bcee7))
* **init:** add --dry-run flag and add flag shorthands ([#63](https://github.com/pantheon-org/skill-quality-auditor/issues/63)) ([8aac9ee](https://github.com/pantheon-org/skill-quality-auditor/commit/8aac9ee5c430beede9920312e1e7d272b93e4750))
* **prune,validate:** add --repo-root and --dry-run flags and add shorthands ([#61](https://github.com/pantheon-org/skill-quality-auditor/issues/61)) ([371a6a5](https://github.com/pantheon-org/skill-quality-auditor/commit/371a6a506857ca94b33b00ea20c14203a3e88eac))

## [0.11.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.10.0...v0.11.0) (2026-04-29)


### Features

* **remediate:** add --json and --dry-run flags and add flag shorthands ([#58](https://github.com/pantheon-org/skill-quality-auditor/issues/58)) ([b10ad75](https://github.com/pantheon-org/skill-quality-auditor/commit/b10ad750a856e997c310308959e37c295b0d394a))

## [0.10.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.9.0...v0.10.0) (2026-04-29)


### Features

* **analyze:** add --json flag and add flag shorthands ([#54](https://github.com/pantheon-org/skill-quality-auditor/issues/54)) ([0ee0d5b](https://github.com/pantheon-org/skill-quality-auditor/commit/0ee0d5bbb1de336460aac4324da8f74f019036f8))
* **batch:** add --markdown flag and add flag shorthands ([#56](https://github.com/pantheon-org/skill-quality-auditor/issues/56)) ([0322889](https://github.com/pantheon-org/skill-quality-auditor/commit/0322889012c55177cbed7716b0a9fb2a7f57d690))
* **duplication,trend:** add --markdown and --store flags and add shorthands ([#57](https://github.com/pantheon-org/skill-quality-auditor/issues/57)) ([52b5ad0](https://github.com/pantheon-org/skill-quality-auditor/commit/52b5ad0d4f3d4ff47b8df2dd51569300d95570e5))

## [0.9.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.8.2...v0.9.0) (2026-04-29)


### Features

* **evaluate:** remove --json flag and add flag shorthands ([#52](https://github.com/pantheon-org/skill-quality-auditor/issues/52)) ([74def52](https://github.com/pantheon-org/skill-quality-auditor/commit/74def52c75e7e58024d3c5c72ce7efcd33395418))

## [0.8.2](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.8.1...v0.8.2) (2026-04-29)


### Bug Fixes

* **lint:** move mdlint rule suppression from CLI flags to mdlint.toml ([#49](https://github.com/pantheon-org/skill-quality-auditor/issues/49)) ([a6c8576](https://github.com/pantheon-org/skill-quality-auditor/commit/a6c85762405b3d780bae7dc7cd3a5c7c1787bbaa))

## [0.8.1](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.8.0...v0.8.1) (2026-04-29)


### Bug Fixes

* **lint:** add mdlint.toml, fix duplicate references heading, restore skill grade to B ([#46](https://github.com/pantheon-org/skill-quality-auditor/issues/46)) ([5384a31](https://github.com/pantheon-org/skill-quality-auditor/commit/5384a3133580cc2cea9a891398b67bcd3518e5ee))

## [0.8.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.7.0...v0.8.0) (2026-04-29)


### Features

* **scorer:** D2 preconditions/postconditions/decision-point sub-scorers ([#44](https://github.com/pantheon-org/skill-quality-auditor/issues/44)) ([47d75eb](https://github.com/pantheon-org/skill-quality-auditor/commit/47d75ebf10e669c729226b28526dbb9568b5ba8d))
* **scorer:** implement D8 outcome-linkage sub-criterion ([#39](https://github.com/pantheon-org/skill-quality-auditor/issues/39)) ([4967c17](https://github.com/pantheon-org/skill-quality-auditor/commit/4967c17ffbebf61f104e35ef19e315e6ff7bb001))

## [0.7.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.6.0...v0.7.0) (2026-04-29)


### Features

* **scorer:** implement D3 SYMPTOM/CONSEQUENCE per-block detection ([#36](https://github.com/pantheon-org/skill-quality-auditor/issues/36)) ([73b554a](https://github.com/pantheon-org/skill-quality-auditor/commit/73b554afd6697ab641ad733325670554249315fc))

## [0.6.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.5.0...v0.6.0) (2026-04-29)


### Features

* **scorer:** implement D4 mutation-resistance scoring ([#40](https://github.com/pantheon-org/skill-quality-auditor/issues/40)) ([8ebfa6f](https://github.com/pantheon-org/skill-quality-auditor/commit/8ebfa6f851aeb95a154b576f1ef8f8c4da0d43a1))

## [0.5.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.4.0...v0.5.0) (2026-04-29)


### Features

* **scorer:** implement D7 discriminativeness diagnostic signal ([#32](https://github.com/pantheon-org/skill-quality-auditor/issues/32)) ([e061f98](https://github.com/pantheon-org/skill-quality-auditor/commit/e061f983ce0c36ddd1c912a3ab5fce8e64502ac3))

## [0.4.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.3.0...v0.4.0) (2026-04-29)


### Features

* **scorer:** implement D1 demonstration concreteness sub-criterion ([#35](https://github.com/pantheon-org/skill-quality-auditor/issues/35)) ([64f0e04](https://github.com/pantheon-org/skill-quality-auditor/commit/64f0e044bae4eeda78ab79d2afe5cd3e486bc0a5))
* **scorer:** implement D6 constraint typology sub-scorers ([#34](https://github.com/pantheon-org/skill-quality-auditor/issues/34)) ([6929321](https://github.com/pantheon-org/skill-quality-auditor/commit/692932188e2b31c4d9e7e3a6b60f6a3e38ef400d))

## [0.3.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.2.0...v0.3.0) (2026-04-29)


### Features

* **scorer:** implement D5 negative-condition detection with table-row scoping ([#30](https://github.com/pantheon-org/skill-quality-auditor/issues/30)) ([df75ed6](https://github.com/pantheon-org/skill-quality-auditor/commit/df75ed6ef6fde17a47f5451749be1b1289d9c67e))
* **scorer:** implement D9 mutation-coverage scoring and CI-safe independent-authoring ([#33](https://github.com/pantheon-org/skill-quality-auditor/issues/33)) ([eb9e9ad](https://github.com/pantheon-org/skill-quality-auditor/commit/eb9e9ad7a81e2743881887dbf774b4f258dcfced))

## [0.2.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.1.5...v0.2.0) (2026-04-29)


### Features

* add install.sh, Homebrew tap, mise support, and update command ([#14](https://github.com/pantheon-org/skill-quality-auditor/issues/14)) ([7a6c92b](https://github.com/pantheon-org/skill-quality-auditor/commit/7a6c92b8dba14df8d37a30cc68b3a1fedf29773c))
* move Go module to repo root and adopt GoReleaser ([#11](https://github.com/pantheon-org/skill-quality-auditor/issues/11)) ([6663943](https://github.com/pantheon-org/skill-quality-auditor/commit/6663943a063042b6c536f0ccb67d182045c6ccd3))
* swap markdownlint-cli2 for mdlint ([#17](https://github.com/pantheon-org/skill-quality-auditor/issues/17)) ([41c1083](https://github.com/pantheon-org/skill-quality-auditor/commit/41c10838f7f2c7eb8939a5bf6b90719f89a29319))


### Bug Fixes

* address all code review findings and coverage gaps ([#22](https://github.com/pantheon-org/skill-quality-auditor/issues/22)) ([c1c1359](https://github.com/pantheon-org/skill-quality-auditor/commit/c1c1359aed480daa2fb3fb5d15a72b2a6a4ce290))
* **ci:** update skill-quality workflow paths for root module layout ([#13](https://github.com/pantheon-org/skill-quality-auditor/issues/13)) ([6097a8e](https://github.com/pantheon-org/skill-quality-auditor/commit/6097a8ea6076cbe1af09a8515800efb462cf5e65))
* correct tile.json path in release-please extra-files config ([#26](https://github.com/pantheon-org/skill-quality-auditor/issues/26)) ([3e1156f](https://github.com/pantheon-org/skill-quality-auditor/commit/3e1156fbc300988f2159dd0b0743be8d2e065af8))
* **mdlint:** remove inline disable, raise line_length to 130 ([#18](https://github.com/pantheon-org/skill-quality-auditor/issues/18)) ([ab025a1](https://github.com/pantheon-org/skill-quality-auditor/commit/ab025a14f06610548d678d3fc7c5cf7425038c3d))
* **mise:** remove invalid brew:tesslio/tap/tessl backend ([#15](https://github.com/pantheon-org/skill-quality-auditor/issues/15)) ([86077d3](https://github.com/pantheon-org/skill-quality-auditor/commit/86077d34390a221fd98a5946b852074836a31019))
