// Package indexer provides codebase file walking, language detection, and include/exclude filtering.
package indexer

import (
	"os"
	"path/filepath"
	"strings"
)

// defaultIgnoreDirs are directories always excluded from indexing.
var defaultIgnoreDirs = map[string]bool{
	".git": true, ".svn": true, "node_modules": true, "vendor": true,
	".agents": true, "dist": true, "build": true, "out": true,
	"__pycache__": true, ".venv": true, "venv": true, ".tox": true,
	"coverage": true, ".nyc_output": true, "tmp": true, ".tmp": true,
}

// defaultExtensions are the file extensions indexed by default.
var defaultExtensions = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
	".py": true, ".rs": true, ".java": true, ".cpp": true, ".c": true,
	".cs": true, ".rb": true, ".php": true, ".swift": true, ".kt": true,
	".md": true, ".sql": true, ".sh": true, ".yaml": true, ".yml": true,
	".json": true, ".toml": true,
}

// WalkerConfig controls which files are included or excluded.
type WalkerConfig struct {
	// Extensions to include (e.g. [".go", ".ts"]). Empty = use defaults.
	Extensions []string
	// IgnorePatterns are directory or file name fragments to skip.
	IgnorePatterns []string
	// MaxFileSizeBytes skips files larger than this (default 512KB).
	MaxFileSizeBytes int64
}

// FileEntry is a file discovered by the walker.
type FileEntry struct {
	AbsPath   string
	RelPath   string // relative to project root
	Lang      string // normalised language name
	SizeBytes int64
}

// Walk traverses projectRoot and returns all indexable FileEntry items.
func Walk(projectRoot string, cfg WalkerConfig) ([]FileEntry, error) {
	if cfg.MaxFileSizeBytes == 0 {
		cfg.MaxFileSizeBytes = 512 * 1024 // 512 KB default
	}

	extSet := buildExtSet(cfg.Extensions)
	ignoreSet := buildIgnoreSet(cfg.IgnorePatterns)

	var entries []FileEntry

	err := filepath.WalkDir(projectRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		name := d.Name()

		// Skip hidden dirs and default ignore dirs
		if d.IsDir() {
			if strings.HasPrefix(name, ".") || defaultIgnoreDirs[name] || ignoreSet[name] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if !extSet[ext] {
			return nil
		}

		// Skip ignored file patterns
		if ignoreSet[name] {
			return nil
		}

		info, err := d.Info()
		if err != nil || info.Size() > cfg.MaxFileSizeBytes {
			return nil
		}

		rel, _ := filepath.Rel(projectRoot, path)
		entries = append(entries, FileEntry{
			AbsPath:   path,
			RelPath:   rel,
			Lang:      LangFromExt(ext),
			SizeBytes: info.Size(),
		})
		return nil
	})

	return entries, err
}

// LangFromExt maps a file extension to a normalised language name.
func LangFromExt(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cpp", ".c", ".h":
		return "c"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".md":
		return "markdown"
	case ".sql":
		return "sql"
	case ".sh":
		return "shell"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	default:
		return "text"
	}
}

func buildExtSet(exts []string) map[string]bool {
	if len(exts) == 0 {
		return defaultExtensions
	}
	s := make(map[string]bool, len(exts))
	for _, e := range exts {
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		s[strings.ToLower(e)] = true
	}
	return s
}

func buildIgnoreSet(patterns []string) map[string]bool {
	s := make(map[string]bool, len(patterns))
	for _, p := range patterns {
		s[p] = true
	}
	return s
}
