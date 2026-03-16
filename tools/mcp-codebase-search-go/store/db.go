// Package store manages SQLite persistence for the codebase search index.
package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps a SQLite connection for the codebase search index.
type DB struct {
	conn        *sql.DB
	ProjectPath string
}

// projectDBPath computes a stable per-project SQLite path.
// Uses sha256(absProjectPath) to avoid filesystem collisions.
func projectDBPath(projectPath, dataDir string) string {
	h := sha256.Sum256([]byte(projectPath))
	name := hex.EncodeToString(h[:16]) // first 16 bytes = 32 hex chars
	return filepath.Join(dataDir, name+".db")
}

// Open opens (or creates) the SQLite database for a given project.
// dataDir is the directory where per-project DBs are stored.
func Open(projectPath, dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}
	dbPath := projectDBPath(projectPath, dataDir)
	conn, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}
	db := &DB{conn: conn, ProjectPath: projectPath}
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, err
	}
	return db, nil
}

// Close releases the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) initSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS project_meta (
			project_path TEXT PRIMARY KEY,
			indexed_at   DATETIME,
			total_chunks INTEGER DEFAULT 0,
			merkle_root  TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS code_chunks (
			id           TEXT PRIMARY KEY,
			project_path TEXT NOT NULL,
			file_path    TEXT NOT NULL,
			rel_path     TEXT NOT NULL,
			lang         TEXT NOT NULL,
			symbol_name  TEXT,
			symbol_kind  TEXT,
			content      TEXT NOT NULL,
			line_start   INTEGER,
			line_end     INTEGER,
			file_hash    TEXT NOT NULL
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
			id UNINDEXED,
			content,
			symbol_name,
			rel_path,
			tokenize='porter ascii'
		)`,
		`CREATE TABLE IF NOT EXISTS chunk_embeddings (
			id        TEXT PRIMARY KEY,
			embedding TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_project ON code_chunks(project_path)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_file    ON code_chunks(file_path)`,
	}
	for _, s := range stmts {
		if _, err := d.conn.Exec(s); err != nil {
			// Detect FTS5 missing — this means the binary was built without the required flags.
			// Fix: cd tools/mcp-codebase-search-go && make build
			if strings.Contains(err.Error(), "no such module: fts5") {
				return fmt.Errorf(
					"FTS5 extension not available in this SQLite build.\n"+
						"The binary must be compiled with:\n"+
						"  CGO_CFLAGS=\"-DSQLITE_ENABLE_FTS5\" go build -tags fts5 -o mcp-codebase-search-go .\n"+
						"Or simply run: cd tools/mcp-codebase-search-go && make build\n"+
						"Original error: %w", err)
			}
			return fmt.Errorf("schema init error: %w\nSQL: %s", err, s)
		}
	}
	return nil
}

// UpsertChunk inserts or replaces a code chunk and its FTS entry.
func (d *DB) UpsertChunk(c ChunkRow) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO code_chunks
			(id, project_path, file_path, rel_path, lang, symbol_name, symbol_kind, content, line_start, line_end, file_hash)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			content=excluded.content, file_hash=excluded.file_hash,
			line_start=excluded.line_start, line_end=excluded.line_end,
			symbol_name=excluded.symbol_name, symbol_kind=excluded.symbol_kind
	`, c.ID, c.ProjectPath, c.FilePath, c.RelPath, c.Lang,
		c.SymbolName, c.SymbolKind, c.Content, c.LineStart, c.LineEnd, c.FileHash)
	if err != nil {
		return err
	}

	// Delete old FTS entry (if any) then re-insert
	if _, err := tx.Exec(`DELETE FROM chunks_fts WHERE id = ?`, c.ID); err != nil {
		return fmt.Errorf("failed to delete old FTS entry: %w", err)
	}
	if _, err = tx.Exec(`INSERT INTO chunks_fts (id, content, symbol_name, rel_path) VALUES (?,?,?,?)`,
		c.ID, c.Content, c.SymbolName, c.RelPath); err != nil {
		return fmt.Errorf("failed to insert FTS entry: %w", err)
	}

	return tx.Commit()
}

// UpsertEmbedding stores a float32 embedding as JSON for a chunk.
func (d *DB) UpsertEmbedding(id string, emb []float32) error {
	data, err := json.Marshal(emb)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec(`
		INSERT INTO chunk_embeddings (id, embedding) VALUES (?,?)
		ON CONFLICT(id) DO UPDATE SET embedding=excluded.embedding
	`, id, string(data))
	return err
}

// DeleteByFile removes all chunks (and their FTS/embedding entries) for a given file.
func (d *DB) DeleteByFile(filePath string) error {
	rows, err := d.conn.Query(`SELECT id FROM code_chunks WHERE file_path=? AND project_path=?`,
		filePath, d.ProjectPath)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()

	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	for _, id := range ids {
		tx.Exec(`DELETE FROM code_chunks WHERE id=?`, id)
		tx.Exec(`DELETE FROM chunks_fts WHERE id=?`, id)
		tx.Exec(`DELETE FROM chunk_embeddings WHERE id=?`, id)
	}
	return tx.Commit()
}

