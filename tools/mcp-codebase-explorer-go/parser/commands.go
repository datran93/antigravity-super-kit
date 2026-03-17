package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// FileTask represents a file to be parsed.
type FileTask struct {
	Folder   string
	Filepath string
	Filename string
	Lang     string
}

// FileResult collects parsed nodes for one file.
type FileResult struct {
	Filename string
	Nodes    []NodeResult
}

// SymbolInfo represents a persisted symbol extracted from AST.
type SymbolInfo struct {
	ID        string
	FilePath  string
	RelPath   string
	Name      string
	Kind      string
	Signature string
	Doc       string
	LineStart int
	LineEnd   int
	ParentID  string
	Lang      string
}

// symbolID generates a deterministic ID for a symbol.
func symbolID(projectPath, relPath, name string, lineStart int) string {
	raw := fmt.Sprintf("%s:%s:%s:%d", projectPath, relPath, name, lineStart)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// GetProjectArchitecture returns a structural overview of the project using AST.
func GetProjectArchitecture(workspacePath, subPath string, maxFiles int, includeDocs bool) (string, error) {
	baseDir := workspacePath
	if subPath != "" {
		baseDir = filepath.Join(workspacePath, subPath)
	}

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return fmt.Sprintf("❌ Path not found: %s", baseDir), nil
	}

	allFiles := GetProjectFiles(workspacePath)
	mainFamily := GetMainLanguageFamily(allFiles)

	folderToFiles := make(map[string][]FileTask)
	for _, f := range allFiles {
		if !strings.HasPrefix(f, baseDir) {
			continue
		}

		ext := filepath.Ext(f)
		family := FamilyFromExt(ext)
		if family != mainFamily || family == "" {
			continue
		}

		relPath, err := filepath.Rel(workspacePath, f)
		if err != nil {
			relPath = f
		}

		folder := filepath.Dir(relPath)
		filename := filepath.Base(relPath)

		lang := LanguageFromExt(ext)
		if lang == "" {
			continue
		}

		folderToFiles[folder] = append(folderToFiles[folder], FileTask{
			Folder:   folder,
			Filepath: f,
			Filename: filename,
			Lang:     lang,
		})
	}

	var sortedFolders []string
	for folder := range folderToFiles {
		sortedFolders = append(sortedFolders, folder)
	}
	sort.Strings(sortedFolders)

	var tasks []FileTask
	for _, folder := range sortedFolders {
		filesInFolder := folderToFiles[folder]
		sort.Slice(filesInFolder, func(i, j int) bool {
			return filesInFolder[i].Filepath < filesInFolder[j].Filepath
		})

		for _, task := range filesInFolder {
			tasks = append(tasks, task)
			if len(tasks) >= maxFiles {
				break
			}
		}
		if len(tasks) >= maxFiles {
			break
		}
	}

	subPathMsg := subPath
	if subPathMsg == "" {
		subPathMsg = "ROOT"
	}

	var output []string
	output = append(output, fmt.Sprintf("🏗 PROJECT ARCHITECTURE AST: %s (Main Lang: %s)\n", subPathMsg, mainFamily))

	var wg sync.WaitGroup
	resultCh := make(chan struct {
		Folder string
		Res    FileResult
	}, len(tasks))

	sem := make(chan struct{}, 8)

	for _, task := range tasks {
		wg.Add(1)
		go func(t FileTask) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			_, nodes := ParseAndExtract(t.Filepath, workspacePath, t.Lang)
			if len(nodes) > 0 {
				resultCh <- struct {
					Folder string
					Res    FileResult
				}{t.Folder, FileResult{Filename: t.Filename, Nodes: nodes}}
			}
		}(task)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	resultsByFolder := make(map[string][]FileResult)
	for res := range resultCh {
		resultsByFolder[res.Folder] = append(resultsByFolder[res.Folder], res.Res)
	}

	for _, folder := range sortedFolders {
		folderResults, ok := resultsByFolder[folder]
		if !ok {
			continue
		}

		displayFolder := folder
		if displayFolder == "" || displayFolder == "." {
			displayFolder = "."
		}
		output = append(output, fmt.Sprintf("📁 %s", displayFolder))

		sort.Slice(folderResults, func(i, j int) bool {
			return folderResults[i].Filename < folderResults[j].Filename
		})

		for _, fileRes := range folderResults {
			output = append(output, fmt.Sprintf("  📄 %s", fileRes.Filename))
			for _, n := range fileRes.Nodes {
				indent := "    " + strings.Repeat("  ", n.Level)
				sig := ""
				if n.Signature != "" {
					sig = " " + n.Signature
				}
				output = append(output, fmt.Sprintf("%s▪ [%s] %s%s", indent, n.Type, n.Name, sig))
				if includeDocs && n.Doc != "" {
					firstLine := strings.Split(n.Doc, "\n")[0]
					output = append(output, fmt.Sprintf("%s  // %s", indent, firstLine))
				}
			}
		}
	}

	if len(tasks) >= maxFiles {
		output = append(output, fmt.Sprintf("\n⚠️ Reached limit of %d files.", maxFiles))
	}

	return strings.Join(output, "\n"), nil
}

