package scorer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// scoreD9 — Eval Validation (max: 20)
func scoreD9(evalsDir string) (int, []Diagnostic) {
	score := 0
	var diags []Diagnostic

	if info, err := os.Stat(evalsDir); err != nil || !info.IsDir() {
		diags = append(diags, warnDiag("D9", "evals/ directory missing entirely"))
		return score, diags
	}
	score += 4

	delta, instrDiags := scoreD9Instructions(evalsDir)
	score += delta
	diags = append(diags, instrDiags...)

	delta, summaryDiags := scoreD9Summary(evalsDir)
	score += delta
	diags = append(diags, summaryDiags...)

	validScenarios, scenarioDiags := countValidScenariosWithDiags(evalsDir)
	diags = append(diags, scenarioDiags...)
	if validScenarios >= 3 {
		score += 4
	} else if validScenarios >= 1 {
		score += 2
	}

	if score > 20 {
		score = 20
	}
	return score, diags
}

func scoreD9Instructions(evalsDir string) (int, []Diagnostic) {
	data, err := os.ReadFile(filepath.Join(evalsDir, "instructions.json"))
	if err != nil {
		return 0, nil
	}
	var instrData struct {
		Instructions []json.RawMessage `json:"instructions"`
	}
	if json.Unmarshal(data, &instrData) != nil {
		return 0, []Diagnostic{errDiag("D9", "instructions.json exists but is not valid JSON")}
	}
	if len(instrData.Instructions) > 0 || len(data) > 0 {
		return 3, nil
	}
	return 0, nil
}

func scoreD9Summary(evalsDir string) (int, []Diagnostic) {
	data, err := os.ReadFile(filepath.Join(evalsDir, "summary.json"))
	if err != nil {
		return 0, nil
	}
	var summaryData struct {
		InstructionsCoverage struct {
			CoveragePercentage any `json:"coverage_percentage"`
		} `json:"instructions_coverage"`
	}
	if json.Unmarshal(data, &summaryData) != nil {
		return 0, []Diagnostic{errDiag("D9", "summary.json exists but is not valid JSON")}
	}
	coverage := parseCoveragePercentage(summaryData.InstructionsCoverage.CoveragePercentage)
	if coverage >= 0 {
		if coverage >= 80 {
			return 6, nil
		}
		return 3, []Diagnostic{warnDiag("D9", fmt.Sprintf("summary.json coverage is %d%% (below 80%% threshold)", coverage))}
	}
	if len(data) > 0 {
		return 3, nil
	}
	return 0, nil
}

// parseCoveragePercentage parses a coverage percentage value to int. Returns -1 if unparseable.
func parseCoveragePercentage(v any) int {
	if v == nil {
		return -1
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		s := strings.TrimRight(strings.TrimSpace(val), "%")
		if s == "" {
			return -1
		}
		if dotIdx := strings.Index(s, "."); dotIdx >= 0 {
			s = s[:dotIdx]
		}
		n := 0
		for _, ch := range s {
			if ch < '0' || ch > '9' {
				return -1
			}
			n = n*10 + int(ch-'0')
		}
		return n
	}
	return -1
}

// countValidScenarios is a thin wrapper used by tests.
func countValidScenarios(evalsDir string) int {
	count, _ := countValidScenariosWithDiags(evalsDir)
	return count
}

// countValidScenariosWithDiags counts valid scenario dirs and emits diagnostics for problems.
func countValidScenariosWithDiags(evalsDir string) (int, []Diagnostic) {
	var diags []Diagnostic
	entries, err := os.ReadDir(evalsDir)
	if err != nil {
		return 0, diags
	}

	flatCount := countFlatScenarios(entries)

	valid := 0
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "scenario-") {
			continue
		}
		ok, scenDiags := validateScenarioDir(evalsDir, e.Name())
		diags = append(diags, scenDiags...)
		if ok {
			valid++
		}
	}

	if valid == 0 && flatCount > 0 {
		diags = append(diags, warnDiag("D9", fmt.Sprintf("%d flat scenario-NN.md file(s) found; migrate to scenario-N/ subdirectory format to score on D9", flatCount)))
	}

	return valid, diags
}

func countFlatScenarios(entries []os.DirEntry) int {
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "scenario-") && strings.HasSuffix(e.Name(), ".md") {
			count++
		}
	}
	return count
}

func validateScenarioDir(evalsDir, name string) (bool, []Diagnostic) {
	scenarioDir := filepath.Join(evalsDir, name)
	hasTask := fileExists(filepath.Join(scenarioDir, "task.md"))
	hasCriteria := fileExists(filepath.Join(scenarioDir, "criteria.json"))
	hasCapability := fileExists(filepath.Join(scenarioDir, "capability.txt"))

	if !hasTask || !hasCriteria || !hasCapability {
		var missing []string
		if !hasTask {
			missing = append(missing, "task.md")
		}
		if !hasCriteria {
			missing = append(missing, "criteria.json")
		}
		if !hasCapability {
			missing = append(missing, "capability.txt")
		}
		return false, []Diagnostic{warnDiag("D9", fmt.Sprintf("%s missing: %s", name, strings.Join(missing, ", ")))}
	}

	return validateScenarioCriteria(scenarioDir, name)
}

func validateScenarioCriteria(scenarioDir, name string) (bool, []Diagnostic) {
	data, err := os.ReadFile(filepath.Join(scenarioDir, "criteria.json"))
	if err != nil {
		return true, nil
	}
	var criteriaData struct {
		Checklist []struct {
			MaxScore int `json:"max_score"`
		} `json:"checklist"`
	}
	if json.Unmarshal(data, &criteriaData) != nil {
		return true, []Diagnostic{errDiag("D9", fmt.Sprintf("%s/criteria.json is not valid JSON", name))}
	}
	sum := 0
	for _, item := range criteriaData.Checklist {
		sum += item.MaxScore
	}
	if sum == 100 {
		return true, nil
	}
	return false, []Diagnostic{warnDiag("D9", fmt.Sprintf("%s/criteria.json checklist does not sum to 100 (got %d)", name, sum))}
}
