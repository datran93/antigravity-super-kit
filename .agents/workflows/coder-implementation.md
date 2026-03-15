---
description:
  Structured workflow for Code Implementation. Reads DESIGN.md and the Planner's task list, executes each task in order,
  and reports a clear summary of all changes to the USER upon completion.
---

# 💻 Coder Workflow

Implements each task defined by the Planner atomically, then delivers a structured report. The Coder does NOT commit.

---

## 🚀 Implementation Phases

### Phase 0: Read Design & Task List 📖

1. **Read `DESIGN.md`** — architecture, data models, file structure, constraints.
2. **Load the task list** — Call `@mcp:context-manager` (`load_checkpoint`).
3. **Confirm scope** — identify files to create/modify and in what order.

> ❌ Do NOT start writing code before completing this phase.

---

### Phase 1: Task Intake & Intent Lock 📥

For each Action (one at a time, in order):

- Read the Action's description, target files, and **Verification Command**.
- Call `@mcp:context-manager` (`declare_intent`) to lock files for this Action.
- Call `@mcp:context-manager` (`check_intent_lock`) before any edit. On Scope Creep ALARM → stop and ask the USER.

---

### Phase 2: Skill & Pattern Alignment 🔍

- `@mcp:skill-router` (`search_skills`) — find relevant patterns and best practices.
- `@mcp:context7` (`query-docs`) — verify latest API specs (avoid syntax hallucinations).
- Cross-reference `DESIGN.md` for contracts to respect.

---

### Phase 3: Execution 🛠️

- **NO BLIND WRITES**: Read every file before modifying it.
- Follow Clean Code: clear naming, small focused functions, SOLID.
- Code must be testable: Dependency Injection, no hardcoded globals.
- Stay strictly within the locked Bounded Context.

---

### Phase 4: Verification ✅

- Run the **Verification Command** (e.g. `go test ./...`, `npm run lint`).
- On **fail**: Call `record_failure`. Fix and re-run. After **3 consecutive failures**, stop and ask the USER.
- On **pass**: Note the result. Do NOT commit — committing is the Planner's responsibility.

> On phase-complete signal (`💡 Phase Px is complete — run /compact-session`): stop and run `/compact-session` before
> the next phase.

Repeat **Phase 1 → Phase 4** for each remaining Action.

---

### Phase 5: Final Report to USER 📋

```
## ✅ Implementation Complete

### Changes Made
| File | Action | Purpose |
|------|--------|---------|

### What was built
<1-2 sentence summary>

### Verification
- All Verification Commands passed: ✅

### Notes / Known Limitations
<Technical debt, deferred items, edge cases>
```

> 🛑 **STOP HERE.** The USER decides the next step (`/tester-verification`, `/reviewer-audit`, etc.).

---

## 🔴 Critical Constraints

1. **Read before write**: Never modify a file without reading it first.
2. **Task order matters**: Complete Actions in the Planner's order. Do not skip or reorder.
3. **No hidden failures**: On unexpected blockers, stop and ask the USER immediately.
4. **Stay in scope**: Do not refactor outside the current Action's Bounded Context.
5. **Role Anchoring**: ALWAYS prefix every response with `[Role: 💻 Coder]`.
