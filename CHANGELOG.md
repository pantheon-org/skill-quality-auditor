# Changelog

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
