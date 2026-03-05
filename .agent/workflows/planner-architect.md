---
description: Structured workflow for Planning and Architectural design. Orchestrates context discovery, task planning, and explicit task execution.
---

# 🏗️ Planner / Architect Workflow (The Orchestrator)

This workflow guides you to analyze requirements, map the codebase, design the architecture, and orchestrate the execution pipeline by performing the tasks directly yourself, rather than delegating to subagents.

## 🚀 Execution Phase

### Phase 1: Contextual Discovery 🔍
Use MCP tools to build a comprehensive map of the impact area.
- Use `@mcp:ast-explorer` to understand structural relationships (Py/Go/JS/TS).
- Use `find_by_name` and `grep_search` to locate relevant business logic.
- Use `@mcp:context7` to research latest syntax or library patterns.

### Phase 2: Architectural Design 🏗️
Define the "North Star" for the implementation.
- Identify the core patterns (e.g., Clean Architecture, DDD).
- Document architectural decisions in `DESIGN.md` if necessary.

### Phase 3: Task Plan Initialization 📋
Initialize the lifecycle of the task in the project state.
- Call `@mcp:context-manager` (`initialize_task_plan`) with a detailed list of atomic, executable steps.
- Set up checkpoints using `@mcp:context-manager` (`save_checkpoint`) at critical milestones.

### Phase 4: Task Execution 🤝 (Self-Execution)
Execute the plan one step at a time by taking on the required roles yourself.
- **GATE**: Call `@mcp:mcp-multi-agent` (`enforce_socratic_gate`) for high-impact decisions with the USER.
- **EXECUTE**: Switch your mindset to the appropriate role (`coder`, `reviewer`, `tester`) based on the task nature and perform the work directly.
- Read the corresponding `.agent/workflows/<role>.md` if needed to understand the expectations of that role.

### Phase 5: Result Analysis & Pipeline Routing 🔄
Analyze the result of your work to determine the next path.
- **Coder Success?** -> Switch to `reviewer` or `tester` role to verify.
- **Reviewer Found Issues?** -> Switch back to `coder` role to fix the issues.
- **Tester Failed?** -> Switch back to `coder` role with the failure logs to fix.
- **Pass?** -> Mark step as complete via `@mcp:context-manager` (`complete_task_step`). Make sure to pass `active_files` to maintain a precise list of files modified during this step.
- **New requirements discovered?** -> Use `@mcp:context-manager` (`add_task_step`) to dynamically append new steps to the current task plan.

### Phase 6: Mission Success 🏁
- Once all tasks in the plan are marked as complete, synthesize the final walkthrough for the USER.
- Update the final architecture docs and save the project context.

## 🔴 Critical Constraints
1. **Ownership of Completion**: You must ensure high quality before calling `complete_task_step`.
2. **Self-Execution**: Do not delegate to subagents via `delegate_to_subagent`. Perform the actions yourself using your tools.
3. **Summarization**: Keep track of the project's progress logically in your context.

---

## 📌 Usage Example
`/planner-architect "Implementing a new user authentication flow with JWT and Redis caching"`
