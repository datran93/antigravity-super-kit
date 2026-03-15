package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mcp-codebase-search-go/indexer"
	"mcp-codebase-search-go/search"
	"mcp-codebase-search-go/store"
)

// ── indexer/walker tests ──────────────────────────────────────────────────────

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
		t.Errorf("expected at least 2 files (.go + .md), got %d", len(files))
	}
	// .bin should not be included
	for _, f := range files {
		if strings.HasSuffix(f.RelPath, ".bin") {
			t.Errorf("binary file should be excluded, but got: %s", f.RelPath)
		}
	}
}

func TestWalk_IgnoresNodeModules(t *testing.T) {
	dir := t.TempDir()
	nm := filepath.Join(dir, "node_modules")
	os.MkdirAll(nm, 0755)
	os.WriteFile(filepath.Join(nm, "pkg.js"), []byte("var x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "index.ts"), []byte("export {}"), 0644)

	files, err := indexer.Walk(dir, indexer.WalkerConfig{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.Contains(f.RelPath, "node_modules") {
			t.Errorf("node_modules should be excluded: %s", f.RelPath)
		}
	}
}

func TestLangFromExt(t *testing.T) {
	cases := map[string]string{
		".go":  "go",
		".ts":  "typescript",
		".tsx": "typescript",
		".py":  "python",
		".rs":  "rust",
		".md":  "markdown",
	}
	for ext, want := range cases {
		got := indexer.LangFromExt(ext)
		if got != want {
			t.Errorf("LangFromExt(%q) = %q, want %q", ext, got, want)
		}
	}
}

// ── indexer/chunker tests ─────────────────────────────────────────────────────

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
		t.Error("expected at least 1 chunk for Go file")
	}
	// Verify one chunk has "Hello" as symbol
	found := false
	for _, c := range chunks {
		if c.SymbolName == "Hello" {
			found = true
			if c.SymbolKind != "function" {
				t.Errorf("expected kind=function, got %s", c.SymbolKind)
			}
		}
	}
	if !found {
		t.Error("expected to find chunk with SymbolName='Hello'")
	}
}

func TestChunkFile_FallbackLineWindow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "script.py")
	lines := make([]string, 60)
	for i := range lines {
		lines[i] = "# line"
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)

	entry := indexer.FileEntry{AbsPath: path, RelPath: "script.py", Lang: "python"}
	chunks, err := indexer.ChunkFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Error("expected line-window chunks for Python file")
	}
}

func TestChunkFile_UniqueIDs(t *testing.T) {
	dir := t.TempDir()
	goCode := `package main
func A(){}
func B(){}
func C(){}
`
	path := filepath.Join(dir, "funcs.go")
	os.WriteFile(path, []byte(goCode), 0644)

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

// ── indexer/merkle tests ──────────────────────────────────────────────────────

func TestMerkleTree_DiffDetection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	os.WriteFile(path, []byte("package main"), 0644)

	tree := indexer.NewMerkleTree()
	files := []indexer.FileEntry{{AbsPath: path, RelPath: "main.go", Lang: "go"}}
	tree.Apply(files)

	// No change → empty diff
	added, changed, removed := tree.Diff(files)
	if len(added)+len(changed)+len(removed) != 0 {
		t.Errorf("expected empty diff, got add=%v chg=%v rem=%v", added, changed, removed)
	}

	// Modify file → changed
	os.WriteFile(path, []byte("package main\nfunc main(){}"), 0644)
	added, changed, removed = tree.Diff(files)
	if len(changed) != 1 {
		t.Errorf("expected 1 changed file, got %v", changed)
	}

	// Remove file → removed
	tree.Apply(files) // apply updated hash
	empty := []indexer.FileEntry{}
	added, changed, removed = tree.Diff(empty)
	if len(removed) != 1 {
		t.Errorf("expected 1 removed file, got %v", removed)
	}
}

func TestMerkleTree_RootDeterminism(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a"), 0644)

	tree := indexer.NewMerkleTree()
	files := []indexer.FileEntry{{AbsPath: filepath.Join(dir, "a.go"), RelPath: "a.go", Lang: "go"}}
	tree.Apply(files)

	r1 := tree.Root()
	r2 := tree.Root()
	if r1 != r2 {
		t.Error("Root() not deterministic")
	}
}

// ── search/hybrid (RRF) tests ─────────────────────────────────────────────────

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
	// "b" appears in both lists → should rank higher than "c" which is vec-only
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
		t.Fatalf("b or c not found in results: %+v", results)
	}
	if bIdx > cIdx {
		t.Errorf("expected 'b' (in both lists) to rank above 'c' (vec-only), got b=%d c=%d", bIdx, cIdx)
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

// ── store/db tests ────────────────────────────────────────────────────────────

func TestStoreDB_CreateAndQuery(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, err := store.Open(projectPath, dbDir)
	if err != nil {
		t.Fatalf("store.Open failed: %v", err)
	}
	defer db.Close()

	row := store.ChunkRow{
		ID:          "chunk-001",
		ProjectPath: projectPath,
		FilePath:    "/tmp/main.go",
		RelPath:     "main.go",
		Lang:        "go",
		SymbolName:  "main",
		SymbolKind:  "function",
		Content:     "func main() { fmt.Println(\"hello\") }",
		LineStart:   1,
		LineEnd:     3,
		FileHash:    "abc123",
	}
	if err := db.UpsertChunk(row); err != nil {
		t.Fatalf("UpsertChunk failed: %v", err)
	}

	// BM25 search
	results, err := db.BM25Search("main*", 5)
	if err != nil {
		t.Fatalf("BM25Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected BM25 result for 'main'")
	}

	// Embedding upsert
	emb := make([]float32, 1536)
	emb[0] = 1.0
	if err := db.UpsertEmbedding("chunk-001", emb); err != nil {
		t.Fatalf("UpsertEmbedding failed: %v", err)
	}

	// GetAllEmbeddings
	embs, err := db.GetAllEmbeddings(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(embs) == 0 {
		t.Error("expected 1 embedding")
	}
}

func TestStoreDB_ClearProject(t *testing.T) {
	dbDir := t.TempDir()
	projectPath := t.TempDir()

	db, _ := store.Open(projectPath, dbDir)
	defer db.Close()

	db.UpsertChunk(store.ChunkRow{ID: "x", ProjectPath: projectPath, FilePath: "/f", RelPath: "f.go", Lang: "go", Content: "test", FileHash: "h"})
	db.UpdateMeta(1, "root")

	if err := db.ClearProject(); err != nil {
		t.Fatalf("ClearProject failed: %v", err)
	}

	total, _, _ := db.GetMeta()
	if total != 0 {
		t.Errorf("expected 0 chunks after clear, got %d", total)
	}
}
