// Package indexer provides AST-based and line-based code chunking.
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

// ChunkFile splits a source file into CodeChunks using AST for Go,
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
	default:
		return chunkByLines(entry, data, fileHash, 40, 10) // 40-line windows, 10-line overlap
	}
}

// ── Go AST chunker ────────────────────────────────────────────────────────────

func chunkGo(entry FileEntry, src []byte, fileHash string) ([]CodeChunk, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, entry.AbsPath, src, parser.ParseComments)
	if err != nil {
		// Fall back to line chunking on parse error
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
				// Prepend receiver type
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

	// If no AST nodes found (e.g. only imports), fall back to line chunks
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
