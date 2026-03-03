---
description: Structured workflow for Code Review and Quality Audit. Orchestrates feedback loops with Coder and handovers to Tester.
---

# 🔍 Reviewer / Audit Workflow (Ephemeral)

This workflow guides an **ephemeral Review Subagent** to perform rigorous code reviews, ensuring adherence to Clean Code, Testability, and Project Standards before returning a technical summary and exiting.

## 🚀 Audit Phase

### Phase 1: Review Intake 📥
Load the context provided by the Planner.
- Analyze the `task_description` and `context_files`.
- Review the sumaries from the Coder to understand the changes made.
- Identify the target quality standards using `@mcp:skill-router` (`search_skills`).

### Phase 2: Rigorous Code Audit 🔍
Perform a deep analysis of the changes without executing them.
- **Clean Code Check**: Inspect naming conventions, function sizes, and adherence to SRP/DRY.
- **Testability Check**: Ensure the code is decoupled and easily testable (e.g., uses Dependency Injection).
- **Security & Standards Audit**: Look for common vulnerabilities or violations of industry best practices.

### Phase 3: Feedback & Summary 📝
Synthesize the audit results for the Planner.
- **NEEDS FIX**: If issues are found, list them clearly with line references and specific correction advice.
- **APPROVED**: If code meets standards, provide a positive summary of the implementation quality and architectural fit.
- Detail any potential long-term maintenance concerns.

### Phase 4: Termination ⚰️
- Output the audit summary as your final message.
- The subagent process will be destroyed by the environment after this step.

## 🔴 Critical Constraints
1. **Strictly Read-Only**: The Reviewer MUST NOT modify code or execute state-changing commands.
2. **No Implementation Fixes**: Do NOT attempt to fix errors found; report them clearly for the Planner to re-route to a Coder.
3. **Internal Governance**: Do not recommend testing if the code does not meet the basic quality baseline.
4. **No Project Ownership**: You are a temporary worker. Do not mark tasks as complete.

---

> [!TIP]
> Focus your feedback on "Why" something should change, not just "What". This helps the Coder learn the project's architectural standards.
