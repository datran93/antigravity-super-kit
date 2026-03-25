---
description:
  Read-only cross-artifact consistency and coverage analysis. Detects duplications, ambiguities, underspecification,
  coverage gaps, and terminology drift across spec, design, and task plan without modifying any files.
---

# 📊 Analyze Artifacts

> All Universal Protocols from GEMINI.md apply (Role Anchoring, Ghost Context, Drift Detection, No Self-Escalation).
>
> **Role**: Standalone utility — **READ-ONLY**. NEVER modifies any artifact files.

---

## Phase 0: Load Artifacts 📖

1. Identify the task: `load_checkpoint` or identify from USER context.
2. Read all artifacts from `features/{NNN}-{slug}/`:
   - `spec.md` — requirements, user stories, acceptance criteria
   - `design.md` OR `design/architecture.md` (detect format)
   - `tasks.md` — task plan from Planner
3. Read `ANCHORS.md` — constraints to validate against.
4. If any artifact is missing → report which artifact is missing → STOP.

> ❌ NEVER proceed without at least spec + design + task plan.

---

## Phase 1: Build Semantic Models 🧠

For each artifact, extract an inventory:

### Requirements Inventory (from spec)

For each user story and AC:
- Unique ID (AC-1, AC-2, ...)
- Description / precondition / action / outcome
- Priority (P1, P2, P3)
- Keywords for matching

### Design Element Inventory (from design)

For each module, component, data model, contract:
- Element name
- Purpose / responsibility
- Files involved
- Related requirements (if stated)

### Task Coverage Map (from task plan)

For each task action:
- Task ID (T001, T002, ...)
- Type (core, handler, config, etc.)
- Target files
- Related requirements (inferred from description)

---

## Phase 2: Detection Passes 🔬

Execute 6 detection passes in order:

### A. Duplication Detection

- Find near-duplicate requirements (same intent, different wording)
- Find overlapping task actions (same files, similar description)
- **Signal**: Two requirements that would produce the same test

### B. Ambiguity Detection

- Scan for vague adjectives: "fast", "secure", "efficient", "seamless", "proper", "appropriate"
- Find unresolved `[NEEDS CLARIFICATION]` markers
- Find placeholder text: `TBD`, `TODO`, `<placeholder>`, `...`
- **Signal**: Requirement that cannot produce a deterministic test

### C. Underspecification

- Requirements without measurable outcome in THEN clause
- User stories without acceptance criteria
- Design elements without corresponding spec requirement
- **Signal**: A developer would need to make assumptions to implement

### D. ANCHORS Alignment

- Read `.agents/rules/ANCHORS.md`
- Check: Does the design respect technology stack constraints?
- Check: Does the plan include verification commands (TDD anchor)?
- Check: Are all source files in English (language anchor)?
- **Signal**: Any violation of an ANCHORS constraint

### E. Coverage Gaps

- Requirements with NO corresponding design element
- Requirements with NO corresponding task
- Tasks with NO corresponding requirement (orphan tasks)
- Design elements with NO corresponding task (unimplemented design)
- **Signal**: Work that will be forgotten or unnecessary work

### F. Inconsistency

- Terminology drift: Same concept called different names across artifacts
- Entity mismatches: Attributes in spec differ from design data model
- Ordering contradictions: Task dependencies that conflict with spec priority
- **Signal**: Artifacts that disagree on what the system should do

---

## Phase 3: Severity Assignment ⚖️

For each finding, assign severity:

| Severity | Criteria | Action Required |
|----------|----------|----------------|
| **CRITICAL** | Violates ANCHORS · Missing core artifact · Zero-coverage blocking requirement | Must fix before implementation |
| **HIGH** | Conflicting requirement · Ambiguous security/performance attribute | Should fix before implementation |
| **MEDIUM** | Terminology drift · Missing NFR task coverage · Minor coverage gap | Can fix during implementation |
| **LOW** | Style/wording improvement · Redundant task | Defer or accept |

---

## Phase 4: Report Delivery 📋

Produce report per `.agents/references/report-templates/analyze-report.md` containing:

1. **Inventory Summary**: Count of items in each artifact
2. **Coverage Mapping Table**: Requirement → Design → Task traceability
3. **Findings Table**: All issues with severity, category, description, recommendation
4. **Metrics**: Coverage %, finding counts by severity
5. **Verdict**: CONSISTENT / GAPS FOUND / MISALIGNED

### Recommended Next Actions

Based on verdict:
- **CONSISTENT**: _"Safe to proceed to `/coder-implementation`"_
- **GAPS FOUND**: _"Address CRITICAL/HIGH findings. Run `/clarify-specification` for ambiguities, update design for coverage gaps."_
- **MISALIGNED**: _"Significant revision needed. Recommend returning to `/specifications-writer` or `/planner-architect`."_

> 🛑 **STOP HERE.** Report delivered. NEVER modify any files.

---

## 🔴 Constraints

1. **READ-ONLY**: NEVER modify spec, design, or task plan files.
2. **All three artifacts required**: Abort if any is missing.
3. **ANCHORS are non-negotiable**: Any ANCHORS violation is automatically CRITICAL.
4. **No implementation**: NEVER write code, create files, or generate tests.
5. **Advisory, not gating**: Unlike Reviewer/Tester, this workflow's verdict is recommended but not required before implementation.
6. **Objective findings only**: Report facts and specific recommendations. No vague "consider improving" suggestions.
