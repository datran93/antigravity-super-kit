package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the AGK dashboard UI",
	Long: `Start the Antigravity Kit dashboard in development mode.
Automatically installs dependencies if needed.

Examples:
  agk ui           # Start the dashboard
  agk ui --open    # Start and open in browser`,
	RunE: runUI,
}

var flagOpen bool

func init() {
	uiCmd.Flags().BoolVar(&flagOpen, "open", false, "open browser after starting")
	rootCmd.AddCommand(uiCmd)
}

func runUI(cmd *cobra.Command, args []string) error {
	// Find the dashboard directory relative to the agk binary or workspace
	uiDir := findDashboardDir()
	if uiDir == "" {
		return fmt.Errorf("dashboard directory not found. Expected ui/agk-dashboard/ in the workspace")
	}

	fmt.Println("🚀 Starting Antigravity Kit Dashboard...")

	// Install dependencies if needed
	if _, err := os.Stat(filepath.Join(uiDir, "node_modules")); os.IsNotExist(err) {
		fmt.Println("📦 Installing dashboard dependencies...")
		npmInstall := exec.Command("npm", "install")
		npmInstall.Dir = uiDir
		npmInstall.Stdout = os.Stdout
		npmInstall.Stderr = os.Stderr
		if err := npmInstall.Run(); err != nil {
			return fmt.Errorf("npm install failed: %w", err)
		}
	}

	// Derive workspace root from dashboard dir (ui/agk-dashboard -> repo root)
	workspaceRoot := filepath.Dir(filepath.Dir(uiDir))
	os.Setenv("WORKSPACE_PATH", workspaceRoot)

	// Use dev mode so source changes are always reflected immediately
	npmDev := exec.Command("npm", "run", "dev")
	npmDev.Dir = uiDir
	npmDev.Stdout = os.Stdout
	npmDev.Stderr = os.Stderr

	if flagOpen {
		if err := npmDev.Start(); err != nil {
			return fmt.Errorf("failed to start dashboard: %w", err)
		}

		fmt.Println("⏳ Waiting for server to start...")
		time.Sleep(2 * time.Second)

		openBrowser("http://localhost:3000")
		return npmDev.Wait()
	}

	return npmDev.Run()
}

func findDashboardDir() string {
	// The binary is at antigravity-kit/tools/agk-cli/agk (possibly symlinked).
	// The dashboard is at antigravity-kit/ui/agk-dashboard.
	// Resolve the real path of the executable and navigate from there.
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}

	// Follow symlinks to get the real location
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}

	// exeDir = antigravity-kit/tools/agk-cli/
	exeDir := filepath.Dir(realPath)
	// Go up two levels to antigravity-kit/, then into ui/agk-dashboard
	candidate := filepath.Join(exeDir, "..", "..", "ui", "agk-dashboard")
	absDir, err := filepath.Abs(candidate)
	if err != nil {
		return ""
	}

	if info, err := os.Stat(absDir); err == nil && info.IsDir() {
		return absDir
	}
	return ""
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		fmt.Printf("🌐 Open %s in your browser\n", url)
		return
	}
	cmd.Start()
}
