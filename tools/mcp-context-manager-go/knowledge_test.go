package main

import (
	"encoding/json"
	"os"
	"testing"
)

// ── cosineSimKI unit tests ────────────────────────────────────────────────────

func TestCosineSimKI_Identical(t *testing.T) {
	v := []float32{1, 0, 0, 0}
	if got := cosineSimKI(v, v); got < 0.999 {
		t.Errorf("identical vectors: expected ~1.0, got %f", got)
	}
}

func TestCosineSimKI_Orthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	if got := cosineSimKI(a, b); got != 0 {
		t.Errorf("orthogonal vectors: expected 0, got %f", got)
	}
}

func TestCosineSimKI_ZeroVector(t *testing.T) {
	a := []float32{0, 0}
	b := []float32{1, 1}
	if got := cosineSimKI(a, b); got != 0 {
		t.Errorf("zero vector: expected 0, got %f", got)
	}
}

func TestCosineSimKI_LengthMismatch(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{1, 2}
	if got := cosineSimKI(a, b); got != 0 {
		t.Errorf("length mismatch: expected 0, got %f", got)
	}
}

// ── rrfScore unit tests ────────────────────────────────────────────────────────

func TestRRFScore_BothRanked(t *testing.T) {
	score := rrfScore(0, 0) // top rank in both lists
	expected := 2.0 / 61.0  // 1/(60+1) + 1/(60+1)
	if abs64(score-expected) > 1e-9 {
		t.Errorf("expected %f, got %f", expected, score)
	}
}

func TestRRFScore_VecOnly(t *testing.T) {
	// bm25Rank = penalised (high number), vecRank = 0
	for _, r := range []int{-1, 999} {
		_ = rrfScore(99, r) // should not panic
	}
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ── Integration: RecallKnowledge with FTS5 only (no OpenAI key) ───────────────

func TestRecallKnowledge_FTSOnly(t *testing.T) {
	// Unset OPENAI_API_KEY to force FTS-only mode
	prev := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if prev != "" {
			os.Setenv("OPENAI_API_KEY", prev)
		}
	}()

	dir := t.TempDir()

	// Seed a KI via CompactMemory
	_, err := CompactMemory(dir, "task-1", "Go Concurrency Patterns",
		"Master goroutines, channels, WaitGroups, and context cancellation.",
		"Used sync.WaitGroup for fan-out, context for cancellation propagation.")
	if err != nil {
		t.Fatalf("CompactMemory failed: %v", err)
	}

	// Recall with a related query
	out, err := RecallKnowledge(dir, "project", "goroutine channel concurrency", 3)
	if err != nil {
		t.Fatalf("RecallKnowledge failed: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty result")
	}
	if !contains(out, "Go Concurrency Patterns") {
		t.Errorf("expected KI to be recalled, got:\n%s", out)
	}
}

func TestRecallKnowledge_EmptyQuery(t *testing.T) {
	dir := t.TempDir()
	out, err := RecallKnowledge(dir, "project", "!!!###", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(out, "Please provide") {
		t.Errorf("expected empty-query message, got: %s", out)
	}
}

func TestRecallKnowledge_NoResults(t *testing.T) {
	dir := t.TempDir()
	out, err := RecallKnowledge(dir, "project", "xyznonexistentquery", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(out, "No relevant") {
		t.Errorf("expected no-results message, got: %s", out)
	}
}

// ── ki_embeddings schema test ─────────────────────────────────────────────────

func TestKiEmbeddingsTableExists(t *testing.T) {
	dir := t.TempDir()
	db, err := GetDBConnection(dir)
	if err != nil {
		t.Fatalf("GetDBConnection failed: %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='ki_embeddings'").Scan(&name)
	if err != nil {
		t.Fatalf("ki_embeddings table not found: %v", err)
	}
}

func TestKiEmbeddingsInsertAndQuery(t *testing.T) {
	dir := t.TempDir()
	db, err := GetDBConnection(dir)
	if err != nil {
		t.Fatalf("GetDBConnection failed: %v", err)
	}
	defer db.Close()

	// Manually insert a fake embedding
	fakeEmb := make([]float32, embeddingDim)
	for i := range fakeEmb {
		fakeEmb[i] = float32(i) * 0.001
	}
	embJSON, _ := json.Marshal(fakeEmb)

	_, err = db.Exec(`INSERT INTO ki_embeddings (ki_path, tactic, embedding) VALUES (?, ?, ?)`,
		"/tmp/test.md", "Test Tactic", string(embJSON))
	if err != nil {
		t.Fatalf("insert ki_embedding failed: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM ki_embeddings").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}

	// Test UPSERT (ON CONFLICT)
	_, err = db.Exec(`
		INSERT INTO ki_embeddings (ki_path, tactic, embedding) VALUES (?, ?, ?)
		ON CONFLICT(ki_path) DO UPDATE SET embedding=excluded.embedding, tactic=excluded.tactic
	`, "/tmp/test.md", "Updated Tactic", string(embJSON))
	if err != nil {
		t.Fatalf("upsert ki_embedding failed: %v", err)
	}

	db.QueryRow("SELECT COUNT(*) FROM ki_embeddings").Scan(&count)
	if count != 1 {
		t.Errorf("expected still 1 row after upsert, got %d", count)
	}
}

// ── embedAndStoreKIDB: no-op when OPENAI_API_KEY unset ───────────────────────

func TestEmbedAndStoreKIDB_NoKey(t *testing.T) {
	prev := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if prev != "" {
			os.Setenv("OPENAI_API_KEY", prev)
		}
	}()

	dir := t.TempDir()
	err := embedAndStoreKIDB(dir, "/tmp/test.md", "tactic", "summary", "decisions")
	if err != nil {
		t.Errorf("expected nil when no API key, got: %v", err)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
