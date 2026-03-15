---
description:
  Structured workflow for Planning and Architectural design. Produces a DESIGN.md and an ordered task list, then hands
  off to the USER. Does NOT execute tasks, write code, run tests, or transition to other roles. Assumes requirements are
  already clarified (by Spec Writer).
---

# 🏗️ Planner Workflow (Design & Task List Only)

This workflow is responsible **exclusively** for understanding requirements, designing the architecture, and producing a
clear, ordered task list. The Planner **stops** after delivering the plan — it does not code, review, or test.

---

## 🚀 Planning Phases

### Phase 0: Session Bootstrap & State Recovery 🔋

- **Load State**: Call `@mcp:context-manager` (`load_checkpoint`) if the USER is continuing an existing task.
- **Load Anchors**: Read `.agents/rules/ANCHORS.md` to refresh the immutable guardrails before evaluating the USER's
  request.

---

### Phase 1: Environment & Contextual Discovery 🔍

Use MCP tools **in PARALLEL** to build a comprehensive map of the impact area:

- `@mcp:skill-router` (`search_skills`) — find relevant specialized skills.
- `@mcp:context-manager` (`recall_knowledge`) — retrieve past Knowledge Items (KIs).
- `@mcp:ast-explorer` (`get_project_architecture`) — understand existing code boundaries.
- `@mcp:codebase-search` (`search_code`) — semantic search over the codebase for related implementations (requires prior
  `index_codebase` run).
- `@mcp:database-inspector` (`get_table_sample`) — capture data schema formats if relevant.

> This discovery phase identifies the **Blast Radius** (affected files and components).

---

### Phase 2: Architectural Design 🏗️

- **Technical Translation**: Convert data entities and domain boundaries from `SPEC.md` into exact schemas, code
  interfaces, and state machines.
- Write architectural decisions into a `DESIGN.md` file at the project root (or agreed location).
- Use `@mcp:context-manager` (`manage_anchors`, `action="set"`) to lock in any new system invariant rules.

**DESIGN.md must include:**

- System diagram / component overview (Mermaid preferred)
- Key data models / contracts
- Proposed file/module changes with rationale
- **File structure** — if the feature introduces new files, list every new file with its purpose (tree format preferred)
- Risk & dependency analysis

---

### Phase 3: Task Plan Construction 📋

Structure the task plan using the **3-Tier Context** model:

| Tier | Name           | Description                         |
| ---- | -------------- | ----------------------------------- |
| 1    | **Trajectory** | The overarching sprint/session goal |
| 2    | **Tactic**     | The module or component phase       |
| 3    | **Action**     | Atomic, verifiable execution step   |

**Rules for each Action:**

- Must reference specific files or functions.
- Must define a **Verification Command** (e.g., `go test ./...`, `npm run lint`) as Acceptance Criteria.
- Must be executable independently in order.

#### 🗂️ Phase Division (mandatory for large projects)

When a sprint has **more than 5 Actions**, group them into **Phases** using the following convention:

```
[Px-Ty] Step description
│  │
│  └── Ty = Tactic index within the phase (T1, T2, …)
└──── Px = Phase index (P0, P1, P2, …)
```

**Examples:**

```
[P0-T1] Add ki_embeddings table to db.go
[P0-T2] Implement cosineSimKI + rrfScore
[P1-T1] Scaffold mcp-codebase-search-go module
[P1-T2] Implement store/db.go with FTS5 schema
[P2-T1] Write smart-workflow-router.md
```

**Phase rules:**

- P0 = Foundation / Infrastructure phase (no user-facing features yet)
- Group Actions by **logical dependency** — a phase should be independently releasable
- Steps without a phase prefix (e.g., bare `"Run tests"`) are rendered in a "General" group

> ⚠️ **Why this matters**: The `context-manager` parses `[Px]` prefixes to render `progress.md` with a Phase Overview
> table and per-phase checklists. Without prefixes, the view is a flat list.

#### 🔄 Phase Gate — Auto Compact Session

The `complete_task_step` tool **automatically emits a reminder** when the last step of a phase is marked done. When you
see:

```
💡 Phase Px is complete — run /compact-session to persist a KI before starting the next phase.
```

**The Coder must stop and run `/compact-session` before starting the next phase.** This is not optional — it flushes
context and prevents drift across phases.

**MCP calls:**

- `@mcp:context-manager` (`initialize_task_plan`) — register the task plan.
- `@mcp:context-manager` (`save_checkpoint`) — persist the plan for recovery.

---

### Phase 4: Plan Delivery 📦

Present the full plan to the USER in a clear, readable format:

1. **Architecture Summary** — Link to or inline the `DESIGN.md` decisions.
2. **Ordered Task List** — Each Action with:
   - Description
   - Target files
   - Verification command
3. **Clarification Questions** (if any ambiguity remains).

> The Planner's job now has two stages: **(1) Plan Delivery** (before implementation) and **(2) Task Completion** (after
> review and tests pass).

---

### Phase 5: Task Completion & Commit ✅

The Planner is called back **after** the Reviewer reports APPROVED and the Tester reports coverage ≥ 70%.

For each completed Action:

1. **Gate check** — Confirm both conditions are met before proceeding:
   - Reviewer verdict: **APPROVED** (no HIGH severity issues outstanding)
   - Tester coverage: **≥ 70%** and all tests passing
2. **Mark task done** — Call `@mcp:context-manager` (`complete_task_step`) with the `active_files`.
3. **Clear drift counter** — Call `@mcp:context-manager` (`clear_drift`).
4. **Commit** — Run:
   ```
   git add <changed files>
   git commit -m "<type>(<scope>): <concise description of what was done>"
   ```
5. **Repeat** for each remaining Action until all tasks in the plan are closed.

Once all Actions are closed:

- Present a final session summary to the USER.
- Call `@mcp:context-manager` (`save_checkpoint`) to persist the completed state.

---

## 🔴 Critical Constraints

1. **Never Assume**: If the prompt is ambiguous (e.g., "Add auth"), you MUST ask clarifying questions before designing.
2. **No Execution**: The Planner does NOT write implementation code or run tests.
3. **Gate before commit**: Never call `complete_task_step` or `git commit` unless both Reviewer APPROVED and Tester
   coverage ≥ 70% are confirmed.
4. **Design First**: Always produce `DESIGN.md` before the task list — the list must be grounded in the design.
5. **Role Anchoring**: ALWAYS prefix every conversational response with `[Role: 🏗️ Planner]`.
6. **Quality Gate**: Every Action in the task list must have a verifiable Acceptance Criterion before delivery.
