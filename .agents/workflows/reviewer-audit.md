---
description: Structured workflow for Code Review and Quality Audit. Reads the completed actions from MCP context and the original
---

# 🔍 Reviewer Workflow

---

## Phase 0: Load Context 📖

1. Read `features/YYYY-MM-DD-{slug}/spec.md` — Acceptance Criteria.
2. Read `features/YYYY-MM-DD-{slug}/design.md` — intended architecture.
3. Check MCP context for completed actions, files changed, and notes left by the Coder.
4. `load_checkpoint` — confirm completed Actions.
5. **Context Pruning**: Classify task domain (auth/db/api/refactor) then use `manage_anchors(action="list")` to load
   only the relevant `[domain:X]` anchors. Always include `[domain:quality]`. Do NOT read the full `ANCHORS.md` file.

> ❌ NEVER begin reviewing code before completing this phase.

---

## Phase 1: Mechanical Checks ⚙️

- **Lint/Build**: Confirm code compiles without errors.
- **Scope**: `get_project_architecture` — detect unapproved changes outside Bounded Context.
- **Breakage**: `search_code` — find callers of modified symbols, confirm no unintended breakage.

Document failures. NEVER fix them.

---

## Phase 1.5: Pattern Consistency 🔄

For each file created/modified:

1. **Find reference**: `search_code` for existing file of same type.
2. **Compare**: Error handling · Response format · Logger usage · DI pattern · Naming conventions.
3. **Flag deviations**: Unjustified pattern inconsistency = `MEDIUM` severity.

---

## Phase 2: Semantic Audit 🔍

### Design Conformity

- **Traceability**: Does each file implement what the design specified?
- **Contracts**: Data models, interfaces, signatures consistent with design?
- **Boundary**: No unapproved scope creep?
- **Creep**: Did the Coder ADD anything NOT in the design?

### Spec AC Verification

For each AC → verify corresponding implementation exists. Flag unaddressed ACs.

### Security Review

> NEVER reduce this to "any obvious issues". Read the actual code.

Per `**/references/security-checklist.md` — verify every check with specific file:line references.

### Performance Review

> Read actual queries and loops. NEVER flag patterns without analysis.

- DB queries: indexed? bounded (LIMIT)? Could return unbounded rows?
- Loops: N+1 queries inside? Bounded?
- Async: context cancellation, timeouts, error propagation?
- Memory: unbounded allocations?

### Testability & Clean Code

- DI, no hardcoded globals, naming clarity, function size, SOLID.

---

## Phase 3: Report 📋

Deliver report directly to the USER per `**/references/report-templates/reviewer-report.md`.

DO NOT write any markdown report files.

> 🛑 **STOP HERE.** The USER decides: ask Coder to fix, proceed to `/tester-verification`, or accept.

---

## 🔴 Constraints

1. **Report, NEVER fix**: Identify and document issues — NEVER modify code.
2. **Objective distance**: Evaluate against design — NOT personal preference.
3. **Severity honesty**: Label accurately. NEVER downplay blocking issues.
4. **Security depth**: ALWAYS use the security checklist. NEVER skip.
5. **Read the code**: Every finding MUST reference specific files and line numbers.
6. **Justify verdict**: APPROVED or NEEDS FIX MUST include 2-3 sentences of justification.
