package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"mcp-codebase-explorer-go/indexer"
	"mcp-codebase-explorer-go/parser"
	"mcp-codebase-explorer-go/search"
	"mcp-codebase-explorer-go/store"
)

// dataDir stores per-project SQLite databases alongside this binary.
var dataDir = filepath.Join(mustExeDir(), ".db")

func mustExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// indexingStatus tracks in-progress indexing jobs.
var (
	statusMu sync.Mutex
	statuses = make(map[string]*indexStatus)
)

type indexStatus struct {
	Total   int
	Indexed int
	Done    bool
	Err     string
}

// ── Tool: index_codebase ──────────────────────────────────────────────────────
// Enhanced: also extracts and persists symbols via Tree-sitter AST.

func handleIndexCodebase(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectPath, _ := args["path"].(string)
	if projectPath == "" {
		return mcp.NewToolResultError("path is required"), nil
	}
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid path: %v", err)), nil
	}

	var extensions []string
	if raw, ok := args["extensions"].([]interface{}); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				extensions = append(extensions, s)
			}
		}
	}
	var ignorePatterns []string
	if raw, ok := args["ignore"].([]interface{}); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				ignorePatterns = append(ignorePatterns, s)
			}
		}
	}

	statusMu.Lock()
	statuses[absPath] = &indexStatus{}
	statusMu.Unlock()

	go runIndexing(absPath, extensions, ignorePatterns)

	return mcp.NewToolResultText(fmt.Sprintf(
		"✅ Indexing started for: %s\nCall get_indexing_status to track progress.", absPath)), nil
}

func runIndexing(projectPath string, extensions, ignorePatterns []string) {
	setStatus(projectPath, 0, 0, false, "")

	cfg := indexer.WalkerConfig{
		Extensions:     extensions,
		IgnorePatterns: ignorePatterns,
	}
	files, err := indexer.Walk(projectPath, cfg)
	if err != nil {
		setStatus(projectPath, 0, 0, true, err.Error())
		return
	}

	setStatus(projectPath, len(files), 0, false, "")

	db, err := store.Open(projectPath, dataDir)
	if err != nil {
		setStatus(projectPath, len(files), 0, true, err.Error())
		return
	}
	defer db.Close()

	// Merkle diff for incremental indexing
	storedHashes, _ := db.GetFileHashes()
	tree := indexer.NewMerkleTree()
	for rel, hash := range storedHashes {
		tree.Set(rel, hash)
	}

	added, changed, _ := tree.Diff(files)
	needsIndexing := make(map[string]bool, len(added)+len(changed))
	for _, r := range added {
		needsIndexing[r] = true
	}
	for _, r := range changed {
		needsIndexing[r] = true
	}

	// Delete stale chunks/symbols for changed files
	for _, rel := range changed {
		for _, f := range files {
			if f.RelPath == rel {
				db.DeleteByFile(f.AbsPath)
				break
			}
		}
	}

	if len(needsIndexing) == 0 {
		tree.Apply(files)
		db.UpdateMeta(len(storedHashes), tree.Root())
		setStatus(projectPath, len(files), len(files), true, "")
		return
	}

	emb, _ := indexer.NewEmbedder()

	totalChunks := 0
	totalSymbols := 0
	for i, f := range files {
		if !needsIndexing[f.RelPath] {
			continue
		}

		chunks, err := indexer.ChunkFile(f)
		if err != nil {
			continue
		}

		// Build texts for embedding
		var texts []string
		for _, c := range chunks {
			texts = append(texts, indexer.BuildChunkText(c))
		}

		var embeddings [][]float32
		if emb != nil && len(texts) > 0 {
			embeddings, _ = emb.EmbedBatch(context.Background(), texts)
		}

		for j, chunk := range chunks {
			row := store.ChunkRow{
				ID:          chunk.ID,
				ProjectPath: projectPath,
				FilePath:    chunk.FilePath,
				RelPath:     chunk.RelPath,
				Lang:        chunk.Lang,
				SymbolName:  chunk.SymbolName,
				SymbolKind:  chunk.SymbolKind,
				Content:     chunk.Content,
				LineStart:   chunk.LineStart,
				LineEnd:     chunk.LineEnd,
				FileHash:    chunk.FileHash,
			}
			db.UpsertChunk(row)
			if embeddings != nil && j < len(embeddings) && len(embeddings[j]) > 0 {
				db.UpsertEmbedding(chunk.ID, embeddings[j])
			}
			totalChunks++
		}

		// Extract and persist symbols via Tree-sitter AST
		symbols := parser.ExtractSymbols(projectPath, f.AbsPath, f.RelPath, f.Lang)
		for _, sym := range symbols {
			db.UpsertSymbol(store.SymbolRow{
				ID:          sym.ID,
				ProjectPath: projectPath,
				FilePath:    sym.FilePath,
				RelPath:     sym.RelPath,
				Name:        sym.Name,
				Kind:        sym.Kind,
				Signature:   sym.Signature,
				Doc:         sym.Doc,
				LineStart:   sym.LineStart,
				LineEnd:     sym.LineEnd,
				ParentID:    sym.ParentID,
				Lang:        sym.Lang,
			})
			totalSymbols++
		}

		setStatus(projectPath, len(files), i+1, false, "")
	}

	tree.Apply(files)
	prevTotal, _, _ := db.GetMeta()
	db.UpdateMeta(prevTotal+totalChunks, tree.Root())
	setStatus(projectPath, len(files), len(files), true, "")
}

