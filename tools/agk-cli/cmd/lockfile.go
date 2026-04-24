package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LockFile represents .agk.lock.json.
type LockFile struct {
	SchemaVersion string                `json:"schema_version"`
	SourceVersion string                `json:"source_version"`
	Agents        map[string]LockAgent  `json:"agents"`
}

// LockAgent records installation metadata for one agent.
type LockAgent struct {
	TargetDir      string `json:"target_dir"`
	FilesInstalled int    `json:"files_installed"`
	InstalledAt    string `json:"installed_at"`
}

const lockFileName = ".agk.lock.json"

// ReadLockFile reads and parses the lock file. Returns nil if not found.
func ReadLockFile(projectDir string) (*LockFile, error) {
	path := filepath.Join(projectDir, lockFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var lf LockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", lockFileName, err)
	}
	return &lf, nil
}

// WriteLockFile writes the lock file to disk.
func WriteLockFile(projectDir string, lf *LockFile) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(projectDir, lockFileName), data, 0644)
}

// UpdateLockEntry updates or adds an agent entry in the lock file.
func UpdateLockEntry(projectDir, agentKey, targetDir, sourceVersion string, fileCount int) error {
	lf, err := ReadLockFile(projectDir)
	if err != nil {
		return err
	}
	if lf == nil {
		lf = &LockFile{
			SchemaVersion: "1.0",
			Agents:        make(map[string]LockAgent),
		}
	}

	lf.SourceVersion = sourceVersion
	lf.Agents[agentKey] = LockAgent{
		TargetDir:      targetDir,
		FilesInstalled: fileCount,
		InstalledAt:    time.Now().UTC().Format(time.RFC3339),
	}

	return WriteLockFile(projectDir, lf)
}

// RemoveLockEntry removes an agent from the lock file.
// Deletes the file if no agents remain.
func RemoveLockEntry(projectDir, agentKey string) error {
	lf, err := ReadLockFile(projectDir)
	if err != nil || lf == nil {
		return err
	}

	delete(lf.Agents, agentKey)

	if len(lf.Agents) == 0 {
		return os.Remove(filepath.Join(projectDir, lockFileName))
	}
	return WriteLockFile(projectDir, lf)
}

// PrintLockInfo displays the lock file contents.
func PrintLockInfo(lf *LockFile) {
	fmt.Printf("  Source version: %s\n\n", lf.SourceVersion)
	for key, info := range lf.Agents {
		fmt.Printf("  %-12s  %-15s  %d files  (%s)\n", key, info.TargetDir, info.FilesInstalled, info.InstalledAt)
	}
}
