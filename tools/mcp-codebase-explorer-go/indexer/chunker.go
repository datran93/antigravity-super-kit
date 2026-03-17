// Package indexer provides AST-based and line-based code chunking.
// Enhanced: uses Tree-sitter parser for multi-language AST-aware chunking.
package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	tsparser "mcp-codebase-explorer-go/parser"
)

// CodeChunk represents one indexable unit of source code.
type CodeChunk struct {
	ID         string // sha256(filePath + ":" + symbolName + ":" + lineStart)
	FilePath   string // absolute path
	RelPath    string // relative path (for display)
	Lang       string
	SymbolName string // function/class/method name or "" for file-level
	SymbolKind string // "function" | "method" | "struct" | "interface" | "block"
	Content    string
	LineStart  int
	LineEnd    int
	FileHash   string // sha256 of file content (for Merkle diff)
}

// ChunkFile splits a source file into CodeChunks.
// Uses Go native AST for .go files, Tree-sitter for Python/JS/TS/TSX,
// and falls back to line-window chunking for other languages.
func ChunkFile(entry FileEntry) ([]CodeChunk, error) {
	data, err := os.ReadFile(entry.AbsPath)
	if err != nil {
		return nil, err
	}
	fileHash := hashBytes(data)

	switch entry.Lang {
	case "go":
		return chunkGo(entry, data, fileHash)
	case "python", "javascript", "typescript":
		return chunkByTreeSitter(entry, data, fileHash)
	default:
		return chunkByLines(entry, data, fileHash, 40, 10)
	}
}

// ── Go AST chunker ────────────────────────────────────────────────────────────

func chunkGo(entry FileEntry, src []byte, fileHash string) ([]CodeChunk, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, entry.AbsPath, src, parser.ParseComments)
	if err != nil {
		return chunkByLines(entry, src, fileHash, 40, 10)
	}

	lines := strings.Split(string(src), "\n")
	var chunks []CodeChunk

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			startLine := fset.Position(d.Pos()).Line
			endLine := fset.Position(d.End()).Line

			symbolName := d.Name.Name
			kind := "function"
			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = "method"
				if rt := receiverType(d.Recv.List[0].Type); rt != "" {
					symbolName = rt + "." + symbolName
				}
			}

			content := extractLines(lines, startLine, endLine)
			chunks = append(chunks, makeChunk(entry, fileHash, symbolName, kind, content, startLine, endLine))

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					startLine := fset.Position(s.Pos()).Line
					endLine := fset.Position(s.End()).Line
					kind := "struct"
					if _, ok := s.Type.(*ast.InterfaceType); ok {
						kind = "interface"
					}
					content := extractLines(lines, startLine, endLine)
					chunks = append(chunks, makeChunk(entry, fileHash, s.Name.Name, kind, content, startLine, endLine))
				}
			}
		}
	}

	if len(chunks) == 0 {
		return chunkByLines(entry, src, fileHash, 40, 10)
	}
	return chunks, nil
}

func receiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	}
	return ""
}

// ── Tree-sitter AST chunker (Python, JS, TS, TSX) ───────────────────────────

func chunkByTreeSitter(entry FileEntry, src []byte, fileHash string) ([]CodeChunk, error) {
	// Map indexer lang names to parser lang names
	parserLang := entry.Lang
	if parserLang == "typescript" {
		ext := strings.ToLower(entry.RelPath)
		if strings.HasSuffix(ext, ".tsx") {
			parserLang = "tsx"
		}
	}

	// Check if parser supports this language
	if _, ok := tsparser.Parsers[parserLang]; !ok {
		return chunkByLines(entry, src, fileHash, 40, 10)
	}

	_, nodes := tsparser.ParseAndExtract(entry.AbsPath, "", parserLang)
	if len(nodes) == 0 {
		return chunkByLines(entry, src, fileHash, 40, 10)
	}

	lines := strings.Split(string(src), "\n")
	var chunks []CodeChunk

	// Tree-sitter nodes don't carry exact line info in our current NodeResult.
	// For now, use the full file content per symbol as a single chunk.
	// This is still better than arbitrary line windows because we have symbol names.
	for _, n := range nodes {
		if n.Level > 0 {
			continue // skip nested (methods inside classes are handled by their parent)
		}
		// Create a chunk per top-level symbol
		content := strings.Join(lines, "\n") // fallback: full file
		// Try to find the symbol in the source to get its boundaries
		symbolContent := findSymbolInSource(string(src), n.Name, n.Type)
		if symbolContent != "" {
			content = symbolContent
		}
		chunks = append(chunks, makeChunk(entry, fileHash, n.Name, n.Type, content, 1, len(lines)))
	}

	if len(chunks) == 0 {
		return chunkByLines(entry, src, fileHash, 40, 10)
	}
	return chunks, nil
}

