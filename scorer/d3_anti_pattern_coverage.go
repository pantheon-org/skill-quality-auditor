package scorer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reD3BadGood   = regexp.MustCompile(`(?is)BAD.*GOOD`)
	reD3AntiInstr = regexp.MustCompile(`(?i)NEVER|ALWAYS|anti-pattern|avoid|do not`)
	// reNEVERMarker matches a **NEVER** bold marker that anchors an anti-pattern block.
	reNEVERMarker = regexp.MustCompile(`\*\*NEVER\*\*`)
	// reSectionHeader matches any bold header like **WORD:** or **WORD** at line start.
	reSectionHeader = regexp.MustCompile(`(?m)^\*\*[A-Z][A-Z ]*[:\*]`)
	// reSymptomHeader matches a **SYMPTOM:** or **SYMPTOM** bold header at line start.
	reSymptomHeader = regexp.MustCompile(`(?m)^\*\*SYMPTOM[:\*]`)
	// reConsequenceHeader matches a **CONSEQUENCE:** or **CONSEQUENCE** bold header at line start.
	reConsequenceHeader = regexp.MustCompile(`(?m)^\*\*CONSEQUENCE[:\*]`)
)

// antiPatternBlock holds the raw text of a single NEVER-anchored anti-pattern block.
type antiPatternBlock = string

// parseAntiPatternBlocks splits a skill document into individual anti-pattern blocks.
// Each block spans from one **NEVER** marker to the next (or end of document).
// Blocks without a substantive body (only the NEVER line) are excluded.
func parseAntiPatternBlocks(content string) []antiPatternBlock {
	indices := reNEVERMarker.FindAllStringIndex(content, -1)
	if len(indices) == 0 {
		return nil
	}
	blocks := make([]antiPatternBlock, 0, len(indices))
	for i, loc := range indices {
		start := loc[0]
		var end int
		if i+1 < len(indices) {
			end = indices[i+1][0]
		} else {
			end = len(content)
		}
		block := strings.TrimSpace(content[start:end])
		if isValidBlock(block) {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

// isValidBlock returns true when a block has a NEVER statement plus at least one
// additional non-empty line (WHY/BAD/GOOD body). Bare NEVER-only blocks do not count.
func isValidBlock(block string) bool {
	nonEmpty := 0
	for _, l := range strings.Split(block, "\n") {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	return nonEmpty >= 2
}

// scoreSymptom returns 1 if the block contains a **SYMPTOM:** bold header with at least
// one non-empty, non-header body line after it. Scoped per-block only — "symptom"
// embedded in WHY/BAD prose does not score.
func scoreSymptom(block string) int {
	return scoreBoldComponent(block, reSymptomHeader)
}

// scoreConsequence returns 1 if the block contains a **CONSEQUENCE:** bold header with at
// least one non-empty, non-header body line after it. Scoped per-block only.
func scoreConsequence(block string) int {
	return scoreBoldComponent(block, reConsequenceHeader)
}

// scoreBoldComponent is the shared implementation for scoreSymptom and scoreConsequence.
// It matches headerRe in the block then checks for substantive body before the next header.
func scoreBoldComponent(block string, headerRe *regexp.Regexp) int {
	loc := headerRe.FindStringIndex(block)
	if loc == nil {
		return 0
	}
	afterHeader := block[loc[1]:]
	// Delimit body by the next bold section header (if any).
	nextHeader := reSectionHeader.FindStringIndex(afterHeader)
	var body string
	if nextHeader != nil {
		body = afterHeader[:nextHeader[0]]
	} else {
		body = afterHeader
	}
	// Body must contain at least one non-empty, non-header line.
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "**") {
			return 1
		}
	}
	return 0
}

// scoreD3 — Anti-Pattern Quality (max: 15)
func scoreD3(content, skillDir string, b *validatorBridge) (int, []Diagnostic) {
	score := 0
	var diags []Diagnostic

	score += scoreD3DirectiveLanguage(content, b)

	// Per-block scoring: parse NEVER-anchored blocks and score each component.
	blocks := parseAntiPatternBlocks(content)
	if len(blocks) > 0 {
		score += scoreD3PerBlocks(blocks)
	} else {
		// Fallback to document-level checks for skills without structured NEVER blocks.
		if reD3BadGood.MatchString(content) {
			score += 2
		}
		if countPattern(content, "WHY:") > 0 {
			score += 2
		}
	}

	instrFile := filepath.Join(skillDir, "evals", "instructions.json")
	delta, instrDiags := scoreD3FromInstructions(instrFile)
	score += delta
	diags = append(diags, instrDiags...)

	if score > d3Max {
		score = d3Max
	}
	if score < 0 {
		score = 0
	}
	return score, diags
}

// scoreD3PerBlocks sums per-block component scores across all blocks and applies
// the count bonus once when ≥ d3BlocksMin valid blocks are present.
func scoreD3PerBlocks(blocks []antiPatternBlock) int {
	total := 0
	for _, block := range blocks {
		total += scoreD3SingleBlock(block)
	}
	if len(blocks) >= d3BlocksMin {
		total += d3PointsCountBonus
	}
	return total
}

// scoreD3SingleBlock scores one anti-pattern block across four components:
// BAD/GOOD (2 pts), WHY (2 pts), SYMPTOM (2 pts), CONSEQUENCE (2 pts).
func scoreD3SingleBlock(block string) int {
	pts := 0
	if reD3BadGood.MatchString(block) {
		pts += 2
	}
	if countPattern(block, "WHY:") > 0 || countPattern(block, "**WHY:**") > 0 {
		pts += 2
	}
	pts += scoreSymptom(block) * d3PointsSymptom
	pts += scoreConsequence(block) * d3PointsConsequence
	return pts
}

func scoreD3DirectiveLanguage(content string, b *validatorBridge) int {
	if b.Content != nil {
		sm := b.Content.StrongMarkers
		switch {
		case sm > d3StrongMarkersHigh:
			return 5
		case sm > d3StrongMarkersMid:
			return 3
		case sm > 0:
			return 1
		}
		return 0
	}
	neverCount := countPattern(content, "NEVER")
	if neverCount > d3NeverCountMin {
		return 3
	}
	return neverCount
}

func scoreD3FromInstructions(instrFile string) (int, []Diagnostic) {
	data, err := os.ReadFile(instrFile)
	if err != nil {
		return 0, nil
	}
	var instrData struct {
		Instructions []struct {
			Type             string `json:"type"`
			OriginalSnippets any    `json:"original_snippets"`
			Content          string `json:"content"`
		} `json:"instructions"`
	}
	if json.Unmarshal(data, &instrData) != nil {
		return 0, []Diagnostic{errDiag("D3", "instructions.json exists but cannot be parsed")}
	}
	antiInstr := 0
	for _, instr := range instrData.Instructions {
		snippetStr := extractSnippetStr(instr.OriginalSnippets)
		if snippetStr == "" {
			snippetStr = instr.Content
		}
		if strings.EqualFold(instr.Type, "anti-pattern") || reD3AntiInstr.MatchString(snippetStr) {
			antiInstr++
		}
	}
	if antiInstr >= d3AntiInstrHigh {
		return 2, nil
	}
	if antiInstr >= d3AntiInstrMid {
		return 1, nil
	}
	return 0, nil
}

func extractSnippetStr(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []any:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}
