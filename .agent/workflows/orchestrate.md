---
description:
  Coordinate multiple skills for complex tasks. Use for multi-perspective analysis, comprehensive reviews, or complex
  implementations.
---

# /orchestrate - Skill Orchestration Workflow

Turn complex problems into solved solutions by coordinating specialized skills in sequence or parallel.

## When to Use

- `/orchestrate [complex task]`
- Tasks requiring **3+ domains** (e.g. Layout + API + DB + Security)

## 🔴 Critical Rules

1. **You are the Orchestrator**: Break tasks down and apply skills systematically.
2. **Context is King**: Maintain context between skill application phases.
3. **Sequential Logic**: Plan → Execute → Verify.

---

## Phase 1: Classification & Skill Mapping 🔀

Break down the request into domains and map to `.agent/CATALOG.md` skills.

```markdown
### Domain Analysis

- **Frontend**: `react-patterns`, `tailwind-patterns`
- **Backend**: `api-design-principles`, `backend-patterns`
- **Database**: `database-design`
```

---

## Phase 2: The Plan 🗺️ (Socratic Gate)

Draft an implementation plan (`PLAN.md`) by applying `project-planner` or `plan-writing` skills.

**🛑 STOP and ASK for Approval:**

> "✅ Plan created at `PLAN.md`. **Approve execution? (Y/N)**"

---

## Phase 3: Execution 🎻

Execute sequentially or in parallel depending on dependencies. For each phase:

1. Load the mapped skill.
2. Execute the task according to the skill's patterns.
3. Verify output before moving to the next.

### Example Sequence:

- **Foundation**: Database scaling & config (Skill: `database-design`)
- **Core Implementation**: Backend logic & API (Skill: `backend-patterns`)
- **UI/UX**: Frontend integration (Skill: `frontend-patterns`)

---

## Phase 4: Integration & QA 🧩

Ensure pieces fit together. Load quality skills (e.g., `test-engineer`, `security-review`) to verify security and
integration.

---

## Phase 5: Reporting 📝

Compile findings into `ORCHESTRATE-{slug}.md` and notify user.

```markdown
# 🎼 Orchestration Report: [Task Name]

## 🔄 Execution Log

1. **Database Setup**: ✅ Success
2. **API Implementation**: ✅ Success
3. **Frontend Integration**: ✅ Success

## 📦 Deliverables

- [ ] `PLAN.md`
- [ ] [Feature Code]
- [ ] [Tests]
```
