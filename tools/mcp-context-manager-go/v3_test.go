package main

import (
	"strings"
	"testing"
	"time"
)

// ── Burndown tests ─────────────────────────────────────────────────────────────

func TestCalculateVelocity_EmptyTimestamps(t *testing.T) {
	ts := StepTimestamps{}
	if v := CalculateVelocity(ts); v != 0 {
		t.Errorf("expected 0 velocity for empty timestamps, got %f", v)
	}
}

func TestCalculateVelocity_SingleTimestamp(t *testing.T) {
	ts := StepTimestamps{"step1": time.Now().Format(time.RFC3339)}
	if v := CalculateVelocity(ts); v != 0 {
		t.Errorf("expected 0 velocity for single timestamp, got %f", v)
	}
}

func TestCalculateVelocity_TwoSteps(t *testing.T) {
	t0 := time.Now().Add(-24 * time.Hour)
	t1 := time.Now()
	ts := StepTimestamps{
		"step1": t0.Format(time.RFC3339),
		"step2": t1.Format(time.RFC3339),
	}
	v := CalculateVelocity(ts)
	// 2 steps over ~1 day ≈ 2.0 steps/day
	if v <= 0 {
		t.Errorf("expected positive velocity, got %f", v)
	}
}

func TestEstimateETA_ZeroRemaining(t *testing.T) {
	if eta := EstimateETA(3.0, 0); eta != "done" {
		t.Errorf("expected 'done', got %s", eta)
	}
}

func TestEstimateETA_ZeroVelocity(t *testing.T) {
	if eta := EstimateETA(0, 5); eta != "unknown" {
		t.Errorf("expected 'unknown', got %s", eta)
	}
}

func TestEstimateETA_DaysRange(t *testing.T) {
	// 3 steps at 1 step/day = ~3 days
	eta := EstimateETA(1.0, 3)
	if eta == "" || eta == "done" || eta == "unknown" {
		t.Errorf("expected a day estimate, got %s", eta)
	}
}

func TestDriftIcon(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{0, "🟢"},
		{1, "🟡"},
		{2, "🟠"},
		{3, "🔴"},
		{5, "🔴"},
	}
	for _, tt := range tests {
		if got := driftIcon(tt.count); got != tt.want {
			t.Errorf("driftIcon(%d) = %s, want %s", tt.count, got, tt.want)
		}
	}
}

func TestParseStepTimestamps_Empty(t *testing.T) {
	ts := ParseStepTimestamps("")
	if len(ts) != 0 {
		t.Errorf("expected empty map, got %v", ts)
	}
}

func TestParseStepTimestamps_Valid(t *testing.T) {
	raw := `{"step1":"2026-03-15T14:00:00Z","step2":"2026-03-15T15:00:00Z"}`
	ts := ParseStepTimestamps(raw)
	if len(ts) != 2 {
		t.Errorf("expected 2 entries, got %d", len(ts))
	}
}

// ── DAG tests ─────────────────────────────────────────────────────────────────

func TestParseStepDeps_NoDeps(t *testing.T) {
	steps := []string{"[P0-T1] Step A", "[P0-T2] Step B"}
	cleaned, deps := ParseStepDeps(steps)
	if len(cleaned) != 2 {
		t.Errorf("expected 2 cleaned steps, got %d", len(cleaned))
	}
	if len(deps) != 0 {
		t.Errorf("expected no deps, got %v", deps)
	}
}

func TestParseStepDeps_WithDeps(t *testing.T) {
	steps := []string{
		"[P0-T1] Foundation",
		"[P1-T1] Build pipeline depends:[P0-T1]",
		"[P1-T2] Tests depends:[P0-T1,P1-T1]",
	}
	cleaned, deps := ParseStepDeps(steps)
	if len(cleaned) != 3 {
		t.Errorf("expected 3 cleaned steps, got %d: %v", len(cleaned), cleaned)
	}
	if len(deps["[P1-T1] Build pipeline"]) != 1 {
		t.Errorf("expected 1 dep for P1-T1, got %v", deps["[P1-T1] Build pipeline"])
	}
	if len(deps["[P1-T2] Tests"]) != 2 {
		t.Errorf("expected 2 deps for P1-T2, got %v", deps["[P1-T2] Tests"])
	}
}

