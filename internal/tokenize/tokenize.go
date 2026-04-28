package tokenize

import (
	"regexp"
	"strings"
)

var (
	MDFormatting = regexp.MustCompile(`(?m)^#{1,6}\s+|[*_` + "`" + `~|]|\[.*?\]\(.*?\)|^\s*[-*+]\s+|^\s*\d+\.\s+|^---+$|^===+$`)
	multiSpace   = regexp.MustCompile(`\s+`)
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

// Normalize strips markdown formatting, lowercases, and returns the token list
// with leading/trailing punctuation removed and stopwords/short tokens filtered.
func Normalize(text string) []string {
	clean := MDFormatting.ReplaceAllString(text, " ")
	clean = multiSpace.ReplaceAllString(clean, " ")
	raw := strings.Fields(strings.ToLower(clean))
	tokens := make([]string, 0, len(raw))
	for _, t := range raw {
		t = strings.Trim(t, ".,;:!?\"'()[]{}/<>=+\\")
		if len(t) > 2 && !stopwords[t] {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

// Set returns the unique token set for text.
func Set(text string) map[string]bool {
	tokens := Normalize(text)
	set := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		set[t] = true
	}
	return set
}

// Counts returns term → frequency for text.
func Counts(text string) map[string]int {
	tokens := Normalize(text)
	counts := make(map[string]int, len(tokens))
	for _, t := range tokens {
		counts[t]++
	}
	return counts
}
