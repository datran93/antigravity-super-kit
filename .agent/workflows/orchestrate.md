---
description:
  Coordinate multiple agents for complex tasks. Use for multi-perspective analysis, comprehensive reviews, or tasks
  requiring different domain expertise.
---

# /orchestrate - Multi-Agent Collaboration Workflow

Turn complex problems into solved solutions by coordinating a team of specialized agents.

---

## When to Use

- `/orchestrate [complex task]` - **Start full orchestration**
- `/orchestrate plan` - Generate multi-agent plan only
- `/orchestrate review` - Multi-perspective review
- Tasks requiring **3+ domains** (e.g. Layout + API + DB + Security)

---

## 🔴 Critical Rules

1. **You are the Manager**: You do not write code; you direct others.
2. **Context is King**: Pass FULL context to every agent.
3. **Sequential Logic**: Plan → Execute → Verify.
4. **Validation**: Orchestration fails if < 3 agents are used.
5. **No Hallucination**: Only use available agents.

---

## Phase 1: The Setup 📋

### Step 1.1: Analyze the Request

Break down the user's request into domain components:

```markdown
### Domain Analysis

| Component    | Required Expertise | Best Agent            |
| :----------- | :----------------- | :-------------------- |
| **Frontend** | React/Vue, CSS     | `frontend-specialist` |
| **Backend**  | API, Node/Python   | `backend-specialist`  |
| **Data**     | SQL/Schema         | `database-architect`  |
| **Security** | Auth/OWASP         | `security-auditor`    |
| **DevOps**   | CI/CD, Docker      | `devops-engineer`     |
```

### Step 1.2: Select Your Team

**Minimum Requirement:** 3 Agents.

**Common Squads:**

- **Feature Squad:** `project-planner` + `frontend-specialist` + `backend-specialist`
- **Quality Squad:** `debugger` + `test-engineer` + `security-auditor`
- **Launch Squad:** `devops-engineer` + `performance-optimizer` + `seo-specialist`

---

## Phase 2: The Plan 🗺️

### Step 2.1: Create the Master Plan

Invoke `project-planner` to create `agent-docs/PLAN.md`.

**Prompt to Planner:**

> "Review this request: [User Request]. Create a detailed implementation plan involving [List of Agents]. Break down
> tasks for each agent. Output to docs/PLAN.md."

### Step 2.2: User Approval Gate 🚧

**STOP and ASK:**

> "✅ Plan created at `agent-docs/PLAN.md`.
>
> **Proposed Team:**
>
> 1. `[Agent 1]` - [Role]
> 2. `[Agent 2]` - [Role]
> 3. `[Agent 3]` - [Role]
>
> **Approve execution? (Y/N)**"

---

## Phase 3: Execution (The Orchestra) 🎻

**Execute sequentially or in parallel groups depending on dependencies.**

### Step 3.1: Foundation Layer

Usually `database-architect` or `devops-engineer`.

**Prompt:**

> "Execute Phase 1 of PLAN.md. [Specific Task]. **Context:** [User Request] + [Decisions] **Output:** Verified
> SQL/Config files."

### Step 3.2: Core Implementation

Usually `backend-specialist` and `frontend-specialist`.

**Prompt:**

> "Execute Phase 2 of PLAN.md. **Context:** Phase 1 completed by [Agent]. **Task:** Build [Feature]. **Constraint:**
> Match design in [Design Doc]."

### Step 3.3: Quality Assurance

Usually `test-engineer` or `security-auditor`.

**Prompt:**

> "Review work from Phase 2. **Task:** generate tests / audit security. **Target:** [Files Created]. **Goal:** Ensure
> production readiness."

---

## Phase 4: Integration & Synthesis 🧩

### Step 4.1: Verify Integration

Ensure pieces fit together.

- Does Frontend talk to Backend?
- Does Backend talk to Database?
- Do Build Scripts work?

### Step 4.2: Final Polish

Invoke `documentation-writer` to update README/Docs if needed.

---

## Phase 5: Reporting 📝

### Step 5.1: Create Orchestration Report

Compile findings into `agent-docs/ORCHESTRATE-{slug}.md`.

```markdown
# 🎼 Orchestration Report: [Task Name]

## 👥 Team

- **Manager:** Orchestrator
- **Squad:** [List Agents]

## 🔄 Execution Log

1. **[Agent 1]**: [Action] - ✅ Success
2. **[Agent 2]**: [Action] - ✅ Success
3. **[Agent 3]**: [Action] - ✅ Success

## 📦 Deliverables

- [ ] `agent-docs/PLAN.md`
- [ ] [Feature Code]
- [ ] [Tests]

## 🛡️ Verification

- Security Scan: [Pass/Fail]
- Lint Check: [Pass/Fail]
```

### Step 5.2: Final Notification

Notify user: `✅ Orchestration Complete! Report: agent-docs/ORCHESTRATE-{slug}.md`

---

## Quick Reference

### Agent Capabilities

| Agent                 | Best For                   |
| :-------------------- | :------------------------- |
| `project-planner`     | Breaking down big tasks    |
| `frontend-specialist` | UI, Components, CSS        |
| `backend-specialist`  | API, Logic, DB Integration |
| `database-architect`  | Schema, Migrations, SQL    |
| `test-engineer`       | Unit/E2E Tests, QA         |
| `security-auditor`    | Vuln Code Review, Auth     |
| `devops-engineer`     | Docker, CI/CD, Cloud       |

### Orchestrator Anti-Patterns

| ❌ Don't           | ✅ Do                            |
| :----------------- | :------------------------------- |
| **Micromanage**    | Give high-level goals + context  |
| **Forget Context** | Pass full history to every agent |
| **Do it yourself** | Delegate EVERYTHING              |
| **Ignore Errors**  | Stop and fix immediately         |
| **Skip Plan**      | Always plan before coding        |
