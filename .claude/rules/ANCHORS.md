---
trigger: always_on
---

# ⚓ ANCHORS (Immutable Facts & Guardrails)

> **Version**: 2.0.0 | **Last Amended**: 2026-04-20
>
> **Amendment Process**: Changes to ANCHORS require:
> 1. Explicit USER request with rationale
> 2. Document what changed and why
> 3. Increment version (MAJOR: new constraint, MINOR: refinement, PATCH: wording)

Non-negotiable project facts and constraints. Must survive context compaction and session restarts.

> **💡 Context Pruning**: Each section is tagged with `[domain:X]`. Agents operating on specific task types
> ONLY need to load the matching domain(s) via `manage_anchors`. Loading everything is wasteful.
>
> | Task Type | Load Domains |
> |-----------|-------------|
> | UI/frontend | `stack`, `quality`, `codebase` |
> | DB / migration | `db`, `quality`, `git`, `security` |
> | Auth / payment | `security`, `quality`, `git` |
> | API / integration | `stack`, `security`, `quality`, `git` |
> | Code refactor | `codebase`, `quality`, `git` |
> | Any task | `quality` (always required) |

---

## 🏗️ Technology Stack `[domain:stack]`

> **⚠️ Research-First Rule**: Before selecting or pinning any technology version, the agent MUST use
> `@mcp:doc-researcher` (`search_latest_syntax`) or `@mcp:context7` (`resolve-library-id` → `query-docs`) to discover
> the **latest stable version** of each stack component. Always start new projects on the **latest stable release**.
> Never rely on memorized or training-data versions — they may be outdated.

---

## 🧪 Quality Standards `[domain:quality]`

- **TDD**: Every task Action MUST have a Verification Command. Code is NOT complete without passing verification.
- **Coverage**: ≥ 70% test coverage before `complete_task_step`. No exceptions.
- **Language**: All source code, variables, functions, comments, and commit messages MUST be in English.

---

## 🔍 Codebase Intelligence `[domain:codebase]`

- **Index Before Search**: ALWAYS call `index_codebase` before `search_code` on a new project.
- **Semantic First**: Prefer `@mcp:codebase-explorer` over `grep_search` for conceptual queries. Use `grep_search` only
  for exact string matches.
- **AST First for Shared Code**: Before modifying any symbol used in more than one file, call `find_usages` or `context`
  to map the blast radius. NEVER edit shared code without first verifying all callers.
- **Symbol Lookup Before grep**: Use `search_symbol` to find class/function definitions. Only fall back to `grep_search`
  for exact literal string matches.

---

## 🔐 Security `[domain:security]`

- **No Secrets in Code**: API keys, passwords, and tokens MUST use environment variables or secret managers. NEVER
  hardcode secrets. NEVER commit `.mcp.json`, `.env`, or credential files.
- **Auth changes are HIGH-RISK**: Any modification to authentication, authorization, or session logic requires explicit
  USER confirmation before execution and a Reviewer APPROVED gate.
- **Input Validation**: All external inputs (HTTP, CLI args, file reads) MUST be validated and sanitized before use.

---

## 🗃️ Database `[domain:db]`

- **Migration Required**: Any schema change (ADD COLUMN, DROP COLUMN, CREATE TABLE, ALTER TABLE) MUST include:
  - A forward migration script
  - A rollback script
  - Backward compatibility analysis
- **No Blind Drops**: SQL `DROP`/`TRUNCATE`/`DELETE` require explicit USER confirmation. Set `SafeToAutoRun: false`.
- **Zero-Downtime Default**: Prefer additive schema changes. Never remove columns until a deprecation cycle completes.

---

## 🔀 Git `[domain:git]`

- **Commit Often**: Each completed task action = one commit. Commit messages follow Conventional Commits:
  `type(scope): description`.
- **Never Force-Push Main**: `git push -f` to `main`/`master` is forbidden without explicit USER instruction.
- **No Skip Hooks**: Never use `--no-verify` or `--no-gpg-sign` unless USER explicitly requests it.

---

## 🛑 Restricted Commands (Require USER Confirmation) `[domain:security]`

Set `SafeToAutoRun: false` for ALL of the following:

- **Destructive**: `rm -rf`, `rm -f`, `shred`, `mv` (overwriting critical files)
- **Git**: `git push -f`, `git reset --hard`, `git clean -fd`, `git rebase` (shared branches), `git branch -D`
- **Database**: SQL `DROP`/`TRUNCATE`/`DELETE`, `terraform destroy`/`apply`
- **Cloud**: `aws * delete-*`/`terminate-*`, `kubectl delete`, `docker system prune`/`docker rm -f`
- **System**: `chmod` (recursive), `chown`, `chgrp`, `sudo`
- **Publishing**: `npm publish`, `docker push`, `vercel deploy --prod`
- **Process**: `kill`, `killall`, `pkill`, `systemctl stop`/`restart`
