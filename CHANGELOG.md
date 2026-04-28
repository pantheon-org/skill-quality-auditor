# Changelog

## [0.2.0](https://github.com/pantheon-org/skill-quality-auditor/compare/v0.1.5...v0.2.0) (2026-04-28)


### Features

* add analyze command and combined analysis reporter ([f7f60bb](https://github.com/pantheon-org/skill-quality-auditor/commit/f7f60bb10c23d6851fc3139fcfcdea31c9da3923))
* add coverage.html to .gitignore ([2924285](https://github.com/pantheon-org/skill-quality-auditor/commit/2924285d130cc7bf2fa5f5985ef318dea07adb15))
* add GitHub CI/release workflows and skill-auditor init command ([aeb8b1e](https://github.com/pantheon-org/skill-quality-auditor/commit/aeb8b1ec4130282e6418780a81d289a323d36563))
* add rule-based pattern detectors to analysis package ([a80182d](https://github.com/pantheon-org/skill-quality-auditor/commit/a80182dc0ce0e08aa12f9fb68be53f0336207ebd))
* add TF-IDF keyword extractor to analysis package ([4d31560](https://github.com/pantheon-org/skill-quality-auditor/commit/4d31560eea445396532be4b494c3d1086b9fd95b))
* add validate, lint, and prune commands porting 4 shell scripts ([edf7846](https://github.com/pantheon-org/skill-quality-auditor/commit/edf7846408242fbf31267a9e9ccdd045d8d19da7))
* automate releases via release-please and tile.json version source ([#8](https://github.com/pantheon-org/skill-quality-auditor/issues/8)) ([34cd6ee](https://github.com/pantheon-org/skill-quality-auditor/commit/34cd6eec14635869de058eaac7b272a818ed78e1))
* consolidate skill into cmd/assets, add CI quality gate ([2355a22](https://github.com/pantheon-org/skill-quality-auditor/commit/2355a2266701e04fd766f27d1a44acda86c80989))
* phase 6 — skill consolidation, CI quality gate, dist/ build path ([76f980b](https://github.com/pantheon-org/skill-quality-auditor/commit/76f980b46e2525afb1027896a49afb052d04cc3c))
* update build commands and add new analysis commands to README and CONTRIBUTING ([dfdef0e](https://github.com/pantheon-org/skill-quality-auditor/commit/dfdef0e7e8e721cd29420f5e51d2095ba75377fe))


### Bug Fixes

* address 9 code review findings across skill-auditor packages ([3a137ea](https://github.com/pantheon-org/skill-quality-auditor/commit/3a137ead8b98e5775205f592a3106752b3688e1b))
* **assets:** replace shell script refs with Go CLI commands; migrate evals to flat scenario-NN.md format ([67db3af](https://github.com/pantheon-org/skill-quality-auditor/commit/67db3af14db72f0eeb1ab26a30420e76488cdcc8))
* consolidate tessl into existing mise.toml, remove duplicate .mise.toml ([924cfc2](https://github.com/pantheon-org/skill-quality-auditor/commit/924cfc21daeb65e70e610cc81b08bc02ea9281ea))
* gofmt evaluate.go ([440ba8b](https://github.com/pantheon-org/skill-quality-auditor/commit/440ba8b8c6a53bf9e32f3859264009c11a8a5f21))
* move tile.json into skill-auditor/cmd/assets/ so tessl finds evals/ ([#5](https://github.com/pantheon-org/skill-quality-auditor/issues/5)) ([5fa9c7e](https://github.com/pantheon-org/skill-quality-auditor/commit/5fa9c7e60ae0652877ccba5afe9a8a2a3f69f937))
* point CI duplication and batch steps at correct assets path ([a79e1ed](https://github.com/pantheon-org/skill-quality-auditor/commit/a79e1ed22792a3a782e970c9787790fa58aa9f8a))
* point skill-duplication and skill-batch hooks at correct assets path ([a243987](https://github.com/pantheon-org/skill-quality-auditor/commit/a243987782a7d1b02ed2229d57cf78b418668083))
* point tessl eval run at repo root where tile.json lives ([#4](https://github.com/pantheon-org/skill-quality-auditor/issues/4)) ([972397b](https://github.com/pantheon-org/skill-quality-auditor/commit/972397bca68ef30bc83741c575cda45e43a3ecba))
* resolve all CI pipeline failures ([#1](https://github.com/pantheon-org/skill-quality-auditor/issues/1)) ([a251199](https://github.com/pantheon-org/skill-quality-auditor/commit/a25119933e30bb2b5bf320efa0683db1e7048443))
* resolve all markdownlint CI failures ([#3](https://github.com/pantheon-org/skill-quality-auditor/issues/3)) ([e9b4d1c](https://github.com/pantheon-org/skill-quality-auditor/commit/e9b4d1c3e75186024418679d3a67bd50c394db73))
* resolve all markdownlint errors in root docs ([3156755](https://github.com/pantheon-org/skill-quality-auditor/commit/3156755d0fc70f681354e9150f8fbbd589e10fc9))
* tessl eval run must point at skill-auditor/cmd/assets/ where tile.json lives ([#6](https://github.com/pantheon-org/skill-quality-auditor/issues/6)) ([170ea7e](https://github.com/pantheon-org/skill-quality-auditor/commit/170ea7ef918d9a31bcb2042a824211c61ca398d1))
* upgrade golangci-lint-action to v7 (required for golangci-lint v2) ([#2](https://github.com/pantheon-org/skill-quality-auditor/issues/2)) ([c86e1c2](https://github.com/pantheon-org/skill-quality-auditor/commit/c86e1c2c11c3c6cac72658d0ccd5a930aaac08e0))
