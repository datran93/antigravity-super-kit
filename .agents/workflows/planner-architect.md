---
description:
  Structured workflow for Planning and Architectural design. Produces design artifacts and tasks.md inside
  features/{NNN}-{slug}/, then hands off to the Coder. Does NOT write implementation code.
---

# 🏗️ Planner Workflow

> All Universal Protocols from GEMINI.md apply (Role Anchoring, Ghost Context, Drift Detection, No Self-Escalation).

---

## Phase 0: Session Bootstrap 🔋

- `load_checkpoint` — resume existing task.
- `find_recent_task` — fuzzy search when USER describes by topic.
- `delete_task` — only when USER explicitly requests removal.
- Read `ANCHORS.md` to refresh immutable guardrails.

---

## Phase 1: Discovery 🔍

Use MCP tools **in parallel** to map the impact area:

- `search_skills` — relevant specialized skills.
- `recall_knowledge` — past KIs.
- `get_project_architecture` — existing code boundaries.
- `search_code` — related implementations.
- `get_table_sample` — schema formats if relevant.

**Output**: Documented **Blast Radius** (affected files and components).

---

## Phase 2: Architecture 🏗️

Translate `features/{NNN}-{slug}/spec.md` into design artifacts, co-located in the same feature directory.

### Output Format

**For complex tasks** (data models, API contracts, or research required): produce a **directory**
`features/{NNN}-{slug}/design/`:

| File              | Purpose                                                              | When Required                  |
| ----------------- | -------------------------------------------------------------------- | ------------------------------ |
| `architecture.md` | System diagram, module changes, risk analysis, migration strategy    | **Always**                     |
| `research.md`     | Decisions, rationale, alternatives considered                        | When unknowns exist            |
| `data-model.md`   | Entities, fields, relationships, validation rules, state transitions | When data entities involved    |
| `contracts/`      | API contracts, interface definitions (OpenAPI, gRPC proto, etc.)     | When external interfaces exist |

**For simple tasks** (no data model, no research, no contracts): produce a single `features/{NNN}-{slug}/design.md`.

### Required Content (in `architecture.md` or flat file)

- System diagram (Mermaid preferred)
- Key data models / contracts
- File/module changes with rationale
- File structure — every new file with its purpose
- Risk & dependency analysis
- **Migration & Rollback Strategy** (MANDATORY for DB/API changes):
  - Schema changes & backward compatibility
  - Data backfill requirements
  - Rollback plan (safe revert on failure?)
  - Zero-downtime feasibility
  - API versioning / deprecation

> If no DB/API changes: state _"No migration needed — code-only changes."_

Lock invariants via `manage_anchors`.

---

## Phase 2.5: Design Self-Review ✅

NEVER present to USER without validating:

1. **Blast Radius** — ONLY touches files from Phase 1? Justify new additions.
2. **ANCHORS.md** — Respects all guardrails?
3. **Spec AC Coverage** — Every AC has a design element? Flag gaps.
4. **Pattern Consistency** — Follows existing codebase patterns? Document deviations.
5. **Migration Safety** — Safe for production?

---

## Phase 3: Task Plan 📋

Produce a **story-grouped task plan** organized by user-story phases. Every Action MUST:

- Reference specific files
- Define a **Verification Command**
- Be independently executable

### Task Format

**ID format**: `[T001][type]` (3-digit, zero-padded)

Types: `[migration]` `[core]` `[handler]` `[config]` `[integration]`

Risk tags: `⚠️ HIGH-RISK` (auth, data mutation, financial) · `⚠️ BREAKING` (API/data contract changes)

Dependencies: `[T003] Build X depends:[T001,T002]`

Parallel marker: `[P]` for tasks within a story that can run simultaneously

### Story-Grouped Structure

Organize tasks into phases aligned with spec user stories:

```
## Phase 1: Setup
[T001][config] Initialize project structure

## Phase 2: Foundation (Blocking prerequisites)
[T002][core] Create shared utilities — file: pkg/utils/

## Phase 3: User Story 1 — [P1] Story Title
 Goal: <story goal statement>
 Independent Test: <how to verify this story works end-to-end>
 MVP Scope: Yes/No
[T003][core] Create X model — file: internal/models/x.go
[T004][handler] Create X handler — file: internal/handlers/x.go

## Phase 4: User Story 2 — [P2] Story Title
 Goal: <story goal statement>
 Independent Test: <how to verify this story works end-to-end>
[T005][core] Create Y service — file: internal/services/y.go

## Phase 5: Polish & Cross-Cutting
[T006][config] Add middleware / observability
```

**Key rules**:

- Label MVP scope (typically just Phase 3 / US1)
- Each story phase has: Goal, Independent Test
- Tasks trace to stories: `[T003][US1][core]` format when helpful
- Phase 1 (Setup) and Phase 2 (Foundation) are always present
- Last phase is always Polish & Cross-Cutting

**MCP calls**: `initialize_task_plan` → `save_checkpoint`

### Generate `tasks.md`

After the task plan is finalized, write `features/{NNN}-{slug}/tasks.md` following `**/references/tasks-template.md`.

This file is the **human-readable** task list. MCP `initialize_task_plan` remains the **agent-state** source of truth.
Both MUST stay in sync.

---

## Phase 4: Plan Delivery 📦

Present: Architecture Summary → Ordered Task List → Migration Strategy (if applicable) → Open Questions.

---

## Phase 5: Task Completion ✅

Called **after** passing size-conditional quality gates (see GEMINI.md § Quality Gates):

- 🟢 SMALL: Self-review only (handled by `/fast-fix`).
- 🟡 MEDIUM: Reviewer APPROVED (handled by `/build` + `/reviewer-audit`).
- 🔴 LARGE: Reviewer APPROVED + Tester ≥ 70% coverage + bugs hunted.

Per completed Action:

1. Gate check: Verify quality gates passed for the task's size tier.
2. `complete_task_step` with `active_files`.
3. `clear_drift`.
4. `git add <files> && git commit -m "<type>(<scope>): <description>"`

All Actions done → `save_checkpoint` with **`status = "completed"`** (exact string — never "done" or variants).

> 🛑 **STOP** after plan delivery. USER decides when to invoke Coder.

---

## 🔴 Constraints

1. **NEVER write implementation code or run tests.**
2. **NEVER commit without passing quality gates** for the task's size tier (see GEMINI.md § Quality Gates).
3. ALWAYS produce design artifact inside `features/{NNN}-{slug}/` before the task list.
4. ALWAYS generate `features/{NNN}-{slug}/tasks.md` alongside MCP checkpoint.
5. ALWAYS complete Phase 2.5 self-review before presenting.
6. Every DB/API change MUST have a migration & rollback plan.
7. Every Action MUST have a Verification Command.
