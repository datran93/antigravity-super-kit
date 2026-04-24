package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install agent configuration into current project",
	Long: `Install agent configuration files into the current project.
Syncs the source repository and transforms rules/workflows for the target agent.

Examples:
  agk install                          # Install for Antigravity (default)
  agk install --ai claude              # Install for Claude Code
  agk install --ai claude --ai gemini  # Install for multiple agents`,
	RunE: runInstall,
}

func init() {
	installCmd.Flags().StringSliceVar(&flagAI, "ai", nil, "target agent(s) (can be repeated)")
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	agents, af, err := resolveAgents()
	if err != nil {
		return err
	}

	cacheDir := DefaultCacheDir()
	if err := SyncRepo(DefaultRepoURL, cacheDir, DefaultRepoName); err != nil {
		return err
	}

	sourceVersion := GetSourceVersion(cacheDir, DefaultRepoName)
	sourceDir := SourceAgentsDir(cacheDir, DefaultRepoName)
	projectDir, _ := os.Getwd()

	successCount := 0
	for _, agentKey := range agents {
		agent, err := af.GetAgent(agentKey)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(projectDir, agent.TargetDir)
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("⚠️  %s already exists. Use 'agk update --ai %s' or remove it first.\n", agent.TargetDir, agentKey)
			continue
		}

		fmt.Printf("📦 Installing %s for %s...\n", agent.TargetDir, agent.Name)
		if err := InstallAgent(agent, sourceDir, projectDir); err != nil {
			fmt.Printf("❌ Failed to install %s: %v\n", agentKey, err)
			continue
		}

		fileCount := CountFiles(targetPath)
		if err := UpdateLockEntry(projectDir, agentKey, agent.TargetDir, sourceVersion, fileCount); err != nil {
			fmt.Printf("⚠️  Lock file update failed: %v\n", err)
		}

		fmt.Printf("✅ Installed %s → %s/ (%d files)\n", agent.Name, agent.TargetDir, fileCount)
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("no agents were installed")
	}
	return nil
}