// ── Tool: search_code ─────────────────────────────────────────────────────────

func handleSearchCode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	query, _ := args["query"].(string)
	projectPath, _ := args["project_path"].(string)
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}
	if projectPath == "" {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	absPath, _ := filepath.Abs(projectPath)

	topK := 5
	if tk, ok := args["top_k"].(float64); ok && tk > 0 {
		topK = int(tk)
	}
	langFilter, _ := args["lang_filter"].(string)

	db, err := store.Open(absPath, dataDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open index: %v", err)), nil
	}
	defer db.Close()

	// BM25 search
	ftsQuery := buildFTSQuery(query)
	bm25Raw, _ := db.BM25Search(ftsQuery, 20)

	var bm25Inputs []search.BM25Input
	for i, r := range bm25Raw {
		bm25Inputs = append(bm25Inputs, search.BM25Input{
			ID: r.ID, RelPath: r.RelPath, SymbolName: r.SymbolName, BM25Rank: i,
		})
	}

	// Vector search
	var vecInputs []search.VecInput
	emb, _ := indexer.NewEmbedder()
	if emb != nil {
		queryEmbs, err := emb.EmbedBatch(ctx, []string{query})
		if err == nil && len(queryEmbs) > 0 {
			queryEmb := queryEmbs[0]
			allEmbs, _ := db.GetAllEmbeddings(0)

			type scored struct {
				e     store.EmbeddingRow
				score float32
			}
			var scoreList []scored
			for _, e := range allEmbs {
				sim := search.CosineSimilarity(queryEmb, e.Embedding)
				scoreList = append(scoreList, scored{e, sim})
			}
			sort.Slice(scoreList, func(i, j int) bool {
				return scoreList[i].score > scoreList[j].score
			})
			for i, s := range scoreList {
				if i >= 20 {
					break
				}
				vecInputs = append(vecInputs, search.VecInput{
					ID: s.e.ID, RelPath: s.e.RelPath, SymbolName: s.e.SymbolName, Score: s.score,
				})
			}
		}
	}

	// RRF fusion
	fused := search.RRFFuse(bm25Inputs, vecInputs, topK*2)

	// Fetch full chunk data
	ids := make([]string, len(fused))
	for i, f := range fused {
		ids[i] = f.ID
	}
	chunks, _ := db.GetChunksByIDs(ids)
	chunkMap := make(map[string]store.ChunkRow, len(chunks))
	for _, c := range chunks {
		chunkMap[c.ID] = c
	}

	// Build output
	var out []string
	count := 0
	for _, f := range fused {
		c, ok := chunkMap[f.ID]
		if !ok {
			continue
		}
		if langFilter != "" && !strings.EqualFold(c.Lang, langFilter) {
			continue
		}
		snippet := c.Content
		if len(snippet) > 600 {
			snippet = snippet[:600] + "\n... (truncated)"
		}
		out = append(out, fmt.Sprintf(
			"### %s [%s] — %s (lines %d-%d)\n```%s\n%s\n```\n",
			c.SymbolName, c.SymbolKind, c.RelPath, c.LineStart, c.LineEnd, c.Lang, snippet,
		))
		count++
		if count >= topK {
			break
		}
	}

	if len(out) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("🔍 No results found for query: '%s'", query)), nil
	}

	header := fmt.Sprintf("🔍 **Search results for:** `%s` (%d results)\n\n", query, len(out))
	return mcp.NewToolResultText(header + strings.Join(out, "\n---\n")), nil
}

