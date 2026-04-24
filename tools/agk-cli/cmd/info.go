package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current installations",
	Long:  "Displays details about installed agents from the lock file.",
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	projectDir, _ := os.Getwd()

	lf, err := ReadLockFile(projectDir)
	if err != nil {
		return err
	}
	if lf == nil {
		fmt.Println("  No installations found. Run 'agk install --ai <agent>'.")
		return nil
	}

	fmt.Println("📋 Installed Agents")
	fmt.Println()
	PrintLockInfo(lf)
	fmt.Println()
	return nil
}
