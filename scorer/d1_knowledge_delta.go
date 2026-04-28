package scorer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	d1BaseScore     = 15 // starting score before adjustments
	d1PenaltyPerPat = 2  // deducted per beginner-content pattern hit
	d1BonusPerPat   = 1  // added per expert-signal pattern hit
	d1Max           = 20
)

// scoreD1 — Knowledge Delta (max: d1Max)
func scoreD1(content, skillDir string) (int, []Diagnostic) {
	score := d1BaseScore
	var diags []Diagnostic

	for _, pat := range []string{"npm install", "yarn add", "pip install", "getting started", "introduction", "basic syntax", "hello world"} {
		if countPattern(content, pat) > 0 {
			score -= d1PenaltyPerPat
		}
	}

	for _, pat := range []string{"anti-pattern", "NEVER", "ALWAYS", "production", "gotcha", "pitfall"} {
		if countPattern(content, pat) > 0 {
			score += d1BonusPerPat
		}
	}

	instrFile := filepath.Join(skillDir, "evals", "instructions.json")
	delta, instrDiags := scoreD1FromInstructions(instrFile)
	score += delta
	diags = append(diags, instrDiags...)

	if score < 0 {
		score = 0
	}
	if score > d1Max {
		score = d1Max
	}
	return score, diags
}

func scoreD1FromInstructions(instrFile string) (int, []Diagnostic) {
	data, err := os.ReadFile(instrFile)
	if err != nil {
		return 0, nil
	}
	var instrData struct {
		Instructions []struct {
			WhyGiven string `json:"why_given"`
		} `json:"instructions"`
	}
	if json.Unmarshal(data, &instrData) != nil {
		return 0, []Diagnostic{errDiag("D1", "instructions.json exists but cannot be parsed")}
	}
	total := len(instrData.Instructions)
	if total == 0 {
		return 0, nil
	}
	newKnow, pref := 0, 0
	for _, instr := range instrData.Instructions {
		switch instr.WhyGiven {
		case "new knowledge":
			newKnow++
		case "preference":
			pref++
		}
	}
	expertRatio := (newKnow + pref) * 100 / total
	if expertRatio >= 70 {
		return 2, nil
	}
	if expertRatio < 30 {
		return -2, nil
	}
	return 0, nil
}
