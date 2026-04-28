package duplication

import (
	"regexp"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/internal/tokenize"
)

var headerLine = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

// TokenSet strips markdown formatting from text, lowercases, splits on whitespace,
// removes stopwords, and returns the resulting word set.
func TokenSet(text string) map[string]bool {
	return tokenize.Set(text)
}

// Jaccard returns the Jaccard similarity coefficient for two token sets.
func Jaccard(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	intersection := 0
	for t := range a {
		if b[t] {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	return float64(intersection) / float64(union)
}

// SectionHeaders extracts the normalised header texts from markdown content.
func SectionHeaders(text string) []string {
	matches := headerLine.FindAllStringSubmatch(text, -1)
	headers := make([]string, 0, len(matches))
	for _, m := range matches {
		headers = append(headers, strings.ToLower(strings.TrimSpace(m[2])))
	}
	return headers
}

// StructuralSimilarity returns a Jaccard score over the section header sets.
func StructuralSimilarity(a, b []string) float64 {
	sa := make(map[string]bool, len(a))
	sb := make(map[string]bool, len(b))
	for _, h := range a {
		sa[h] = true
	}
	for _, h := range b {
		sb[h] = true
	}
	return Jaccard(sa, sb)
}

// Similarity combines word-level Jaccard (weight 0.7) and structural header
// Jaccard (weight 0.3) into a single [0,1] score.
func Similarity(textA, textB string) float64 {
	wordScore := Jaccard(TokenSet(textA), TokenSet(textB))
	structScore := StructuralSimilarity(SectionHeaders(textA), SectionHeaders(textB))
	return wordScore*0.7 + structScore*0.3
}
