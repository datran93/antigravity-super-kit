package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mcp-codebase-explorer-go/indexer"
	"mcp-codebase-explorer-go/parser"
	"mcp-codebase-explorer-go/search"
	"mcp-codebase-explorer-go/store"
)

// ══════════════════════════════════════════════════════════════════════════════
// Parser Tests (from ast-explorer)
// ══════════════════════════════════════════════════════════════════════════════

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
	writeTestFile(t, pyFile, pyCode)

	relPath, nodes := parser.ParseAndExtract(pyFile, tempDir, "python")
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
		}
		if n.Name == "MyClass" {
			foundClass = true
		}
	}
	if !foundHello || !foundClass {
		t.Errorf("Missing expected nodes (hello=%v, class=%v)", foundHello, foundClass)
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
	writeTestFile(t, goFile, goCode)

	_, nodes := parser.ParseAndExtract(goFile, tempDir, "go")
	if len(nodes) == 0 {
		t.Fatalf("expected go nodes, got none")
	}

	found := false
	for _, n := range nodes {
		if n.Name == "MyGoFunc" {
			found = true
			if n.Doc != "MyGoFunc does something" {
				t.Errorf("Expected docstring 'MyGoFunc does something', got '%s'", n.Doc)
			}
		}
	}
	if !found {
		t.Errorf("Missing MyGoFunc in Go parser")
	}
}

func TestParseAndExtract_GoMethodWithReceiver(t *testing.T) {
	tempDir := t.TempDir()
	goCode := `package main

type Server struct {
	port int
}

// Start starts the server
func (s *Server) Start() error {
	return nil
}
`
	goFile := filepath.Join(tempDir, "server.go")
	writeTestFile(t, goFile, goCode)

	_, nodes := parser.ParseAndExtract(goFile, tempDir, "go")
	if len(nodes) == 0 {
		t.Fatalf("expected nodes, got none")
	}

	foundMethod := false
	for _, n := range nodes {
		if n.Name == "Start" && n.Type == "method" {
			foundMethod = true
		}
	}
	if !foundMethod {
		t.Errorf("Expected method 'Start' to appear")
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
	writeTestFile(t, tsFile, tsCode)

	_, nodes := parser.ParseAndExtract(tsFile, tempDir, "typescript")
	if len(nodes) == 0 {
		t.Fatalf("expected typescript nodes, got none")
	}

	found := false
	for _, n := range nodes {
		if n.Name == "greet" || n.Name == "Greeter" {
			found = true
		}
	}
	if !found {
		t.Errorf("Missing expected nodes in TypeScript parser")
	}
}

func TestParseAndExtract_JavaScript(t *testing.T) {
	tempDir := t.TempDir()
	jsCode := `
function sayHello(name) {
    return "Hello " + name;
}
`
	jsFile := filepath.Join(tempDir, "test.js")
	writeTestFile(t, jsFile, jsCode)

	_, nodes := parser.ParseAndExtract(jsFile, tempDir, "javascript")
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
	writeTestFile(t, f, "def foo; end")

	relPath, nodes := parser.ParseAndExtract(f, tempDir, "ruby")
	if relPath != "" || len(nodes) != 0 {
		t.Errorf("expected empty result for unknown language")
	}
}

func TestParseAndExtract_NonExistentFile(t *testing.T) {
	relPath, nodes := parser.ParseAndExtract("/nonexistent/file.go", "/nonexistent", "go")
	if relPath != "" || len(nodes) != 0 {
		t.Errorf("expected empty result for missing file")
	}
}

func TestLanguageFromExt(t *testing.T) {
	tests := map[string]string{
		".py": "python", ".go": "go", ".ts": "typescript",
		".tsx": "tsx", ".js": "javascript", ".jsx": "javascript",
		".cjs": "javascript", ".mjs": "javascript",
		".cts": "typescript", ".mts": "typescript",
		".rb": "",
	}
	for ext, want := range tests {
		got := parser.LanguageFromExt(ext)
		if got != want {
			t.Errorf("LanguageFromExt(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestFamilyFromExt(t *testing.T) {
	tests := map[string]string{
		".py": "python", ".go": "go", ".ts": "typescript",
		".tsx": "typescript", ".js": "javascript", ".rb": "",
	}
	for ext, want := range tests {
		got := parser.FamilyFromExt(ext)
		if got != want {
			t.Errorf("FamilyFromExt(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestGetMainLanguageFamily(t *testing.T) {
	files := []string{"a.go", "b.go", "c.py"}
	family := parser.GetMainLanguageFamily(files)
	if family != "go" {
		t.Errorf("Expected 'go' family, got '%s'", family)
	}
}

func TestGetMainLanguageFamily_Empty(t *testing.T) {
	family := parser.GetMainLanguageFamily(nil)
	if family != "" {
		t.Errorf("Expected empty family, got '%s'", family)
	}
}

func TestGetProjectFiles_Walk(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "ignored.js"), []byte(""), 0644)

	files := parser.GetProjectFiles(tmpDir)
	for _, f := range files {
		if strings.Contains(f, "node_modules") {
			t.Errorf("Expected node_modules to be ignored, got: %s", f)
		}
	}
	found := false
	for _, f := range files {
		if strings.HasSuffix(f, "main.go") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected main.go to be found")
	}
}

func TestExtractSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "funcs.go"), `package main
func Alpha() {}
func Beta() {}
`)

	symbols := parser.ExtractSymbols(tmpDir, filepath.Join(tmpDir, "funcs.go"), "funcs.go", "go")
	if len(symbols) < 2 {
		t.Errorf("Expected at least 2 symbols, got %d", len(symbols))
	}

	names := make(map[string]bool)
	for _, s := range symbols {
		names[s.Name] = true
		if s.Kind != "function" {
			t.Errorf("Expected kind='function', got '%s' for %s", s.Kind, s.Name)
		}
		if s.RelPath != "funcs.go" {
			t.Errorf("Expected relPath='funcs.go', got '%s'", s.RelPath)
		}
	}
	if !names["Alpha"] || !names["Beta"] {
		t.Errorf("Expected Alpha and Beta symbols, got %v", names)
	}
}

// ── Architecture command tests ────────────────────────────────────────────────

func TestGetProjectArchitecture_InvalidPath(t *testing.T) {
	res, err := parser.GetProjectArchitecture("/nonexistent/path", "", 1000, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "❌ Path not found") {
		t.Errorf("Expected 'Path not found', got: %s", res)
	}
}

func TestGetProjectArchitecture_WithGoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "add.go"), `package main
func Add(a, b int) int { return a + b }
`)

	res, err := parser.GetProjectArchitecture(tmpDir, "", 1000, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "PROJECT ARCHITECTURE AST") {
		t.Errorf("Expected architecture header")
	}
}

