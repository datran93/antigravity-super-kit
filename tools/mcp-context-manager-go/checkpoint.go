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

type CheckpointRow struct {
	TaskID         string
	Description    string
	Status         string
	CompletedSteps string // JSON array
	NextSteps      string // JSON array
	ActiveFiles    string // JSON array
	Notes          string
	UpdatedAt      string
}

// phaseKey extracts the phase prefix from a step name.
// "[P0-T1] Implement X" → "P0", "[P1] Foo" → "P1", "bare step" → "".
func phaseKey(step string) string {
	if len(step) == 0 || step[0] != '[' {
		return ""
	}
	end := strings.Index(step, "]")
	if end < 0 {
		return ""
	}
	inner := step[1:end] // e.g. "P0-T1" or "P1"
	// Keep only the phase part (Px) before any dash
	dash := strings.Index(inner, "-")
	if dash >= 0 {
		return inner[:dash]
	}
	return inner
}

// phaseEntry holds all steps belonging to a single phase.
type phaseEntry struct {
	key       string
	completed []string
	pending   []string
}

// groupByPhase groups completed + pending steps into ordered phaseEntry slices.
// Order is determined by first appearance across the combined list.
func groupByPhase(completed, pending []string) []phaseEntry {
	order := []string{}
	seen := map[string]bool{}
	phases := map[string]*phaseEntry{}

	add := func(step string, done bool) {
		k := phaseKey(step)
		if !seen[k] {
			seen[k] = true
			order = append(order, k)
			phases[k] = &phaseEntry{key: k}
		}
		if done {
			phases[k].completed = append(phases[k].completed, step)
		} else {
			phases[k].pending = append(phases[k].pending, step)
		}
	}

	for _, s := range completed {
		add(s, true)
	}
	for _, s := range pending {
		add(s, false)
	}

	result := make([]phaseEntry, 0, len(order))
	for _, k := range order {
		result = append(result, *phases[k])
	}
	return result
}

// phaseIcon returns the status icon for a phase.
func phaseIcon(p phaseEntry) string {
	if len(p.pending) == 0 && len(p.completed) > 0 {
		return "✅"
	}
	if len(p.completed) > 0 {
		return "🔄"
	}
	return "🔲"
}

// isPhaseComplete returns true when all steps in p are done.
func isPhaseComplete(p phaseEntry) bool {
	return len(p.pending) == 0 && len(p.completed) > 0
}

func WriteMarkdownProgress(db *sql.DB, workspacePath, taskID, description, status string, completedSteps, nextSteps, activeFiles []string, notes string) error {
	mdPath := filepath.Join(workspacePath, "progress.md")

	// Historical completed tasks (other task_ids)
	rows, err := db.Query("SELECT task_id, description FROM checkpoints WHERE status = 'completed' AND task_id != ? ORDER BY updated_at DESC", taskID)
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

	if len(activeFiles) > 0 {
		content += "### 📁 Active Files\n"
		for _, f := range activeFiles {
			content += fmt.Sprintf("- `%s`\n", f)
		}
		content += "\n"
	}

	// ── Phase-aware rendering ──────────────────────────────────────────────
	phases := groupByPhase(completedSteps, nextSteps)
	hasPhases := len(phases) > 0 && phases[0].key != ""

	if hasPhases {
		// Phase summary table
		content += "### 📊 Phase Overview\n\n"
		content += "| Phase | Done | Todo | Status |\n"
		content += "|-------|------|------|--------|\n"
		for _, p := range phases {
			label := p.key
			if label == "" {
				label = "General"
			}
			content += fmt.Sprintf("| **%s** | %d | %d | %s |\n",
				label, len(p.completed), len(p.pending), phaseIcon(p))
		}
		content += "\n"

		// Per-phase step lists
		for _, p := range phases {
			label := p.key
			if label == "" {
				label = "General"
			}
			content += fmt.Sprintf("#### %s %s\n", phaseIcon(p), label)
			for _, s := range p.completed {
				content += fmt.Sprintf("- [x] %s\n", s)
			}
			for _, s := range p.pending {
				content += fmt.Sprintf("- [ ] %s\n", s)
			}
			content += "\n"
		}
	} else {
		// Fallback: flat rendering for tasks without phase prefixes
		content += "### ✅ Completed\n"
		if len(completedSteps) == 0 {
			content += "_None yet_\n"
		} else {
			for _, s := range completedSteps {
				content += fmt.Sprintf("- [x] %s\n", s)
			}
		}
		content += "\n### ⏳ Next Steps\n"
		if len(nextSteps) == 0 {
			content += "_All tasks done!_ 🎉\n"
		} else {
			for _, s := range nextSteps {
				content += fmt.Sprintf("- [ ] %s\n", s)
			}
		}
		content += "\n"
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

	content += fmt.Sprintf("---\n*Last sync: %s*", time.Now().Format("2006-01-02 15:04:05"))
	return os.WriteFile(mdPath, []byte(content), 0644)
}

func SaveCheckpoint(workspacePath, taskID, description, status string, completedSteps, nextSteps, activeFiles []string, notes string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	now := time.Now().Format(time.RFC3339)

	compBytes, _ := json.Marshal(completedSteps)
	nextBytes, _ := json.Marshal(nextSteps)
	activeBytes, _ := json.Marshal(activeFiles)

	query := `
        INSERT INTO checkpoints (task_id, description, status, completed_steps, next_steps, active_files, notes, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(task_id) DO UPDATE SET
            description=excluded.description,
            status=excluded.status,
            completed_steps=excluded.completed_steps,
            next_steps=excluded.next_steps,
            active_files=excluded.active_files,
            notes=excluded.notes,
            updated_at=excluded.updated_at
    `
	_, err = db.Exec(query, taskID, description, status, string(compBytes), string(nextBytes), string(activeBytes), notes, now)
	if err != nil {
		return "", fmt.Errorf("failed to save checkpoint table: %v", err)
	}

	if err := WriteMarkdownProgress(db, workspacePath, taskID, description, status, completedSteps, nextSteps, activeFiles, notes); err != nil {
		fmt.Printf("Error writing progress.md: %v\n", err)
	}

	// ── Phase completion hint ──────────────────────────────────────────────
	// Detect if the last completed step just finished a phase → emit reminder.
	phases := groupByPhase(completedSteps, nextSteps)
	phaseHint := ""
	if len(completedSteps) > 0 {
		lastStep := completedSteps[len(completedSteps)-1]
		lastKey := phaseKey(lastStep)
		if lastKey != "" {
			for _, p := range phases {
				if p.key == lastKey && isPhaseComplete(p) {
					phaseHint = fmt.Sprintf(
						"\n\n💡 Phase **%s** is complete — run `/compact-session` to persist a KI before starting the next phase.",
						lastKey,
					)
					break
				}
			}
		}
	}

	msg := fmt.Sprintf("✅ Checkpoint '%s' saved.", taskID)
	if len(nextSteps) == 0 && len(completedSteps) > 0 {
		msg += "\n\n🎉 ALL TASKS COMPLETED! Great job."
	}
	msg += phaseHint
	return msg, nil
}
