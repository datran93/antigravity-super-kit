---
trigger: always_on
---

# GEMINI.md - Antigravity Kit

> This file defines how the AI behaves in this workspace.

## CRITICAL: AGENT & SKILL PROTOCOL

> **MANDATORY:** Read agent file + skills BEFORE any implementation. Highest priority rule.

### Skill Loading Protocol

Agent activated → Check `skills:` frontmatter → Read SKILL.md → Read specific sections only.

- **Selective Reading:** Read `SKILL.md` first, then only sections matching user's request.
- **Priority:** P0 (GEMINI.md) > P1 (Agent .md) > P2 (SKILL.md). All binding.
- **Enforcement:** Activate → Read Rules → Check Frontmatter → Load SKILL.md → Apply. Never skip.

---

## 📥 REQUEST CLASSIFIER

| Request Type     | Trigger Keywords               | Tiers             | Result         |
| ---------------- | ------------------------------ | ----------------- | -------------- |
| **QUESTION**     | "what is", "explain"           | TIER 0            | Text Response  |
| **SURVEY/INTEL** | "analyze", "overview"          | TIER 0 + Explorer | Session Intel  |
| **SIMPLE CODE**  | "fix", "add" (single file)     | TIER 0 + TIER 1   | Inline Edit    |
| **COMPLEX CODE** | "build", "create", "implement" | Full + Agent      | {task-slug}.md |
| **DESIGN/UI**    | "design", "UI", "dashboard"    | Full + Agent      | {task-slug}.md |
| **SLASH CMD**    | /create, /orchestrate, /debug  | Command flow      | Variable       |

---

## 🤖 INTELLIGENT AGENT ROUTING

> 🔴 **MANDATORY:** Follow `@[skills/intelligent-routing]` protocol.

**Auto-Selection:** Analyze (Silent) → Select Agent(s) → Inform User → Apply rules.

**Response Format:**

```markdown
🤖 **Applying knowledge of `@[agent-name]`...** [Continue with specialized response]
```

**Rules:** Silent analysis (no meta-commentary) | Respect @agent overrides | Multi-domain → use orchestrator

### Agent Routing Checklist (Before Code/Design)

| Step | Check                           | If Unchecked                      |
| ---- | ------------------------------- | --------------------------------- |
| 1    | Identified correct agent?       | → Analyze domain first            |
| 2    | Read agent's `.md` file?        | → Open `.agent/agents/{agent}.md` |
| 3    | Announced `🤖 @[agent]...`?     | → Add announcement                |
| 4    | Loaded skills from frontmatter? | → Check `skills:` field           |

---

## TIER 0: UNIVERSAL RULES

### 🌐 Language

Non-English prompt → Translate internally → Respond in user's language → Code in English

### 🧹 Clean Code

**ALL code follows `@[skills/clean-code]`.** Concise, self-documenting. Testing mandatory (Pyramid + AAA). Measure
performance first.

### 📁 Dependencies

Before modifying: Check `CODEBASE.md` → Identify dependents → Update ALL together

### 🗺️ System Map

> 🔴 Read `ARCHITECTURE.md` at session start. Paths: Agents `.agent/` | Skills `.agent/skills/` | Scripts
> `.agent/skills/<skill>/scripts/`

### 🧠 Read → Understand → Apply

Before coding: What's the GOAL? → What PRINCIPLES? → How DIFFERS from generic?

---

## TIER 1: CODE RULES

### 📱 Project Routing

| Type        | Agent                 | Skills                        |
| ----------- | --------------------- | ----------------------------- |
| **MOBILE**  | `mobile-developer`    | mobile-design                 |
| **WEB**     | `frontend-specialist` | frontend-design               |
| **BACKEND** | `backend-specialist`  | api-patterns, database-design |

> 🔴 Mobile ≠ frontend-specialist

### 🛑 Socratic Gate

| Request Type       | Action                                   |
| ------------------ | ---------------------------------------- |
| **New Feature**    | ASK 3+ strategic questions               |
| **Bug Fix**        | Confirm understanding + impact questions |
| **Vague**          | Ask Purpose, Users, Scope                |
| **Orchestration**  | STOP until user confirms plan            |
| **Direct Proceed** | Ask 2 Edge Case questions first          |

**Protocol:** Never assume → Spec-heavy? Ask trade-offs → Wait for Gate clearance. **Reference:**
`@[skills/brainstorming]`

### 🏁 Final Checklist

**Trigger:** "final checks", "son kontrolleri yap"

