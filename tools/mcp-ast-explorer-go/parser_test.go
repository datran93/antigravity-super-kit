package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── Parser / Extraction Tests ──────────────────────────────────────────────────

func TestParseAndExtract_Python(t *testing.T) {
	tempDir := t.TempDir()

	pyCode := `
def hello_world(name: str) -> str:
    """This is a docstring."""
    return f"Hello, {name}"

class MyClass:
    def method(self):
        pass
`
	pyFile := filepath.Join(tempDir, "test.py")
	if err := os.WriteFile(pyFile, []byte(pyCode), 0644); err != nil {
		t.Fatalf("failed to write python file: %v", err)
	}

	relPath, nodes := parseAndExtract(pyFile, tempDir, "python")
	if relPath == "" || len(nodes) == 0 {
		t.Fatalf("expected nodes, got none")
	}

	foundHello, foundClass := false, false
	for _, n := range nodes {
		if n.Name == "hello_world" {
			foundHello = true
			if n.Doc != "This is a docstring." {
				t.Errorf("Expected docstring 'This is a docstring.', got '%s'", n.Doc)
			}
			if n.Signature != "(name: str) -> str" {
				t.Errorf("Expected signature '(name: str) -> str', got '%s'", n.Signature)
			}
		}
		if n.Name == "MyClass" {
			foundClass = true
		}
	}

	if !foundHello || !foundClass {
		t.Errorf("Missing expected nodes in Python parser (hello=%v, class=%v)", foundHello, foundClass)
	}
}

func TestParseAndExtract_Go(t *testing.T) {
	tempDir := t.TempDir()

	goCode := `package main

// MyGoFunc does something
func MyGoFunc(a int) bool {
	return true
}
`
	goFile := filepath.Join(tempDir, "test.go")
	_ = os.WriteFile(goFile, []byte(goCode), 0644)

	_, nodes := parseAndExtract(goFile, tempDir, "go")
	if len(nodes) == 0 {
		t.Fatalf("expected go nodes, got none")
	}

	foundGo := false
	for _, n := range nodes {
		if n.Name == "MyGoFunc" {
			foundGo = true
			if n.Doc != "MyGoFunc does something" {
				t.Errorf("Expected go docstring 'MyGoFunc does something', got '%s'", n.Doc)
			}
			if n.Signature != "(a int) bool" {
				t.Errorf("Expected signature '(a int) bool', got '%s'", n.Signature)
			}
		}
	}

	if !foundGo {
		t.Errorf("Missing expected nodes in Go parser")
	}
}

func TestParseAndExtract_TypeScript(t *testing.T) {
	tempDir := t.TempDir()

	tsCode := `
// Greet the user
function greet(name: string): string {
    return "Hello " + name;
}

class Greeter {
    greet(name: string) {
        return "Hi " + name;
    }
}
`
	tsFile := filepath.Join(tempDir, "test.ts")
	_ = os.WriteFile(tsFile, []byte(tsCode), 0644)

	_, nodes := parseAndExtract(tsFile, tempDir, "typescript")
	if len(nodes) == 0 {
		t.Fatalf("expected typescript nodes, got none")
	}

	foundGreet := false
	for _, n := range nodes {
		if n.Name == "greet" || n.Name == "Greeter" {
			foundGreet = true
		}
	}
	if !foundGreet {
		t.Errorf("Missing expected nodes in TypeScript parser")
	}
}

func TestParseAndExtract_JavaScript(t *testing.T) {
	tempDir := t.TempDir()

	jsCode := `
// SayHello greets
function sayHello(name) {
    return "Hello " + name;
}
`
	jsFile := filepath.Join(tempDir, "test.js")
	_ = os.WriteFile(jsFile, []byte(jsCode), 0644)

	_, nodes := parseAndExtract(jsFile, tempDir, "javascript")
	if len(nodes) == 0 {
		t.Fatalf("expected javascript nodes, got none")
	}

	found := false
	for _, n := range nodes {
		if n.Name == "sayHello" {
			found = true
		}
	}
	if !found {
		t.Errorf("Missing 'sayHello' node in JavaScript parser")
	}
}

func TestParseAndExtract_UnknownLang(t *testing.T) {
	tempDir := t.TempDir()
	f := filepath.Join(tempDir, "test.rb")
	_ = os.WriteFile(f, []byte("def foo; end"), 0644)

	relPath, nodes := parseAndExtract(f, tempDir, "ruby")
	if relPath != "" || len(nodes) != 0 {
		t.Errorf("expected empty result for unknown language, got relPath=%s nodes=%d", relPath, len(nodes))
	}
}

