---
description: Structured workflow for Testing. Reads the completed actions from MCP context, performs deep code analysis to find
---

# 🧪 Tester Workflow

> Primary goal: **find bugs**. Coverage ≥ 70% is a byproduct, NOT the target.

---

## Phase 0: Read Context 📖

1. Read `features/YYYY-MM-DD-{slug}/design.md` — contracts, boundaries, expected behaviors.
2. Check MCP context for completed actions, files created/modified, and notes.
3. `load_checkpoint` — confirm completed Actions.
4. **Context Pruning**: Load only relevant domain anchors via `manage_anchors(action="list")`. Always include
   `[domain:quality]` and `[domain:security]` for testing.
5. `search_code` — find existing test helpers and fixtures.

> ❌ NEVER write tests before completing this phase.

---

## Phase 1: Deep Code Analysis (Bug Hunting) 🔍

> MOST CRITICAL PHASE. Read full implementation line by line. NEVER test from signatures alone.

For every file created/modified, build a **Bug Hypothesis List**:

### 🔴 Logic Bugs

Off-by-one · wrong comparison (`>` vs `>=`) · missing `return` after error · wrong variable (copy-paste) · nil/null
dereference without check.

### 🔴 Authorization & Security Bugs

Missing ownership checks (User A → User B's data?) · Tenant isolation violations · Role bypass · Input validation gaps ·
Sensitive data in responses. Per `**/references/security-checklist.md`.

### 🔴 State & Concurrency Bugs

Race conditions · Inconsistent state on partial failure · Missing transactions · Stale cache data.

### 🔴 Error Handling Bugs

Swallowed errors · Wrong HTTP status codes · Missing error cases (DB down? API timeout?) · Internals leaked in errors.

### 🔴 Data Integrity Bugs

Missing uniqueness constraints · Orphaned records on delete · Type coercion (string vs int, timezone, float precision).

**Output**: Written Bug Hypothesis List with file:line references.

---

## Phase 2: Test Case Design & Approval 📝

Based on the Bug Hypothesis List from Phase 1, deeply think through the necessary tests and create a structured **Test Case Table**.

**For each test case, include:**
- **ID**: (e.g., `TC01`)
- **Scenario**: What is being tested (e.g., "User A accesses User B's file").
- **Inputs/State**: Preconditions and input data.
- **Expected Outcome**: What the test should assert.
- **Priority**: (P0 to P3).

> 🛑 **STOP HERE.** Present the Test Case Table to the USER and explicitly ask for their approval or modifications before writing ANY test code. Do NOT proceed to Phase 3 until the USER approves.

---

## Phase 3: Bug-Hunting Tests 🎯

For each hypothesis:

1. Write a test that **proves the bug exists** (or proves the code is correct).
2. **Name tests for the bug**: ✅ `TestCreate_OtherUserAccess_ShouldReturn403` · ❌ `TestCreate_Success`
3. **Priority**:
   - **P0**: Auth bypass, data isolation, data corruption
   - **P1**: Silent wrong results
   - **P2**: Crashes or misleading error responses
   - **P3**: Edge cases (empty, boundary, unicode, large payloads)
4. **Structure**: Setup (multi-user/multi-tenant) → Action (as WRONG user/role) → Assert (correctly REJECTS).

> 🎯 Test passes + you expected failure → code is correct. Move on. 🐛 Test fails → real bug. Document in output. NEVER
> fix implementation.

---

## Phase 4: Coverage Completion 📊

**Only after Phase 3**. If below 70%:

- Write tests for uncovered paths — but think "what could go wrong", not "what lines to hit".
- Place tests adjacent to code (`foo_test.go`, `foo.test.ts`).

---

## Phase 5: Run & Measure ▶️

Run tests with coverage. Gate: **≥ 70%** or write more tests.

- Test fails → **real bug** (document, don't fix) or **wrong expectation** (fix the test).
- 3 consecutive failures → `record_failure` → stop → ask USER.

---

## Phase 6: Report 📋

Deliver report directly to the USER per `**/references/report-templates/tester-report.md`.

DO NOT write any markdown report files.

> 🛑 **STOP HERE.** The USER decides the next step.

---

## 🔴 Constraints

1. **Bug hunter first, coverage second**: Phases 1-3 MUST complete before Phase 4.
2. **Read code before writing tests**: NEVER test from function signatures alone.
3. **Tests only**: NEVER modify implementation code.
4. **≥ 70% coverage is mandatory**: But high coverage + no bugs found = you didn't look hard enough.
5. **NEVER mark tests passing without running them and checking output.**
