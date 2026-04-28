package reporter

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/scorer"
	"gopkg.in/yaml.v3"
)

// ---- YAML structs matching remediation-plan.schema.json ----

type remPlanFrontmatter struct {
	PlanDate         string         `yaml:"plan_date"`
	SkillName        string         `yaml:"skill_name"`
	SourceAudit      string         `yaml:"source_audit"`
	ExecutiveSummary remExecSummary `yaml:"executive_summary"`
	CriticalIssues   []remCritical  `yaml:"critical_issues"`
	RemPhases        []remPhase     `yaml:"remediation_phases"`
	VerifCmds        []string       `yaml:"verification_commands"`
	SuccessCriteria  []string       `yaml:"success_criteria"`
	EffortEstimates  []remEffort    `yaml:"effort_estimates"`
	Dependencies     []string       `yaml:"dependencies"`
	RollbackPlan     string         `yaml:"rollback_plan"`
	Notes            string         `yaml:"notes"`
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
	Phase     string   `yaml:"phase"`
	Dimension string   `yaml:"dimension"`
	Priority  string   `yaml:"priority"`
	Target    string   `yaml:"target"`
	Steps     []string `yaml:"steps"`
}

type remEffort struct {
	Phase  string `yaml:"phase"`
	Effort string `yaml:"effort"`
	Time   string `yaml:"time"`
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

	gaps := buildGaps(r)
	sort.Slice(gaps, func(i, j int) bool {
		return (gaps[i].max - gaps[i].score) > (gaps[j].max - gaps[j].score)
	})

	totalGap := r.MaxTotal - r.Total
	overallPriority := gapPriority(totalGap)
	overallEffort := gapEffort(totalGap)

	focusAreas := make([]string, 0, 3)
	for i, g := range gaps {
		if i >= 3 {
			break
		}
		focusAreas = append(focusAreas, dimLabelToCode(g.label)+": "+g.label)
	}

	criticals := buildCriticalIssues(gaps)
	phases := buildPhases(gaps)
	efforts := buildEffortEstimates(gaps)

	pct := func(score int) int { return int(math.Round(float64(score) / 140.0 * 100)) }

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
			Priority:   overallPriority,
			Effort:     overallEffort,
			FocusAreas: focusAreas,
			Verdict:    planVerdict(overallPriority),
		},
		CriticalIssues:  criticals,
		RemPhases:       phases,
		VerifCmds:       verifCommands(skillName),
		SuccessCriteria: successCriteria(r, targetScore),
		EffortEstimates: efforts,
		Dependencies:    []string{"Completed audit stored in .context/audits/"},
		RollbackPlan:    "git checkout HEAD -- skills/" + skillName + "/SKILL.md",
		Notes:           planNotes(r),
	}

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

	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString("| Field | Current | Target |\n")
	sb.WriteString("|-------|---------|--------|\n")
	fmt.Fprintf(&sb, "| Score | %d/140 (%d%%) | %d/140 (%d%%) |\n",
		r.Total, pct(r.Total), targetScore, pct(targetScore))
	fmt.Fprintf(&sb, "| Grade | %s | %s |\n", r.Grade, scorer.Grade(targetScore))
	fmt.Fprintf(&sb, "| Priority | %s | — |\n\n", overallPriority)

	sb.WriteString("## Critical Issues\n\n")
	sb.WriteString("| Issue | Dimension | Severity | Impact |\n")
	sb.WriteString("|-------|-----------|----------|--------|\n")
	for _, c := range criticals {
		fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n", c.Issue, c.Dimension, c.Severity, c.Impact)
	}
	sb.WriteString("\n")

	sb.WriteString("## Remediation Phases\n\n")
	for _, ph := range phases {
		fmt.Fprintf(&sb, "### %s\n\n", ph.Phase)
		fmt.Fprintf(&sb, "**Dimension:** %s  \n", ph.Dimension)
		fmt.Fprintf(&sb, "**Target:** %s  \n", ph.Target)
		fmt.Fprintf(&sb, "**Priority:** %s\n\n", ph.Priority)
		for _, step := range ph.Steps {
			fmt.Fprintf(&sb, "- %s\n", step)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Verification Commands\n\n")
	for _, cmd := range verifCommands(skillName) {
		fmt.Fprintf(&sb, "```bash\n%s\n```\n\n", cmd)
	}

	sb.WriteString("## Success Criteria\n\n")
	for _, sc := range successCriteria(r, targetScore) {
		fmt.Fprintf(&sb, "- [ ] %s\n", sc)
	}
	sb.WriteString("\n")

	return sb.String(), nil
}

// ---- Validation ----

// ValidateRemediationPlan parses YAML frontmatter in planPath and returns
// a list of validation errors. An empty slice means the plan is valid.
func ValidateRemediationPlan(planPath string) []string {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return []string{fmt.Sprintf("cannot read file: %v", err)}
	}

	fm, err := extractFrontmatter(data)
	if err != nil {
		return []string{fmt.Sprintf("frontmatter parse error: %v", err)}
	}

	return validateFrontmatter(fm)
}

