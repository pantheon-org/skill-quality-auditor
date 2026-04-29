package duplication

import (
	"sort"
	"strings"
)

const (
	ThresholdCritical = 0.35
	ThresholdHigh     = 0.20

	// MaxDetectEntries caps the corpus size fed to Detect. O(n²) comparisons become
	// expensive beyond a few hundred entries; entries beyond this cap are silently dropped.
	MaxDetectEntries = 500
)

// ShortKey returns the skill name portion of a "domain/skill-name" key,
// stripping the leading domain segment if present.
func ShortKey(key string) string {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return key
}

// Pair represents a pair of skills with a computed similarity score.
type Pair struct {
	A          string
	B          string
	Similarity float64 // [0,1]
	Severity   string  // "Critical", "High", or ""
}

// Detect performs an O(n²) pairwise similarity check across all entries and
// returns pairs above ThresholdHigh, sorted by similarity descending.
// Corpus is silently truncated to MaxDetectEntries before comparison.
func Detect(entries []SkillEntry) []Pair {
	if len(entries) > MaxDetectEntries {
		entries = entries[:MaxDetectEntries]
	}
	var pairs []Pair
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			sim := Similarity(entries[i].Content, entries[j].Content)
			if sim < ThresholdHigh {
				continue
			}
			sev := "High"
			if sim >= ThresholdCritical {
				sev = "Critical"
			}
			pairs = append(pairs, Pair{
				A:          entries[i].Key,
				B:          entries[j].Key,
				Similarity: sim,
				Severity:   sev,
			})
		}
	}
	// sort descending by similarity
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Similarity > pairs[j].Similarity })
	return pairs
}
