---
description:
  Workflow for effective design execution. Guides agent through discovery, design system generation, validation, and
  implementation.
---

# /design - Universal Design Workflow

Guide agents to design effectively for **any domain**: system architecture, database, API, UI/UX, infrastructure,
workflows, agents, etc.

---

## When to Use

- `/design` - Start design workflow
- Keywords: "design", "architect", "plan structure", "model", "schema"

---

## Phase 1: Classification & Routing 🔀

### Step 1.1: Identify Design Domain

| Domain             | Indicators                                              | Route To               |
| ------------------ | ------------------------------------------------------- | ---------------------- |
| **System/Arch**    | "system", "architecture", "microservices", "components" | `@architecture`        |
| **Database**       | "schema", "database", "data model", "tables"            | `@database-architect`  |
| **API**            | "API", "endpoints", "REST", "GraphQL"                   | `@api-patterns`        |
| **UI/UX**          | "interface", "UI", "dashboard", "landing page"          | `@frontend-design`     |
| **Infrastructure** | "infra", "deployment", "cloud", "kubernetes"            | `@devops-engineer`     |
| **AI/Agent**       | "agent", "AI system", "LLM", "workflow"                 | `@ai-agents-architect` |
| **Data Pipeline**  | "ETL", "pipeline", "data flow"                          | `@data-engineer`       |

### Step 1.2: Load Domain Skill

After classification, explicitly load the relevant agent/skill:

```markdown
🔀 **Design Domain:** [identified domain] 🤖 **Routing to:** `@[agent-name]`
```

---

## Phase 2: Discovery (MANDATORY) 🔍

**🛑 Do NOT skip. Bad inputs = bad designs.**

### Step 2.1: Context Analysis

Extract from user request:

| Dimension        | Question                                        |
| ---------------- | ----------------------------------------------- |
| **Goal**         | What problem does this design solve?            |
| **Scope**        | What is IN scope? What is OUT of scope?         |
| **Context**      | Greenfield / Feature / Refactor / Migration?    |
| **Constraints**  | Tech stack, budget, timeline, existing systems? |
| **Stakeholders** | Who will use/maintain this?                     |

### Step 2.2: Socratic Gate (If Unclear)

If request lacks clarity, ASK before proceeding:

```markdown
### 🎯 Discovery Questions

1. **Goal**: What is the primary objective of this design?
2. **Scope**: What are the boundaries? What should NOT be included?
3. **Context**: Is this new (greenfield) or modifying existing?
4. **Constraints**: Any limitations (tech, time, resources)?
5. **Success Criteria**: How will you know the design is successful?
```

**DO NOT proceed until sufficient clarity is achieved.**

---

## Phase 3: Design Thinking 🧠

### Step 3.1: Explore Options (Mandatory for Non-Trivial Designs)

For each major decision point, document options:

```markdown
### Decision: [Decision Name]

| Option   | Pros       | Cons        | Best For   |
| -------- | ---------- | ----------- | ---------- |
| Option A | [benefits] | [drawbacks] | [use case] |
| Option B | [benefits] | [drawbacks] | [use case] |
| Option C | [benefits] | [drawbacks] | [use case] |

**Recommendation:** [Choice] — [Rationale]
```

### Step 3.2: Apply Design Principles

Depending on domain, apply relevant principles:

| Domain   | Key Principles                                              |
| -------- | ----------------------------------------------------------- |
| System   | SOLID, DRY, Separation of Concerns, Modularity              |
| Database | Normalization, Indexing Strategy, Query Patterns, ACID/BASE |
| API      | RESTful conventions, Versioning, Pagination, Error Handling |
| UI/UX    | Consistency, Accessibility, Performance, User Mental Model  |
| Infra    | Immutability, Scalability, Observability, Security          |
| Agent/AI | Tool Design, Memory Strategy, Error Recovery, Guardrails    |

---

## Phase 4: Risk Assessment ⚠️

### Step 4.1: Classify Risk Level

| Factor            | Low              | Moderate         | High                    |
| ----------------- | ---------------- | ---------------- | ----------------------- |
| **Impact**        | Internal/small   | Customer-facing  | Business-critical       |
| **Reversibility** | Easy rollback    | Some effort      | Hard/impossible to undo |
| **Complexity**    | Single component | Multi-component  | System-wide             |
| **Uncertainty**   | Well-understood  | Some unknowns    | Many unknowns           |
| **Dependencies**  | Standalone       | Few dependencies | Many dependencies       |

### Step 4.2: Route Based on Risk

| Risk Level   | Action                                        |
| ------------ | --------------------------------------------- |
| **Low**      | Proceed to implementation                     |
| **Moderate** | Document assumptions, recommend review        |
| **High**     | REQUIRE `@[skills/multi-agent-brainstorming]` |

