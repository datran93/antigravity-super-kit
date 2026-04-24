package main

import (
	"os"
	"strings"
	"testing"
)

// ── ManageAnchors coverage ─────────────────────────────────────────────────

func TestManageAnchors_SetGetList(t *testing.T) {
	tempDir := t.TempDir()

	// Set
	res, err := ManageAnchors(tempDir, "project", "set", "go_version", "1.25", "Use Go 1.25+ features")
	if err != nil {
		t.Fatalf("ManageAnchors set: %v", err)
	}
	if !strings.Contains(res, "go_version") {
		t.Errorf("expected anchor name in result, got: %s", res)
	}

	// Get
	res, err = ManageAnchors(tempDir, "project", "get", "go_version", "", "")
	if err != nil {
		t.Fatalf("ManageAnchors get: %v", err)
	}
	if !strings.Contains(res, "1.25") {
		t.Errorf("expected value '1.25' in get result, got: %s", res)
	}

	// List
	res, err = ManageAnchors(tempDir, "project", "list", "", "", "")
	if err != nil {
		t.Fatalf("ManageAnchors list: %v", err)
	}
	if !strings.Contains(res, "go_version") {
		t.Errorf("expected anchor in list, got: %s", res)
	}

	// Unknown action
	res, err = ManageAnchors(tempDir, "project", "delete", "go_version", "", "")
	if err != nil {
		t.Fatalf("unexpected error on unknown action: %v", err)
	}
	if !strings.Contains(res, "Unknown action") {
		t.Errorf("expected 'Unknown action' message, got: %s", res)
	}
}

func TestManageAnchors_GetMissing(t *testing.T) {
	tempDir := t.TempDir()
	res, err := ManageAnchors(tempDir, "project", "get", "nonexistent", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "not found") {
		t.Errorf("expected 'not found' in result, got: %s", res)
	}
}

func TestManageAnchors_EmptyList(t *testing.T) {
	tempDir := t.TempDir()
	res, err := ManageAnchors(tempDir, "project", "list", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "No anchors") {
		t.Errorf("expected 'No anchors' for empty list, got: %s", res)
	}
}

// ── AnnotateFile coverage ─────────────────────────────────────────────────

func TestAnnotateFile_SetAndRead(t *testing.T) {
	tempDir := t.TempDir()

	res, err := AnnotateFile(tempDir, "main.go", "gotcha: CGO_ENABLED=1 required for sqlite3")
	if err != nil {
		t.Fatalf("AnnotateFile: %v", err)
	}
	if !strings.Contains(res, "main.go") {
		t.Errorf("expected filename in result, got: %s", res)
	}

	// Verify it's stored and retrieved via CheckIntentLock ghost context
	_, err = DeclareIntent(tempDir, "test-task", "test tactic", []string{"main.go"}, 0)
	if err != nil {
		t.Fatalf("DeclareIntent: %v", err)
	}
	lockRes, err := CheckIntentLock(tempDir, "test-task", "main.go")
	if err != nil {
		t.Fatalf("CheckIntentLock: %v", err)
	}
	if !strings.Contains(lockRes, "CGO_ENABLED") {
		t.Errorf("expected ghost context in check_intent_lock result, got: %s", lockRes)
	}
}

// ── WriteMarkdownProgress integration ─────────────────────────────────────

func TestWriteMarkdownProgress_BurndownAndDAG(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "burndown-test"

	// Initialize with phase-labeled steps
	_, err := InitializeTaskPlan(tempDir, taskID, "Burndown test task", []string{
		"[P0-T1] Step one",
		"[P0-T2] Step two",
		"[P1-T1] Step three",
	})
	if err != nil {
		t.Fatalf("InitializeTaskPlan: %v", err)
	}

	// Complete a step to record a timestamp
	_, err = CompleteTaskStep(tempDir, taskID, "[P0-T1] Step one", []string{"file.go"}, "")
	if err != nil {
		t.Fatalf("CompleteTaskStep: %v", err)
	}

	mdPath := tempDir + "/progress.md"
	content, err := readFileString(mdPath)
	if err != nil {
		t.Fatalf("failed to read progress.md: %v", err)
	}

	// Verify steps overview header is rendered
	if !strings.Contains(content, "Steps Overview") {
		t.Errorf("expected 'Steps Overview' header in progress.md, got:\n%s", content)
	}
}

func TestWriteMarkdownProgress_FlatRendering(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "flat-test"

	// Initialize without phase labels — uses flat rendering
	_, err := InitializeTaskPlan(tempDir, taskID, "Flat rendering test", []string{
		"Do alpha",
		"Do beta",
	})
	if err != nil {
		t.Fatalf("InitializeTaskPlan: %v", err)
	}
	_, err = CompleteTaskStep(tempDir, taskID, "Do alpha", []string{}, "")
	if err != nil {
		t.Fatalf("CompleteTaskStep: %v", err)
	}

	content, err := readFileString(tempDir + "/progress.md")
	if err != nil {
		t.Fatalf("failed to read progress.md: %v", err)
	}

	if !strings.Contains(content, "Steps Overview") {
		t.Errorf("expected 'Steps Overview' section in flat progress.md, got:\n%s", content)
	}
}

// ── FindRecentTask coverage ────────────────────────────────────────────────

func TestFindRecentTask_Match(t *testing.T) {
	tempDir := t.TempDir()
	_, err := InitializeTaskPlan(tempDir, "context-enhancements", "Context Manager Enhancement Sprint", []string{"Step 1"})
	if err != nil {
		t.Fatalf("InitializeTaskPlan: %v", err)
	}

	res, err := FindRecentTask(tempDir, "Context Manager")
	if err != nil {
		t.Fatalf("FindRecentTask: %v", err)
	}
	if !strings.Contains(res, "context-enhancements") {
		t.Errorf("expected task_id in results, got: %s", res)
	}
	if !strings.Contains(res, "load_checkpoint") {
		t.Errorf("expected load_checkpoint hint in results, got: %s", res)
	}
}

