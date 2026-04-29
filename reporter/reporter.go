package reporter

import (
	"fmt"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/scorer"
)

// Format returns a human-readable representation of a Result.
func Format(r *scorer.Result) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Skill: %s\n", r.Skill)
	fmt.Fprintf(&sb, "Grade: %s (%d/%d)\n", r.Grade, r.Total, r.MaxTotal)
	fmt.Fprintf(&sb, "\nDimensions:\n")

	for _, d := range scorer.AllDimensions {
		score, ok := r.Dimensions[d.Key]
		if !ok {
			continue
		}
		fmt.Fprintf(&sb, "  %-28s %2d/%d\n", d.Label, score, d.Max)
	}

	fmt.Fprintf(&sb, "\nErrors: %d  Warnings: %d\n", r.Errors, r.Warnings)

	if len(r.ErrorDetails) > 0 || len(r.WarningDetails) > 0 {
		fmt.Fprintf(&sb, "\nDiagnostics:\n")
		for _, d := range r.ErrorDetails {
			fmt.Fprintf(&sb, "  [E] %-3s %s\n", d.Dimension, d.Message)
		}
		for _, d := range r.WarningDetails {
			fmt.Fprintf(&sb, "  [W] %-3s %s\n", d.Dimension, d.Message)
		}
	}

	return sb.String()
}
