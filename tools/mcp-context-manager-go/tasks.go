package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func InitializeTaskPlan(workspacePath, taskID, description string, steps []string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	initNotes := fmt.Sprintf("[%s] Task started.", time.Now().Format("15:04:05"))

	specPath := filepath.Join(workspacePath, "features", taskID, "spec.md")
	ac := ExtractAcceptanceCriteria(specPath)

	err = RunInTx(db, func(tx *sql.Tx) error {
		// Insert into tasks
		_, err := tx.Exec("INSERT INTO tasks (task_id, description, status, notes, acceptance_criteria) VALUES (?, ?, ?, ?, ?)",
			taskID, description, "in_progress", initNotes, ac)
		if err != nil {
			return fmt.Errorf("insert task: %v", err)
		}

		// Insert into steps and dependencies
		for i, stepStr := range steps {
			info := ParseStep(stepStr)
			if info.ID == "" {
				// Fallback if no ID was matched
				info.ID = fmt.Sprintf("STEP-%03d", i+1)
			}
			_, err = tx.Exec("INSERT INTO steps (step_id, task_id, name, status) VALUES (?, ?, ?, ?)",
				info.ID, taskID, info.Name, "pending")
			if err != nil {
				return fmt.Errorf("insert step %s: %v", info.ID, err)
			}
			
			for _, dep := range info.Dependencies {
				_, err = tx.Exec("INSERT INTO step_dependencies (step_id, depends_on_step_id) VALUES (?, ?)",
					info.ID, dep)
				if err != nil {
					return fmt.Errorf("insert step_dependency %s->%s: %v", info.ID, dep, err)
				}
			}
		}
		return nil
	})
	
	if err != nil {
		return "", err
	}
	
	_ = WriteMarkdownProgress(db, workspacePath, taskID)
	
	return fmt.Sprintf("✅ Task '%s' initialized with %d steps.", taskID, len(steps)), nil
}

// GetTaskSummary returns a compact JSON summary of a task's current state.
// Designed for quick status checks without loading the full checkpoint.
func GetTaskSummary(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var status, updatedAt, desc, notes string
	err = db.QueryRow("SELECT status, updated_at, description, notes FROM tasks WHERE task_id = ?", taskID).Scan(&status, &updatedAt, &desc, &notes)
	if err == sql.ErrNoRows {
		return fmt.Sprintf(`{"error": "task '%s' not found"}`, taskID), nil
	} else if err != nil {
		return "", err
	}

	var compCount, pendingCount int
	err = db.QueryRow("SELECT COUNT(*) FROM steps WHERE task_id = ? AND status = 'completed'", taskID).Scan(&compCount)
	if err != nil {
		return "", err
	}
	err = db.QueryRow("SELECT COUNT(*) FROM steps WHERE task_id = ? AND status != 'completed'", taskID).Scan(&pendingCount)
	if err != nil {
		return "", err
	}

	total := compCount + pendingCount
	pct := 0.0
	if total > 0 {
		pct = float64(compCount) / float64(total) * 100.0
	}

	nextStep := ""
	err = db.QueryRow("SELECT name FROM steps WHERE task_id = ? AND status = 'pending' ORDER BY created_at ASC LIMIT 1", taskID).Scan(&nextStep)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	linkedCtx := FetchAutoLinksContext(db, desc+"\n"+notes)

	return fmt.Sprintf(
		`{"task_id": %q, "status": %q, "progress": "%d/%d steps (%.0f%%)", "next_step": %q, "last_updated": %q, "context": %q}`,
		taskID, status, compCount, total, pct, nextStep, updatedAt, linkedCtx,
	), nil
}

