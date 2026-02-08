---
description: Create project plan using project-planner agent. No code writing - only plan
---

# /plan - Project Planning Workflow

Guide agents to create comprehensive, actionable plans. **NO CODE WRITING** - only planning.

---

## When to Use

- `/plan [description]` - Start planning workflow
- Keywords: "plan", "roadmap", "breakdown", "strategy", "outline"

---

## 🔴 Critical Rules

1. **NO CODE WRITING** - This command creates plan file only
2. **Socratic Gate** - Ask clarifying questions before planning
3. **Actionable Output** - Every task must be executable
4. **Dynamic Naming** - Plan file named based on task

---

## Phase 1: Context Analysis 🔍

### Step 1.1: Understand the Request

Extract from user request:

| Dimension       | Question                                    |
| --------------- | ------------------------------------------- |
| **What**        | What needs to be planned?                   |
| **Why**         | What is the goal/outcome?                   |
| **Scope**       | What's included? What's excluded?           |
| **Constraints** | Timeline, resources, dependencies?          |
| **Context**     | Greenfield / Enhancement / Migration / Fix? |

### Step 1.2: Clarifying Questions (If Needed)

If request is vague, ASK before planning:

```markdown
### 🔍 Before Planning

1. **Scope**: What are the boundaries of this plan?
2. **Priority**: What is the most important outcome?
3. **Timeline**: Any deadline or milestones?
4. **Dependencies**: Any existing systems to consider?
5. **Resources**: Any constraints (team, tech, budget)?
```

### Step 1.3: Confirm Understanding

```markdown
### ✅ Planning Context Confirmed

**Planning:** [what] **Goal:** [desired outcome] **Scope:** [boundaries] **Constraints:** [limitations]

> Does this look correct? If yes, I'll proceed with planning.
```

---

## Phase 2: Research & Discovery 📚

### Step 2.1: Codebase Check (If Applicable)

For plans involving existing codebase:

1. Check `CODEBASE.md` for structure overview
2. Identify relevant existing components
3. Note dependencies and integration points

### Step 2.2: Identify Stakeholders & Agents

```markdown
### Stakeholders & Agents

| Role                   | Responsibility                |
| ---------------------- | ----------------------------- |
| `@backend-specialist`  | API and business logic        |
| `@database-architect`  | Data layer and schema         |
| `@frontend-specialist` | User interface                |
| `@test-engineer`       | Testing strategy              |
| `@devops-engineer`     | Deployment and infrastructure |
```

---

## Phase 3: Task Breakdown 📋

### Step 3.1: Identify Major Phases

Break work into logical phases:

```markdown
### Phases

| Phase | Name        | Description                 | Duration |
| ----- | ----------- | --------------------------- | -------- |
| 1     | Foundation  | Setup, dependencies, config | X days   |
| 2     | Core Logic  | Main functionality          | X days   |
| 3     | Integration | Connect components          | X days   |
| 4     | Testing     | Verification and QA         | X days   |
| 5     | Deployment  | Release and monitoring      | X days   |
```

### Step 3.2: Task Decomposition

For each phase, break into actionable tasks:

```markdown
### Phase 1: Foundation

- [ ] **1.1 Project Setup**
  - [ ] Initialize project structure
  - [ ] Configure dependencies
  - [ ] Setup environment variables
  - **Agent:** `@backend-specialist`
  - **Deliverable:** Working project skeleton

- [ ] **1.2 Database Schema**
  - [ ] Design data model
  - [ ] Create migrations
  - [ ] Setup seed data
  - **Agent:** `@database-architect`
  - **Deliverable:** Database ready for development
```

### Step 3.3: Task Quality Checklist

Each task must have:

- [ ] **Clear description** - What needs to be done
- [ ] **Assignee** - Which agent handles it
- [ ] **Deliverable** - What is the output
- [ ] **Dependencies** - What must be done first
- [ ] **Verification** - How to know it's done

---

## Phase 4: Dependencies & Order 🔗

### Step 4.1: Dependency Mapping

Identify what depends on what:

```markdown
### Dependency Graph
```

1.1 Project Setup ↓ 1.2 Database Schema → 2.1 Core Models ↓ 2.2 Business Logic → 3.1 API Endpoints ↓ 3.2 Frontend → 4.1
Integration Tests

```

```

### Step 4.2: Critical Path

Identify the longest chain (critical path):

```markdown
### Critical Path

1. Project Setup → Database → Core Logic → API → Frontend → Tests

**Estimated Duration:** X days **Parallel Work Possible:** [list items that can run in parallel]
```

---

## Phase 5: Risk Assessment ⚠️

### Step 5.1: Identify Risks

