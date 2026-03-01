---
description: Resume and continue the existing multi-agent workflow (Loads history from DB, no arguments required).
---
# Continue Auto Agent Workflow

If you accidentally shut down the system (`/stop-auto-agent`) or your computer restarted, use this command to revive all 4 Agents. They will AUTOMATICALLY read the last message in the database and continue from where they left off. No arguments are required.

// turbo-all
1. Revive all agent timelines (injecting the `--resume` flag):
`bash /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/continue_agents.sh`

2. Open the Monitoring Dashboard in your browser to watch them work:
`open http://localhost:6060`
