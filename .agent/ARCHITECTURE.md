# Antigravity Kit Architecture

> Comprehensive AI Agent Capability Expansion Toolkit

---

## 📋 Overview

Antigravity Kit is a modular system consisting of:

- **26 Specialist Agents** - Role-based AI personas
- **85 Skills** - Domain-specific knowledge modules
- **11 Workflows** - Slash command procedures

---

## 🏗️ Directory Structure

```plaintext
.agent/
├── ARCHITECTURE.md          # This file
├── agents/                  # 26 Specialist Agents
├── skills/                  # 85 Skills
├── workflows/               # 11 Slash Commands
├── rules/                   # Global Rules
└── scripts/                 # Master Validation Scripts
```

---

## 🤖 Agents (26)

Specialist AI personas for different domains.

### Core Orchestration & Planning

| Agent             | Focus                    | Skills Used                               |
| ----------------- | ------------------------ | ----------------------------------------- |
| `orchestrator`    | Multi-agent coordination | parallel-agents, behavioral-modes         |
| `project-planner` | Discovery, task planning | brainstorming, plan-writing, architecture |

### Development Specialists

| Agent                 | Focus                 | Skills Used                                              |
| --------------------- | --------------------- | -------------------------------------------------------- |
| `frontend-specialist` | Web UI/UX             | frontend-design, react-best-practices, tailwind-patterns |
| `backend-specialist`  | API, business logic   | api-patterns, nodejs-best-practices, database-design     |
| `database-architect`  | Schema, SQL           | database-design, prisma-expert                           |
| `mobile-developer`    | iOS, Android, RN      | mobile-design                                            |
| `game-developer`      | Game logic, mechanics | game-development                                         |
| `api-designer`        | REST/GraphQL design   | api-design-principles, api-patterns                      |

### Infrastructure & Operations

| Agent              | Focus                 | Skills Used                          |
| ------------------ | --------------------- | ------------------------------------ |
| `devops-engineer`  | CI/CD, Docker         | deployment-procedures, docker-expert |
| `network-engineer` | Cloud networking      | network-engineer                     |
| `data-engineer`    | Pipelines, warehouses | data-engineer                        |

### Quality & Security

| Agent                    | Focus                     | Skills Used                                    |
| ------------------------ | ------------------------- | ---------------------------------------------- |
| `security-auditor`       | Security compliance       | vulnerability-scanner, red-team-tactics        |
| `penetration-tester`     | Offensive security        | red-team-tactics                               |
| `test-engineer`          | Testing strategies        | testing-patterns, tdd-workflow, webapp-testing |
| `debugger`               | Root cause analysis       | systematic-debugging                           |
| `performance-optimizer`  | Speed, Web Vitals         | performance-profiling, performance-engineer    |
| `qa-automation-engineer` | E2E testing, CI pipelines | webapp-testing, testing-patterns               |

### AI & Data Science

| Agent                 | Focus                   | Skills Used         |
| --------------------- | ----------------------- | ------------------- |
| `data-scientist`      | ML, analytics           | data-scientist      |
| `ai-agents-architect` | Agent systems, tool use | ai-agents-architect |

### Content & Documentation

| Agent                  | Focus               | Skills Used                        |
| ---------------------- | ------------------- | ---------------------------------- |
| `seo-specialist`       | Ranking, visibility | seo-fundamentals, geo-fundamentals |
| `documentation-writer` | Manuals, docs       | documentation-templates            |
| `skill-developer`      | Skill creation      | skill-developer                    |

### Product & Analysis

| Agent                | Focus                      | Skills Used                       |
| -------------------- | -------------------------- | --------------------------------- |
| `product-manager`    | Requirements, user stories | plan-writing, brainstorming       |
| `product-owner`      | Strategy, backlog, MVP     | plan-writing, brainstorming       |
| `code-archaeologist` | Legacy code, refactoring   | clean-code, code-review-checklist |
| `explorer-agent`     | Codebase analysis          | -                                 |

---

## 🧩 Skills (85)

Modular knowledge domains that agents can load on-demand based on task context.

### Frontend & UI

| Skill                   | Description                                                           |
| ----------------------- | --------------------------------------------------------------------- |
| `react-best-practices`  | React & Next.js performance optimization (Vercel - 57 rules)          |
| `web-design-guidelines` | Web UI audit - 100+ rules for accessibility, UX, performance (Vercel) |
| `tailwind-patterns`     | Tailwind CSS v4 utilities                                             |
| `frontend-design`       | UI/UX patterns, design systems                                        |
| `ui-ux-pro-max`         | 50 styles, 21 palettes, 50 fonts                                      |

### Backend & API

