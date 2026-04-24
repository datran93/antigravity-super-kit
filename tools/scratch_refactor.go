package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
)

func main() {
	content, err := os.ReadFile("tools/mcp-context-manager-go/main.go")
	if err != nil {
		panic(err)
	}

	// First pass: replace the start of the function definition
	reStart := regexp.MustCompile(`(?s)mcp\.NewTool\("([^"]+)"(.*?\n\s+), func\(ctx context\.Context, req mcp\.CallToolRequest\) \(\*mcp\.CallToolResult, error\) \{`)
	
	newContent := reStart.ReplaceAllFunc(content, func(match []byte) []byte {
		matches := reStart.FindSubmatch(match)
		toolName := string(matches[1])
		return []byte(fmt.Sprintf(`mcp.NewTool("%s"%s, WithMiddlewares("%s", func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {`, toolName, string(matches[2]), toolName))
	})

	// Second pass: replace the closing bracket. We know each mcpServer.AddTool( ... }) is currently at the end.
	// We'll replace `\n	})\n` with `\n	}))\n`
	// Wait, some might have multiple newlines or different spacing.
	// Let's just find `\n	})\n`
	
	newContent = bytes.ReplaceAll(newContent, []byte("\n\t})\n"), []byte("\n\t}))\n"))

	if err := os.WriteFile("tools/mcp-context-manager-go/main.go", newContent, 0644); err != nil {
		panic(err)
	}
	fmt.Println("Success")
}
