package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"antigravity-kit/mcp-confluence-go/confluence"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// ── Config from environment ───────────────────────────────────────────
	baseURL := os.Getenv("CONFLUENCE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://seinetime.atlassian.net"
	}
	username := os.Getenv("CONFLUENCE_USERNAME")
	apiToken := os.Getenv("CONFLUENCE_API_TOKEN")

	if username == "" || apiToken == "" {
		log.Fatal("CONFLUENCE_USERNAME and CONFLUENCE_API_TOKEN environment variables are required")
	}

	client := confluence.NewClient(baseURL, username, apiToken)

	// ── MCP Server ────────────────────────────────────────────────────────
	s := server.NewMCPServer(
		"McpConfluence",
		"1.0.0",
		server.WithLogging(),
	)

	// ─────────────────────────────────────────────────────────────────────
	// TOOL: get_page
	// ─────────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_page",
			mcp.WithDescription("Fetch a Confluence page by its page ID. Returns the page title, version, space, and body content in Confluence storage format."),
			mcp.WithString("page_id", mcp.Required(), mcp.Description("The numeric Confluence page ID (e.g. '123456').")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			pageID, ok := args["page_id"].(string)
			if !ok || pageID == "" {
				return mcp.NewToolResultError("page_id is required"), nil
			}

			page, err := client.GetPage(pageID)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get page: %v", err)), nil
			}

			webURL := baseURL + "/wiki" + page.Links.WebUI
			result := fmt.Sprintf("## %s\n\n**ID:** %s\n**Space ID:** %s\n**Version:** %d\n**URL:** %s\n\n---\n\n%s",
				page.Title, page.ID, page.SpaceID, page.Version.Number, webURL,
				page.Body.Storage.Value)
			return mcp.NewToolResultText(result), nil
		},
	)

	// ─────────────────────────────────────────────────────────────────────
	// TOOL: search_pages
	// ─────────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("search_pages",
			mcp.WithDescription("Search Confluence pages using CQL (Confluence Query Language). Returns matching pages with title, URL, space, and excerpt. Example CQL: 'type=page AND space.key=DEV AND text~\"authentication\"'"),
			mcp.WithString("cql", mcp.Required(), mcp.Description("CQL query string. E.g. 'type=page AND text~\"deployment\"' or 'space.key=TEAM AND title~\"spec\"'.")),
			mcp.WithNumber("limit", mcp.Description("Max results to return (default 10, max 50).")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			cql, ok := args["cql"].(string)
			if !ok || cql == "" {
				return mcp.NewToolResultError("cql is required"), nil
			}
			limit := 10
			if l, ok := args["limit"].(float64); ok && l > 0 {
				limit = int(l)
				if limit > 50 {
					limit = 50
				}
			}

			results, err := client.SearchPages(cql, limit)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
			}

			if len(results) == 0 {
				return mcp.NewToolResultText("No pages found matching your query."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("## Search Results (%d pages)\n\n", len(results)))
			for i, r := range results {
				webURL := baseURL + "/wiki" + r.Links.WebUI
				sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, r.Title))
				sb.WriteString(fmt.Sprintf("- **ID:** %s\n", r.ID))
				sb.WriteString(fmt.Sprintf("- **Space:** %s (%s)\n", r.Space.Name, r.Space.Key))
				sb.WriteString(fmt.Sprintf("- **URL:** %s\n", webURL))
				if r.Excerpt != "" {
					sb.WriteString(fmt.Sprintf("- **Excerpt:** %s\n", r.Excerpt))
				}
				sb.WriteString("\n")
			}
			return mcp.NewToolResultText(sb.String()), nil
		},
	)

	// ─────────────────────────────────────────────────────────────────────
	// TOOL: get_page_children
	// ─────────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_page_children",
			mcp.WithDescription("List all direct child pages of a given Confluence page. Useful for navigating page hierarchies."),
			mcp.WithString("page_id", mcp.Required(), mcp.Description("The parent Confluence page ID.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			pageID, ok := args["page_id"].(string)
			if !ok || pageID == "" {
				return mcp.NewToolResultError("page_id is required"), nil
			}

			children, err := client.GetPageChildren(pageID)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get children: %v", err)), nil
			}

			if len(children) == 0 {
				return mcp.NewToolResultText(fmt.Sprintf("Page %s has no child pages.", pageID)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("## Child Pages of %s (%d total)\n\n", pageID, len(children)))
			for _, child := range children {
				webURL := baseURL + "/wiki" + child.Links.WebUI
				sb.WriteString(fmt.Sprintf("- **%s** (ID: %s) — %s\n", child.Title, child.ID, webURL))
			}
			return mcp.NewToolResultText(sb.String()), nil
		},
	)

	// ─────────────────────────────────────────────────────────────────────
	// TOOL: get_spaces
	// ─────────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_spaces",
			mcp.WithDescription("List all available Confluence spaces. Returns space key, name, and homepage ID. Use this to discover space keys for CQL searches or page creation."),
			mcp.WithNumber("limit", mcp.Description("Max spaces to return (default 25).")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			limit := 25
			if l, ok := args["limit"].(float64); ok && l > 0 {
				limit = int(l)
			}

			spaces, err := client.GetSpaces(limit)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get spaces: %v", err)), nil
			}

			if len(spaces) == 0 {
				return mcp.NewToolResultText("No spaces found."), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("## Confluence Spaces (%d)\n\n", len(spaces)))
			sb.WriteString("| Key | Name | Homepage ID |\n|-----|------|-------------|\n")
			for _, sp := range spaces {
				sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", sp.Key, sp.Name, sp.HomepageID))
			}
			return mcp.NewToolResultText(sb.String()), nil
		},
	)

	// ─────────────────────────────────────────────────────────────────────
	// TOOL: create_page
	// ─────────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("create_page",
			mcp.WithDescription("Create a new Confluence page. Accepts a Markdown body which is automatically wrapped in a Confluence storage-format macro. Returns the new page ID and URL."),
			mcp.WithString("space_id", mcp.Required(), mcp.Description("The Confluence space ID (numeric). Get this from get_spaces.")),
			mcp.WithString("title", mcp.Required(), mcp.Description("The page title.")),
			mcp.WithString("body", mcp.Required(), mcp.Description("Page content in Markdown format. Will be converted to Confluence storage format.")),
			mcp.WithString("parent_id", mcp.Description("Optional: parent page ID to nest this page under.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			spaceID, _ := args["space_id"].(string)
			title, _ := args["title"].(string)
			body, _ := args["body"].(string)
			parentID, _ := args["parent_id"].(string)

			if spaceID == "" || title == "" || body == "" {
				return mcp.NewToolResultError("space_id, title, and body are required"), nil
			}

			storageHTML := confluence.MarkdownToStorage(body)
			page, err := client.CreatePage(spaceID, parentID, title, storageHTML)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to create page: %v", err)), nil
			}

			webURL := baseURL + "/wiki" + page.Links.WebUI
			result := fmt.Sprintf("✅ Page created successfully!\n\n**Title:** %s\n**ID:** %s\n**URL:** %s",
				page.Title, page.ID, webURL)
			return mcp.NewToolResultText(result), nil
		},
	)

	// ─────────────────────────────────────────────────────────────────────
	// TOOL: update_page
	// ─────────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("update_page",
			mcp.WithDescription("Update an existing Confluence page's title and/or body. Automatically fetches and increments the page version. Accepts Markdown body content."),
			mcp.WithString("page_id", mcp.Required(), mcp.Description("The Confluence page ID to update.")),
			mcp.WithString("title", mcp.Required(), mcp.Description("The new page title (can be same as current).")),
			mcp.WithString("body", mcp.Required(), mcp.Description("New page content in Markdown format.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			pageID, _ := args["page_id"].(string)
			title, _ := args["title"].(string)
			body, _ := args["body"].(string)

			if pageID == "" || title == "" || body == "" {
				return mcp.NewToolResultError("page_id, title, and body are required"), nil
			}

			storageHTML := confluence.MarkdownToStorage(body)
			page, err := client.UpdatePage(pageID, title, storageHTML)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to update page: %v", err)), nil
			}

			webURL := baseURL + "/wiki" + page.Links.WebUI
			result := fmt.Sprintf("✅ Page updated successfully!\n\n**Title:** %s\n**ID:** %s\n**Version:** %d\n**URL:** %s",
				page.Title, page.ID, page.Version.Number, webURL)
			return mcp.NewToolResultText(result), nil
		},
	)

	// ─────────────────────────────────────────────────────────────────────
	// TOOL: add_comment
	// ─────────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("add_comment",
			mcp.WithDescription("Add a footer comment to a Confluence page. Useful for adding review notes, AI-generated summaries, or status updates."),
			mcp.WithString("page_id", mcp.Required(), mcp.Description("The Confluence page ID to comment on.")),
			mcp.WithString("comment", mcp.Required(), mcp.Description("The comment text to add.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			pageID, _ := args["page_id"].(string)
			comment, _ := args["comment"].(string)

			if pageID == "" || comment == "" {
				return mcp.NewToolResultError("page_id and comment are required"), nil
			}

			c, err := client.AddComment(pageID, comment)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to add comment: %v", err)), nil
			}

			webURL := baseURL + "/wiki" + c.Links.WebUI
			result := fmt.Sprintf("✅ Comment added!\n\n**Comment ID:** %s\n**URL:** %s", c.ID, webURL)
			return mcp.NewToolResultText(result), nil
		},
	)

	// ── Start serving ─────────────────────────────────────────────────────
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
