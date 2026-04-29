// Package reporter formats and persists scorer results, analysis reports,
// aggregation plans, and remediation plans.
// This file owns remediation plan generation: building the YAML frontmatter
// struct and rendering the full markdown document.
package reporter

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
	"gopkg.in/yaml.v3"
)

// ---- YAML structs matching remediation-plan.schema.json ----

type remPlanFrontmatter struct {
	PlanDate         string                `yaml:"plan_date"`
	SkillName        string                `yaml:"skill_name"`
	SourceAudit      string                `yaml:"source_audit"`
	ExecutiveSummary remExecSummary        `yaml:"executive_summary"`
	CriticalIssues   []remCritical         `yaml:"critical_issues"`
	RemPhases        []remPhase            `yaml:"remediation_phases"`
	VerifCmds        []string              `yaml:"verification_commands"`
	SuccessCriteria  []remSuccessCriterion `yaml:"success_criteria"`
	EffortEstimates  []remEffort           `yaml:"effort_estimates"`
	Dependencies     []string              `yaml:"dependencies"`
	RollbackPlan     string                `yaml:"rollback_plan"`
	Notes            remNotes              `yaml:"notes"`
}

type remExecSummary struct {
	Score      remScoreRange `yaml:"score"`
	Grade      remGradeRange `yaml:"grade"`
	Priority   string        `yaml:"priority"`
	Effort     string        `yaml:"effort"`
	FocusAreas []string      `yaml:"focus_areas"`
	Verdict    string        `yaml:"verdict"`
}

type remScoreRange struct {
	Current string `yaml:"current"`
	Target  string `yaml:"target"`
}

type remGradeRange struct {
	Current string `yaml:"current"`
	Target  string `yaml:"target"`
}

type remCritical struct {
	Issue     string `yaml:"issue"`
	Dimension string `yaml:"dimension"`
	Severity  string `yaml:"severity"`
	Impact    string `yaml:"impact"`
}

type remPhase struct {
	Phase     int       `yaml:"phase"`
	Dimension string    `yaml:"dimension"`
	Priority  string    `yaml:"priority"`
	Target    string    `yaml:"target"`
	Steps     []remStep `yaml:"steps"`
}

