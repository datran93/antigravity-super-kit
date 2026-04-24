package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update agent configuration to latest version",
	Long: `Update an already installed agent's configuration.
Syncs the source repository and re-applies transformations.

Examples:
  agk update                 # Update all installed agents
  agk update --ai claude     # Update Claude installation only`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().StringSliceVar(&flagAI, "ai", nil, "target agent(s) (can be repeated)")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Sync first so the cached agents.json is available as fallback
	cacheDir := DefaultCacheDir()
	if err := SyncRepo(DefaultRepoURL, cacheDir, DefaultRepoName); err != nil {
		return err
	}

	agents, af, err := resolveAgents()
	if err != nil {
		return err
	}

	sourceVersion := GetSourceVersion(cacheDir, DefaultRepoName)
	sourceDir := SourceAgentsDir(cacheDir, DefaultRepoName)
	projectDir, _ := os.Getwd()

	for _, agentKey := range agents {
		agent, err := af.GetAgent(agentKey)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(projectDir, agent.TargetDir)

		// If not installed yet, do a fresh install
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			fmt.Printf("📦 %s not found. Installing...\n", agent.TargetDir)
		} else {
			fmt.Printf("🔄 Updating %s...\n", agent.TargetDir)
			os.RemoveAll(targetPath)
		}

		if err := InstallAgent(agent, sourceDir, projectDir); err != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", agentKey, err)
			continue
		}

		fileCount := CountFiles(targetPath)
		if err := UpdateLockEntry(projectDir, agentKey, agent.TargetDir, sourceVersion, fileCount); err != nil {
			fmt.Printf("⚠️  Lock file update failed: %v\n", err)
		}

		fmt.Printf("✅ Updated %s → %s/ (%d files)\n", agent.Name, agent.TargetDir, fileCount)
	}
	return nil
}