---

## Phase 5: Design Specification 📋

### Step 5.1: Document the Design

Create design artifact with structure appropriate to domain:

#### For System/Architecture:

```markdown
## System Design: [Name]

### Overview

[High-level description]

### Components

| Component | Responsibility | Dependencies |
| --------- | -------------- | ------------ |

### Data Flow

[Diagram or description]

### Key Decisions

| Decision | Choice | Rationale |
| -------- | ------ | --------- |
```

#### For Database:

```markdown
## Database Design: [Name]

### Schema Overview

[Tables/Collections]

### Relationships

[Diagram or description]

### Indexes

| Table | Index | Purpose |
| ----- | ----- | ------- |

### Query Patterns

[Expected queries and access patterns]
```

#### For API:

```markdown
## API Design: [Name]

### Endpoints

| Method | Path | Purpose |
| ------ | ---- | ------- |

### Request/Response Examples

[Samples]

### Error Handling

[Error codes and messages]
```

#### For UI/UX:

```markdown
## UI Design: [Name]

### Design Direction

[Aesthetic, style, differentiation]

### Components

[Key UI components needed]

### User Flows

[Primary user journeys]
```

### Step 5.2: Document Trade-offs & Limitations

Every design has trade-offs. Explicitly document:

```markdown
### Trade-offs Made

| Choice     | Benefit Gained | Cost Paid           |
| ---------- | -------------- | ------------------- |
| [Decision] | [What we get]  | [What we sacrifice] |

### Known Limitations

- [Limitation 1 with mitigation strategy]
- [Limitation 2 with mitigation strategy]
```

---

## Phase 6: Validation Checklist ✅

Before proceeding to implementation:

### Universal Checklist

- [ ] **Goal clarity**: Does the design solve the stated problem?
- [ ] **Scope adherence**: Is the design within defined boundaries?
- [ ] **Constraint respect**: Does it fit technical/resource constraints?
- [ ] **Trade-offs documented**: Are sacrifices explicit and acceptable?
- [ ] **Assumptions stated**: Are key assumptions documented?
- [ ] **Risk assessed**: Is the risk level appropriate?

### Domain-Specific Additions

| Domain   | Additional Checks                                        |
| -------- | -------------------------------------------------------- |
| System   | □ Scalability considered □ Failure modes identified      |
| Database | □ Query patterns validated □ Index strategy defined      |
| API      | □ Versioning strategy □ Auth/error handling defined      |
| UI/UX    | □ Accessibility addressed □ Responsive design considered |
| Infra    | □ Security reviewed □ Disaster recovery planned          |
| Agent/AI | □ Guardrails defined □ Error recovery documented         |

---

## Phase 7: Save & Notify 💾

After design completion:

1. **Save design document** to `.agent/docs/DESIGN-{slug}.md`
2. **Slug generation**: Extract key words → lowercase → hyphen-separated → max 30 chars
   - Examples:
     - "payment service database" → `DESIGN-payment-service-db.md`
     - "user authentication API" → `DESIGN-user-auth-api.md`
     - "agent memory system" → `DESIGN-agent-memory.md`
3. **Notify user**: `✅ Design saved: .agent/docs/DESIGN-{slug}.md`

---

## Quick Reference

### Workflow Flow

```
Classification → Discovery → Design Thinking → Risk Assessment → Specification → Validation
      ↓              ↓              ↓                ↓                  ↓
  Route to       Socratic      Explore          Multi-agent        Document
  domain skill     Gate        Options          (if high risk)     Trade-offs
```

### Domain → Agent Routing

| Domain         | Primary Agent          | Skills                          |
| -------------- | ---------------------- | ------------------------------- |
| Architecture   | `@orchestrator`        | `architecture`, `plan-writing`  |
| Database       | `@database-architect`  | `database-design`, `postgresql` |
| API            | `@backend-specialist`  | `api-patterns`, `api-design`    |
| UI/UX          | `@frontend-specialist` | `frontend-design`               |
| Infrastructure | `@devops-engineer`     | `docker-expert`, `kubernetes`   |
| AI/Agent       | `@ai-agents-architect` | `multi-agent-patterns`          |
| Data           | `@data-engineer`       | `database-design`               |

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern             | ✅ Instead                               |
| --------------------------- | ---------------------------------------- |
| Jump to implementation      | Complete Discovery first                 |
| Single option considered    | Explore at least 2-3 options             |
| No trade-offs documented    | Every choice has costs - document them   |
| Assume requirements         | Ask Socratic questions                   |
| Skip risk assessment        | Always classify risk before implementing |
| Design in isolation         | Consider dependencies and stakeholders   |
| Over-engineer first version | Start simple, iterate based on needs     |
