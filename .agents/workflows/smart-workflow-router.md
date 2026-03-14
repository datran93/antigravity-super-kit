---
description: Smart intent-based workflow router — automatically route user requests to the right workflow
---

# 🧠 Smart Workflow Router

This workflow analyses the USER's natural language request and **automatically routes** it to the most appropriate
specialist workflow, eliminating the need for the user to remember specific slash commands.

---

## 🚀 Routing Phase

### Step 1: Classify Intent

Read the USER's message and extract the **primary intent** using the following classification table:

| If the request involves...                               | Route to                 |
| -------------------------------------------------------- | ------------------------ |
| Requirements, spec, ambiguity, "what should we build"    | `/specifications-writer` |
| Architecture, design, DESIGN.md, task list, plan         | `/planner-architect`     |
| Implementation, writing code, "build X", "add feature Y" | `/coder-implementation`  |
| Code review, audit, quality, "check my code"             | `/reviewer-audit`        |
| Tests, coverage, "write tests for", TDD                  | `/tester-verification`   |
| Git commit, stage changes, "commit this"                 | `/git-commit`            |
| Git push, "push to remote", "publish branch"             | `/git-push`              |
| GitLab MR, "fix MR comments", review thread              | `/gitlab-mr-read-fix`    |
| Brownfield, existing codebase, legacy analysis           | `/brownfield-scan`       |
| Codebase map, visualisation, "show me the structure"     | `/codemap`               |
| Emergency, production issue, "site is down"              | `/war-room`              |
| Context too long, "compact session", memory overflow     | `/compact-session`       |
| Knowledge wiki, deep documentation                       | `/deepwiki`              |

### Step 2: Confirm Before Routing

Before executing the routed workflow, present a **one-line confirmation** to the USER:

```
🧭 Intent detected: "<detected intent>"
→ Routing to: /[workflow-name]

Proceed? (yes / or tell me what you actually need)
```

> ⚠️ Always confirm. Never auto-execute a destructive workflow (`/git-push`, `/war-room`) without explicit user
> approval.

### Step 3: Execute Routed Workflow

Once the USER confirms:

1. Use `@mcp:skill-router` (`search_skills`) with the detected intent as query to find any relevant skills to load
   before starting.
2. Read the workflow file: `.agents/workflows/[workflow-name].md`
3. Follow the workflow exactly — do not modify scope based on the routing decision.

---

## ⚡ Express Mode

If the USER prefixes their message with `!` (e.g., `! add auth to the API`), skip the confirmation step and route
directly. Use this only when:

- The intent is unambiguous
- The routed workflow is non-destructive

---

## 🛡️ Fallback

If confidence in intent classification is below 80% (multiple equally plausible routes), present the top 2–3 candidates
and ask the USER to choose:

```
🤔 I'm not sure which workflow fits best. Did you mean:
1. /planner-architect — design the solution first
2. /coder-implementation — start building immediately
3. /specifications-writer — clarify requirements first

Which one? (or describe your goal in more detail)
```

---

## 📝 Metadata

- **Slash command**: `/smart-route`
- **Version**: 1.0.0
- **Maintained by**: Planner role
