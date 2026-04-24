package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove agent configuration from project",
	Long: `Remove an installed agent's configuration files.
Prompts for confirmation before deleting.

Examples:
  agk remove --ai claude     # Remove Claude installation
  agk remove --ai claude --force  # Skip confirmation`,
	RunE: runRemove,
}

var flagForce bool

func init() {
	removeCmd.Flags().StringSliceVar(&flagAI, "ai", nil, "target agent(s) (can be repeated)")
	removeCmd.Flags().BoolVar(&flagForce, "force", false, "skip confirmation prompt")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	agents, af, err := resolveAgents()
	if err != nil {
		return err
	}

	projectDir, _ := os.Getwd()

	// Show warning
	for _, agentKey := range agents {
		agent, _ := af.GetAgent(agentKey)
		if agent != nil {
			fmt.Printf("⚠️  This will delete %s/ for %s.\n", agent.TargetDir, agentKey)
		}
	}

	// Confirm
	if !flagForce {
		fmt.Print("Continue? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("ℹ️  Cancelled.")
			return nil
		}
	}

	for _, agentKey := range agents {
		agent, err := af.GetAgent(agentKey)
		if err != nil {
			continue
		}

		targetPath := filepath.Join(projectDir, agent.TargetDir)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			fmt.Printf("⚠️  %s not found for %s.\n", agent.TargetDir, agentKey)
			continue
		}

		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", targetPath, err)
		}

		if err := RemoveLockEntry(projectDir, agentKey); err != nil {
			fmt.Printf("⚠️  Lock file update failed: %v\n", err)
		}

		fmt.Printf("✅ Removed %s (%s/)\n", agent.Name, agent.TargetDir)
	}
	return nil
}
