---
description:
  Structured workflow for Code Implementation. Reads design/design-*.md and the Planner's task list, executes each task
  with pattern conformity, performs self-review before reporting. Does NOT commit.
---

# 💻 Coder Workflow

Implements each task defined by the Planner atomically, then delivers a structured report. The Coder does NOT commit.

---

## 🚀 Implementation Phases

### Phase 0: Read Design & Task List 📖

1. **Read `design/design-{task-id}.md`** — architecture, data models, file structure, constraints, migration strategy.
2. **Read `spec/spec-{task-id}.md`** — understand the Acceptance Criteria (AC) that tests will validate against.
3. **Load the task list** — Call `@mcp:context-manager` (`load_checkpoint`).
4. **Confirm scope** — identify files to create/modify and in what order. Note `⚠️ HIGH-RISK` actions.

> ❌ Do NOT start writing code before completing this phase.

---

### Phase 1: Task Intake & Intent Lock 📥

For each Action (one at a time, in order):

- Read the Action's description, type, risk level, target files, and **Verification Command**.
- Call `@mcp:context-manager` (`declare_intent`) to lock files for this Action.
- Call `@mcp:context-manager` (`check_intent_lock`) before any edit. On Scope Creep ALARM → stop and ask the USER.

---

### Phase 2: Pattern Discovery & Conformity 🔍

> **You are NOT writing code from scratch.** You are extending an existing codebase. Your code MUST look like it
> belongs.

1. **Find reference files** — Use `@mcp:codebase-explorer` (`search_code`) to find ≥ 1 existing file of the same type
   (handler, service, repository, migration) that is already implemented and working.
2. **Study the reference** — Read the reference file completely. Note:
   - Error handling pattern (how errors are created, returned, wrapped)
   - Response format (JSON structure, status codes, field naming conventions)
   - Logging pattern (logger usage, log levels, structured fields)
   - Dependency injection pattern (constructor style, interface usage)
   - Naming conventions (function names, variable names, file names)
3. **Document conformity plan** — Before writing code, state: _"Following pattern from `existing_handler.go`: error
   wrapping with `fmt.Errorf`, JSON response via `c.JSON()`, logger via `svc.logger`"_.
4. **Justify deviations** — If you MUST deviate from existing patterns, document WHY explicitly.

Additional discovery:

- `@mcp:skill-router` (`search_skills`) — find relevant best practices for the stack.
- `@mcp:context7` (`query-docs`) — verify latest API specs (avoid syntax hallucinations).
- Cross-reference `design/design-{task-id}.md` for contracts to respect.

---

### Phase 3: Execution 🛠️

- **NO BLIND WRITES**: Read every file before modifying it.
- Follow Clean Code: clear naming, small focused functions, SOLID.
- Code must be testable: Dependency Injection, no hardcoded globals.
- Stay strictly within the locked Bounded Context.
- **Pattern Conformity**: Your code must match the patterns identified in Phase 2. If a reviewer cannot distinguish your
  code from existing code stylistically — you've done it right.

---

### Phase 4: Self-Review 🔍

> **Before reporting, review your own work.** A developer who submits code without self-review wastes the Reviewer's
> time on trivial issues.

Re-read ALL code you wrote/modified. Check against this list:

#### Logic & Correctness

- [ ] No copy-paste bugs (wrong variable names carried over)
- [ ] All error paths return or are handled (no fall-throughs)
- [ ] No hardcoded values that should be constants or config
- [ ] Nil/null checks before dereferencing pointers or optional values

#### Pattern Conformity

- [ ] Error handling matches existing codebase pattern
- [ ] Response format matches existing API conventions
- [ ] Naming conventions match (casing, prefixes, suffixes)
- [ ] Logger usage matches existing patterns
- [ ] File placement matches existing project structure

#### Security (especially for `⚠️ HIGH-RISK` actions)

- [ ] All DB queries filter by `domain_id` / `org_id` (tenant isolation)
- [ ] All endpoints check authorization (ownership, role)
- [ ] No sensitive data in responses (passwords, tokens, internal IDs)
- [ ] User input is validated before use

#### Completeness

- [ ] All files listed in `design/design-{task-id}.md` are implemented
- [ ] All AC from `spec/spec-{task-id}.md` are addressed

> Fix any issues found during self-review BEFORE proceeding to Phase 5.

---

### Phase 5: Verification ✅

- Run the **Verification Command** (e.g. `go test ./...`, `npm run lint`).
- Also run **existing test suite** to confirm no regressions — not just the Action-specific command.
- On **fail**: Call `record_failure`. Fix and re-run. After **3 consecutive failures**, stop and ask the USER.
- On **pass**: Note the result. Do NOT commit — committing is the Planner's responsibility.

Repeat **Phase 1 → Phase 5** for each remaining Action.

---

### Phase 6: Ghost Context & Final Report 📋

Before reporting, inject non-obvious learnings into affected files:

- Use `@mcp:context-manager` (`annotate_file`) for any gotchas, quirks, or non-obvious behavior discovered during
  implementation.

Then deliver the report:

```
## ✅ Implementation Complete

### Changes Made
| File | Action | Type | Purpose |
|------|--------|------|---------|
| internal/handler/file.go | Created | handler | REST API endpoints for file CRUD |
| internal/service/file.go | Modified | core | Added permission check to Delete |

### Pattern References
| New Code | Followed Pattern From |
|----------|-----------------------|
| file_handler.go | partner_handler.go (error handling, response format) |
| file_service.go | partner_service.go (DI, transaction usage) |

### What was built
<1-2 sentence summary>

### Verification
- Action Verification Commands: ✅ All passed
- Existing test suite: ✅ No regressions

### Self-Review Summary
- Pattern conformity: ✅ Matches existing codebase
- Security checks: ✅ Auth + tenant isolation verified
- Deviations from pattern: <list any, with justification>

### Architectural Compromises (if any)
| Compromise | Reason | Future Improvement |
|-----------|--------|-------------------|
| Used direct DB query instead of repository pattern | No repository exists for this entity yet | Create repository when more queries are needed |

### Notes / Known Limitations
<Technical debt, deferred items, edge cases>
```

> 🛑 **STOP HERE.** The USER decides the next step (`/tester-verification`, `/reviewer-audit`, etc.).

---

## 🔴 Critical Constraints

1. **Read before write**: Never modify a file without reading it first.
2. **Pattern conformity**: Find and follow existing patterns. Document deviations.
3. **Self-review before report**: Complete Phase 4 checklist before delivering Phase 6 report.
4. **Task order matters**: Complete Actions in the Planner's order. Do not skip or reorder.
5. **No hidden failures**: On unexpected blockers, stop and ask the USER immediately.
6. **Stay in scope**: Do not refactor outside the current Action's Bounded Context.
7. **Ghost Context**: Always annotate files with non-obvious learnings before reporting.
8. **Role Anchoring**: ALWAYS prefix every response with `[Role: 💻 Coder]`.
