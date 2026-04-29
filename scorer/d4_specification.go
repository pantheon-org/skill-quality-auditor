package scorer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// scoreD4 — Specification Compliance (max: 17)
func scoreD4(content, skillDir string, b *validatorBridge) (int, []Diagnostic) {
	score := 6
	var diags []Diagnostic

	score += scoreD4Description(content, b)
	score += scoreD4HarnessRefs(content, &diags)
	score += scoreD4RelPaths(content)
	score += scoreSpecificationMutationResistance(content)

	absPathRe := regexp.MustCompile(`skills/[a-z][a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+`)
	ctxAgentsRe := regexp.MustCompile(`\.(context|agents)/`)
	delta, newDiags := scoreD4ContentViolations(content, b, absPathRe, ctxAgentsRe)
	score += delta
	diags = append(diags, newDiags...)

	score -= penaltyFromDir(filepath.Join(skillDir, "scripts"), absPathRe)
	score -= penaltyFromDir(filepath.Join(skillDir, "scripts"), ctxAgentsRe)
	score -= penaltyFromDir(filepath.Join(skillDir, "references"), absPathRe)
	score -= penaltyFromDir(filepath.Join(skillDir, "references"), ctxAgentsRe)

	if score > d4Max {
		score = d4Max
	}
	if score < 0 {
		score = 0
	}

	score += scoreD4Bonus(content, skillDir)
	if score > d4MaxWithBonus {
		score = d4MaxWithBonus
	}
	return score, diags
}

func scoreD4Description(content string, b *validatorBridge) int {
	delta := 0
	descLen := b.descriptionLen()
	if descLen < 0 {
		descLen = len(extractFrontmatterField(content, "description"))
	}
	if descLen > d4DescLenMid {
		delta += 2
	}
	description := extractFrontmatterField(content, "description")
	andOrRe := regexp.MustCompile(`(?i) and | or `)
	andOrCount := len(andOrRe.FindAllString(description, -1))
	if andOrCount > d4AndOrCountHigh {
		delta -= 2
	} else if andOrCount > d4AndOrCountMid {
		delta--
	}
	return delta
}

func scoreD4HarnessRefs(content string, diags *[]Diagnostic) int {
	delta := 0
	if dir := findHarnessPath(content); dir != "" {
		*diags = append(*diags, warnDiag("D4", "harness-specific path found: "+dir))
	} else {
		delta++
	}
	if ref := findAgentRef(content); ref != "" {
		*diags = append(*diags, warnDiag("D4", "agent-specific reference found: "+ref))
	} else {
		delta++
	}
	return delta
}

func scoreD4RelPaths(content string) int {
	relPathRe := regexp.MustCompile(`(scripts|references|assets)/[a-zA-Z0-9_-]+`)
	if relPathRe.MatchString(content) {
		return 1
	}
	return 0
}

func scoreD4ContentViolations(content string, b *validatorBridge, absPathRe, ctxAgentsRe *regexp.Regexp) (int, []Diagnostic) {
	delta := 0
	var diags []Diagnostic
	nonCode := removeCodeBlocks(content)
	if b.hasInternalLinkWarning() || strings.Contains(nonCode, "../") {
		delta -= 2
		diags = append(diags, warnDiag("D4", "../ reference outside code blocks (self-containment violation)"))
	}
	if m := absPathRe.FindString(nonCode); m != "" {
		delta--
		diags = append(diags, warnDiag("D4", "absolute skill path outside code blocks: "+m))
	}
	if m := ctxAgentsRe.FindString(nonCode); m != "" {
		delta--
		diags = append(diags, warnDiag("D4", ".context/ or .agents/ reference outside code blocks: "+m))
	}
	return delta, diags
}