// findSymbolInSource tries to find a symbol definition in source code and extract its content.
// This is a simple heuristic — finds the line containing the definition keyword + name.
func findSymbolInSource(src, name, kind string) string {
	lines := strings.Split(src, "\n")
	patterns := getDefinitionPatterns(name, kind)

	for _, pattern := range patterns {
		for startIdx, line := range lines {
			if strings.Contains(line, pattern) {
				// Find the end of the block by counting braces/indentation
				endIdx := findBlockEnd(lines, startIdx, kind)
				if endIdx > startIdx {
					return strings.Join(lines[startIdx:endIdx+1], "\n")
				}
			}
		}
	}
	return ""
}

// getDefinitionPatterns returns search patterns for a symbol definition.
func getDefinitionPatterns(name, kind string) []string {
	switch kind {
	case "function":
		return []string{"func " + name, "def " + name, "function " + name}
	case "class":
		return []string{"class " + name}
	case "method":
		return []string{"def " + name, name + "("}
	case "interface":
		return []string{"interface " + name}
	case "type":
		return []string{"type " + name}
	default:
		return []string{name}
	}
}

// findBlockEnd finds the end of a code block starting at startIdx.
func findBlockEnd(lines []string, startIdx int, kind string) int {
	if startIdx >= len(lines) {
		return startIdx
	}

	// Simple brace counting for C-style languages
	braceCount := 0
	foundOpening := false
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		for _, ch := range line {
			if ch == '{' {
				braceCount++
				foundOpening = true
			} else if ch == '}' {
				braceCount--
			}
		}
		if foundOpening && braceCount <= 0 {
			return i
		}
	}

	// Python-style: use indentation
	if !foundOpening && startIdx+1 < len(lines) {
		baseIndent := countIndent(lines[startIdx])
		for i := startIdx + 1; i < len(lines); i++ {
			line := lines[i]
			if strings.TrimSpace(line) == "" {
				continue
			}
			if countIndent(line) <= baseIndent {
				return i - 1
			}
		}
		return len(lines) - 1
	}

	// Fallback: return 30 lines after start
	end := startIdx + 30
	if end >= len(lines) {
		end = len(lines) - 1
	}
	return end
}

// countIndent returns the number of leading spaces/tabs.
func countIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

// ── Line-window fallback chunker ─────────────────────────────────────────────

// chunkByLines splits content into overlapping windows of windowSize lines.
func chunkByLines(entry FileEntry, src []byte, fileHash string, windowSize, overlap int) ([]CodeChunk, error) {
	lines := strings.Split(string(src), "\n")
	if len(lines) == 0 {
		return nil, nil
	}

	var chunks []CodeChunk
	step := windowSize - overlap
	if step <= 0 {
		step = windowSize
	}

	for start := 0; start < len(lines); start += step {
		end := start + windowSize
		if end > len(lines) {
			end = len(lines)
		}
		content := strings.Join(lines[start:end], "\n")
		if strings.TrimSpace(content) == "" {
			continue
		}
		symbolName := fmt.Sprintf("lines_%d_%d", start+1, end)
		chunks = append(chunks, makeChunk(entry, fileHash, symbolName, "block", content, start+1, end))
		if end >= len(lines) {
			break
		}
	}
	return chunks, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func makeChunk(entry FileEntry, fileHash, symbolName, kind, content string, lineStart, lineEnd int) CodeChunk {
	id := hashStr(entry.AbsPath + ":" + symbolName + ":" + fmt.Sprintf("%d", lineStart))
	return CodeChunk{
		ID:         id,
		FilePath:   entry.AbsPath,
		RelPath:    entry.RelPath,
		Lang:       entry.Lang,
		SymbolName: symbolName,
		SymbolKind: kind,
		Content:    content,
		LineStart:  lineStart,
		LineEnd:    lineEnd,
		FileHash:   fileHash,
	}
}

func extractLines(lines []string, start, end int) string {
	s := start - 1
	e := end
	if s < 0 {
		s = 0
	}
	if e > len(lines) {
		e = len(lines)
	}
	return strings.Join(lines[s:e], "\n")
}

func hashBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hashStr(s string) string {
	return hashBytes([]byte(s))
}
