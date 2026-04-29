package scorer

import (
	"strings"
)

// Dimension is the canonical descriptor for one scoring rubric dimension.
type Dimension struct {
	Code  string // "D1"–"D9"
	Key   string // camelCase map key used in Result.Dimensions
	Label string // human-readable display name
	Max   int    // maximum points for this dimension
}

// AllDimensions is the single source of truth for all nine scoring dimensions,
// in canonical display order.
var AllDimensions = []Dimension{
	{"D1", "knowledgeDelta", "Knowledge Delta", 20},
	{"D2", "mindsetProcedures", "Mindset + Procedures", 15},
	{"D3", "antiPatternQuality", "Anti-Pattern Quality", 15},
	{"D4", "specificationCompliance", "Specification Compliance", 15},
	{"D5", "progressiveDisclosure", "Progressive Disclosure", 15},
	{"D6", "freedomCalibration", "Freedom Calibration", 15},
	{"D7", "patternRecognition", "Pattern Recognition", 10},
	{"D8", "practicalUsability", "Practical Usability", 15},
	{"D9", "evalValidation", "Eval Validation", 20},
}

// dimensionScores builds the Result.Dimensions map from the nine raw scores in AllDimensions order.
func dimensionScores(d1, d2, d3, d4, d5, d6, d7, d8, d9 int) map[string]int {
	scores := []int{d1, d2, d3, d4, d5, d6, d7, d8, d9}
	m := make(map[string]int, len(AllDimensions))
	for i, d := range AllDimensions {
		m[d.Key] = scores[i]
	}
	return m
}

// DimensionScore holds the score and max for a single rubric dimension.
// Used internally; not serialized directly.
type DimensionScore struct {
	Score int
	Max   int
}

// Diagnostic is a single error or warning produced during scoring.
type Diagnostic struct {
	Dimension string `json:"dimension"`
	Message   string `json:"message"`
	severity  string // "error" or "warning" — internal only, not serialized
}

// Severity returns "error" or "warning".
func (d Diagnostic) Severity() string { return d.severity }

// Result is the output of scoring a skill.
// JSON shape matches the legacy evaluate.sh audit.json format.
type Result struct {
	Skill                     string         `json:"skill"`
	Date                      string         `json:"date"`
	Dimensions                map[string]int `json:"dimensions"`
	Total                     int            `json:"total"`
	MaxTotal                  int            `json:"maxTotal"`
	Grade                     string         `json:"grade"`
	Lines                     int            `json:"lines"`
	HasReferences             bool           `json:"hasReferences"`
	ReferenceCount            int            `json:"referenceCount"`
	ReferenceSectionCompliant bool           `json:"referenceSectionCompliant"`
	Errors                    int            `json:"errors"`
	Warnings                  int            `json:"warnings"`
	ErrorDetails              []Diagnostic   `json:"errorDetails,omitempty"`
	WarningDetails            []Diagnostic   `json:"warningDetails,omitempty"`
}

func errDiag(dimension, message string) Diagnostic {
	return Diagnostic{Dimension: dimension, Message: message, severity: "error"}
}

func warnDiag(dimension, message string) Diagnostic {
	return Diagnostic{Dimension: dimension, Message: message, severity: "warning"}
}

// NewErrorDiag creates an error-severity Diagnostic for use in tests and external packages.
func NewErrorDiag(dimension, message string) Diagnostic {
	return errDiag(dimension, message)
}

// NewWarnDiag creates a warning-severity Diagnostic for use in tests and external packages.
func NewWarnDiag(dimension, message string) Diagnostic {
	return warnDiag(dimension, message)
}

// countPattern counts case-insensitive substring occurrences.
func countPattern(content, pattern string) int {
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

// countLines counts non-empty lines in content.
func countLines(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// extractFrontmatterField parses a YAML frontmatter field between --- delimiters.
func extractFrontmatterField(content, field string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	fmStarted := false
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "---" {
			if !fmStarted {
				fmStarted = true
				inFrontmatter = true
				continue
			}
			break
		}
		if inFrontmatter {
			prefix := field + ":"
			if strings.HasPrefix(trimmed, prefix) {
				val := strings.TrimPrefix(trimmed, prefix)
				val = strings.TrimSpace(val)
				val = strings.Trim(val, `"'`)
				return val
			}
		}
	}
	return ""
}

// codeBlockCount counts the number of triple-backtick fence pairs.
func codeBlockCount(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			count++
		}
	}
	return count / 2
}

// removeCodeBlocks strips fenced code blocks from content (replicates awk '/^```/{skip=!skip;next} !skip').
func removeCodeBlocks(content string) string {
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
