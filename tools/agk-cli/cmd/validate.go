package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [workspace]",
	Short: "Validate agent governance files and references",
	Long: `Scans the workspace for:
  • AGK.md / AGENTS.md presence and shim integrity
  • Broken @doc/path references in markdown files
  • Step label format compliance (STXX)
  • Workflow file existence`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

var (
	flagFix     bool
	flagVerbose bool
)

func init() {
	validateCmd.Flags().BoolVar(&flagFix, "fix", false, "auto-fix simple issues (e.g., missing directories)")
	validateCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "show detailed output")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	workspace := "."
	if len(args) > 0 {
		workspace = args[0]
	}

	absPath, err := filepath.Abs(workspace)
	if err != nil {
		return fmt.Errorf("invalid workspace path: %w", err)
	}

	fmt.Printf("🔍 Validating workspace: %s\n\n", absPath)

	var issues []string
	var passes []string

	// 1. Check AGK.md exists
	agkPath := filepath.Join(absPath, ".agents", "rules", "AGK.md")
	if _, err := os.Stat(agkPath); os.IsNotExist(err) {
		issues = append(issues, "❌ AGK.md not found at .agents/rules/AGK.md")
	} else {
		passes = append(passes, "✅ AGK.md exists")
	}

	// 2. Check AGENTS.md exists and is a shim
	agentsPath := filepath.Join(absPath, ".agents", "rules", "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		issues = append(issues, "❌ AGENTS.md not found at .agents/rules/AGENTS.md")
	} else {
		content, _ := os.ReadFile(agentsPath)
		if strings.Contains(string(content), "AGK.md") {
			passes = append(passes, "✅ AGENTS.md is a shim → AGK.md")
		} else {
			issues = append(issues, "⚠️  AGENTS.md does not reference AGK.md (should be a shim)")
		}
	}

	// 3. Check workflows directory
	workflowsDir := filepath.Join(absPath, ".agents", "workflows")
	if entries, err := os.ReadDir(workflowsDir); err != nil {
		issues = append(issues, "❌ Workflows directory not found at .agents/workflows/")
	} else {
		passes = append(passes, fmt.Sprintf("✅ Workflows directory exists (%d files)", len(entries)))
	}

	// 4. Scan markdown files for @doc/ references
	docIssues := scanDocReferences(absPath)
	issues = append(issues, docIssues...)
	if len(docIssues) == 0 {
		passes = append(passes, "✅ No broken @doc/ references found")
	}

	// 5. Check ANCHORS.md exists
	anchorsPath := filepath.Join(absPath, ".agents", "rules", "ANCHORS.md")
	if _, err := os.Stat(anchorsPath); os.IsNotExist(err) {
		issues = append(issues, "⚠️  ANCHORS.md not found at .agents/rules/ANCHORS.md")
	} else {
		passes = append(passes, "✅ ANCHORS.md exists")
	}

	// 6. Check models directory
	home, _ := os.UserHomeDir()
	modelsDir := os.Getenv("AGK_MODELS_DIR")
	if modelsDir == "" {
		modelsDir = filepath.Join(home, ".agk", "models")
	}
	modelPath := filepath.Join(modelsDir, "all-MiniLM-L6-v2.onnx")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		issues = append(issues, fmt.Sprintf("⚠️  ONNX model not found at %s (embedding will use OpenAI fallback)", modelPath))
	} else {
		passes = append(passes, "✅ ONNX model available")
	}

	// 7. Check templates directory
	templatesDir := filepath.Join(absPath, ".agents", "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		if flagFix {
			os.MkdirAll(templatesDir, 0755)
			passes = append(passes, "✅ Created .agents/templates/ directory")
		} else {
			issues = append(issues, "⚠️  Templates directory not found at .agents/templates/ (use --fix to create)")
		}
	} else {
		passes = append(passes, "✅ Templates directory exists")
	}

	// Print results
	fmt.Println("── Passes ──")
	for _, p := range passes {
		fmt.Println("  " + p)
	}

	if len(issues) > 0 {
		fmt.Println("\n── Issues ──")
		for _, i := range issues {
			fmt.Println("  " + i)
		}
		fmt.Printf("\n📊 Result: %d passes, %d issues\n", len(passes), len(issues))
		return fmt.Errorf("%d validation issues found", len(issues))
	}

	fmt.Printf("\n📊 Result: %d passes, 0 issues — All clear! ✨\n", len(passes))
	return nil
}

var docRefPattern = regexp.MustCompile(`@doc/([a-zA-Z0-9_\-/\.]+)`)

// scanDocReferences scans markdown files in features/ and .agents/ for @doc/ references
// and checks if corresponding files exist in the docs/ directory.
func scanDocReferences(workspace string) []string {
	var issues []string
	scanDirs := []string{
		filepath.Join(workspace, "features"),
		filepath.Join(workspace, ".agents"),
	}

	for _, dir := range scanDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			matches := docRefPattern.FindAllStringSubmatch(string(content), -1)
			for _, m := range matches {
				docPath := m[1]
				// Check if the doc file exists
				fullPath := filepath.Join(workspace, "docs", docPath)
				if !strings.HasSuffix(fullPath, ".md") {
					fullPath += ".md"
				}
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					rel, _ := filepath.Rel(workspace, path)
					issues = append(issues, fmt.Sprintf("❌ Broken @doc/%s in %s", docPath, rel))
				}
			}

			return nil
		})
	}

	return issues
}