| Skill                             | Description                          |
| --------------------------------- | ------------------------------------ |
| `api-patterns`                    | REST, GraphQL, tRPC                  |
| `api-design-principles`           | API design, OpenAPI, versioning      |
| `nodejs-best-practices`           | Node.js async, modules               |
| `microservices-patterns`          | Service boundaries, event-driven     |
| `backend-architect`               | Scalable API design, distributed sys |
| `saga-orchestration`              | Distributed transactions, workflows  |
| `workflow-orchestration-patterns` | Durable workflows with Temporal      |

### Database

| Skill                     | Description                               |
| ------------------------- | ----------------------------------------- |
| `database-design`         | Schema design, optimization               |
| `database-migration`      | Zero-downtime migrations, rollbacks       |
| `database-optimizer`      | Query optimization, indexing              |
| `postgresql`              | PostgreSQL-specific best practices        |
| `postgres-best-practices` | Supabase/Postgres optimization (37 files) |

### Languages

| Skill                     | Description                            |
| ------------------------- | -------------------------------------- |
| `javascript-pro`          | ES6+, async patterns, Node.js          |
| `golang-pro`              | Go 1.21+, concurrency, microservices   |
| `python-patterns`         | Python standards, FastAPI, type hints  |
| `java-pro`                | Java 21+, Spring Boot, virtual threads |
| `rust-pro`                | Rust 1.75+, async, systems programming |
| `go-concurrency-patterns` | Goroutines, channels, sync primitives  |

### Cloud & Infrastructure

| Skill                   | Description                              |
| ----------------------- | ---------------------------------------- |
| `docker-expert`         | Containerization, Compose, security      |
| `kubernetes-architect`  | K8s, GitOps, service mesh, EKS/AKS/GKE   |
| `deployment-procedures` | CI/CD, safe deploy workflows             |
| `server-management`     | Infrastructure management                |
| `network-engineer`      | Cloud networking, security, optimization |
| `data-engineer`         | Data pipelines, warehouses, streaming    |

### Testing & Quality

| Skill                   | Description                            |
| ----------------------- | -------------------------------------- |
| `testing-patterns`      | Jest, Vitest, strategies               |
| `webapp-testing`        | E2E, Playwright                        |
| `tdd-workflow`          | Test-driven development                |
| `code-review-checklist` | Code review standards                  |
| `lint-and-validate`     | Linting, validation                    |
| `performance-engineer`  | Observability, optimization, profiling |

### Security

| Skill                   | Description              |
| ----------------------- | ------------------------ |
| `vulnerability-scanner` | Security auditing, OWASP |
| `red-team-tactics`      | Offensive security       |

### Architecture & Planning

| Skill           | Description                |
| --------------- | -------------------------- |
| `app-builder`   | Full-stack app scaffolding |
| `architecture`  | System design patterns     |
| `plan-writing`  | Task planning, breakdown   |
| `brainstorming` | Socratic questioning       |

### Mobile

| Skill           | Description           |
| --------------- | --------------------- |
| `mobile-design` | Mobile UI/UX patterns |

### Game Development

| Skill              | Description           |
| ------------------ | --------------------- |
| `game-development` | Game logic, mechanics |

### SEO & Growth

| Skill              | Description                   |
| ------------------ | ----------------------------- |
| `seo-fundamentals` | SEO, E-E-A-T, Core Web Vitals |
| `geo-fundamentals` | GenAI optimization            |

### Shell/CLI

| Skill                | Description               |
| -------------------- | ------------------------- |
| `bash-linux`         | Linux commands, scripting |
| `powershell-windows` | Windows PowerShell        |

### AI & Agents

| Skill                               | Description                                    |
| ----------------------------------- | ---------------------------------------------- |
| `ai-agents-architect`               | Agent design, tool use, memory systems         |
| `agent-memory-systems`              | Short-term, long-term, cognitive architectures |
| `agent-orchestration-improve-agent` | Agent performance analysis, iteration          |
| `multi-agent-patterns`              | Orchestrator, peer-to-peer, hierarchical       |
| `memory-systems`                    | Graph-based memory architectures               |
| `rag-engineer`                      | Embeddings, vector search, retrieval           |
| `rag-implementation`                | RAG systems for LLM applications               |
| `mcp-builder`                       | Model Context Protocol servers                 |
| `vector-database-engineer`          | Pinecone, Weaviate, Qdrant, pgvector           |
| `data-scientist`                    | ML workflows, statistical analysis             |
| `deep-research`                     | Gemini Deep Research Agent integration         |

### Knowledge Management

| Skill                     | Description                   |
| ------------------------- | ----------------------------- |
| `skill-developer`         | Skill creation, documentation |
| `documentation-templates` | Doc formats                   |

### Architecture & Patterns

| Skill                     | Description                                    |
| ------------------------- | ---------------------------------------------- |
| `architecture`            | System design patterns, ADR                    |
| `architecture-patterns`   | Clean, Hexagonal, DDD                          |
| `software-architecture`   | Quality-focused architecture guide             |
| `error-handling-patterns` | Exceptions, Result types, graceful degradation |

