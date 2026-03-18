---
trigger: always_on
---

# ⚓ ANCHORS (Immutable Facts & Guardrails)

This file contains the absolute, non-negotiable project facts, architectural constraints, and guardrails for the agent
system. These **Anchors** must survive context compaction and session restarts.

As a Self-Executing Agent, you MUST respect these facts before executing any code changes or planning any project
tactics.

## 🏗️ Technology Stack

> **⚠️ Research-First Rule**: Before selecting or pinning any technology version, the agent MUST use
> `@mcp:doc-researcher` (`search_latest_syntax`) or `@mcp:context7` (`resolve-library-id` → `query-docs`) to discover
> the **latest stable version** of each stack component. Always start new projects on the **latest stable release**.
> Never rely on memorized or training-data versions — they may be outdated.

- **Backend Stack**: Golang (latest stable). Verify at [go.dev/dl](https://go.dev/dl/).
- **Frontend Stack**: React (latest stable) / Next.js (latest stable) App Router. Verify at
  [react.dev/versions](https://react.dev/versions) and [nextjs.org/docs](https://nextjs.org/docs).
- **Package Manager**: Use `pnpm` (latest stable) instead of `npm` for better performance.
- **Database**: PostgreSQL (latest stable major). DO NOT use MongoDB or MySQL. Verify at
  [postgresql.org/support/versioning](https://www.postgresql.org/support/versioning/).

## 🛡️ Agentic Guardrails & Execution Constraints

- **State Boundaries**: `active_files` MUST be explicitly tracked across role transitions and locked via Intent
  Declaration to prevent blind writes and scope creep.
- **No Destruction**: Do not delete existing API contracts, database columns, or core functionality without explicit
  confirmation from the USER.
- **Panic Protocol**: If you fail at the same step 3 times (e.g., test failures looping between Coder and Tester), you
  MUST STOP, record failure (`record_failure`), and drop back to the Planner role to discuss with the USER instead of
  blindly hallucinating fixes.
- **No Assumptions**: If Requirements are vague, stop and clarify via Socratic questioning before writing any code.

## 🧪 Quality & Verification

- **TDD Requirement**: Every action in a Task Plan MUST include a clear Verification Command as acceptance criteria.
  Code is not complete until verification passes.
- **Language**: All source code, variables, functions, comments, and commit messages MUST be in English.
- **Coverage**: You must ensure high quality and stability (aim for >= 70% test coverage) before calling
  `complete_task_step`.

## 🧠 Memory Governance

- **Session Compaction**: At the end of a major Tactic/Phase, run the `compact_session` workflow to distill
  architectural decisions and lessons learned into long-term Knowledge Items (KIs).
- **Ghost Context**: When encountering complex file-specific logic or tricky quirks, leverage `annotate_file` to attach
  localized lessons directly to the file to prevent recurring mistakes.

## 🔍 Codebase Intelligence

- **Index Before Search**: Before using `@mcp:codebase-explorer` (`search_code`) on a new project, call `index_codebase`
  first. The index is persistent (SQLite) — subsequent sessions can search immediately.
- **Semantic Precedence**: Prefer `@mcp:codebase-explorer` over raw `grep_search` for finding conceptually related code
  (e.g., "authentication middleware", "retry logic"). Use `grep_search` only for exact string matches.

## 🛑 Restricted Terminal Commands

To prevent destructive actions, the agent MUST set `SafeToAutoRun: false` and explicitly ask the USER for confirmation
before executing any of the following commands:

- **File & Directory Destruction**: `rm` (especially `rm -rf` or `rm -f`), `mv` (if overwriting critical files), `shred`
- **Destructive Git Operations**: `git push -f`, `git reset --hard`, `git clean -fd`, `git rebase` (on shared branches),
  `git branch -D`
- **Database & Infrastructure Alterations**: SQL `drop`/`truncate`/`delete`, `terraform destroy`/`apply`,
  `aws * delete-*`/`terminate-*`, `kubectl delete`, `docker system prune`/`docker rm -f`
- **System Permissions & Security**: `chmod` (e.g., recursive changes), `chown`, `chgrp`, `sudo`
- **Publishing & Deployments**: `npm publish`, `docker push`, package registry uploads, direct production deployments
  (`vercel deploy --prod`)
- **Process Management**: `kill`, `killall`, `pkill`, `systemctl stop`/`restart`
