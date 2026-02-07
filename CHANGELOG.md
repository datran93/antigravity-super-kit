# Changelog

All notable changes to the Antigravity Kit will be documented in this file.

## [Unreleased]

## [2.1.0] - 2026-02-06

### Added

- **New CLI Features**:
    - `agk` - Global CLI script for easy installation and updates of `.agent` folder
        - Auto-caching of `antigravity-kit` repository
        - `agk install`: Install agents to current project
        - `agk update`: Update agents in current project
        - `agk status`: Check if agents are up to date

- **New Specialized Agents** (6 total, expanding from 13 to 19 agents):
    - `data-scientist` - Machine learning, statistical analysis, and model deployment expertise
    - `ai-agents-architect` - Autonomous agent design, tool use, and memory architecture
    - `network-engineer` - Cloud networking, security, and performance optimization
    - `data-engineer` - Data pipelines, warehouses, and streaming architecture
    - `skill-developer` - Creating and managing skills, knowledge packaging
    - `api-designer` - REST/GraphQL design, OpenAPI specs, and developer experience

### Changed

- **Agent Enhancements** (comprehensive documentation improvements):
    - `test-engineer`: Enhanced from 159 to ~500 lines (+314% content)
        - Added testing pyramid with percentages and trade-offs
        - Framework comparison matrices (Vitest, Jest, Pytest, Playwright)
        - Expanded TDD workflow with Red-Green-Refactor cycle
        - Advanced patterns: mutation testing, contract testing, property-based testing
        - Integration/E2E testing with Supertest and Playwright examples
        - CI/CD integration patterns and test parallelization
        - 10+ practical code examples in TypeScript and Python
    - `security-auditor`: Enhanced from 171 to ~600 lines (+351% content)
        - Added 5-phase security methodology (UNDERSTAND → ANALYZE → PRIORITIZE → REPORT → VERIFY)
        - STRIDE threat modeling framework with examples
        - CVSS+EPSS risk prioritization decision tree
        - 4 vulnerability code examples with fixes (IDOR, SQL injection, XSS, hardcoded secrets)
        - Supply chain security audit checklist and vetting process
        - Authentication/authorization best practices
        - Security headers and CORS configuration examples
        - Cryptography decision tree and algorithm selection matrix
        - Incident response 5-phase workflow
    - `backend-specialist`: Added partial Golang support

### Impact

- **Specialist Coverage**: Expanded from 13 to 19 agents (+46%)
- **Domain Expertise**: Added 6 new critical domains (ML/AI, Agent Architecture, Network Engineering, Data Engineering, Skill Management, API Design)
- **Documentation Quality**: Added 3,115 lines of new agent documentation plus 770+ lines of enhancements
- **Code Examples**: 18+ new practical code examples across 4 languages
- **Decision Frameworks**: 7 new comprehensive decision frameworks for technology selection

## [2.0.2] - 2026-02-04

- **New Skills**:
    - `rust-pro` - Master Rust 1.75+
- **Agent Workflows**:
    - Updated `orchestrate.md` fix output turkish

## [2.0.1] - 2026-01-26

### Added

- **Agent Flow Documentation**: New comprehensive workflow documentation
    - Added `.agent/AGENT_FLOW.md` - Complete agent flow architecture guide
    - Documented Agent Routing Checklist (mandatory steps before code/design work)
    - Documented Socratic Gate Protocol for requirement clarification
    - Added Cross-Skill References pattern documentation
- **New Skills**:
    - `react-best-practices` - Consolidated Next.js and React expertise
    - `web-design-guidelines` - Professional web design standards and patterns

### Changed

- **Skill Consolidation**: Merged `nextjs-best-practices` and `react-patterns` into unified `react-best-practices` skill
- **Architecture Updates**:
    - Enhanced `.agent/ARCHITECTURE.md` with improved flow diagrams
    - Updated `.agent/rules/GEMINI.md` with Agent Routing Checklist
- **Agent Updates**:
    - Updated `frontend-specialist.md` with new skill references
    - Updated `qa-automation-engineer.md` with enhanced testing workflows
- **Frontend Design Skill**: Enhanced `frontend-design/SKILL.md` with cross-references to `web-design-guidelines`

### Removed

- Deprecated `nextjs-best-practices` skill (consolidated into `react-best-practices`)
- Deprecated `react-patterns` skill (consolidated into `react-best-practices`)

### Fixed

- **Agent Flow Accuracy**: Corrected misleading terminology in AGENT_FLOW.md
    - Changed "Parallel Execution" → "Sequential Multi-Domain Execution"
    - Changed "Integration Layer" → "Code Coherence" with accurate description
    - Added reality notes about AI's sequential processing vs. simulated multi-agent behavior
    - Clarified that scripts require user approval (not auto-executed)

## [2.0.0] - Unreleased

### Initial Release

- Initial release of Antigravity Kit
- 20 specialized AI agents
- 37 domain-specific skills
- 11 workflow slash commands
- CLI tool for easy installation and updates
- Comprehensive documentation and architecture guide

[Unreleased]: https://github.com/datran/antigravity-kit/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/datran/antigravity-kit/releases/tag/v2.0.0
