# Scenario 02: Skills Collection Quality Audit

## User Prompt

"Run a full batch audit of all 12 skills and compare results against last month's baseline."

## Expected Behavior

1. Run `skill-auditor batch` across all 12 skills with the `--store` flag to persist results.
2. Load the existing baseline from `.context/audits/*/2025-10-01/audit.json` for the 8 previously-audited skills.
3. Mark the 4 new skills (no baseline entry) as "new — no delta available" in the comparison output.
4. Produce `audit-execution.sh` containing exact, reproducible commands used to run the audit.
5. Produce `audit-results.json` with the full JSON batch output.
6. Produce `audit-report.md` as a human-readable table: skill | score | grade | delta vs baseline.
7. Produce `baseline-comparison.md` identifying improvements, regressions, and new skills categories.

## Success Criteria

- `skill-auditor batch` used with `--store` flag; all 12 skills evaluated in a single run.
- Grade distribution table includes a delta vs baseline column.
- The 4 new skills are listed and marked as "new — no delta available".
- `baseline-comparison.md` separates results into improvements, regressions, and new skills.
- `audit-execution.sh` contains exact commands that can be re-run to reproduce the audit.

## Failure Conditions

- Only the 8 previously-baselined skills evaluated; new skills skipped.
- Delta comparison column missing from the audit report table.
- Skills evaluated one-by-one with individual `evaluate` commands instead of `batch`.
- No reproducible command record produced.
- New skills not distinguished from regressions in the comparison report.

**Context:**

- Repository root: current working directory
- Previous baseline: `.context/audits/*/2025-10-01/audit.json` files exist for 8 of the 12 skills
- Four new skills have no baseline yet
- Grading thresholds: A >=126, B+ >=119, B >=112, C/C+ <112 (blocked from publishing)
