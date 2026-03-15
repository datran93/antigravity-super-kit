package main

import (
	"encoding/json"
	"fmt"
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
//  3. [T1] or legacy [Px-Ty] step label format consistency across all steps
//  4. Duplicate step detection (same step in both completed and next_steps)
//  5. Git SHA present and matches current HEAD
func ReviewCheckpoint(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	row := db.QueryRow(
		`SELECT status, completed_steps, next_steps, active_files, git_sha
         FROM checkpoints WHERE task_id = ?`,
		taskID,
	)
	var status, compStr, nxtStr, activeStr, storedSHA string
	if err := row.Scan(&status, &compStr, &nxtStr, &activeStr, &storedSHA); err != nil {
		return fmt.Sprintf("❌ Checkpoint '%s' not found.", taskID), nil
	}

	var comp, nxt, active []string
	json.Unmarshal([]byte(compStr), &comp)
	json.Unmarshal([]byte(nxtStr), &nxt)
	if activeStr != "" {
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
		// Contextual success message: completed tasks don't need a file lock.
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
			// Valid formats (primary): [T1], [T2] — legacy: [P0-T1], [P0], [P1-T2]
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

	// ── Check 5: Git SHA present and HEAD match ───────────────────────────────
	currentSHA := captureGitSHA(workspacePath)
	if storedSHA == "" {
		checks = append(checks, ValidationResult{
			Check:   "Git SHA Tracking",
			Passed:  false,
			Message: "No git_sha stored. Run save_checkpoint to capture the current HEAD.",
		})
	} else if currentSHA != "" && storedSHA != currentSHA {
		checks = append(checks, ValidationResult{
			Check:  "Git SHA Tracking",
			Passed: false,
			Message: fmt.Sprintf(
				"Checkpoint SHA (%s) differs from HEAD (%s). Run: git log --oneline %s..HEAD",
				storedSHA[:min(7, len(storedSHA))],
				currentSHA[:min(7, len(currentSHA))],
				storedSHA[:min(7, len(storedSHA))],
			),
		})
	} else {
		shaLabel := "not a git repo"
		if storedSHA != "" {
			shaLabel = storedSHA[:min(7, len(storedSHA))]
		}
		checks = append(checks, ValidationResult{
			Check:   "Git SHA Tracking",
			Passed:  true,
			Message: fmt.Sprintf("Git SHA matches HEAD: %s", shaLabel),
		})
	}

	// ── Render report ─────────────────────────────────────────────────────────
	return renderValidationReport(taskID, checks), nil
}

// isValidPhaseLabel returns true for:
//   - New format: "T1", "T2", "T10" (simple task IDs)
//   - Legacy format: "P0", "P1", "P0-T1", "P2-T3"
func isValidPhaseLabel(s string) bool {
	if len(s) == 0 {
		return false
	}
	// New format: [T1], [T2], [T10]
	if s[0] == 'T' {
		rest := s[1:]
		if len(rest) == 0 {
			return false
		}
		for _, c := range rest {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	// Legacy format: [P0], [P1], [P0-T1], [P2-T3]
	if s[0] != 'P' {
		return false
	}
	dashIdx := strings.Index(s, "-")
	if dashIdx < 0 {
		// "P0", "P1", etc.
		return len(s) >= 2
	}
	// "P0-T1" format
	if dashIdx < 2 {
		return false
	}
	rest := s[dashIdx+1:]
	return len(rest) >= 2 && rest[0] == 'T'
}

// renderValidationReport formats the list of checks as a markdown report.
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
