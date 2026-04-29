package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- ValidateRemediationPlan ----

func TestValidateRemediationPlan_validPlan(t *testing.T) {
	r := makeResultWithScore(84)
	plan, err := RemediationPlan(r, 0, ".context/audits/my-skill/2026-04-27/Analysis.md", "2026-04-27")
	if err != nil {
		t.Fatalf("generate plan: %v", err)
	}
	f := writeTempPlan(t, plan)
	errs := ValidateRemediationPlan(f)
	if len(errs) > 0 {
		t.Errorf("expected valid plan, got errors:\n%s", strings.Join(errs, "\n"))
	}
}

func TestValidateRemediationPlan_missingFile(t *testing.T) {
	errs := ValidateRemediationPlan("/nonexistent/plan.md")
	if len(errs) == 0 {
		t.Error("expected error for missing file")
	}
}

func TestValidateRemediationPlan_noFrontmatter(t *testing.T) {
	f := writeTempPlan(t, "# Just a markdown file\n\nNo frontmatter here.")
	errs := ValidateRemediationPlan(f)
	if len(errs) == 0 {
		t.Error("expected error for missing frontmatter")
	}
}

func TestValidateRemediationPlan_badDate(t *testing.T) {
	fm := minimalValidFrontmatter()
	fm = strings.ReplaceAll(fm, "plan_date: 2026-04-27", "plan_date: not-a-date")
	f := writeTempPlan(t, "---\n"+fm+"\n---\n")
	errs := ValidateRemediationPlan(f)
	if !containsError(errs, "plan_date") {
		t.Errorf("expected plan_date validation error, got: %v", errs)
	}
}

func TestValidateRemediationPlan_badSkillName(t *testing.T) {
	fm := minimalValidFrontmatter()
	fm = strings.ReplaceAll(fm, "skill_name: my-skill", "skill_name: My/Skill With Spaces")
	f := writeTempPlan(t, "---\n"+fm+"\n---\n")
	errs := ValidateRemediationPlan(f)
	if !containsError(errs, "skill_name") {
		t.Errorf("expected skill_name validation error, got: %v", errs)
	}
}

func TestValidateRemediationPlan_badSourceAudit(t *testing.T) {
	fm := minimalValidFrontmatter()
	fm = strings.ReplaceAll(fm, ".context/audits/my-skill/2026-04-27/Analysis.md", "wrong/path.txt")
	f := writeTempPlan(t, "---\n"+fm+"\n---\n")
	errs := ValidateRemediationPlan(f)
	if !containsError(errs, "source_audit") {
		t.Errorf("expected source_audit validation error, got: %v", errs)
	}
}

func TestValidateRemediationPlan_emptyCriticalIssues(t *testing.T) {
	content := "---\n" + noCriticalIssuesFrontmatter() + "\n---\n"
	f := writeTempPlan(t, content)
	errs := ValidateRemediationPlan(f)
	if !containsError(errs, "critical_issues") {
		t.Errorf("expected critical_issues validation error, got: %v", errs)
	}
}

// ---- extractFrontmatter ----

func TestExtractFrontmatter_noFrontmatter(t *testing.T) {
	_, err := extractFrontmatter([]byte("# No frontmatter\n\nJust content."))
	if err == nil {
		t.Error("expected error for content without frontmatter delimiter")
	}
}

func TestExtractFrontmatter_valid(t *testing.T) {
	content := "---\nplan_date: 2026-04-27\nskill_name: my-skill\n---\n# Body"
	fm, err := extractFrontmatter([]byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.PlanDate != "2026-04-27" {
		t.Errorf("expected plan_date '2026-04-27', got %q", fm.PlanDate)
	}
}

// ---- helpers ----

func writeTempPlan(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp plan: %v", err)
	}
	return f
}

func containsError(errs []string, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}

func noCriticalIssuesFrontmatter() string {
	return `plan_date: 2026-04-27
skill_name: my-skill
source_audit: .context/audits/my-skill/2026-04-27/Analysis.md
executive_summary:
  score:
    current: "84/140 (60%)"
    target: "104/140 (74%)"
  grade:
    current: F
    target: C
  priority: High
  effort: M
  focus_areas:
    - "D1: Knowledge Delta"
  verdict: Priority improvements required
critical_issues: []
remediation_phases:
  - phase: 1
    dimension: "D1: Knowledge Delta (12/20)"
    priority: High
    target: Increase from 12/20 to 20/20 (+8 points)
    steps:
      - step: "1.1"
        title: Add expert-signal keywords
        description: Add expert-signal keywords that go beyond LLM baseline knowledge.
verification_commands:
  - ./skill-auditor evaluate my-skill
success_criteria:
  - criterion: Total score target
    measurement: "Score >= 104/140"
effort_estimates:
  - phase: Phase 1
    effort: M
    time: 1 hour
  - phase: Total
    effort: M
    time: 2 hours
dependencies:
  - Completed audit stored in .context/audits/
rollback_plan: git checkout HEAD -- skills/my-skill/SKILL.md
notes:
  rating: "6/10"
  assessment: Review carefully before publishing.
`
}

func minimalValidFrontmatter() string {
	return `plan_date: 2026-04-27
skill_name: my-skill
source_audit: .context/audits/my-skill/2026-04-27/Analysis.md
executive_summary:
  score:
    current: "84/140 (60%)"
    target: "104/140 (74%)"
  grade:
    current: F
    target: C
  priority: High
  effort: M
  focus_areas:
    - "D1: Knowledge Delta"
  verdict: Priority improvements required
critical_issues:
  - issue: Score is below target for this dimension
    dimension: "D1: Knowledge Delta (12/20)"
    severity: High
    impact: Missing points prevent grade improvement
remediation_phases:
  - phase: 1
    dimension: "D1: Knowledge Delta (12/20)"
    priority: High
    target: Increase from 12/20 to 20/20 (+8 points)
    steps:
      - step: "1.1"
        title: Add expert-signal keywords
        description: Add expert-signal keywords that go beyond LLM baseline knowledge.
verification_commands:
  - ./skill-auditor evaluate my-skill
success_criteria:
  - criterion: Total score target
    measurement: "Score >= 104/140"
effort_estimates:
  - phase: Phase 1
    effort: M
    time: 1 hour
  - phase: Total
    effort: M
    time: 2 hours
dependencies:
  - Completed audit stored in .context/audits/
rollback_plan: git checkout HEAD -- skills/my-skill/SKILL.md
notes:
  rating: "6/10"
  assessment: Review carefully before publishing.
`
}
