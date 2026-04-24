package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <source>",
	Short: "Import external skill/template packs into .agents/",
	Long: `Import agent skill packs from Git repositories or local directories.

Examples:
  agk import https://github.com/org/agk-skills.git
  agk import ./path/to/local/skills
  agk import https://github.com/org/agk-skills.git --target workflows
  agk import https://github.com/org/agk-skills.git --filter "*.md"`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

var (
	flagTarget string
	flagFilter string
	flagDryRun bool
)

func init() {
	importCmd.Flags().StringVar(&flagTarget, "target", "", "subdirectory under .agents/ to import into (e.g., workflows, templates, rules)")
	importCmd.Flags().StringVar(&flagFilter, "filter", "", "glob pattern to filter imported files (e.g., '*.md')")
	importCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "show what would be imported without writing files")
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	source := args[0]

	workspace, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	agentsDir := filepath.Join(workspace, ".agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return fmt.Errorf(".agents/ directory not found in %s — are you in an AGK workspace?", workspace)
	}

	destDir := agentsDir
	if flagTarget != "" {
		destDir = filepath.Join(agentsDir, flagTarget)
	}

	if isGitURL(source) {
		return importFromGit(source, destDir)
	}
	return importFromLocal(source, destDir)
}

func isGitURL(s string) bool {
	return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "git@") || strings.HasSuffix(s, ".git")
}

func importFromGit(url, destDir string) error {
	// Clone to a temp directory first
	tmpDir, err := os.MkdirTemp("", "agk-import-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("📥 Cloning %s...\n", url)
	gitCmd := exec.Command("git", "clone", "--depth=1", url, tmpDir)
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return copyFiles(tmpDir, destDir)
}

func importFromLocal(source, destDir string) error {
	absSource, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}

	info, err := os.Stat(absSource)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source must be a directory: %s", absSource)
	}

	fmt.Printf("📥 Importing from %s...\n", absSource)
	return copyFiles(absSource, destDir)
}

func copyFiles(srcDir, destDir string) error {
	var copied int

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories (.git, etc.)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}

		// Apply filter
		if flagFilter != "" {
			matched, _ := filepath.Match(flagFilter, info.Name())
			if !matched {
				return nil
			}
		}

		rel, _ := filepath.Rel(srcDir, path)
		destPath := filepath.Join(destDir, rel)

		if flagDryRun {
			fmt.Printf("  📄 %s → %s\n", rel, destPath)
			copied++
			return nil
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check for conflicts
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("  ⚠️  Skipping %s (already exists)\n", rel)
			return nil
		}

		if err := os.WriteFile(destPath, data, info.Mode()); err != nil {
			return err
		}

		fmt.Printf("  ✅ %s\n", rel)
		copied++
		return nil
	})

	if err != nil {
		return err
	}

	if flagDryRun {
		fmt.Printf("\n📊 Dry run: %d files would be imported\n", copied)
	} else {
		fmt.Printf("\n📊 Imported %d files into %s\n", copied, destDir)
	}
	return nil
}