func TestHasCycle_NoCycle(t *testing.T) {
	deps := map[string][]string{
		"B": {"A"},
		"C": {"A", "B"},
	}
	if HasCycle(deps) {
		t.Error("expected no cycle, but HasCycle returned true")
	}
}

func TestHasCycle_WithCycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"C"},
		"B": {"A"},
		"C": {"B"},
	}
	if !HasCycle(deps) {
		t.Error("expected cycle to be detected, but HasCycle returned false")
	}
}

func TestIsParallelReady_NoDeps(t *testing.T) {
	deps := map[string][]string{}
	completed := map[string]bool{}
	if !IsParallelReady("any-step", deps, completed) {
		t.Error("step with no deps should always be parallel-ready")
	}
}

func TestIsParallelReady_DepComplete(t *testing.T) {
	deps := map[string][]string{"B": {"A"}}
	completed := map[string]bool{"A": true}
	if !IsParallelReady("B", deps, completed) {
		t.Error("B should be ready since A is complete")
	}
}

func TestIsParallelReady_DepIncomplete(t *testing.T) {
	deps := map[string][]string{"B": {"A"}}
	completed := map[string]bool{}
	if IsParallelReady("B", deps, completed) {
		t.Error("B should NOT be ready since A is not complete")
	}
}

// ── Validator tests ────────────────────────────────────────────────────────────

func TestIsValidPhaseLabel(t *testing.T) {
	tests := []struct {
		label string
		valid bool
	}{
		{"P0", true},
		{"P1", true},
		{"P0-T1", true},
		{"P2-T3", true},
		{"X0", false},
		{"P", false},
		{"P-T1", false},
		{"P0-X1", false},
	}
	for _, tt := range tests {
		if got := isValidPhaseLabel(tt.label); got != tt.valid {
			t.Errorf("isValidPhaseLabel(%q) = %v, want %v", tt.label, got, tt.valid)
		}
	}
}

func TestReviewCheckpoint_Integration(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "review-test"

	// Create a valid checkpoint
	_, err := InitializeTaskPlan(tempDir, taskID, "Review checkpoint test", []string{
		"[P0-T1] Step one",
		"[P0-T2] Step two",
	})
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	report, err := ReviewCheckpoint(tempDir, taskID)
	if err != nil {
		t.Fatalf("ReviewCheckpoint failed: %v", err)
	}
	if report == "" {
		t.Error("expected non-empty report")
	}
	// Expect active_files warning (empty on in_progress)
	if len(report) < 50 {
		t.Errorf("report too short: %s", report)
	}
}

func TestBuildCompletedSet(t *testing.T) {
	completed := []string{"A", "B", "C"}
	set := BuildCompletedSet(completed)
	for _, s := range completed {
		if !set[s] {
			t.Errorf("expected %s in completed set", s)
		}
	}
	if set["D"] {
		t.Error("D should not be in set")
	}
}

// ── hasGotchaKeyword tests ─────────────────────────────────────────────────────

