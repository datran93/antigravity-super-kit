---
description: Create new application command. Triggers App Builder skill and starts interactive dialogue with user.
---

# /create - Universal Creation Workflow

Guide agents to create new things systematically: services, systems, apps, components, modules, pipelines, etc.

---

## When to Use

- `/create [description]` - Start creation workflow
- Keywords: "create", "build", "make", "new", "implement", "set up"

---

## Phase 1: Classification & Routing 🔀

### Step 1.1: Identify Creation Type

| Type               | Indicators                                  | Primary Agent          |
| ------------------ | ------------------------------------------- | ---------------------- |
| **Web App**        | "app", "website", "landing", "dashboard"    | `@frontend-specialist` |
| **API/Service**    | "service", "API", "backend", "microservice" | `@backend-specialist`  |
| **Database**       | "schema", "database", "data model"          | `@database-architect`  |
| **Infrastructure** | "deploy", "infra", "kubernetes", "docker"   | `@devops-engineer`     |
| **Agent/AI**       | "agent", "AI", "automation", "workflow"     | `@ai-agents-architect` |
| **Data Pipeline**  | "ETL", "pipeline", "data flow"              | `@data-engineer`       |
| **Full Stack**     | Multiple of above                           | `@orchestrator`        |

### Step 1.2: Announce Routing

```markdown
🔀 **Creation Type:** [identified type] 🤖 **Primary Agent:** `@[agent-name]` 📦 **Coordinating:** [list other agents if
multi-domain]
```

---

## Phase 2: Discovery (MANDATORY) 🔍

**🛑 Do NOT start building without understanding requirements.**

### Step 2.1: Requirement Extraction

Extract from user request:

| Dimension       | Question                                 |
| --------------- | ---------------------------------------- |
| **What**        | What exactly needs to be created?        |
| **Why**         | What problem does this solve?            |
| **Who**         | Who will use this?                       |
| **Scope**       | Must-have features vs nice-to-have?      |
| **Constraints** | Tech stack, timeline, budget?            |
| **Context**     | Greenfield or integrating with existing? |

### Step 2.2: Clarifying Questions (If Unclear)

If request lacks detail, ASK before proceeding:

```markdown
### 🔍 Before We Start

1. **Core Features**: What are the must-have features?
2. **Users**: Who will use this and how?
3. **Tech Preference**: Any preferred technologies?
4. **Integrations**: Any existing systems to connect with?
5. **Scale**: Expected usage/load?
```

### Step 2.3: Confirm Understanding

```markdown
### ✅ Requirements Confirmed

**Creating:** [what] **Purpose:** [why] **Users:** [who] **Core Features:**

- [Feature 1]
- [Feature 2]
- [Feature 3]

**Tech Stack:** [planned stack]

> Does this look correct? If yes, I'll proceed with planning.
```

---

## Phase 3: Planning 📋

### Step 3.1: Architecture Decision

For non-trivial creations, outline key decisions:

```markdown
### Architecture Decisions

| Decision     | Choice   | Rationale |
| ------------ | -------- | --------- |
| [Decision 1] | [Choice] | [Why]     |
| [Decision 2] | [Choice] | [Why]     |
```

### Step 3.2: Component Breakdown

Identify components to build:

```markdown
### Components

| Component      | Responsibility          | Agent                  |
| -------------- | ----------------------- | ---------------------- |
| Database       | Schema & migrations     | `@database-architect`  |
| Backend API    | Business logic & routes | `@backend-specialist`  |
| Frontend       | User interface          | `@frontend-specialist` |
| Infrastructure | Deployment config       | `@devops-engineer`     |
```

### Step 3.3: File Structure Preview

```markdown
### Proposed Structure
```

project/ ├── src/ │ ├── components/ # [purpose] │ ├── services/ # [purpose] │ ├── models/ # [purpose] │ └── utils/ #
[purpose] ├── config/ # [purpose] ├── tests/ # [purpose] └── docs/ # [purpose]

```

```

### Step 3.4: Implementation Order

```markdown
### Build Order

1. [ ] **Foundation** - Project setup, dependencies
2. [ ] **Data Layer** - Database schema, models
3. [ ] **Business Logic** - Core services, API
4. [ ] **Interface** - UI/CLI/API endpoints
5. [ ] **Integration** - Connect components
6. [ ] **Testing** - Unit, integration tests
7. [ ] **Deployment** - Config, scripts
```

---

## Phase 4: Implementation 🔨

### Step 4.1: Execute by Layer

Follow the build order. For each layer:

1. **Announce** what you're building
2. **Build** the component
3. **Verify** it works before moving on
4. **Report** progress

