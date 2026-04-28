package analysis

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

var (
	mdFormatting = regexp.MustCompile(`(?m)^#{1,6}\s+|[*_` + "`" + `~|]|\[.*?\]\(.*?\)|^\s*[-*+]\s+|^\s*\d+\.\s+|^---+$|^===+$`)
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

// KeywordScore holds a single term's TF-IDF result.
type KeywordScore struct {
	Term  string
	TF    float64
	IDF   float64
	Score float64
}

// TermFrequency returns a map of term → count for the given text,
// after stripping markdown, lowercasing, removing stopwords (len>2).
func TermFrequency(text string) map[string]int {
	if text == "" {
		return map[string]int{}
	}
	clean := mdFormatting.ReplaceAllString(text, " ")
	clean = multiSpace.ReplaceAllString(clean, " ")
	tokens := strings.Fields(strings.ToLower(clean))
	counts := make(map[string]int, len(tokens))
	for _, t := range tokens {
		t = strings.Trim(t, ".,;:!?\"'()[]{}/<>=+\\")
		if len(t) > 2 && !stopwords[t] {
			counts[t]++
		}
	}
	return counts
}

// ExtractKeywords scores all terms in content against the corpus using TF-IDF,
// returns the top limit results sorted by Score descending.
// corpus is a slice of token sets (one per skill); used to compute IDF.
// If corpus is empty, IDF defaults to 1.0 for all terms.
func ExtractKeywords(content string, corpus []map[string]bool, limit int) []KeywordScore {
	if limit <= 0 {
		return []KeywordScore{}
	}
	counts := TermFrequency(content)
	if len(counts) == 0 {
		return []KeywordScore{}
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	N := len(corpus)
	useCorpus := N > 0

	scores := make([]KeywordScore, 0, len(counts))
	for term, count := range counts {
		tf := float64(count) / float64(total)
		var idf float64
		if !useCorpus {
			idf = 1.0
		} else {
			df := 0
			for _, doc := range corpus {
				if doc[term] {
					df++
				}
			}
			idf = math.Log(float64(N) / float64(1+df))
		}
		scores = append(scores, KeywordScore{
			Term:  term,
			TF:    tf,
			IDF:   idf,
			Score: tf * idf,
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score != scores[j].Score {
			return scores[i].Score > scores[j].Score
		}
		return scores[i].Term < scores[j].Term
	})

	if limit < len(scores) {
		return scores[:limit]
	}
	return scores
}