```markdown
### Risks

| Risk                    | Impact | Probability | Mitigation        |
| ----------------------- | ------ | ----------- | ----------------- |
| [Risk 1]                | High   | Medium      | [How to mitigate] |
| [Risk 2]                | Medium | Low         | [How to mitigate] |
| [Technical uncertainty] | High   | High        | [Spike/POC first] |
```

### Step 5.2: Unknowns & Assumptions

```markdown
### Assumptions Made

- [Assumption 1 - with fallback if wrong]
- [Assumption 2 - with fallback if wrong]

### Unknowns to Resolve

- [ ] [Unknown 1] - Resolve by [method/date]
- [ ] [Unknown 2] - Resolve by [method/date]
```

---

## Phase 6: Verification Strategy ✅

### Step 6.1: Success Criteria

```markdown
### Definition of Done

The plan is complete when:

- [ ] All core features are implemented
- [ ] Tests pass with X% coverage
- [ ] Documentation is updated
- [ ] Deployed to [environment]
- [ ] Stakeholder approval received
```

### Step 6.2: Milestones & Checkpoints

```markdown
### Milestones

| Milestone | Criteria                    | Date       |
| --------- | --------------------------- | ---------- |
| Alpha     | Core features working       | YYYY-MM-DD |
| Beta      | All features, basic testing | YYYY-MM-DD |
| Release   | Full testing, documentation | YYYY-MM-DD |
```

---

## Phase 7: Documentation & Delivery 📝

### Step 7.1: Plan Document Structure

```markdown
# PLAN: [Title]

## Overview

[Brief description of what this plan covers]

## Goals

- [Goal 1]
- [Goal 2]

## Scope

**In Scope:** [what's included] **Out of Scope:** [what's excluded]

## Task Breakdown

### Phase 1: [Name]

- [ ] Task 1.1
- [ ] Task 1.2

### Phase 2: [Name]

- [ ] Task 2.1
- [ ] Task 2.2

## Dependencies

[Dependency graph or list]

## Risks & Mitigations

[Risk table]

## Timeline

[Milestone table]

## Success Criteria

[Definition of done]

## Next Steps

- [ ] Review this plan
- [ ] Run `/create` to start implementation
- [ ] Or modify plan as needed
```

## Save Plan First

1. Save to `.agent/docs/PLAN-{slug}.md`
2. **Slug generation**: Extract 2-3 key words → lowercase → hyphen-separated → max 30 chars
   - "e-commerce site with cart" → `PLAN-ecommerce-cart.md`
   - "mobile app for fitness" → `PLAN-fitness-app.md`
   - "add dark mode feature" → `PLAN-dark-mode.md`

### Step 7.2: Request User Review

After saving, notify and ask for review:

```markdown
✅ **Plan saved:** `.agent/docs/PLAN-{slug}.md`

**Please review the plan and:**

1. ✅ Approve → Run `/create` to start implementation
2. 📝 Edit → Modify the plan file directly
3. 💬 Discuss → Ask questions or request changes
```

---

## Quick Reference

### Workflow Flow

```
Context Analysis → Research → Task Breakdown → Dependencies → Risk Assessment → Verification → SAVE → Review
       ↓              ↓            ↓               ↓               ↓               ↓            ↓        ↓
   Clarify        Codebase     Phases +       Dependency      Identify        Success      Save     Ask user
   + Confirm      + Agents     Tasks          graph           risks           criteria     first    to review
```

### Task Template

```markdown
- [ ] **[Task ID] [Task Name]**
  - [ ] [Subtask 1]
  - [ ] [Subtask 2]
  - **Agent:** `@[agent-name]`
  - **Deliverable:** [what is produced]
  - **Depends on:** [prerequisites]
```

### Naming Examples

| Request                           | Plan File                |
| --------------------------------- | ------------------------ |
| `/plan e-commerce site with cart` | `PLAN-ecommerce-cart.md` |
| `/plan mobile app for fitness`    | `PLAN-fitness-app.md`    |
| `/plan add dark mode feature`     | `PLAN-dark-mode.md`      |
| `/plan refactor auth system`      | `PLAN-auth-refactor.md`  |
| `/plan SaaS dashboard`            | `PLAN-saas-dashboard.md` |

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern               | ✅ Instead                           |
| ---------------------------- | ----------------------------------- |
| Write code during planning   | Only plan - no code                 |
| Vague tasks ("do the thing") | Specific, actionable tasks          |
| No deliverables defined      | Every task has clear output         |
| Skip dependency analysis     | Map dependencies before ordering    |
| No risk assessment           | Identify and mitigate risks         |
| Missing success criteria     | Define how to know it's done        |
| Monolithic tasks             | Break into small, verifiable chunks |

---

## Examples

```bash
/plan e-commerce site with cart and checkout
/plan mobile app for fitness tracking
/plan refactor authentication to use OAuth
/plan migrate database from MySQL to PostgreSQL
/plan implement real-time notifications
```
