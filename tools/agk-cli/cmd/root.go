// Package cmd defines the agk CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "agk",
	Short:   "Antigravity Kit — CLI toolkit for agent governance",
	Long:    "agk is the command-line interface for the Antigravity Kit.\nIt provides validation, import, and code generation utilities for agent-driven projects.",
	Version: version,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
