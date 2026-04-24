package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// captureGitSHA runs `git rev-parse HEAD` in the given directory with a 2s timeout.
// Returns empty string if the directory is not a git repo or git is unavailable.
func captureGitSHA(workspacePath string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", workspacePath, "rev-parse", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// WriteMarkdownProgress renders progress.md.
func WriteMarkdownProgress(db *sql.DB, workspacePath, taskID string) error {
	mdPath := filepath.Join(workspacePath, "progress.md")

	// 1. Fetch Task Info
	var description, status, notes string
	var ac sql.NullString
	err := db.QueryRow("SELECT description, status, notes, acceptance_criteria FROM tasks WHERE task_id = ?", taskID).Scan(&description, &status, &notes, &ac)
	if err == sql.ErrNoRows {
		// Task deleted or doesn't exist
		description = "[deleted]"
		status = "deleted"
	} else if err != nil {
		return fmt.Errorf("query task: %v", err)
	}

	// 2. Fetch Steps (with time tracking)
	type stepEntry struct {
		name      string
		duration  string // e.g. "2m30s" or ""
	}
	var completedSteps, nextSteps []stepEntry
	rows, err := db.Query("SELECT name, status, started_at, completed_at FROM steps WHERE task_id = ? ORDER BY created_at ASC", taskID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name, s string
			var startedAt, completedAt sql.NullString
			if err := rows.Scan(&name, &s, &startedAt, &completedAt); err == nil {
				dur := ""
				if startedAt.Valid && completedAt.Valid {
					t1, e1 := time.Parse(time.RFC3339, startedAt.String)
					t2, e2 := time.Parse(time.RFC3339, completedAt.String)
					if e1 == nil && e2 == nil {
						dur = t2.Sub(t1).Truncate(time.Second).String()
					}
				}
				entry := stepEntry{name: name, duration: dur}
				if s == "completed" {
					completedSteps = append(completedSteps, entry)
				} else {
					nextSteps = append(nextSteps, entry)
				}
			}
		}
	}

	totalSteps := len(completedSteps) + len(nextSteps)
	progressPct := float64(0)
	if totalSteps > 0 {
		progressPct = float64(len(completedSteps)) / float64(totalSteps) * 100.0
	}
	barLen := 20
	filled := int(float64(barLen) * progressPct / 100.0)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)

	content := fmt.Sprintf("# 🚀 Project Progress: %s\n\n", taskID)
	content += fmt.Sprintf("**Status:** `%s` | **Progress:** `[%s] %.1f%%` (%d/%d)\n\n",
		strings.ToUpper(status), bar, progressPct, len(completedSteps), totalSteps)
	content += fmt.Sprintf("> %s\n\n", description)

	// Fetch intent locks (Active Files)
	var lockedFilesStr string
	err = db.QueryRow("SELECT locked_files FROM intents WHERE task_id = ?", taskID).Scan(&lockedFilesStr)
	if err == nil && lockedFilesStr != "" && lockedFilesStr != "[]" {
		content += "### 📁 Active Intent Locks\n"
		// simple parse of stringified JSON array
		lockedFilesStr = strings.Trim(lockedFilesStr, "[]\"")
		files := strings.Split(lockedFilesStr, "\",\"")
		for _, f := range files {
			if f != "" {
				content += fmt.Sprintf("- `%s`\n", f)
			}
		}
		content += "\n"
	}

	content += "### 📋 Steps Overview\n\n"
	for _, s := range completedSteps {
		if s.duration != "" {
			content += fmt.Sprintf("- [x] %s ⏱️ %s\n", s.name, s.duration)
		} else {
			content += fmt.Sprintf("- [x] %s\n", s.name)
		}
	}
	for _, s := range nextSteps {
		content += fmt.Sprintf("- [ ] %s\n", s.name)
	}
	content += "\n"

	if ac.Valid && ac.String != "" {
		content += "### 🎯 Acceptance Criteria\n"
		content += fmt.Sprintf("%s\n\n", ac.String)
	}

	if notes != "" {
		content += "### 📝 Log & Notes\n"
		content += fmt.Sprintf("```text\n%s\n```\n\n", notes)
	}

	// Historical tasks
	hRows, err := db.Query("SELECT task_id, description FROM tasks WHERE LOWER(status) = 'completed' AND task_id != ? ORDER BY updated_at DESC", taskID)
	if err == nil {
		defer hRows.Close()
		var historical []string
		for hRows.Next() {
			var tID, tDesc string
			if err := hRows.Scan(&tID, &tDesc); err == nil {
				historical = append(historical, fmt.Sprintf("- **%s**: %s\n", tID, tDesc))
			}
		}
		if len(historical) > 0 {
			content += "---\n### 🏆 Historically Completed Tasks\n"
			for _, h := range historical {
				content += h
			}
			content += "\n"
		}
	}

	if idleTasks, err := fetchIdleTasks(workspacePath, taskID, 3); err == nil && len(idleTasks) > 0 {
		content += "---\n### ⏳ Historically Incomplete Tasks\n"
		for _, t := range idleTasks {
			content += fmt.Sprintf("- **%s** [%.0f%%] — %s _(Idle for %d days)_\n", t.TaskID, t.Progress, t.Description, t.IdleDays)
		}
		content += "\n"
	}

	gitSHA := captureGitSHA(workspacePath)
	shaStr := ""
	if gitSHA != "" && len(gitSHA) >= 7 {
		shaStr = fmt.Sprintf(" | 🔗 Git: `%s`", gitSHA[:7])
	}
	content += fmt.Sprintf("---\n*Last sync: %s%s*", time.Now().Format("2006-01-02 15:04:05"), shaStr)

	return os.WriteFile(mdPath, []byte(content), 0644)
}

