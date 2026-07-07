# Architecture overview

## High-level package dependency graph

```text
main.go
  └── cmd/ (cobra commands)
        ├── scorer/        (scoring engine)
        │     ├── scorer.go           (Score, ScoreFromContent)
        │     ├── dimensions.go       (Dimension, Diagnostic, Result, AllDimensions)
        │     ├── grades.go           (GradeRank, Grade)
        │     ├── thresholds.go       (rubric cut-points)
        │     ├── validator_bridge.go (skill-validator integration)
        │     ├── d1_knowledge_delta.go
        │     ├── d2_mindset_procedures.go
        │     ├── d3_anti_pattern_coverage.go
        │     ├── d4_specification.go
        │     ├── d5_progressive_disclosure.go
        │     ├── d6_freedom_calibration.go
        │     ├── d7_pattern_recognition.go
        │     ├── d8_practical_usability.go
        │     └── d9_eval_validation.go
        │
        ├── reporter/      (formatting, storage, plans)
        │     ├── reporter.go          (Format — human-readable text)
        │     ├── store.go             (Store — persist to .context/audits/)
        │     ├── analysis.go          (Analysis — markdown audit report)
        │     ├── duplication.go       (DuplicationReport)
        │     ├── aggregation.go       (AggregationPlan)
        │     ├── remediation.go       (Remediation — simple plan)
        │     ├── remediation_plan_generate.go (schema-compliant plan: emits effort/value/themes frontmatter)
        │     └── remediation_plan_validate.go (schema validation)
        │
        ├── duplication/   (pairwise similarity)
        │     ├── inventory.go (SkillEntry, Inventory)
        │     ├── detect.go    (Pair, Detect, thresholds)
        │     └── similarity.go (Jaccard, TokenSet, SectionHeaders)
        │
        ├── agents/        (agent registry)
        │     └── registry.go
        │
        ├── analysis/      (static analysis)
        │     ├── patterns.go (rule-based detection)
        │     └── tfidf.go    (keyword extraction)
        │
        └── internal/
              ├── llmclient/ (provider-agnostic LLM client)
              │     ├── types.go      (Client, Provider, Message, etc.)
              │     ├── client.go     (NewFromEnv, providers registry)
              │     ├── anthropic.go  (Anthropic provider)
              │     ├── openai.go     (OpenAI provider)
              │     ├── gemini.go     (Gemini provider)
              │     ├── mistral.go    (Mistral provider, OpenAI-wire-compatible)
              │     ├── cerebras.go   (Cerebras provider, OpenAI-wire-compatible)
              │     └── prompt.go     (JudgePrompt, ActorMessages, JudgeMessages)
              │
              ├── patternconfig/ (externalised D1/D6/analysis-quality pattern words)
              │     └── loads & validates scoring-patterns.yaml against
              │         scoring-patterns.schema.json (ADR-028); LoadFromPath +
              │         WriteDefault back the 5-tier override chain (ADR-032)
              │
              └── tokenize/   (text normalization)
                    └── tokenize.go (Normalize, Set, Counts)
```

`internal/patternconfig` is resolved once per invocation via `cmd/root.go`'s
`PersistentPreRunE` and consumed by `scoreD1`, `scoreD6`, and `analysis/patterns.go` — so
the beginner/expert signal words, "when not to use" phrases, and hedge/vague/passive word
lists are maintainer-editable YAML, not Go constants. Resolution follows a 5-tier
precedence (`-c/--config` flag → CWD file → per-OS default path → embedded config →
hardcoded defaults; `eval` and `--no-user-config` skip straight to the embedded tier for
reproducibility) — see ADR-028, ADR-032, and
[Configuring scoring patterns](../development/setup.md#configuring-scoring-patterns).

## Data flow

```text
User CLI input
    │
    ▼
cobra.Command
    │
    ├── evaluate ──► scorer.Score()
    │                    │
    │                    ├── validatorBridge (external skill-validator library)
    │                    ├── scoreD1 ··· scoreD9
    │                    │
    │                    ▼
    │                scorer.Result     ──► reporter.Format()  → stdout
    │                    │                 reporter.Store()   → .context/audits/
    │                    ▼
    │                reporter.Remediation() → .context/audits/*/Remediation.md
    │
    ├── batch     ──► loop scorer.Score()   ──► sorted table / JSON
    │
    ├── duplication─► duplication.Inventory() → Detect() → reporter.DuplicationReport()
    │
    ├── aggregate ──► Inventory() → filter family → Detect() → reporter.AggregationPlan()
    │
    ├── remediate ──► load audit.json → reporter.RemediationPlan() / ValidateRemediationPlan()
    │
    ├── trend     ──► group audits by skill → compute deltas → table / JSON
    │
    ├── eval      ──► load scenarios → llmclient (actor + judge) → PASS/FAIL
    │
    ├── analyze   ──► read SKILL.md → ExtractKeywords() + Detect*() → CombinedAnalysis
    │
    ├── validate  ──► artifacts: walk skills dir → check schemas/templates/scripts/SKILL.md conventions
    │                 context:   JSON-schema-validate context frontmatter under a given path (default .context; santhosh-tekuri)
    │
    ├── init      ──► resolve agents → write embedded assets → symlink/copy to harness dirs
    │
    ├── update    ──► GitHub API → download tarball → verify checksum → replace binary
    │
    ├── prune     ──► read audit dirs → keep N newest per skill → remove rest
    │
    └── version   ──► print version + release date (buildDate ldflag / vcs.time)
```

## Output layout

```text
.context/
  audits/
    <domain/skill-name>/
      <YYYY-MM-DD>/
        audit.json        (scorer.Result — JSON)
        Analysis.md       (human-readable markdown)
        Remediation.md    (simple remediation markdown)
  analysis/
    duplication-report-YYYY-MM-DD.md
    aggregation-plan-<family>-YYYY-MM-DD.md
    pattern-report-<skill>-YYYY-MM-DD.md
  plans/
    <skill-name>-remediation-plan-<date>.md
```

Every `.context/**/*.md` file carries YAML frontmatter whose enum values are
UPPER_CASE — `type` (`PLAN`/`FINDING`/`ANALYSIS`/`INSTRUCTION`/`AUDIT`/`KNOWN_ISSUE`),
`status` (`DRAFT`/`ACTIVE`/`DONE`/`SUPERSEDED`), `severity`, and `value`. The
schema-compliant remediation plan generator emits the same convention, so a freshly
generated plan validates without hand-patching. See ADR-050 and
`.context/instructions/value-rubric.md`.
