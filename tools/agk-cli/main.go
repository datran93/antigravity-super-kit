// Package main is the entry point for the agk CLI binary.
package main

import (
	"fmt"
	"os"

	"agk-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