type remStep struct {
	Step        string   `yaml:"step"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	File        string   `yaml:"file,omitempty"`
	CodeBlock   *remCode `yaml:"code_block,omitempty"`
}

type remCode struct {
	Language string `yaml:"language"`
	Content  string `yaml:"content"`
}

type remSuccessCriterion struct {
	Criterion   string `yaml:"criterion"`
	Measurement string `yaml:"measurement"`
}

type remEffort struct {
	Phase  string `yaml:"phase"`
	Effort string `yaml:"effort"`
	Time   string `yaml:"time"`
}

type remNotes struct {
	Rating     string `yaml:"rating"`
	Assessment string `yaml:"assessment"`
}

// ---- Generation ----

// RemediationPlan generates a schema-compliant YAML-frontmatter + markdown
// remediation plan. targetScore ≤ 0 defaults to min(current+20, 140).
func RemediationPlan(r *scorer.Result, targetScore int, auditPath, date string) (string, error) {
	if targetScore > 0 && targetScore <= r.Total {
		return "", fmt.Errorf("targetScore %d must exceed current score %d", targetScore, r.Total)
	}
	if targetScore <= 0 || targetScore > 140 {
		targetScore = r.Total + 20
		if targetScore > 140 {
			targetScore = 140
		}
	}

	skillName := planSkillName(r.Skill)
	if auditPath == "" {
		auditPath = fmt.Sprintf(".context/audits/%s/%s/Analysis.md", skillName, r.Date)
	}

	fm, gaps, err := buildRemediationFrontmatter(r, targetScore, skillName, auditPath, date)
	if err != nil {
		return "", err
	}

	return renderRemediationPlan(r, fm, gaps, targetScore, date)
}

func buildRemediationFrontmatter(r *scorer.Result, targetScore int, skillName, auditPath, date string) (remPlanFrontmatter, []gap, error) {
	gaps := buildGaps(r)
	sort.Slice(gaps, func(i, j int) bool {
		return (gaps[i].max - gaps[i].score) > (gaps[j].max - gaps[j].score)
	})

	totalGap := r.MaxTotal - r.Total
	pct := func(score int) int { return int(math.Round(float64(score) / 140.0 * 100)) }

	focusAreas := make([]string, 0, 3)
	for i, g := range gaps {
		if i >= 3 {
			break
		}
		focusAreas = append(focusAreas, dimLabelToCode(g.label)+": "+g.label)
	}

	fm := remPlanFrontmatter{
		PlanDate:    date,
		SkillName:   skillName,
		SourceAudit: auditPath,
		ExecutiveSummary: remExecSummary{
			Score: remScoreRange{
				Current: fmt.Sprintf("%d/140 (%d%%)", r.Total, pct(r.Total)),
				Target:  fmt.Sprintf("%d/140 (%d%%)", targetScore, pct(targetScore)),
			},
			Grade: remGradeRange{
				Current: r.Grade,
				Target:  scorer.Grade(targetScore),
			},
			Priority:   gapPriority(totalGap),
			Effort:     gapEffort(totalGap),
			FocusAreas: focusAreas,
			Verdict:    planVerdict(gapPriority(totalGap)),
		},
		CriticalIssues:  buildCriticalIssues(gaps),
		RemPhases:       buildPhases(gaps),
		VerifCmds:       verifCommands(skillName),
		SuccessCriteria: successCriteria(r, targetScore),
		EffortEstimates: buildEffortEstimates(gaps),
		Dependencies:    []string{"Completed audit stored in .context/audits/"},
		RollbackPlan:    "git checkout HEAD -- skills/" + skillName + "/SKILL.md",
		Notes:           planNotes(r),
	}
	return fm, gaps, nil
}

func renderRemediationPlan(r *scorer.Result, fm remPlanFrontmatter, _ []gap, targetScore int, date string) (string, error) {
	pct := func(score int) int { return int(math.Round(float64(score) / 140.0 * 100)) }

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("marshal frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")

	fmt.Fprintf(&sb, "# Remediation Plan — %s\n\n", r.Skill)
	fmt.Fprintf(&sb, "**Generated:** %s  \n", date)
	fmt.Fprintf(&sb, "**Current:** %s (%d/140)  \n", r.Grade, r.Total)
	fmt.Fprintf(&sb, "**Target:** %s (%d/140)\n\n", scorer.Grade(targetScore), targetScore)

	writeRemediationExecSummary(&sb, r, fm, targetScore, pct)
	writeRemediationCriticalIssues(&sb, fm.CriticalIssues)
	writeRemediationPhases(&sb, fm.RemPhases)
	writeRemediationVerifAndCriteria(&sb, fm)

	return sb.String(), nil
}

func writeRemediationExecSummary(sb *strings.Builder, r *scorer.Result, fm remPlanFrontmatter, targetScore int, pct func(int) int) {
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString("| Field | Current | Target |\n")
	sb.WriteString("|-------|---------|--------|\n")
	fmt.Fprintf(sb, "| Score | %d/140 (%d%%) | %d/140 (%d%%) |\n",
		r.Total, pct(r.Total), targetScore, pct(targetScore))
	fmt.Fprintf(sb, "| Grade | %s | %s |\n", r.Grade, scorer.Grade(targetScore))
	fmt.Fprintf(sb, "| Priority | %s | — |\n\n", fm.ExecutiveSummary.Priority)
}

func writeRemediationCriticalIssues(sb *strings.Builder, criticals []remCritical) {
	sb.WriteString("## Critical Issues\n\n")
	sb.WriteString("| Issue | Dimension | Severity | Impact |\n")
	sb.WriteString("|-------|-----------|----------|--------|\n")
	for _, c := range criticals {
		fmt.Fprintf(sb, "| %s | %s | %s | %s |\n", c.Issue, c.Dimension, c.Severity, c.Impact)
	}
	sb.WriteString("\n")
}

func writeRemediationPhases(sb *strings.Builder, phases []remPhase) {
	sb.WriteString("## Remediation Phases\n\n")
	for _, ph := range phases {
		fmt.Fprintf(sb, "### Phase %d\n\n", ph.Phase)
		fmt.Fprintf(sb, "**Dimension:** %s  \n", ph.Dimension)
		fmt.Fprintf(sb, "**Target:** %s  \n", ph.Target)
		fmt.Fprintf(sb, "**Priority:** %s\n\n", ph.Priority)
		for _, step := range ph.Steps {
			fmt.Fprintf(sb, "- **%s** (`%s`): %s\n", step.Title, step.Step, step.Description)
		}
		sb.WriteString("\n")
	}
}

func writeRemediationVerifAndCriteria(sb *strings.Builder, fm remPlanFrontmatter) {
	sb.WriteString("## Verification Commands\n\n")
	for _, cmd := range fm.VerifCmds {
		fmt.Fprintf(sb, "```bash\n%s\n```\n\n", cmd)
	}
	sb.WriteString("## Success Criteria\n\n")
	for _, sc := range fm.SuccessCriteria {
		fmt.Fprintf(sb, "- [ ] %s: %s\n", sc.Criterion, sc.Measurement)
	}
	sb.WriteString("\n")
}

// ---- Helpers ----

func planSkillName(skill string) string {
	base := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(skill, "/", "-"), "\\", "-"))
	// strip trailing .md file component (e.g. SKILL.md, README.md)
	if idx := strings.LastIndex(base, "-"); idx >= 0 && strings.HasSuffix(base, ".md") {
		candidate := base[:idx]
		if !strings.Contains(base[idx+1:], ".") || strings.EqualFold(base[idx+1:], "skill.md") || strings.HasSuffix(base[idx+1:], ".md") {
			base = candidate
		}
	}
	return strings.Trim(base, "-")
}

func buildGaps(r *scorer.Result) []gap {
	gaps := make([]gap, 0, len(scorer.AllDimensions))
	for _, d := range scorer.AllDimensions {
		score, ok := r.Dimensions[d.Key]
		if !ok || score >= d.Max {
			continue
		}
		gaps = append(gaps, gap{key: d.Key, label: d.Label, score: score, max: d.Max})
	}
	return gaps
}

func buildCriticalIssues(gaps []gap) []remCritical {
	criticals := make([]remCritical, 0, len(gaps))
	for _, g := range gaps {
		avail := g.max - g.score
		sev := "Low"
		switch {
		case avail >= 10:
			sev = "Critical"
		case avail >= 7:
			sev = "High"
		case avail >= 4:
			sev = "Medium"
		}
		code := dimLabelToCode(g.label)
		criticals = append(criticals, remCritical{
			Issue:     fmt.Sprintf("%s scores %d/%d (%d pts below max)", g.label, g.score, g.max, avail),
			Dimension: fmt.Sprintf("%s: %s (%d/%d)", code, g.label, g.score, g.max),
			Severity:  sev,
			Impact:    fmt.Sprintf("Missing %d/%d points reduces total score by %.0f%%", avail, g.max, float64(avail)/140.0*100),
		})
	}
	return criticals
}

func buildPhases(gaps []gap) []remPhase {
	phases := make([]remPhase, 0, len(gaps))
	for i, g := range gaps {
		code := dimLabelToCode(g.label)
		advice := dimensionAdvice[g.key]
		steps := splitAdviceIntoSteps(advice, i+1)
		phases = append(phases, remPhase{
			Phase:     i + 1,
			Dimension: fmt.Sprintf("%s: %s", code, g.label),
			Priority:  gapPriority(g.max - g.score),
			Target:    fmt.Sprintf("Reach %d/%d", g.max, g.max),
			Steps:     steps,
		})
	}
	return phases
}

func buildEffortEstimates(gaps []gap) []remEffort {
	estimates := make([]remEffort, 0, len(gaps))
	total := 0
	for i, g := range gaps {
		avail := g.max - g.score
		h := gapHours(avail)
		total += h
		estimates = append(estimates, remEffort{
			Phase:  fmt.Sprintf("Phase %d", i+1),
			Effort: gapEffort(avail),
			Time:   gapTime(avail),
		})
	}
	estimates = append(estimates, remEffort{Phase: "Total", Effort: gapEffort(len(gaps) * 5), Time: fmt.Sprintf("%dh", total)})
	return estimates
}

func gapHours(gap int) int {
	switch {
	case gap >= 15:
		return 3
	case gap >= 5:
		return 2
	default:
		return 1
	}
}

func splitAdviceIntoSteps(advice string, phaseNum int) []remStep {
	if advice == "" {
		return []remStep{{
			Step:        fmt.Sprintf("%d.1", phaseNum),
			Title:       "Review and improve",
			Description: "Review the dimension rubric and apply targeted improvements.",
		}}
	}
	sentences := strings.Split(advice, ". ")
	steps := make([]remStep, 0, len(sentences))
	for i, s := range sentences {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !strings.HasSuffix(s, ".") {
			s += "."
		}
		steps = append(steps, remStep{
			Step:        fmt.Sprintf("%d.%d", phaseNum, i+1),
			Title:       shortStepTitle(s),
			Description: s,
		})
	}
	if len(steps) == 0 {
		return []remStep{{
			Step:        fmt.Sprintf("%d.1", phaseNum),
			Title:       "Apply advice",
			Description: advice,
		}}
	}
	return steps
}

func shortStepTitle(s string) string {
	words := strings.Fields(s)
	if len(words) > 5 {
		words = words[:5]
	}
	title := strings.Join(words, " ")
	title = strings.TrimRight(title, ".,;:")
	return title
}

func gapPriority(gap int) string {
	switch {
	case gap >= 10:
		return "Critical"
	case gap >= 7:
		return "High"
	case gap >= 4:
		return "Medium"
	default:
		return "Low"
	}
}

func gapEffort(gap int) string {
	switch {
	case gap >= 15:
		return "L"
	case gap >= 5:
		return "M"
	default:
		return "S"
	}
}

func gapTime(gap int) string {
	switch {
	case gap >= 15:
		return "3+ hours"
	case gap >= 5:
		return "1-2 hours"
	default:
		return "30 min"
	}
}

func planVerdict(priority string) string {
	switch priority {
	case "Critical":
		return "Immediate action required — significant gaps block production use"
	case "High":
		return "Priority improvements needed before publishing"
	case "Medium":
		return "Targeted improvements recommended"
	default:
		return "Minor refinements recommended"
	}
}

func verifCommands(skillName string) []string {
	return []string{
		"cd skill-auditor && go build -o skill-auditor . && ./skill-auditor evaluate " + skillName + " --store",
		"./skill-auditor evaluate " + skillName + " --json | jq '.grade'",
	}
}

func successCriteria(r *scorer.Result, targetScore int) []remSuccessCriterion {
	targetGrade := scorer.Grade(targetScore)
	return []remSuccessCriterion{
		{Criterion: "Total score target", Measurement: fmt.Sprintf("Score >= %d/140", targetScore)},
		{Criterion: "Grade improvement", Measurement: fmt.Sprintf(">= %s (from %s)", targetGrade, r.Grade)},
		{Criterion: "No critical diagnostics", Measurement: ">= 0 Critical issues resolved"},
		{Criterion: "All phase steps completed", Measurement: ">= all phases complete"},
	}
}

func planNotes(r *scorer.Result) remNotes {
	pct := r.Total * 100 / 140
	rating := fmt.Sprintf("%d/10", max(4, min(9, pct/10)))
	var assessment string
	switch {
	case r.Errors > 0:
		assessment = fmt.Sprintf("Audit reported %d error(s) and %d warning(s). Address errors first.", r.Errors, r.Warnings)
	case r.Warnings > 0:
		assessment = fmt.Sprintf("Audit reported %d warning(s). Review before publishing.", r.Warnings)
	default:
		assessment = "No errors or warnings from the audit. Focus on dimension gaps above."
	}
	return remNotes{Rating: rating, Assessment: assessment}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
