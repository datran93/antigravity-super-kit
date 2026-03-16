# Antigravity Kit

> A production-grade AI agent framework — Go MCP servers, 880+ skills, structured role workflows, and hybrid semantic
> search.

```
[Spec Writer] → [Planner] → [Coder] → [Reviewer] → [Tester] → [Commit]
                                  ↑                              ↓
                              /smart-route ←←←←←←←←←←←←←←←←←←←←
```

---

## What's Inside

| Layer                 | Components                                                 | Count      |
| --------------------- | ---------------------------------------------------------- | ---------- |
| **MCP Servers** (Go)  | Database, AST, Search, Governance, Context, Skills…        | 11 servers |
| **Skills**            | Domain-specific knowledge modules                          | 880+       |
| **Workflows**         | Slash-command procedures                                   | 17         |
| **Role Architecture** | Spec Writer → Planner → Coder → Reviewer → Tester → Router | 6 roles    |

---

## Quick Install

```bash
# 1. Clone
git clone git@github.com:Dang-Hai-Tran/antigravity-kit.git ~/.antigravity/cache/antigravity-kit

# 2. Link CLI
sudo ln -s ~/.antigravity/cache/antigravity-kit/scripts/agk /usr/local/bin/agk

# 3. Install .agents into any project
cd my-project && agk install
```

### CLI Commands

| Command                | Description                                        |
| ---------------------- | -------------------------------------------------- |
| `agk install`          | Install `.agents/` folder (auto-excludes from git) |
| `agk install --force`  | Overwrite existing installation                    |
| `agk update`           | Pull latest version                                |
| `agk update --dry-run` | Preview changes                                    |
| `agk update --offline` | Update without network                             |
| `agk status`           | Check for available updates                        |
| `agk remove`           | Remove `.agents/` (with confirmation)              |
| `agk version`          | Show version                                       |

> **Cursor / Windsurf users**: Do NOT add `.agents/` to `.gitignore` — it prevents slash command indexing. Use
> `.git/info/exclude` instead.

---

## MCP Server Stack

All servers are written in **Go** and communicate via stdio (MCP protocol). Registered in
`~/.gemini/antigravity/mcp_config.json`.

### Core Intelligence

| Server             | Command                   | Key Tools                                                                                          |
| ------------------ | ------------------------- | -------------------------------------------------------------------------------------------------- |
| `skill-router`     | `mcp-skill-router-go`     | `search_skills` — semantic skill search with section-level Merkle indexing                         |
| `context-manager`  | `mcp-context-manager-go`  | `recall_knowledge` (hybrid BM25+vector RRF), `save_checkpoint`, `declare_intent`, `compact_memory` |
| `codebase-search`  | `mcp-codebase-search-go`  | `index_codebase`, `search_code` (hybrid RRF), `get_indexing_status`, `clear_index`                 |
| `context-governor` | `mcp-context-governor-go` | `get_budget_status`, `estimate_cost`, `suggest_compression`, `trigger_compact`                     |

### Tooling & Integrations

| Server               | Key Tools                                                                                                                                            |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ast-explorer`       | `get_project_architecture`, `search_symbol`                                                                                                          |
| `database-inspector` | `list_tables`, `inspect_schema`, `run_read_query`, `explain_query`                                                                                   |
| `doc-researcher`     | `search_latest_syntax`, `read_website_markdown`, `read_doc_file`                                                                                     |
| `figma-reader`       | `read_figma_design`, `export_figma_images`                                                                                                           |
| `gitlab`             | `get_file_content`, `list_directory`, `get_repository_info`, `search_code`, `read_mr_discussions`, `reply_to_mr_discussion`, `resolve_mr_discussion` |
| `mcp-http-client`    | `http_request`, `import_curl`, `set_env`                                                                                                             |
| `confluence`         | `search_pages`, `get_page`, `create_page`, `update_page`                                                                                             |

### Hybrid Search Architecture

```
Query
  │
  ├─► BM25 (FTS5 SQLite)  ──────────────────────┐
  │                                              ▼
  └─► Vector (OpenAI text-embedding-3-small) → RRF Fusion → Top-K results
         ↑ graceful fallback if no API key
