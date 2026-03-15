package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func InitializeTaskPlan(workspacePath, taskID, description string, steps []string) (string, error) {
	initNotes := fmt.Sprintf("[%s] Task started.", time.Now().Format("15:04:05"))
	return SaveCheckpoint(
		workspacePath,
		taskID,
		description,
		"in_progress",
		[]string{},
		steps,
		[]string{},
		initNotes,
	)
}

func CompleteTaskStep(workspacePath, taskID, stepName string, activeFiles []string, notes string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	row := db.QueryRow(
		"SELECT description, status, completed_steps, next_steps, active_files, notes, COALESCE(step_timestamps,'{}'), COALESCE(step_drift,'{}') FROM checkpoints WHERE task_id = ?",
		taskID,
	)
	var description, status, completedStepsStr, nextStepsStr, activeFilesStr, currentNotes, stepTimestampsStr, stepDriftStr string
	if err := row.Scan(&description, &status, &completedStepsStr, &nextStepsStr, &activeFilesStr, &currentNotes, &stepTimestampsStr, &stepDriftStr); err != nil {
		return fmt.Sprintf("❌ Task '%s' not found.", taskID), nil
	}

	var comp []string
	var nxt []string
	var currActiveFiles []string
	json.Unmarshal([]byte(completedStepsStr), &comp)
	json.Unmarshal([]byte(nextStepsStr), &nxt)
	if activeFilesStr != "" {
		json.Unmarshal([]byte(activeFilesStr), &currActiveFiles)
	}

	// Load existing step metadata
	timestamps := ParseStepTimestamps(stepTimestampsStr)
	drift := ParseStepDrift(stepDriftStr)

	stepFound := false
	var newNxt []string
	for _, s := range nxt {
		if s == stepName {
			stepFound = true
			comp = append(comp, s)
		} else {
			newNxt = append(newNxt, s)
		}
	}

	if !stepFound {
		return fmt.Sprintf("⚠️ Step '%s' not in queue.", stepName), nil
	}

	// Record completion timestamp for velocity calculation
	timestamps[stepName] = time.Now().Format(time.RFC3339)

	stat := "completed"
	if len(newNxt) > 0 {
		stat = status
	}

	endTimeStr := time.Now().Format("15:04:05")
	log := currentNotes + fmt.Sprintf("\n[%s] ✅ Done: %s", endTimeStr, stepName)

	if len(activeFiles) > 0 {
		var activeFilesLog []string
		for _, f := range activeFiles {
			activeFilesLog = append(activeFilesLog, f)
			found := false
			for _, cf := range currActiveFiles {
				if cf == f {
					found = true
					break
				}
			}
			if !found {
				currActiveFiles = append(currActiveFiles, f)
			}
		}
		fromJSON, _ := json.Marshal(activeFilesLog)
		log += fmt.Sprintf("\n  - Files: %s", string(fromJSON))
	}

	if notes != "" {
		log += fmt.Sprintf("\n  - Notes: %s", notes)
	}

	// Improvement #5: auto-annotate active files when notes contain key phrases
	if hasGotchaKeyword(notes) {
		for _, f := range activeFiles {
			AnnotateFile(workspacePath, f, fmt.Sprintf("[%s] %s", stepName, notes))
		}
	}

	// Serialize updated metadata back to JSON
	timestampsBytes, _ := json.Marshal(timestamps)
	driftBytes, _ := json.Marshal(drift)

	// Persist step_timestamps + step_drift before the full SaveCheckpoint.
	// Log on failure (e.g., schema mismatch on old DB) — non-fatal but visible.
	if _, execErr := db.Exec(
		"UPDATE checkpoints SET step_timestamps=?, step_drift=? WHERE task_id=?",
		string(timestampsBytes), string(driftBytes), taskID,
	); execErr != nil {
		fmt.Printf("Warning: failed to persist step metadata for task '%s': %v\n", taskID, execErr)
	}

	return SaveCheckpoint(workspacePath, taskID, description, stat, comp, newNxt, currActiveFiles, log)
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

	row := db.QueryRow("SELECT description, status, completed_steps, next_steps, active_files, notes FROM checkpoints WHERE task_id = ?", taskID)
	var description, status, completedStepsStr, nextStepsStr, activeFilesStr, currentNotes string
	if err := row.Scan(&description, &status, &completedStepsStr, &nextStepsStr, &activeFilesStr, &currentNotes); err != nil {
		return fmt.Sprintf("❌ Task '%s' not found.", taskID), nil
	}

	var comp []string
	var nxt []string
	var currActiveFiles []string
	json.Unmarshal([]byte(completedStepsStr), &comp)
	json.Unmarshal([]byte(nextStepsStr), &nxt)
	if activeFilesStr != "" {
		json.Unmarshal([]byte(activeFilesStr), &currActiveFiles)
	}

	for _, s := range nxt {
		if s == newStep {
			return fmt.Sprintf("⚠️ Step '%s' already exists in task '%s'.", newStep, taskID), nil
		}
	}
	for _, s := range comp {
		if s == newStep {
			return fmt.Sprintf("⚠️ Step '%s' already exists in task '%s'.", newStep, taskID), nil
		}
	}

	nxt = append(nxt, newStep)
	log := currentNotes + fmt.Sprintf("\n[%s] Added new step: %s", time.Now().Format("15:04:05"), newStep)

	return SaveCheckpoint(workspacePath, taskID, description, status, comp, nxt, currActiveFiles, log)
}

