---
trigger: always_on
---

# GEMINI.md - Antigravity Kit

## 🎯 UNIVERSAL CORE RULES

### 🚀 Mandatory Initialization

- **Session Start**: At the beginning of **EVERY** session, the Agent MUST read `@/.agents/rules/GEMINI.md` to ensure
  all core protocols are fresh, and MUST read `@/.agents/rules/ANCHORS.md` to load the immutable facts and project
  guardrails.
- **Proactive MCP Tooling**: Always use MCP tools for discovery, research, and analysis tasks **automatically**. Do NOT
  ask for permission to use read-only discovery tools.

### 🛠️ MCP Tool Mastery (Priority Over Bash)

```
mcp/
├── @mcp:ast-explorer/
│   ├── get_project_architecture   → AST structural overview of the codebase
│   └── search_symbol              → Find class/function definitions by name
│
├── @mcp:context-manager/
│   ├── save_checkpoint            → Persist current task state for recovery
│   ├── load_checkpoint            → Restore a previously saved task state
│   ├── list_active_tasks          → List all active task plans
│   ├── initialize_task_plan       → Register a new ordered task plan
│   ├── complete_task_step         → Mark an Action complete (Planner only)
│   ├── add_task_step              → Append a new step to an existing plan
│   ├── declare_intent             → Lock files to a tactic (Intent Lock)
│   ├── check_intent_lock          → Verify scope before modifying a file
│   ├── recall_knowledge           → Search past Knowledge Items via FTS5
│   ├── compact_memory             → Extract and index a KI from tactic
│   ├── record_failure             → Increment drift counter on failure
│   ├── clear_drift                → Reset drift counter after success
│   ├── manage_anchors             → Get/set immutable system invariant rules
│   └── annotate_file              → Inject Ghost Context into a file
│
├── @mcp:context7/
│   ├── resolve-library-id         → Resolve a library name to Context7 ID
│   └── query-docs                 → Fetch up-to-date docs & code examples
│
├── @mcp:database-inspector/
│   ├── list_tables                → List all tables/views in the database
│   ├── get_table_sample           → DDL + 5 sample rows for a table
│   ├── inspect_schema             → Detailed column/key schema for a table
│   ├── explain_query              → EXPLAIN ANALYZE for performance analysis
│   ├── run_read_query             → Read-only SQL with pagination
│   └── run_write_query            → Write SQL — ask USER first, confirm=True
│
├── @mcp:doc-researcher/
│   ├── search_latest_syntax       → Real-time web search for SOTA syntax
│   ├── read_doc_file              → Read local .doc/.docx/.pdf with pagination
│   └── read_website_markdown      → Scrape a URL as clean Markdown
│
├── @mcp:figma-reader/
│   ├── read_figma_design          → File structure & metadata from Figma URL
│   ├── export_figma_images        → Render Figma nodes as image URLs
│   └── get_design_details         → Deep JSON for specific Figma node IDs
│
├── @mcp:gitlab-mr-discussions/
│   ├── read_mr_discussions        → Fetch all threads from a GitLab MR
│   ├── reply_to_mr_discussion     → Post a reply to a discussion thread
│   └── resolve_mr_discussion      → Resolve or unresolve a thread
│
├── @mcp:mcp-http-client/
│   ├── http_request               → Execute HTTP with {{var}} placeholders
│   ├── import_curl                → Parse and execute a raw cURL command
│   ├── set_env                    → Set {{key}} environment variables
│   ├── set_config                 → Configure base URL and auth token
│   ├── list_history               → View request history in .rest format
│   └── clear_history              → Clear all request history
│
├── @mcp:skill-router/
│   └── search_skills              → Semantic skill search with tags_filter
│
├── @mcp:codebase-search/
│   ├── index_codebase             → Index a project for hybrid semantic + keyword search
│   ├── search_code                → RRF hybrid search (BM25 + cosine) over codebase
│   ├── get_indexing_status        → Track indexing progress
│   └── clear_index                → Clear project search index
│
└── @mcp:stitch/
    ├── generate_screen_from_text  → Generate a UI screen from a text prompt
    ├── edit_screens               → Edit existing screens with a prompt
    ├── generate_variants          → Create variants of existing screens
    ├── create_project             → Create a new Stitch project
    ├── get_project                → Get details of a Stitch project
    ├── list_projects              → List all accessible Stitch projects
    ├── get_screen                 → Get a specific screen from a project
    └── list_screens               → List all screens in a project
```

