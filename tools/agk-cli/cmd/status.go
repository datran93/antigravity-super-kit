package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check installation status",
	Long:  "Shows the source version and status of all installed agents.",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectDir, _ := os.Getwd()

	lf, err := ReadLockFile(projectDir)
	if err != nil {
		return err
	}
	if lf == nil {
		return fmt.Errorf("no installations found. Run 'agk install --ai <agent>'")
	}

	fmt.Println("🔧 Antigravity Kit Status")
	fmt.Println()
	PrintLockInfo(lf)
	return nil
}
