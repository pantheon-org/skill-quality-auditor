package scorer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// imperativeRe matches lines containing MUST, NEVER, or ALWAYS (case-insensitive, whole word).
var imperativeRe = regexp.MustCompile(`(?i)\b(MUST|NEVER|ALWAYS)\b`)

// adversarialRe matches adversarial/failure-mode keywords in task.md files.
var adversarialRe = regexp.MustCompile(`(?i)\b(fail(ure|ed|s)?|error|invalid|edge case|conflict|unexpected)\b`)

// scoreD9 — Eval Validation (max: 20)
// skillPath is the path to SKILL.md; evalsDir is <skilldir>/evals/.
func scoreD9(evalsDir, skillPath string) (int, []Diagnostic) {
	score := 0
	var diags []Diagnostic

	if info, err := os.Stat(evalsDir); err != nil || !info.IsDir() {
		diags = append(diags, warnDiag("D9", "evals/ directory missing entirely"))
		return score, diags
	}
	score += d9EvalsDirPoints

	delta, instrDiags := scoreD9Instructions(evalsDir)
	score += delta
	diags = append(diags, instrDiags...)

	delta, summaryDiags := scoreD9Summary(evalsDir)
	score += delta
	diags = append(diags, summaryDiags...)

	validScenarios, scenarioDiags := countValidScenariosWithDiags(evalsDir)
	diags = append(diags, scenarioDiags...)
	if validScenarios >= d9ScenariosHigh {
		score += 2
	} else if validScenarios >= d9ScenariosMid {
		score += 1
	}

	// Mutation coverage: up to 5 pts added to score.
	delta, mutDiags := scoreMutationCoverage(skillPath, evalsDir)
	score += delta
	diags = append(diags, mutDiags...)

	// Adversarial scenario: diagnostic-only bonus (0 pts added to score).
	_, advDiags := scoreAdversarialScenario(evalsDir)
	diags = append(diags, advDiags...)

	// Independent authoring: diagnostic-only bonus (0 pts added to score).
	_, authDiags := scoreIndependentAuthoring(evalsDir, skillPath)
	diags = append(diags, authDiags...)

	if score > d9Max {
		score = d9Max
	}
	return score, diags
}

// scoreMutationCoverage checks whether SKILL.md's imperative constraints (MUST/NEVER/ALWAYS)
// are covered by criteria.json item descriptions across all scenario dirs.
// Returns 0–5 pts based on coverage percentage.
// If SKILL.md does not exist, returns 0 pts with no error diagnostic.
func scoreMutationCoverage(skillPath, evalsDir string) (int, []Diagnostic) {
	f, err := os.Open(skillPath)
	if err != nil {
		return 0, nil
	}
	defer func() { _ = f.Close() }()

	var instructions []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if imperativeRe.MatchString(line) {
			instructions = append(instructions, strings.ToLower(line))
		}
	}
	if len(instructions) == 0 {
		return 0, nil
	}

	criteriaDescs := gatherCriteriaDescriptions(evalsDir)

	covered := 0
	for _, instr := range instructions {
		instrTokens := tokenise(instr)
		for _, desc := range criteriaDescs {
			if hasOverlap(instrTokens, tokenise(strings.ToLower(desc))) {
				covered++
				break
			}
		}
	}

	pct := (covered * 100) / len(instructions)
	switch {
	case pct >= 80:
		return 5, nil
	case pct >= 50:
		return 3, nil
	case covered >= 1:
		return 1, nil
	default:
		return 0, nil
	}
}

// gatherCriteriaDescriptions collects all checklist item descriptions from every scenario-N/criteria.json.
func gatherCriteriaDescriptions(evalsDir string) []string {
	entries, err := os.ReadDir(evalsDir)
	if err != nil {
		return nil
	}
	var descs []string
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "scenario-") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(evalsDir, e.Name(), "criteria.json"))
		if err != nil {
			continue
		}
		var cd struct {
			Checklist []struct {
				Description string `json:"description"`
				Criterion   string `json:"criterion"`
			} `json:"checklist"`
		}
		if json.Unmarshal(data, &cd) != nil {
			continue
		}
		for _, item := range cd.Checklist {
			text := item.Description
			if text == "" {
				text = item.Criterion
			}
			if text != "" {
				descs = append(descs, text)
			}
		}
	}
	return descs
}

// tokenise splits a lowercase string into content words (≥3 chars, no stop words).
func tokenise(s string) map[string]struct{} {
	stop := map[string]struct{}{
		"a": {}, "an": {}, "the": {}, "is": {}, "are": {}, "be": {}, "to": {},
		"of": {}, "in": {}, "on": {}, "at": {}, "by": {}, "for": {}, "with": {},
		"and": {}, "or": {}, "not": {}, "all": {}, "it": {}, "its": {}, "from": {},
	}
	tokens := make(map[string]struct{})
	for _, word := range strings.FieldsFunc(s, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}) {
		if len(word) >= 3 {
			if _, isStop := stop[word]; !isStop {
				tokens[word] = struct{}{}
			}
		}
	}
	return tokens
}