```

- **`mcp-context-manager-go`**: Hybrid KI recall — BM25 + cosine similarity fused via RRF (k=60)
- **`mcp-skill-router-go`**: Section-level Merkle diff — only re-embeds changed sections on disk
- **`mcp-codebase-search-go`**: AST-aware chunking (Go) + line-window fallback + incremental indexing

---

## Role Architecture

The system uses **6 distinct, non-overlapping roles**. Each produces one defined output then stops.

| Role               | Slash Command            | Responsibility                                | Output                              |
| ------------------ | ------------------------ | --------------------------------------------- | ----------------------------------- |
| 📝 **Spec Writer** | `/specifications-writer` | Socratic requirements engineering             | `SPEC.md`                           |
| 🏗️ **Planner**     | `/planner-architect`     | Architecture + ordered task list + commits    | `DESIGN.md` + task plan             |
| 💻 **Coder**       | `/coder-implementation`  | Execute tasks from design                     | Code changes + report               |
| 🔍 **Reviewer**    | `/reviewer-audit`        | Audit code quality and correctness            | Audit report (APPROVED / NEEDS FIX) |
| 🧪 **Tester**      | `/tester-verification`   | Write tests, enforce ≥ 70% coverage           | Test files + coverage report        |
| 🧭 **Router**      | `/smart-route`           | Classify intent and route to correct workflow | Confirmed routing                   |

### All Workflows

| Command                  | Description                                         |
| ------------------------ | --------------------------------------------------- |
| `/specifications-writer` | Socratic requirements interview → SPEC.md           |
| `/planner-architect`     | Design architecture → DESIGN.md + ordered task plan |
| `/coder-implementation`  | Implement from task plan, report changes            |
| `/reviewer-audit`        | Audit code against DESIGN.md, report findings       |
| `/tester-verification`   | Write unit + integration tests, enforce coverage    |
| `/smart-route`           | Auto-classify intent and route to correct workflow  |
| `/compact-session`       | Flush context → Knowledge Item (KI) generation      |
| `/brownfield-scan`       | Analyze and orchestrate legacy codebases            |
| `/war-room`              | Emergency incident response with evidence gates     |
| `/tdd-autopilot`         | Automated TDD red-green-refactor loop               |
| `/git-commit`            | Stage, commit, rebase against develop               |
| `/git-push`              | Pull rebase, resolve conflicts, push                |
| `/gitlab-mr-read-fix`    | Read MR discussions, apply fixes locally            |
| `/gitlab-mr-reply-push`  | Push fixes, reply to and resolve MR threads         |
| `/codemap`               | Generate hierarchical codebase visualizations       |
| `/deepwiki`              | Generate interactive repository wikis               |
| `/terminal`              | Run a command in terminal                           |

---

## Context Governance

The **Context Governor** (`mcp-context-governor-go`) tracks token budget automatically:

| Level      | Usage  | Action                                        |
| ---------- | ------ | --------------------------------------------- |
| `ok`       | < 60%  | Continue normally                             |
| `warning`  | 60–80% | Consider compacting soon                      |
| `critical` | 80–95% | Run `/compact-session` before next large task |
| `overflow` | > 95%  | **Stop immediately**, compact, then continue  |

```
# Check current budget
@mcp:context-governor get_budget_status session_id="my-session"

# Trigger compaction (resets counter + prompts compact_memory)
@mcp:context-governor trigger_compact session_id="my-session" reason="tactic complete"
```

---

## Codebase Search

Index any project for hybrid semantic + keyword search:

```
# Index the project (background job, persistent SQLite)
@mcp:codebase-search index_codebase path="/path/to/project"

# Search with natural language
@mcp:codebase-search search_code query="authentication middleware" project_path="/path/to/project"

# Filter by language
@mcp:codebase-search search_code query="retry logic" project_path="." lang_filter="go"

# Check progress
@mcp:codebase-search get_indexing_status project_path="."
```

**Requires**: `OPENAI_API_KEY` for vector embeddings. Falls back to BM25-only if unset.

---

## Skills

880+ domain-specific knowledge modules loaded automatically based on task context via `search_skills`.

**Example domains**: Go, TypeScript, React, FastAPI, PostgreSQL, Kubernetes, Terraform, AI/LLM, Security, AWS, Docker,
Agent Orchestration, and many more.

Skills are stored in `.agents/skills/` and can be invoked explicitly or discovered automatically:

```
@mcp:skill-router search_skills query="Go concurrency patterns" tags_filter="golang"
```

---

## Configuration

MCP servers are configured in `~/.gemini/antigravity/mcp_config.json`. Example entry:

```json
{
  "mcpServers": {
    "codebase-search": {
      "command": "/path/to/antigravity-kit/tools/mcp-codebase-search-go/mcp-codebase-search-go",
      "args": [],
      "env": {
        "OPENAI_API_KEY": "sk-..."
      }
    },
    "context-governor": {
      "command": "/path/to/antigravity-kit/tools/mcp-context-governor-go/mcp-context-governor-go",
      "args": []
    }
  }
}
```

### Build Requirements

```bash
# All Go servers require CGO + sqlite_fts5 tag for FTS5 support
CGO_ENABLED=1 go build -tags sqlite_fts5 -o ./binary ./...

# Run tests
CGO_ENABLED=1 go test -tags sqlite_fts5 ./...
```

---

## Project Structure

```
antigravity-kit/
├── .agents/
│   ├── rules/
│   │   ├── GEMINI.md          # Universal agent rules (v2.1.0)
│   │   └── ANCHORS.md         # Immutable guardrails
│   ├── skills/                # 880+ domain skills
│   └── workflows/             # 17 slash-command workflows
├── tools/
│   ├── mcp-context-manager-go/   # KI recall (hybrid BM25+vec)
│   ├── mcp-skill-router-go/      # Skill search (section Merkle)
│   ├── mcp-codebase-search-go/   # AST + hybrid code search
│   ├── mcp-context-governor-go/  # Token budget governance
│   ├── mcp-ast-explorer-go/      # AST structural analysis
│   ├── mcp-database-inspector-go/
│   ├── mcp-doc-researcher-go/
│   ├── mcp-figma-reader-go/
│   ├── mcp-gitlab-go/            # GitLab: file reader + MR discussions (unified)
│   ├── mcp-http-client-go/
│   └── mcp-confluence-go/
├── DESIGN.md                  # Current sprint architecture
└── scripts/
    └── agk                    # CLI for managing .agents/
```

---

## License

MIT © datran
