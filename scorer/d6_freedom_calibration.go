package scorer

import (
	"regexp"
	"strings"
)

// hard constraint marker patterns (uppercase, word-boundary matched)
var (
	reHardMarkers = regexp.MustCompile(`\b(MUST|NEVER|ALWAYS|REQUIRED|PROHIBITED)\b`)
	reSoftMarkers = regexp.MustCompile(`\b(PREFER|AVOID|BY DEFAULT|UNLESS|TYPICALLY|RECOMMENDED)\b`)
)

// whenNotToUsePatterns are case-insensitive substrings that indicate negative-scope guidance.
var whenNotToUsePatterns = []string{
	"when not to use",
	"do not use",
	"not intended for",
	"outside the scope",
	"avoid using",
}

// scoreCalibrationBalance scores whether the ratio of strong-to-weak markers
// falls in a balanced range (neither all-imperative nor all-permissive). (max: 5)
//   - 0.3–0.8 → 5 pts  (balanced)
//   - 0.2–0.3 or 0.8–0.9 → 3 pts  (marginal)
//   - outside that range → 1 pt
//   - zero markers → 0 pts
func scoreCalibrationBalance(b *validatorBridge) int {
	if b.Content == nil {
		return 0
	}
	if b.Content.StrongMarkers+b.Content.WeakMarkers == 0 {
		return 0
	}
	s := b.Content.InstructionSpecificity
	switch {
	case s >= 0.3 && s <= 0.8:
		return 5
	case (s >= 0.2 && s < 0.3) || (s > 0.8 && s <= 0.9):
		return 3
	default:
		return 1
	}
}

// scoreWhenNotToUse scores the presence of negative-scope guidance (max: 3).
// Returns 3 pts if any "when not to use" signal is found, 0 otherwise.
func scoreWhenNotToUse(content string) int {
	lower := strings.ToLower(content)
	for _, p := range whenNotToUsePatterns {
		if strings.Contains(lower, p) {
			return 3
		}
	}
	return 0
}

// scorePermissiveImperativeRatio scores the balance of permissive vs. imperative
// instruction style using b.Content.ImperativeRatio (max: 3).
//   - 3 pts: ratio in [0.3, 0.7] — well-balanced mix
//   - 2 pts: ratio in [0.2, 0.3) or (0.7, 0.8]
//   - 1 pt:  ratio outside that range but markers present
//   - 0 pts: zero markers (bridge unavailable or no markers)
func scorePermissiveImperativeRatio(b *validatorBridge) int {
	if b.Content == nil {
		return 0
	}
	if b.Content.StrongMarkers+b.Content.WeakMarkers == 0 {
		return 0
	}
	r := b.Content.ImperativeRatio
	switch {
	case r >= 0.3 && r <= 0.7:
		return 3
	case (r >= 0.2 && r < 0.3) || (r > 0.7 && r <= 0.8):
		return 2
	default:
		return 1
	}
}

// scoreConstraintTypology scores explicit hard vs. soft distinction (max: 4).
//   - 4 pts: ≥2 hard and ≥2 soft uppercase markers
//   - 3 pts: ≥1 hard and ≥1 soft, but fewer than 2 of one type
//   - 2 pts: both types present via bridge signals but no ALLCAPS markers
//   - 1 pt:  only hard or only soft markers, not both
//   - 0 pts: no constraint language detected
func scoreConstraintTypology(content string, b *validatorBridge) int {
	hardCount := len(reHardMarkers.FindAllString(content, -1))
	softCount := len(reSoftMarkers.FindAllString(content, -1))

	switch {
	case hardCount >= 2 && softCount >= 2:
		return 4
	case hardCount >= 1 && softCount >= 1:
		return 3
	case hardCount > 0 || softCount > 0:
		return 1
	}

	// No ALLCAPS markers found — fall back to bridge signals for implicit distinction.
	if b.Content != nil && b.Content.ImperativeRatio > 0 && b.Content.WeakMarkers > 0 {
		return 2
	}
	return 0
}

// scoreD6 — Freedom Calibration (max: 15).
// Composed of four sub-components:
//   - scoreCalibrationBalance        (max 5 pts)
//   - scoreWhenNotToUse              (max 3 pts — requires raw content)
//   - scorePermissiveImperativeRatio (max 3 pts)
//   - scoreConstraintTypology        (max 4 pts — requires raw content)
//
// The content parameter provides raw SKILL.md text for pattern scanning.
func scoreD6(content string, b *validatorBridge) (int, []Diagnostic) {
	if b.Content == nil {
		return 0, []Diagnostic{warnDiag("D6", "validator bridge unavailable — freedom calibration score defaulted to 0")}
	}
	if b.Content.StrongMarkers+b.Content.WeakMarkers == 0 {
		return 0, []Diagnostic{warnDiag("D6", "no directive markers (MUST/NEVER/ALWAYS/etc.) found — skill lacks explicit guidance")}
	}

	score := scoreCalibrationBalance(b) +
		scoreWhenNotToUse(content) +
		scorePermissiveImperativeRatio(b) +
		scoreConstraintTypology(content, b)

	if score > 15 {
		return 15, nil
	}
	if score < 0 {
		return 0, nil
	}
	return score, nil
}
