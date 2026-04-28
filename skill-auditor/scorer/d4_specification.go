package scorer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// scoreD4 — Specification Compliance (max: 17)
func scoreD4(content, skillDir string, b *validatorBridge) (int, []Diagnostic) {
	score := 8
	var diags []Diagnostic

	score += scoreD4Description(content, b)
	score += scoreD4HarnessRefs(content, &diags)
	score += scoreD4RelPaths(content)

	absPathRe := regexp.MustCompile(`skills/[a-z][a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+`)
	ctxAgentsRe := regexp.MustCompile(`\.(context|agents)/`)
	delta, newDiags := scoreD4ContentViolations(content, b, absPathRe, ctxAgentsRe)
	score += delta
	diags = append(diags, newDiags...)

	score -= penaltyFromDir(filepath.Join(skillDir, "scripts"), absPathRe)
	score -= penaltyFromDir(filepath.Join(skillDir, "scripts"), ctxAgentsRe)
	score -= penaltyFromDir(filepath.Join(skillDir, "references"), absPathRe)
	score -= penaltyFromDir(filepath.Join(skillDir, "references"), ctxAgentsRe)

	if score > 15 {
		score = 15
	}
	if score < 0 {
		score = 0
	}

	score += scoreD4Bonus(content, skillDir)
	if score > 17 {
		score = 17
	}
	return score, diags
}

func scoreD4Description(content string, b *validatorBridge) int {
	delta := 0
	descLen := b.descriptionLen()
	if descLen < 0 {
		descLen = len(extractFrontmatterField(content, "description"))
	}
	if descLen > 100 {
		delta += 2
	}
	if descLen > 200 {
		delta++
	}
	description := extractFrontmatterField(content, "description")
	andOrRe := regexp.MustCompile(`(?i) and | or `)
	andOrCount := len(andOrRe.FindAllString(description, -1))
	if andOrCount > 3 {
		delta -= 2
	} else if andOrCount > 1 {
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