// NormalizeStatus converts any completion-alias status to the canonical value
func NormalizeStatus(status string) string {
	s := strings.ToLower(status)
	switch s {
	case "done", "committed", "complete", "finished", "closed":
		return "completed"
	}
	return s
}

// SaveCheckpoint handles the mcp tool save_checkpoint
func SaveCheckpoint(workspacePath, taskID, description, status string, completedSteps, nextSteps, activeFiles []string, notes string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	status = NormalizeStatus(status)

	err = RunInTx(db, func(tx *sql.Tx) error {
		// Update tasks
		_, err := tx.Exec("INSERT INTO tasks (task_id, description, status, notes) VALUES (?, ?, ?, ?) ON CONFLICT(task_id) DO UPDATE SET description=excluded.description, status=excluded.status, notes=excluded.notes, updated_at=CURRENT_TIMESTAMP",
			taskID, description, status, notes)
		if err != nil {
			return err
		}

		// Clear existing steps and insert new ones to reflect the explicit state
		tx.Exec("DELETE FROM steps WHERE task_id = ?", taskID)

		for i, s := range completedSteps {
			info := ParseStep(s)
			if info.ID == "" {
				info.ID = fmt.Sprintf("CSTEP-%03d", i+1)
			}
			tx.Exec("INSERT INTO steps (step_id, task_id, name, status) VALUES (?, ?, ?, 'completed')", info.ID, taskID, info.Name)
		}

		for i, s := range nextSteps {
			info := ParseStep(s)
			if info.ID == "" {
				info.ID = fmt.Sprintf("PSTEP-%03d", i+1)
			}
			tx.Exec("INSERT INTO steps (step_id, task_id, name, status) VALUES (?, ?, ?, 'pending')", info.ID, taskID, info.Name)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to save checkpoint: %v", err)
	}

	if err := WriteMarkdownProgress(db, workspacePath, taskID); err != nil {
		fmt.Fprintf(os.Stderr, "[context-manager] error writing progress.md: %v\n", err)
	}

	msg := fmt.Sprintf("✅ Checkpoint '%s' saved.", taskID)
	if len(nextSteps) == 0 && len(completedSteps) > 0 {
		msg += "\n\n🎉 ALL TASKS COMPLETED! Great job."
	}
	return msg, nil
}