// hasOverlap returns true if the two token sets share at least one token.
func hasOverlap(a, b map[string]struct{}) bool {
	for tok := range a {
		if _, ok := b[tok]; ok {
			return true
		}
	}
	return false
}

// scoreAdversarialScenario checks whether any task.md contains adversarial/failure markers.
// Always returns 0 pts; emits a hint diagnostic with the bonus amount if adversarial content found.
func scoreAdversarialScenario(evalsDir string) (int, []Diagnostic) {
	entries, err := os.ReadDir(evalsDir)
	if err != nil {
		return 0, nil
	}

	bonusAmt := 0
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "scenario-") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(evalsDir, e.Name(), "task.md"))
		if err != nil {
			continue
		}
		content := strings.ToLower(string(data))
		matches := adversarialRe.FindAllString(content, -1)
		if len(matches) >= 2 {
			bonusAmt = 3
			break
		} else if len(matches) == 1 && bonusAmt < 1 {
			bonusAmt = 1
		}
	}

	if bonusAmt > 0 {
		return 0, []Diagnostic{hintDiag("D9", fmt.Sprintf("adversarial bonus: +%d pts — not applied to score", bonusAmt))}
	}
	return 0, nil
}

// scoreIndependentAuthoring checks whether evals/ and SKILL.md were authored independently.
// Always returns 0 pts (diagnostic-only); emits a hint diagnostic with the bonus amount.
//
// Fallback tiers:
//  1. Non-git repo or git command failure → 0 pts, no error diagnostic.
//  2. CI shallow clone / detached HEAD (≤1 commit line) → fall back to os.Stat mtime.
//  3. Full git history → compare author emails and timestamps.
func scoreIndependentAuthoring(evalsDir, skillPath string) (int, []Diagnostic) {
	bonus := computeAuthoringBonus(evalsDir, skillPath)
	if bonus > 0 {
		return 0, []Diagnostic{hintDiag("D9", fmt.Sprintf("independent authoring bonus: +%d pts — not applied to score", bonus))}
	}
	return 0, nil
}

// computeAuthoringBonus returns the bonus amount (0–2) using the 3-tier fallback strategy.
func computeAuthoringBonus(evalsDir, skillPath string) int {
	evalsLines, evalsErr := gitLogLines(evalsDir)
	skillLines, skillErr := gitLogLines(skillPath)

	if evalsErr != nil || skillErr != nil {
		return mtimeBonus(evalsDir, skillPath)
	}

	if len(evalsLines) <= 1 || len(skillLines) <= 1 {
		return mtimeBonus(evalsDir, skillPath)
	}

	evalsEmail, evalsTime := parseGitLogLine(evalsLines[0])
	skillEmail, skillTime := parseGitLogLine(skillLines[0])

	if evalsEmail != "" && skillEmail != "" && evalsEmail != skillEmail {
		return 2
	}
	if !evalsTime.IsZero() && !skillTime.IsZero() {
		diff := evalsTime.Sub(skillTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Hour {
			return 2
		}
		return 1
	}
	return 0
}

// gitLogLines runs git log on a path and returns the non-empty output lines.
func gitLogLines(path string) ([]string, error) {
	cmd := exec.Command("git", "log", "--follow", "--format=%ae %at", "--", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// parseGitLogLine parses "email unixTimestamp" from a git log --format="%ae %at" line.
func parseGitLogLine(line string) (string, time.Time) {
	parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
	if len(parts) != 2 {
		return "", time.Time{}
	}
	email := parts[0]
	var ts int64
	for _, ch := range parts[1] {
		if ch < '0' || ch > '9' {
			return email, time.Time{}
		}
		ts = ts*10 + int64(ch-'0')
	}
	return email, time.Unix(ts, 0)
}

// mtimeBonus returns a bonus based on file system mtime comparison.
// Returns 2 if mtimes differ by > 1 hour, 1 if within 1 hour, 0 on stat error.
func mtimeBonus(evalsDir, skillPath string) int {
	evalsInfo, err1 := os.Stat(evalsDir)
	skillInfo, err2 := os.Stat(skillPath)
	if err1 != nil || err2 != nil {
		return 0
	}
	diff := evalsInfo.ModTime().Sub(skillInfo.ModTime())
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Hour {
		return 2
	}
	return 1
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
		if coverage >= d9CoverageMin {
			return 5, nil
		}
		return 3, []Diagnostic{warnDiag("D9", fmt.Sprintf("summary.json coverage is %d%% (below %d%% threshold)", coverage, d9CoverageMin))}
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
