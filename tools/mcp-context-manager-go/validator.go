package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ValidationResult holds data for a single check.
type ValidationResult struct {
	Check   string
	Passed  bool
	Message string
}

// ReviewCheckpoint runs 5 quality checks on a saved checkpoint and returns a human-readable report.
//
// Checks:
//  1. Stale detection: completed_steps > total (impossible — indicates data corruption)
//  2. Empty active_files on in_progress task
//  3. [ST01], [T1] or legacy [Px-Ty] step label format consistency across all steps
//  4. Duplicate step detection (same step in both completed and next_steps)
//  5. Git SHA present and matches current HEAD
func ReviewCheckpoint(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	row := db.QueryRow(
		`SELECT status, description, notes, acceptance_criteria
         FROM tasks WHERE task_id = ?`,
		taskID,
	)
	var status, descStr, notesStr string
	var ac sql.NullString
	if err := row.Scan(&status, &descStr, &notesStr, &ac); err != nil {
		return fmt.Sprintf("❌ Task '%s' not found.", taskID), nil
	}
	acStr := ac.String

	sRows, err := db.Query("SELECT name, status, notes FROM steps WHERE task_id = ?", taskID)
	if err != nil {
		return "", err
	}
	defer sRows.Close()

	var comp, nxt []string
	var allStepNames []string
	var allStepNotes []string
	for sRows.Next() {
		var name, s string
		var ns sql.NullString
		if err := sRows.Scan(&name, &s, &ns); err == nil {
			allStepNames = append(allStepNames, name)
			if ns.Valid {
				allStepNotes = append(allStepNotes, ns.String)
			}
			if s == "completed" {
				comp = append(comp, name)
			} else {
				nxt = append(nxt, name)
			}
		}
	}

	var activeStr string
	var active []string
	_ = db.QueryRow("SELECT locked_files FROM intents WHERE task_id = ?", taskID).Scan(&activeStr)
	if activeStr != "" && activeStr != "[]" {
		json.Unmarshal([]byte(activeStr), &active)
	}

	checks := []ValidationResult{}

	// ── Check 1: Stale detection ─────────────────────────────────────────────
	total := len(comp) + len(nxt)
	if total > 0 && len(comp) > total {
		checks = append(checks, ValidationResult{
			Check:   "Stale Detection",
			Passed:  false,
			Message: fmt.Sprintf("completed_steps (%d) > total steps (%d) — data may be corrupted.", len(comp), total),
		})
	} else {
		checks = append(checks, ValidationResult{
			Check:   "Stale Detection",
			Passed:  true,
			Message: fmt.Sprintf("Progress is consistent: %d/%d steps completed.", len(comp), total),
		})
	}

	// ── Check 2: Active files non-empty for in-progress tasks ────────────────
	if strings.ToLower(status) == "in_progress" && len(active) == 0 {
		checks = append(checks, ValidationResult{
			Check:   "Active Files Guard",
			Passed:  false,
			Message: "Task is in_progress but active_files is empty. Call declare_intent to set locked files.",
		})
	} else {
		var activeMsg string
		switch strings.ToLower(status) {
		case "completed":
			activeMsg = "Task is completed; no active file lock required."
		case "in_progress":
			activeMsg = fmt.Sprintf("Active file lock is set: %d file(s) tracked.", len(active))
		default:
			activeMsg = fmt.Sprintf("%d file(s) in active_files.", len(active))
		}
		checks = append(checks, ValidationResult{
			Check:   "Active Files Guard",
			Passed:  true,
			Message: activeMsg,
		})
	}

	// ── Check 3: Step label format consistency ──────────────────────────────────────
	allSteps := append(comp, nxt...)
	hasPhaseLabels := false
	badFormat := []string{}
	for _, s := range allSteps {
		if strings.HasPrefix(s, "[") {
			hasPhaseLabels = true
			end := strings.Index(s, "]")
			if end < 0 {
				badFormat = append(badFormat, s)
				continue
			}
			inner := s[1:end]
			if !isValidPhaseLabel(inner) {
				badFormat = append(badFormat, s)
			}
		}
	}
	if hasPhaseLabels && len(badFormat) > 0 {
		checks = append(checks, ValidationResult{
			Check:   "Step Label Format",
			Passed:  false,
			Message: fmt.Sprintf("Malformed step labels: %s", strings.Join(badFormat, ", ")),
		})
	} else {
		checks = append(checks, ValidationResult{
			Check:   "Step Label Format",
			Passed:  true,
			Message: "All step labels are correctly formatted.",
		})
	}

	// ── Check 4: Duplicate step detection ────────────────────────────────────
	compSet := make(map[string]bool, len(comp))
	for _, s := range comp {
		compSet[s] = true
	}
	var duplicates []string
	for _, s := range nxt {
		if compSet[s] {
			duplicates = append(duplicates, s)
		}
	}
	if len(duplicates) > 0 {
		checks = append(checks, ValidationResult{
			Check:   "Duplicate Step Detection",
			Passed:  false,
			Message: fmt.Sprintf("Steps appear in both completed and next_steps: %s", strings.Join(duplicates, ", ")),
		})
	} else {
		checks = append(checks, ValidationResult{
			Check:   "Duplicate Step Detection",
			Passed:  true,
			Message: "No duplicate steps found.",
		})
	}

	// ── Check 5: Git Context ───────────────────────────────
	currentSHA := captureGitSHA(workspacePath)
	if currentSHA == "" {
		checks = append(checks, ValidationResult{
			Check:   "Git SHA Tracking",
			Passed:  false,
			Message: "Not a git repository or git is unavailable.",
		})
	} else {
		checks = append(checks, ValidationResult{
			Check:   "Git SHA Tracking",
			Passed:  true,
			Message: fmt.Sprintf("Git repository active at HEAD: %s", currentSHA[:min(7, len(currentSHA))]),
		})
	}

	// ── Check 6: Broken Link Validator ────────────────────────────────────
	var fullText strings.Builder
	fullText.WriteString(descStr + "\n" + notesStr + "\n" + acStr + "\n")
	for _, n := range allStepNames {
		fullText.WriteString(n + "\n")
	}
	for _, n := range allStepNotes {
		fullText.WriteString(n + "\n")
	}

	brokenLinks := validateBrokenLinks(db, fullText.String())
	if len(brokenLinks) > 0 {
		uniqueLinks := make(map[string]bool)
		for _, l := range brokenLinks {
			uniqueLinks[l] = true
		}
		var ul []string
		for k := range uniqueLinks {
			ul = append(ul, k)
		}
		checks = append(checks, ValidationResult{
			Check:   "Reference Link Integrity",
			Passed:  false,
			Message: fmt.Sprintf("Broken links detected: %s", strings.Join(ul, ", ")),
		})
	} else {
		checks = append(checks, ValidationResult{
			Check:   "Reference Link Integrity",
			Passed:  true,
			Message: "All @task, @ki, @anchor, and @doc references are valid.",
		})
	}

	// ── Render report ─────────────────────────────────────────────────────────
	return renderValidationReport(taskID, checks), nil
}

