package budget

import (
	"os"
	"path/filepath"
	"testing"
)

// ── Estimator tests ───────────────────────────────────────────────────────────

func TestEstimateTokens_Basic(t *testing.T) {
	text := "Hello, world!" // 13 bytes → ~3 tokens
	got := EstimateTokens(text)
	if got < 1 {
		t.Errorf("expected at least 1 token, got %d", got)
	}
}

func TestEstimateTokens_Empty(t *testing.T) {
	got := EstimateTokens("")
	if got < 1 {
		t.Errorf("expected minimum 1, got %d", got)
	}
}

func TestEstimateTokens_Proportional(t *testing.T) {
	short := EstimateTokens("abcd")
	long := EstimateTokens(string(make([]byte, 400)))
	if short >= long {
		t.Errorf("longer text should estimate more tokens: short=%d long=%d", short, long)
	}
}

func TestEstimateFileTokens_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	os.WriteFile(path, []byte("package main\n\nfunc main() {}\n"), 0644)
	got := EstimateFileTokens(path)
	if got < 1 {
		t.Errorf("expected at least 1 token for file, got %d", got)
	}
}

func TestEstimateFileTokens_NonExistent(t *testing.T) {
	got := EstimateFileTokens("/tmp/nonexistent_xyz987.go")
	if got != 0 {
		t.Errorf("expected 0 for missing file, got %d", got)
	}
}

func TestEstimateContextLoad(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.go")
	f2 := filepath.Join(dir, "b.go")
	os.WriteFile(f1, []byte("package a\n"), 0644)
	os.WriteFile(f2, []byte("package b\n"), 0644)

	total := EstimateContextLoad([]string{f1, f2})
	each := EstimateFileTokens(f1) + EstimateFileTokens(f2)
	if total != each {
		t.Errorf("EstimateContextLoad should equal sum of individual estimates: %d vs %d", total, each)
	}
}

// ── Thresholds tests ──────────────────────────────────────────────────────────

func TestEvaluate_OK(t *testing.T) {
	r := Evaluate(10000, 100000) // 10%
	if r.Level != LevelOK {
		t.Errorf("expected LevelOK, got %s", r.Level)
	}
}

func TestEvaluate_Warning(t *testing.T) {
	r := Evaluate(65000, 100000) // 65%
	if r.Level != LevelWarning {
		t.Errorf("expected LevelWarning, got %s", r.Level)
	}
}

func TestEvaluate_Critical(t *testing.T) {
	r := Evaluate(85000, 100000) // 85%
	if r.Level != LevelCritical {
		t.Errorf("expected LevelCritical, got %s", r.Level)
	}
}

func TestEvaluate_Overflow(t *testing.T) {
	r := Evaluate(96000, 100000) // 96%
	if r.Level != LevelOverflow {
		t.Errorf("expected LevelOverflow, got %s", r.Level)
	}
}

func TestEvaluate_ZeroMax(t *testing.T) {
	// Should not divide by zero
	r := Evaluate(1000, 0)
	if r.MaxTokens != 100000 {
		t.Errorf("expected fallback max 100000, got %d", r.MaxTokens)
	}
}

func TestEvaluate_Percentage(t *testing.T) {
	r := Evaluate(50000, 100000)
	if r.UsedPercent != 50.0 {
		t.Errorf("expected 50.0%%, got %f", r.UsedPercent)
	}
}

func TestSuggestCompression_HasSuggestions(t *testing.T) {
	for _, level := range []Level{LevelOK, LevelWarning, LevelCritical, LevelOverflow} {
		suggestions := SuggestCompression(level)
		if len(suggestions) == 0 {
			t.Errorf("SuggestCompression(%s) returned empty suggestions", level)
		}
	}
}

// ── Tracker tests ─────────────────────────────────────────────────────────────

func TestTracker_RecordAndSummary(t *testing.T) {
	dir := t.TempDir()
	tracker, err := OpenTracker(dir, "test-session")
	if err != nil {
		t.Fatalf("OpenTracker failed: %v", err)
	}
	defer tracker.Close()

	if err := tracker.RecordUsage("view_file", 1500, "estimate"); err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}
	if err := tracker.RecordUsage("grep_search", 200, "actual"); err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	summary, err := tracker.GetSummary()
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summary.TotalTokens != 1700 {
		t.Errorf("expected 1700 tokens, got %d", summary.TotalTokens)
	}
	if summary.EventCount != 2 {
		t.Errorf("expected 2 events, got %d", summary.EventCount)
	}
}

func TestTracker_DefaultMaxBudget(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := OpenTracker(dir, "session-x")
	defer tracker.Close()

	mb := tracker.GetMaxBudget()
	if mb != 100000 {
		t.Errorf("expected default max 100000, got %d", mb)
	}
}

func TestTracker_SetMaxBudget(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := OpenTracker(dir, "session-y")
	defer tracker.Close()

	tracker.SetMaxBudget(50000)
	if got := tracker.GetMaxBudget(); got != 50000 {
		t.Errorf("expected 50000, got %d", got)
	}
}

func TestTracker_ResetSession(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := OpenTracker(dir, "session-z")
	defer tracker.Close()

	tracker.RecordUsage("test-tool", 9999, "estimate")
	tracker.ResetSession()

	summary, _ := tracker.GetSummary()
	if summary.TotalTokens != 0 {
		t.Errorf("expected 0 after reset, got %d", summary.TotalTokens)
	}
}

func TestTracker_MultiSession(t *testing.T) {
	dir := t.TempDir()
	t1, _ := OpenTracker(dir, "session-1")
	t2, _ := OpenTracker(dir, "session-2")
	defer t1.Close()
	defer t2.Close()

	t1.RecordUsage("toolA", 500, "estimate")
	t2.RecordUsage("toolB", 300, "estimate")

	s1, _ := t1.GetSummary()
	s2, _ := t2.GetSummary()

	if s1.TotalTokens != 500 {
		t.Errorf("session-1 expected 500, got %d", s1.TotalTokens)
	}
	if s2.TotalTokens != 300 {
		t.Errorf("session-2 expected 300, got %d", s2.TotalTokens)
	}
}
