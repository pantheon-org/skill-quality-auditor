package duplication

const (
	ThresholdCritical = 0.35
	ThresholdHigh     = 0.20
)

// Pair represents a pair of skills with a computed similarity score.
type Pair struct {
	A          string
	B          string
	Similarity float64 // [0,1]
	Severity   string  // "Critical", "High", or ""
}

// Detect performs an O(n²) pairwise similarity check across all entries and
// returns pairs above ThresholdHigh, sorted by similarity descending.
func Detect(entries []SkillEntry) []Pair {
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
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].Similarity > pairs[i].Similarity {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	return pairs
}
