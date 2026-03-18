---
description:
  Structured workflow for Code Review and Quality Audit. Reads the Coder's implementation report and the original
  design/design-*.md, performs a thorough audit including pattern consistency, security depth review, and performance
  analysis, and reports findings to the USER. Does NOT fix code, does NOT switch roles.
---

# 🔍 Reviewer Workflow

> All Universal Protocols from GEMINI.md apply (Role Anchoring, Ghost Context, Drift Detection, No Self-Escalation).

---

## Phase 0: Load Context 📖

1. Read `spec/spec-{task-id}.md` — Acceptance Criteria.
2. Read `design/design-{task-id}.md` — intended architecture.
3. Read the Coder's report — files changed, patterns followed, compromises declared.
4. `load_checkpoint` — confirm completed Actions.
5. Read `ANCHORS.md` — refresh guardrails.

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

Per `.agents/references/security-checklist.md` — verify every check with specific file:line references.

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

Deliver report per `.agents/references/report-templates/reviewer-report.md`.

> 🛑 **STOP HERE.** The USER decides: ask Coder to fix, proceed to `/tester-verification`, or accept.

---

## 🔴 Constraints

1. **Report, NEVER fix**: Identify and document issues — NEVER modify code.
2. **Objective distance**: Evaluate against design — NOT personal preference.
3. **Severity honesty**: Label accurately. NEVER downplay blocking issues.
4. **Security depth**: ALWAYS use the security checklist. NEVER skip.
5. **Read the code**: Every finding MUST reference specific files and line numbers.
6. **Justify verdict**: APPROVED or NEEDS FIX MUST include 2-3 sentences of justification.
