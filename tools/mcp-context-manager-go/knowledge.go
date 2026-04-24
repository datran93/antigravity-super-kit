package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// cleanQuery removes characters that break FTS5 MATCH syntax.
func cleanQuery(query string) string {
	var builder strings.Builder
	for _, ch := range query {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || unicode.IsSpace(ch) {
			builder.WriteRune(ch)
		}
	}
	return builder.String()
}

// cosineSimKI computes cosine similarity between two float32 slices.
// Returns 0 for zero-length or mismatched slices.
func cosineSimKI(a, b []float32) float32 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	denom := float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB)))
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// getEmbedding uses the ONNX-first → OpenAI-fallback chain from onnx.go.
// Returns nil (no error) when neither ONNX nor OpenAI is available — graceful degradation.
func getEmbedding(text string) ([]float32, error) {
	emb, _, err := getEmbeddingWithFallback(text)
	return emb, err
}

// kiRank holds ranking data for RRF fusion.
type kiRank struct {
	kiPath     string
	tacticName string
	summary    string
	decisions  string
	bm25Rank   int // 0-based rank from FTS5 (lower = better)
	vecRank    int // 0-based rank from cosine (lower = better), -1 if missing
}

// rrfScore computes Reciprocal Rank Fusion score.
// k=60 is the standard RRF constant.
func rrfScore(bm25Rank, vecRank int) float64 {
	const k = 60.0
	score := 1.0 / (k + float64(bm25Rank+1))
	if vecRank >= 0 {
		score += 1.0 / (k + float64(vecRank+1))
	}
	return score
}

func bm25SearchFTS5(db interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}, tableName, ftsQuery string) (*sql.Rows, error) {
	return db.Query(fmt.Sprintf(`
		SELECT tactic_name, ki_path, summary, decisions
		FROM %s
		WHERE %s MATCH ?
		ORDER BY rank
		LIMIT 20
	`, tableName, tableName), ftsQuery)
}

// likeSearchFallback performs a multi-token LIKE search across knowledge_fts columns.
// Used when SQLite FTS5 module is unavailable (e.g. macOS system SQLite).
func likeSearchFallback(db interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}, tableName string, tokens []string) (*sql.Rows, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens")
	}
	// Build: WHERE (summary LIKE '%t1%' OR decisions LIKE '%t1%' OR tactic_name LIKE '%t1%')
	//        AND   (summary LIKE '%t2%' OR ...)
	var clauses []string
	var args []interface{}
	for _, tok := range tokens {
		pat := "%" + tok + "%"
		clauses = append(clauses, "(summary LIKE ? OR decisions LIKE ? OR tactic_name LIKE ?)")
		args = append(args, pat, pat, pat)
	}
	q := fmt.Sprintf(`SELECT tactic_name, ki_path, summary, decisions FROM %s WHERE `, tableName) +
		strings.Join(clauses, " AND ") + ` LIMIT 20`
	return db.Query(q, args...)
}

