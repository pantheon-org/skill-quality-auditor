package scorer

import (
	"strings"

	"github.com/pantheon-org/skill-quality-auditor/agents"
)

// harnessDirs and agentNames are derived from the shared agents registry.
// extraAgentNames covers tools that are real agents but not installable via
// the init command (e.g. Aider, which has no skills-spec install path).
var extraAgentNames = []string{"aider"}

var harnessDirs = agents.HarnessDirs()
var agentNames = append(agents.DisplayNames(), extraAgentNames...)

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
// or an empty string if none is present.
func findAgentRef(content string) string {
	lower := strings.ToLower(content)
	for _, name := range agentNames {
		if strings.Contains(lower, name) {
			return name
		}
	}
	return ""
}
