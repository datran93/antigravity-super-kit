---
description:
  Standard workflow for 🟡 MEDIUM tasks (new utilities, refactors, internal features). Plans inline, implements with
  intent locking and pattern discovery, self-reviews, then updates state. Does NOT commit.
---

# 🔨 Build Workflow (🟡 MEDIUM)

---

## When to Use

This workflow is for tasks where **ANY** of the following are true:

- 50-300 LOC estimated change
- New files within an existing module
- Internal refactor (no public API change)
- No DB migration needed

**Examples**: new utility, add helper function, internal feature, component refactor, new workflow.

> If the task involves DB migration, public API changes, auth/payment, or > 300 LOC → use `/planner-architect` (🔴
> LARGE) instead. If the task is < 50 LOC and modifies only existing files → use `/fast-fix` (🟢 SMALL) instead.

---

## Phase 0: Context 📖

1. Read the USER's request. Identify what needs to be built.
2. `load_checkpoint` if resuming an existing task.
3. Read relevant existing files to understand patterns.

> ❌ NEVER start building before understanding the codebase context.

---

## Phase 1: Inline Plan 📋

Produce a brief inline plan (NOT a separate spec.md or design.md):

```
## Plan
**Goal**: what we're building and why
**Files**: list of files to create/modify
**Approach**: brief technical approach (2-3 sentences)
**Verification**: how to verify it works
```

Present to USER for confirmation before proceeding.

---

## Phase 2: AST Pattern & Dependency Discovery 🔍

> You are extending an existing codebase. Your code MUST look like it belongs and avoid unintended side-effects.

1. Use `context` or `search_symbol` (via `codebase-explorer`) for a 360° AST view of target files and dependencies.
2. Use `find_usages` to understand the blast radius if modifying shared utilities.
3. Study: error handling, response format, naming conventions.
4. State which pattern you're following before writing code.

---

## Phase 3: Execute 🛠️

For each file group:

1. `declare_intent` to lock target files.
2. Read every file before modifying.
3. Implement with pattern conformity.
4. Clean Code: clear naming, small focused functions, SOLID.

---

## Phase 4: Self-Review 🔍

> NEVER submit without self-reviewing. Re-read ALL code you wrote/modified.

**Logic**: No copy-paste bugs · All error paths handled · No hardcoded values.

**Pattern Conformity**: Error handling matches · Naming matches · File placement matches.

**Security**: Quick check per `**/references/security-checklist.md`.

**Completeness**: All items from the inline plan addressed.

> Fix issues BEFORE proceeding.

---

## Phase 5: Verify ✅

- Run relevant tests — confirm no regressions.
- On **fail**: Fix and re-run. After 3 consecutive failures → `record_failure` → STOP → ask USER.
- On **pass**: Proceed to state update.

---

## Phase 6: State Update 📋

Inject gotchas via `annotate_file`. Update the MCP `context-manager` database with your progress and verified test
results.

> 🛑 **STOP HERE.** The USER decides: proceed to `/reviewer-audit`, commit directly, or continue (or proceed
> automatically if in `/auto-pilot` mode).

---

## 🔴 Constraints

1. **Read before write**: NEVER modify a file without reading it first.
2. **Inline plan only**: Do NOT create spec.md, design.md, tasks.md, or feature directories.
3. **Pattern conformity**: Find and follow existing patterns. Document deviations.
4. **Self-review before update**: Complete Phase 4 BEFORE Phase 6.
5. **Intent locking required**: ALWAYS use `declare_intent` before editing.
6. **No commit**: NEVER commit — the USER or Planner decides.
7. **Escape hatch**: If scope grows beyond MEDIUM → STOP → recommend `/planner-architect`.
