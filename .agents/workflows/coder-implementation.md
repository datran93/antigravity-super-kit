---
description:
  Structured workflow for Code Implementation. Reads design/design-*.md and the Planner's task list, executes each task
  with pattern conformity, performs self-review before reporting. Does NOT commit.
---

# 💻 Coder Workflow

> All Universal Protocols from GEMINI.md apply (Role Anchoring, Ghost Context, Drift Detection, No Self-Escalation).

---

## Phase 0: Read Design & Task List 📖

1. Read `design/design-{task-id}.md` — architecture, data models, constraints, migration strategy.
2. Read `spec/spec-{task-id}.md` — Acceptance Criteria tests will validate against.
3. `load_checkpoint` — load the task list.
4. Confirm scope — identify files to create/modify. Note `⚠️ HIGH-RISK` actions.

> ❌ NEVER start writing code before completing this phase.

---

## Phase 1: Intent Lock 📥

For each Action (one at a time, in order):

- Read Action description, type, risk level, target files, Verification Command.
- `declare_intent` to lock files.
- `check_intent_lock` before any edit. On Scope Creep ALARM → **stop and ask the USER**.

---

## Phase 2: Pattern Discovery 🔍

> You are extending an existing codebase. Your code MUST look like it belongs.

1. **Find reference**: `search_code` for ≥ 1 existing file of the same type (handler, service, repository).
2. **Study**: Error handling, response format, logging, DI pattern, naming conventions.
3. **Document**: State which pattern you're following before writing code.
4. **Deviations**: NEVER deviate without documenting WHY.

Additional: `search_skills` for best practices · `query-docs` for latest API specs · cross-reference design doc.

---

## Phase 3: Execution 🛠️

- **NEVER write blindly**: Read every file before modifying.
- Clean Code: clear naming, small focused functions, SOLID.
- Testable: Dependency Injection, no hardcoded globals.
- Stay strictly within the locked Bounded Context.
- Pattern Conformity: If a reviewer cannot distinguish your code from existing code — success.

---

## Phase 4: Self-Review 🔍

> NEVER submit without self-reviewing. Re-read ALL code you wrote/modified.

**Logic**: No copy-paste bugs · All error paths handled · No hardcoded values · Nil/null checks.

**Pattern Conformity**: Error handling matches · Response format matches · Naming matches · Logger matches · File
placement matches.

**Security**: Per `.agents/references/security-checklist.md`.

**Completeness**: All files from design implemented · All ACs addressed.

> Fix issues BEFORE proceeding.

---

## Phase 5: Verification ✅

- Run the **Verification Command** for the current Action.
- Run the **existing test suite** — confirm no regressions.
- On **fail**: Fix and re-run. After 3 consecutive failures → `record_failure` → stop → ask USER.
- On **pass**: Note result. NEVER commit — committing is the Planner's job.

Repeat **Phase 1 → 5** for each remaining Action.

---

## Phase 6: Report 📋

Inject gotchas via `annotate_file`, then deliver report per `.agents/references/report-templates/coder-report.md`.

> 🛑 **STOP HERE.** The USER decides the next step.

---

## 🔴 Constraints

1. **Read before write**: NEVER modify a file without reading it first.
2. **Pattern conformity**: Find and follow existing patterns. Document deviations.
3. **Self-review before report**: Complete Phase 4 BEFORE Phase 6.
4. **Task order**: Complete Actions in the Planner's order. NEVER skip or reorder.
5. **Stay in scope**: NEVER refactor outside the current Action's Bounded Context.