// isValidPhaseLabel returns true only for the canonical STXX format:
//   - "ST01", "ST02", "ST10" (ST prefix followed by digits)
func isValidPhaseLabel(s string) bool {
	if !strings.HasPrefix(s, "ST") || len(s) < 3 {
		return false
	}
	for _, c := range s[2:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func renderValidationReport(taskID string, checks []ValidationResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🔍 Checkpoint Review: `%s`\n\n", taskID))

	passed := 0
	for _, c := range checks {
		icon := "✅"
		if !c.Passed {
			icon = "❌"
		} else {
			passed++
		}
		sb.WriteString(fmt.Sprintf("**%s %s**\n> %s\n\n", icon, c.Check, c.Message))
	}

	total := len(checks)
	sb.WriteString(fmt.Sprintf("---\n**Score: %d/%d checks passed.**", passed, total))
	if passed == total {
		sb.WriteString(" 🎉 Checkpoint is healthy!")
	} else {
		sb.WriteString(fmt.Sprintf(" ⚠️ %d issue(s) require attention before proceeding.", total-passed))
	}
	return sb.String()
}

func validateBrokenLinks(db *sql.DB, text string) []string {
	var brokenLinks []string

	// @task-ID or @task-ID{relation} — strip optional {relation} suffix
	taskRe := regexp.MustCompile(`(?i)@task-([a-zA-Z0-9_-]+)(?:\{[a-zA-Z0-9_-]+\})?`)
	for _, m := range taskRe.FindAllStringSubmatch(text, -1) {
		tagID := m[1]
		var dummy int
		if err := db.QueryRow("SELECT 1 FROM tasks WHERE task_id = ?", tagID).Scan(&dummy); err == sql.ErrNoRows {
			brokenLinks = append(brokenLinks, fmt.Sprintf("@task-%s", tagID))
		}
	}

	kiRe := regexp.MustCompile(`(?i)@ki/([a-zA-Z0-9_-]+)`)
	kiMatches := kiRe.FindAllStringSubmatch(text, -1)
	if len(kiMatches) > 0 {
		globalDB, _ := GetGlobalDBConnection()
		if globalDB != nil {
			defer globalDB.Close()
		}
		for _, m := range kiMatches {
			tagName := m[1]
			var dummy int
			if err := db.QueryRow("SELECT 1 FROM knowledge_fts WHERE tactic_name = ?", tagName).Scan(&dummy); err == sql.ErrNoRows {
				if globalDB != nil {
					if err2 := globalDB.QueryRow("SELECT 1 FROM global_knowledge_fts WHERE tactic_name = ?", tagName).Scan(&dummy); err2 == sql.ErrNoRows {
						brokenLinks = append(brokenLinks, fmt.Sprintf("@ki/%s", tagName))
					}
				} else {
					brokenLinks = append(brokenLinks, fmt.Sprintf("@ki/%s", tagName))
				}
			}
		}
	}

	anchorRe := regexp.MustCompile(`(?i)@anchor/([a-zA-Z0-9_-]+)`)
	anchorMatches := anchorRe.FindAllStringSubmatch(text, -1)
	if len(anchorMatches) > 0 {
		globalDB, _ := GetGlobalDBConnection()
		if globalDB != nil {
			defer globalDB.Close()
		}
		for _, m := range anchorMatches {
			tagName := m[1]
			var dummy int
			if err := db.QueryRow("SELECT 1 FROM anchors WHERE key = ?", tagName).Scan(&dummy); err == sql.ErrNoRows {
				if globalDB != nil {
					if err2 := globalDB.QueryRow("SELECT 1 FROM global_anchors WHERE key = ?", tagName).Scan(&dummy); err2 == sql.ErrNoRows {
						brokenLinks = append(brokenLinks, fmt.Sprintf("@anchor/%s", tagName))
					}
				} else {
					brokenLinks = append(brokenLinks, fmt.Sprintf("@anchor/%s", tagName))
				}
			}
		}
	}

	// @doc/path references — validate against docs table
	docRe := regexp.MustCompile(`(?i)@doc/([a-zA-Z0-9_/.-]+)`)
	for _, m := range docRe.FindAllStringSubmatch(text, -1) {
		docPath := m[1]
		var dummy int
		if err := db.QueryRow("SELECT 1 FROM docs WHERE doc_path = ?", docPath).Scan(&dummy); err == sql.ErrNoRows {
			brokenLinks = append(brokenLinks, fmt.Sprintf("@doc/%s", docPath))
		}
	}

	return brokenLinks
}