func TestParseAndExtract_NonExistentFile(t *testing.T) {
	relPath, nodes := parseAndExtract("/nonexistent/path/file.go", "/nonexistent", "go")
	if relPath != "" || len(nodes) != 0 {
		t.Errorf("expected empty result for missing file")
	}
}

// ── Language Detection Tests ───────────────────────────────────────────────────

func TestGetLanguageFromExt(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".py", "python"},
		{".go", "go"},
		{".ts", "typescript"},
		{".cts", "typescript"},
		{".mts", "typescript"},
		{".tsx", "tsx"},
		{".js", "javascript"},
		{".jsx", "javascript"},
		{".cjs", "javascript"},
		{".mjs", "javascript"},
		{".rb", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := getLanguageFromExt(tt.ext)
		if got != tt.want {
			t.Errorf("getLanguageFromExt(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

func TestGetFamilyFromExt(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".py", "python"},
		{".go", "go"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".js", "javascript"},
		{".rb", ""},
	}
	for _, tt := range tests {
		got := getFamilyFromExt(tt.ext)
		if got != tt.want {
			t.Errorf("getFamilyFromExt(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

// ── Workspace Tests ────────────────────────────────────────────────────────────

func TestGetProjectFiles_Walk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple file structure
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "node_modules", "ignored.js"), []byte(""), 0644)

	files := getProjectFiles(tmpDir)
	for _, f := range files {
		if strings.Contains(f, "node_modules") {
			t.Errorf("Expected node_modules to be ignored, but got: %s", f)
		}
	}
	found := false
	for _, f := range files {
		if strings.HasSuffix(f, "main.go") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected main.go to be found in project files")
	}
}

func TestGetMainLanguageFamily(t *testing.T) {
	files := []string{"a.go", "b.go", "c.py"}
	family := getMainLanguageFamily(files)
	if family != "go" {
		t.Errorf("Expected 'go' family, got '%s'", family)
	}
}

func TestGetMainLanguageFamily_Empty(t *testing.T) {
	family := getMainLanguageFamily(nil)
	if family != "" {
		t.Errorf("Expected empty family for no files, got '%s'", family)
	}
}

// ── Architecture Command Tests ─────────────────────────────────────────────────

func TestGetProjectArchitecture_InvalidPath(t *testing.T) {
	res, err := GetProjectArchitecture("/nonexistent/path", "", 1000, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "❌ Path not found") {
		t.Errorf("Expected 'Path not found' error, got: %s", res)
	}
}

func TestGetProjectArchitecture_WithGoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	goCode := `package main

func Add(a, b int) int {
    return a + b
}
`
	_ = os.WriteFile(filepath.Join(tmpDir, "add.go"), []byte(goCode), 0644)

	res, err := GetProjectArchitecture(tmpDir, "", 1000, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "PROJECT ARCHITECTURE AST") {
		t.Errorf("Expected architecture output header, got: %s", res)
	}
}

func TestGetProjectArchitecture_MaxFiles(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 5; i++ {
		code := `package main
func Fn` + string(rune('A'+i)) + `() {}`
		fname := filepath.Join(tmpDir, string(rune('a'+i))+"_file.go")
		_ = os.WriteFile(fname, []byte(code), 0644)
	}

	// Limit to 2 files
	res, err := GetProjectArchitecture(tmpDir, "", 2, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "Reached limit of 2 files") {
		t.Errorf("Expected file limit warning, got: %s", res)
	}
}

func TestGetProjectArchitecture_IncludeDocs(t *testing.T) {
	tmpDir := t.TempDir()
	goCode := `package main

// Documented function with a useful comment
func Documented() {}
`
	_ = os.WriteFile(filepath.Join(tmpDir, "doc.go"), []byte(goCode), 0644)

	res, err := GetProjectArchitecture(tmpDir, "", 1000, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "Documented function") {
		t.Errorf("Expected doc comment in output when include_docs=true. Got: %s", res)
	}
}

func TestSearchSymbol_Found(t *testing.T) {
	tmpDir := t.TempDir()
	goCode := `package main

func UniqueTarget(x int) {}
`
	_ = os.WriteFile(filepath.Join(tmpDir, "target.go"), []byte(goCode), 0644)

	res, err := SearchSymbol(tmpDir, "UniqueTarget")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "UniqueTarget") {
		t.Errorf("Expected 'UniqueTarget' in search results. Got: %s", res)
	}
}

func TestSearchSymbol_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	goCode := `package main

func SomeOtherFunc() {}
`
	_ = os.WriteFile(filepath.Join(tmpDir, "other.go"), []byte(goCode), 0644)

	res, err := SearchSymbol(tmpDir, "NonExistentSymbol99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "No symbols matching") {
		t.Errorf("Expected 'No symbols matching' message. Got: %s", res)
	}
}