```markdown
### 🔨 Building: [Component Name]

**Status:** 🔄 In Progress | ✅ Complete | ❌ Blocked

**Files Created:**

- `path/to/file1.ts` - [purpose]
- `path/to/file2.ts` - [purpose]

**Next:** [what comes next]
```

### Step 4.2: Agent Coordination (Multi-Domain)

For full-stack or complex creations:

| Order | Agent                  | Deliverable        |
| ----- | ---------------------- | ------------------ |
| 1     | `@database-architect`  | Schema, migrations |
| 2     | `@backend-specialist`  | API, services      |
| 3     | `@frontend-specialist` | UI components      |
| 4     | `@devops-engineer`     | Deployment config  |
| 5     | `@test-engineer`       | Test suite         |

### Step 4.3: Quality Checks

During implementation, verify:

- [ ] Code follows project conventions
- [ ] No hardcoded values
- [ ] Error handling in place
- [ ] Dependencies documented
- [ ] Tests written where needed

---

## Phase 5: Verification ✅

### Step 5.1: Smoke Test

```bash
# For web apps
npm run dev  # or equivalent

# For services
npm run start  # or equivalent

# For APIs
curl http://localhost:PORT/health
```

### Step 5.2: Functionality Check

Verify core features work:

```markdown
### Verification Checklist

- [ ] Application starts without errors
- [ ] Core feature 1 works
- [ ] Core feature 2 works
- [ ] Core feature 3 works
- [ ] No console errors/warnings
```

### Step 5.3: Preview (If Applicable)

For web apps/UIs:

```bash
# Start preview server
python3 .agent/.shared/scripts/auto_preview.py --dir [project_dir]
```

Present URL to user:

```markdown
🌐 **Preview:** http://localhost:PORT
```

---

## Phase 6: Documentation & Delivery 📝

### Step 6.1: Creation Summary

````markdown
## ✅ Creation Complete: [Name]

### What Was Created

[Brief description]

### Tech Stack

- **Frontend:** [if applicable]
- **Backend:** [if applicable]
- **Database:** [if applicable]
- **Other:** [if applicable]

### Files Created

| Path      | Purpose   |
| --------- | --------- |
| `src/...` | [purpose] |

### Features Implemented

- ✅ [Feature 1]
- ✅ [Feature 2]
- ✅ [Feature 3]

### How to Run

```bash
[commands to start/run]
```
````

### Next Steps

- [ ] [Recommended enhancement 1]
- [ ] [Recommended enhancement 2]

```

### Step 6.2: Save & Notify

1. Save summary to `.agent/docs/CREATE-{slug}.md`
2. **Slug generation**: Extract 2-3 key words → lowercase → hyphen-separated → max 30 chars
   - "blog site with comments" → `CREATE-blog-comments.md`
   - "user authentication service" → `CREATE-auth-service.md`
   - "inventory management system" → `CREATE-inventory-system.md`
3. Notify: `✅ Creation complete! Summary saved: .agent/docs/CREATE-{slug}.md`

---

## Quick Reference

### Workflow Flow

```

Classification → Discovery → Planning → Implementation → Verification → Delivery ↓ ↓ ↓ ↓ ↓ ↓ Route to Clarify
Architecture Build by Smoke test Document agents + Confirm + Structure layer + Preview + Save

````

### Common Creation Types

| Type          | Stack Defaults            | Agents                |
| ------------- | ------------------------- | --------------------- |
| Web App       | React/Next.js + Tailwind  | frontend, backend, db |
| API Service   | Node.js/Go + PostgreSQL   | backend, db, devops   |
| CLI Tool      | Node.js/Python            | backend               |
| Agent         | Python + LangChain/Custom | ai-agents, backend    |
| Data Pipeline | Python + Airflow/dbt      | data-engineer, db     |

### Status Icons

| Icon | Meaning        |
| ---- | -------------- |
| 🔄    | In Progress    |
| ✅    | Complete       |
| ❌    | Failed/Blocked |
| ⏳    | Waiting        |
| ⚠️    | Warning        |

---

## Anti-Patterns (AVOID)

| ❌ Anti-Pattern                    | ✅ Instead                              |
| --------------------------------- | -------------------------------------- |
| Start coding immediately          | Complete Discovery first               |
| Build everything at once          | Build layer by layer, verify each      |
| Skip planning for "simple" things | Always have at least minimal plan      |
| Assume tech stack                 | Ask if not specified                   |
| No verification before delivery   | Always smoke test                      |
| Missing documentation             | Document what was created + how to run |
| Hardcoded values                  | Use config/env variables               |

---

## Examples

```bash
/create blog site with authentication and comments
/create REST API for inventory management
/create CLI tool for data migration
/create AI agent for customer support
/create microservice for payment processing
/create dashboard for analytics metrics
````
