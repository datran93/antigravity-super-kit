---
description: Smart intent-based workflow router — automatically route user requests to the right workflow
---

# 🧠 Smart Workflow Router

Analyses the USER's request and **automatically routes** it to the most appropriate workflow.

---

## 🚀 Routing

### Step 1: Classify Intent

| If the request involves...                                 | Route to                 |
| ---------------------------------------------------------- | ------------------------ |
| Requirements, spec, ambiguity, "what should we build"      | `/specifications-writer` |
| Architecture, design, design/design-\*.md, task list, plan | `/planner-architect`     |
| Implementation, writing code, "build X", "add feature Y"   | `/coder-implementation`  |
| Code review, audit, quality, "check my code"               | `/reviewer-audit`        |
| Tests, coverage, "write tests for", TDD                    | `/tester-verification`   |
| Git commit, stage changes                                  | `/git-commit`            |
| Git push, "push to remote"                                 | `/git-push`              |
| GitLab MR, "fix MR comments", review thread                | `/gitlab-mr-read-fix`    |
| Brownfield, existing codebase, legacy analysis             | `/brownfield-scan`       |
| Codebase map, visualisation, "show me the structure"       | `/codemap`               |
| Emergency, production issue, "site is down"                | `/war-room`              |
| Context too long, "compact session"                        | `/compact-session`       |
| Knowledge wiki, deep documentation                         | `/deepwiki`              |

### Step 2: Confirm Before Routing

```
🧭 Intent detected: "<detected intent>"
→ Routing to: /[workflow-name]

Proceed? (yes / or tell me what you actually need)
```

> ⚠️ Always confirm. Never auto-execute a destructive workflow (`/git-push`, `/war-room`) without explicit approval.

### Step 3: Execute

1. `@mcp:skill-router` (`search_skills`) — find relevant skills for the intent.
2. Read `.agents/workflows/[workflow-name].md` and follow it exactly.

---

## 🛡️ Fallback

If confidence < 80%, present the top 2–3 candidates:

```
🤔 I'm not sure which workflow fits best. Did you mean:
1. /planner-architect — design the solution first
2. /coder-implementation — start building immediately
3. /specifications-writer — clarify requirements first

Which one?
```
