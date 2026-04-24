package main

import (
	"database/sql"
	"fmt"
	"strings"
)

// LogActivity records an activity event in the audit trail.
func LogActivity(workspacePath, eventType, taskID, detail string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return logActivityDB(db, eventType, taskID, detail)
}

// logActivityDB is the internal variant that accepts an existing *sql.DB connection.
// Used by other handlers to emit events without opening a new connection.
func logActivityDB(db *sql.DB, eventType, taskID, detail string) (string, error) {
	if eventType == "" {
		return "", fmt.Errorf("event_type is required")
	}

	_, err := db.Exec(
		"INSERT INTO activity_events (event_type, task_id, detail) VALUES (?, ?, ?)",
		eventType, taskID, detail,
	)
	if err != nil {
		return "", fmt.Errorf("failed to log activity: %w", err)
	}
	return fmt.Sprintf("📝 Activity logged: [%s] %s", eventType, truncateActivity(detail, 100)), nil
}

// ListActivity retrieves recent activity events, optionally filtered by event_type or task_id.
func ListActivity(workspacePath, eventType, taskID string, limit int) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if limit <= 0 {
		limit = 20
	}

	var conditions []string
	var args []interface{}

	if eventType != "" {
		conditions = append(conditions, "event_type = ?")
		args = append(args, eventType)
	}
	if taskID != "" {
		conditions = append(conditions, "task_id = ?")
		args = append(args, taskID)
	}

	query := "SELECT id, event_type, task_id, detail, created_at FROM activity_events"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to list activity: %w", err)
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📋 Recent Activity:\n\n")

	count := 0
	for rows.Next() {
		var id int
		var evType, createdAt string
		var evTaskID, detail sql.NullString
		if err := rows.Scan(&id, &evType, &evTaskID, &detail, &createdAt); err != nil {
			continue
		}

		taskStr := ""
		if evTaskID.Valid && evTaskID.String != "" {
			taskStr = fmt.Sprintf(" [%s]", evTaskID.String)
		}
		detailStr := ""
		if detail.Valid && detail.String != "" {
			detailStr = fmt.Sprintf(": %s", detail.String)
		}

		sb.WriteString(fmt.Sprintf("- `%s` **%s**%s%s\n", createdAt, evType, taskStr, detailStr))
		count++
	}

	if count == 0 {
		return "📋 No activity events found.", nil
	}

	return sb.String(), nil
}

// truncateActivity shortens a string for display.
func truncateActivity(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
