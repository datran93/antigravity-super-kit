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
	Long: `Start the Antigravity Kit dashboard.
Automatically installs dependencies and builds if needed.

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

	// Build if needed
	if _, err := os.Stat(filepath.Join(uiDir, ".next")); os.IsNotExist(err) {
		fmt.Println("🔨 Building dashboard...")
		npmBuild := exec.Command("npm", "run", "build")
		npmBuild.Dir = uiDir
		npmBuild.Stdout = os.Stdout
		npmBuild.Stderr = os.Stderr
		if err := npmBuild.Run(); err != nil {
			return fmt.Errorf("npm build failed: %w", err)
		}
	}

	// Set workspace path
	cwd, _ := os.Getwd()
	os.Setenv("WORKSPACE_PATH", cwd)

	// Start the server
	npmStart := exec.Command("npm", "run", "start")
	npmStart.Dir = uiDir
	npmStart.Stdout = os.Stdout
	npmStart.Stderr = os.Stderr

	if flagOpen {
		if err := npmStart.Start(); err != nil {
			return fmt.Errorf("failed to start dashboard: %w", err)
		}

		fmt.Println("⏳ Waiting for server to start...")
		time.Sleep(2 * time.Second)

		openBrowser("http://localhost:3000")
		return npmStart.Wait()
	}

	return npmStart.Run()
}

func findDashboardDir() string {
	// Try relative to CWD
	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(cwd, "ui", "agk-dashboard"),
		filepath.Join(cwd, "..", "ui", "agk-dashboard"),
	}

	// Try relative to executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "..", "ui", "agk-dashboard"),
			filepath.Join(exeDir, "..", "..", "ui", "agk-dashboard"),
		)
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			absDir, _ := filepath.Abs(dir)
			return absDir
		}
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
