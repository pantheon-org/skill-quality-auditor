package analysis

import (
	"math"
	"sort"

	"github.com/pantheon-org/skill-quality-auditor/internal/tokenize"
)

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
	return tokenize.Counts(text)
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
			idf = math.Max(0, math.Log(float64(N)/float64(1+df)))
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
