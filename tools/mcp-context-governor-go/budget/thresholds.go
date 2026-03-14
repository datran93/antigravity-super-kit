// Package budget provides warning threshold levels for context budget.
package budget

import "fmt"

// Level represents a warning severity level.
type Level string

const (
	LevelOK       Level = "ok"       // < 60%
	LevelWarning  Level = "warning"  // 60–79%
	LevelCritical Level = "critical" // 80–94%
	LevelOverflow Level = "overflow" // ≥ 95%
)

// ThresholdResult is the result of evaluating usage against thresholds.
type ThresholdResult struct {
	Level       Level
	UsedTokens  int
	MaxTokens   int
	UsedPercent float64
	Message     string
	Action      string // suggested action
}

// Evaluate determines the threshold level for a given usage.
func Evaluate(usedTokens, maxTokens int) ThresholdResult {
	if maxTokens <= 0 {
		maxTokens = 100000
	}
	pct := float64(usedTokens) / float64(maxTokens) * 100.0

	var level Level
	var message, action string

	switch {
	case pct >= 95:
		level = LevelOverflow
		message = fmt.Sprintf("🔴 OVERFLOW: context at %.0f%% (used %d / %d tokens)", pct, usedTokens, maxTokens)
		action = "Call compact_memory IMMEDIATELY to flush context before continuing."
	case pct >= 80:
		level = LevelCritical
		message = fmt.Sprintf("🟠 CRITICAL: context at %.0f%% (used %d / %d tokens)", pct, usedTokens, maxTokens)
		action = "Plan a compact_memory call after the current task step completes."
	case pct >= 60:
		level = LevelWarning
		message = fmt.Sprintf("🟡 WARNING: context at %.0f%% (used %d / %d tokens)", pct, usedTokens, maxTokens)
		action = "Consider compressing large file reads or deferring non-critical context."
	default:
		level = LevelOK
		message = fmt.Sprintf("🟢 OK: context at %.0f%% (used %d / %d tokens)", pct, usedTokens, maxTokens)
		action = "No action required."
	}

	return ThresholdResult{
		Level:       level,
		UsedTokens:  usedTokens,
		MaxTokens:   maxTokens,
		UsedPercent: pct,
		Message:     message,
		Action:      action,
	}
}

// SuggestCompression returns a list of suggested compression strategies
// based on the current threshold level.
func SuggestCompression(level Level) []string {
	switch level {
	case LevelOverflow:
		return []string{
			"1. Call compact_memory to summarise the current tactic into a KI",
			"2. Remove large file contents from active context",
			"3. Switch to a summary-only view of previous steps",
		}
	case LevelCritical:
		return []string{
			"1. Avoid reading new large files until you compact",
			"2. Summarize completed steps before starting new ones",
			"3. Use grep_search instead of view_file for targeted lookups",
		}
	case LevelWarning:
		return []string{
			"1. Use targeted searches instead of loading full files",
			"2. Prefer view_file with StartLine/EndLine ranges",
		}
	default:
		return []string{"No compression needed at this time."}
	}
}
