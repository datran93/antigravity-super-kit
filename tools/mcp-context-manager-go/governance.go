package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DeclareIntent locks a set of files for the current tactic.
// ttlMinutes: how long the lock is valid (0 = no expiry / legacy mode; default = 60).
func DeclareIntent(workspacePath, taskID, tactic string, lockedFiles []string, ttlMinutes int) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	lockedBytes, _ := json.Marshal(lockedFiles)

	var expiresAt int64
	if ttlMinutes > 0 {
		expiresAt = time.Now().Add(time.Duration(ttlMinutes) * time.Minute).Unix()
	}

	query := `
        INSERT INTO intents (task_id, tactic, locked_files, expires_at)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(task_id) DO UPDATE SET
            tactic=excluded.tactic,
            locked_files=excluded.locked_files,
            expires_at=excluded.expires_at,
            updated_at=CURRENT_TIMESTAMP
    `
	_, err = db.Exec(query, taskID, tactic, string(lockedBytes), expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to declare intent: %v", err)
	}

	ttlMsg := "no expiry"
	if expiresAt > 0 {
		ttlMsg = fmt.Sprintf("expires in %d min at %s", ttlMinutes, time.Unix(expiresAt, 0).Format("15:04:05"))
	}
	// T23: Auto-emit activity event
	logActivityDB(db, "intent_declared", taskID, fmt.Sprintf("Tactic: %s, Files: %s", tactic, strings.Join(lockedFiles, ", ")))

	return fmt.Sprintf("🔒 Intent declared. Lock applied to files: %s (%s)", strings.Join(lockedFiles, ", "), ttlMsg), nil
}

func CheckIntentLock(workspacePath, taskID, filePath string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var lockedFilesStr string
	var expiresAt int64
	err = db.QueryRow("SELECT locked_files, COALESCE(expires_at, 0) FROM intents WHERE task_id = ?", taskID).Scan(&lockedFilesStr, &expiresAt)
	if err != nil {
		return "⚠️ No intent declared for this task. Please declare_intent first.", nil
	}

	// TTL check — treat expired lock as if no lock was declared
	if expiresAt > 0 && time.Now().Unix() > expiresAt {
		return fmt.Sprintf("⚠️ EXPIRED LOCK: Intent lock for task '%s' expired at %s. Call declare_intent again to renew.",
			taskID, time.Unix(expiresAt, 0).Format("15:04:05")), nil
	}

	var lockedFiles []string
	if err := json.Unmarshal([]byte(lockedFilesStr), &lockedFiles); err != nil {
		return "", fmt.Errorf("failed to parse locked_files: %v", err)
	}

	isLocked := false
	for _, f := range lockedFiles {
		if strings.Contains(f, filePath) || strings.Contains(filePath, f) {
			isLocked = true
			break
		}
	}

	if isLocked {
		var gotchas string
		err := db.QueryRow("SELECT gotchas FROM file_annotations WHERE file_path = ?", filePath).Scan(&gotchas)
		ghostCtx := ""
		if err == nil && gotchas != "" {
			ghostCtx = fmt.Sprintf("\n👻 GHOST CONTEXT: %s", gotchas)
		}
		return fmt.Sprintf("✅ File '%s' is unlocked for current intent.%s", filePath, ghostCtx), nil
	}

	return fmt.Sprintf("❌ ALARM: Scope Creep! File '%s' is NOT in the active_files lock. Switch to Planner to update intent via declare_intent.", filePath), nil
}

func AnnotateFile(workspacePath, filePath, gotchas string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	query := `
        INSERT INTO file_annotations (file_path, gotchas)
        VALUES (?, ?)
        ON CONFLICT(file_path) DO UPDATE SET
            gotchas=excluded.gotchas,
            updated_at=CURRENT_TIMESTAMP
    `
	_, err = db.Exec(query, filePath, gotchas)
	if err != nil {
		return "", fmt.Errorf("failed to annotate file: %v", err)
	}

	return fmt.Sprintf("👻 Ghost Context added to '%s'.", filePath), nil
}

