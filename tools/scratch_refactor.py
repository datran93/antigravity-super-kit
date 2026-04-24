import sys
import re

with open("tools/mcp-context-manager-go/main.go", "r") as f:
    lines = f.readlines()

out = []
in_tool = False
tool_name = ""
for line in lines:
    m = re.search(r'mcpServer\.AddTool\(mcp\.NewTool\("([^"]+)"', line)
    if m:
        tool_name = m.group(1)
        in_tool = True
    
    if in_tool and '), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {' in line:
        line = line.replace('), func(ctx', f'), WithMiddlewares("{tool_name}", func(ctx')
    elif in_tool and line.strip() == '})':
        line = line.replace('})', '}))')
        in_tool = False

    out.append(line)

with open("tools/mcp-context-manager-go/main.go", "w") as f:
    f.writelines(out)

print("Done")
