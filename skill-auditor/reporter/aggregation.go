package reporter

import (
	"fmt"
	"math"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/skill-auditor/duplication"
)

// AggregationPlan returns a markdown aggregation plan for a skill family.
func AggregationPlan(family string, entries []duplication.SkillEntry, pairs []duplication.Pair, date string) string {
	var sb strings.Builder

	totalLines := 0
	for _, e := range entries {
		totalLines += strings.Count(e.Content, "\n") + 1
	}

	avgSim := 0.0
	if len(pairs) > 0 {
		sum := 0.0
		for _, p := range pairs {
			sum += p.Similarity
		}
		avgSim = sum / float64(len(pairs))
	}

	fmt.Fprintf(&sb, "# Aggregation Plan — %s — %s\n\n", family, date)

	fmt.Fprintf(&sb, "## Family Analysis\n\n")
	fmt.Fprintf(&sb, "| Metric | Value |\n")
	fmt.Fprintf(&sb, "|--------|-------|\n")
	fmt.Fprintf(&sb, "| Skills | %d |\n", len(entries))
	fmt.Fprintf(&sb, "| Total lines | %d |\n", totalLines)
	fmt.Fprintf(&sb, "| Avg duplication | %.0f%% |\n", avgSim*100)
	fmt.Fprintf(&sb, "| Duplicate pairs | %d |\n\n", len(pairs))

	fmt.Fprintf(&sb, "### Skills in Family\n\n")
	for _, e := range entries {
		lines := strings.Count(e.Content, "\n") + 1
		flag := ""
		if lines > 500 {
			flag = " ⚠️ oversized"
		}
		fmt.Fprintf(&sb, "- `%s` (%d lines)%s\n", e.Key, lines, flag)
	}
	fmt.Fprintf(&sb, "\n")

	fmt.Fprintf(&sb, "## Decision\n\n")
	decision, rationale := aggregationDecision(len(entries), totalLines, avgSim)
	fmt.Fprintf(&sb, "**%s** — %s\n\n", decision, rationale)

	if !strings.HasPrefix(decision, "AGGREGATE") {
		return sb.String()
	}

	fmt.Fprintf(&sb, "## 6-Step Aggregation Process\n\n")

	hubName := family
	fmt.Fprintf(&sb, "### Step 1: Identify Aggregation Candidates\n\n")
	fmt.Fprintf(&sb, "All %d skills in the `%s` family are candidates.\n\n", len(entries), family)

	fmt.Fprintf(&sb, "### Step 2: Design Category Structure\n\n")
	fmt.Fprintf(&sb, "Create `skills/%s/` as the navigation hub directory.\n\n", hubName)

	fmt.Fprintf(&sb, "### Step 3: Create Navigation Hub (SKILL.md)\n\n")
	fmt.Fprintf(&sb, "```\nskills/%s/SKILL.md        ← 60–100 lines, overview only\nskills/%s/references/     ← deep-dive content from originals\nskills/%s/AGENTS.md       ← lists all references with descriptions\n```\n\n", hubName, hubName, hubName)

	fmt.Fprintf(&sb, "### Step 4: Migrate Content to References\n\n")
	for _, e := range entries {
		refName := strings.ReplaceAll(duplication.ShortKey(e.Key), "-", "-") + ".md"
		fmt.Fprintf(&sb, "- Extract content from `%s` → `skills/%s/references/%s`\n", e.Key, hubName, refName)
	}
	fmt.Fprintf(&sb, "\n")

	fmt.Fprintf(&sb, "### Step 5: Update Cross-References\n\n")
	fmt.Fprintf(&sb, "Update any skills that reference the originals to point to the new hub.\n\n")

	fmt.Fprintf(&sb, "### Step 6: Deprecate Original Skills\n\n")
	fmt.Fprintf(&sb, "Move originals to `.deprecated/` with a README explaining the consolidation.\n\n")

	fmt.Fprintf(&sb, "## Effort Estimate\n\n")
	effort := aggregationEffort(len(entries), totalLines)
	fmt.Fprintf(&sb, "| Phase | Effort | Time |\n")
	fmt.Fprintf(&sb, "|-------|--------|------|\n")
	fmt.Fprintf(&sb, "| Steps 1–3: Design hub | S | 1 hour |\n")
	fmt.Fprintf(&sb, "| Step 4: Migrate content | %s | %s |\n", effort, aggregationTime(len(entries)))
	fmt.Fprintf(&sb, "| Steps 5–6: Cleanup | S | 30 min |\n")
	fmt.Fprintf(&sb, "| **Total** | **%s** | **%s** |\n\n", effort, aggregationTotalTime(len(entries)))

	fmt.Fprintf(&sb, "## Verification Checklist\n\n")
	fmt.Fprintf(&sb, "- [ ] Navigation hub (SKILL.md) is 60–100 lines\n")
	fmt.Fprintf(&sb, "- [ ] AGENTS.md lists all references with descriptions\n")
	fmt.Fprintf(&sb, "- [ ] Each original skill content migrated to references/\n")
	fmt.Fprintf(&sb, "- [ ] Original skills moved to .deprecated/\n")
	fmt.Fprintf(&sb, "- [ ] No broken @see references\n")

	return sb.String()
}

func aggregationDecision(count, lines int, avgSim float64) (string, string) {
	reasons := buildAggregationReasons(count, lines, avgSim)
	strongSignal := avgSim >= duplication.ThresholdCritical || (avgSim >= duplication.ThresholdHigh && count >= 3)
	if len(reasons) >= 2 || (len(reasons) == 1 && strongSignal) {
		return "AGGREGATE ✅", strings.Join(reasons, "; ")
	}
	if len(reasons) == 1 {
		return "CONSIDER", reasons[0]
	}
	return "MONITOR", "No strong duplication or size signals detected"
}

func buildAggregationReasons(count, lines int, avgSim float64) []string {
	var reasons []string
	if avgSim >= duplication.ThresholdCritical {
		reasons = append(reasons, fmt.Sprintf("%.0f%% avg duplication exceeds critical threshold", avgSim*100))
	} else if avgSim >= duplication.ThresholdHigh && count >= 3 {
		reasons = append(reasons, fmt.Sprintf("%.0f%% avg duplication across %d skills causes confusion", avgSim*100, count))
	}
	if count >= 3 && len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf("%d skills in family cause user confusion", count))
	}
	if lines > 2000 {
		reasons = append(reasons, fmt.Sprintf("%d combined lines exceeds 2000-line threshold", lines))
	}
	return reasons
}

func aggregationEffort(count, lines int) string {
	if count <= 3 && lines <= 1000 {
		return "S"
	}
	if count <= 6 && lines <= 3000 {
		return "M"
	}
	return "L"
}

func aggregationTime(count int) string {
	hours := int(math.Ceil(float64(count) * 0.5))
	if hours <= 1 {
		return "1 hour"
	}
	return fmt.Sprintf("%d hours", hours)
}

func aggregationTotalTime(count int) string {
	hours := int(math.Ceil(float64(count)*0.5)) + 2
	return fmt.Sprintf("%d hours", hours)
}
