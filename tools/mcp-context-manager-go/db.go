package main

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// embeddingDim is the fixed dimension of OpenAI text-embedding-3-small vectors.
const embeddingDim = 1536

// GetDBConnection opens a connection to the SQLite database and initializes the schema.
func GetDBConnection(workspacePath string) (*sql.DB, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace_path is required")
	}

	absPath, err := filepath.Abs(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace path: %v", err)
	}

	dbPath := filepath.Join(absPath, "context.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := initializeSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// columnExists checks if a column exists in a given table.
// SAFETY: tableName MUST be a hardcoded constant (e.g., "checkpoints", "intents").
// Never pass user-controlled input — PRAGMA does not support parameterized table names.
func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, fmt.Errorf("failed to query table info for %s: %w", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt_value sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return false, fmt.Errorf("failed to scan table info row for %s: %w", tableName, err)
		}
		if name == columnName {
			return true, nil
		}
	}
	return false, nil
}

// addColumnIfNotExist adds a column to a table if it doesn't already exist.
func addColumnIfNotExist(db *sql.DB, tableName, columnName, columnType string) error {
	exists, err := columnExists(db, tableName, columnName)
	if err != nil {
		return err
	}
	if !exists {
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
		if _, err := db.Exec(alterQuery); err != nil {
			return fmt.Errorf("failed to add column %s to table %s: %w", columnName, tableName, err)
		}
	}
	return nil
}

func initializeSchema(db *sql.DB) error {
	// Initial table creation queries (idempotent with IF NOT EXISTS)
	createTableQueries := []string{
		`CREATE TABLE IF NOT EXISTS checkpoints (
            task_id TEXT PRIMARY KEY,
            description TEXT,
            status TEXT,
            completed_steps TEXT,
            next_steps TEXT,
            active_files TEXT,
            notes TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS intents (
            task_id TEXT PRIMARY KEY,
            tactic TEXT,
            locked_files TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS drift_tracker (
            task_id TEXT PRIMARY KEY,
            failure_count INTEGER DEFAULT 0,
            last_failed_at TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS anchors (
            key TEXT PRIMARY KEY,
            value TEXT,
            rule TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS file_annotations (
            file_path TEXT PRIMARY KEY,
            gotchas TEXT,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_fts USING fts5(
            tactic_name,
            ki_path UNINDEXED,
            summary,
            decisions
        )`,
		// ki_embeddings stores float32 vectors as JSON blobs for cosine-similarity
		// hybrid recall. Kept separate from knowledge_fts for additive-only migration.
		`CREATE TABLE IF NOT EXISTS ki_embeddings (
            ki_path  TEXT PRIMARY KEY,
            tactic   TEXT NOT NULL,
            embedding TEXT NOT NULL
        )`,
	}

	for _, query := range createTableQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to initialize schema query:\n%s\nerror: %v", query, err)
		}
	}

	// Schema migrations for existing tables (v3 changes)
	// checkpoints: git SHA + burndown columns
	for _, m := range []struct{ col, def string }{
		{"git_sha", "TEXT DEFAULT ''"},
		{"step_timestamps", "TEXT DEFAULT '{}'"},
		{"step_drift", "TEXT DEFAULT '{}'"},
		{"step_deps", "TEXT DEFAULT '{}'"},
	} {
		if err := addColumnIfNotExist(db, "checkpoints", m.col, m.def); err != nil {
			return fmt.Errorf("checkpoints.%s migration: %w", m.col, err)
		}
	}

	// intents: TTL support (Improvement #2)
	if err := addColumnIfNotExist(db, "intents", "expires_at", "INTEGER DEFAULT 0"); err != nil {
		return fmt.Errorf("intents.expires_at migration: %w", err)
	}

	// drift_tracker: per-step war-room context (Improvement #8)
	for _, m := range []struct{ col, def string }{
		{"step_name", "TEXT DEFAULT ''"},
		{"error_context", "TEXT DEFAULT ''"},
	} {
		if err := addColumnIfNotExist(db, "drift_tracker", m.col, m.def); err != nil {
			return fmt.Errorf("drift_tracker.%s migration: %w", m.col, err)
		}
	}

	return nil
}
