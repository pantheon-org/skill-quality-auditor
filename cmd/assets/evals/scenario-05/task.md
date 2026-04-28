# Scenario 05: Skills Portfolio Duplication Analysis

## User Prompt

"Analyse my 8 CI/CD and infrastructure skills for duplication and give consolidation recommendations."

## Expected Behavior

1. Perform pairwise similarity analysis across all relevant skill pairs in the collection (not just a subset).
2. Assign a numeric similarity percentage to each pair — not qualitative labels like "high" or "low".
3. Apply the thresholds: <20% monitor only, 20–35% plan aggregation, >35% immediate action required.
4. Flag generator/validator pairs within each tool domain as high-overlap candidates for consolidation.
5. Evaluate whether terraform and terragrunt skills overlap significantly.
6. Apply the Navigation Hub pattern for any pairs recommended for consolidation above the immediate-action threshold.
7. Reject cross-domain aggregation (e.g., do not recommend merging `terraform-generator` with `github-actions-generator`).
8. Produce `similarity-analysis.md`, `consolidation-recommendations.md`, and `duplication-report.json`.

## Success Criteria

- `similarity-analysis.md` analyses all relevant pairs, not just a subset.
- Each pair assigned a numeric similarity percentage.
- Pairs above 20% similarity flagged as aggregation candidates.
- Pairs above 35% similarity flagged as requiring immediate action.
- No cross-domain aggregation recommended (e.g., terraform + github-actions kept separate).
- Navigation Hub pattern mentioned in `consolidation-recommendations.md`.
- `duplication-report.json` contains a `pairs` array with `skill_a`, `skill_b`, `similarity_pct`, and `action` fields.
- For pairs recommended to stay separate, an explicit justification is provided (different purpose, domain fit, etc.).

## Failure Conditions

- Only a subset of pairs analysed.
- Similarity reported as qualitative labels without numeric percentages.
- Cross-domain aggregation recommended (e.g., merging skills from `ci-cd/` and `infrastructure/`).
- Navigation Hub pattern not referenced for consolidation candidates.
- `duplication-report.json` missing required fields (`skill_a`, `skill_b`, `similarity_pct`, `action`).
- Keep-separate recommendations provided without justification.

**Skills collection:**

- `ci-cd/github-actions-generator`
- `ci-cd/github-actions-validator`
- `ci-cd/gitlab-ci-generator`
- `ci-cd/gitlab-ci-validator`
- `infrastructure/terraform-generator`
- `infrastructure/terraform-validator`
- `infrastructure/terragrunt-generator`
- `infrastructure/terragrunt-validator`

**Similarity thresholds:** <20% monitor only | 20–35% plan aggregation | >35% immediate action