// RecallKnowledge retrieves the most relevant Knowledge Items using hybrid search:
// BM25 (FTS5) + cosine similarity (OpenAI embeddings), fused via RRF.
// Falls back to multi-token LIKE search when FTS5 module is unavailable.
func RecallKnowledge(workspacePath, scope, query string, topK int) (string, error) {
	var db *sql.DB
	var err error
	tableName := "knowledge_fts"

	if scope == "global" {
		db, err = GetGlobalDBConnection()
		tableName = "global_knowledge_fts"
	} else {
		db, err = GetDBConnection(workspacePath)
	}

	if err != nil {
		return "", err
	}
	defer db.Close()

	if topK <= 0 {
		topK = 3
	}

	cleanedQuery := cleanQuery(query)
	parts := strings.Fields(cleanedQuery)
	var tokens []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			tokens = append(tokens, p)
		}
	}

	if len(tokens) == 0 {
		return "🔍 Please provide a valid search query.", nil
	}

	// Build FTS5 query string (for BM25 path)
	var ftsTokens []string
	for _, t := range tokens {
		ftsTokens = append(ftsTokens, t+"*")
	}
	ftsQuery := strings.Join(ftsTokens, " OR ")

	// ── Step 1: BM25 via FTS5; graceful fallback to LIKE ────────────────────
	var rows *sql.Rows
	rows, err = bm25SearchFTS5(db, tableName, ftsQuery)
	if err != nil {
		// FTS5 unavailable (e.g. "no such module: fts5") — use LIKE fallback
		rows, err = likeSearchFallback(db, tableName, tokens)
		if err != nil {
			return "", fmt.Errorf("knowledge search failed (both FTS5 and LIKE): %v", err)
		}
	}
	defer rows.Close()

	rankMap := make(map[string]*kiRank)
	var orderedPaths []string
	bm25Idx := 0
	for rows.Next() {
		var tacticName, kiPath, summary, decisions string
		if err := rows.Scan(&tacticName, &kiPath, &summary, &decisions); err != nil {
			continue
		}
		rankMap[kiPath] = &kiRank{
			kiPath:     kiPath,
			tacticName: tacticName,
			summary:    summary,
			decisions:  decisions,
			bm25Rank:   bm25Idx,
			vecRank:    -1, // not yet assigned
		}
		orderedPaths = append(orderedPaths, kiPath)
		bm25Idx++
	}
	rows.Close()

	// ── Step 2: Vector cosine similarity (optional, degrades if no API key) ─
	queryEmb, embErr := getEmbedding(query)
	if embErr == nil && len(queryEmb) > 0 {
		// Fetch all stored embeddings from ki_embeddings table
		embRows, err := db.Query(`SELECT ki_path, tactic, embedding FROM ki_embeddings`)
		if err == nil {
			defer embRows.Close()

			type embEntry struct {
				kiPath    string
				tactic    string
				embedding []float32
			}
			var allEmbs []embEntry

			for embRows.Next() {
				var kiPath, tactic, embJSON string
				if err := embRows.Scan(&kiPath, &tactic, &embJSON); err != nil {
					continue
				}
				var emb []float32
				if err := json.Unmarshal([]byte(embJSON), &emb); err != nil {
					continue
				}
				allEmbs = append(allEmbs, embEntry{kiPath, tactic, emb})
			}
			embRows.Close()

			// Score all by cosine similarity
			type vecScore struct {
				kiPath string
				score  float32
			}
			var vecScores []vecScore
			for _, e := range allEmbs {
				sim := cosineSimKI(queryEmb, e.embedding)
				vecScores = append(vecScores, vecScore{e.kiPath, sim})
			}
			sort.Slice(vecScores, func(i, j int) bool {
				return vecScores[i].score > vecScores[j].score // descending
			})

			// Merge vec results into rankMap (take top 20)
			for vecIdx, vs := range vecScores {
				if vecIdx >= 20 {
					break
				}
				if r, exists := rankMap[vs.kiPath]; exists {
					r.vecRank = vecIdx
				} else {
					// Vec found something BM25 missed — fetch its metadata via LIKE
					// (avoid FTS5 dependency in this secondary lookup path)
					var tacticName, summary, decisions string
					err := db.QueryRow(fmt.Sprintf(`
						SELECT tactic_name, ki_path, summary, decisions
						FROM %s WHERE ki_path = ?`, tableName), vs.kiPath,
					).Scan(&tacticName, &vs.kiPath, &summary, &decisions)
					if err == nil {
						rankMap[vs.kiPath] = &kiRank{
							kiPath:     vs.kiPath,
							tacticName: tacticName,
							summary:    summary,
							decisions:  decisions,
							bm25Rank:   bm25Idx + vecIdx, // penalise: not in BM25 top
							vecRank:    vecIdx,
						}
						orderedPaths = append(orderedPaths, vs.kiPath)
					}
				}
			}
		}
	}

	if len(rankMap) == 0 {
		return fmt.Sprintf("🔍 No relevant Knowledge Items found for query: '%s'", query), nil
	}

	// ── Step 3: RRF fusion ───────────────────────────────────────────────────
	type rankEntry struct {
		r     *kiRank
		score float64
	}
	var ranked []rankEntry
	for _, r := range rankMap {
		ranked = append(ranked, rankEntry{r, rrfScore(r.bm25Rank, r.vecRank)})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})
	if len(ranked) > topK {
		ranked = ranked[:topK]
	}

	// ── Step 4: Format output ────────────────────────────────────────────────
	var res []string
	res = append(res, fmt.Sprintf("🧠 Recalled Knowledge for '%s':\n", query))
	for _, e := range ranked {
		r := e.r
		res = append(res, fmt.Sprintf("### KI: %s", r.tacticName))
		res = append(res, fmt.Sprintf("**Path**: `%s`", r.kiPath))
		res = append(res, fmt.Sprintf("**Summary**: %s", r.summary))
		res = append(res, fmt.Sprintf("**Decisions**: %s\n---", r.decisions))
	}
	return strings.Join(res, "\n"), nil
}

