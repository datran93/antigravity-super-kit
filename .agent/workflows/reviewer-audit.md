---
description: Structured workflow for Code Review and Quality Audit. Orchestrates feedback loops to the Coder role.
---

# 🔍 Reviewer / Audit Workflow

This workflow guides you to perform rigorous code reviews on your own implementations, ensuring adherence to Clean Code, Testability, and Project Standards before advancing to testing.

## 🚀 Audit Phase

### Phase 1: Review Intake & Filtering 📥
Evaluate if a distinct review phase is necessary.
- **Bypass Rule**: If the task has low complexity (e.g. <= 3) and is low-impact (e.g., minor UI/Text changes), you may bypass this distinct audit and transition directly to the `tester` role.
- **For High-Impact / High-Complexity Tasks**: Load the context of the recent code changes. Identify the target quality standards using `@mcp:skill-router` (`search_skills`) and understand what was implemented in the `coder` role.

### Phase 2: Rigorous Code Audit 🔍
Perform a deep analysis of the changes.
- **Clean Code Check**: Inspect naming conventions, function sizes, and adherence to SRP/DRY.
- **Testability Check**: Ensure the code is decoupled and easily testable (e.g., uses Dependency Injection).
- **Security & Standards Audit**: Look for common vulnerabilities or violations of industry best practices.

### Phase 3: Feedback & Resolution 📝
Act on the audit results.
- **NEEDS FIX**: If issues are found, list them clearly, transition back to the `coder` role, and fix them immediately.
- **APPROVED**: If code meets standards, proceed to the next testing or planning phase.
- Details any potential long-term maintenance concerns. If significant technical debt is found, transfer to `planner` role and add new refactoring tasks to the plan using `@mcp:context-manager` (`add_task_step`)  but isn't a blocker right now.

### Phase 4: Role Transition 🔄
- Successfully passing the review means transitioning to the `tester` role or continuing the plan.

## 🔴 Critical Constraints
1. **Strictly Audit First**: Evaluate the code objectively before trying to fix it.
2. **Internal Governance**: Do not proceed to testing if the code does not meet the basic quality baseline.
3. **Role Anchoring**: ALWAYS prefix every conversational response with `[Role: 🔍 Reviewer]`.

---

> [!TIP]
> Focus your feedback on "Why" something should change, not just "What". Keep yourself accountable to the project's architectural standards.