func CompleteTaskStep(workspacePath, taskID, stepName string, activeFiles []string, notes string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	info := ParseStep(stepName)

	var msg string
	err = RunInTx(db, func(tx *sql.Tx) error {
		// Find the step by exact name or ID
		var stepID, currentStatus string
		
		query := "SELECT step_id, status FROM steps WHERE task_id = ? AND (name = ? OR step_id = ?)"
		err := tx.QueryRow(query, taskID, stepName, info.ID).Scan(&stepID, &currentStatus)
		if err == sql.ErrNoRows {
			return fmt.Errorf("⚠️ Step '%s' not found.", stepName)
		} else if err != nil {
			return err
		}

		if currentStatus == "completed" {
			return fmt.Errorf("⚠️ Step '%s' is already completed.", stepName)
		}

		// Update step status
		_, err = tx.Exec("UPDATE steps SET status = 'completed' WHERE step_id = ?", stepID)
		if err != nil {
			return err
		}

		// Insert active files
		for _, f := range activeFiles {
			_, err = tx.Exec("INSERT INTO step_files (step_id, file_path, role) VALUES (?, ?, 'coder')", stepID, f)
			if err != nil {
				return err
			}
		}

		// Update tasks notes
		endTimeStr := time.Now().Format("15:04:05")
		newLog := fmt.Sprintf("\n[%s] ✅ Done: %s", endTimeStr, stepName)
		if len(activeFiles) > 0 {
			fromJSON, _ := json.Marshal(activeFiles)
			newLog += fmt.Sprintf("\n  - Files: %s", string(fromJSON))
		}
		if notes != "" {
			newLog += fmt.Sprintf("\n  - Notes: %s", notes)
		}

		_, err = tx.Exec("UPDATE tasks SET notes = notes || ? WHERE task_id = ?", newLog, taskID)
		if err != nil {
			return err
		}

		// Check if all steps are completed
		var pendingCount int
		err = tx.QueryRow("SELECT COUNT(*) FROM steps WHERE task_id = ? AND status != 'completed'", taskID).Scan(&pendingCount)
		if err != nil {
			return err
		}

		if pendingCount == 0 {
			_, err = tx.Exec("UPDATE tasks SET status = 'completed' WHERE task_id = ?", taskID)
			if err != nil {
				return err
			}
			msg = fmt.Sprintf("✅ Task '%s' fully completed! Step '%s' done.", taskID, stepName)
		} else {
			msg = fmt.Sprintf("✅ Step '%s' marked completed. %d steps remaining.", stepName, pendingCount)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Improvement #5: auto-annotate active files when notes contain key phrases
	if hasGotchaKeyword(notes) {
		for _, f := range activeFiles {
			AnnotateFile(workspacePath, f, fmt.Sprintf("[%s] %s", stepName, notes))
		}
	}

	_ = WriteMarkdownProgress(db, workspacePath, taskID)

	return msg, nil
}

// hasGotchaKeyword returns true when the notes string contains known gotcha trigger phrases.
func hasGotchaKeyword(notes string) bool {
	lower := strings.ToLower(notes)
	for _, kw := range []string{"gotcha:", "quirk:", "warning:", "caution:", "⚠️"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func AddTaskStep(workspacePath, taskID, newStep string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	info := ParseStep(newStep)

	var msg string
	err = RunInTx(db, func(tx *sql.Tx) error {
		// check if task exists
		var currentNotes string
		err := tx.QueryRow("SELECT notes FROM tasks WHERE task_id = ?", taskID).Scan(&currentNotes)
		if err == sql.ErrNoRows {
			return fmt.Errorf("❌ Task '%s' not found.", taskID)
		} else if err != nil {
			return err
		}

		if info.ID == "" {
			info.ID = "STEP-" + time.Now().Format("20060102150405")
		}

		// check if step exists
		var exists int
		err = tx.QueryRow("SELECT COUNT(*) FROM steps WHERE task_id = ? AND step_id = ?", taskID, info.ID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists > 0 {
			return fmt.Errorf("⚠️ Step '%s' already exists in task '%s'.", newStep, taskID)
		}

		// insert step
		_, err = tx.Exec("INSERT INTO steps (step_id, task_id, name, status) VALUES (?, ?, ?, 'pending')", info.ID, taskID, info.Name)
		if err != nil {
			return err
		}

		// insert dependencies
		for _, dep := range info.Dependencies {
			_, err = tx.Exec("INSERT INTO step_dependencies (step_id, depends_on_step_id) VALUES (?, ?)", info.ID, dep)
			if err != nil {
				return err
			}
		}

		// Update notes
		newLog := currentNotes + fmt.Sprintf("\n[%s] Added new step: %s", time.Now().Format("15:04:05"), newStep)
		_, err = tx.Exec("UPDATE tasks SET notes = ? WHERE task_id = ?", newLog, taskID)
		if err != nil {
			return err
		}

		msg = fmt.Sprintf("✅ Step '%s' added to task '%s'.", info.ID, taskID)
		return nil
	})

	if err != nil {
		return "", err
	}
	
	_ = WriteMarkdownProgress(db, workspacePath, taskID)
	
	return msg, nil
}

func LoadCheckpoint(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var status, updatedAt, notes, desc string
	var ac sql.NullString
	err = db.QueryRow("SELECT status, updated_at, notes, description, acceptance_criteria FROM tasks WHERE task_id = ?", taskID).Scan(&status, &updatedAt, &notes, &desc, &ac)
	if err == sql.ErrNoRows {
		return fmt.Sprintf("❌ Task '%s' not found.", taskID), nil
	} else if err != nil {
		return "", err
	}

	rows, err := db.Query("SELECT name, status FROM steps WHERE task_id = ? ORDER BY created_at ASC", taskID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var comp []string
	var nxt []string

	for rows.Next() {
		var name, stepStatus string
		if err := rows.Scan(&name, &stepStatus); err == nil {
			if stepStatus == "completed" {
				comp = append(comp, name)
			} else {
				nxt = append(nxt, name)
			}
		}
	}

	total := len(comp) + len(nxt)
	pct := float64(0)
	if total > 0 {
		pct = float64(len(comp)) / float64(total) * 100.0
	}

	res := fmt.Sprintf("🔄 %s [%s]\nLast Update: %s\n\nProgress: %.1f%% (%d/%d steps)\n\n## ✅ Completed\n", taskID, strings.ToUpper(status), updatedAt, pct, len(comp), total)
	for _, s := range comp {
		res += fmt.Sprintf("- [x] %s\n", s)
	}

	res += "\n## ⏳ Next\n"
	for _, s := range nxt {
		res += fmt.Sprintf("- [ ] %s\n", s)
	}

	if ac.Valid && ac.String != "" {
		res += fmt.Sprintf("\n## 🎯 Acceptance Criteria\n%s\n", ac.String)
	}

	res += fmt.Sprintf("\n## 📝 Notes\n%s", notes)
	
	// Auto-Linked Context
	res += FetchAutoLinksContext(db, desc+"\n"+notes+"\n"+ac.String)

	return res, nil
}

func stringsToUpper(s string) string {
	return strings.ToUpper(s)
}

func ListActiveTasks(workspacePath string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	rows, err := db.Query("SELECT task_id, status, updated_at FROM tasks ORDER BY updated_at DESC")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var res []string
	res = append(res, "📋 WORKSPACE TASKS:")
	count := 0
	for rows.Next() {
		var tID, stat, updated string
		if err := rows.Scan(&tID, &stat, &updated); err == nil {
			res = append(res, fmt.Sprintf("- **%s** (%s) - %s", tID, stat, updated))
			count++
		}
	}

	if count == 0 {
		return "No tasks found.", nil
	}

	return strings.Join(res, "\n"), nil
}

// FindRecentTask performs fuzzy keyword search across checkpoint descriptions.
// Returns up to 3 matches with task_id, status, progress%, and description.
// Improvement #4: Smart session continuity without needing exact task IDs.
func FindRecentTask(workspacePath, keywords string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Build LIKE pattern — safe parameterized query, escape literal %
	pattern := "%" + strings.ReplaceAll(keywords, "%", "\\%") + "%"

	rows, err := db.Query(
		`SELECT t.task_id, t.description, t.status, t.updated_at,
		 (SELECT COUNT(*) FROM steps WHERE task_id = t.task_id AND status = 'completed') as comp_count,
		 (SELECT COUNT(*) FROM steps WHERE task_id = t.task_id AND status != 'completed') as pending_count
         FROM tasks t
         WHERE t.description LIKE ? ESCAPE '\'
         ORDER BY t.updated_at DESC
         LIMIT 3`,
		pattern,
	)
	if err != nil {
		return "", fmt.Errorf("find_recent_task query failed: %v", err)
	}
	defer rows.Close()

	var results []string
	results = append(results, fmt.Sprintf("🔍 Tasks matching **\"%s\"**:", keywords))

	count := 0
	for rows.Next() {
		var tID, desc, stat, updated string
		var compCount, pendingCount int
		if err := rows.Scan(&tID, &desc, &stat, &updated, &compCount, &pendingCount); err != nil {
			continue
		}
		total := compCount + pendingCount
		pct := 0.0
		if total > 0 {
			pct = float64(compCount) / float64(total) * 100.0
		}
		results = append(results, fmt.Sprintf(
			"- **%s** [%s] %.0f%% (%d/%d) — %s\n  _Updated: %s_ → `load_checkpoint(task_id=\"%s\")`",
			tID, strings.ToUpper(stat), pct, compCount, total, desc, updated, tID,
		))
		count++
	}

	if count == 0 {
		return fmt.Sprintf("⚠️ No tasks found matching \"%s\".", keywords), nil
	}
	return strings.Join(results, "\n\n"), nil
}

// IdleTask holds the summary of a stale in_progress task for the incomplete tasks panel.
type IdleTask struct {
	TaskID      string
	Description string
	Progress    float64
	Done        int
	Total       int
	IdleDays    int
	LastUpdate  string
}

// fetchIdleTasks queries the DB for tasks that are still in_progress,
// have pending steps, and haven't been updated within idleThresholdDays.
// The current task (currentTaskID) is always excluded.
func fetchIdleTasks(workspacePath, currentTaskID string, idleThresholdDays int) ([]IdleTask, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	cutoff := time.Now().AddDate(0, 0, -idleThresholdDays).Format(time.RFC3339)

	rows, err := db.Query(`
		SELECT t.task_id, t.description, t.updated_at,
		 (SELECT COUNT(*) FROM steps WHERE task_id = t.task_id AND status = 'completed') as comp_count,
		 (SELECT COUNT(*) FROM steps WHERE task_id = t.task_id AND status != 'completed') as pending_count
		FROM tasks t
		WHERE LOWER(t.status) = 'in_progress'
		  AND t.task_id != ?
		  AND t.updated_at < ?
		ORDER BY t.updated_at DESC`,
		currentTaskID, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []IdleTask
	for rows.Next() {
		var tID, desc, updatedAt string
		var compCount, pendingCount int
		if err := rows.Scan(&tID, &desc, &updatedAt, &compCount, &pendingCount); err != nil {
			continue
		}

		if pendingCount == 0 {
			continue
		}

		total := compCount + pendingCount
		pct := 0.0
		if total > 0 {
			pct = float64(compCount) / float64(total) * 100.0
		}

		t, err := time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			t, _ = time.Parse("2006-01-02T15:04:05Z07:00", updatedAt)
		}
		idleDays := int(time.Since(t).Hours() / 24)

		shortDate := updatedAt
		if len(shortDate) >= 10 {
			shortDate = shortDate[:10]
		}

		results = append(results, IdleTask{
			TaskID:      tID,
			Description: desc,
			Progress:    pct,
			Done:        compCount,
			Total:       total,
			IdleDays:    idleDays,
			LastUpdate:  shortDate,
		})
	}
	return results, nil
}

// DeleteTask permanently removes a task checkpoint and its associated intent locks
// from the database, then refreshes progress.md so the task no longer appears.
// Returns an error if the task_id is not found.
func DeleteTask(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Verify the task exists before deleting
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE task_id = ?", taskID).Scan(&count); err != nil {
		return "", fmt.Errorf("failed to query task '%s': %v", taskID, err)
	}
	if count == 0 {
		return fmt.Sprintf("❌ Task '%s' not found. Nothing was deleted.", taskID), nil
	}

	// Remove associated intent locks first (foreign-key style cleanup)
	if _, err := db.Exec("DELETE FROM intent_locks WHERE task_id = ?", taskID); err != nil {
		// Non-fatal: intent_locks table may not exist on older DBs
		fmt.Fprintf(os.Stderr, "[context-manager] warning: could not remove intent_locks for '%s': %v\n", taskID, err)
	}

	// Remove the checkpoint itself
	res, err := db.Exec("DELETE FROM tasks WHERE task_id = ?", taskID)
	if err != nil {
		return "", fmt.Errorf("failed to delete task '%s': %v", taskID, err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Sprintf("⚠️ Task '%s' was not deleted (0 rows affected).", taskID), nil
	}

	// Refresh progress.md using a sentinel empty state so the deleted task
	// is no longer rendered (WriteMarkdownProgress queries live DB).
	_ = WriteMarkdownProgress(db, workspacePath, taskID)

	return fmt.Sprintf("🗑️ Task '%s' deleted successfully.", taskID), nil
}
