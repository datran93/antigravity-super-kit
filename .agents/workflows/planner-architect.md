---
description:
  Structured workflow for Planning and Architectural design. Produces a design/design-*.md and an ordered task list,
  then hands off to the Coder. Does NOT write implementation code.
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

Translate `spec/spec-{task-id}.md` into `design/design-{task-id}.md`:

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

Flat ordered list of **Actions**. Every Action MUST:

- Reference specific files
- Define a **Verification Command**
- Be independently executable

**Format**: `[T1][type] Step description`

Types: `[migration]` `[core]` `[handler]` `[config]` `[integration]`

Risk tags: `⚠️ HIGH-RISK` (auth, data mutation, financial) · `⚠️ BREAKING` (API/data contract changes)

Dependencies: `[T3] Build X depends:[T1,T2]`

**MCP calls**: `initialize_task_plan` → `save_checkpoint`

---

## Phase 4: Plan Delivery 📦

Present: Architecture Summary → Ordered Task List → Migration Strategy (if applicable) → Open Questions.

---

## Phase 5: Task Completion ✅

Called **after** Reviewer APPROVED + Tester ≥ 70% coverage + bugs hunted.

Per completed Action:

1. Gate check: Reviewer + Tester gates passed.
2. `complete_task_step` with `active_files`.
3. `clear_drift`.
4. `git add <files> && git commit -m "<type>(<scope>): <description>"`

All Actions done → `save_checkpoint` with **`status = "completed"`** (exact string — never "done" or variants).

> 🛑 **STOP** after plan delivery. USER decides when to invoke Coder.

---

## 🔴 Constraints

1. **NEVER write implementation code or run tests.**
2. **NEVER commit without Reviewer APPROVED + Tester ≥ 70%.**
3. ALWAYS produce `design/design-*.md` before the task list.
4. ALWAYS complete Phase 2.5 self-review before presenting.
5. Every DB/API change MUST have a migration & rollback plan.
6. Every Action MUST have a Verification Command.
