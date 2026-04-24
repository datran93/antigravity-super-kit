package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AgentsFile is the top-level structure of agents.json.
type AgentsFile struct {
	SchemaVersion string                `json:"schema_version"`
	Agents        map[string]AgentEntry `json:"agents"`
}

// AgentEntry describes one supported agent.
type AgentEntry struct {
	Name             string        `json:"name"`
	TargetDir        string        `json:"target_dir"`
	Rules            *SectionSpec  `json:"rules"`
	Workflows        *SectionSpec  `json:"workflows"`
	Skills           *SectionSpec  `json:"skills"`
	References       *SectionSpec  `json:"references"`
	ArgsPlaceholder  string        `json:"args_placeholder"`
	InstructionsFile string        `json:"instructions_file"`
	Verbatim         bool          `json:"verbatim"`
}

// SectionSpec describes how a particular section (rules, workflows, etc.) is laid out.
type SectionSpec struct {
	Dir       string            `json:"dir"`
	Format    string            `json:"format"`
	Extension string            `json:"extension"`
	Rename    map[string]string `json:"rename"`
}

// DefaultAgentKey is the default agent when none is specified.
const DefaultAgentKey = "agy"

// FindAgentsJSON locates agents.json by walking up from the given directory.
// Falls back to the cached repository if not found locally.
func FindAgentsJSON(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "agents.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: check the cached repo
	cached := filepath.Join(DefaultCacheDir(), DefaultRepoName, "agents.json")
	if _, err := os.Stat(cached); err == nil {
		return cached, nil
	}

	return "", fmt.Errorf("agents.json not found (searched from %s and cache)", startDir)
}

// LoadAgentsFile reads and parses agents.json.
func LoadAgentsFile(path string) (*AgentsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var af AgentsFile
	if err := json.Unmarshal(data, &af); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &af, nil
}

// GetAgent returns the agent entry for the given key, or error if not found.
func (af *AgentsFile) GetAgent(key string) (*AgentEntry, error) {
	agent, ok := af.Agents[key]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", key)
	}
	return &agent, nil
}

// AgentKeys returns all agent keys sorted.
func (af *AgentsFile) AgentKeys() []string {
	keys := make([]string, 0, len(af.Agents))
	for k := range af.Agents {
		keys = append(keys, k)
	}
	// Simple sort
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

// RepoDir returns the directory containing agents.json (the repo root).
func RepoDir(agentsJSONPath string) string {
	return filepath.Dir(agentsJSONPath)
}