func LoadCheckpoint(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	row := db.QueryRow("SELECT status, updated_at, completed_steps, next_steps, notes FROM checkpoints WHERE task_id = ?", taskID)
	var status, updatedAt, completedStepsStr, nextStepsStr, notes string
	if err := row.Scan(&status, &updatedAt, &completedStepsStr, &nextStepsStr, &notes); err != nil {
		return fmt.Sprintf("❌ Checkpoint '%s' not found.", taskID), nil
	}

	var comp []string
	var nxt []string
	json.Unmarshal([]byte(completedStepsStr), &comp)
	json.Unmarshal([]byte(nextStepsStr), &nxt)

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

	res += fmt.Sprintf("\n## 📝 Notes\n%s", notes)
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

	rows, err := db.Query("SELECT task_id, status, updated_at FROM checkpoints ORDER BY updated_at DESC")
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
		`SELECT task_id, description, status, completed_steps, next_steps, updated_at
         FROM checkpoints
         WHERE description LIKE ? ESCAPE '\'
         ORDER BY updated_at DESC
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
		var tID, desc, stat, compStr, nxtStr, updated string
		if err := rows.Scan(&tID, &desc, &stat, &compStr, &nxtStr, &updated); err != nil {
			continue
		}
		var comp, nxt []string
		json.Unmarshal([]byte(compStr), &comp)
		json.Unmarshal([]byte(nxtStr), &nxt)
		total := len(comp) + len(nxt)
		pct := 0.0
		if total > 0 {
			pct = float64(len(comp)) / float64(total) * 100.0
		}
		results = append(results, fmt.Sprintf(
			"- **%s** [%s] %.0f%% (%d/%d) — %s\n  _Updated: %s_ → `load_checkpoint(task_id=\"%s\")`",
			tID, strings.ToUpper(stat), pct, len(comp), total, desc, updated, tID,
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
		SELECT task_id, description, completed_steps, next_steps, updated_at
		FROM checkpoints
		WHERE LOWER(status) = 'in_progress'
		  AND task_id != ?
		  AND updated_at < ?
		ORDER BY updated_at DESC`,
		currentTaskID, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []IdleTask
	for rows.Next() {
		var tID, desc, compStr, nxtStr, updatedAt string
		if err := rows.Scan(&tID, &desc, &compStr, &nxtStr, &updatedAt); err != nil {
			continue
		}

		var comp, nxt []string
		json.Unmarshal([]byte(compStr), &comp)
		json.Unmarshal([]byte(nxtStr), &nxt)

		// Skip tasks with no remaining work
		if len(nxt) == 0 {
			continue
		}

		total := len(comp) + len(nxt)
		pct := 0.0
		if total > 0 {
			pct = float64(len(comp)) / float64(total) * 100.0
		}

		// Compute idle days from updated_at
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
			Done:        len(comp),
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
	if err := db.QueryRow("SELECT COUNT(*) FROM checkpoints WHERE task_id = ?", taskID).Scan(&count); err != nil {
		return "", fmt.Errorf("failed to query task '%s': %v", taskID, err)
	}
	if count == 0 {
		return fmt.Sprintf("❌ Task '%s' not found. Nothing was deleted.", taskID), nil
	}

	// Remove associated intent locks first (foreign-key style cleanup)
	if _, err := db.Exec("DELETE FROM intent_locks WHERE task_id = ?", taskID); err != nil {
		// Non-fatal: intent_locks table may not exist on older DBs
		fmt.Printf("Warning: could not remove intent_locks for '%s': %v\n", taskID, err)
	}

	// Remove the checkpoint itself
	res, err := db.Exec("DELETE FROM checkpoints WHERE task_id = ?", taskID)
	if err != nil {
		return "", fmt.Errorf("failed to delete task '%s': %v", taskID, err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Sprintf("⚠️ Task '%s' was not deleted (0 rows affected).", taskID), nil
	}

	// Refresh progress.md using a sentinel empty state so the deleted task
	// is no longer rendered (WriteMarkdownProgress queries live DB).
	_ = WriteMarkdownProgress(db, workspacePath, taskID, "[deleted]", "deleted",
		nil, nil, nil, "", "", 0)

	return fmt.Sprintf("🗑️ Task '%s' deleted successfully.", taskID), nil
}