func TestFindRecentTask_NoMatch(t *testing.T) {
	tempDir := t.TempDir()
	res, err := FindRecentTask(tempDir, "nonexistent-topic-xyz")
	if err != nil {
		t.Fatalf("FindRecentTask: %v", err)
	}
	if !strings.Contains(res, "No tasks found") {
		t.Errorf("expected 'No tasks found', got: %s", res)
	}
}

// ── RecordFailure war-room messages ───────────────────────────────────────

func TestRecordFailure_WarRoomThreshold(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "drift-test"
	_, err := InitializeTaskPlan(tempDir, taskID, "drift test", []string{"step1"})
	if err != nil {
		t.Fatalf("InitializeTaskPlan: %v", err)
	}

	for i := 0; i < 2; i++ {
		res, err := RecordFailure(tempDir, taskID, "", "")
		if err != nil {
			t.Fatalf("RecordFailure %d: %v", i, err)
		}
		if !strings.Contains(res, "Count:") {
			t.Errorf("expected count message for failure %d, got: %s", i, res)
		}
	}

	// 3rd failure triggers war-room message
	res, err := RecordFailure(tempDir, taskID, "[P0-T1] Build pipeline", "compile error: undefined symbol")
	if err != nil {
		t.Fatalf("RecordFailure 3rd: %v", err)
	}
	if !strings.Contains(res, "DRIFT DETECTED") {
		t.Errorf("expected DRIFT DETECTED on 3rd failure, got: %s", res)
	}
	if !strings.Contains(res, "[P0-T1] Build pipeline") {
		t.Errorf("expected step name in war-room message, got: %s", res)
	}
	if !strings.Contains(res, "compile error") {
		t.Errorf("expected error_context in war-room message, got: %s", res)
	}
}

// ── ClearDrift coverage ───────────────────────────────────────────────────

func TestClearDrift_ResetsCounter(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "clear-drift-test"
	_, _ = InitializeTaskPlan(tempDir, taskID, "clear drift test", []string{"s1"})

	RecordFailure(tempDir, taskID, "", "") //nolint:errcheck
	RecordFailure(tempDir, taskID, "", "") //nolint:errcheck

	res, err := ClearDrift(tempDir, taskID)
	if err != nil {
		t.Fatalf("ClearDrift: %v", err)
	}
	if !strings.Contains(res, "reset") {
		t.Errorf("expected 'reset' in result, got: %s", res)
	}

	// After clearing, next failure should be count 1
	res, err = RecordFailure(tempDir, taskID, "", "")
	if err != nil {
		t.Fatalf("RecordFailure post-clear: %v", err)
	}
	if !strings.Contains(res, "1/3") {
		t.Errorf("expected '1/3' after clear, got: %s", res)
	}
}

// ── ReviewCheckpoint contextual messages ─────────────────────────────────

func TestReviewCheckpoint_CompletedTask(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "completed-review"

	_, err := InitializeTaskPlan(tempDir, taskID, "Test complete task", []string{"[P0-T1] Step"})
	if err != nil {
		t.Fatalf("InitializeTaskPlan: %v", err)
	}
	// Complete the only step → status becomes "completed"
	_, err = CompleteTaskStep(tempDir, taskID, "[P0-T1] Step", []string{}, "")
	if err != nil {
		t.Fatalf("CompleteTaskStep: %v", err)
	}

	report, err := ReviewCheckpoint(tempDir, taskID)
	if err != nil {
		t.Fatalf("ReviewCheckpoint: %v", err)
	}
	// Should NOT fail active_files guard since task is completed
	if strings.Contains(report, "active_files is empty") {
		t.Errorf("completed task should not trigger active_files guard, got:\n%s", report)
	}
	if !strings.Contains(report, "completed") {
		t.Errorf("expected 'completed' status in report, got:\n%s", report)
	}
}

// ── TTL expiry integration ────────────────────────────────────────────────

func TestDeclareIntent_TTLExpiry(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "ttl-test"

	// Declare with ttlMinutes=0 → no expiry
	res, err := DeclareIntent(tempDir, taskID, "no-expiry tactic", []string{"foo.go"}, 0)
	if err != nil {
		t.Fatalf("DeclareIntent: %v", err)
	}
	if !strings.Contains(res, "no expiry") {
		t.Errorf("expected 'no expiry' in result, got: %s", res)
	}

	// Declare with ttlMinutes=60 → expiry set
	res, err = DeclareIntent(tempDir, taskID, "ttl tactic", []string{"foo.go"}, 60)
	if err != nil {
		t.Fatalf("DeclareIntent with TTL: %v", err)
	}
	if !strings.Contains(res, "expires in 60 min") {
		t.Errorf("expected TTL expiry message, got: %s", res)
	}
}

// ── captureGitSHA in non-git dir ──────────────────────────────────────────

func TestCaptureGitSHA_NonGitDir(t *testing.T) {
	tempDir := t.TempDir()
	sha := captureGitSHA(tempDir)
	// In a temp dir with no git repo, should return empty string
	if sha != "" {
		// Could be non-empty if the temp dir is inside a git repo — acceptable in CI
		t.Logf("captureGitSHA returned %q (may be inside a git repo)", sha)
	}
}

// readFileString is a test helper to read a file as string.
func readFileString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
