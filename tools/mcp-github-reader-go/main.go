package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
)

// ─── GitHub client factory ─────────────────────────────────────────────────

func getGitHubClient() (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf(
			"GITHUB_TOKEN environment variable is not set. " +
				"Create a PAT at https://github.com/settings/tokens (scope: repo) " +
				"and add it to mcp_config.json env block",
		)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	baseURL := os.Getenv("GITHUB_API_URL")
	if baseURL != "" {
		client, err := github.NewClient(tc).WithAuthToken(token).WithEnterpriseURLs(baseURL, baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub enterprise client: %w", err)
		}
		return client, nil
	}

	return github.NewClient(tc), nil
}

// logRateLimit writes remaining rate limit info to stderr (never stdout/MCP stream).
func logRateLimit(resp *github.Response) {
	if resp != nil {
		fmt.Fprintf(os.Stderr, "[github-reader] rate-limit remaining: %d, reset: %s\n",
			resp.Rate.Remaining, resp.Rate.Reset.Time)
	}
}

// ─── Tool: get_file_content ────────────────────────────────────────────────

func getFileContentTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	path, _ := args["path"].(string)
	ref, _ := args["ref"].(string)

	if owner == "" || repo == "" || path == "" {
		return mcp.NewToolResultError("owner, repo, and path are required"), nil
	}

	gh, err := getGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	opts := &github.RepositoryContentGetOptions{}
	if ref != "" {
		opts.Ref = ref
	}

	fileContent, _, resp, err := gh.Repositories.GetContents(ctx, owner, repo, path, opts)
	logRateLimit(resp)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error reading '%s': %v", path, err)), nil
	}

	if fileContent == nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ '%s' is a directory. Use list_directory instead.", path)), nil
	}

	// go-github's GetContent() already handles base64 decoding internally.
	// Do NOT manually decode — it returns the decoded plaintext string directly.
	rawContent, err := fileContent.GetContent()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Failed to get file content: %v", err)), nil
	}

	// Binary detection: look for null bytes in the decoded content
	isBinary := false
	for _, b := range []byte(rawContent) {
		if b == 0 {
			isBinary = true
			break
		}
	}
	if isBinary {
		return mcp.NewToolResultText(fmt.Sprintf(
			"ℹ️ '%s' appears to be a binary file (SHA: %s). Raw content not returned.",
			path, fileContent.GetSHA(),
		)), nil
	}

	return mcp.NewToolResultText(rawContent), nil
}

// ─── Tool: list_directory ──────────────────────────────────────────────────

func listDirectoryTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	path, _ := args["path"].(string)
	ref, _ := args["ref"].(string)

	if owner == "" || repo == "" {
		return mcp.NewToolResultError("owner and repo are required"), nil
	}

	gh, err := getGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	opts := &github.RepositoryContentGetOptions{}
	if ref != "" {
		opts.Ref = ref
	}

	_, dirContents, resp, err := gh.Repositories.GetContents(ctx, owner, repo, path, opts)
	logRateLimit(resp)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error listing '%s': %v", path, err)), nil
	}

	if dirContents == nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ '%s' is a file. Use get_file_content instead.", path)), nil
	}

	displayPath := path
	if displayPath == "" {
		displayPath = "/"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📂 **%s/%s** `%s`\n\n", owner, repo, displayPath))

	const maxEntries = 200
	count := 0
	for _, entry := range dirContents {
		if count >= maxEntries {
			sb.WriteString(fmt.Sprintf("\n⚠️ Output capped at %d entries.\n", maxEntries))
			break
		}
		icon := "📄"
		if entry.GetType() == "dir" {
			icon = "📁"
		}
		sb.WriteString(fmt.Sprintf("%s `%s`  _(sha: %s)_\n", icon, entry.GetName(), entry.GetSHA()[:8]))
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

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)

	if owner == "" || repo == "" {
		return mcp.NewToolResultError("owner and repo are required"), nil
	}

	gh, err := getGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	r, resp, err := gh.Repositories.Get(ctx, owner, repo)
	logRateLimit(resp)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Error fetching repo '%s/%s': %v", owner, repo, err)), nil
	}

	visibility := "public"
	if r.GetPrivate() {
		visibility = "private"
	}

	output := fmt.Sprintf(
		"## 📦 %s\n\n"+
			"- **Description**: %s\n"+
			"- **Visibility**: %s\n"+
			"- **Default branch**: `%s`\n"+
			"- **Language**: %s\n"+
			"- **Stars**: ⭐ %d\n"+
			"- **Forks**: 🍴 %d\n"+
			"- **Open Issues**: 🐛 %d\n"+
			"- **Last updated**: %s\n"+
			"- **Clone URL**: %s\n",
		r.GetFullName(),
		r.GetDescription(),
		visibility,
		r.GetDefaultBranch(),
		r.GetLanguage(),
		r.GetStargazersCount(),
		r.GetForksCount(),
		r.GetOpenIssuesCount(),
		r.GetUpdatedAt().Time.Format("2006-01-02"),
		r.GetCloneURL(),
	)

	return mcp.NewToolResultText(output), nil
}

