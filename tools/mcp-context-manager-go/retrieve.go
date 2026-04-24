package main

import (
	"database/sql"
	"fmt"
	"strings"
)

// RetrieveContext assembles a context pack from multiple sources: KIs, docs, anchors, and tasks.
// Returns a unified markdown response with citations for the requesting agent.
func RetrieveContext(workspacePath, query, scope string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	if scope == "" {
		scope = "project"
	}

	var sections []string

	// 1. Search KIs via knowledge recall
	kiResult, err := RecallKnowledge(workspacePath, query, scope, 3)
	if err == nil && kiResult != "" && !strings.Contains(kiResult, "No matching") {
		sections = append(sections, "## 📚 Knowledge Items\n\n"+kiResult)
	}

	// 2. Search docs
	docResult := searchDocsForContext(db, query)
	if docResult != "" {
		sections = append(sections, "## 📄 Documentation\n\n"+docResult)
	}

	// 3. Search anchors
	anchorResult := searchAnchorsForContext(db, query, scope, workspacePath)
	if anchorResult != "" {
		sections = append(sections, "## ⚓ Anchors\n\n"+anchorResult)
	}

	// 4. Search active tasks
	taskResult := searchTasksForContext(db, query)
	if taskResult != "" {
		sections = append(sections, "## 📋 Tasks\n\n"+taskResult)
	}

	if len(sections) == 0 {
		return fmt.Sprintf("🔍 No context found matching '%s'.", query), nil
	}

	header := fmt.Sprintf("# 🧠 Context Pack: '%s'\n\n", query)
	return header + strings.Join(sections, "\n---\n"), nil
}

// searchDocsForContext performs a lightweight doc search for the context pack.
func searchDocsForContext(db *sql.DB, query string) string {
	pattern := "%" + query + "%"
	rows, err := db.Query(
		"SELECT doc_path, title, content FROM docs WHERE title LIKE ? OR content LIKE ? LIMIT 3",
		pattern, pattern,
	)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var path, title, content string
		if err := rows.Scan(&path, &title, &content); err != nil {
			continue
		}
		excerpt := content
		if len(excerpt) > 300 {
			excerpt = excerpt[:300] + "..."
		}
		sb.WriteString(fmt.Sprintf("### `@doc/%s` — %s\n%s\n\n", path, title, excerpt))
	}
	return sb.String()
}

// searchAnchorsForContext searches anchors matching the query.
func searchAnchorsForContext(db *sql.DB, query, scope, workspacePath string) string {
	pattern := "%" + query + "%"
	rows, err := db.Query(
		"SELECT key, value, rule FROM anchors WHERE key LIKE ? OR value LIKE ? LIMIT 3",
		pattern, pattern,
	)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var key string
		var value, rule sql.NullString
		if err := rows.Scan(&key, &value, &rule); err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s**: %s", key, value.String))
		if rule.Valid && rule.String != "" {
			sb.WriteString(fmt.Sprintf(" _(rule: %s)_", rule.String))
		}
		sb.WriteString("\n")
	}

	// Also search global anchors if scope is global
	if scope == "global" {
		globalDB, _ := GetGlobalDBConnection()
		if globalDB != nil {
			defer globalDB.Close()
			gRows, err := globalDB.Query(
				"SELECT key, value, rule FROM global_anchors WHERE key LIKE ? OR value LIKE ? LIMIT 3",
				pattern, pattern,
			)
			if err == nil {
				defer gRows.Close()
				for gRows.Next() {
					var key string
					var value, rule sql.NullString
					if err := gRows.Scan(&key, &value, &rule); err == nil {
						sb.WriteString(fmt.Sprintf("- 🌐 **%s**: %s", key, value.String))
						if rule.Valid && rule.String != "" {
							sb.WriteString(fmt.Sprintf(" _(rule: %s)_", rule.String))
						}
						sb.WriteString("\n")
					}
				}
			}
		}
	}

	return sb.String()
}

// searchTasksForContext searches tasks matching the query.
func searchTasksForContext(db *sql.DB, query string) string {
	pattern := "%" + query + "%"
	rows, err := db.Query(
		"SELECT task_id, description, status FROM tasks WHERE description LIKE ? OR task_id LIKE ? LIMIT 3",
		pattern, pattern,
	)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var sb strings.Builder
	for rows.Next() {
		var taskID, desc, status string
		if err := rows.Scan(&taskID, &desc, &status); err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("- `@task-%s` [%s]: %s\n", taskID, strings.ToUpper(status), desc))
	}
	return sb.String()
}
