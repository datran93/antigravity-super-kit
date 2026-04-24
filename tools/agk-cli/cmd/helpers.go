package cmd

import (
	"os"
)

// flagAI is shared across install, update, remove commands.
var flagAI []string

// resolveAgents loads agents.json and resolves the --ai flags to agent keys.
// Returns the list of agent keys and the loaded agents file.
func resolveAgents() ([]string, *AgentsFile, error) {
	af, err := loadAgentsFromWorkspace()
	if err != nil {
		return nil, nil, err
	}

	agents := flagAI
	if len(agents) == 0 {
		agents = []string{DefaultAgentKey}
	}

	// Validate all agent keys
	for _, key := range agents {
		if _, err := af.GetAgent(key); err != nil {
			return nil, nil, err
		}
	}

	return agents, af, nil
}

// loadAgentsFromWorkspace finds and loads agents.json from the current workspace.
func loadAgentsFromWorkspace() (*AgentsFile, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	jsonPath, err := FindAgentsJSON(cwd)
	if err != nil {
		return nil, err
	}

	return LoadAgentsFile(jsonPath)
}
