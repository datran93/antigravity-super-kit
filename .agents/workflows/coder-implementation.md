---
description:
  Structured workflow for Code Implementation. Reads features/YYYY-MM-DD-{slug}/ artifacts (design, spec) and the
  Planner's task
---

# 💻 Coder Workflow

> ⚠️ **Quality Warning**: The Reviewer will review your code and the Tester will test your code, so think carefully and
> write code carefully.

> **Implementation**: This workflow executes tasks planned by `/planner-architect`.

---

## Phase 0: Read Design & Task List 📖

1. `load_checkpoint` — load the MCP task execution context.
2. The context will provide the specific Design, Acceptance Criteria (ACs), and **Auto-Linked Contexts** (automatically
   resolved by `load_checkpoint` when encountering `@task-[ID]`, `@ki/[Name]`, `@anchor/[Key]`, or `@doc/[path]` tags).
3. `retrieve_context` — assemble a unified context pack for the current task domain.
4. Confirm scope — identify files to create/modify. Note `⚠️ HIGH-RISK` actions.

> ❌ NEVER start writing code before completing this phase.

---

## Phase 1: Intent Lock 📥

For each Action (one at a time, in order):

- Read Action description, type, risk level, target files, Verification Command.
- `declare_intent` to lock files.
- `check_intent_lock` before any edit. On Scope Creep ALARM → **stop and ask the USER**.

---

## Phase 2: AST Pattern & Dependency Discovery 🔍

> You are extending an existing codebase. Your code MUST look like it belongs and you MUST know its blast radius.

1. **360° Context**: Use `context` or `search_symbol` (via `codebase-explorer`) to get an AST-based view of the
   classes/functions you are modifying. This reveals definitions and related chunks without keyword-guessing. _Note:
   Prioritize these AST tools over `view_file` to drastically save input tokens._
2. **Blast Radius**: Use `find_usages` to map out dependencies before modifying shared code.
3. **Study**: Error handling, response format, logging, DI pattern, naming conventions.
4. **Document**: State which pattern you're following before writing code.
5. **Deviations**: NEVER deviate without documenting WHY.

Additional: `query-docs` for latest API specs · `search_docs` for `@doc/` references · cross-reference design doc.

6. **Template Check**: Before creating new files (handlers, tests, repositories), check `.agents/templates/` for
   matching boilerplate templates. Use them as starting points to ensure pattern conformity.

---

## Phase 3: Execution 🛠️

- **NEVER write blindly**: Understand the file structure before modifying. Prioritize AST Context over reading the
  entire file with `view_file`.
- Clean Code: clear naming, small focused functions, SOLID.
- Testable: Dependency Injection, no hardcoded globals.
- Stay strictly within the locked Bounded Context.
- Pattern Conformity: If a reviewer cannot distinguish your code from existing code — success.

---

## Phase 4: Self-Review 🔍

> NEVER submit without self-reviewing. Re-read ALL code you wrote/modified.

**CRITICAL**: Do the self-review internally using `<thinking>` tags. Output NOTHING if the review passes. Only output
text if you find bugs to fix and explain the fix.

**Check**:

- **Logic**: No copy-paste bugs · All error paths handled · No hardcoded values · Nil/null checks.
- **Pattern Conformity**: Error handling matches · Response format matches · Naming matches · Logger matches · File
  placement matches.
- **Security**: Per `**/references/security-checklist.md`.
- **Completeness**: All files from design implemented · All ACs addressed.

> Fix issues BEFORE proceeding.

---

## Phase 5: Verification ✅

- Run the **Verification Command** for the current Action. **ALWAYS append quiet flags** (e.g., `-short`, `--quiet`,
  `| grep FAIL`) to reduce terminal noise and save tokens.
- Run the **existing test suite** — confirm no regressions. Use quiet flags to prevent flooding the context window.
- Run `agk validate` to check for broken `@doc/path` references, missing governance files, and template integrity.
- On **fail**: Fix and re-run. After 3 consecutive failures → `record_failure` → stop → ask USER.
- On **pass**: Note result. NEVER commit — the USER decides when to commit.

Repeat **Phase 1 → 5** for each remaining Action.

---

## Phase 6: State Update 📋

For **each completed Action**, do the following BEFORE moving to the next Action:

1. `complete_task_step` with `active_files`.
2. `log_activity` (event_type: "step_completed") to record the action in the audit trail.
3. `clear_drift`.

When **ALL Actions are completed**:

1. **AC Validation**: Cross-reference the task's `acceptance_criteria` against what was implemented. Flag any
   unaddressed ACs before marking complete.
2. Inject gotchas via `annotate_file`.
3. Persist ephemeral findings via `manage_session_memory` (action: "add") for the Reviewer/Tester.
4. `log_activity` (event_type: "task_completed") with a summary of all changes.
5. Update the MCP `context-manager` database with your progress and any notes for the Reviewer/Tester. Any `@task`,
   `@ki`, `@anchor`, or `@doc/` tags used in your notes will be validated against the global and project scope.
6. `save_checkpoint` with **`status = "completed"`** (exact string — never "done" or variants). This triggers the Broken
   Link Validator.

---

## Phase 7: Report 📋

Deliver report directly to the USER per `**/references/report-templates/coder-report.md`.

DO NOT write any markdown report files.

> 🛑 **STOP HERE.** The USER decides the next step.

---

## 🔴 Constraints

1. **Read before write**: NEVER modify a file without checking its context first. Use AST context over full `view_file`
   reads.
2. **Pattern conformity**: Find and follow existing patterns. Document deviations.
3. **Self-review before update**: Complete Phase 4 BEFORE Phase 6.
4. **Task order**: Complete Actions in the Planner's order. NEVER skip or reorder.
5. **Stay in scope**: NEVER refactor outside the current Action's Bounded Context.
