package scorer

import "regexp"

var (
	reD2MindsetHeader = regexp.MustCompile(`(?im)##\s*(mindset|philosophy|principles)`)
	reD2NumberedList  = regexp.MustCompile(`(?m)^\s*[0-9]+\.`)
)

// scoreD2 — Mindset + Procedures (max: 15)
func scoreD2(content string, b *validatorBridge) (int, []Diagnostic) {
	score := 0
	var diags []Diagnostic

	if reD2MindsetHeader.MatchString(content) {
		score += 2
	}

	delta, structDiags := scoreD2Structure(content, b)
	score += delta
	diags = append(diags, structDiags...)

	if countPattern(content, "when to use") > 0 || countPattern(content, "when to apply") > 0 {
		score += 4
	}
	if countPattern(content, "when not to") > 0 {
		score += 3
	}

	if score > 15 {
		score = 15
	}
	return score, diags
}

func scoreD2Structure(content string, b *validatorBridge) (int, []Diagnostic) {
	if b.Content != nil {
		if b.Content.StrongMarkers == 0 && b.Content.WeakMarkers == 0 && b.Content.ImperativeRatio == 0 {
			return 0, []Diagnostic{warnDiag("D2", "no imperative or directive markers detected — procedural guidance may be missing")}
		}
		delta := 0
		switch {
		case b.Content.ImperativeRatio >= 0.4:
			delta = 4
		case b.Content.ImperativeRatio >= 0.25:
			delta = 3
		case b.Content.ImperativeRatio >= 0.1:
			delta = 2
		}
		if b.Content.ListItemCount > 3 {
			delta += 2
		} else if b.Content.ListItemCount > 0 {
			delta++
		}
		return delta, nil
	}
	if reD2NumberedList.MatchString(content) {
		return 2, nil
	}
	return 0, nil
}
