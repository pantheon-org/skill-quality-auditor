package reporter

import (
	"strings"
	"testing"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

func makeResultWithScore(total int) *scorer.Result {
	grade := scorer.Grade(total)
	dims := map[string]int{}
	for _, d := range scorer.AllDimensions {
		dims[d.Key] = d.Max * total / 140
	}
	return &scorer.Result{
		Skill:      "tools/my-skill",
		Date:       "2026-04-27",
		Total:      total,
		MaxTotal:   140,
		Grade:      grade,
		Dimensions: dims,
	}
}

// ---- RemediationPlan generation ----

func TestRemediationPlan_generatesValidYAML(t *testing.T) {
	r := makeResultWithScore(84)
	plan, err := RemediationPlan(r, 0, "", "2026-04-27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(plan, "---\n") {
		t.Error("plan should start with YAML frontmatter delimiter")
	}
	if !strings.Contains(plan, "plan_date:") {
		t.Error("plan should contain plan_date field")
	}
	if !strings.Contains(plan, "skill_name:") {
		t.Error("plan should contain skill_name")
	}
	if !strings.Contains(plan, "critical_issues:") {
		t.Error("plan should contain critical_issues")
	}
	if !strings.Contains(plan, "remediation_phases:") {
		t.Error("plan should contain remediation_phases")
	}
}

func TestRemediationPlan_targetScoreBelowOrEqualTotal(t *testing.T) {
	r := makeResultWithScore(84)
	_, err := RemediationPlan(r, 50, "", "2026-04-27")
	if err == nil {
		t.Error("expected error when targetScore <= r.Total, got nil")
	}
}

func TestRemediationPlan_defaultTargetScore(t *testing.T) {
	r := makeResultWithScore(100)
	plan, err := RemediationPlan(r, 0, "", "2026-04-27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(plan, "120/140") {
		t.Errorf("expected target 120/140 in plan, got:\n%s", plan[:500])
	}
}

func TestRemediationPlan_targetScoreCappedAt140(t *testing.T) {
	r := makeResultWithScore(130)
	plan, err := RemediationPlan(r, 200, "", "2026-04-27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(plan, "140/140") {
		t.Errorf("target score should be capped at 140, got:\n%s", plan[:500])
	}
}

func TestRemediationPlan_perfectScore(t *testing.T) {
	r := makeResultWithScore(140)
	for _, d := range scorer.AllDimensions {
		r.Dimensions[d.Key] = d.Max
	}
	r.Total = 140
	r.Grade = "A+"
	plan, err := RemediationPlan(r, 0, "", "2026-04-27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == "" {
		t.Error("plan should not be empty even at perfect score")
	}
}

func TestRemediationPlan_containsMarkdownSections(t *testing.T) {
	r := makeResultWithScore(70)
	plan, err := RemediationPlan(r, 0, "", "2026-04-27")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, section := range []string{
		"## Executive Summary",
		"## Critical Issues",
		"## Remediation Phases",
		"## Verification Commands",
		"## Success Criteria",
	} {
		if !strings.Contains(plan, section) {
			t.Errorf("plan missing section %q", section)
		}
	}
}

// ---- gapPriority / gapEffort / gapTime ----

func TestGapPriority(t *testing.T) {
	cases := []struct {
		gap  int
		want string
	}{
		{12, "Critical"},
		{7, "High"},
		{4, "Medium"},
		{1, "Low"},
	}
	for _, c := range cases {
		if got := gapPriority(c.gap); got != c.want {
			t.Errorf("gapPriority(%d) = %q, want %q", c.gap, got, c.want)
		}
	}
}

func TestGapEffort(t *testing.T) {
	cases := []struct {
		gap  int
		want string
	}{
		{25, "L"},
		{10, "M"},
		{3, "S"},
	}
	for _, c := range cases {
		if got := gapEffort(c.gap); got != c.want {
			t.Errorf("gapEffort(%d) = %q, want %q", c.gap, got, c.want)
		}
	}
}

func TestGapTime(t *testing.T) {
	if !strings.Contains(gapTime(25), "3+") {
		t.Error("large gap should indicate 3+ hours")
	}
	if !strings.Contains(gapTime(10), "1-2") {
		t.Error("medium gap should indicate 1-2 hours")
	}
	if !strings.Contains(gapTime(2), "30") {
		t.Error("small gap should indicate 30 min")
	}
}

// ---- planVerdict ----

func TestPlanVerdict(t *testing.T) {
	if planVerdict("Critical") == "" {
		t.Error("verdict should not be empty for Critical priority")
	}
	if planVerdict("Low") == "" {
		t.Error("verdict should not be empty for Low priority")
	}
}

func TestPlanVerdict_allBranches(t *testing.T) {
	cases := []struct {
		priority string
		substr   string
	}{
		{"Critical", "Immediate"},
		{"High", "Priority"},
		{"Medium", "Targeted"},
		{"Low", "Minor"},
		{"Unknown", "Minor"},
		{"", "Minor"},
	}
	for _, c := range cases {
		got := planVerdict(c.priority)
		if !strings.Contains(got, c.substr) {
			t.Errorf("planVerdict(%q) = %q, want substring %q", c.priority, got, c.substr)
		}
	}
}

// ---- planNotes ----

func TestPlanNotes_withErrors(t *testing.T) {
	r := &scorer.Result{Errors: 3, Warnings: 1}
	got := planNotes(r)
	if !strings.Contains(got.Assessment, "error") {
		t.Errorf("planNotes with errors should mention 'error', got %q", got.Assessment)
	}
}

func TestPlanNotes_withWarningsOnly(t *testing.T) {
	r := &scorer.Result{Errors: 0, Warnings: 2}
	got := planNotes(r)
	if !strings.Contains(got.Assessment, "warning") {
		t.Errorf("planNotes with warnings should mention 'warning', got %q", got.Assessment)
	}
}

func TestPlanNotes_noIssues(t *testing.T) {
	r := &scorer.Result{Errors: 0, Warnings: 0}
	got := planNotes(r)
	if !strings.Contains(got.Assessment, "No errors or warnings") {
		t.Errorf("planNotes with no issues should say 'No errors or warnings', got %q", got.Assessment)
	}
}

func TestPlanNotes_ratingFormat(t *testing.T) {
	for _, total := range []int{0, 50, 84, 100, 130, 140} {
		r := &scorer.Result{Total: total}
		got := planNotes(r)
		if !reNotesRating.MatchString(got.Rating) {
			t.Errorf("planNotes(total=%d).Rating = %q does not match N/10 pattern", total, got.Rating)
		}
	}
}

// ---- planSkillName ----

func TestPlanSkillName_normalises(t *testing.T) {
	name := planSkillName("domain/skill-name")
	if strings.Contains(name, "/") {
		t.Errorf("planSkillName should replace slashes, got %q", name)
	}
	if !reSkillPattern.MatchString(name) {
		t.Errorf("planSkillName result %q does not match kebab-case pattern", name)
	}
}

func TestPlanSkillName_withSKILLMDSuffix(t *testing.T) {
	got := planSkillName("domain/my-skill/SKILL.md")
	if strings.Contains(got, "skill.md") {
		t.Errorf("planSkillName should strip SKILL.md suffix, got %q", got)
	}
}

func TestPlanSkillName_withMDSuffix(t *testing.T) {
	got := planSkillName("domain/my-skill/README.md")
	if strings.HasSuffix(got, ".md") {
		t.Errorf("planSkillName should strip .md suffix, got %q", got)
	}
}

func TestPlanSkillName_dotPath(t *testing.T) {
	got := planSkillName("my-skill")
	if got == "" {
		t.Error("planSkillName should return non-empty for simple name")
	}
}

// ---- splitAdviceIntoSteps ----

func TestSplitAdviceIntoSteps_multiSentence(t *testing.T) {
	advice := "Do this first. Then do that. Finally wrap up."
	steps := splitAdviceIntoSteps(advice, 1)
	if len(steps) < 2 {
		t.Errorf("expected multiple steps, got %d: %v", len(steps), steps)
	}
}

func TestSplitAdviceIntoSteps_stepNumbering(t *testing.T) {
	advice := "Do this first. Then do that."
	steps := splitAdviceIntoSteps(advice, 2)
	if steps[0].Step != "2.1" {
		t.Errorf("first step should be 2.1, got %q", steps[0].Step)
	}
	if steps[1].Step != "2.2" {
		t.Errorf("second step should be 2.2, got %q", steps[1].Step)
	}
}

func TestSplitAdviceIntoSteps_hasRequiredFields(t *testing.T) {
	advice := "Do this important thing for the skill."
	steps := splitAdviceIntoSteps(advice, 1)
	if len(steps) == 0 {
		t.Fatal("expected at least one step")
	}
	s := steps[0]
	if !reStepPattern.MatchString(s.Step) {
		t.Errorf("step.Step %q does not match N.N pattern", s.Step)
	}
	if len(s.Title) < 3 {
		t.Errorf("step.Title %q is too short (min 3)", s.Title)
	}
	if len(s.Description) < 10 {
		t.Errorf("step.Description %q is too short (min 10)", s.Description)
	}
}

func TestSplitAdviceIntoSteps_shortAdvice(t *testing.T) {
	steps := splitAdviceIntoSteps("Hi.", 1)
	if len(steps) != 1 || steps[0].Description != "Hi." {
		t.Errorf("short advice should fall back to single step, got %v", steps)
	}
}

func TestSplitAdviceIntoSteps_emptyAdvice(t *testing.T) {
	steps := splitAdviceIntoSteps("", 1)
	if len(steps) != 1 {
		t.Errorf("empty advice should produce exactly 1 step, got %v", steps)
	}
}

func TestSplitAdviceIntoSteps_alreadyEndsWithDot(t *testing.T) {
	advice := "Do this correctly. It matters."
	steps := splitAdviceIntoSteps(advice, 1)
	for _, s := range steps {
		if !strings.HasSuffix(s.Description, ".") {
			t.Errorf("step description %q should end with '.'", s.Description)
		}
	}
}

// ---- buildGaps ----

func TestBuildGaps_allAtMax(t *testing.T) {
	r := makeResultWithScore(140)
	for _, d := range scorer.AllDimensions {
		r.Dimensions[d.Key] = d.Max
	}
	gaps := buildGaps(r)
	if len(gaps) != 0 {
		t.Errorf("expected 0 gaps for perfect score, got %d", len(gaps))
	}
}

func TestBuildGaps_missingDimensionKey(t *testing.T) {
	r := makeResultWithScore(84)
	delete(r.Dimensions, "evalValidation")
	gaps := buildGaps(r)
	for _, g := range gaps {
		if g.key == "evalValidation" {
			t.Error("missing dimension key should be skipped in buildGaps")
		}
	}
}