func TestGetProjectArchitecture_MaxFiles(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 5; i++ {
		code := `package main
func Fn` + string(rune('A'+i)) + `() {}`
		writeTestFile(t, filepath.Join(tmpDir, string(rune('a'+i))+"_file.go"), code)
	}

	res, _ := parser.GetProjectArchitecture(tmpDir, "", 2, false)
	if !strings.Contains(res, "Reached limit of 2 files") {
		t.Errorf("Expected file limit warning, got: %s", res)
	}
}

func TestGetProjectArchitecture_IncludeDocs(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "doc.go"), `package main

// Documented function with a useful comment
func Documented() {}
`)

	res, _ := parser.GetProjectArchitecture(tmpDir, "", 1000, true)
	if !strings.Contains(res, "Documented function") {
		t.Errorf("Expected doc comment, got: %s", res)
	}
}

func TestSearchSymbol_Found(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "target.go"), `package main
func UniqueTarget(x int) {}
`)

	res, err := parser.SearchSymbol(tmpDir, "UniqueTarget")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "UniqueTarget") {
		t.Errorf("Expected 'UniqueTarget' in results")
	}
}

func TestSearchSymbol_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "other.go"), `package main
func SomeOtherFunc() {}
`)

	res, _ := parser.SearchSymbol(tmpDir, "NonExistent99")
	if !strings.Contains(res, "No symbols matching") {
		t.Errorf("Expected 'No symbols matching'")
	}
}

// ── FindUsages tests ──────────────────────────────────────────────────────────

func TestFindUsages_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "a.go"), `package main
func MyFunc() {}
func caller() { MyFunc() }
`)
	writeTestFile(t, filepath.Join(tmpDir, "b.go"), `package main
var x = MyFunc
`)

	result, err := parser.FindUsages(tmpDir, "MyFunc")
	if err != nil {
		t.Fatalf("FindUsages error: %v", err)
	}
	if !strings.Contains(result, "a.go") {
		t.Errorf("expected a.go in result")
	}
	if !strings.Contains(result, "b.go") {
		t.Errorf("expected b.go in result")
	}
}

