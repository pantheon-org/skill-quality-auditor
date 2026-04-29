package scorer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// scoreD5 — Progressive Disclosure (max: 15)
func scoreD5(content, skillDir string, b *validatorBridge) int {
	score, _, _, _ := scoreD5WithMeta(content, skillDir, b)
	return score
}

// scoreD5WithMeta returns the D5 score plus metadata used in the Result.
// Token thresholds are calibrated at ~8 tokens/line; falls back to line count
// when the library cannot produce a token count.
func scoreD5WithMeta(content, skillDir string, b *validatorBridge) (score, lines, refCount int, hasRefs bool) {
	refsDir := filepath.Join(skillDir, "references")
	if info, err := os.Stat(refsDir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(refsDir)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !strings.HasPrefix(e.Name(), ".") {
				hasRefs = true
				refCount++
			}
		}
	}

	lines = len(strings.Split(content, "\n"))

	tokens := b.skillMDTokens()
	var heuristicScore int
	if tokens > 0 {
		heuristicScore, lines, refCount, hasRefs = scoreD5ByTokens(tokens, lines, refCount, hasRefs)
	} else {
		heuristicScore, lines, refCount, hasRefs = scoreD5ByLines(lines, refCount, hasRefs)
	}

	negScore := scoreNegativeConditions(content)
	score = min(15, heuristicScore+negScore)
	return score, lines, refCount, hasRefs
}

func scoreD5ByTokens(tokens, lines, refCount int, hasRefs bool) (score, outLines, outRefCount int, outHasRefs bool) {
	if hasRefs {
		switch {
		case tokens < d5TokenCompact:
			return 13, lines, refCount, hasRefs
		case tokens < d5TokenModerate:
			return 13, lines, refCount, hasRefs
		case tokens < d5TokenVerbose:
			return 11, lines, refCount, hasRefs
		default:
			return 10, lines, refCount, hasRefs
		}
	}
	switch {
	case tokens < d5TokenModerate:
		return 12, lines, refCount, hasRefs
	case tokens < d5TokenLong:
		return 10, lines, refCount, hasRefs
	case tokens < d5TokenVeryLong:
		return 7, lines, refCount, hasRefs
	default:
		return 5, lines, refCount, hasRefs
	}
}

func scoreD5ByLines(lines, refCount int, hasRefs bool) (score, outLines, outRefCount int, outHasRefs bool) {
	if hasRefs {
		switch {
		case lines < d5LinesCompact:
			return 13, lines, refCount, hasRefs
		case lines < d5LinesModerate:
			return 13, lines, refCount, hasRefs
		case lines < d5LinesVerbose:
			return 11, lines, refCount, hasRefs
		default:
			return 10, lines, refCount, hasRefs
		}
	}
	switch {
	case lines < d5LinesVerbose:
		return 12, lines, refCount, hasRefs
	case lines < d5LinesLong:
		return 10, lines, refCount, hasRefs
	case lines < d5LinesVeryLong:
		return 7, lines, refCount, hasRefs
	default:
		return 5, lines, refCount, hasRefs
	}
}

// isReferenceSectionCompliant checks if ## References is the last H2 with ≥1 bullet link.
func isReferenceSectionCompliant(content string) bool {
	lines := strings.Split(content, "\n")
	lastH2 := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			lastH2 = strings.TrimPrefix(line, "## ")
		}
	}
	bulletLinkRe := regexp.MustCompile(`(?m)^- \[.+\]\(.+\)`)
	return strings.TrimSpace(lastH2) == "References" && bulletLinkRe.MatchString(content)
}

// scoreNegativeConditions scans markdown table rows in content for negative
// trigger language and returns 0–2 pts. Matching is restricted to table rows
// (lines beginning with '|') to avoid false positives from prose.
func scoreNegativeConditions(content string) int {
	keywords := []string{
		"skip if",
		"only when",
		"unless",
		"not needed when",
		"omit if",
	}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		lower := strings.ToLower(trimmed)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return 2
			}
		}
	}
	return 0
}