// ─── Tool: search_code ─────────────────────────────────────────────────────

func searchCodeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	owner, _ := args["owner"].(string)
	repo, _ := args["repo"].(string)
	query, _ := args["query"].(string)

	perPage := 10
	if pp, ok := args["per_page"].(float64); ok {
		perPage = int(pp)
		if perPage > 30 {
			perPage = 30
		}
	}

	if owner == "" || repo == "" || query == "" {
		return mcp.NewToolResultError("owner, repo, and query are required"), nil
	}

	gh, err := getGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ %v", err)), nil
	}

	// Scope the search to the specific repository
	scopedQuery := fmt.Sprintf("%s repo:%s/%s", query, owner, repo)

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: perPage},
	}

	results, resp, err := gh.Search.Code(ctx, scopedQuery, opts)
	logRateLimit(resp)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("❌ Search failed for '%s': %v", query, err)), nil
	}

	if results.GetTotal() == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("🔍 No results for `%s` in `%s/%s`.", query, owner, repo)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 **%d result(s)** for `%s` in `%s/%s`\n\n",
		results.GetTotal(), query, owner, repo))

	for i, item := range results.CodeResults {
		sb.WriteString(fmt.Sprintf("### %d. `%s`\n", i+1, item.GetPath()))
		sb.WriteString(fmt.Sprintf("   - **File**: [%s](%s)\n", item.GetName(), item.GetHTMLURL()))
		if len(item.TextMatches) > 0 {
			for _, tm := range item.TextMatches {
				fragment := tm.GetFragment()
				if fragment != "" {
					sb.WriteString(fmt.Sprintf("   - **Match**: `...%s...`\n",
						strings.TrimSpace(fragment)))
				}
			}
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// ─── Main ──────────────────────────────────────────────────────────────────

func main() {
	s := server.NewMCPServer("McpGitHubReader", "1.0.0")

	// Tool: get_file_content
	getFileContent := mcp.NewTool("get_file_content",
		mcp.WithDescription(
			"Read the text content of a file from a GitHub repository. "+
				"Returns decoded UTF-8 text. Binary files are gracefully detected and skipped. "+
				"Requires GITHUB_TOKEN env var (PAT with 'repo' scope for private repos).",
		),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner (user or org)")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path within the repo (e.g. 'src/main.go')")),
		mcp.WithString("ref", mcp.Description("Branch, tag, or commit SHA (defaults to repo default branch)")),
	)
	s.AddTool(getFileContent, getFileContentTool)

	// Tool: list_directory
	listDirectory := mcp.NewTool("list_directory",
		mcp.WithDescription(
			"List files and subdirectories at a given path in a GitHub repository. "+
				"Returns a Markdown-formatted tree with icons. Capped at 200 entries.",
		),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner (user or org)")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("path", mcp.Description("Directory path within the repo (empty string = root)")),
		mcp.WithString("ref", mcp.Description("Branch, tag, or commit SHA (defaults to repo default branch)")),
	)
	s.AddTool(listDirectory, listDirectoryTool)

	// Tool: get_repository_info
	getRepositoryInfo := mcp.NewTool("get_repository_info",
		mcp.WithDescription(
			"Fetch metadata for a GitHub repository: description, default branch, language, "+
				"star count, fork count, open issues, visibility, and last updated date.",
		),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner (user or org)")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
	)
	s.AddTool(getRepositoryInfo, getRepositoryInfoTool)

	// Tool: search_code
	searchCode := mcp.NewTool("search_code",
		mcp.WithDescription(
			"Search code within a specific GitHub repository using GitHub's Search API. "+
				"Returns matching file paths and text fragments. "+
				"Note: GitHub Search API has a separate rate limit of 30 requests/min with a PAT. "+
				"To get text match fragments, the server uses the 'text-match' media type automatically.",
		),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner (user or org)")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (e.g. 'func handleRequest')")),
		mcp.WithNumber("per_page", mcp.Description("Number of results to return (default: 10, max: 30)")),
	)
	s.AddTool(searchCode, searchCodeTool)

	server.ServeStdio(s)
}
