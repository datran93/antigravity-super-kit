---
description: Structured workflow for Code Implementation. Reads design/design-*.md and the Planner's task list, executes each task
---

# 💻 Coder Workflow

Implements each task defined by the Planner atomically, then delivers a structured report. The Coder does NOT commit.

---

## 🚀 Implementation Phases

### Phase 0: Read Design & Task List 📖

1. **Read `design/design-{task-id}.md`** — architecture, data models, file structure, constraints.
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
- Cross-reference `design/design-{task-id}.md` for contracts to respect.

---

### Phase 3: Execution 🛠️

Choose the execution mode per T-step based on complexity:

#### Mode A — Direct Implementation (default)

Implement inline in the current session:

- **NO BLIND WRITES**: Read every file before modifying it.
- Follow Clean Code: clear naming, small focused functions, SOLID.
- Code must be testable: Dependency Injection, no hardcoded globals.
- Stay strictly within the locked Bounded Context.

#### Mode B — Subagent Delegation (for complex or long T-steps)

Spawn a fresh isolated Gemini subprocess for the T-step via the subagent engine:

````bash
# 1. Write the task file
cat > /tmp/task-T{N}.md << 'EOF'
## Task: [T{N}] <step name>
### Objective
<copy the T-step description>
### Files to Create/Modify
- `<file1>`
### Verification Command
```bash
<verification command>
````

### Acceptance Criteria

<verification command must exit 0> EOF

# 2. Spawn implementer subagent

.agents/subagents/spawn-subagent.sh \
 --role implementer \
 --task-file /tmp/task-T{N}.md \
 --context-files "<file1>,<file2>" \
 --workspace "$PWD" \
  --output-file /tmp/impl-T{N}.md
EXIT=$?

# 3. Handle result

# EXIT 0 = DONE, EXIT 1 = BLOCKED/NEEDS_CONTEXT, EXIT 2 = TIMEOUT, EXIT 3 = ERROR

```

**When to use Mode B:**
- T-step touches 3+ files with complex interdependencies
- Session context is already long (risk of context rot)
- Task has a clear isolated scope (no tight coupling to current session state)
- You want a fresh-eyes review via `spec-reviewer` or `quality-reviewer`

**Handling subagent results:**
- `EXIT 0` → read `/tmp/impl-T{N}.md`, verify, proceed to Phase 4
- `EXIT 1` → read output for NEEDS_CONTEXT details, add missing context, retry (max 3×)
- `EXIT 2` → task too large; split into smaller sub-steps, retry
- `EXIT 3` → process error; check `gemini` CLI availability, escalate to USER

> See `.agents/workflows/dispatch-subagent.md` for the full orchestrator loop with 2-stage review.

---

### Phase 4: Verification ✅

- Run the **Verification Command** (e.g. `go test ./...`, `npm run lint`).
- On **fail**: Call `record_failure`. Fix and re-run. After **3 consecutive failures**, stop and ask the USER.
- On **pass**: Note the result. Do NOT commit — committing is the Planner's responsibility.

Repeat **Phase 1 → Phase 4** for each remaining Action.

---

### Phase 5: Final Report to USER 📋

```

## ✅ Implementation Complete

### Changes Made

| File | Action | Purpose |
| ---- | ------ | ------- |

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

1. **Read before write**: Never modify a file without reading it first (applies to both modes).
2. **Task order matters**: Complete Actions in the Planner's order. Do not skip or reorder.
3. **No hidden failures**: On unexpected blockers, stop and ask the USER immediately.
4. **Stay in scope**: Do not refactor outside the current Action's Bounded Context.
5. **Subagent max retries**: In Mode B, maximum 3 retry attempts per stage before escalating to USER.
6. **Role Anchoring**: ALWAYS prefix every response with `[Role: 💻 Coder]`.
```
