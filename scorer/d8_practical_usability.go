package scorer

// scoreD8 — Practical Usability (max: 15)
// Code block count and language tags via library; run-command check kept custom.
func scoreD8(content string, b *validatorBridge) (int, []Diagnostic) {
	score := 5
	score += scoreD8CodeBlocks(content, b)
	if hasRunCommand(content) {
		score += 4
	}
	if score > 15 {
		score = 15
	}
	if score < 0 {
		score = 0
	}
	return score, nil
}

func scoreD8CodeBlocks(content string, b *validatorBridge) int {
	if b.Content != nil {
		delta := 0
		switch {
		case b.Content.CodeBlockCount > 5:
			delta = 4
		case b.Content.CodeBlockCount > 2:
			delta = 2
		case b.Content.CodeBlockCount > 0:
			delta = 1
		}
		if len(b.Content.CodeLanguages) > 0 {
			delta += 2
		}
		return delta
	}
	blocks := codeBlockCount(content)
	switch {
	case blocks > 5:
		return 4
	case blocks > 2:
		return 2
	}
	return 0
}

func hasRunCommand(content string) bool {
	for _, pat := range []string{"./", "npm run", "yarn ", "pnpm run", "bun run", "make ", "python ", "go run"} {
		if countPattern(content, pat) > 0 {
			return true
		}
	}
	return false
}
