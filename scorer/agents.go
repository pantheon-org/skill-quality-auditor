package scorer

import (
	"regexp"
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/agents"
)

// harnessDirs and agentNames are derived from the shared agents registry.
// extraAgentNames covers tools that are real agents but not installable via
// the init command (e.g. Aider, which has no skills-spec install path).
var extraAgentNames = []string{"aider"}

var harnessDirs = agents.HarnessDirs()
var agentNames = append(agents.DisplayNames(), extraAgentNames...)

// agentNameRes are precompiled word-boundary regexes for each agent name.
// Word-boundary matching prevents short names (e.g. "amp") from matching
// as substrings of unrelated words like "example" or "sample".
var agentNameRes = func() []*regexp.Regexp {
	res := make([]*regexp.Regexp, len(agentNames))
	for i, name := range agentNames {
		res[i] = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(name) + `\b`)
	}
	return res
}()

// findHarnessPath returns the first harness directory reference found in content,
// or an empty string if none is present.
func findHarnessPath(content string) string {
	for _, dir := range harnessDirs {
		if strings.Contains(content, dir+"/") {
			return dir + "/"
		}
	}
	return ""
}

// findAgentRef returns the first agent name reference found in content (case-insensitive),
// or an empty string if none is present. Uses word-boundary matching so short agent
// names (e.g. "amp") do not match inside longer words (e.g. "example").
func findAgentRef(content string) string {
	for i, re := range agentNameRes {
		if re.MatchString(content) {
			return agentNames[i]
		}
	}
	return ""
}
