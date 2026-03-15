package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── FindUsages tests ──────────────────────────────────────────────────────────

func TestFindUsages_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two Go source files with references
	writeTestFile(t, filepath.Join(tmpDir, "a.go"), `package main
func MyFunc() {}
func caller() { MyFunc() }
`)
	writeTestFile(t, filepath.Join(tmpDir, "b.go"), `package main
// MyFunc is also referenced here
var x = MyFunc
`)

	result, err := FindUsages(tmpDir, "MyFunc")
	if err != nil {
		t.Fatalf("FindUsages error: %v", err)
	}
	if !strings.Contains(result, "a.go") {
		t.Errorf("expected a.go in result, got: %s", result)
	}
	if !strings.Contains(result, "b.go") {
		t.Errorf("expected b.go in result, got: %s", result)
	}
}

func TestFindUsages_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "main.go"), `package main
func MYFUNC() {}
func myfunc() {}
`)

	result, err := FindUsages(tmpDir, "myfunc")
	if err != nil {
		t.Fatalf("FindUsages error: %v", err)
	}
	// Should find both MYFUNC and myfunc
	count := strings.Count(result, "L")
	if count < 2 {
		t.Errorf("expected at least 2 matching lines (case insensitive), got: %s", result)
	}
}

func TestFindUsages_NoResults(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "main.go"), `package main
func DoSomething() {}
`)

	result, err := FindUsages(tmpDir, "NonExistentSymbol_XYZ")
	if err != nil {
		t.Fatalf("FindUsages error: %v", err)
	}
	if !strings.Contains(result, "No usages") {
		t.Errorf("expected 'No usages' message, got: %s", result)
	}
}

func TestFindUsages_EmptySymbol(t *testing.T) {
	tmpDir := t.TempDir()
	result, err := FindUsages(tmpDir, "")
	if err != nil {
		t.Fatalf("FindUsages error: %v", err)
	}
	if !strings.Contains(result, "required") {
		t.Errorf("expected 'required' message for empty symbol, got: %s", result)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file %s: %v", path, err)
	}
}
