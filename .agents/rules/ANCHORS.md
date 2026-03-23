---
trigger: always_on
---

# ⚓ ANCHORS (Immutable Facts & Guardrails)

> **Version**: 1.0.0 | **Last Amended**: 2026-03-24
>
> **Amendment Process**: Changes to ANCHORS require:
> 1. Explicit USER request with rationale
> 2. Document what changed and why
> 3. Increment version (MAJOR: new constraint, MINOR: refinement, PATCH: wording)

Non-negotiable project facts and constraints. Must survive context compaction and session restarts.

## 🏗️ Technology Stack

> **⚠️ Research-First Rule**: Before selecting or pinning any technology version, the agent MUST use
> `@mcp:doc-researcher` (`search_latest_syntax`) or `@mcp:context7` (`resolve-library-id` → `query-docs`) to discover
> the **latest stable version** of each stack component. Always start new projects on the **latest stable release**.
> Never rely on memorized or training-data versions — they may be outdated.

## 🧪 Quality Standards

- **TDD**: Every task Action MUST have a Verification Command. Code is NOT complete without passing verification.
- **Coverage**: ≥ 70% test coverage before `complete_task_step`. No exceptions.
- **Language**: All source code, variables, functions, comments, and commit messages MUST be in English.

## 🔍 Codebase Intelligence

- **Index Before Search**: ALWAYS call `index_codebase` before `search_code` on a new project.
- **Semantic First**: Prefer `@mcp:codebase-explorer` over `grep_search` for conceptual queries. Use `grep_search` only
  for exact string matches.

## 🛑 Restricted Commands (Require USER Confirmation)

Set `SafeToAutoRun: false` for ALL of the following:

- **Destructive**: `rm -rf`, `rm -f`, `shred`, `mv` (overwriting critical files)
- **Git**: `git push -f`, `git reset --hard`, `git clean -fd`, `git rebase` (shared branches), `git branch -D`
- **Database**: SQL `DROP`/`TRUNCATE`/`DELETE`, `terraform destroy`/`apply`
- **Cloud**: `aws * delete-*`/`terminate-*`, `kubectl delete`, `docker system prune`/`docker rm -f`
- **System**: `chmod` (recursive), `chown`, `chgrp`, `sudo`
- **Publishing**: `npm publish`, `docker push`, `vercel deploy --prod`
- **Process**: `kill`, `killall`, `pkill`, `systemctl stop`/`restart`
