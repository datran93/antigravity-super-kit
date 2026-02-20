---
description:
  Workflow for effective design execution. Guides agent through discovery, design system generation, validation, and
  implementation.
---

# /design - Universal Design Workflow

Guide agents to design effectively for **any domain**: system architecture, database, API, UI/UX, infrastructure,
workflows, agents, etc.

## When to Use

- `/design` - Start design workflow
- Keywords: "design", "architect", "plan structure", "model", "schema"

---

## Phase 1: Classification & Skill Mapping 🔀

Analyze the request to determine the Design Domain. Map strictly to skills in `.agent/CATALOG.md` (e.g.
`architecture-patterns`, `database-architect`, `frontend-design`, `kubernetes`).

```markdown
🔀 **Design Domain:** [Target Domain] 🛠️ **Primary Skills:** `@[skill-1]`, `@[skill-2]`
```

---

## Phase 2: Discovery (MANDATORY) 🔍 (Socratic Gate)

**🛑 Do NOT skip. Bad inputs = bad designs.** Extract the following: Goal, Scope, Context (Greenfield vs Refactor),
Constraints, Stakeholders.

If the request is unclear, **ask 3+ clarifying questions** before proceeding. Wait for user answers.

---

## Phase 3: Design Thinking 🧠

- **Explore Options:** For major decisions, weigh Pros/Cons of Options A, B, C.
- **Apply Principles:** Use SOLID, DRY, RESTful conventions, Normalization, Accessibility, etc., depending on the
  selected skills.

---

## Phase 4: Risk Assessment ⚠️

Assess the risk (Impact, Complexity, Uncertainties). If High Risk, require team/user review before moving to
specification.

---

## Phase 5: Design Specification 📋

Document the finalized design tailored to the domain (e.g., Schema tables for DB, Endpoints for API, Components for
UI/UX).

**Trade-offs:** Explicitly document what benefits were gained and what sacrifices were made. Include known limitations.

---

## Phase 6: Validation & Save ✅

Validate against the original goals and constraints. Once user approves, save the design document to `DESIGN-{slug}.md`.

### Universal Checklist

- [ ] Goal clarity
- [ ] Scope adherence
- [ ] Constraints respected
- [ ] Trade-offs documented
