package reporter

import (
	"fmt"
	"math"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/duplication"
)

// DuplicationReport returns a markdown duplication report in the standard format.
func DuplicationReport(pairs []duplication.Pair, entries []duplication.SkillEntry, date string) string {
	var sb strings.Builder

	criticalCount := 0
	for _, p := range pairs {
		if p.Severity == "Critical" {
			criticalCount++
		}
	}

	fmt.Fprintf(&sb, "# Duplication Report — %s\n\n", date)
	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "- Skills analysed: %d\n", len(entries))
	fmt.Fprintf(&sb, "- Pairs with >%.0f%% similarity: %d\n", duplication.ThresholdHigh*100, len(pairs))
	fmt.Fprintf(&sb, "- Critical (>%.0f%%): %d\n\n", duplication.ThresholdCritical*100, criticalCount)

	if len(pairs) == 0 {
		fmt.Fprintf(&sb, "No duplication detected above threshold.\n")
		return sb.String()
	}

	// Group by family (common prefix before first '-' separator)
	families := groupByFamily(pairs)

	fmt.Fprintf(&sb, "## High-Priority Candidates\n\n")
	for family, fps := range families {
		fmt.Fprintf(&sb, "### %s\n\n", family)
		fmt.Fprintf(&sb, "| Skill Pair | Similarity | Severity | Action |\n")
		fmt.Fprintf(&sb, "|------------|-----------|----------|--------|\n")
		for _, p := range fps {
			sim := int(math.Round(p.Similarity * 100))
			action := "Consider"
			if p.Severity == "Critical" {
				action = "Aggregate"
			}
			fmt.Fprintf(&sb, "| %s ↔ %s | %d%% | %s | %s |\n",
				shortKey(p.A), shortKey(p.B), sim, p.Severity, action)
		}
		fmt.Fprintf(&sb, "\n")
	}

	fmt.Fprintf(&sb, "## Recommendations\n\n")
	for i, p := range pairs {
		if i >= 5 {
			break
		}
		sim := int(math.Round(p.Similarity * 100))
		verb := "Review"
		if p.Severity == "Critical" {
			verb = "Consolidate"
		}
		fmt.Fprintf(&sb, "%d. **%s**: %s ↔ %s (%d%% similarity)\n",
			i+1, verb, shortKey(p.A), shortKey(p.B), sim)
	}
	fmt.Fprintf(&sb, "\n")

	return sb.String()
}

func shortKey(key string) string {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return key
}

func groupByFamily(pairs []duplication.Pair) map[string][]duplication.Pair {
	m := make(map[string][]duplication.Pair)
	for _, p := range pairs {
		family := familyOf(p.A, p.B)
		m[family] = append(m[family], p)
	}
	return m
}

// familyOf derives a family label from two skill keys by finding their common prefix.
func familyOf(a, b string) string {
	sa := shortKey(a)
	sb := shortKey(b)
	// find longest common prefix up to a '-' boundary
	parts := strings.Split(sa, "-")
	prefix := ""
	for i := range parts {
		candidate := strings.Join(parts[:i+1], "-")
		if strings.HasPrefix(sb, candidate) {
			prefix = candidate
		}
	}
	if prefix != "" {
		return prefix + "-family"
	}
	return sa + " / " + sb
}
