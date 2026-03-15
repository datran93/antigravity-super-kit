package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

// hasBracketSteps returns true when at least one step starts with '[',
// indicating that steps use the [T1] / [Px-Ty] label convention.
func hasBracketSteps(completed, pending []string) bool {
	for _, s := range completed {
		if len(s) > 0 && s[0] == '[' {
			return true
		}
	}
	for _, s := range pending {
		if len(s) > 0 && s[0] == '[' {
			return true
		}
	}
	return false
}

// WriteMarkdownProgress renders progress.md with burndown dashboard, drift heatmap,
// DAG visualization, and git SHA footer.
// gitSHA is passed in by SaveCheckpoint (already captured once) to avoid a second
// subprocess call for the footer line.
func WriteMarkdownProgress(db *sql.DB, workspacePath, taskID, description, status string, completedSteps, nextSteps, activeFiles []string, notes, gitSHA string, idleThresholdDays int) error {
	mdPath := filepath.Join(workspacePath, "progress.md")

	// Historical completed tasks (other task_ids)
	// Use LOWER() for backward-compat with any uppercase status values stored in DB.
	rows, err := db.Query("SELECT task_id, description FROM checkpoints WHERE LOWER(status) = 'completed' AND task_id != ? ORDER BY updated_at DESC", taskID)
	var historicalTasks []map[string]string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tID, tDesc string
			if err := rows.Scan(&tID, &tDesc); err == nil {
				historicalTasks = append(historicalTasks, map[string]string{"task_id": tID, "description": tDesc})
			}
		}
	}

	// ── Load step metadata for burndown rendering ──────────────────────────────
	var stepTimestampsStr, stepDriftStr string
	db.QueryRow("SELECT COALESCE(step_timestamps,'{}'), COALESCE(step_drift,'{}') FROM checkpoints WHERE task_id=?", taskID).
		Scan(&stepTimestampsStr, &stepDriftStr) //nolint:errcheck — fallback to empty maps on failure
	stepTs := ParseStepTimestamps(stepTimestampsStr)
	stepDrift := ParseStepDrift(stepDriftStr)

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

	// ── Burndown header (velocity + ETA) ──────────────────────────────────────
	content += RenderBurndownHeader(taskID, stepTs, len(nextSteps))

	if len(activeFiles) > 0 {
		content += "### 📁 Active Files\n"
		for _, f := range activeFiles {
			content += fmt.Sprintf("- `%s`\n", f)
		}
		content += "\n"
	}

	// ── Parse DAG deps (opt-in: steps with 'depends:[...]' suffix) ─────────────
	allRaw := append(append([]string{}, completedSteps...), nextSteps...)
	_, deps := ParseStepDeps(allRaw)
	completedSet := BuildCompletedSet(completedSteps)

	// ── Step rendering ─────────────────────────────────────────────────────────
	if hasBracketSteps(completedSteps, nextSteps) {
		// Steps use [T1] / [Px-Ty] labels: render under a single flat header.
		content += "### 📋 Steps Overview\n\n"
		for _, s := range completedSteps {
			content += RenderStepWithMeta(s, true, stepTs, stepDrift)
		}
		for _, s := range nextSteps {
			content += RenderStepWithMeta(s, false, stepTs, stepDrift)
		}
		content += "\n"
	} else {
		// Plain steps (no label prefix): use section-header rendering.
		content += RenderBurndownSection(completedSteps, nextSteps, stepTs, stepDrift)
	}

	// ── DAG block (emitted only when steps have 'depends:[...]' declarations) ──
	if dagBlock := RenderDAGBlock(allRaw, deps, completedSet); dagBlock != "" {
		content += dagBlock
	}

	if notes != "" {
		content += "### 📝 Log & Notes\n"
		content += fmt.Sprintf("```text\n%s\n```\n\n", notes)
	}

	if len(historicalTasks) > 0 {
		content += "---\n### 🏆 Historically Completed Tasks\n"
		for _, t := range historicalTasks {
			content += fmt.Sprintf("- **%s**: %s\n", t["task_id"], t["description"])
		}
		content += "\n"
	}

	// ── Historically Incomplete Tasks ────────────────────────────────────────
	if idleThresholdDays > 0 {
		if idleTasks, err := fetchIdleTasks(workspacePath, taskID, idleThresholdDays); err == nil && len(idleTasks) > 0 {
			content += RenderHistoricallyIncomplete(idleTasks)
		}
	}

	// Footer: timestamp + short git SHA (reuse the SHA already captured by SaveCheckpoint)
	shaStr := ""
	if gitSHA != "" && len(gitSHA) >= 7 {
		shaStr = fmt.Sprintf(" | 🔗 Git: `%s`", gitSHA[:7])
	}
	content += fmt.Sprintf("---\n*Last sync: %s%s*", time.Now().Format("2006-01-02 15:04:05"), shaStr)
	return os.WriteFile(mdPath, []byte(content), 0644)
}

// NormalizeStatus converts any completion-alias status to the canonical value
// "completed" that progress.md SQL queries and the progress renderer rely on.
// Non-completion statuses (e.g. "in_progress", "blocked") are only lowercased.
func NormalizeStatus(status string) string {
	s := strings.ToLower(status)
	switch s {
	case "done", "committed", "complete", "finished", "closed":
		return "completed"
	}
	return s
}

func SaveCheckpoint(workspacePath, taskID, description, status string, completedSteps, nextSteps, activeFiles []string, notes string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Normalize status via canonical function (maps aliases like "done", "committed" → "completed")
	status = NormalizeStatus(status)
	now := time.Now().Format(time.RFC3339)

	// Capture current git SHA for traceability (empty string if not a git repo)
	gitSHA := captureGitSHA(workspacePath)

	compBytes, _ := json.Marshal(completedSteps)
	nextBytes, _ := json.Marshal(nextSteps)
	activeBytes, _ := json.Marshal(activeFiles)

	query := `
        INSERT INTO checkpoints (task_id, description, status, completed_steps, next_steps, active_files, notes, updated_at, git_sha)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(task_id) DO UPDATE SET
            description=excluded.description,
            status=excluded.status,
            completed_steps=excluded.completed_steps,
            next_steps=excluded.next_steps,
            active_files=excluded.active_files,
            notes=excluded.notes,
            updated_at=excluded.updated_at,
            git_sha=excluded.git_sha
    `
	_, err = db.Exec(query, taskID, description, status, string(compBytes), string(nextBytes), string(activeBytes), notes, now, gitSHA)
	if err != nil {
		return "", fmt.Errorf("failed to save checkpoint table: %v", err)
	}

	if err := WriteMarkdownProgress(db, workspacePath, taskID, description, status, completedSteps, nextSteps, activeFiles, notes, gitSHA, 3); err != nil {
		fmt.Printf("Error writing progress.md: %v\n", err)
	}

	// No per-step-group compact hint needed — steps are sequential, not grouped.

	msg := fmt.Sprintf("✅ Checkpoint '%s' saved.", taskID)
	if gitSHA != "" {
		msg += fmt.Sprintf(" [git: %s]", gitSHA[:min(7, len(gitSHA))])
	}
	if len(nextSteps) == 0 && len(completedSteps) > 0 {
		msg += "\n\n🎉 ALL TASKS COMPLETED! Great job."
	}
	return msg, nil
}
