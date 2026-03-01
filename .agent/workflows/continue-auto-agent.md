---
description: Resume and continue the existing multi-agent workflow (Loads history from DB, optional --engine flag supported).
---
# Continue Auto Agent Workflow

If you stop the agents or the system restarts, use this command to revive all 4 Agents. They will automatically read the last message in the database and continue.

**Usage**: `/continue-auto-agent [-e|--engine <kilocode|opencode|gemini|copilot>]`

// turbo-all
1. Revive all agent timelines (injecting the `--resume` flag):
`bash /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/continue_agents.sh $ARGUMENTS`

2. Open the Monitoring Dashboard in your browser:
`open http://localhost:6060`
