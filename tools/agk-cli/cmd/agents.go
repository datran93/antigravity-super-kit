package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "List all supported agents",
	Long:  "Displays all agent configurations defined in agents.json.",
	RunE:  runAgents,
}

func init() {
	rootCmd.AddCommand(agentsCmd)
}

func runAgents(cmd *cobra.Command, args []string) error {
	af, err := loadAgentsFromWorkspace()
	if err != nil {
		return err
	}

	fmt.Println("🤖 Supported Agents")
	fmt.Println()
	fmt.Printf("  %-12s  %-25s  %s\n", "KEY", "NAME", "TARGET DIR")
	fmt.Printf("  %-12s  %-25s  %s\n", "---", "----", "----------")

	for _, key := range af.AgentKeys() {
		agent := af.Agents[key]
		fmt.Printf("  %-12s  %-25s  → %s\n", key, agent.Name, agent.TargetDir)
	}
	fmt.Println()
	return nil
}