func extractFrontmatter(data []byte) (*remPlanFrontmatter, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var lines []string
	inFM := false
	fmLines := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if len(lines) == 1 && line == "---" {
			inFM = true
			continue
		}
		if inFM {
			if line == "---" {
				break
			}
			fmLines = append(fmLines, line)
		}
	}
	if !inFM {
		return nil, fmt.Errorf("no YAML frontmatter found (file must start with ---)")
	}

	var fm remPlanFrontmatter
	if err := yaml.Unmarshal([]byte(strings.Join(fmLines, "\n")), &fm); err != nil {
		return nil, err
	}
	return &fm, nil
}

var (
	reDatePattern   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	reSkillPattern  = regexp.MustCompile(`^[a-z0-9-]+$`)
	reAuditPattern  = regexp.MustCompile(`^\.context/audits/.+\.md$`)
	reScorePattern  = regexp.MustCompile(`^\d+/140\s*\(\d+%\)$`)
	reDimPattern    = regexp.MustCompile(`^D[1-9]:\s+.+\s+\(\d+/\d+\)$`)
	validGrades     = map[string]bool{"A+": true, "A": true, "B+": true, "B": true, "C+": true, "C": true, "D": true, "F": true}
	validSeverities = map[string]bool{"Critical": true, "High": true, "Medium": true, "Low": true}
	validPriorities = map[string]bool{"Critical": true, "High": true, "Medium": true, "Low": true}
	validEfforts    = map[string]bool{"S": true, "M": true, "L": true}
)

func validateFrontmatter(fm *remPlanFrontmatter) []string {
	var errs []string
	check := func(cond bool, msg string) {
		if !cond {
			errs = append(errs, msg)
		}
	}

	check(reDatePattern.MatchString(fm.PlanDate), "plan_date must match YYYY-MM-DD")
	check(reSkillPattern.MatchString(fm.SkillName), "skill_name must be kebab-case [a-z0-9-]+")
	check(reAuditPattern.MatchString(fm.SourceAudit), "source_audit must match .context/audits/.../*.md")

	check(reScorePattern.MatchString(fm.ExecutiveSummary.Score.Current), "executive_summary.score.current must match NNN/140 (NN%)")
	check(reScorePattern.MatchString(fm.ExecutiveSummary.Score.Target), "executive_summary.score.target must match NNN/140 (NN%)")
	check(validGrades[fm.ExecutiveSummary.Grade.Current], "executive_summary.grade.current must be a valid grade (A+/A/B+/B/C+/C/D/F)")
	check(validGrades[fm.ExecutiveSummary.Grade.Target], "executive_summary.grade.target must be a valid grade (A+/A/B+/B/C+/C/D/F)")
	check(validPriorities[fm.ExecutiveSummary.Priority], "executive_summary.priority must be Critical/High/Medium/Low")
	check(validEfforts[fm.ExecutiveSummary.Effort], "executive_summary.effort must be S/M/L")
	check(len(fm.ExecutiveSummary.FocusAreas) > 0, "executive_summary.focus_areas must have at least one entry")
	check(len(fm.ExecutiveSummary.Verdict) >= 10, "executive_summary.verdict must be at least 10 characters")

	check(len(fm.CriticalIssues) >= 1, "critical_issues must have at least one entry")
	for i, c := range fm.CriticalIssues {
		pfx := fmt.Sprintf("critical_issues[%d]", i)
		check(len(c.Issue) >= 10, pfx+".issue must be at least 10 characters")
		check(reDimPattern.MatchString(c.Dimension), pfx+".dimension must match 'D#: Name (score/max)'")
		check(validSeverities[c.Severity], pfx+".severity must be Critical/High/Medium/Low")
		check(len(c.Impact) >= 10, pfx+".impact must be at least 10 characters")
	}

	check(len(fm.RemPhases) >= 1, "remediation_phases must have at least one entry")
	check(len(fm.VerifCmds) >= 1, "verification_commands must have at least one entry")
	check(len(fm.SuccessCriteria) >= 1, "success_criteria must have at least one entry")
	check(len(fm.EffortEstimates) >= 1, "effort_estimates must have at least one entry")
	for i, e := range fm.EffortEstimates {
		check(validEfforts[e.Effort], fmt.Sprintf("effort_estimates[%d].effort must be S/M/L", i))
	}
	check(len(fm.RollbackPlan) >= 1, "rollback_plan must not be empty")
	check(len(fm.Notes) >= 1, "notes must not be empty")

	return errs
}

// ---- Helpers ----

