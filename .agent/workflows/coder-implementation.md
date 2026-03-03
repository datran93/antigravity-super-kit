---
description: Structured workflow for Code Implementation by Ephemeral Workers. Handles task execution and technical summary generation.
---

# 💻 Coder / Implementation Workflow (Ephemeral)

This workflow guides an **ephemeral Code Subagent** through implementing atomic tasks defined by the Planner. It emphasizes technical excellence, clean code, and strict adherence to the project's architectural standards before returning a technical summary and exiting.

## 🚀 Implementation Phase

### Phase 1: Task Intake 📥
Load the context provided by the Planner.
- Analyze the `task_description` and `context_files` provided during summoning.
- Review any feedback or previous summaries in the history to understand the current state.

### Phase 2: Skill & Pattern Alignment 🔍
Align the implementation with project-specific standards.
- Use `@mcp:skill-router` (`search_skills`) to find relevant clean code patterns and language-specific best practices for the task.
- Review specific `SKILL.md` files referenced in the instructions to ensure architectural compliance.

### Phase 3: Execution & Engineering 🛠️
Write high-quality, maintainable code.
- Implement the changes using `replace_file_content` or `write_to_file`.
- **MANDATORY**: Follow **Clean Code** principles (naming, small functions, SRP).
- **MANDATORY**: Ensure code is **Testable** (Dependency Injection, modularity).
- Focus only on the atomic task assigned. Do not over-engineer or touch unrelated files.

### Phase 4: Verification & Summary 📝
Prepare the report for the Planner.
- Run lightweight verification commands (e.g., `go build`, `tsc --noEmit`, `python -m py_compile`) to ensure no syntax errors.
- Synthesize a concise **Technical Summary** of your work:
    - Files modified.
    - Logic implemented/refactored.
    - Any technical debt or edge cases identified.
- **DO NOT** call `complete_task_step`.

### Phase 5: Termination ⚰️
- Output the Technical Summary as your final message.
- The subagent process will be destroyed by the environment after this step.

## 🔴 Critical Constraints
1. **No Project Ownership**: You are a temporary worker. Do not mark tasks as complete or manage the plan.
2. **Quality First**: Never ignore lint errors or violations of the design system.
3. **No Co-Authored-By**: When making git commits, DO NOT add 'Co-authored-by' or any agent information.
4. **Direct Feedback**: If you encounter issues that prevent completion, report them clearly in your summary so the Planner can decide the next step.

---

> [!TIP]
> Always verify that your changes don't break existing modularity. If you needed to refactor common code, document it clearly for the next agent (Reviewer/Tester).
