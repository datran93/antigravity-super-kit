package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WithAuditTrail is a middleware that logs the tool execution to the global SQLite database.
func WithAuditTrail(toolName string, next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Prepare payload
		payloadBytes, _ := json.Marshal(req.GetArguments())
		payloadStr := string(payloadBytes)

		// Execute next handler
		res, err := next(ctx, req)

		// Determine status and error
		status := "SUCCESS"
		errorStr := ""

		if err != nil {
			status = "ERROR"
			errorStr = err.Error()
		} else if res != nil && res.IsError {
			status = "TOOL_ERROR"
			contentBytes, _ := json.Marshal(res.Content)
			errorStr = string(contentBytes)
		}

		// Log to global DB
		logAudit(toolName, payloadStr, status, errorStr)

		return res, err
	}
}

func logAudit(toolName, payload, status, errorStr string) {
	db, err := GetGlobalDBConnection()
	if err != nil {
		fmt.Printf("[audit] Failed to get global DB connection: %v\n", err)
		return
	}
	defer db.Close()

	query := `INSERT INTO audit_logs (tool_name, request_payload, response_status, response_error) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(query, toolName, payload, status, errorStr)
	if err != nil {
		fmt.Printf("[audit] Failed to insert audit log: %v\n", err)
	}
}

// WithPermissionGuard is a middleware that evaluates tools against defined capabilities.
func WithPermissionGuard(toolName string, next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// AI Permission Guards capabilities checker
		if isRestrictedTool(toolName) {
			return mcp.NewToolResultError(fmt.Sprintf("Permission Guard: Tool '%s' execution requires explicit user confirmation or is currently blocked.", toolName)), nil
		}
		return next(ctx, req)
	}
}

func isRestrictedTool(toolName string) bool {
	// For now, context-manager tools are generally safe state management tools.
	// We can define a restricted list here.
	restricted := map[string]bool{
		"destructive_action_example": true, // example for future expansion
	}
	return restricted[toolName]
}

// WithMiddlewares applies all standard middlewares to a handler.
func WithMiddlewares(toolName string, handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return WithAuditTrail(toolName, WithPermissionGuard(toolName, handler))
}
