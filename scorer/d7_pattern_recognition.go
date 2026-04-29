package scorer

import (
	"fmt"
	"regexp"
	"strings"
)

// scoreD7 — Pattern Recognition (max: 10)
// Uses library description length (chars) from CheckFrontmatter.
// Falls back to word-count heuristic when library result unavailable.
// Also calls scoreDiscriminativeness to emit diagnostic-only signal (no numeric change).
func scoreD7(b *validatorBridge) (int, []Diagnostic) {
	var diags []Diagnostic

	descLen := b.descriptionLen()
	var score int
	if descLen >= 0 {
		// Library has the char count; map to score bands.
		switch {
		case descLen > 200:
			score = 10
		case descLen > 120:
			score = 9
		case descLen > 60:
			score = 8
		default:
			if descLen <= 30 {
				diags = append(diags, warnDiag("D7", fmt.Sprintf("description is only %d chars — aim for >60 for a useful pattern signal", descLen)))
			}
			score = 6
		}
	} else {
		// Fallback: library unavailable.
		if b.Structure == nil {
			return 6, append(diags, warnDiag("D7", "description length unavailable (validator bridge failed)"))
		}
		score = 6
	}

	// Discriminativeness signal: diagnostics only, no numeric change this iteration.
	diags = append(diags, scoreDiscriminativeness(b.rawDescription())...)
	return score, diags
}

// negativeAnchorRe matches phrases indicating a skill explicitly states when NOT to trigger.
var negativeAnchorRe = regexp.MustCompile(
	`(?i)\b(does not apply|skip when|not for|exclude|do not trigger|not intended for)\b`,
)

// workflowAnchorRe matches artifact nouns that ground a trigger in a concrete workflow context.
var workflowAnchorRe = regexp.MustCompile(
	`(?i)\b(file|pr|commit|test|config|pipeline|migration)\b`,
)

// scoreDiscriminativeness checks whether the description contains discriminativeness anchors
// and returns a diagnostic signal. It never changes the numeric score.
//
// Rules:
//   - Both anchors present (negative + workflow) → hint: positive discriminativeness signal.
//   - Neither anchor present                     → WARN: description may over-trigger.
//   - Only one anchor present                    → no diagnostic (neutral).
func scoreDiscriminativeness(desc string) []Diagnostic {
	if strings.TrimSpace(desc) == "" {
		return nil
	}
	hasNegative := negativeAnchorRe.MatchString(desc)
	hasWorkflow := workflowAnchorRe.MatchString(desc)

	switch {
	case hasNegative && hasWorkflow:
		return []Diagnostic{hintDiag("D7", "description includes negative and workflow anchors — good discriminativeness signal")}
	case !hasNegative && !hasWorkflow:
		return []Diagnostic{warnDiag("D7", "description lacks negative and workflow anchors — skill may over-trigger on adjacent topics")}
	default:
		// Exactly one anchor: neutral, no diagnostic.
		return nil
	}
}
