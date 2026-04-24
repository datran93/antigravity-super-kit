package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CreateDoc creates or updates a structured documentation entry.
func CreateDoc(workspacePath, docPath, title, content string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if docPath == "" {
		return "", fmt.Errorf("doc_path is required")
	}
	if title == "" {
		return "", fmt.Errorf("title is required")
	}

	now := time.Now().UTC().Format(time.RFC3339)

	_, err = db.Exec(`
		INSERT INTO docs (doc_path, title, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(doc_path) DO UPDATE SET
			title=excluded.title,
			content=excluded.content,
			updated_at=excluded.updated_at
	`, docPath, title, content, now, now)
	if err != nil {
		return "", fmt.Errorf("failed to create doc: %w", err)
	}
	return fmt.Sprintf("📄 Doc '%s' created/updated at @doc/%s", title, docPath), nil
}

// GetDoc retrieves a single doc by its path.
func GetDoc(workspacePath, docPath string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var title, content, createdAt, updatedAt string
	err = db.QueryRow(
		"SELECT title, content, created_at, updated_at FROM docs WHERE doc_path = ?",
		docPath,
	).Scan(&title, &content, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return fmt.Sprintf("⚠️ Doc '@doc/%s' not found.", docPath), nil
	} else if err != nil {
		return "", err
	}

	// Fetch references
	refs := fetchDocReferences(db, "doc", docPath)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 📄 %s\n", title))
	sb.WriteString(fmt.Sprintf("**Path**: `@doc/%s`\n", docPath))
	sb.WriteString(fmt.Sprintf("**Updated**: %s\n\n", updatedAt))
	sb.WriteString(content)

	if refs != "" {
		sb.WriteString("\n\n---\n### 🔗 References\n")
		sb.WriteString(refs)
	}

	return sb.String(), nil
}

// ListDocs returns all docs in the workspace.
func ListDocs(workspacePath string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	rows, err := db.Query("SELECT doc_path, title, updated_at FROM docs ORDER BY updated_at DESC")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("📚 Documentation:\n\n")

	count := 0
	for rows.Next() {
		var path, title, updatedAt string
		if err := rows.Scan(&path, &title, &updatedAt); err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("- `@doc/%s` — **%s** (%s)\n", path, title, updatedAt))
		count++
	}

	if count == 0 {
		return "📚 No docs found. Create one with create_doc.", nil
	}

	return sb.String(), nil
}

// SearchDocs performs a LIKE-based search across doc titles and content.
func SearchDocs(workspacePath, query string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	pattern := "%" + query + "%"
	rows, err := db.Query(
		"SELECT doc_path, title, content FROM docs WHERE title LIKE ? OR content LIKE ? LIMIT 10",
		pattern, pattern,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 Docs matching '%s':\n\n", query))

	count := 0
	for rows.Next() {
		var path, title, content string
		if err := rows.Scan(&path, &title, &content); err != nil {
			continue
		}
		excerpt := content
		if len(excerpt) > 120 {
			excerpt = excerpt[:120] + "..."
		}
		sb.WriteString(fmt.Sprintf("- `@doc/%s` — **%s**: %s\n", path, title, excerpt))
		count++
	}

	if count == 0 {
		return fmt.Sprintf("🔍 No docs found matching '%s'.", query), nil
	}

	return sb.String(), nil
}

// AddDocReference creates a typed reference between two entities.
func AddDocReference(workspacePath, sourceType, sourceID, targetType, targetID, relation string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if relation == "" {
		relation = "references"
	}

	_, err = db.Exec(`
		INSERT INTO doc_references (source_type, source_id, target_type, target_id, relation)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`, sourceType, sourceID, targetType, targetID, relation)
	if err != nil {
		return "", fmt.Errorf("failed to add reference: %w", err)
	}
	return fmt.Sprintf("🔗 Reference added: %s/%s —[%s]→ %s/%s", sourceType, sourceID, relation, targetType, targetID), nil
}

// fetchDocReferences returns a formatted string of all references for a given entity.
func fetchDocReferences(db *sql.DB, entityType, entityID string) string {
	rows, err := db.Query(`
		SELECT target_type, target_id, relation FROM doc_references
		WHERE source_type = ? AND source_id = ?
		UNION ALL
		SELECT source_type, source_id, relation FROM doc_references
		WHERE target_type = ? AND target_id = ?
	`, entityType, entityID, entityType, entityID)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var refs []string
	for rows.Next() {
		var refType, refID, relation string
		if err := rows.Scan(&refType, &refID, &relation); err != nil {
			continue
		}
		refs = append(refs, fmt.Sprintf("- `%s/%s` (%s)", refType, refID, relation))
	}

	return strings.Join(refs, "\n")
}

// ResolveDocPath resolves a @doc/path reference to its content.
// Used by the auto-linking system.
func ResolveDocPath(db *sql.DB, docPath string) string {
	var title, content string
	err := db.QueryRow("SELECT title, content FROM docs WHERE doc_path = ?", docPath).Scan(&title, &content)
	if err != nil {
		return ""
	}
	excerpt := content
	if len(excerpt) > 200 {
		excerpt = excerpt[:200] + "..."
	}
	return fmt.Sprintf("**Doc [%s]**: %s", title, excerpt)
}
