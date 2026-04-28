package analysis

import (
	"regexp"
	"strings"
)

var headerLine = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

var hedgeWords = []string{"maybe", "perhaps", "might want to", "could be", "feel free", "you might", "possibly"}
var vagueWords = []string{"do something", "handle appropriately", "as needed", "when necessary", "if applicable"}
var passivePatterns = []string{"is done", "was created", "can be used", "is used", "are used", "is called", "was called"}

// RuleMatch is the result of a single pattern detection rule.
type RuleMatch struct {
	Rule     string
	Matched  bool
	Score    float64
	Evidence []string
}

func extractHeaders(content string) []string {
	matches := headerLine.FindAllStringSubmatch(content, -1)
	headers := make([]string, 0, len(matches))
	for _, m := range matches {
		headers = append(headers, strings.ToLower(strings.TrimSpace(m[2])))
	}
	return headers
}

func countSubstring(content, pattern string) int {
	lower := strings.ToLower(content)
	pat := strings.ToLower(pattern)
	count := 0
	start := 0
	for {
		idx := strings.Index(lower[start:], pat)
		if idx < 0 {
			break
		}
		count++
		start += idx + len(pat)
	}
	return count
}

func stripCodeBlocks(content string) string {
	var result strings.Builder
	skip := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			skip = !skip
			continue
		}
		if !skip {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	return result.String()
}

// DetectRequiredSections checks whether all required section headers are present.
func DetectRequiredSections(content string, required []string) []RuleMatch {
	headers := extractHeaders(content)
	headerSet := make(map[string]bool, len(headers))
	for _, h := range headers {
		headerSet[h] = true
	}

	results := make([]RuleMatch, 0, len(required))
	for _, req := range required {
		lower := strings.ToLower(req)
		matched := headerSet[lower]
		evidence := []string{}
		if matched {
			evidence = []string{lower}
		}
		score := 0.0
		if matched {
			score = 1.0
		}
		results = append(results, RuleMatch{
			Rule:     "required-section:" + lower,
			Matched:  matched,
			Score:    score,
			Evidence: evidence,
		})
	}
	return results
}

// DetectTriggerFrequency checks each trigger word against a minimum count threshold.
func DetectTriggerFrequency(content string, triggers map[string]int) []RuleMatch {
	results := make([]RuleMatch, 0, len(triggers))
	for trigger, minCount := range triggers {
		count := countSubstring(content, trigger)
		matched := count >= minCount
		score := 0.0
		if minCount > 0 {
			score = float64(count) / float64(minCount)
			if score > 1.0 {
				score = 1.0
			}
		}
		evidence := []string{}
		if matched {
			evidence = []string{trigger}
		}
		results = append(results, RuleMatch{
			Rule:     "trigger-frequency:" + strings.ToLower(trigger),
			Matched:  matched,
			Score:    score,
			Evidence: evidence,
		})
	}
	return results
}

// DetectStructuralConformance measures how closely the skill's section headers
// match a canonical section list using Jaccard similarity over header sets.
func DetectStructuralConformance(content string, canonical []string) RuleMatch {
	actual := extractHeaders(content)

	actualSet := make(map[string]bool, len(actual))
	for _, h := range actual {
		actualSet[h] = true
	}
	canonicalSet := make(map[string]bool, len(canonical))
	for _, h := range canonical {
		canonicalSet[strings.ToLower(h)] = true
	}

	intersection := 0
	for h := range actualSet {
		if canonicalSet[h] {
			intersection++
		}
	}
	union := len(actualSet) + len(canonicalSet) - intersection

	score := 0.0
	if union > 0 {
		score = float64(intersection) / float64(union)
	}

	evidence := make([]string, 0)
	for h := range actualSet {
		if canonicalSet[h] {
			evidence = append(evidence, h)
		}
	}

	return RuleMatch{
		Rule:     "structural-conformance",
		Matched:  score >= 0.5,
		Score:    score,
		Evidence: evidence,
	}
}

// DetectAntiPatternSignals looks for common anti-pattern indicators.
func DetectAntiPatternSignals(content string) []RuleMatch {
	clean := stripCodeBlocks(content)

	var results []RuleMatch

	hedgeCount := 0
	hedgeEvidence := []string{}
	for _, w := range hedgeWords {
		c := countSubstring(clean, w)
		if c > 0 {
			hedgeCount += c
			hedgeEvidence = append(hedgeEvidence, w)
		}
	}
	results = append(results, RuleMatch{
		Rule:     "anti-pattern:hedge-language",
		Matched:  hedgeCount >= 2,
		Score:    float64(hedgeCount) / 2.0,
		Evidence: hedgeEvidence,
	})

	vagueCount := 0
	vagueEvidence := []string{}
	for _, w := range vagueWords {
		c := countSubstring(clean, w)
		if c > 0 {
			vagueCount += c
			vagueEvidence = append(vagueEvidence, w)
		}
	}
	results = append(results, RuleMatch{
		Rule:     "anti-pattern:vague-instructions",
		Matched:  vagueCount >= 2,
		Score:    float64(vagueCount) / 2.0,
		Evidence: vagueEvidence,
	})

	passiveCount := 0
	passiveEvidence := []string{}
	for _, w := range passivePatterns {
		c := countSubstring(clean, w)
		if c > 0 {
			passiveCount += c
			passiveEvidence = append(passiveEvidence, w)
		}
	}
	results = append(results, RuleMatch{
		Rule:     "anti-pattern:passive-voice",
		Matched:  passiveCount >= 3,
		Score:    float64(passiveCount) / 3.0,
		Evidence: passiveEvidence,
	})

	return results
}