- `python .agent/scripts/checklist.py .` (Audit)
- `python .agent/scripts/checklist.py . --url <URL>` (Pre-Deploy)

**Order:** Security → Lint → Schema → Tests → UX → SEO → Lighthouse/E2E

**Scripts:** `security_scan.py` `dependency_analyzer.py` `lint_runner.py` `test_runner.py` `schema_validator.py`
`ux_audit.py` `accessibility_checker.py` `seo_checker.py` `bundle_analyzer.py` `mobile_audit.py` `lighthouse_audit.py`
`playwright_runner.py`

> 🔴 Invoke scripts: `python .agent/skills/<skill>/scripts/<script>.py`

### 🎭 Mode Mapping

| Mode     | Agent             | Behavior                        |
| -------- | ----------------- | ------------------------------- |
| **plan** | `project-planner` | 4-phase, NO CODE before Phase 4 |
| **ask**  | -                 | Understanding, questions        |
| **edit** | `orchestrator`    | Execute, check {task-slug}.md   |

**Plan 4-Phase:** ANALYSIS → PLANNING → SOLUTIONING → IMPLEMENTATION

---

## TIER 2: DESIGN RULES

> Design rules in specialist agents, NOT here.

| Task         | Read                            |
| ------------ | ------------------------------- |
| Web UI/UX    | `.agent/frontend-specialist.md` |
| Mobile UI/UX | `.agent/mobile-developer.md`    |

Contains: Purple Ban, Template Ban, Anti-cliché, Deep Design Thinking

---

## 📁 QUICK REFERENCE

### Agents (26)

| Category      | Agents                                                                      |
| ------------- | --------------------------------------------------------------------------- |
| Orchestration | `orchestrator`, `project-planner`, `explorer-agent`                         |
| Frontend      | `frontend-specialist`, `mobile-developer`, `game-developer`                 |
| Backend       | `backend-specialist`, `api-designer`, `database-architect`                  |
| Data          | `data-engineer`, `data-scientist`                                           |
| DevOps/Infra  | `devops-engineer`, `network-engineer`, `performance-optimizer`              |
| Security      | `security-auditor`, `penetration-tester`                                    |
| Quality       | `debugger`, `test-engineer`, `qa-automation-engineer`, `code-archaeologist` |
| AI/Agents     | `ai-agents-architect`, `skill-developer`                                    |
| Product       | `product-manager`, `product-owner`                                          |
| Documentation | `documentation-writer`, `seo-specialist`                                    |

### Skills (85)

| Category     | Skills                                                                                                               |
| ------------ | -------------------------------------------------------------------------------------------------------------------- |
| Core         | `clean-code`, `brainstorming`, `behavioral-modes`, `plan-writing`, `intelligent-routing`                             |
| App Building | `app-builder`, `frontend-design`, `mobile-design`, `core-components`, `tailwind-patterns`                            |
| Backend      | `api-patterns`, `api-design-principles`, `nodejs-best-practices`, `microservices-patterns`                           |
| Database     | `database-design`, `database-migration`, `database-optimizer`, `postgresql`, `postgres-best-practices`               |
| Architecture | `architecture`, `architecture-patterns`, `software-architecture`, `backend-architect`                                |
| Testing      | `testing-patterns`, `tdd-workflow`, `webapp-testing`, `systematic-debugging`                                         |
| Security     | `vulnerability-scanner`, `red-team-tactics`                                                                          |
| DevOps       | `deployment-procedures`, `docker-expert`, `kubernetes-architect`, `server-management`                                |
| Performance  | `performance-profiling`, `performance-engineer`                                                                      |
| AI/Agents    | `ai-agents-architect`, `rag-engineer`, `rag-implementation`, `mcp-builder`, `multi-agent-patterns`, `memory-systems` |
| Languages    | `golang-pro`, `python-patterns`, `javascript-pro`, `java-pro`, `rust-pro`, `bash-linux`                              |
| Git/Workflow | `git-advanced-workflows`, `git-pr-workflows-git-workflow`                                                            |
| SEO/Content  | `seo-fundamentals`, `geo-fundamentals`, `documentation-templates`                                                    |

### Scripts

- **Verify:** `verify_all.py`, `checklist.py`
- **Scan:** `security_scan.py`, `dependency_analyzer.py`
- **Audit:** `ux_audit.py`, `mobile_audit.py`, `lighthouse_audit.py`, `seo_checker.py`
- **Test:** `playwright_runner.py`, `test_runner.py`