### Git & Workflows

| Skill                           | Description                      |
| ------------------------------- | -------------------------------- |
| `git-advanced-workflows`        | Rebasing, cherry-picking, bisect |
| `git-pr-workflows-git-workflow` | PR creation, code review flow    |

### Core & Other

| Skill                       | Description                   |
| --------------------------- | ----------------------------- |
| `clean-code`                | Coding standards (Global)     |
| `behavioral-modes`          | Agent personas                |
| `intelligent-routing`       | Auto agent selection          |
| `parallel-agents`           | Multi-agent orchestration     |
| `context-compression`       | Long-session compression      |
| `context-optimization`      | Compaction, masking, caching  |
| `i18n-localization`         | Internationalization          |
| `design-orchestration`      | Design workflow routing       |
| `multi-agent-brainstorming` | Sequential multi-agent review |
| `performance-profiling`     | Web Vitals, optimization      |
| `systematic-debugging`      | Troubleshooting (11 files)    |

---

## 🔄 Workflows (11)

Slash command procedures. Invoke with `/command`.

| Command          | Description              |
| ---------------- | ------------------------ |
| `/brainstorm`    | Socratic discovery       |
| `/create`        | Create new features      |
| `/debug`         | Debug issues             |
| `/deploy`        | Deploy application       |
| `/enhance`       | Improve existing code    |
| `/orchestrate`   | Multi-agent coordination |
| `/plan`          | Task breakdown           |
| `/preview`       | Preview changes          |
| `/status`        | Check project status     |
| `/test`          | Run tests                |
| `/ui-ux-pro-max` | Design with 50 styles    |

---

## 🎯 Skill Loading Protocol

```plaintext
User Request → Skill Description Match → Load SKILL.md
                                            ↓
                                    Read references/
                                            ↓
                                    Read scripts/
```

### Skill Structure

```plaintext
skill-name/
├── SKILL.md           # (Required) Metadata & instructions
├── scripts/           # (Optional) Python/Bash scripts
├── references/        # (Optional) Templates, docs
└── assets/            # (Optional) Images, logos
```

### Enhanced Skills (with scripts/references)

| Skill           | Files | Coverage                         |
| --------------- | ----- | -------------------------------- |
| `ui-ux-pro-max` | 27    | 50 styles, 21 palettes, 50 fonts |
| `app-builder`   | 20    | Full-stack scaffolding           |

---

## � Scripts (2)

Master validation scripts that orchestrate skill-level scripts.

### Master Scripts

| Script          | Purpose                                 | When to Use              |
| --------------- | --------------------------------------- | ------------------------ |
| `checklist.py`  | Priority-based validation (Core checks) | Development, pre-commit  |
| `verify_all.py` | Comprehensive verification (All checks) | Pre-deployment, releases |

### Usage

```bash
# Quick validation during development
python .agent/scripts/checklist.py .

# Full verification before deployment
python .agent/scripts/verify_all.py . --url http://localhost:3000
```

### What They Check

**checklist.py** (Core checks):

- Security (vulnerabilities, secrets)
- Code Quality (lint, types)
- Schema Validation
- Test Suite
- UX Audit
- SEO Check

**verify_all.py** (Full suite):

- Everything in checklist.py PLUS:
- Lighthouse (Core Web Vitals)
- Playwright E2E
- Bundle Analysis
- Mobile Audit
- i18n Check

For details, see [scripts/README.md](scripts/README.md)

---

## 📊 Statistics

| Metric              | Value                         |
| ------------------- | ----------------------------- |
| **Total Agents**    | 26                            |
| **Total Skills**    | 85                            |
| **Total Workflows** | 11                            |
| **Total Scripts**   | 2 (master) + 18 (skill-level) |
| **Coverage**        | ~90% web/mobile development   |

---

## 🔗 Quick Reference

| Need           | Agent                 | Skills                                |
| -------------- | --------------------- | ------------------------------------- |
| Web App        | `frontend-specialist` | react-best-practices, frontend-design |
| API            | `backend-specialist`  | api-patterns, nodejs-best-practices   |
| API Design     | `api-designer`        | api-design-principles, api-patterns   |
| Mobile         | `mobile-developer`    | mobile-design                         |
| Database       | `database-architect`  | database-design, prisma-expert        |
| Security       | `security-auditor`    | vulnerability-scanner                 |
| Testing        | `test-engineer`       | testing-patterns, webapp-testing      |
| Debug          | `debugger`            | systematic-debugging                  |
| Plan           | `project-planner`     | brainstorming, plan-writing           |
| ML/Analytics   | `data-scientist`      | data-scientist                        |
| AI Agents      | `ai-agents-architect` | ai-agents-architect                   |
| Networking     | `network-engineer`    | network-engineer                      |
| Data Pipelines | `data-engineer`       | data-engineer                         |
| Skills         | `skill-developer`     | skill-developer                       |