func scoreD4Bonus(content, skillDir string) int {
	bonus := 0
	scriptsDir := filepath.Join(skillDir, "scripts")
	if info, err := os.Stat(scriptsDir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(scriptsDir)
		for _, e := range entries {
			name := e.Name()
			if strings.HasSuffix(name, ".py") || strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".js") {
				bonus++
				break
			}
		}
	}
	lastH2 := extractLastH2(content)
	bulletLinkRe := regexp.MustCompile(`(?m)^- \[.+\]\(.+\)`)
	if strings.TrimSpace(lastH2) == "References" && bulletLinkRe.MatchString(content) {
		bonus++
	}
	return bonus
}

func extractLastH2(content string) string {
	last := ""
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			last = strings.TrimPrefix(line, "## ")
		}
	}
	return last
}

// scoreSpecificationMutationResistance scores mutation-resistance of the specification.
// Three sub-criteria (code blocks stripped before matching):
//   - hard constraints (MUST/NEVER/always/ONLY + ≥4-word verb phrase): 1.5 pts
//   - conditional branches (if/when/unless + ≥2-word noun phrase): 1.5 pts
//   - explicit exclusions (does not/out of scope/DO NOT/SKIP in ≥6-word sentence): 1 pt
//
// Total float is truncated to int (partial credit for 1-of-3 or 2-of-3).
func scoreSpecificationMutationResistance(content string) int {
	nonCode := removeCodeBlocks(content)
	var total float64
	if hasHardConstraint(nonCode) {
		total += 1.5
	}
	if hasConditionalBranch(nonCode) {
		total += 1.5
	}
	if hasExclusion(nonCode) {
		total += 1.0
	}
	return int(total)
}

// hasHardConstraint returns true when the content outside code blocks contains
// a MUST/NEVER/always/ONLY keyword followed by a specific ≥4-word verb phrase.
// Generic phrases like "follow best practices" or "be careful" are excluded.
func hasHardConstraint(nonCode string) bool {
	re := regexp.MustCompile(`(?i)\b(MUST|NEVER|always|ONLY)\b\s+(\S+\s+\S+\s+\S+\s+\S+)`)
	matches := re.FindAllStringSubmatch(nonCode, -1)
	for _, m := range matches {
		phrase := strings.ToLower(strings.TrimSpace(m[2]))
		if strings.HasPrefix(phrase, "follow best practices") || strings.HasPrefix(phrase, "be careful") {
			continue
		}
		return true
	}
	return false
}

// hasConditionalBranch returns true when the content outside code blocks contains
// an if/when/unless/only if keyword followed by a ≥2-word noun phrase
// (no comma between the two words, preventing vague "if needed, proceed" patterns;
// first word must not be a gerund ending in -ing to exclude purpose clauses like "when writing code").
func hasConditionalBranch(nonCode string) bool {
	re := regexp.MustCompile(`(?i)\b(only\s+if|when|unless|if)\b\s+([^\s,;.]+\s+[^\s,;.]+)`)
	matches := re.FindAllStringSubmatch(nonCode, -1)
	for _, m := range matches {
		phrase := strings.ToLower(strings.TrimSpace(m[2]))
		firstWord := strings.Fields(phrase)[0]
		if strings.HasSuffix(firstWord, "ing") {
			continue
		}
		return true
	}
	return false
}

// hasExclusion returns true when the content outside code blocks contains
// a does not/out of scope/DO NOT/SKIP marker inside a ≥6-word sentence.
func hasExclusion(nonCode string) bool {
	sentenceRe := regexp.MustCompile(`[^.!?\n]+[.!?\n]?`)
	exclusionRe := regexp.MustCompile(`(?i)(does not|out of scope|DO NOT|SKIP)`)
	sentences := sentenceRe.FindAllString(nonCode, -1)
	for _, s := range sentences {
		if !exclusionRe.MatchString(s) {
			continue
		}
		if len(strings.Fields(s)) >= 6 {
			return true
		}
	}
	return false
}

// penaltyFromDir returns the number of files in dir matching re, capped at 2.
func penaltyFromDir(dir string, re *regexp.Regexp) int {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return 0
	}
	entries, _ := os.ReadDir(dir)
	penalty := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if re.MatchString(string(data)) {
			penalty++
			if penalty >= 2 {
				break
			}
		}
	}
	return penalty
}