// ── Tool: get_project_architecture ────────────────────────────────────────────

func handleGetProjectArchitecture(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	workspacePath, ok := args["workspace_path"].(string)
	if !ok {
		return mcp.NewToolResultError("workspace_path is required"), nil
	}
	subPath, _ := args["sub_path"].(string)
	maxFiles := 1000
	if mf, ok := args["max_files"].(float64); ok {
		maxFiles = int(mf)
	}
	includeDocs := false
	if id, ok := args["include_docs"].(bool); ok {
		includeDocs = id
	}

	res, err := parser.GetProjectArchitecture(workspacePath, subPath, maxFiles, includeDocs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

// ── Tool: search_symbol ───────────────────────────────────────────────────────

func handleSearchSymbol(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	workspacePath, ok := args["workspace_path"].(string)
	if !ok {
		return mcp.NewToolResultError("workspace_path is required"), nil
	}
	query, ok := args["query"].(string)
	if !ok {
		return mcp.NewToolResultError("query is required"), nil
	}

	// Try DB first (fast path) if project has been indexed
	absPath, _ := filepath.Abs(workspacePath)
	db, err := store.Open(absPath, dataDir)
	if err == nil {
		defer db.Close()
		symbols, err := db.SearchSymbolsByName(query, 50)
		if err == nil && len(symbols) > 0 {
			var results []string
			for _, s := range symbols {
				sig := ""
				if s.Signature != "" {
					sig = " " + s.Signature
				}
				results = append(results, fmt.Sprintf("📍 %s -> [%s] %s%s", s.RelPath, s.Kind, s.Name, sig))
			}
			return mcp.NewToolResultText("🔎 SYMBOL SEARCH RESULTS (from index):\n" + strings.Join(results, "\n")), nil
		}
	}

	// Fallback to AST parsing (slow path)
	res, err := parser.SearchSymbol(workspacePath, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

// ── Tool: find_usages ─────────────────────────────────────────────────────────

func handleFindUsages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	workspacePath, ok := args["workspace_path"].(string)
	if !ok {
		return mcp.NewToolResultError("workspace_path is required"), nil
	}
	symbolName, ok := args["symbol_name"].(string)
	if !ok {
		return mcp.NewToolResultError("symbol_name is required"), nil
	}

	res, err := parser.FindUsages(workspacePath, symbolName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

// ── Tool: context (NEW — P1) ─────────────────────────────────────────────────

func handleContext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	name, _ := args["name"].(string)
	projectPath, _ := args["project_path"].(string)
	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}
	if projectPath == "" {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	absPath, _ := filepath.Abs(projectPath)

	// 1. Get symbol definition from DB
	db, err := store.Open(absPath, dataDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open index: %v", err)), nil
	}
	defer db.Close()

	symbol, _ := db.GetSymbolByName(name)

	// 2. Get usages via grep-based find_usages
	usagesResult, _ := parser.FindUsages(absPath, name)

	// 3. Get related search results
	ftsQuery := buildFTSQuery(name)
	bm25Raw, _ := db.BM25Search(ftsQuery, 10)

	// Build combined output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🔍 Context for `%s`\n\n", name))

	// Symbol definition
	if symbol != nil {
		sb.WriteString("### 📌 Definition\n")
		sb.WriteString(fmt.Sprintf("- **Name**: `%s`\n", symbol.Name))
		sb.WriteString(fmt.Sprintf("- **Kind**: %s\n", symbol.Kind))
		if symbol.Signature != "" {
			sb.WriteString(fmt.Sprintf("- **Signature**: `%s`\n", symbol.Signature))
		}
		sb.WriteString(fmt.Sprintf("- **File**: `%s`\n", symbol.RelPath))
		if symbol.Doc != "" {
			sb.WriteString(fmt.Sprintf("- **Doc**: %s\n", symbol.Doc))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("### ⚠️ No indexed definition found (project may need re-indexing)\n\n")
	}

	// Usages
	sb.WriteString("### 📋 Usages\n")
	sb.WriteString(usagesResult)
	sb.WriteString("\n")

	// Related code
	if len(bm25Raw) > 0 {
		sb.WriteString("\n### 🔗 Related Code Chunks\n")
		ids := make([]string, 0, len(bm25Raw))
		for _, r := range bm25Raw {
			ids = append(ids, r.ID)
		}
		chunks, _ := db.GetChunksByIDs(ids)
		for _, c := range chunks {
			if len(c.Content) > 200 {
				c.Content = c.Content[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("- **%s** [%s] in `%s` (L%d-%d)\n",
				c.SymbolName, c.SymbolKind, c.RelPath, c.LineStart, c.LineEnd))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// ── Tool: get_indexing_status ─────────────────────────────────────────────────

func handleGetIndexingStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectPath, _ := args["project_path"].(string)
	if projectPath == "" {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	absPath, _ := filepath.Abs(projectPath)

	statusMu.Lock()
	st, ok := statuses[absPath]
	statusMu.Unlock()

	if !ok {
		db, err := store.Open(absPath, dataDir)
		if err == nil {
			total, _, _ := db.GetMeta()
			symbolCount, _ := db.GetSymbolCount()
			db.Close()
			if total > 0 {
				out, _ := json.Marshal(map[string]any{
					"status":        "indexed",
					"total_chunks":  total,
					"total_symbols": symbolCount,
					"percent":       100,
				})
				return mcp.NewToolResultText(string(out)), nil
			}
		}
		return mcp.NewToolResultText(`{"status":"not_started","percent":0}`), nil
	}

	pct := 0
	if st.Total > 0 {
		pct = st.Indexed * 100 / st.Total
	}
	state := "indexing"
	if st.Done {
		state = "done"
	}
	if st.Err != "" {
		state = "error"
	}
	out, _ := json.Marshal(map[string]any{
		"status":        state,
		"total_files":   st.Total,
		"indexed_files": st.Indexed,
		"percent":       pct,
		"error":         st.Err,
	})
	return mcp.NewToolResultText(string(out)), nil
}

// ── Tool: clear_index ─────────────────────────────────────────────────────────

func handleClearIndex(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectPath, _ := args["project_path"].(string)
	if projectPath == "" {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	absPath, _ := filepath.Abs(projectPath)

	db, err := store.Open(absPath, dataDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open index: %v", err)), nil
	}
	defer db.Close()

	if err := db.ClearProject(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to clear: %v", err)), nil
	}

	statusMu.Lock()
	delete(statuses, absPath)
	statusMu.Unlock()

	return mcp.NewToolResultText(fmt.Sprintf("✅ Index cleared for: %s", absPath)), nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func setStatus(path string, total, indexed int, done bool, errMsg string) {
	statusMu.Lock()
	defer statusMu.Unlock()
	statuses[path] = &indexStatus{
		Total:   total,
		Indexed: indexed,
		Done:    done,
		Err:     errMsg,
	}
}

func buildFTSQuery(query string) string {
	var tokens []string
	for _, p := range strings.Fields(query) {
		clean := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		}, p)
		if clean != "" {
			tokens = append(tokens, clean+"*")
		}
	}
	return strings.Join(tokens, " OR ")
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	s := server.NewMCPServer("McpCodebaseExplorer", "1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool 1: index_codebase
	s.AddTool(mcp.NewTool("index_codebase",
		mcp.WithDescription("Index a project directory for hybrid semantic + keyword code search.\nAlso extracts AST symbols for fast symbol lookup."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Absolute or relative path to the project root.")),
		mcp.WithArray("extensions", mcp.Items(map[string]interface{}{"type": "string"}),
			mcp.Description("File extensions to include (e.g. [\".go\",\".ts\"]). Default: all common code files.")),
		mcp.WithArray("ignore", mcp.Items(map[string]interface{}{"type": "string"}),
			mcp.Description("Directory or file names to exclude (e.g. [\"dist\",\"generated\"]).")),
	), handleIndexCodebase)

	// Tool 2: search_code
	s.AddTool(mcp.NewTool("search_code",
		mcp.WithDescription("Search the indexed codebase using natural language. Returns ranked code chunks."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Natural language search query.")),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to the indexed project root.")),
		mcp.WithNumber("top_k", mcp.Description("Number of results to return (default 5).")),
		mcp.WithString("lang_filter", mcp.Description("Filter by language (e.g. 'go', 'typescript').")),
	), handleSearchCode)

	// Tool 3: get_project_architecture (from ast-explorer)
	s.AddTool(mcp.NewTool("get_project_architecture",
		mcp.WithDescription("Get a structural overview (AST-based) of the project architecture.\nExtracts Classes, Functions, and Methods with signatures."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("The base absolute path of the project workspace.")),
		mcp.WithString("sub_path", mcp.Description("Optional sub-path within the workspace to limit scope.")),
		mcp.WithNumber("max_files", mcp.Description("Max number of files to process, default 1000.")),
		mcp.WithBoolean("include_docs", mcp.Description("If true, includes the first line of docstrings/comments.")),
	), handleGetProjectArchitecture)

	// Tool 4: search_symbol (from ast-explorer, enhanced with DB)
	s.AddTool(mcp.NewTool("search_symbol",
		mcp.WithDescription("Search for a class or function symbol across the project using AST.\nUseful for finding definitions quickly. Uses indexed DB if available, falls back to AST parsing."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("The base absolute path of the project workspace.")),
		mcp.WithString("query", mcp.Required(), mcp.Description("The symbol name to search for (case-insensitive substring match).")),
	), handleSearchSymbol)

	// Tool 5: find_usages (from ast-explorer)
	s.AddTool(mcp.NewTool("find_usages",
		mcp.WithDescription("Find all usages/references of a symbol (function, class, variable) across the project. Returns file:line references grouped by file."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("The base absolute path of the project workspace.")),
		mcp.WithString("symbol_name", mcp.Required(), mcp.Description("The symbol name to search for (case-insensitive substring match).")),
	), handleFindUsages)

	// Tool 6: context (NEW — P1)
	s.AddTool(mcp.NewTool("context",
		mcp.WithDescription("360-degree view of a code symbol. Shows definition, usages, and related code chunks.\nCombines symbol lookup, usage search, and semantic search in one call."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Symbol name (e.g., 'HandleLogin', 'UserService').")),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to the indexed project root.")),
	), handleContext)

	// Tool 7: get_indexing_status
	s.AddTool(mcp.NewTool("get_indexing_status",
		mcp.WithDescription("Check the indexing status of a project."),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to the project root.")),
	), handleGetIndexingStatus)

	// Tool 8: clear_index
	s.AddTool(mcp.NewTool("clear_index",
		mcp.WithDescription("Clear the search index for a project."),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to the project root.")),
	), handleClearIndex)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
