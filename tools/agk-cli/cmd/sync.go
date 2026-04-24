package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultCacheDir returns the default cache directory for AGK.
func DefaultCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".antigravity", "cache")
}

// DefaultRepoURL is the default source repository.
const DefaultRepoURL = "git@github.com:datran93/antigravity-super-kit.git"
const DefaultRepoName = "antigravity-kit"

// SyncRepo ensures the source repo is up-to-date in the cache.
func SyncRepo(repoURL, cacheDir, repoName string) error {
	repoDir := filepath.Join(cacheDir, repoName)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	if isGitRepo(repoDir) {
		fmt.Println("🔍 Checking for updates...")

		localHead := gitRevParse(repoDir)
		remoteHead := gitLsRemoteHead(repoURL)

		if remoteHead == "" {
			fmt.Println("⚠️  Cannot reach remote. Using cached version.")
			return nil
		}

		if localHead == remoteHead {
			fmt.Println("✅ Cache is up to date.")
			return nil
		}

		fmt.Println("🔄 Updates available. Re-cloning...")
		os.RemoveAll(repoDir)
	}

	fmt.Println("📥 Cloning repository...")
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, repoDir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Println("✅ Repository synced.")
	return nil
}

// GetSourceVersion returns the short SHA of the cached repo.
func GetSourceVersion(cacheDir, repoName string) string {
	repoDir := filepath.Join(cacheDir, repoName)
	if !isGitRepo(repoDir) {
		return "unknown"
	}
	out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// SourceAgentsDir returns the path to .agents/ in the cached repo.
func SourceAgentsDir(cacheDir, repoName string) string {
	return filepath.Join(cacheDir, repoName, ".agents")
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func gitRevParse(repoDir string) string {
	out, err := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitLsRemoteHead(repoURL string) string {
	out, err := exec.Command("git", "ls-remote", repoURL, "HEAD").Output()
	if err != nil {
		return ""
	}
	parts := strings.Fields(string(out))
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
