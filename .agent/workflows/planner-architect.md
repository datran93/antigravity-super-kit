---
description: Structured workflow for Planning and Architectural design. Orchestrates context discovery, task planning, and Coder delegation.
---

# 🏗️ Planner / Architect Workflow (The Orchestrator)

This workflow guides the **persistent Planner Agent** to analyze requirements, map the codebase, design the architecture, and orchestrate the execution pipeline by summoning ephemeral subagents.

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

### Phase 4: Task Delegation 🤝 (Sequential Mode)
Execute the plan one step at a time via ephemeral workers.
- **GATE**: Call `@mcp:mcp-multi-agent` (`enforce_socratic_gate`) for high-impact decisions with the USER.
- **SUMMON**: Use `@mcp:mcp-multi-agent` (`delegate_to_subagent`) with `run_background=False`.
- Target the appropriate role (`coder`, `reviewer`, `tester`) based on the task nature.
- **WAIT**: The Planner process will pause until the subagent returns its final summary.

### Phase 5: Result Analysis & Pipeline Routing 🔄
Analyze the returned summary to determine the next path.
- **Coder Success?** -> Summon `reviewer` or `tester`.
- **Reviewer Found Issues?** -> Summon `coder` with the review notes to fix.
- **Tester Failed?** -> Summon `coder` with the failure logs to fix.
- **Pass?** -> Mark step as complete via `@mcp:context-manager` (`complete_task_step`).

### Phase 6: Mission Success 🏁
- Once all tasks in the plan are marked as complete, synthesize the final walkthrough for the USER.
- Update the final architecture docs and save the project context.

## 🔴 Critical Constraints
1. **Ownership of Completion**: The Planner is the ONLY agent permitted to call `complete_task_step`.
2. **Ephemeral Coordination**: Do not expect workers to be running. Always summon them via `delegate_to_subagent`.
3. **Summarization**: Use the technical summaries returned by tools to build a cumulative understanding of the project's progress.

---

## 📌 Usage Example
`/planner-architect "Implementing a new user authentication flow with JWT and Redis caching"`
