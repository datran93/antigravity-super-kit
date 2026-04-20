# 📋 Tasks Template

> This is the canonical template for `features/{NNN}-{slug}/tasks.md`. The Planner MUST follow this structure.
> This file is the **human-readable** task list. MCP `initialize_task_plan` remains the **agent-state** source of truth.

---

## Metadata

- **Feature**: `features/{NNN}-{slug}`
- **Spec**: `spec.md`
- **Design**: `design.md` | `design/architecture.md`
- **Created**: YYYY-MM-DD
- **Status**: planning | in-progress | completed

---

## Phase 1: Setup

- [ ] `[T001][config]` _Description_ — file: `{path}`
  - Verification: `{command}`

## Phase 2: Foundation (Blocking prerequisites)

- [ ] `[T002][core]` _Description_ — file: `{path}`
  - Verification: `{command}`
  - depends: `[T001]`

## Phase 3: User Story 1 — [P1] _Story Title_

> **Goal**: _story goal statement_
> **Independent Test**: _how to verify this story works end-to-end_
> **MVP Scope**: Yes

- [ ] `[T003][US1][core]` _Description_ — file: `{path}`
  - Verification: `{command}`
- [ ] `[T004][P][US1][handler]` _Description_ — file: `{path}`
  - Verification: `{command}`

## Phase 4: User Story 2 — [P2] _Story Title_

> **Goal**: _story goal statement_
> **Independent Test**: _how to verify this story works end-to-end_

- [ ] `[T005][US2][core]` _Description_ — file: `{path}`
  - Verification: `{command}`

## Phase N: Polish & Cross-Cutting

- [ ] `[T0xx][config]` _Description_ — file: `{path}`
  - Verification: `{command}`

---

## Legend

| Tag | Meaning |
|-----|---------|
| `[T001]` | Task ID (3-digit, zero-padded) |
| `[P]` | Can run in parallel with sibling tasks |
| `[US1]` | Traces to User Story 1 |
| `[core]` `[handler]` `[config]` `[migration]` `[integration]` | Task type |
| `⚠️ HIGH-RISK` | Auth, data mutation, financial |
| `⚠️ BREAKING` | API/data contract changes |
| `depends:[T001,T002]` | Dependency declaration |

---

## Completion Tracking

> Mark tasks `[x]` as they complete. The Planner updates this file after each `complete_task_step`.