// ClearProject deletes all indexed data for the project.
// NOTE: chunks_fts and chunk_embeddings are filtered via JOIN to code_chunks
// to avoid accidentally deleting data from other projects sharing the DB.
func (d *DB) ClearProject() error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove FTS entries for chunks belonging to this project
	tx.Exec(`
		DELETE FROM chunks_fts
		WHERE id IN (SELECT id FROM code_chunks WHERE project_path=?)
	`, d.ProjectPath)
	// Remove embeddings for chunks belonging to this project
	tx.Exec(`
		DELETE FROM chunk_embeddings
		WHERE id IN (SELECT id FROM code_chunks WHERE project_path=?)
	`, d.ProjectPath)
	tx.Exec(`DELETE FROM code_chunks WHERE project_path=?`, d.ProjectPath)
	tx.Exec(`DELETE FROM project_meta WHERE project_path=?`, d.ProjectPath)
	return tx.Commit()
}

// UpdateMeta upserts project metadata (indexed_at, total_chunks, merkle_root).
func (d *DB) UpdateMeta(totalChunks int, merkleRoot string) error {
	_, err := d.conn.Exec(`
		INSERT INTO project_meta (project_path, indexed_at, total_chunks, merkle_root)
		VALUES (?, CURRENT_TIMESTAMP, ?, ?)
		ON CONFLICT(project_path) DO UPDATE SET
			indexed_at=CURRENT_TIMESTAMP, total_chunks=excluded.total_chunks, merkle_root=excluded.merkle_root
	`, d.ProjectPath, totalChunks, merkleRoot)
	return err
}

// GetFileHashes returns a map[relPath]fileHash for all indexed files in this project.
// Used to restore the Merkle tree state for incremental diff detection.
func (d *DB) GetFileHashes() (map[string]string, error) {
	rows, err := d.conn.Query(`
		SELECT DISTINCT rel_path, file_hash
		FROM code_chunks
		WHERE project_path = ?
	`, d.ProjectPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var rel, hash string
		if err := rows.Scan(&rel, &hash); err != nil {
			continue
		}
		out[rel] = hash
	}
	return out, nil
}

// GetMeta returns (totalChunks, merkleRoot, error).
func (d *DB) GetMeta() (int, string, error) {
	var total int
	var root string
	err := d.conn.QueryRow(`SELECT total_chunks, merkle_root FROM project_meta WHERE project_path=?`,
		d.ProjectPath).Scan(&total, &root)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}
	return total, root, err
}

// BM25Search runs FTS5 keyword search and returns up to limit chunk IDs ranked by BM25.
func (d *DB) BM25Search(query string, limit int) ([]BM25Result, error) {
	rows, err := d.conn.Query(`
		SELECT id, rel_path, symbol_name, rank
		FROM chunks_fts
		WHERE chunks_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BM25Result
	for rows.Next() {
		var r BM25Result
		rows.Scan(&r.ID, &r.RelPath, &r.SymbolName, &r.Rank)
		results = append(results, r)
	}
	return results, nil
}

// GetChunksByIDs fetches full chunk data for a list of IDs.
func (d *DB) GetChunksByIDs(ids []string) ([]ChunkRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	q := `SELECT id, project_path, file_path, rel_path, lang, symbol_name, symbol_kind, content, line_start, line_end, file_hash
		FROM code_chunks WHERE id IN (` + joinStr(placeholders, ",") + `)`
	rows, err := d.conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chunks []ChunkRow
	for rows.Next() {
		var c ChunkRow
		rows.Scan(&c.ID, &c.ProjectPath, &c.FilePath, &c.RelPath, &c.Lang,
			&c.SymbolName, &c.SymbolKind, &c.Content, &c.LineStart, &c.LineEnd, &c.FileHash)
		chunks = append(chunks, c)
	}
	return chunks, nil
}

// GetAllEmbeddings loads chunk embeddings for the project up to maxRows.
// Pass 0 for maxRows to use the default safety cap (10000).
// This prevents unbounded RAM usage on large codebases.
func (d *DB) GetAllEmbeddings(maxRows int) ([]EmbeddingRow, error) {
	const defaultCap = 10000
	if maxRows <= 0 {
		maxRows = defaultCap
	}
	rows, err := d.conn.Query(`
		SELECT ce.id, cc.rel_path, cc.symbol_name, ce.embedding
		FROM chunk_embeddings ce
		JOIN code_chunks cc ON cc.id = ce.id
		WHERE cc.project_path = ?
		LIMIT ?
	`, d.ProjectPath, maxRows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []EmbeddingRow
	for rows.Next() {
		var r EmbeddingRow
		var embJSON string
		rows.Scan(&r.ID, &r.RelPath, &r.SymbolName, &embJSON)
		json.Unmarshal([]byte(embJSON), &r.Embedding)
		result = append(result, r)
	}
	return result, nil
}

// GetEmbeddingCount returns the total number of embeddings stored for this project.
func (d *DB) GetEmbeddingCount() (int, error) {
	var count int
	err := d.conn.QueryRow(`
		SELECT COUNT(*) FROM chunk_embeddings ce
		JOIN code_chunks cc ON cc.id = ce.id
		WHERE cc.project_path = ?
	`, d.ProjectPath).Scan(&count)
	return count, err
}

// ── Value types ───────────────────────────────────────────────────────────────

// ChunkRow mirrors the code_chunks table.
type ChunkRow struct {
	ID          string
	ProjectPath string
	FilePath    string
	RelPath     string
	Lang        string
	SymbolName  string
	SymbolKind  string
	Content     string
	LineStart   int
	LineEnd     int
	FileHash    string
}

// BM25Result is one FTS5 search result.
type BM25Result struct {
	ID         string
	RelPath    string
	SymbolName string
	Rank       float64
}

// EmbeddingRow holds a chunk ID + float32 embedding for cosine search.
type EmbeddingRow struct {
	ID         string
	RelPath    string
	SymbolName string
	Embedding  []float32
}

func joinStr(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