func TestFindUsages_EmptySymbol(t *testing.T) {
	tmpDir := t.TempDir()
	result, _ := parser.FindUsages(tmpDir, "")
	if !strings.Contains(result, "required") {
		t.Errorf("expected 'required' message, got: %s", result)
	}
}

func TestFindUsages_NoResults(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "main.go"), `package main
func DoSomething() {}
`)

	result, _ := parser.FindUsages(tmpDir, "NonExistentSymbol_XYZ")
	if !strings.Contains(result, "No usages") {
		t.Errorf("expected 'No usages' message, got: %s", result)
	}
}

func TestFindUsages_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, filepath.Join(tmpDir, "main.go"), `package main
func MYFUNC() {}
func myfunc() {}
`)

	result, _ := parser.FindUsages(tmpDir, "myfunc")
	count := strings.Count(result, "L")
	if count < 2 {
		t.Errorf("expected at least 2 lines (case insensitive), got: %s", result)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Indexer Tests (from codebase-search)
// ══════════════════════════════════════════════════════════════════════════════

func TestWalk_BasicGoProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main(){}"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello"), 0644)
	os.WriteFile(filepath.Join(dir, "ignore.bin"), []byte("binary"), 0644)

	files, err := indexer.Walk(dir, indexer.WalkerConfig{})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}
	if len(files) < 2 {
		t.Errorf("expected at least 2 files, got %d", len(files))
	}
	for _, f := range files {
		if strings.HasSuffix(f.RelPath, ".bin") {
			t.Errorf("binary file should be excluded")
		}
	}
}

func TestWalk_IgnoresNodeModules(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg.js"), []byte("var x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "index.ts"), []byte("export {}"), 0644)

	files, _ := indexer.Walk(dir, indexer.WalkerConfig{})
	for _, f := range files {
		if strings.Contains(f.RelPath, "node_modules") {
			t.Errorf("node_modules should be excluded: %s", f.RelPath)
		}
	}
}

func TestWalk_CustomExtensions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "app.ts"), []byte("const x = 1"), 0644)

	files, _ := indexer.Walk(dir, indexer.WalkerConfig{Extensions: []string{".go"}})
	for _, f := range files {
		if !strings.HasSuffix(f.RelPath, ".go") {
			t.Errorf("Expected only .go files with extension filter, got: %s", f.RelPath)
		}
	}
}