// SearchSymbol searches for a symbol name across the project using AST.
func SearchSymbol(workspacePath, query string) (string, error) {
	allFiles := GetProjectFiles(workspacePath)
	mainFamily := GetMainLanguageFamily(allFiles)

	var tasks []FileTask
	for _, filepathStr := range allFiles {
		ext := filepath.Ext(filepathStr)
		family := FamilyFromExt(ext)
		if family != mainFamily || family == "" {
			continue
		}

		lang := LanguageFromExt(ext)
		if lang == "" {
			continue
		}

		tasks = append(tasks, FileTask{
			Filepath: filepathStr,
			Lang:     lang,
		})
	}

	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 8)

	for _, task := range tasks {
		wg.Add(1)
		go func(t FileTask) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			relPath, nodes := ParseAndExtract(t.Filepath, workspacePath, t.Lang)
			if len(nodes) == 0 {
				return
			}

			lowerQuery := strings.ToLower(query)
			for _, n := range nodes {
				if strings.Contains(strings.ToLower(n.Name), lowerQuery) {
					mu.Lock()
					if len(results) < 50 {
						results = append(results, fmt.Sprintf("📍 %s -> [%s] %s%s", relPath, n.Type, n.Name, n.Signature))
					}
					mu.Unlock()
				}
			}
		}(task)
	}

	wg.Wait()

	if len(results) == 0 {
		return fmt.Sprintf("🔍 No symbols matching '%s' found.", query), nil
	}

	return "🔎 SYMBOL SEARCH RESULTS:\n" + strings.Join(results, "\n"), nil
}

// UsageRef represents a single reference to a symbol.
type UsageRef struct {
	File    string
	Line    int
	Content string
}

// FindUsages scans all source files for references to symbolName.
func FindUsages(workspacePath, symbolName string) (string, error) {
	if symbolName == "" {
		return "❌ symbol_name is required.", nil
	}

	allFiles := GetProjectFiles(workspacePath)
	if len(allFiles) == 0 {
		return fmt.Sprintf("❌ No source files found in: %s", workspacePath), nil
	}

	const concurrency = 8
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var usages []UsageRef

	for _, filePath := range allFiles {
		wg.Add(1)
		sem <- struct{}{}
		go func(fp string) {
			defer wg.Done()
			defer func() { <-sem }()

			data, err := os.ReadFile(fp)
			if err != nil {
				return
			}

			lines := strings.Split(string(data), "\n")
			lower := strings.ToLower(symbolName)

			relPath, err := filepath.Rel(workspacePath, fp)
			if err != nil {
				relPath = fp
			}

			for lineNum, line := range lines {
				if strings.Contains(strings.ToLower(line), lower) {
					mu.Lock()
					usages = append(usages, UsageRef{
						File:    relPath,
						Line:    lineNum + 1,
						Content: strings.TrimSpace(line),
					})
					mu.Unlock()
				}
			}
		}(filePath)
	}

	wg.Wait()

	if len(usages) == 0 {
		return fmt.Sprintf("🔍 No usages of '%s' found in %s.", symbolName, workspacePath), nil
	}

	sort.Slice(usages, func(i, j int) bool {
		if usages[i].File != usages[j].File {
			return usages[i].File < usages[j].File
		}
		return usages[i].Line < usages[j].Line
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 **%d usages** of `%s`:\n\n", len(usages), symbolName))

	currentFile := ""
	for _, u := range usages {
		if u.File != currentFile {
			sb.WriteString(fmt.Sprintf("\n### `%s`\n", u.File))
			currentFile = u.File
		}
		sb.WriteString(fmt.Sprintf("  - **L%d**: `%s`\n", u.Line, u.Content))
	}

	return sb.String(), nil
}

// ExtractSymbols parses a file and returns SymbolInfo for DB persistence.
func ExtractSymbols(projectPath, filePath, relPath, lang string) []SymbolInfo {
	_, nodes := ParseAndExtract(filePath, projectPath, lang)
	if len(nodes) == 0 {
		return nil
	}

	var symbols []SymbolInfo
	for _, n := range nodes {
		sym := SymbolInfo{
			ID:        symbolID(projectPath, relPath, n.Name, 0), // lineStart from node not available in NodeResult
			FilePath:  filePath,
			RelPath:   relPath,
			Name:      n.Name,
			Kind:      n.Type,
			Signature: n.Signature,
			Doc:       n.Doc,
			LineStart: 0, // Not tracked in current NodeResult
			LineEnd:   0,
			ParentID:  "",
			Lang:      lang,
		}
		symbols = append(symbols, sym)
	}
	return symbols
}
