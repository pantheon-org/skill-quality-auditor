package cmd

import "github.com/pantheon-org/skill-quality-auditor/agents"

// Agent and registry helpers re-exported from the shared agents package
// for use within the cmd package.
type Agent = agents.Agent

var agentRegistry = agents.Registry

func agentByID(id string) (Agent, bool) { return agents.ByID(id) }
