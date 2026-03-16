package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// ─── GitLab client factory ─────────────────────────────────────────────────

func getGitLabClient() (*gitlab.Client, error) {
	token := os.Getenv("GITLAB_PRIVATE_TOKEN")
	if token == "" {
		return nil, fmt.Errorf(
			"GITLAB_PRIVATE_TOKEN environment variable is not set. " +
				"Create a PAT at https://gitlab.com/-/profile/personal_access_tokens " +
				"(scopes: read_api + api) and add it to mcp_config.json env block",
		)
	}

	baseURL := os.Getenv("GITLAB_URL")
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	return gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
}

// ─── Pagination constants ───────────────────────────────────────────────────

const (
	defaultPageSize = 8000  // characters per page
	maxPageSize     = 32000 // hard cap
)

// ─── Tool: get_file_content ────────────────────────────────────────────────

func getFileContentTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	projectID, _ := args["project_id"].(string)
	filePath, _ := args["file_path"].(string)
	ref, _ := args["ref"].(string)

	page := 1
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	pageSize := defaultPageSize
	if v, ok := args["page_size"].(float64); ok && v > 0 {
		pageSize = int(v)
		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
	}

	if projectID == "" || filePath == "" {
		return mcp.NewToolResultError("project_id and file_path are required"), nil
	}

	gl, err := getGitLabClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	opts := &gitlab.GetFileOptions{}
	if ref != "" {
		opts.Ref = gitlab.Ptr(ref)
	} else {
		opts.Ref = gitlab.Ptr("HEAD")
	}

	file, _, err := gl.RepositoryFiles.GetFile(projectID, filePath, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error reading '%s': %v", filePath, err)), nil
	}

	// GitLab returns base64-encoded content
	decoded, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Failed to decode file content: %v", err)), nil
	}
	rawContent := string(decoded)

	// Binary detection: look for null bytes
	for _, b := range decoded {
		if b == 0 {
			return mcp.NewToolResultText(fmt.Sprintf(
				"ℹ️ '%s' appears to be a binary file (SHA: %s). Raw content not returned.",
				filePath, file.BlobID,
			)), nil
		}
	}

	// Rune-based pagination
	runes := []rune(rawContent)
	totalRunes := len(runes)
	totalPages := (totalRunes + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		return mcp.NewToolResultText(fmt.Sprintf(
			"⚠️ Page %d exceeds total pages (%d) for '%s' (%d chars, page_size=%d).",
			page, totalPages, filePath, totalRunes, pageSize,
		)), nil
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalRunes {
		end = totalRunes
	}
	pageContent := string(runes[start:end])

	if totalPages > 1 {
		header := fmt.Sprintf("📄 `%s` — Page %d/%d (%d chars, page_size=%d)\n\n",
			filePath, page, totalPages, totalRunes, pageSize)
		pageContent = header + pageContent
	}

	return mcp.NewToolResultText(pageContent), nil
}

// ─── Tool: list_directory ──────────────────────────────────────────────────

func listDirectoryTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	projectID, _ := args["project_id"].(string)
	path, _ := args["path"].(string)
	ref, _ := args["ref"].(string)

	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	gl, err := getGitLabClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	opts := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}
	if path != "" {
		opts.Path = gitlab.Ptr(path)
	}
	if ref != "" {
		opts.Ref = gitlab.Ptr(ref)
	}

	const maxEntries = 200
	var allEntries []*gitlab.TreeNode

	for {
		entries, resp, err := gl.Repositories.ListTree(projectID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("❌ Error listing '%s': %v", path, err)), nil
		}
		allEntries = append(allEntries, entries...)
		if resp.NextPage == 0 || len(allEntries) >= maxEntries {
			break
		}
		opts.Page = resp.NextPage
	}

	displayPath := path
	if displayPath == "" {
		displayPath = "/"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📂 **%s** `%s`\n\n", projectID, displayPath))

	count := 0
	for _, entry := range allEntries {
		if count >= maxEntries {
			sb.WriteString(fmt.Sprintf("\n⚠️ Output capped at %d entries.\n", maxEntries))
			break
		}
		icon := "📄"
		if entry.Type == "tree" {
			icon = "📁"
		}
		sb.WriteString(fmt.Sprintf("%s `%s`  _(id: %s)_\n", icon, entry.Name, entry.ID[:8]))
		count++
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// ─── Tool: get_repository_info ─────────────────────────────────────────────

func getRepositoryInfoTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	gl, err := getGitLabClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	proj, _, err := gl.Projects.GetProject(projectID, &gitlab.GetProjectOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error fetching project '%s': %v", projectID, err)), nil
	}

	visibility := "public"
	if proj.Visibility == gitlab.PrivateVisibility {
		visibility = "private"
	} else if proj.Visibility == gitlab.InternalVisibility {
		visibility = "internal"
	}

	lastActivity := ""
	if proj.LastActivityAt != nil {
		lastActivity = proj.LastActivityAt.Format("2006-01-02")
	}

	output := fmt.Sprintf(
		"## 📦 %s\n\n"+
			"- **Description**: %s\n"+
			"- **Visibility**: %s\n"+
			"- **Default branch**: `%s`\n"+
			"- **Stars**: ⭐ %d\n"+
			"- **Forks**: 🍴 %d\n"+
			"- **Open Issues**: 🐛 %d\n"+
			"- **Last Activity**: %s\n"+
			"- **Clone URL (HTTP)**: %s\n"+
			"- **Clone URL (SSH)**: %s\n",
		proj.PathWithNamespace,
		proj.Description,
		visibility,
		proj.DefaultBranch,
		proj.StarCount,
		proj.ForksCount,
		proj.OpenIssuesCount,
		lastActivity,
		proj.HTTPURLToRepo,
		proj.SSHURLToRepo,
	)

	return mcp.NewToolResultText(output), nil
}

// ─── Tool: search_code ─────────────────────────────────────────────────────

func searchCodeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	projectID, _ := args["project_id"].(string)
	query, _ := args["query"].(string)

	perPage := 10
	if pp, ok := args["per_page"].(float64); ok {
		perPage = int(pp)
		if perPage > 30 {
			perPage = 30
		}
	}

	if projectID == "" || query == "" {
		return mcp.NewToolResultError("project_id and query are required"), nil
	}

	gl, err := getGitLabClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	scope := "blobs"
	_ = scope

	results, _, err := gl.Search.BlobsByProject(projectID, query, &gitlab.SearchOptions{
		ListOptions: gitlab.ListOptions{PerPage: int64(perPage)},
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Search failed for '%s' in project '%s': %v", query, projectID, err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("🔍 No results for `%s` in `%s`.", query, projectID)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 **%d result(s)** for `%s` in `%s`\n\n", len(results), query, projectID))

	for i, item := range results {
		sb.WriteString(fmt.Sprintf("### %d. `%s`\n", i+1, item.Filename))
		sb.WriteString(fmt.Sprintf("   - **Path**: `%s`\n", item.Filename))
		sb.WriteString(fmt.Sprintf("   - **Ref**: `%s`\n", item.Ref))
		if item.Data != "" {
			// Show first 300 chars of matching data
			snippet := item.Data
			if len(snippet) > 300 {
				snippet = snippet[:300] + "..."
			}
			sb.WriteString(fmt.Sprintf("   - **Match**: `...%s...`\n", strings.TrimSpace(snippet)))
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// ─── Tool: read_mr_discussions ─────────────────────────────────────────────

func readMRDiscussionsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	projectID, _ := args["project_id"].(string)
	mrIIDFloat, ok := args["mr_iid"].(float64)
	if !ok {
		return mcp.NewToolResultError("mr_iid must be a number"), nil
	}
	mrIID := int64(mrIIDFloat)

	gl, err := getGitLabClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error: %v", err)), nil
	}

	opt := &gitlab.ListMergeRequestDiscussionsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	var allDiscussions []*gitlab.Discussion
	for {
		discussions, resp, err := gl.Discussions.ListMergeRequestDiscussions(projectID, mrIID, opt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("❌ Error reading MR discussions: %v", err)), nil
		}
		allDiscussions = append(allDiscussions, discussions...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var output []string
	output = append(output, fmt.Sprintf("💬 DISCUSSIONS FOR MR #%d (Project: %s)\n", mrIID, projectID))

	count := 0
	for _, disc := range allDiscussions {
		if len(disc.Notes) == 0 {
			continue
		}

		isSystemOnly := true
		for _, note := range disc.Notes {
			if !note.System {
				isSystemOnly = false
				break
			}
		}
		if isSystemOnly {
			continue
		}

		count++

		statusText := ""
		if disc.Notes[0].Resolvable {
			status := "❌ UNRESOLVED"
			if disc.Notes[0].Resolved {
				status = "✅ RESOLVED"
			}
			statusText = fmt.Sprintf(" | Status: %s", status)
		}

		output = append(output, fmt.Sprintf("--- Discussion ID: %s%s ---", disc.ID, statusText))

		for _, note := range disc.Notes {
			author := note.Author.Username
			if author == "" {
				author = "unknown"
			}
			body := note.Body
			created := ""
			if note.CreatedAt != nil {
				created = note.CreatedAt.Format("2006-01-02 15:04:05")
			}
			isSystem := ""
			if note.System {
				isSystem = "[SYSTEM]"
			}
			output = append(output, fmt.Sprintf("[%s] @%s %s (Note ID: %d):\n%s\n", created, author, isSystem, note.ID, body))
		}
	}

	if count == 0 {
		return mcp.NewToolResultText("✅ No user discussions found on this Merge Request."), nil
	}

	return mcp.NewToolResultText(strings.Join(output, "\n")), nil
}

// ─── Tool: reply_to_mr_discussion ──────────────────────────────────────────

func replyToMRDiscussionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	projectID, _ := args["project_id"].(string)
	mrIIDFloat, ok := args["mr_iid"].(float64)
	if !ok {
		return mcp.NewToolResultError("mr_iid must be a number"), nil
	}
	mrIID := int64(mrIIDFloat)
	discussionID, _ := args["discussion_id"].(string)
	body, _ := args["body"].(string)

	gl, err := getGitLabClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error: %v", err)), nil
	}

	opt := &gitlab.AddMergeRequestDiscussionNoteOptions{
		Body: gitlab.Ptr(body),
	}

	_, _, err = gl.Discussions.AddMergeRequestDiscussionNote(projectID, mrIID, discussionID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error replying to discussion: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ Successfully replied to discussion '%s' on MR #%d.", discussionID, mrIID)), nil
}

// ─── Tool: resolve_mr_discussion ───────────────────────────────────────────

func resolveMRDiscussionTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	projectID, _ := args["project_id"].(string)
	mrIIDFloat, ok := args["mr_iid"].(float64)
	if !ok {
		return mcp.NewToolResultError("mr_iid must be a number"), nil
	}
	mrIID := int64(mrIIDFloat)
	discussionID, _ := args["discussion_id"].(string)

	resolve := true
	if r, ok := args["resolve"].(bool); ok {
		resolve = r
	}

	gl, err := getGitLabClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error: %v", err)), nil
	}

	// Verify it exists and is resolvable
	disc, _, err := gl.Discussions.GetMergeRequestDiscussion(projectID, mrIID, discussionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Discussion '%s' not found or error: %v", discussionID, err)), nil
	}

	if len(disc.Notes) == 0 || !disc.Notes[0].Resolvable {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Discussion '%s' is not resolvable.", discussionID)), nil
	}

	opt := &gitlab.ResolveMergeRequestDiscussionOptions{
		Resolved: gitlab.Ptr(resolve),
	}

	_, _, err = gl.Discussions.ResolveMergeRequestDiscussion(projectID, mrIID, discussionID, opt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error resolving discussion: %v", err)), nil
	}

	action := "resolved"
	if !resolve {
		action = "unresolved"
	}
	return mcp.NewToolResultText(fmt.Sprintf("✅ Successfully %s discussion '%s' on MR #%d.", action, discussionID, mrIID)), nil
}

