---
description: Start a NEW Multi-Agent workflow (Resets memory, clears database).
---
# Start Auto Agent Workflow

Use this when you want Agents to forget all previous history (clear DB) and start a brand-new project in the current Workspace. Make sure to pass `$ARGUMENTS` to provide the initial request/spec.

// turbo-all
1. Reset memory and start the agents:
`bash /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/start_agents.sh "$ARGUMENTS"`

2. Note: If the dashboard is not running, you need to open a new Terminal tab and run this command:
`/Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/.venv/bin/python /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/dashboard.py $(pwd)/.agent_logs/multi_agent_bus.db`

3. Open the Monitoring Dashboard in your browser:
`open http://localhost:6060`
