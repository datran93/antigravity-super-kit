---
description:
  Dispatch a fresh Gemini CLI subagent per T-step with 2-stage review (spec compliance + code quality). Replaces
  single-session Coder role for complex or long-running task plans.
---

# 🤖 Dispatch Subagent Workflow

Orchestrates per-task subagent execution using `.agents/subagents/spawn-subagent.sh`. Each T-step runs in a **fresh
isolated Gemini process** — no session history leakage.

---

## When to Use This Workflow

Use instead of the standard `/coder-implementation` when:

- The task plan has **3+ T-steps** and context rot is a concern
- Tasks are **independently executable** (no tight shared state between steps)
- You want **2-stage review** (spec compliance → code quality) per task
- The session is already long and context pollution is visible

For simple 1-2 step tasks, `/coder-implementation` is faster.

---

## Prerequisites

- `.agents/subagents/spawn-subagent.sh` must exist and be executable
- `.agents/subagents/templates/*.md` (4 templates) must exist
- `gemini` CLI must be available in PATH (`which gemini`)
- The task plan checkpoint must be loaded (`load_checkpoint`)

---

## Execution Loop (Per T-step)

```
For each T-step in the plan (in order):
  1. PREPARE   — Extract task text + verification command
  2. IMPLEMENT — Dispatch implementer subagent (max 3 attempts)
  3. SPEC      — Dispatch spec-reviewer subagent (max 3 attempts)
  4. QUALITY   — Dispatch quality-reviewer subagent (max 3 attempts)
  5. COMPLETE  — complete_task_step + clear_drift
```

---

## Step-by-Step Instructions

### Step 1: Prepare Task File

For each T-step, write a `/tmp/task-T{N}.md` with:

````markdown
## Task: [T{N}] <step name>

### Objective

<Copy the full T-step description from the plan>

### Files to Create/Modify

- `<file1>`
- `<file2>`

### Verification Command

```bash
<verification command from plan>
```
````

### Acceptance Criteria

<verification command must exit 0>

````

**Identifying relevant context files**: Include only files directly touched by the task.
For Go: include the target `.go` file + its `_test.go`. For shell: include the script itself.

---

### Step 2: Dispatch Implementer (max 3 attempts)

```bash
ATTEMPT=1
MAX=3
while [[ $ATTEMPT -le $MAX ]]; do
  .agents/subagents/spawn-subagent.sh \
    --role implementer \
    --task-file /tmp/task-T{N}.md \
    --context-files "<relevant files>" \
    --workspace "$PWD" \
    --output-file /tmp/impl-T{N}.md
  EXIT=$?

  if [[ $EXIT -eq 0 ]]; then
    echo "✅ Implementer done"
    break
  elif [[ $EXIT -eq 1 ]]; then
    # Read /tmp/impl-T{N}.md for NEEDS_CONTEXT or BLOCKED details
    ATTEMPT=$((ATTEMPT + 1))
    if [[ $ATTEMPT -le $MAX ]]; then
      echo "🔄 Retry $ATTEMPT/$MAX — adding context and re-dispatching"
      # Append missing context to task file, then loop
    else
      echo "🚨 Implementer BLOCKED after $MAX attempts — ESCALATE TO USER"
      exit 1
    fi
  elif [[ $EXIT -eq 2 ]]; then
    echo "⏰ TIMEOUT — consider breaking task into smaller pieces"
    exit 1
  else
    echo "💥 Process error — ESCALATE TO USER"
    exit 1
  fi
done
````

**Handling NEEDS_CONTEXT**: Read `/tmp/impl-T{N}.md`, find what's missing, append it to the task file, retry.

**Handling BLOCKED**: If the implementer cannot proceed after 3 attempts, stop and ask the USER.

---

### Step 3: Dispatch Spec Reviewer (max 3 rounds)

```bash
ROUND=1
MAX=3
while [[ $ROUND -le $MAX ]]; do
  .agents/subagents/spawn-subagent.sh \
    --role spec-reviewer \
    --task-file /tmp/task-T{N}.md \
    --context-files "<modified files>" \
    --workspace "$PWD" \
    --output-file /tmp/spec-T{N}.md
  EXIT=$?

  # Read verdict from output file
  VERDICT=$(grep -i "^APPROVED\|NEEDS_FIXES" /tmp/spec-T{N}.md | tail -1 || echo "APPROVED")

  if [[ $EXIT -eq 0 && "$VERDICT" == *"APPROVED"* ]]; then
    echo "✅ Spec review passed"
    break
  elif [[ "$VERDICT" == *"NEEDS_FIXES"* ]]; then
    ROUND=$((ROUND + 1))
    if [[ $ROUND -le $MAX ]]; then
      echo "🔄 Spec gaps found — re-dispatching implementer with gap list"
      # Append spec review output to task file as "Fixes Required" section, re-implement
      .agents/subagents/spawn-subagent.sh --role implementer \
        --task-file /tmp/task-T{N}.md --workspace "$PWD" \
        --output-file /tmp/impl-T{N}-fix.md
    else
      echo "🚨 Spec still failing after $MAX rounds — ESCALATE TO USER"
      exit 1
    fi
  else
    echo "⚠️  Spec reviewer could not complete — ESCALATE TO USER"
    exit 1
  fi
