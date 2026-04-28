package duplication

import (
	"regexp"
	"strings"
)

var (
	mdFormatting = regexp.MustCompile(`(?m)^#{1,6}\s+|[*_` + "`" + `~|]|\[.*?\]\(.*?\)|^\s*[-*+]\s+|^\s*\d+\.\s+|^---+$|^===+$`)
	multiSpace   = regexp.MustCompile(`\s+`)
	headerLine   = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
)

var stopwords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true, "to": true, "of": true,
	"and": true, "or": true, "in": true, "on": true, "at": true, "for": true,
	"with": true, "by": true, "from": true, "it": true, "its": true, "this": true,
	"that": true, "you": true, "we": true, "use": true, "can": true, "will": true,
	"if": true, "as": true, "not": true, "all": true, "when": true, "then": true,
	"your": true, "how": true, "what": true, "which": true, "have": true, "has": true,
}

// TokenSet strips markdown formatting from text, lowercases, splits on whitespace,
// removes stopwords, and returns the resulting word set.
func TokenSet(text string) map[string]bool {
	clean := mdFormatting.ReplaceAllString(text, " ")
	clean = multiSpace.ReplaceAllString(clean, " ")
	tokens := strings.Fields(strings.ToLower(clean))
	set := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		// strip leading/trailing punctuation
		t = strings.Trim(t, ".,;:!?\"'()[]{}/<>=+\\")
		if len(t) > 2 && !stopwords[t] {
			set[t] = true
		}
	}
	return set
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
	if union == 0 {
		return 0
	}
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
