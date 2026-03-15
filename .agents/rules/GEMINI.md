---
trigger: always_on
---

# GEMINI.md - Antigravity Kit

## 🎯 UNIVERSAL CORE RULES

### 🚀 Mandatory Initialization

- **Session Start**: At the beginning of **EVERY** session, read `@/.agents/rules/GEMINI.md` and
  `@/.agents/rules/ANCHORS.md`.
- **Proactive MCP Tooling**: Always use MCP tools for discovery, research, and analysis tasks **automatically**. Do NOT
  ask for permission to use read-only discovery tools.

### 🛠️ MCP Tool Mastery (Priority Over Bash)

```
mcp/
├── @mcp:ast-explorer/        get_project_architecture · search_symbol
├── @mcp:context-manager/     save/load_checkpoint · initialize_task_plan · complete_task_step
│                             add_task_step · declare_intent · check_intent_lock
│                             recall_knowledge · compact_memory · record_failure
│                             clear_drift · manage_anchors · annotate_file
│                             find_recent_task · review_checkpoint
├── @mcp:context7/            resolve-library-id · query-docs
├── @mcp:database-inspector/  list_tables · get_table_sample · inspect_schema
│                             explain_query · run_read_query · run_write_query
├── @mcp:doc-researcher/      search_latest_syntax · read_doc_file · read_website_markdown
├── @mcp:figma-reader/        read_figma_design · export_figma_images · get_design_details
├── @mcp:gitlab-mr-discussions/ read_mr_discussions · reply_to_mr_discussion · resolve_mr_discussion
├── @mcp:github-reader/       get_file_content · list_directory · get_repository_info · search_code
├── @mcp:mcp-http-client/     http_request · import_curl · set_env · set_config · list_history · clear_history
├── @mcp:skill-router/        search_skills
├── @mcp:codebase-search/     index_codebase · search_code · get_indexing_status · clear_index
└── @mcp:stitch/              generate_screen_from_text · edit_screens · generate_variants
                              create_project · get_project · list_projects · get_screen · list_screens
```

---

## 🏛️ ROLE ARCHITECTURE

Distinct, non-overlapping roles. Each role has a single responsibility, produces a specific output, and stops. **No role
transitions** — the USER decides when to invoke the next role.

```
[Spec Writer] → [Planner] → [Coder] → [Reviewer] → [Tester] → [Planner]
```

| Role               | Slash Command            | Output                                         | Stops When                            |
| ------------------ | ------------------------ | ---------------------------------------------- | ------------------------------------- |
| **📝 Spec Writer** | `/specifications-writer` | `spec/spec-*.md`                               | Requirements are unambiguous          |
| **🏗️ Planner**     | `/planner-architect`     | `design/design-*.md` + task plan + git commits | Plan delivered OR all tasks committed |
| **💻 Coder**       | `/coder-implementation`  | Code changes + implementation report           | All Actions implemented and reported  |
| **🔍 Reviewer**    | `/reviewer-audit`        | Audit report (APPROVED / NEEDS FIX)            | Report delivered to USER              |
| **🧪 Tester**      | `/tester-verification`   | Test files + coverage report                   | Coverage ≥ 70% achieved and reported  |
| **🧭 Router**      | `/smart-route`           | Confirmation + routed workflow                 | USER confirms routing decision        |

> Any role that hits a blocker **stops and asks the USER** — never self-escalates.

---

## ⛔ CORE COMMUNICATION & PROTOCOLS

1. **Role Anchoring**: ALWAYS prefix every response with the role tag: `[Role: 📝 Spec Writer]` / `[Role: 🏗️ Planner]` /
   `[Role: 💻 Coder]` / `[Role: 🔍 Reviewer]` / `[Role: 🧪 Tester]`
2. **Strict Output Contract**: Each role delivers its defined output then **stops**. It does not initiate the next role.
3. **Explicit Resource Ownership**:
   - **Planner**: Owns `design/design-*.md`, task plan, `git commit`, `complete_task_step`.
   - **Coder**: Owns source code changes and implementation report.
   - **Reviewer**: Owns the audit report.
   - **Tester**: Owns the test suite and coverage report.
4. **Skill Transparency**: Explicitly state which specialized skills are used BEFORE executing tasks.

---

## 🛡️ UNIVERSAL GUARDRAILS (Applies to ALL Roles)

### A. Drift Detection

If any role is stuck on the **same issue 3 times consecutively**:

1. Call `@mcp:context-manager` (`record_failure`).
2. **Stop immediately** — do not attempt a 4th fix.
3. Report the blocker to the USER: what was attempted, what failed, what is needed to unblock.

> ⚠️ No role self-escalates or switches role on drift. The USER decides how to proceed.

### B. Ghost Context (Before Stopping)

Before any role finishes, use `@mcp:context-manager` (`annotate_file`) to inject non-obvious gotchas, architectural
quirks, or library limitations into affected files.

### C. Quality Gates (Planner-Enforced)

Both Reviewer (APPROVED) and Tester (≥ 70% coverage) gates must pass before any commit. If either fails, the Planner
asks the USER — it does not auto-loop. See [`planner-architect.md`](./../workflows/planner-architect.md) **Phase 5**.

---

## 📌 Metadata

- **Version**: 2.4.0
- **Last Updated**: 2026-03-15
