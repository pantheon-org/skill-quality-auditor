package scorer

import (
	"strings"
)

// scoreD8 — Practical Usability (max: 15)
// Code block count and language tags via library; run-command check kept custom;
// outcome linkage checks whether examples specify a verifiable result.
func scoreD8(content string, b *validatorBridge) (int, []Diagnostic) {
	score := d8BaseScore
	score += scoreD8CodeBlocks(content, b)
	if hasRunCommand(content) {
		score += 4
	}
	score += scoreOutcomeLinkage(content)
	if score > d8Max {
		score = d8Max
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
		case b.Content.CodeBlockCount > d8BlocksHigh:
			delta = 4
		case b.Content.CodeBlockCount > d8BlocksMid:
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
	case blocks > d8BlocksHigh:
		return 4
	case blocks > d8BlocksMid:
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

// scoreOutcomeLinkage awards up to 3 pts for examples that specify a verifiable outcome.
// It segments content on fenced code blocks and checks each segment (code block +
// immediately following prose paragraph) for outcome indicator phrases.
func scoreOutcomeLinkage(content string) int {
	var segments []string

	lines := strings.Split(content, "\n")
	inBlock := false
	var segBuf strings.Builder
	collectingPost := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if !inBlock {
				inBlock = true
				segBuf.Reset()
				collectingPost = false
			} else {
				inBlock = false
				collectingPost = true
			}
			continue
		}
		if inBlock {
			segBuf.WriteString(line)
			segBuf.WriteString("\n")
		} else if collectingPost {
			if trimmed == "" || (strings.HasPrefix(trimmed, "#") && trimmed != "#" && !strings.HasPrefix(trimmed, "# ")) {
				segments = append(segments, segBuf.String())
				segBuf.Reset()
				collectingPost = false
			} else {
				segBuf.WriteString(line)
				segBuf.WriteString("\n")
			}
		}
	}
	if collectingPost {
		segments = append(segments, segBuf.String())
	}

	if len(segments) == 0 {
		return 0
	}

	outcomeIndicators := []string{
		"# output:", "# result:", "→", "produces", "creates", "writes to",
		"# verify:", "# check:", "should return", "expected:", "assert",
		"you should see", "the PR will", "the job will", "confirms that",
	}

	linked := 0
	for _, seg := range segments {
		lower := strings.ToLower(seg)
		for _, ind := range outcomeIndicators {
			if strings.Contains(lower, ind) {
				linked++
				break
			}
		}
	}

	switch {
	case linked == len(segments):
		return 3
	case linked*2 >= len(segments):
		return 2
	case linked > 0:
		return 1
	default:
		return 0
	}
}