done
```

---

### Step 4: Dispatch Quality Reviewer (max 3 rounds)

Same pattern as Step 3 but `--role quality-reviewer`.

If Verdict = `NEEDS_FIXES` with 🔴 Critical issues: re-dispatch implementer with fix list before proceeding.

If Verdict = `NEEDS_FIXES` with only 🟡 Suggestions: proceed with `STATUS: DONE_WITH_CONCERNS`.

---

### Step 5: Mark Step Complete

After both reviews pass:

```
complete_task_step(step_name="[T{N}] ...", active_files=[...])
clear_drift()
```

---

## Model Reference

| Role               | Model                    | Timeout |
| ------------------ | ------------------------ | ------- |
| `implementer`      | `gemini-3.1-pro-preview` | 600s    |
| `spec-reviewer`    | `gemini-3-flash-preview` | 300s    |
| `quality-reviewer` | `gemini-3-flash-preview` | 300s    |

Override with `--model` flag if a task needs different reasoning power.

---

## Quick Reference — Common Patterns

### Pattern A: Simple mechanical task

```bash
# 1 attempt expected
.agents/subagents/spawn-subagent.sh \
  --role implementer \
  --task-file /tmp/task-T2.md \
  --context-files "src/handler.go,src/handler_test.go" \
  --workspace "$PWD"
```

### Pattern B: Task requiring codebase context

```bash
# Pass more context files to avoid NEEDS_CONTEXT
.agents/subagents/spawn-subagent.sh \
  --role implementer \
  --task-file /tmp/task-T3.md \
  --context-files "src/handler.go,src/middleware.go,src/config.go,src/handler_test.go" \
  --workspace "$PWD"
```

### Pattern C: Quick spec check only (skip quality review)

```bash
# When you only need to verify spec compliance
.agents/subagents/spawn-subagent.sh \
  --role spec-reviewer \
  --task-file /tmp/task-T1.md \
  --context-files "src/handler.go" \
  --workspace "$PWD" \
  --output-file /tmp/spec-review.md
```

---

## Escalation Rules

| Situation                          | Action                                          |
| ---------------------------------- | ----------------------------------------------- |
| Implementer BLOCKED 3×             | Stop. Report to USER with `/tmp/impl-T{N}.md`   |
| Spec NEEDS_FIXES 3×                | Stop. Report to USER with `/tmp/spec-T{N}.md`   |
| Quality 🔴 Critical can't be fixed | Stop. Report to USER                            |
| TIMEOUT                            | Consider splitting the task; report to USER     |
| Process error (exit 3)             | Check `gemini` CLI is available; report to USER |

---

## 🔴 Critical Constraints

1. **Never skip reviews**: Both spec-reviewer and quality-reviewer must pass before `complete_task_step`.
2. **No context sharing**: Each subagent gets only the files listed in `--context-files`. Do NOT pass session history.
3. **Read output files**: Always read `/tmp/impl-T{N}.md` and `/tmp/spec-T{N}.md` to understand details before retrying.
4. **Max 3 per stage**: Hard limit of 3 attempts per stage. Escalate after 3 failures.
5. **Role Anchoring**: Orchestrator ALWAYS prefixes responses with its role tag.
