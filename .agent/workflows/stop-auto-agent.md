---
description: Immediately shut down all agents and the monitoring system.
---
# Stop Auto Agent Workflow

Use this when you want to perform maintenance, update code, or stop CPU consumption. It effectively "cuts the power" to the system. Any pending tasks or code execution will stop immediately. You can later call `continue-auto-agent` to revive them.

// turbo-all
1. Shut down immediately (killing processes):
`bash /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/stop_agents.sh`