// ─── Main ──────────────────────────────────────────────────────────────────

func main() {
	s := server.NewMCPServer("McpGitLab", "1.0.0")

	// Tool: get_file_content
	s.AddTool(mcp.NewTool("get_file_content",
		mcp.WithDescription(
			"Read the text content of a file from a GitLab repository. "+
				"Returns decoded UTF-8 text with rune-based pagination. "+
				"Binary files are gracefully detected and skipped. "+
				"Requires GITLAB_PRIVATE_TOKEN env var (PAT with 'read_api' scope).",
		),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("The ID or URL-encoded path of the project (e.g. '12345' or 'mygroup/myrepo')")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File path within the repo (e.g. 'src/main.go')")),
		mcp.WithString("ref", mcp.Description("Branch, tag, or commit SHA (defaults to HEAD)")),
		mcp.WithNumber("page", mcp.Description("Page number (1-indexed). Default: 1")),
		mcp.WithNumber("page_size", mcp.Description("Characters per page (default 8000, max 32000)")),
	), getFileContentTool)

	// Tool: list_directory
	s.AddTool(mcp.NewTool("list_directory",
		mcp.WithDescription(
			"List files and subdirectories at a given path in a GitLab repository. "+
				"Returns a Markdown-formatted tree with icons. Capped at 200 entries.",
		),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("The ID or URL-encoded path of the project")),
		mcp.WithString("path", mcp.Description("Directory path within the repo (empty string = root)")),
		mcp.WithString("ref", mcp.Description("Branch, tag, or commit SHA (defaults to default branch)")),
	), listDirectoryTool)

	// Tool: get_repository_info
	s.AddTool(mcp.NewTool("get_repository_info",
		mcp.WithDescription(
			"Fetch metadata for a GitLab project: description, default branch, visibility, "+
				"star count, fork count, open issues, last activity, and clone URLs.",
		),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("The ID or URL-encoded path of the project")),
	), getRepositoryInfoTool)

	// Tool: search_code
	s.AddTool(mcp.NewTool("search_code",
		mcp.WithDescription(
			"Search code blobs within a specific GitLab project using the GitLab Search API. "+
				"Returns matching file paths and content snippets.",
		),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("The ID or URL-encoded path of the project")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (e.g. 'func handleRequest')")),
		mcp.WithNumber("per_page", mcp.Description("Number of results to return (default: 10, max: 30)")),
	), searchCodeTool)

	// Tool: read_mr_discussions
	s.AddTool(mcp.NewTool("read_mr_discussions",
		mcp.WithDescription("Read all discussions (threads) from a specific GitLab Merge Request."),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("The ID or URL-encoded path of the project")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("The internal ID of the merge request")),
	), readMRDiscussionsTool)

	// Tool: reply_to_mr_discussion
	s.AddTool(mcp.NewTool("reply_to_mr_discussion",
		mcp.WithDescription("Reply to an existing discussion thread on a GitLab Merge Request."),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("The ID or URL-encoded path of the project")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("The internal ID of the merge request")),
		mcp.WithString("discussion_id", mcp.Required(), mcp.Description("The ID of the discussion thread to reply to")),
		mcp.WithString("body", mcp.Required(), mcp.Description("The text content of your reply")),
	), replyToMRDiscussionTool)

	// Tool: resolve_mr_discussion
	s.AddTool(mcp.NewTool("resolve_mr_discussion",
		mcp.WithDescription("Resolve or unresolve a discussion thread on a GitLab Merge Request."),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("The ID or URL-encoded path of the project")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("The internal ID of the merge request")),
		mcp.WithString("discussion_id", mcp.Required(), mcp.Description("The ID of the discussion thread")),
		mcp.WithBoolean("resolve", mcp.Required(), mcp.Description("True to resolve, False to unresolve")),
	), resolveMRDiscussionTool)

	// Tool: ping
	s.AddTool(mcp.NewTool("ping",
		mcp.WithDescription("Health-check endpoint. Returns server name, version, and status=ok."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(`{"status":"ok","server":"McpGitLab","version":"1.0.0"}`), nil
	})

	server.ServeStdio(s)
}