func TestLangFromExt(t *testing.T) {
	cases := map[string]string{".go": "go", ".ts": "typescript", ".py": "python", ".rs": "rust", ".md": "markdown"}
	for ext, want := range cases {
		if got := indexer.LangFromExt(ext); got != want {
			t.Errorf("LangFromExt(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestChunkFile_GoAST(t *testing.T) {
	dir := t.TempDir()
	goCode := `package main

import "fmt"

func Hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

type MyStruct struct {
	Field string
}
`
	path := filepath.Join(dir, "hello.go")
	os.WriteFile(path, []byte(goCode), 0644)

	entry := indexer.FileEntry{AbsPath: path, RelPath: "hello.go", Lang: "go"}
	chunks, err := indexer.ChunkFile(entry)
	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected chunks for Go file")
	}

	found := false
	for _, c := range chunks {
		if c.SymbolName == "Hello" && c.SymbolKind == "function" {
			found = true
		}
	}
	if !found {
		t.Error("expected chunk with SymbolName='Hello'")
	}
}

func TestChunkFile_GoMethod(t *testing.T) {
	dir := t.TempDir()
	code := `package main

type Server struct{}

func (s *Server) Start() error {
	return nil
}
`
	path := filepath.Join(dir, "server.go")
	os.WriteFile(path, []byte(code), 0644)

	entry := indexer.FileEntry{AbsPath: path, RelPath: "server.go", Lang: "go"}
	chunks, _ := indexer.ChunkFile(entry)

	found := false
	for _, c := range chunks {
		if c.SymbolKind == "method" {
			found = true
		}
	}
	if !found {
		t.Error("expected method chunk for receiver function")
	}
}

func TestChunkFile_PythonTreeSitter(t *testing.T) {
	dir := t.TempDir()
	pyCode := `def greet(name):
    return f"Hello, {name}"

class Greeter:
    def say_hi(self):
        pass
`
	path := filepath.Join(dir, "greet.py")
	os.WriteFile(path, []byte(pyCode), 0644)

	entry := indexer.FileEntry{AbsPath: path, RelPath: "greet.py", Lang: "python"}
	chunks, err := indexer.ChunkFile(entry)
	if err != nil {
		t.Fatalf("ChunkFile for Python failed: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("expected chunks for Python file")
	}

	hasNamedSymbol := false
	for _, c := range chunks {
		if !strings.HasPrefix(c.SymbolName, "lines_") {
			hasNamedSymbol = true
		}
	}
	if !hasNamedSymbol {
		t.Error("expected AST-aware symbol names for Python")
	}
}

func TestChunkFile_FallbackLineWindow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	lines := make([]string, 60)
	for i := range lines {
		lines[i] = "key: value"
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)

	entry := indexer.FileEntry{AbsPath: path, RelPath: "config.yaml", Lang: "yaml"}
	chunks, _ := indexer.ChunkFile(entry)
	if len(chunks) == 0 {
		t.Error("expected line-window chunks for YAML file")
	}

	for _, c := range chunks {
		if !strings.HasPrefix(c.SymbolName, "lines_") {
			t.Errorf("expected line-window symbol name, got: %s", c.SymbolName)
		}
	}
}

func TestChunkFile_UniqueIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "funcs.go")
	os.WriteFile(path, []byte("package main\nfunc A(){}\nfunc B(){}\nfunc C(){}"), 0644)

	entry := indexer.FileEntry{AbsPath: path, RelPath: "funcs.go", Lang: "go"}
	chunks, _ := indexer.ChunkFile(entry)

	seen := make(map[string]bool)
	for _, c := range chunks {
		if seen[c.ID] {
			t.Errorf("duplicate chunk ID: %s", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestBuildChunkText(t *testing.T) {
	chunk := indexer.CodeChunk{
		RelPath:    "auth.go",
		SymbolName: "Login",
		SymbolKind: "function",
		LineStart:  10,
		LineEnd:    20,
		Content:    "func Login() {}",
	}
	text := indexer.BuildChunkText(chunk)
	if !strings.Contains(text, "File: auth.go") {
		t.Error("expected file info in chunk text")
	}
	if !strings.Contains(text, "Symbol: Login") {
		t.Error("expected symbol info in chunk text")
	}
}

// ── Merkle tests ──────────────────────────────────────────────────────────────

func TestMerkleTree_DiffDetection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	os.WriteFile(path, []byte("package main"), 0644)

	tree := indexer.NewMerkleTree()
	files := []indexer.FileEntry{{AbsPath: path, RelPath: "main.go", Lang: "go"}}
	tree.Apply(files)

	added, changed, removed := tree.Diff(files)
	if len(added)+len(changed)+len(removed) != 0 {
		t.Errorf("expected empty diff")
	}

	os.WriteFile(path, []byte("package main\nfunc main(){}"), 0644)
	added, changed, removed = tree.Diff(files)
	if len(changed) != 1 {
		t.Errorf("expected 1 changed file")
	}

	tree.Apply(files)
	empty := []indexer.FileEntry{}
	_, _, removed = tree.Diff(empty)
	if len(removed) != 1 {
		t.Errorf("expected 1 removed file")
	}
}

func TestMerkleTree_RootDeterminism(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a"), 0644)

	tree := indexer.NewMerkleTree()
	files := []indexer.FileEntry{{AbsPath: filepath.Join(dir, "a.go"), RelPath: "a.go", Lang: "go"}}
	tree.Apply(files)

	if tree.Root() != tree.Root() {
		t.Error("Root() not deterministic")
	}
}

func TestMerkleTree_SetAndRemove(t *testing.T) {
	tree := indexer.NewMerkleTree()
	tree.Set("a.go", "hash1")
	tree.Set("b.go", "hash2")

	r1 := tree.Root()
	tree.Remove("b.go")
	r2 := tree.Root()

	if r1 == r2 {
		t.Error("Root should change after Remove")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Search Tests (RRF)
// ══════════════════════════════════════════════════════════════════════════════

func TestRRFFuse_BothLists(t *testing.T) {
	bm25 := []search.BM25Input{
		{ID: "a", RelPath: "a.go", SymbolName: "FuncA", BM25Rank: 0},
		{ID: "b", RelPath: "b.go", SymbolName: "FuncB", BM25Rank: 1},
	}
	vec := []search.VecInput{
		{ID: "b", RelPath: "b.go", SymbolName: "FuncB", Score: 0.95},
		{ID: "c", RelPath: "c.go", SymbolName: "FuncC", Score: 0.80},
	}
	results := search.RRFFuse(bm25, vec, 5)
	if len(results) == 0 {
		t.Fatal("expected results")
	}

	bIdx, cIdx := -1, -1
	for i, r := range results {
		if r.ID == "b" {
			bIdx = i
		}
		if r.ID == "c" {
			cIdx = i
		}
	}
	if bIdx < 0 || cIdx < 0 {
		t.Fatalf("b or c not found")
	}
	if bIdx > cIdx {
		t.Errorf("expected 'b' to rank above 'c'")
	}
}

func TestRRFFuse_TopK(t *testing.T) {
	var bm25 []search.BM25Input
	for i := 0; i < 10; i++ {
		bm25 = append(bm25, search.BM25Input{ID: string(rune('a' + i)), BM25Rank: i})
	}
	results := search.RRFFuse(bm25, nil, 3)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestRRFFuse_BM25Only(t *testing.T) {
	bm25 := []search.BM25Input{{ID: "x", BM25Rank: 0}}
	results := search.RRFFuse(bm25, nil, 5)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].VecRank != -1 {
		t.Errorf("expected VecRank=-1, got %d", results[0].VecRank)
	}
}

func TestRRFFuse_VecOnly(t *testing.T) {
	vec := []search.VecInput{{ID: "y", Score: 0.9}}
	results := search.RRFFuse(nil, vec, 5)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].BM25Rank != -1 {
		t.Errorf("expected BM25Rank=-1, got %d", results[0].BM25Rank)
	}
}

func TestCosineSimilarity_Basic(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{1, 0}
	if got := search.CosineSimilarity(a, b); got < 0.99 {
		t.Errorf("identical vectors: expected ~1, got %f", got)
	}
	c := []float32{0, 1}
	if got := search.CosineSimilarity(a, c); got != 0 {
		t.Errorf("orthogonal: expected 0, got %f", got)
	}
}

func TestCosineSimilarity_Empty(t *testing.T) {
	if got := search.CosineSimilarity(nil, nil); got != 0 {
		t.Errorf("empty: expected 0, got %f", got)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Store Tests
// ══════════════════════════════════════════════════════════════════════════════

func TestStoreDB_CreateAndQuery(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, err := store.Open(projectPath, dbDir)
	if err != nil {
		t.Fatalf("store.Open failed: %v", err)
	}
	defer db.Close()

	row := store.ChunkRow{
		ID: "chunk-001", ProjectPath: projectPath, FilePath: "/tmp/main.go",
		RelPath: "main.go", Lang: "go", SymbolName: "main", SymbolKind: "function",
		Content: "func main() { fmt.Println(\"hello\") }", LineStart: 1, LineEnd: 3, FileHash: "abc123",
	}
	if err := db.UpsertChunk(row); err != nil {
		t.Fatalf("UpsertChunk failed: %v", err)
	}

	results, err := db.BM25Search("main*", 5)
	if err != nil {
		t.Fatalf("BM25Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected BM25 result")
	}

	emb := make([]float32, 1536)
	emb[0] = 1.0
	db.UpsertEmbedding("chunk-001", emb)

	embs, _ := db.GetAllEmbeddings(0)
	if len(embs) == 0 {
		t.Error("expected 1 embedding")
	}
}

func TestStoreDB_GetChunksByIDs(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, _ := store.Open(projectPath, dbDir)
	defer db.Close()

	db.UpsertChunk(store.ChunkRow{ID: "c1", ProjectPath: projectPath, FilePath: "/f1", RelPath: "f1.go", Lang: "go", Content: "a", FileHash: "h1"})
	db.UpsertChunk(store.ChunkRow{ID: "c2", ProjectPath: projectPath, FilePath: "/f2", RelPath: "f2.go", Lang: "go", Content: "b", FileHash: "h2"})

	chunks, _ := db.GetChunksByIDs([]string{"c1", "c2"})
	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(chunks))
	}

	// Empty IDs
	chunks, _ = db.GetChunksByIDs(nil)
	if len(chunks) != 0 {
		t.Errorf("expected 0 for nil IDs")
	}
}

func TestStoreDB_FileHashes(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, _ := store.Open(projectPath, dbDir)
	defer db.Close()

	db.UpsertChunk(store.ChunkRow{ID: "c1", ProjectPath: projectPath, FilePath: "/f", RelPath: "f.go", Lang: "go", Content: "x", FileHash: "hash123"})

	hashes, _ := db.GetFileHashes()
	if hashes["f.go"] != "hash123" {
		t.Errorf("expected hash123, got %s", hashes["f.go"])
	}
}

func TestStoreDB_UpdateMeta(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, _ := store.Open(projectPath, dbDir)
	defer db.Close()

	db.UpdateMeta(42, "rootHash")
	total, root, _ := db.GetMeta()
	if total != 42 || root != "rootHash" {
		t.Errorf("expected 42/rootHash, got %d/%s", total, root)
	}
}

func TestStoreDB_DeleteByFile(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, _ := store.Open(projectPath, dbDir)
	defer db.Close()

	db.UpsertChunk(store.ChunkRow{ID: "c1", ProjectPath: projectPath, FilePath: "/f.go", RelPath: "f.go", Lang: "go", Content: "x", FileHash: "h"})
	db.UpsertSymbol(store.SymbolRow{ID: "s1", ProjectPath: projectPath, FilePath: "/f.go", RelPath: "f.go", Name: "Foo", Kind: "function", Lang: "go"})

	db.DeleteByFile("/f.go")

	chunks, _ := db.GetChunksByIDs([]string{"c1"})
	if len(chunks) != 0 {
		t.Error("expected 0 chunks after DeleteByFile")
	}
	count, _ := db.GetSymbolCount()
	if count != 0 {
		t.Error("expected 0 symbols after DeleteByFile")
	}
}

func TestStoreDB_SymbolsCRUD(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, _ := store.Open(projectPath, dbDir)
	defer db.Close()

	sym := store.SymbolRow{
		ID: "sym-001", ProjectPath: projectPath, FilePath: "/tmp/auth.go",
		RelPath: "auth.go", Name: "HandleLogin", Kind: "function",
		Signature: "(w http.ResponseWriter, r *http.Request)", Doc: "HandleLogin processes login",
		LineStart: 10, LineEnd: 45, ParentID: "", Lang: "go",
	}
	db.UpsertSymbol(sym)

	results, _ := db.SearchSymbolsByName("Login", 10)
	if len(results) == 0 {
		t.Error("expected symbol search result")
	}

	s, _ := db.GetSymbolByName("HandleLogin")
	if s == nil {
		t.Fatal("expected symbol, got nil")
	}
	if s.Signature != "(w http.ResponseWriter, r *http.Request)" {
		t.Errorf("unexpected signature: %s", s.Signature)
	}

	count, _ := db.GetSymbolCount()
	if count != 1 {
		t.Errorf("expected 1 symbol, got %d", count)
	}

	// Non-existent symbol
	s2, _ := db.GetSymbolByName("NonExistent")
	if s2 != nil {
		t.Error("expected nil for non-existent symbol")
	}
}

func TestStoreDB_ClearProject(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, _ := store.Open(projectPath, dbDir)
	defer db.Close()

	db.UpsertChunk(store.ChunkRow{ID: "x", ProjectPath: projectPath, FilePath: "/f", RelPath: "f.go", Lang: "go", Content: "test", FileHash: "h"})
	db.UpsertSymbol(store.SymbolRow{ID: "s1", ProjectPath: projectPath, FilePath: "/f", RelPath: "f.go", Name: "Foo", Kind: "function", Lang: "go"})
	db.UpdateMeta(1, "root")

	db.ClearProject()

	total, _, _ := db.GetMeta()
	if total != 0 {
		t.Errorf("expected 0 chunks after clear, got %d", total)
	}
	count, _ := db.GetSymbolCount()
	if count != 0 {
		t.Errorf("expected 0 symbols after clear, got %d", count)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Main handler tests
// ══════════════════════════════════════════════════════════════════════════════

func TestBuildFTSQuery(t *testing.T) {
	cases := map[string]string{
		"hello world":  "hello* OR world*",
		"func_name":    "funcname*",
		"":             "",
		"single":       "single*",
		"with  spaces": "with* OR spaces*",
	}
	for input, want := range cases {
		got := buildFTSQuery(input)
		if got != want {
			t.Errorf("buildFTSQuery(%q) = %q, want %q", input, got, want)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
