package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"mcp-context-governor-go/budget"
)

// govDataDir stores the governor SQLite DB.
var govDataDir = func() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Join(filepath.Dir(exe), ".db")
}()

// ── Tool: get_budget_status ───────────────────────────────────────────────────

func handleGetBudgetStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		sessionID = "default"
	}

	t, err := budget.OpenTracker(govDataDir, sessionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open tracker: %v", err)), nil
	}
	defer t.Close()

	summary, err := t.GetSummary()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get summary: %v", err)), nil
	}

	maxBudget := t.GetMaxBudget()
	result := budget.Evaluate(summary.TotalTokens, maxBudget)

	out, _ := json.Marshal(map[string]any{
		"session_id":   sessionID,
		"used_tokens":  summary.TotalTokens,
		"max_tokens":   maxBudget,
		"used_percent": fmt.Sprintf("%.1f%%", result.UsedPercent),
		"level":        result.Level,
		"message":      result.Message,
		"action":       result.Action,
		"event_count":  summary.EventCount,
	})
	return mcp.NewToolResultText(string(out)), nil
}

// ── Tool: estimate_cost ───────────────────────────────────────────────────────

func handleEstimateCost(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		sessionID = "default"
	}
	tool, _ := args["tool"].(string)

	var totalEstimate int

	// Estimate from text content
	if text, ok := args["text"].(string); ok && text != "" {
		totalEstimate += budget.EstimateTokens(text)
	}

	// Estimate from file list
	if rawFiles, ok := args["files"].([]interface{}); ok {
		var files []string
		for _, f := range rawFiles {
			if s, ok := f.(string); ok {
				files = append(files, s)
			}
		}
		totalEstimate += budget.EstimateContextLoad(files)
	}

	// Guard against negative estimates (should not happen, but defensive)
	if totalEstimate < 0 {
		totalEstimate = 0
	}

	// Record to tracker
	t, err := budget.OpenTracker(govDataDir, sessionID)
	if err == nil {
		t.RecordUsage(tool, totalEstimate, "estimate")
		t.Close()
	}

	out, _ := json.Marshal(map[string]any{
		"session_id":       sessionID,
		"estimated_tokens": totalEstimate,
		"tool":             tool,
		"note":             "1 token ≈ 4 bytes. This is a heuristic estimate.",
	})
	return mcp.NewToolResultText(string(out)), nil
}

// ── Tool: suggest_compression ────────────────────────────────────────────────

func handleSuggestCompression(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		sessionID = "default"
	}

	t, err := budget.OpenTracker(govDataDir, sessionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open tracker: %v", err)), nil
	}
	defer t.Close()

	summary, _ := t.GetSummary()
	maxBudget := t.GetMaxBudget()
	result := budget.Evaluate(summary.TotalTokens, maxBudget)
	suggestions := budget.SuggestCompression(result.Level)

	out, _ := json.Marshal(map[string]any{
		"session_id":  sessionID,
		"level":       result.Level,
		"message":     result.Message,
		"suggestions": suggestions,
	})
	return mcp.NewToolResultText(string(out)), nil
}

// ── Tool: trigger_compact ────────────────────────────────────────────────────

func handleTriggerCompact(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		sessionID = "default"
	}

	reason, _ := args["reason"].(string)

	t, err := budget.OpenTracker(govDataDir, sessionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open tracker: %v", err)), nil
	}
	defer t.Close()

	prevSummary, _ := t.GetSummary()

	if err := t.ResetSession(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to reset session: %v", err)), nil
	}

	msg := fmt.Sprintf("🗜️ Context compaction triggered.\n"+
		"- Session: %s\n"+
		"- Tokens freed: %d\n"+
		"- Reason: %s\n\n"+
		"⚠️ IMPORTANT: You should now call `compact_memory` (context-manager) to persist a Knowledge Item for the current tactic before proceeding.",
		sessionID, prevSummary.TotalTokens, ifEmpty(reason, "manual trigger"))

	return mcp.NewToolResultText(msg), nil
}

func ifEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	s := server.NewMCPServer("McpContextGovernor", "1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("get_budget_status",
		mcp.WithDescription("Get current token budget status for the session. Returns usage level (ok/warning/critical/overflow) and recommended action."),
		mcp.WithString("session_id", mcp.Description("Session identifier (default: 'default').")),
	), handleGetBudgetStatus)

	s.AddTool(mcp.NewTool("estimate_cost",
		mcp.WithDescription("Estimate token cost of text or files and record to session budget tracker."),
		mcp.WithString("session_id", mcp.Description("Session identifier.")),
		mcp.WithString("tool", mcp.Description("Tool or action being estimated (for tracking).")),
		mcp.WithString("text", mcp.Description("Text content to estimate (optional).")),
		mcp.WithArray("files", mcp.Items(map[string]interface{}{"type": "string"}),
			mcp.Description("Absolute file paths to estimate (optional).")),
	), handleEstimateCost)

	s.AddTool(mcp.NewTool("suggest_compression",
		mcp.WithDescription("Get compression suggestions based on current context budget level."),
		mcp.WithString("session_id", mcp.Description("Session identifier.")),
	), handleSuggestCompression)

	s.AddTool(mcp.NewTool("trigger_compact",
		mcp.WithDescription("Trigger context compaction: resets token counter and instructs agent to call compact_memory."),
		mcp.WithString("session_id", mcp.Description("Session identifier.")),
		mcp.WithString("reason", mcp.Description("Reason for triggering compaction (for logging).")),
	), handleTriggerCompact)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Governor error: %v\n", err)
		os.Exit(1)
	}
}