// embedAndStoreKIDB is called asynchronously by CompactMemory to persist an
// OpenAI embedding for a KI. Silently skips when OPENAI_API_KEY is unset.
func embedAndStoreKIDB(workspacePath, kiPath, tactic, summary, decisions string) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil // graceful skip — FTS-only mode
	}

	// Build the text to embed: combine tactic name + summary + decisions
	text := fmt.Sprintf("KI: %s\nSummary: %s\nDecisions: %s", tactic, summary, decisions)
	emb, err := getEmbedding(text)
	if err != nil || len(emb) == 0 {
		return err // non-fatal: caller ignores this
	}

	embJSON, err := json.Marshal(emb)
	if err != nil {
		return err
	}

	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO ki_embeddings (ki_path, tactic, embedding)
		VALUES (?, ?, ?)
		ON CONFLICT(ki_path) DO UPDATE SET embedding=excluded.embedding, tactic=excluded.tactic
	`, kiPath, tactic, string(embJSON))
	return err
}

// CompactMemory distills a tactic into a Knowledge Item, saves it to disk,
// indexes it in FTS5, and (if OPENAI_API_KEY is set) embeds it for hybrid search.
func CompactMemory(workspacePath, taskID, tacticName, summary, decisions string) (string, error) {
	knowledgeDir := filepath.Join(workspacePath, "knowledge")
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to make knowledge dir: %v", err)
	}

	safeName := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(tacticName)), " ", "_")
	if safeName == "" {
		safeName = "unknown_tactic"
	}
	kiPath := filepath.Join(knowledgeDir, fmt.Sprintf("%s.md", safeName))

	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var lockedFilesStr string
	err = db.QueryRow("SELECT locked_files FROM intents WHERE task_id = ?", taskID).Scan(&lockedFilesStr)
	var files []string
	if err == nil && lockedFilesStr != "" {
		json.Unmarshal([]byte(lockedFilesStr), &files)
	}

	content := fmt.Sprintf("# KI: %s\n\n## Summary\n%s\n\n## Affected Files\n", tacticName, summary)
	for _, f := range files {
		content += fmt.Sprintf("- `%s`\n", f)
	}
	content += fmt.Sprintf("\n## Architecture & Decisions\n%s\n", decisions)

	if err := os.WriteFile(kiPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write KI file: %v", err)
	}

	insertFTS := `
		INSERT INTO knowledge_fts (tactic_name, ki_path, summary, decisions)
		VALUES (?, ?, ?, ?)
	`
	if _, err := db.Exec(insertFTS, tacticName, kiPath, summary, decisions); err != nil {
		return "", fmt.Errorf("failed to index FTS: %v", err)
	}

	// Async embed — non-blocking, failure is silently ignored
	go func() {
		_ = embedAndStoreKIDB(workspacePath, kiPath, tacticName, summary, decisions)
	}()

	var notes string
	if err := db.QueryRow("SELECT notes FROM checkpoints WHERE task_id = ?", taskID).Scan(&notes); err == nil {
		newNotes := notes + fmt.Sprintf("\n[COMPACTION] Tactic '%s' completed. KI saved to %s", tacticName, kiPath)
		db.Exec("UPDATE checkpoints SET active_files='[]', notes=? WHERE task_id=?", newNotes, taskID)
		db.Exec("UPDATE intents SET locked_files='[]' WHERE task_id=?", taskID)
		db.Exec("UPDATE drift_tracker SET failure_count=0 WHERE task_id=?", taskID)
	}

	// T23: Auto-emit activity event
	logActivityDB(db, "ki_created", taskID, fmt.Sprintf("KI: %s → %s", tacticName, kiPath))

	return fmt.Sprintf("🗜️ Context Compaction successful. Knowledge Item indexed into local RAG and saved to %s. Memory flushed.", kiPath), nil
}

// FetchAutoLinksContext scans text for @ki and @task tags, fetches their content from DB, and returns a formatted context string.
func FetchAutoLinksContext(db *sql.DB, text string) string {
	kis, tasks := ExtractLinks(text)
	if len(kis) == 0 && len(tasks) == 0 {
		return ""
	}

	var ctxBuilder strings.Builder
	ctxBuilder.WriteString("\n\n---\n### 🔗 Auto-Linked Context\n")

	// Fetch KIs
	for _, ki := range kis {
		var tacticName, summary string
		err := db.QueryRow(`
			SELECT tactic_name, summary 
			FROM knowledge_fts 
			WHERE ki_path LIKE ? OR tactic_name = ?
			LIMIT 1`, "%"+ki+".md", ki).Scan(&tacticName, &summary)
		if err == nil {
			ctxBuilder.WriteString(fmt.Sprintf("- **KI [%s]**: %s\n", tacticName, summary))
		}
	}

	// Fetch Tasks
	for _, task := range tasks {
		var status, desc string
		err := db.QueryRow(`SELECT status, description FROM tasks WHERE task_id = ?`, task).Scan(&status, &desc)
		if err == nil {
			ctxBuilder.WriteString(fmt.Sprintf("- **Task [%s]** (`%s`): %s\n", task, status, desc))
		}
	}

	return ctxBuilder.String()
}
