// This file owns simple remediation formatting: deriving a prioritised
// action plan directly from a scorer.Result without schema validation.
package reporter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

// dimensionAdvice holds generic improvement advice for each dimension when
// it scores below maximum and no specific diagnostic covers the gap.
var dimensionAdvice = map[string]string{
	"knowledgeDelta":          "Add expert-signal keywords: NEVER, ALWAYS, production, gotcha, pitfall, anti-pattern. Remove beginner-oriented patterns (npm install, getting started, hello world).",
	"mindsetProcedures":       "Add a `## Mindset` or `## Philosophy` section. Use numbered procedure lists. Add `## When to Use` and `## When NOT to Use` sections.",
	"antiPatternQuality":      "Add NEVER statements paired with `WHY:` explanations. Include BAD/GOOD contrast examples.",
	"specificationCompliance": "Expand the `description` frontmatter to >100 characters. Ensure no harness-specific paths, agent references, or `../` escapes outside code blocks.",
	"progressiveDisclosure":   "Add a `references/` directory with focused deep-dive `.md` files. Keep `SKILL.md` under 150 lines to maximise the score.",
	"freedomCalibration":      "Balance prescriptive language (NEVER/ALWAYS) with permissive alternatives (consider, optionally, may).",
	"patternRecognition":      "Expand the `description` frontmatter field to more than 15 qualifying words (words longer than 3 characters).",
	"practicalUsability":      "Add more fenced code blocks (aim for >5 pairs). Include `./` or `bun run` commands. Use language-tagged fences (```bash, ```typescript).",
	"evalValidation":          "Create an `evals/` directory with `instructions.json`, `summary.json`, and at least 3 scenario subdirectories each containing `task.md`, `criteria.json` (checklist summing to 100), and `capability.txt`.",
}

type gap struct {
	key   string
	label string
	score int
	max   int
}

// Remediation returns a markdown prioritised action plan derived from a Result.
func Remediation(r *scorer.Result) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Remediation Plan — %s\n\n", r.Skill)
	fmt.Fprintf(&sb, "**Current Grade:** %s (%d/%d)\n\n", r.Grade, r.Total, r.MaxTotal)

	gaps := buildSortedGaps(r)
	if len(gaps) == 0 {
		fmt.Fprintf(&sb, "All dimensions are at maximum score. Nothing to remediate.\n")
		return sb.String()
	}

	fmt.Fprintf(&sb, "## Priority Actions\n\n")
	diagsByDim := groupDiagsByDimension(r)
	for _, g := range gaps {
		writeGapSection(&sb, g, diagsByDim)
	}

	return sb.String()
}

func buildSortedGaps(r *scorer.Result) []gap {
	gaps := make([]gap, 0, len(scorer.AllDimensions))
	for _, d := range scorer.AllDimensions {
		score, ok := r.Dimensions[d.Key]
		if !ok || score >= d.Max {
			continue
		}
		gaps = append(gaps, gap{key: d.Key, label: d.Label, score: score, max: d.Max})
	}
	sort.Slice(gaps, func(i, j int) bool {
		return (gaps[i].max - gaps[i].score) > (gaps[j].max - gaps[j].score)
	})
	return gaps
}

func groupDiagsByDimension(r *scorer.Result) map[string][]scorer.Diagnostic {
	diagsByDim := map[string][]scorer.Diagnostic{}
	for _, d := range r.ErrorDetails {
		diagsByDim[d.Dimension] = append(diagsByDim[d.Dimension], d)
	}
	for _, d := range r.WarningDetails {
		diagsByDim[d.Dimension] = append(diagsByDim[d.Dimension], d)
	}
	return diagsByDim
}

func writeGapSection(sb *strings.Builder, g gap, diagsByDim map[string][]scorer.Diagnostic) {
	available := g.max - g.score
	fmt.Fprintf(sb, "### %s (%d/%d) — %d pt%s available\n\n", g.label, g.score, g.max, available, plural(available))
	dimKey := dimLabelToCode(g.label)
	for _, d := range diagsByDim[dimKey] {
		prefix := "⚠️"
		if d.Severity() == "error" {
			prefix = "🔴"
		}
		fmt.Fprintf(sb, "%s %s\n\n", prefix, d.Message)
	}
	if advice, ok := dimensionAdvice[g.key]; ok {
		fmt.Fprintf(sb, "%s\n\n", advice)
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// dimLabelToCode maps a dimension display label to its D-code for diagnostic lookup.
func dimLabelToCode(label string) string {
	for _, d := range scorer.AllDimensions {
		if d.Label == label {
			return d.Code
		}
	}
	return ""
}