func planSkillName(skill string) string {
	// normalise to kebab-case compatible (lowercase, replace path sep with -)
	base := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(skill, "/", "-"), "\\", "-"))
	// strip SKILL.md suffix if present
	base = strings.TrimSuffix(base, "-skill.md")
	base = strings.TrimSuffix(base, ".md")
	return base
}

func buildGaps(r *scorer.Result) []gap {
	gaps := make([]gap, 0, len(dimensionOrder))
	for _, d := range dimensionOrder {
		score, ok := r.Dimensions[d.key]
		if !ok {
			continue
		}
		if score < d.max {
			gaps = append(gaps, gap{key: d.key, label: d.label, score: score, max: d.max})
		}
	}
	return gaps
}

func buildCriticalIssues(gaps []gap) []remCritical {
	issues := make([]remCritical, 0, min(5, len(gaps)))
	for i, g := range gaps {
		if i >= 5 {
			break
		}
		gapSize := g.max - g.score
		dcode := dimLabelToCode(g.label)
		issues = append(issues, remCritical{
			Issue:     fmt.Sprintf("%s scores %d/%d (%d points below maximum)", g.label, g.score, g.max, gapSize),
			Dimension: fmt.Sprintf("%s: %s (%d/%d)", dcode, g.label, g.score, g.max),
			Severity:  gapPriority(gapSize),
			Impact:    fmt.Sprintf("Missing %d pts prevents grade improvement", gapSize),
		})
	}
	return issues
}

func buildPhases(gaps []gap) []remPhase {
	phases := make([]remPhase, 0, len(gaps))
	for i, g := range gaps {
		gapSize := g.max - g.score
		dcode := dimLabelToCode(g.label)
		advice := dimensionAdvice[g.key]
		steps := splitAdviceIntoSteps(advice)
		phases = append(phases, remPhase{
			Phase:     fmt.Sprintf("Phase %d: %s", i+1, g.label),
			Dimension: fmt.Sprintf("%s: %s (%d/%d)", dcode, g.label, g.score, g.max),
			Priority:  gapPriority(gapSize),
			Target:    fmt.Sprintf("Increase from %d/%d to %d/%d (+%d points)", g.score, g.max, g.max, g.max, gapSize),
			Steps:     steps,
		})
	}
	return phases
}

func buildEffortEstimates(gaps []gap) []remEffort {
	estimates := make([]remEffort, 0, len(gaps)+1)
	totalGap := 0
	for i, g := range gaps {
		gapSize := g.max - g.score
		totalGap += gapSize
		estimates = append(estimates, remEffort{
			Phase:  fmt.Sprintf("Phase %d: %s", i+1, g.label),
			Effort: gapEffort(gapSize),
			Time:   gapTime(gapSize),
		})
	}
	estimates = append(estimates, remEffort{
		Phase:  "Total",
		Effort: gapEffort(totalGap),
		Time:   fmt.Sprintf("%d hours", max(1, totalGap/10)),
	})
	return estimates
}

func splitAdviceIntoSteps(advice string) []string {
	// Split on '. ' or '. \n' to create individual steps; fallback to single step.
	sentences := strings.Split(advice, ". ")
	var steps []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 5 {
			if !strings.HasSuffix(s, ".") {
				s += "."
			}
			steps = append(steps, s)
		}
	}
	if len(steps) == 0 {
		steps = []string{advice}
	}
	return steps
}

func gapPriority(gap int) string {
	switch {
	case gap >= 10:
		return "Critical"
	case gap >= 6:
		return "High"
	case gap >= 3:
		return "Medium"
	default:
		return "Low"
	}
}

func gapEffort(gap int) string {
	switch {
	case gap >= 20:
		return "L"
	case gap >= 8:
		return "M"
	default:
		return "S"
	}
}

func gapTime(gap int) string {
	switch {
	case gap >= 20:
		return "3+ hours"
	case gap >= 8:
		return "1-2 hours"
	default:
		return "30 min"
	}
}

func planVerdict(priority string) string {
	switch priority {
	case "Critical":
		return "Immediate improvements required to reach acceptable grade"
	case "High":
		return "Priority improvements required before publication"
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

func successCriteria(r *scorer.Result, targetScore int) []string {
	targetGrade := scorer.Grade(targetScore)
	return []string{
		fmt.Sprintf("Total score reaches %d/140 or higher", targetScore),
		fmt.Sprintf("Grade improves from %s to %s", r.Grade, targetGrade),
		"No Critical-severity diagnostics remain",
		"All remediation phase steps completed",
	}
}

func planNotes(r *scorer.Result) string {
	if r.Errors > 0 {
		return fmt.Sprintf("Audit reported %d error(s) and %d warning(s). Address errors first.", r.Errors, r.Warnings)
	}
	if r.Warnings > 0 {
		return fmt.Sprintf("Audit reported %d warning(s). Review before publishing.", r.Warnings)
	}
	return "No errors or warnings from the audit. Focus on dimension gaps above."
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
