---
description: Display agent and project status. Progress tracking and status board.
---

# /status - Project Health & Progress Report

Guide agents to analyze and report project status, health, and progress systematically.

---

## When to Use

- `/status` - **Full Project Status Data**
- `/status active` - What is currently running?
- `/status plan` - Compare progress vs PLAN.md
- `/status health` - Check system health/tests

---

## Phase 1: Context Gathering 🕵️

### Step 1.1: Project Metadata

Gather core project information:

```markdown
### Project Context

**Name:** [Project Name] **Path:** `[Current Working Directory]` **Branch:** `[Git Branch]` **Modified Files:** [Count
of uncommitted files]
```

### Step 1.2: Active tasks

Check `agent-docs/task.md` or `agent-docs/PLAN.md` for active items:

```markdown
### Active Tasks

- [ ] **Current Focus:** [Task Name]
- [ ] **Status:** 🏃 In Progress | ⏸️ Paused | 🛑 Blocked
- [ ] **Agent:** [e.g. backend-specialist]
```

---

## Phase 2: health Check 🩺

### Step 2.1: System Health

Verify core systems are operational:

- [ ] **Build:** `[Build Command]` → Pass/Fail
- [ ] **Tests:** `[Test Command]` → Pass/Fail (or last run status)
- [ ] **Lint:** `[Lint Command]` → Pass/Fail
- [ ] **Local Server:** Running at `[localhost:port]`?

### Step 2.2: file Statistics

Analyze recent changes:

- **New Files (24h):** [Count]
- **Modified (24h):** [Count]
- **Key Modules Touched:** [List critical paths modified]

---

## Phase 3: Progress Analysis 📈

### Step 3.1: Plan verification

Compare current state against `agent-docs/PLAN.md` (if exists):

```markdown
### Plan vs Reality

**Total Steps:** [Total] **Completed:** [Count] ([%]) **Remaining:** [Count] **On Track?** ✅ Yes | ⚠️ Risk | ❌ Behind
```

### Step 3.2: Risk Detection

Identify potential issues:

- **Complexity Risk:** High cyclomatic complexity in new code?
- **Test Gap:** New code without tests?
- **Dependency Risk:** New vulnerabilities?

---

## Phase 4: Reporting 📝

### Step 4.1: Status Summary

Compile findings into a structured report.

```markdown
# 📊 Project Status Report

## 🚦 Executive Summary

**Status:** 🟢 On Track | 🟡 At Risk | 🔴 Blocked **Focus:** [Current Feature/Task] **Completion:** [X]% of current
milestone

## 🔨 Recent Activity

- [x] Completed: [Task A]
- [x] Completed: [Task B]
- [ ] In Progress: [Task C]

## 🩺 Health Check

| Check      | Status  | Notes        |
| :--------- | :------ | :----------- |
| **Build**  | ✅ Pass | [Build time] |
| **Tests**  | ✅ Pass | [Test count] |
| **Server** | ✅ Up   | [URL]        |

## ⚠️ Blockers & Risks

- [Blocker 1]: [Description]
- [Risk 1]: [Description]

## ⏭️ Next Steps

1. [Action Item 1]
2. [Action Item 2]
```

### Step 4.2: Save & Record

1. Save report to `agent-docs/STATUS-[date].md` (Optional, usually for milestones).
2. Or just output to chat for immediate user review.

---

## Quick Reference

### Status Indicators

| Indicator       | Meaning                                                    |
| :-------------- | :--------------------------------------------------------- |
| 🟢 **On Track** | Plan proceeding as expected, no blockers.                  |
| 🟡 **At Risk**  | Minor issues, tests failing, or slightly behind schedule.  |
| 🔴 **Blocked**  | Critical issue preventing progress (e.g. API down, bug).   |
| ⏸️ **Paused**   | Development halted pending user input/external dependency. |

### Diagnostic Commands

- **Git Status:** `git status -s`
- **Recent Logs:** `tail -n 50 .agent/logs/latest.log` (if available)
- **Active Ports:** `lsof -i :[port]` or `netstat`

---

## Anti-Patterns (AVOID)

| ❌ Don't          | ✅ Do                                                      |
| :---------------- | :--------------------------------------------------------- |
| **Guess Status**  | Verify with commands (git, build, test)                    |
| **Hide Errors**   | Report failing tests/builds immediately                    |
| **Vague Updates** | Be specific: "Auth module finished" vs "Worked on backend" |
| **Ignore Plan**   | Always reference `PLAN.md` or `task.md`                    |