---

## 🏛️ ROLE ARCHITECTURE

The system uses **distinct, non-overlapping roles**. Each role has a single responsibility, produces a specific output,
and stops. **No role transitions between roles** — the USER decides when to invoke the next role.

```
[Spec Writer] → [Planner] → [Coder] → [Reviewer] → [Tester] → [Planner]
```

### Role Definitions

| Role               | Slash Command            | Responsibility                                                          | Output                                | Stops When                            |
| ------------------ | ------------------------ | ----------------------------------------------------------------------- | ------------------------------------- | ------------------------------------- |
| **📝 Spec Writer** | `/specifications-writer` | Requirements engineering & Socratic questioning                         | `SPEC.md`                             | Requirements are unambiguous          |
| **🏗️ Planner**     | `/planner-architect`     | Architecture design + ordered task list. Also commits after gates pass  | `DESIGN.md` + task plan + git commits | Plan delivered OR all tasks committed |
| **💻 Coder**       | `/coder-implementation`  | Reads `DESIGN.md` + task list, implements each Action in order          | Code changes + implementation report  | All Actions implemented and reported  |
| **🔍 Reviewer**    | `/reviewer-audit`        | Reads Coder report + `DESIGN.md`, audits code quality and correctness   | Audit report (APPROVED / NEEDS FIX)   | Report delivered to USER              |
| **🧪 Tester**      | `/tester-verification`   | Writes unit + integration tests for Coder's code, measures coverage     | Test files + coverage report          | Coverage ≥ 70% achieved and reported  |
| **🧭 Router**      | `/smart-route`           | Classifies user intent and routes to the correct workflow automatically | Confirmation + routed workflow        | USER confirms routing decision        |

> Each role's specific boundaries and constraints are defined in its own workflow file (see `Critical Constraints`
> section). The universal rule: any role that hits a blocker **stops and asks the USER** — never self-escalates.

---

## ⛔ CORE COMMUNICATION & PROTOCOLS

1. **Role Anchoring**: ALWAYS prefix every conversational response with the current role tag:
   - `[Role: 📝 Spec Writer]` / `[Role: 🏗️ Planner]` / `[Role: 💻 Coder]` / `[Role: 🔍 Reviewer]` / `[Role: 🧪 Tester]`
2. **Strict Output Contract**: Each role delivers its defined output then **stops**. It does not initiate the next role.
3. **Explicit Resource Ownership**:
   - **Planner**: Owns `DESIGN.md`, task plan, `git commit`, `complete_task_step`.
   - **Coder**: Owns source code changes and implementation report.
   - **Reviewer**: Owns the audit report.
   - **Tester**: Owns the test suite and coverage report.
4. **Skill Transparency**: Explicitly state which specialized skills are used BEFORE executing tasks.

---

## 🛡️ UNIVERSAL GUARDRAILS (Applies to ALL Roles)

### A. Drift Detection

If any role is stuck on the **same issue 3 times consecutively** (same test failing, same lint error, same design
conflict):

1. Call `@mcp:context-manager` (`record_failure`).
2. **Stop immediately** — do not attempt a 4th fix.
3. Report the blocker clearly to the USER with:
   - What was attempted (3 times)
   - What the error/rejection was
   - What decision or information is needed to unblock

> ⚠️ No role self-escalates or switches role on drift. The USER decides how to proceed.

### B. Ghost Context (Before Stopping)

Before any role finishes and stops, use `@mcp:context-manager` (`annotate_file`) to inject non-obvious gotchas,
architectural quirks, or library limitations directly into the affected files — so the next role has immediate context
without re-reading everything.

### C. Quality Gates (Planner-Enforced)

Gate conditions and commit protocol are defined in [`planner-architect.md`](./../workflows/planner-architect.md) **Phase
5**. Both Reviewer (APPROVED) and Tester (≥ 70% coverage) gates must pass before any commit. If either fails, the
Planner asks the USER — it does not auto-loop.

---

## 📌 Metadata

- **Version**: 2.2.0
- **Last Updated**: 2026-03-15
- **Related**: `.agents/workflows/*.md`, `.agents/rules/ANCHORS.md`
- **New in 2.2.0**: Removed mcp-context-governor (not functional in Gemini Code Assist). Git hooks moved to
  `.agents/hooks/git/` via `core.hooksPath`.
