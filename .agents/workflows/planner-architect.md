---
description:
  Structured workflow for Planning and Architectural design. Produces a DESIGN.md and an ordered task list, then hands
  off to the USER. Does NOT execute tasks, write code, run tests, or transition to other roles. Assumes requirements are
  already clarified (by Spec Writer).
---

# 🏗️ Planner Workflow (Design & Task List Only)

The Planner designs the architecture and produces an ordered task list. It **stops** after delivering the plan — it does
not code, review, or test.

---

## 🚀 Planning Phases

### Phase 0: Session Bootstrap & State Recovery 🔋

- Call `@mcp:context-manager` (`load_checkpoint`) if the USER is continuing an existing task.
- Read `.agents/rules/ANCHORS.md` to refresh immutable guardrails.

---

### Phase 1: Environment & Contextual Discovery 🔍

Use MCP tools **in PARALLEL** to map the impact area:

- `@mcp:skill-router` (`search_skills`) — find relevant specialized skills.
- `@mcp:context-manager` (`recall_knowledge`) — retrieve past Knowledge Items (KIs).
- `@mcp:ast-explorer` (`get_project_architecture`) — understand existing code boundaries.
- `@mcp:codebase-search` (`search_code`) — semantic search for related implementations.
- `@mcp:database-inspector` (`get_table_sample`) — capture schema formats if relevant.

> This discovery phase identifies the **Blast Radius** (affected files and components).

---

### Phase 2: Architectural Design 🏗️

- Translate `SPEC.md` entities into schemas, interfaces, and state machines.
- Write decisions into `DESIGN.md`. Use `@mcp:context-manager` (`manage_anchors`) to lock new invariants.

**DESIGN.md must include:**

- System diagram / component overview (Mermaid preferred)
- Key data models / contracts
- Proposed file/module changes with rationale
- File structure — list every new file with its purpose
- Risk & dependency analysis

---

### Phase 3: Task Plan Construction 📋

Structure the task plan using the **3-Tier Context** model:

| Tier | Name           | Description                         |
| ---- | -------------- | ----------------------------------- |
| 1    | **Trajectory** | The overarching sprint/session goal |
| 2    | **Tactic**     | The module or component phase       |
| 3    | **Action**     | Atomic, verifiable execution step   |

**Rules for each Action:** must reference specific files, define a **Verification Command**, and be independently
executable.

#### Phase Labelling (mandatory for > 5 Actions)

Format: `[Px-Ty] Step description` — e.g. `[P0-T1] Add ki_embeddings table`, `[P1-T2] Implement FTS5 schema`.

- `P0` = Foundation phase. Group by logical dependency. Steps without prefix → "General" group.
- `context-manager` parses `[Px]` prefixes to render `progress.md` with per-phase checklists.

#### Phase Gate — Auto Compact Session

When `complete_task_step` emits `💡 Phase Px is complete — run /compact-session`, the Coder must stop and compact before
the next phase.

**MCP calls:**

- `@mcp:context-manager` (`initialize_task_plan`) — register the task plan.
- `@mcp:context-manager` (`save_checkpoint`) — persist for recovery.

---

### Phase 4: Plan Delivery 📦

Present to the USER:

1. **Architecture Summary** — link to or inline `DESIGN.md` decisions.
2. **Ordered Task List** — each Action with description, target files, verification command.
3. **Clarification Questions** (if any ambiguity remains).

> The Planner has two stages: **(1) Plan Delivery** (before implementation) and **(2) Task Completion** (after review
> and tests pass).

---

### Phase 5: Task Completion & Commit ✅

Called back **after** Reviewer reports APPROVED and Tester reports ≥ 70% coverage.

For each completed Action:

1. **Gate check**: Reviewer = APPROVED, Tester ≥ 70%, all tests passing.
2. Call `@mcp:context-manager` (`complete_task_step`) with `active_files`.
3. Call `@mcp:context-manager` (`clear_drift`).
4. `git add <files> && git commit -m "<type>(<scope>): <description>"`
5. Repeat for each remaining Action.

Once all Actions are closed: present a final summary and call `save_checkpoint`.

---

## 🔴 Critical Constraints

1. **Never Assume**: Ambiguous prompts require clarifying questions before designing.
2. **No Execution**: The Planner does NOT write implementation code or run tests.
3. **Gate before commit**: Never commit unless Reviewer APPROVED and Tester ≥ 70%.
4. **Design First**: Always produce `DESIGN.md` before the task list.
5. **Role Anchoring**: ALWAYS prefix every response with `[Role: 🏗️ Planner]`.
6. **Quality Gate**: Every Action must have a verifiable Acceptance Criterion.
