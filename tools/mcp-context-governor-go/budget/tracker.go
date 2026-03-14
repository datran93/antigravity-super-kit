// Package budget provides session-level token usage tracking for the Context Budget Governor.
package budget

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Tracker manages per-session token usage in a SQLite database.
type Tracker struct {
	db        *sql.DB
	SessionID string
}

// TokenEvent represents a single token usage record.
type TokenEvent struct {
	ID        int64
	SessionID string
	Tool      string
	Tokens    int
	Source    string // "estimate" | "actual"
	CreatedAt time.Time
}

// SessionSummary is the aggregated usage for a session.
type SessionSummary struct {
	SessionID   string
	TotalTokens int
	EventCount  int
	StartedAt   time.Time
	LastEventAt time.Time
}

// OpenTracker opens (or creates) the governor database in dataDir.
func OpenTracker(dataDir, sessionID string) (*Tracker, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}
	dbPath := filepath.Join(dataDir, "governor.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open governor db: %w", err)
	}
	t := &Tracker{db: db, SessionID: sessionID}
	if err := t.initSchema(); err != nil {
		db.Close()
		return nil, err
	}
	return t, nil
}

// Close releases the database connection.
func (t *Tracker) Close() error {
	return t.db.Close()
}

func (t *Tracker) initSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS token_events (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT    NOT NULL,
			tool       TEXT    NOT NULL DEFAULT '',
			tokens     INTEGER NOT NULL,
			source     TEXT    NOT NULL DEFAULT 'estimate',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_session ON token_events(session_id)`,
		`CREATE TABLE IF NOT EXISTS session_meta (
			session_id   TEXT    PRIMARY KEY,
			max_budget   INTEGER NOT NULL DEFAULT 100000,
			started_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_event_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, s := range stmts {
		if _, err := t.db.Exec(s); err != nil {
			return fmt.Errorf("schema init error: %w\nSQL: %s", err, s)
		}
	}
	// Ensure session_meta row exists — log any unexpected error for visibility.
	if _, err := t.db.Exec(`INSERT OR IGNORE INTO session_meta (session_id) VALUES (?)`, t.SessionID); err != nil {
		fmt.Fprintf(os.Stderr, "[context-governor] warning: failed to init session_meta for %q: %v\n", t.SessionID, err)
	}
	return nil
}

// RecordUsage adds a token usage event to the tracker.
func (t *Tracker) RecordUsage(tool string, tokens int, source string) error {
	if source == "" {
		source = "estimate"
	}
	_, err := t.db.Exec(`
		INSERT INTO token_events (session_id, tool, tokens, source) VALUES (?, ?, ?, ?)
	`, t.SessionID, tool, tokens, source)
	if err != nil {
		return err
	}
	t.db.Exec(`UPDATE session_meta SET last_event_at=CURRENT_TIMESTAMP WHERE session_id=?`, t.SessionID)
	return nil
}

// GetSummary returns aggregated usage for the current session.
func (t *Tracker) GetSummary() (SessionSummary, error) {
	var s SessionSummary
	s.SessionID = t.SessionID

	err := t.db.QueryRow(`
		SELECT COALESCE(SUM(tokens),0), COUNT(*) FROM token_events WHERE session_id=?
	`, t.SessionID).Scan(&s.TotalTokens, &s.EventCount)
	if err != nil {
		return s, err
	}

	var startedAt, lastAt string
	t.db.QueryRow(`SELECT started_at, last_event_at FROM session_meta WHERE session_id=?`,
		t.SessionID).Scan(&startedAt, &lastAt)
	s.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
	s.LastEventAt, _ = time.Parse("2006-01-02 15:04:05", lastAt)

	return s, nil
}

// GetMaxBudget returns the configured token budget for the session.
// Resolution order (highest → lowest priority):
//  1. MAX_BUDGET_TOKENS env var (e.g. 200000 for Claude, 1048576 for Gemini Flash)
//  2. max_budget column in session_meta (set via SetMaxBudget)
//  3. Default: 100 000 tokens
func (t *Tracker) GetMaxBudget() int {
	// 1. Env var takes precedence — allows per-model tuning without DB writes
	if raw := os.Getenv("MAX_BUDGET_TOKENS"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			return v
		}
	}
	// 2. DB-persisted value (set via SetMaxBudget or future MCP tool)
	var max int
	t.db.QueryRow(`SELECT max_budget FROM session_meta WHERE session_id=?`, t.SessionID).Scan(&max)
	if max > 0 {
		return max
	}
	// 3. Hardcoded fallback
	return 100_000
}

// SetMaxBudget updates the max token budget for the session.
func (t *Tracker) SetMaxBudget(maxTokens int) error {
	_, err := t.db.Exec(`UPDATE session_meta SET max_budget=? WHERE session_id=?`, maxTokens, t.SessionID)
	return err
}

// ResetSession clears all events for the current session.
func (t *Tracker) ResetSession() error {
	tx, _ := t.db.Begin()
	defer tx.Rollback()
	tx.Exec(`DELETE FROM token_events WHERE session_id=?`, t.SessionID)
	tx.Exec(`UPDATE session_meta SET last_event_at=CURRENT_TIMESTAMP WHERE session_id=?`, t.SessionID)
	return tx.Commit()
}
