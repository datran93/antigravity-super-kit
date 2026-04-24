package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SessionMemoryEntry represents a single ephemeral memory item.
type SessionMemoryEntry struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Category  string `json:"category"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ManageSessionMemory dispatches to the correct action for session memory management.
// Supported actions: add, list, promote, clear.
func ManageSessionMemory(workspacePath, action, sessionID, category, content string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	switch strings.ToLower(action) {
	case "add":
		return addSessionMemory(db, sessionID, category, content)
	case "list":
		return listSessionMemories(db, sessionID, category)
	case "promote":
		return promoteSessionMemory(db, workspacePath, content)
	case "clear":
		return clearSessionMemories(db, sessionID)
	default:
		return "", fmt.Errorf("unknown action '%s' — use: add, list, promote, clear", action)
	}
}

// addSessionMemory inserts a new ephemeral memory entry.
func addSessionMemory(db *sql.DB, sessionID, category, content string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session_id is required")
	}
	if content == "" {
		return "", fmt.Errorf("content is required")
	}
	if category == "" {
		category = "finding"
	}

	id := uuid.New().String()[:8]

	_, err := db.Exec(
		"INSERT INTO session_memories (id, session_id, category, content) VALUES (?, ?, ?, ?)",
		id, sessionID, category, content,
	)
	if err != nil {
		return "", fmt.Errorf("failed to add session memory: %w", err)
	}
	return fmt.Sprintf("🧠 Session memory added [%s] (%s): %s", id, category, truncate(content, 80)), nil
}

// listSessionMemories returns all memories for a session, optionally filtered by category.
func listSessionMemories(db *sql.DB, sessionID, category string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session_id is required")
	}

	var rows *sql.Rows
	var err error
	if category != "" {
		rows, err = db.Query(
			"SELECT id, category, content, created_at FROM session_memories WHERE session_id = ? AND category = ? ORDER BY created_at ASC",
			sessionID, category,
		)
	} else {
		rows, err = db.Query(
			"SELECT id, category, content, created_at FROM session_memories WHERE session_id = ? ORDER BY created_at ASC",
			sessionID,
		)
	}
	if err != nil {
		return "", fmt.Errorf("failed to list session memories: %w", err)
	}
	defer rows.Close()

	var entries []SessionMemoryEntry
	for rows.Next() {
		var e SessionMemoryEntry
		if err := rows.Scan(&e.ID, &e.Category, &e.Content, &e.CreatedAt); err != nil {
			continue
		}
		e.SessionID = sessionID
		entries = append(entries, e)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("🧠 No session memories found for session '%s'.", sessionID), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🧠 Session memories for '%s' (%d items):\n\n", sessionID, len(entries)))
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("- **[%s]** (%s) %s\n", e.ID, e.Category, e.Content))
	}
	return sb.String(), nil
}

// promoteSessionMemory converts a session memory ID to a KI via compact_memory.
// The `content` parameter is the session memory ID to promote.
func promoteSessionMemory(db *sql.DB, workspacePath, memoryID string) (string, error) {
	if memoryID == "" {
		return "", fmt.Errorf("memory ID is required for promotion")
	}

	var category, content string
	err := db.QueryRow(
		"SELECT category, content FROM session_memories WHERE id = ?", memoryID,
	).Scan(&category, &content)
	if err == sql.ErrNoRows {
		return fmt.Sprintf("⚠️ Session memory '%s' not found.", memoryID), nil
	} else if err != nil {
		return "", err
	}

	// Promote to KI
	tacticName := fmt.Sprintf("promoted_%s_%s", category, time.Now().Format("20060102_150405"))
	result, err := CompactMemory(workspacePath, "session-promote", tacticName, content, fmt.Sprintf("Promoted from session memory [%s]", memoryID))
	if err != nil {
		return "", fmt.Errorf("failed to promote session memory: %w", err)
	}

	// Remove the promoted memory
	db.Exec("DELETE FROM session_memories WHERE id = ?", memoryID)

	return fmt.Sprintf("🎓 Session memory '%s' promoted to KI.\n%s", memoryID, result), nil
}

// clearSessionMemories deletes all memories for a given session.
func clearSessionMemories(db *sql.DB, sessionID string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session_id is required")
	}

	res, err := db.Exec("DELETE FROM session_memories WHERE session_id = ?", sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to clear session memories: %w", err)
	}
	affected, _ := res.RowsAffected()
	return fmt.Sprintf("🧹 Cleared %d session memories for session '%s'.", affected, sessionID), nil
}

// PurgeStaleSessionMemories removes session memories older than 24 hours.
// Called during DB initialization for automatic cleanup.
func PurgeStaleSessionMemories(db *sql.DB) {
	cutoff := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	db.Exec("DELETE FROM session_memories WHERE created_at < ?", cutoff)
}

// truncate shortens a string to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// sessionMemoriesToJSON is a helper for API consumers that need JSON output.
func sessionMemoriesToJSON(entries []SessionMemoryEntry) string {
	data, err := json.Marshal(entries)
	if err != nil {
		return "[]"
	}
	return string(data)
}
