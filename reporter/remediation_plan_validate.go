// This file owns remediation plan validation: parsing YAML frontmatter from
// a saved plan file and checking it against the schema constraints.
package reporter

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	reStepPattern   = regexp.MustCompile(`^\d+\.\d+$`)
	reNotesRating   = regexp.MustCompile(`^\d+/10$`)
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
	for i, ph := range fm.RemPhases {
		pfx := fmt.Sprintf("remediation_phases[%d]", i)
		check(ph.Phase >= 1, pfx+".phase must be >= 1")
		check(len(ph.Steps) >= 1, pfx+" must have at least one step")
		for j, s := range ph.Steps {
			spfx := fmt.Sprintf("%s.steps[%d]", pfx, j)
			check(reStepPattern.MatchString(s.Step), spfx+".step must match N.N pattern (e.g. 1.1)")
			check(len(s.Title) >= 3, spfx+".title must be at least 3 characters")
			check(len(s.Description) >= 10, spfx+".description must be at least 10 characters")
		}
	}

	check(len(fm.VerifCmds) >= 1, "verification_commands must have at least one entry")

	check(len(fm.SuccessCriteria) >= 1, "success_criteria must have at least one entry")
	for i, sc := range fm.SuccessCriteria {
		pfx := fmt.Sprintf("success_criteria[%d]", i)
		check(len(sc.Criterion) >= 5, pfx+".criterion must be at least 5 characters")
		check(strings.Contains(sc.Measurement, ">= "), pfx+".measurement must contain '>= '")
	}

	check(len(fm.EffortEstimates) >= 1, "effort_estimates must have at least one entry")
	for i, e := range fm.EffortEstimates {
		check(validEfforts[e.Effort], fmt.Sprintf("effort_estimates[%d].effort must be S/M/L", i))
	}

	check(len(fm.RollbackPlan) >= 1, "rollback_plan must not be empty")
	check(reNotesRating.MatchString(fm.Notes.Rating), "notes.rating must match N/10 pattern")
	check(len(fm.Notes.Assessment) >= 10, "notes.assessment must be at least 10 characters")

	return errs
}
