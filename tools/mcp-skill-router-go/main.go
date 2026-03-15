package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	openai "github.com/sashabaranov/go-openai"
)

var (
	workspaceRoot = "/Users/datran/LearnDev/antigravity-kit"
	skillsDir     = filepath.Join(workspaceRoot, ".agents", "skills")
	dbDir         = filepath.Join(workspaceRoot, "tools", "mcp-skill-router-go", ".db")
	dbFile        = filepath.Join(dbDir, "skills_cache.json")
	mu            sync.Mutex
)

func getOpenAIClient() (*openai.Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ERROR: OPENAI_API_KEY environment variable is not set")
	}
	return openai.NewClient(apiKey), nil
}

func searchSkillsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	query, _ := args["query"].(string)
	tagsFilter, _ := args["tags_filter"].(string)

	topK := 3
	if tk, ok := args["top_k"].(float64); ok {
		topK = int(tk)
	}

	results, err := searchSkills(ctx, query, tagsFilter, topK)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Search error: %v", err)), nil
	}
	if tagsFilter != "" && len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("❌ No relevant skills matching tags: '%s'", tagsFilter)), nil
	}

	return mcp.NewToolResultText(formatSearchResults(query, results)), nil
}

func main() {
	s := server.NewMCPServer("McpSkillRouter", "1.0.0")

	searchSkillsTool := mcp.NewTool("search_skills",
		mcp.WithDescription("Search for relevant AI agent skills based on a semantic query.\nUse this tool when you need to find which skills are best suited for a user's request."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Semantic query to search for (e.g., 'beautiful ui design', 'fix database query')")),
		mcp.WithString("tags_filter", mcp.Description("Optional comma-separated tags to filter by (e.g., 'frontend, react'). Will only return skills containing these tags.")),
		mcp.WithNumber("top_k", mcp.Description("Number of skills to return (default is 3)")),
	)
	s.AddTool(searchSkillsTool, searchSkillsTool_handler)

	s.AddTool(mcp.NewTool("ping",
		mcp.WithDescription("Health-check endpoint. Returns server name, version, and status=ok."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(`{"status":"ok","server":"McpSkillRouter","version":"1.0.0"}`), nil
	})

	server.ServeStdio(s)
}

// searchSkillsTool_handler is a package-level var pointing to the tool handler.
// This avoids naming collision between the mcp.Tool variable and the handler.
var searchSkillsTool_handler = searchSkillsHandler

func searchSkillsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return searchSkillsTool(ctx, request)
}
