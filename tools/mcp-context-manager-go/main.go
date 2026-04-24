package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	mcpServer := server.NewMCPServer(
		"McpContextManager",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// save_checkpoint
	mcpServer.AddTool(mcp.NewTool("save_checkpoint",
		mcp.WithDescription("Save or update a task checkpoint/context."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Description of the task")),
		mcp.WithString("status", mcp.Required(), mcp.Description("Status")),
		mcp.WithArray("completed_steps", mcp.Items(map[string]interface{}{"type": "string"}), mcp.Description("Completed steps array")),
		mcp.WithArray("next_steps", mcp.Items(map[string]interface{}{"type": "string"}), mcp.Description("Next steps array")),
		mcp.WithArray("active_files", mcp.Items(map[string]interface{}{"type": "string"}), mcp.Description("Active files array")),
		mcp.WithString("notes", mcp.Required(), mcp.Description("Notes")),
	), WithMiddlewares("save_checkpoint", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)
		description, _ := args["description"].(string)
		status, _ := args["status"].(string)

		compStepsIf, _ := args["completed_steps"].([]interface{})
		var compSteps []string
		for _, s := range compStepsIf {
			if str, ok := s.(string); ok {
				compSteps = append(compSteps, str)
			}
		}

		nextStepsIf, _ := args["next_steps"].([]interface{})
		var nextSteps []string
		for _, s := range nextStepsIf {
			if str, ok := s.(string); ok {
				nextSteps = append(nextSteps, str)
			}
		}

		activeFilesIf, _ := args["active_files"].([]interface{})
		var activeFiles []string
		for _, s := range activeFilesIf {
			if str, ok := s.(string); ok {
				activeFiles = append(activeFiles, str)
			}
		}

		notes, _ := args["notes"].(string)

		res, err := SaveCheckpoint(workspacePath, taskID, description, status, compSteps, nextSteps, activeFiles, notes)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// initialize_task_plan
	mcpServer.AddTool(mcp.NewTool("initialize_task_plan",
		mcp.WithDescription("Start a new task with a list of steps."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Description of the task")),
		mcp.WithArray("steps", mcp.Items(map[string]interface{}{"type": "string"}), mcp.Description("Steps collection")),
	), WithMiddlewares("initialize_task_plan", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)
		description, _ := args["description"].(string)

		stepsIf, _ := args["steps"].([]interface{})
		var steps []string
		for _, s := range stepsIf {
			if str, ok := s.(string); ok {
				steps = append(steps, str)
			}
		}

		res, err := InitializeTaskPlan(workspacePath, taskID, description, steps)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// complete_task_step
	mcpServer.AddTool(mcp.NewTool("complete_task_step",
		mcp.WithDescription("Mark step as done, track active files, update graph and bar."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("step_name", mcp.Required(), mcp.Description("Step name")),
		mcp.WithArray("active_files", mcp.Items(map[string]interface{}{"type": "string"}), mcp.Description("Active files array")),
		mcp.WithString("notes", mcp.Description("Notes")),
	), WithMiddlewares("complete_task_step", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)
		stepName, _ := args["step_name"].(string)

		var activeFiles []string
		if val, ok := args["active_files"]; ok && val != nil {
			if activeFilesIf, ok2 := val.([]interface{}); ok2 {
				for _, s := range activeFilesIf {
					if str, ok := s.(string); ok {
						activeFiles = append(activeFiles, str)
					}
				}
			}
		}

		var notes string
		if val, ok := args["notes"]; ok && val != nil {
			notes, _ = val.(string)
		}

		res, err := CompleteTaskStep(workspacePath, taskID, stepName, activeFiles, notes)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// add_task_step
	mcpServer.AddTool(mcp.NewTool("add_task_step",
		mcp.WithDescription("Add a new task step to the next_steps list of an existing task."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("new_step", mcp.Required(), mcp.Description("New step")),
	), WithMiddlewares("add_task_step", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)
		newStep, _ := args["new_step"].(string)

		res, err := AddTaskStep(workspacePath, taskID, newStep)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// load_checkpoint
	mcpServer.AddTool(mcp.NewTool("load_checkpoint",
		mcp.WithDescription("Load a previously saved task checkpoint."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
	), WithMiddlewares("load_checkpoint", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)

		res, err := LoadCheckpoint(workspacePath, taskID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// list_active_tasks
	mcpServer.AddTool(mcp.NewTool("list_active_tasks",
		mcp.WithDescription("List all active tasks."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
	), WithMiddlewares("list_active_tasks", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)

		res, err := ListActiveTasks(workspacePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// declare_intent
	mcpServer.AddTool(mcp.NewTool("declare_intent",
		mcp.WithDescription("Declare working intention and restrict changes to specific locked_files. TTL default=60min (0=no expiry)."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("tactic", mcp.Required(), mcp.Description("Tactic name")),
		mcp.WithArray("locked_files", mcp.Items(map[string]interface{}{"type": "string"}), mcp.Description("Locked files")),
		mcp.WithNumber("ttl_minutes", mcp.Description("Lock TTL in minutes. 0=no expiry (default=60)")),
	), WithMiddlewares("declare_intent", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)
		tactic, _ := args["tactic"].(string)

		lockedIf, _ := args["locked_files"].([]interface{})
		var locked []string
		for _, v := range lockedIf {
			if str, ok := v.(string); ok {
				locked = append(locked, str)
			}
		}

		ttlMinutes := 60 // default
		if val, ok := args["ttl_minutes"]; ok && val != nil {
			if num, ok2 := val.(float64); ok2 {
				ttlMinutes = int(num)
			}
		}

		res, err := DeclareIntent(workspacePath, taskID, tactic, locked, ttlMinutes)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// check_intent_lock
	mcpServer.AddTool(mcp.NewTool("check_intent_lock",
		mcp.WithDescription("Check if a file is unlocked for the current intent."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File Path")),
	), WithMiddlewares("check_intent_lock", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)
		filePath, _ := args["file_path"].(string)

		res, err := CheckIntentLock(workspacePath, taskID, filePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// annotate_file
	mcpServer.AddTool(mcp.NewTool("annotate_file",
		mcp.WithDescription("Add 'Ghost Context' (lessons/gotchas) to a specific file"),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File Path")),
		mcp.WithString("gotchas", mcp.Required(), mcp.Description("Gotchas context")),
	), WithMiddlewares("annotate_file", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		filePath, _ := args["file_path"].(string)
		gotchas, _ := args["gotchas"].(string)

		res, err := AnnotateFile(workspacePath, filePath, gotchas)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// record_failure
	mcpServer.AddTool(mcp.NewTool("record_failure",
		mcp.WithDescription("Record a failure (e.g. test failing, compile error) to detect context drift."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("step_name", mcp.Description("Optional: current step name for war-room context")),
		mcp.WithString("error_context", mcp.Description("Optional: error message or context for war-room KI")),
	), WithMiddlewares("record_failure", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)

		var stepName, errorContext string
		if val, ok := args["step_name"]; ok && val != nil {
			stepName, _ = val.(string)
		}
		if val, ok := args["error_context"]; ok && val != nil {
			errorContext, _ = val.(string)
		}

		res, err := RecordFailure(workspacePath, taskID, stepName, errorContext)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// clear_drift
	mcpServer.AddTool(mcp.NewTool("clear_drift",
		mcp.WithDescription("Reset the failure counter after a success or planner intervention."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
	), WithMiddlewares("clear_drift", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)

		res, err := ClearDrift(workspacePath, taskID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// manage_anchors
	mcpServer.AddTool(mcp.NewTool("manage_anchors",
		mcp.WithDescription("Manage architectural anchors. Action can be 'set', 'get', or 'list'."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("scope", mcp.Description("Scope: 'project' or 'global' (default 'project')")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action (set, get, list)")),
		mcp.WithString("key", mcp.Description("Anchor key")),
		mcp.WithString("value", mcp.Description("Anchor value")),
		mcp.WithString("rule", mcp.Description("Anchor rule")),
	), WithMiddlewares("manage_anchors", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		action, _ := args["action"].(string)

		scope := "project"
		if val, ok := args["scope"]; ok && val != nil {
			if str, ok := val.(string); ok && str != "" {
				scope = str
			}
		}

		var key, value, rule string
		if val, ok := args["key"]; ok && val != nil {
			key, _ = val.(string)
		}
		if val, ok := args["value"]; ok && val != nil {
			value, _ = val.(string)
		}
		if val, ok := args["rule"]; ok && val != nil {
			rule, _ = val.(string)
		}

		res, err := ManageAnchors(workspacePath, scope, action, key, value, rule)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// recall_knowledge
	mcpServer.AddTool(mcp.NewTool("recall_knowledge",
		mcp.WithDescription("Recall relevant Knowledge Items (Local RAG) based on a search query using SQLite FTS5."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("scope", mcp.Description("Scope: 'project' or 'global' (default 'project')")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("top_k", mcp.Description("Top K results")),
	), WithMiddlewares("recall_knowledge", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		query, _ := args["query"].(string)

		scope := "project"
		if val, ok := args["scope"]; ok && val != nil {
			if str, ok := val.(string); ok && str != "" {
				scope = str
			}
		}

		topK := 3
		if val, ok := args["top_k"]; ok && val != nil {
			if num, ok2 := val.(float64); ok2 {
				topK = int(num)
			}
		}

		res, err := RecallKnowledge(workspacePath, scope, query, topK)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// compact_memory
	mcpServer.AddTool(mcp.NewTool("compact_memory",
		mcp.WithDescription("Generate a Knowledge Item (KI) and prune context at the end of a tactic. Also indexes KI into Semantic SQLite."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("tactic_name", mcp.Required(), mcp.Description("Tactic name")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("KI summary")),
		mcp.WithString("decisions", mcp.Required(), mcp.Description("Architecture decisions")),
	), WithMiddlewares("compact_memory", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)
		tacticName, _ := args["tactic_name"].(string)
		summary, _ := args["summary"].(string)
		decisions, _ := args["decisions"].(string)

		res, err := CompactMemory(workspacePath, taskID, tacticName, summary, decisions)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// find_recent_task (Improvement #4: Smart session continuity)
	mcpServer.AddTool(mcp.NewTool("find_recent_task",
		mcp.WithDescription("Fuzzy keyword search across checkpoint descriptions. Returns top 3 matching tasks for smart session continuity."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("keywords", mcp.Required(), mcp.Description("Keywords to search for in task descriptions")),
	), WithMiddlewares("find_recent_task", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		keywords, _ := args["keywords"].(string)

		res, err := FindRecentTask(workspacePath, keywords)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// review_checkpoint (Improvement #9: Checkpoint quality validator)
	mcpServer.AddTool(mcp.NewTool("review_checkpoint",
		mcp.WithDescription("Run 5 quality checks on a checkpoint: stale detection, active_files guard, step label format, duplicate steps, git SHA match."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to validate")),
	), WithMiddlewares("review_checkpoint", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)

		res, err := ReviewCheckpoint(workspacePath, taskID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// delete_task
	mcpServer.AddTool(mcp.NewTool("delete_task",
		mcp.WithDescription("Permanently delete a task checkpoint and its intent locks from the database. Use when the USER explicitly requests to remove a task. Progress.md is refreshed automatically."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to delete")),
	), WithMiddlewares("delete_task", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)

		res, err := DeleteTask(workspacePath, taskID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// get_task_summary
	mcpServer.AddTool(mcp.NewTool("get_task_summary",
		mcp.WithDescription("Get a compact JSON summary of a task (status, progress%, next step). Cheaper than load_checkpoint when you only need a quick status check."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
	), WithMiddlewares("get_task_summary", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		taskID, _ := args["task_id"].(string)

		res, err := GetTaskSummary(workspacePath, taskID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))
	// manage_session_memory (T08: session-scoped ephemeral memory)
	mcpServer.AddTool(mcp.NewTool("manage_session_memory",
		mcp.WithDescription("Manage session-scoped ephemeral memory. Actions: add, list, promote, clear."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action: add, list, promote, clear")),
		mcp.WithString("session_id", mcp.Description("Session ID for scoping")),
		mcp.WithString("category", mcp.Description("Category: finding, decision, pattern")),
		mcp.WithString("content", mcp.Description("Memory content (for add) or memory ID (for promote)")),
	), WithMiddlewares("manage_session_memory", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		action, _ := args["action"].(string)

		var sessionID, category, content string
		if val, ok := args["session_id"]; ok && val != nil {
			sessionID, _ = val.(string)
		}
		if val, ok := args["category"]; ok && val != nil {
			category, _ = val.(string)
		}
		if val, ok := args["content"]; ok && val != nil {
			content, _ = val.(string)
		}

		res, err := ManageSessionMemory(workspacePath, action, sessionID, category, content)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// log_activity (T10: activity event audit trail)
	mcpServer.AddTool(mcp.NewTool("log_activity",
		mcp.WithDescription("Record an activity event in the audit trail."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("event_type", mcp.Required(), mcp.Description("Event type (e.g. step_completed, ki_created)")),
		mcp.WithString("task_id", mcp.Description("Associated task ID")),
		mcp.WithString("detail", mcp.Description("Event detail text")),
	), WithMiddlewares("log_activity", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		eventType, _ := args["event_type"].(string)

		var taskID, detail string
		if val, ok := args["task_id"]; ok && val != nil {
			taskID, _ = val.(string)
		}
		if val, ok := args["detail"]; ok && val != nil {
			detail, _ = val.(string)
		}

		res, err := LogActivity(workspacePath, eventType, taskID, detail)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// list_activity (T10: query activity events)
	mcpServer.AddTool(mcp.NewTool("list_activity",
		mcp.WithDescription("List recent activity events with optional filtering."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("event_type", mcp.Description("Filter by event type")),
		mcp.WithString("task_id", mcp.Description("Filter by task ID")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 20)")),
	), WithMiddlewares("list_activity", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)

		var eventType, taskID string
		if val, ok := args["event_type"]; ok && val != nil {
			eventType, _ = val.(string)
		}
		if val, ok := args["task_id"]; ok && val != nil {
			taskID, _ = val.(string)
		}

		limit := 20
		if val, ok := args["limit"]; ok && val != nil {
			if num, ok2 := val.(float64); ok2 {
				limit = int(num)
			}
		}

		res, err := ListActivity(workspacePath, eventType, taskID, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// create_doc (T12: structured documentation)
	mcpServer.AddTool(mcp.NewTool("create_doc",
		mcp.WithDescription("Create or update a structured documentation entry."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("doc_path", mcp.Required(), mcp.Description("Doc path (e.g. 'patterns/auth', 'architecture/overview')")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Document title")),
		mcp.WithString("content", mcp.Description("Document content in Markdown")),
	), WithMiddlewares("create_doc", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		docPath, _ := args["doc_path"].(string)
		title, _ := args["title"].(string)

		var content string
		if val, ok := args["content"]; ok && val != nil {
			content, _ = val.(string)
		}

		res, err := CreateDoc(workspacePath, docPath, title, content)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// get_doc (T12)
	mcpServer.AddTool(mcp.NewTool("get_doc",
		mcp.WithDescription("Retrieve a structured documentation entry by path."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("doc_path", mcp.Required(), mcp.Description("Doc path to retrieve")),
	), WithMiddlewares("get_doc", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		docPath, _ := args["doc_path"].(string)

		res, err := GetDoc(workspacePath, docPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// list_docs (T12)
	mcpServer.AddTool(mcp.NewTool("list_docs",
		mcp.WithDescription("List all structured documentation entries."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
	), WithMiddlewares("list_docs", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)

		res, err := ListDocs(workspacePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// search_docs (T12)
	mcpServer.AddTool(mcp.NewTool("search_docs",
		mcp.WithDescription("Search structured documentation by keyword."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
	), WithMiddlewares("search_docs", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		query, _ := args["query"].(string)

		res, err := SearchDocs(workspacePath, query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))
	// retrieve_context (T15: unified context assembly)
	mcpServer.AddTool(mcp.NewTool("retrieve_context",
		mcp.WithDescription("Assemble a context pack from KIs, docs, anchors, and tasks matching a query."),
		mcp.WithString("workspace_path", mcp.Required(), mcp.Description("Workspace path")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Natural language query")),
		mcp.WithString("scope", mcp.Description("Scope: 'project' or 'global' (default 'project')")),
	), WithMiddlewares("retrieve_context", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		workspacePath, _ := args["workspace_path"].(string)
		query, _ := args["query"].(string)

		scope := "project"
		if val, ok := args["scope"]; ok && val != nil {
			scope, _ = val.(string)
		}

		res, err := RetrieveContext(workspacePath, query, scope)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(res), nil
	}))

	// Run standard I/O server
	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Fprintf(os.Stderr, "[context-manager] server error: %v\n", err)
	}
}