// RecordFailure increments the drift counter. stepName and errorContext are optional
// but recommended for Improvement #8 war-room KI generation.
func RecordFailure(workspacePath, taskID, stepName, errorContext string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	query := `
        INSERT INTO drift_tracker (task_id, failure_count, last_failed_at, step_name, error_context)
        VALUES (?, 1, CURRENT_TIMESTAMP, ?, ?)
        ON CONFLICT(task_id) DO UPDATE SET
            failure_count=failure_count + 1,
            last_failed_at=CURRENT_TIMESTAMP,
            step_name=excluded.step_name,
            error_context=excluded.error_context
    `
	if _, err := db.Exec(query, taskID, stepName, errorContext); err != nil {
		return "", fmt.Errorf("failed to record failure: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT failure_count FROM drift_tracker WHERE task_id=?", taskID).Scan(&count); err != nil {
		return "", err
	}

	if count >= 3 {
		warRoomMsg := fmt.Sprintf("🚨 DRIFT DETECTED (Failures: %d): You have failed 3 or more times", count)
		if stepName != "" {
			warRoomMsg += fmt.Sprintf(" on step '%s'", stepName)
		}
		warRoomMsg += ". Trigger 'think_back' panic protocol immediately!"
		if errorContext != "" {
			warRoomMsg += fmt.Sprintf("\n📋 Last error: %s", errorContext)
		}
		return warRoomMsg, nil
	}
	return fmt.Sprintf("⚠️ Failure recorded. Count: %d/3", count), nil
}

func ClearDrift(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if _, err := db.Exec("UPDATE drift_tracker SET failure_count=0 WHERE task_id=?", taskID); err != nil {
		return "", fmt.Errorf("failed to clear drift: %v", err)
	}

	return "🧹 Drift counter reset to 0.", nil
}

func ManageAnchors(workspacePath, scope, action, key, value, rule string) (string, error) {
	var db *sql.DB
	var err error
	tableName := "anchors"
	scopeName := "PROJECT"

	if scope == "global" {
		db, err = GetGlobalDBConnection()
		tableName = "global_anchors"
		scopeName = "GLOBAL"
	} else {
		db, err = GetDBConnection(workspacePath)
	}

	if err != nil {
		return "", err
	}
	defer db.Close()

	switch action {
	case "set":
		if key == "" || value == "" {
			return "❌ Key and value required for 'set'.", nil
		}
		query := fmt.Sprintf(`
            INSERT INTO %s (key, value, rule)
            VALUES (?, ?, ?)
            ON CONFLICT(key) DO UPDATE SET
                value=excluded.value,
                rule=excluded.rule,
                updated_at=CURRENT_TIMESTAMP
        `, tableName)
		if _, err := db.Exec(query, key, value, rule); err != nil {
			return "", err
		}
		return fmt.Sprintf("⚓ %s Anchor '%s' secured successfully.", scopeName, key), nil

	case "get":
		var val, r string
		err := db.QueryRow(fmt.Sprintf("SELECT value, rule FROM %s WHERE key = ?", tableName), key).Scan(&val, &r)
		if err != nil {
			return fmt.Sprintf("⚠️ %s Anchor '%s' not found.", scopeName, key), nil
		}
		return fmt.Sprintf("⚓ %s ANCHOR [%s]: %s (Rule: %s)", scopeName, key, val, r), nil

	case "list":
		rows, err := db.Query(fmt.Sprintf("SELECT key, value, rule FROM %s ORDER BY key", tableName))
		if err != nil {
			return "", err
		}
		defer rows.Close()

		var res []string
		res = append(res, fmt.Sprintf("⚓ %s ANCHORS:", scopeName))
		count := 0
		for rows.Next() {
			var k, v, r string
			if err := rows.Scan(&k, &v, &r); err == nil {
				res = append(res, fmt.Sprintf("- **%s**: %s", k, v))
				if r != "" {
					res = append(res, fmt.Sprintf("  Rule: %s", r))
				}
				count++
			}
		}
		if count == 0 {
			return "No anchors defined.", nil
		}
		return strings.Join(res, "\n"), nil

	default:
		return "❌ Unknown action. Use 'set', 'get', or 'list'.", nil
	}
}
