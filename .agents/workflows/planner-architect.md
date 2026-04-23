---
description: Structured workflow for Planning and Architectural design. Produces design artifacts inside features/YYYY-MM-DD-{slug}/, then
---

# 🏗️ Planner Workflow

---

## Phase 0: Session Bootstrap 🔋

- `load_checkpoint` — resume existing task.
- `find_recent_task` — fuzzy search when USER describes by topic.
- `delete_task` — only when USER explicitly requests removal.
- **Context Pruning**: Use `manage_anchors` (action: "list") or `recall_knowledge` to dynamically fetch only the
  domain-specific constraints relevant to the task (e.g., Auth, DB, UI) instead of loading the entire `ANCHORS.md` file.

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

Translate `features/YYYY-MM-DD-{slug}/spec.md` into design artifacts, co-located in the same feature directory.
**State Machine Enforcement**: You MUST follow this process sequence: 
`Explore -> Ask Clarifying Questions -> Propose 2-3 Approaches -> Present Section-by-Section -> Adversarial Review -> Task Plan`.

### 1. Section-by-Section Presentation
Do NOT dump a massive `design.md` file at once. Present foundational decisions (e.g., Data Models, API Contracts) to the USER first. Only proceed to downstream systems (UI, Integrations) after the foundation is locked. 
For the core problem, ALWAYS **propose 2-3 approaches** with trade-offs before locking the design. Prefer concrete design details (structs, interfaces, data flow) over hand-wavey prose.

### Output Format

**For complex tasks** (data models, API contracts, or research required): produce a **directory**
`features/YYYY-MM-DD-{slug}/design/`:

| File              | Purpose                                                              | When Required                  |
| ----------------- | -------------------------------------------------------------------- | ------------------------------ |
| `design.md` | System diagram, module changes, risk analysis, migration strategy    | **Always**                     |
| `research.md`     | Decisions, rationale, alternatives considered                        | When unknowns exist            |
| `data-model.md`   | Entities, fields, relationships, validation rules, state transitions | When data entities involved    |
| `contracts/`      | API contracts, interface definitions (OpenAPI, gRPC proto, etc.)     | When external interfaces exist |

**For simple tasks** (no data model, no research, no contracts): produce a single `features/YYYY-MM-DD-{slug}/design.md`.

### Required Content (in `design.md`)

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

## Phase 2.5: Adversarial Design Review 🦹‍♂️

NEVER present to USER without performing a hostile self-review. Actively attack your own design:

1. **Vagueness Attack**: Attack hand-wavey sections, hidden assumptions, and missing concrete details (signatures, algorithms).
2. **Blast Radius** — ONLY touches files from Phase 1? Justify new additions.
3. **ANCHORS.md Doctrine** — Attack weak compliance with `ANCHORS.md` guardrails. Are there security or migration gaps?
4. **Spec AC Coverage** — Every AC has a design element? Flag gaps.
5. **Verification Weakness** — Are the verification strategies actually provable?
6. **External Challenge Prompt**: Offer the USER an explicit "External Challenge Prompt" (a list of hard questions they should consider) before proceeding to Phase 3.

---

## Phase 3: Task Plan 📋

Produce a **story-grouped task plan** organized by user-story phases.
**Do NOT create a `tasks.md` file.** The task plan is exclusively managed via the MCP `initialize_task_plan` tool, which automatically tracks state and generates a `progress.md` dashboard.

Every Action MUST:

- Reference specific files
- Define a **Verification Command**
- Be independently executable

### Task Format

**ID format**: `[ST01][type]` (2-digit, zero-padded)

Types: `[migration]` `[core]` `[handler]` `[config]` `[integration]`

Risk tags: `⚠️ HIGH-RISK` (auth, data mutation, financial) · `⚠️ BREAKING` (API/data contract changes)

Dependencies: `[ST03] Build X depends:[ST01,ST02]`

Parallel marker: `[P]` for tasks within a story that can run simultaneously

### Story-Grouped Structure

Organize tasks into phases aligned with spec user stories:

```
## Phase 1: Setup
[ST01][config] Initialize project structure

## Phase 2: Foundation (Blocking prerequisites)
[ST02][core] Create shared utilities — file: pkg/utils/

## Phase 3: User Story 1 — [P1] Story Title
 Goal: <story goal statement>
 Entry Criteria: <what must be true to start this phase>
 Exit Criteria: <what must be true to consider this phase complete>
 Independent Test: <how to verify this story works end-to-end>
 MVP Scope: Yes/No
[ST03][core] Create X model — file: internal/models/x.go
[ST04][handler] Create X handler — file: internal/handlers/x.go

## Phase 4: User Story 2 — [P2] Story Title
 Goal: <story goal statement>
 Entry Criteria: <what must be true to start this phase>
 Exit Criteria: <what must be true to consider this phase complete>
 Independent Test: <how to verify this story works end-to-end>
[ST05][core] Create Y service — file: internal/services/y.go

## Phase 5: Polish & Cross-Cutting
[ST06][config] Add middleware / observability
[ST07][config] Update Knowledge Items / Anchors (Memory Sync)
```

**Key rules**:

- Label MVP scope (typically just Phase 3 / US1)
- Each story phase has: Goal, Entry Criteria, Exit Criteria, Independent Test
- Tasks trace to stories: `[ST03][US1][core]` format when helpful
- Phase 1 (Setup) and Phase 2 (Foundation) are always present
- Last phase is always Polish & Cross-Cutting. It MUST include a task to update Knowledge Items / Anchors if architecture or patterns changed.

**MCP calls**: `initialize_task_plan` → `save_checkpoint`

---

## Phase 4: Plan Delivery 📦

Present: Architecture Summary → Ordered Task List → Migration Strategy (if applicable) → Open Questions.

---

> 🛑 **STOP** after plan delivery. USER decides when to invoke Coder.

---

## 🔴 Constraints

1. **NEVER write implementation code or run tests.**
2. **NEVER commit code.** The USER decides when to commit.
3. ALWAYS produce design artifact inside `features/YYYY-MM-DD-{slug}/` before the task list.
4. ALWAYS complete Phase 2.5 Adversarial Review and offer an External Challenge Prompt before presenting the task plan.
5. Every DB/API change MUST have a migration & rollback plan.
6. Every Action MUST have a Verification Command.
7. **Strict State Machine**: You MUST follow: Explore -> Clarify -> Propose Options -> Section Design -> Adversarial Review -> Task Planning. Do not jump straight to dumping the design.