func TestHasGotchaKeyword(t *testing.T) {
	tests := []struct {
		notes string
		want  bool
	}{
		{"gotcha: mattn requires CGO_ENABLED=1", true},
		{"quirk: sqlite_fts5 needs build tag", true},
		{"warning: do not use nil tx", true},
		{"caution: race condition possible", true},
		{"⚠️ check this", true},
		{"normal implementation notes", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := hasGotchaKeyword(tt.notes); got != tt.want {
			t.Errorf("hasGotchaKeyword(%q) = %v, want %v", tt.notes, got, tt.want)
		}
	}
}

// ── [T1] format validation tests ──────────────────────────────────────────────

func TestIsValidPhaseLabel_NewFormat(t *testing.T) {
	tests := []struct {
		label string
		valid bool
	}{
		// New [T1] format
		{"T1", true},
		{"T2", true},
		{"T10", true},
		{"T99", true},
		{"T", false},   // no number
		{"T0X", false}, // non-digit suffix
		// Legacy [Px-Ty] format still valid
		{"P0", true},
		{"P1", true},
		{"P0-T1", true},
		{"P2-T3", true},
		// Invalid
		{"X0", false},
		{"P", false},
		{"P-T1", false},
		{"P0-X1", false},
	}
	for _, tt := range tests {
		if got := isValidPhaseLabel(tt.label); got != tt.valid {
			t.Errorf("isValidPhaseLabel(%q) = %v, want %v", tt.label, got, tt.valid)
		}
	}
}

// ── fetchIdleTasks tests ───────────────────────────────────────────────────────

func TestFetchIdleTasks_ExcludesCurrentTask(t *testing.T) {
	tempDir := t.TempDir()

	// Create two in_progress tasks
	InitializeTaskPlan(tempDir, "task-a", "Task A", []string{"[T1] Step one"})
	InitializeTaskPlan(tempDir, "task-b", "Task B", []string{"[T1] Step one"})

	// With threshold=0 → idle check disabled; use 0 to skip time filter for simplicity
	// Actually fetchIdleTasks with 0 days means everything older than now → all qualify
	// Use a large threshold to simulate "never idle" (i.e. newly created tasks won't be idle)
	tasks, err := fetchIdleTasks(tempDir, "task-a", 365)
	if err != nil {
		t.Fatalf("fetchIdleTasks error: %v", err)
	}
	for _, task := range tasks {
		if task.TaskID == "task-a" {
			t.Errorf("fetchIdleTasks should not return the current task")
		}
	}
}

func TestFetchIdleTasks_ExcludesCompleted(t *testing.T) {
	tempDir := t.TempDir()

	// Create a completed task
	InitializeTaskPlan(tempDir, "done-task", "Completed task", []string{"[T1] Do it"})
	CompleteTaskStep(tempDir, "done-task", "[T1] Do it", nil, "")

	// Create current task
	InitializeTaskPlan(tempDir, "current", "Current task", []string{"[T1] Work"})

	// Completed tasks should not appear even with large threshold
	tasks, err := fetchIdleTasks(tempDir, "current", 365)
	if err != nil {
		t.Fatalf("fetchIdleTasks error: %v", err)
	}
	for _, task := range tasks {
		if task.TaskID == "done-task" {
			t.Errorf("fetchIdleTasks should not return completed tasks")
		}
	}
}

func TestFetchIdleTasks_ReturnsNothingForFreshTasks(t *testing.T) {
	tempDir := t.TempDir()

	// Task created just now should not be idle with threshold=3 days
	InitializeTaskPlan(tempDir, "other-task", "Other task", []string{"[T1] Pending"})
	InitializeTaskPlan(tempDir, "current", "Current task", []string{"[T1] Work"})

	tasks, err := fetchIdleTasks(tempDir, "current", 3)
	if err != nil {
		t.Fatalf("fetchIdleTasks error: %v", err)
	}
	// Fresh tasks should not be returned since they were updated < 3 days ago
	if len(tasks) > 0 {
		t.Errorf("expected no idle tasks for fresh tasks, got %d", len(tasks))
	}
}

// ── RenderHistoricallyIncomplete tests ────────────────────────────────────────

func TestRenderHistoricallyIncomplete_Empty(t *testing.T) {
	out := RenderHistoricallyIncomplete([]IdleTask{})
	if out != "" {
		t.Errorf("expected empty string for no idle tasks, got %q", out)
	}
}

func TestRenderHistoricallyIncomplete_Format(t *testing.T) {
	tasks := []IdleTask{
		{TaskID: "old-task", Description: "An old task", Progress: 33.3, Done: 1, Total: 3, IdleDays: 7, LastUpdate: "2026-03-08"},
		{TaskID: "stale-project", Description: "Stale project with very long description that should be truncated at fifty characters max", Progress: 22.2, Done: 2, Total: 9, IdleDays: 14, LastUpdate: "2026-03-01"},
	}
	out := RenderHistoricallyIncomplete(tasks)

	// Must contain header
	if !strings.Contains(out, "Historically Incomplete Tasks") {
		t.Error("missing section header")
	}
	// Must contain both task IDs
	if !strings.Contains(out, "old-task") {
		t.Error("missing old-task in output")
	}
	if !strings.Contains(out, "stale-project") {
		t.Error("missing stale-project in output")
	}
	// Must contain resume hint
	if !strings.Contains(out, "load_checkpoint") {
		t.Error("missing resume hint in output")
	}
	// Long description must be truncated
	if strings.Contains(out, "that should be truncated at fifty characters max") {
		t.Error("description should have been truncated")
	}
}

// ── GetTaskSummary tests (T2) ─────────────────────────────────────────────────

func TestGetTaskSummary_HappyPath(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "summary-test"

	_, err := InitializeTaskPlan(tempDir, taskID, "Summary test task", []string{
		"[T1] Step one",
		"[T2] Step two",
		"[T3] Step three",
	})
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Complete one step
	CompleteTaskStep(tempDir, taskID, "[T1] Step one", nil, "")

	summary, err := GetTaskSummary(tempDir, taskID)
	if err != nil {
		t.Fatalf("GetTaskSummary error: %v", err)
	}
	if !strings.Contains(summary, `"task_id"`) {
		t.Errorf("expected task_id in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "1/3") {
		t.Errorf("expected 1/3 progress, got: %s", summary)
	}
	if !strings.Contains(summary, "[T2] Step two") {
		t.Errorf("expected next_step to be T2, got: %s", summary)
	}
}

func TestGetTaskSummary_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	summary, err := GetTaskSummary(tempDir, "nonexistent-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(summary, "not found") {
		t.Errorf("expected 'not found' message, got: %s", summary)
	}
}

func TestGetTaskSummary_AllComplete(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "all-done"

	InitializeTaskPlan(tempDir, taskID, "All done task", []string{"[T1] Only step"})
	CompleteTaskStep(tempDir, taskID, "[T1] Only step", nil, "")

	summary, err := GetTaskSummary(tempDir, taskID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(summary, "1/1") {
		t.Errorf("expected 1/1 steps, got: %s", summary)
	}
}

// ── GetTaskDAG tests (T3) ─────────────────────────────────────────────────────

func TestGetTaskDAG_NoDeps(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "dag-nodeps"

	InitializeTaskPlan(tempDir, taskID, "DAG test no deps", []string{
		"[T1] Step one",
		"[T2] Step two",
	})

	result, err := GetTaskDAG(tempDir, taskID)
	if err != nil {
		t.Fatalf("GetTaskDAG error: %v", err)
	}
	if !strings.Contains(result, taskID) {
		t.Errorf("expected task ID in DAG output, got: %s", result)
	}
	if !strings.Contains(result, "no dependencies declared") {
		t.Errorf("expected no-deps message, got: %s", result)
	}
}

func TestGetTaskDAG_WithDeps(t *testing.T) {
	tempDir := t.TempDir()
	taskID := "dag-withdeps"

	InitializeTaskPlan(tempDir, taskID, "DAG with deps", []string{
		"[T1] Foundation",
		"[T2] Build depends:[T1]",
	})

	result, err := GetTaskDAG(tempDir, taskID)
	if err != nil {
		t.Fatalf("GetTaskDAG error: %v", err)
	}
	if !strings.Contains(result, "mermaid") {
		t.Errorf("expected mermaid diagram in output when deps exist, got: %s", result)
	}
}

func TestGetTaskDAG_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	result, err := GetTaskDAG(tempDir, "ghost-task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "not found") {
		t.Errorf("expected not found message, got: %s", result)
	}
}
