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
)

// scoreD3 — Anti-Pattern Quality (max: 15)
func scoreD3(content, skillDir string, b *validatorBridge) (int, []Diagnostic) {
	score := 0
	var diags []Diagnostic

	score += scoreD3DirectiveLanguage(content, b)

	if reD3BadGood.MatchString(content) {
		score += 2
	}
	if countPattern(content, "WHY:") > 0 {
		score += 2
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
