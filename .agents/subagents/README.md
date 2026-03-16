# Subagent Engine — Antigravity Kit

The `spawn-subagent.sh` engine enables the Antigravity Kit to spawn **fresh, isolated Gemini CLI subprocesses** per task
— equivalent to Claude Code's `Task` tool.

## Quick Start

```bash
# 1. Write a task file
cat > /tmp/my-task.md << 'EOF'
## Task: Add input validation to the login handler

Validate that email is non-empty and password is >= 8 chars.
Return 400 with descriptive error if validation fails.

### Files
- `internal/handlers/auth.go`
- `internal/handlers/auth_test.go`

### Verification
`go test ./internal/handlers/... -run TestLogin`
EOF

# 2. Spawn implementer subagent
.agents/subagents/spawn-subagent.sh \
  --role implementer \
  --task-file /tmp/my-task.md \
  --context-files "internal/handlers/auth.go" \
  --workspace "$(pwd)"

# 3. Check exit code
echo "Exit: $?"   # 0=done, 1=blocked/needs-context, 2=timeout, 3=error
```

## Roles & Defaults

| Role                     | Model                    | Timeout | When to use                   |
| ------------------------ | ------------------------ | ------- | ----------------------------- |
| `implementer`            | `gemini-3.1-pro-preview` | 10 min  | Write/modify code             |
| `spec-reviewer`          | `gemini-3-flash-preview` | 5 min   | Verify code matches spec      |
| `quality-reviewer`       | `gemini-3-flash-preview` | 5 min   | Code quality / best practices |
| `spec-document-reviewer` | `gemini-3-flash-preview` | 5 min   | Validate spec doc itself      |

## STATUS Protocol

Every subagent MUST end its output with one of:

```
STATUS: DONE
STATUS: DONE_WITH_CONCERNS — <brief explanation>
STATUS: NEEDS_CONTEXT — <what is missing>
STATUS: BLOCKED — <blocker description>
```

`spawn-subagent.sh` parses this line and maps it to an exit code:

- `0` → DONE / DONE_WITH_CONCERNS
- `1` → NEEDS_CONTEXT / BLOCKED
- `2` → TIMEOUT
- `3` → PROCESS ERROR

## Full Workflow (per T-step)

```bash
#!/usr/bin/env bash
# Orchestrator loop for one T-step
TASK_FILE="/tmp/task-T3.md"
WORKSPACE="/path/to/project"
MAX_RETRIES=3

for attempt in $(seq 1 $MAX_RETRIES); do
  .agents/subagents/spawn-subagent.sh \
    --role implementer \
    --task-file "$TASK_FILE" \
    --workspace "$WORKSPACE" \
    --output-file /tmp/impl-result.md

  EXIT=$?

  if [[ $EXIT -eq 0 ]]; then
    echo "✅ Implementation done — running spec review..."
    break
  elif [[ $EXIT -eq 1 ]]; then
    echo "⚠️  Attempt $attempt failed. Check /tmp/impl-result.md for details."
    if [[ $attempt -eq $MAX_RETRIES ]]; then
      echo "🚨 Max retries reached — escalate to USER"
      exit 1
    fi
  else
    echo "💥 Fatal error (exit $EXIT) — escalate to USER"
    exit $EXIT
  fi
done

# Spec review
.agents/subagents/spawn-subagent.sh \
  --role spec-reviewer \
  --task-file "$TASK_FILE" \
  --workspace "$WORKSPACE" \
  --output-file /tmp/spec-review.md

# Quality review
.agents/subagents/spawn-subagent.sh \
  --role quality-reviewer \
  --task-file "$TASK_FILE" \
  --workspace "$WORKSPACE" \
  --output-file /tmp/quality-review.md
```

## Options Reference

```
--role ROLE            implementer|spec-reviewer|quality-reviewer|spec-document-reviewer
--task-file PATH       Markdown file with task (required)
--context-files LIST   Comma-separated files to embed as context
--workspace PATH       Project root (default: $PWD)
--model MODEL          Override default model (e.g. gemini-3.1-pro-preview)
--output-file PATH     Write structured result markdown here
--timeout SECS         Kill subagent after SECS seconds
--verbose              Stream subagent output in real-time
```

## Tips

1. **Keep task files small** — one clear objective per task
2. **List only directly relevant files** in `--context-files` (not the whole codebase)
3. **Use `--verbose`** during development to see subagent reasoning live
4. **Check `--output-file`** after a BLOCKED/NEEDS_CONTEXT for the full explanation
5. **Templates are in** `.agents/subagents/templates/` — edit them to tune agent behavior
