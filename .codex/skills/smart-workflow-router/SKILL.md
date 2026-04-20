---
description: Smart intent-based workflow router — automatically route user requests to the right workflow
---

# 🧠 Smart Workflow Router

Analyses the USER's request and **automatically routes** it to the most appropriate workflow.

---

## 🚀 Routing

### Step 0: Classify Task Size

Before routing to a workflow, classify the task size:

| Size      | Criteria (ALL must be true for 🟢, ANY triggers 🟡 or 🔴)                                         |
| --------- | ------------------------------------------------------------------------------------------------- |
| 🟢 SMALL  | < 50 LOC, no new files, no DB/API change, no auth/payment/security                                |
| 🟡 MEDIUM | 50-300 LOC, OR new files within existing module, OR internal refactor, no DB migration            |
| 🔴 LARGE  | > 300 LOC, OR new module/service, OR DB migration, OR public API change, OR auth/payment/security |

> If unsure, classify **one tier up** (safer to over-classify than under-classify).

### Step 1: Route by Size + Intent

| If the request involves...                                   | Route to                 |
| ------------------------------------------------------------ | ------------------------ |
| 🟢 Bug fix, typo, config tweak, "quick fix", small change    | `/fast-fix`              |
| 🟡 New utility, refactor, internal feature, "add X"          | `/build`                 |
| 🔴 Requirements, spec, ambiguity, "what should we build"     | `/specifications-writer` |
| Clarify spec, resolve ambiguities, "questions about spec"    | `/clarify-specification` |
| 🔴 Architecture, design, features/, task list, plan          | `/planner-architect`     |
| Consistency check, coverage, "analyze artifacts", "verify"   | `/analyze-artifacts`     |
| 🔴 Implementation of planned tasks (has tasks.md)            | `/coder-implementation`  |
| Code review, audit, quality, "check my code"                 | `/reviewer-audit`        |
| Tests, coverage, "write tests for", TDD                      | `/tester-verification`   |
| Checklist, "generate checklist", "security checklist"        | `/checklist-generator`   |
| Git commit, stage changes                                    | `/git-commit`            |
| Git push, "push to remote"                                   | `/git-push`              |
| GitLab MR, "fix MR comments", review thread                  | `/gitlab-mr-read-fix`    |
| Symbol lookup, blast radius, "who calls X", dependency graph | `/codebase-graph`        |
| Brownfield, existing codebase, legacy analysis               | `/brownfield-scan`       |
| Codebase map, visualisation, "show me the structure"         | `/codemap`               |
| Emergency, production issue, "site is down"                  | `/war-room`              |
| Context too long, "compact session"                          | `/compact-session`       |
| Knowledge wiki, deep documentation                           | `/deepwiki`              |

### Step 2: Confirm Before Routing

```
🧭 Task size: 🟢 SMALL / 🟡 MEDIUM / 🔴 LARGE
   Intent detected: "<detected intent>"
   → Routing to: /[workflow-name]

Proceed? (yes / or tell me what you actually need)
```

> ⚠️ Always confirm. Never auto-execute a destructive workflow (`/git-push`, `/war-room`) without explicit approval. The
> USER can always override size classification by invoking a specific workflow directly.

### Step 3: Execute

1. `@mcp:skill-router` (`search_skills`) — find relevant skills for the intent.
2. Read `**/workflows/[workflow-name].md` and follow it exactly.

---

## 🛡️ Fallback

If confidence < 80%, present the top 2–3 candidates:

```
🤔 I'm not sure which workflow fits best. Did you mean:
1. /fast-fix — quick fix, small change (🟢 SMALL)
2. /build — build a feature without full ceremony (🟡 MEDIUM)
3. /planner-architect — design the solution first (🔴 LARGE)

Which one?
```
