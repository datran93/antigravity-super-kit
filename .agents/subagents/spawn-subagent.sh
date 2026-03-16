#!/usr/bin/env bash
# =============================================================================
# spawn-subagent.sh — Antigravity Kit Subagent Engine
# =============================================================================
# Spawns a fresh, isolated Gemini CLI subprocess per task.
# Inspired by Superpowers' subagent-driven-development pattern.
#
# Usage:
#   spawn-subagent.sh --role <role> --task-file <path> [OPTIONS]
#
# Exit codes:
#   0 = DONE or DONE_WITH_CONCERNS
#   1 = NEEDS_CONTEXT or BLOCKED
#   2 = TIMEOUT
#   3 = PROCESS ERROR / bad args
# =============================================================================
set -euo pipefail

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------
ROLE=""
TASK_FILE=""
CONTEXT_FILES=""
WORKSPACE="${PWD}"
MODEL=""
OUTPUT_FILE=""
TIMEOUT=""
VERBOSE=false
AGENTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATES_DIR="${AGENTS_DIR}/templates"

# Role → default model + timeout mapping (bash 3.x compatible)
role_model() {
  case "$1" in
    implementer)            echo "gemini-3.1-pro-preview"   ;;
    spec-reviewer)          echo "gemini-3-flash-preview"   ;;
    quality-reviewer)       echo "gemini-3-flash-preview"   ;;
    spec-document-reviewer) echo "gemini-3-flash-preview"   ;;
    *) echo "" ;;
  esac
}
role_timeout() {
  case "$1" in
    implementer)            echo "600" ;;
    spec-reviewer)          echo "300" ;;
    quality-reviewer)       echo "300" ;;
    spec-document-reviewer) echo "300" ;;
    *) echo "" ;;
  esac
}
valid_roles="implementer spec-reviewer quality-reviewer spec-document-reviewer"

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------
usage() {
  cat <<EOF
Usage: spawn-subagent.sh [OPTIONS]

Spawn a fresh, isolated Gemini CLI subagent for a specific task.

Required:
  --role ROLE           Agent role: implementer | spec-reviewer |
                        quality-reviewer | spec-document-reviewer
  --task-file PATH      Path to markdown file with task description
                        (use template placeholders or pre-rendered content)

Optional:
  --context-files LIST  Comma-separated list of files to include as context
                        (e.g. "src/main.go,src/main_test.go")
  --workspace PATH      Project root directory (default: \$PWD)
  --model MODEL         Gemini model override (e.g. gemini-2.5-pro)
  --output-file PATH    Write structured result here (default: /tmp/subagent-<role>-<ts>.md)
  --timeout SECS        Max seconds before killing subagent (default: role-based)
  --verbose             Print subagent stdout in real-time
  -h, --help            Show this help

Role defaults:
  implementer            model=gemini-3.1-pro-preview   timeout=600s
  spec-reviewer          model=gemini-3-flash-preview   timeout=300s
  quality-reviewer       model=gemini-3-flash-preview   timeout=300s
  spec-document-reviewer model=gemini-3-flash-preview   timeout=300s

Exit codes:
  0   DONE or DONE_WITH_CONCERNS
  1   NEEDS_CONTEXT or BLOCKED
  2   TIMEOUT
  3   PROCESS ERROR / invalid args

STATUS Protocol:
  Subagent MUST end its final output with one of:
    STATUS: DONE
    STATUS: DONE_WITH_CONCERNS — <explanation>
    STATUS: NEEDS_CONTEXT — <what is missing>
    STATUS: BLOCKED — <blocker>

Examples:
  # Dispatch implementer (auto selects flash model, 5min timeout)
  spawn-subagent.sh \\
    --role implementer \\
    --task-file /tmp/task-auth.md \\
    --context-files "internal/auth/jwt.go,internal/auth/jwt_test.go" \\
    --workspace /path/to/project

  # Dispatch spec reviewer with output capture
  spawn-subagent.sh \\
    --role spec-reviewer \\
    --task-file /tmp/task-auth.md \\
    --context-files "internal/auth/jwt.go" \\
    --output-file /tmp/spec-review.md
EOF
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --role)            ROLE="$2";           shift 2 ;;
    --task-file)       TASK_FILE="$2";      shift 2 ;;
    --context-files)   CONTEXT_FILES="$2";  shift 2 ;;
    --workspace)       WORKSPACE="$2";      shift 2 ;;
    --model)           MODEL="$2";          shift 2 ;;
    --output-file)     OUTPUT_FILE="$2";    shift 2 ;;
    --timeout)         TIMEOUT="$2";        shift 2 ;;
    --verbose)         VERBOSE=true;        shift   ;;
    -h|--help)         usage; exit 0 ;;
    *) echo "❌ Unknown argument: $1" >&2; usage; exit 3 ;;
  esac
done

# ---------------------------------------------------------------------------
# Validation
# ---------------------------------------------------------------------------
if [[ -z "$ROLE" ]]; then
  echo "❌ --role is required" >&2; exit 3
fi
if [[ -z "$(role_model "$ROLE")" ]]; then
  echo "❌ Unknown role: '$ROLE'. Valid roles: ${valid_roles}" >&2; exit 3
fi
if [[ -z "$TASK_FILE" ]]; then
  echo "❌ --task-file is required" >&2; exit 3
fi
if [[ ! -f "$TASK_FILE" ]]; then
  echo "❌ Task file not found: $TASK_FILE" >&2; exit 3
fi
if [[ ! -d "$WORKSPACE" ]]; then
  echo "❌ Workspace directory not found: $WORKSPACE" >&2; exit 3
fi
TEMPLATE_FILE="${TEMPLATES_DIR}/${ROLE}.md"
if [[ ! -f "$TEMPLATE_FILE" ]]; then
  echo "❌ Template not found: $TEMPLATE_FILE" >&2; exit 3
fi

# Apply defaults
[[ -z "$MODEL"   ]] && MODEL="$(role_model "$ROLE")"
[[ -z "$TIMEOUT" ]] && TIMEOUT="$(role_timeout "$ROLE")"

TS="$(date +%Y%m%d_%H%M%S)"
[[ -z "$OUTPUT_FILE" ]] && OUTPUT_FILE="/tmp/subagent-${ROLE}-${TS}.md"

# ---------------------------------------------------------------------------
# Build the subagent prompt
# ---------------------------------------------------------------------------
TASK_CONTENT="$(cat "$TASK_FILE")"

# Build context block from context files
CONTEXT_BLOCK=""
if [[ -n "$CONTEXT_FILES" ]]; then
  CONTEXT_BLOCK=$'\n\n---\n\n## Context Files\n\nThe following files are relevant to your task:\n\n'
  IFS=',' read -ra FILES <<< "$CONTEXT_FILES"
  for f in "${FILES[@]}"; do
    f="${f## }"; f="${f%% }"   # trim whitespace
    ABS_FILE="${WORKSPACE}/${f}"
    if [[ -f "$ABS_FILE" ]]; then
      CONTEXT_BLOCK+="### \`${f}\`\n\`\`\`\n$(cat "$ABS_FILE")\n\`\`\`\n\n"
    else
      CONTEXT_BLOCK+="### \`${f}\`\n*(file not found — create it if required)*\n\n"
    fi
  done
fi

# Compose full prompt: role template + task + context
ROLE_INSTRUCTIONS="$(cat "$TEMPLATE_FILE")"
FULL_PROMPT="${ROLE_INSTRUCTIONS}

---

## Your Task

${TASK_CONTENT}
${CONTEXT_BLOCK}
---

**Working directory**: \`${WORKSPACE}\`

> Your final line of output MUST be one of:
> - \`STATUS: DONE\`
> - \`STATUS: DONE_WITH_CONCERNS — <brief explanation>\`
> - \`STATUS: NEEDS_CONTEXT — <what is missing>\`
> - \`STATUS: BLOCKED — <blocker description>\`
"


# ---------------------------------------------------------------------------
# Spawn subagent
# ---------------------------------------------------------------------------
SUBAGENT_TMP_STDOUT="/tmp/subagent-stdout-${TS}.txt"

echo "🤖 Spawning subagent [role=${ROLE} model=${MODEL} timeout=${TIMEOUT}s]" >&2
echo "📄 Task: ${TASK_FILE}" >&2
[[ -n "$CONTEXT_FILES" ]] && echo "📂 Context: ${CONTEXT_FILES}" >&2
echo "---" >&2

# gemini CLI: pipe full prompt via stdin only (--prompt conflicts with stdin)
# Use --prompt with the actual content so no positional arg collision occurs.
# Per gemini docs: "-p/--prompt: Run in non-interactive mode. Appended to stdin."
# Strategy: pass everything via --prompt flag (no stdin piping needed).
GEMINI_CMD=(
  gemini
  --model         "$MODEL"
  --prompt        "$FULL_PROMPT"
  --approval-mode yolo
  --output-format text
)

SPAWN_START="$(date +%s)"

# Run with timeout using `timeout` command (GNU coreutils / macOS gnutimeout)
# macOS ships `timeout` via `brew install coreutils` as `gtimeout`; fall back gracefully.
TIMEOUT_CMD="timeout"
if ! command -v timeout &>/dev/null; then
  if command -v gtimeout &>/dev/null; then
    TIMEOUT_CMD="gtimeout"
  else
    TIMEOUT_CMD=""
  fi
fi

set +e   # allow non-zero exit from subagent

if [[ -n "$TIMEOUT_CMD" ]]; then
  if [[ "$VERBOSE" == "true" ]]; then
    "$TIMEOUT_CMD" "${TIMEOUT}" "${GEMINI_CMD[@]}" 2>&1 | tee "$SUBAGENT_TMP_STDOUT"
    GEMINI_EXIT="${PIPESTATUS[0]}"
  else
    "$TIMEOUT_CMD" "${TIMEOUT}" "${GEMINI_CMD[@]}" > "$SUBAGENT_TMP_STDOUT" 2>&1
    GEMINI_EXIT="$?"
  fi
else
  echo "⚠️  'timeout' command not found. Running without timeout guard." >&2
  if [[ "$VERBOSE" == "true" ]]; then
    "${GEMINI_CMD[@]}" 2>&1 | tee "$SUBAGENT_TMP_STDOUT"
    GEMINI_EXIT="${PIPESTATUS[0]}"
  else
    "${GEMINI_CMD[@]}" > "$SUBAGENT_TMP_STDOUT" 2>&1
    GEMINI_EXIT="$?"
  fi
fi

set -e

SPAWN_END="$(date +%s)"
ELAPSED=$(( SPAWN_END - SPAWN_START ))

# ---------------------------------------------------------------------------
# Handle timeout
# ---------------------------------------------------------------------------
if [[ "$GEMINI_EXIT" -eq 124 ]]; then
  echo "" >&2
  echo "⏰ TIMEOUT: Subagent exceeded ${TIMEOUT}s after ${ELAPSED}s" >&2
  {
    echo "# Subagent Result — TIMEOUT"
    echo "role: ${ROLE}"
    echo "model: ${MODEL}"
    echo "elapsed: ${ELAPSED}s"
    echo "exit: TIMEOUT"
    echo ""
    echo "## Subagent stdout (partial)"
    cat "$SUBAGENT_TMP_STDOUT" 2>/dev/null || echo "(no output)"
  } > "$OUTPUT_FILE"
  rm -f "$SUBAGENT_TMP_STDOUT"
  exit 2
fi

# ---------------------------------------------------------------------------
# Handle process error
# ---------------------------------------------------------------------------
if [[ "$GEMINI_EXIT" -ne 0 ]]; then
  echo "" >&2
  echo "💥 PROCESS ERROR: gemini exited with code ${GEMINI_EXIT}" >&2
  cat "$SUBAGENT_TMP_STDOUT" >&2
  {
    echo "# Subagent Result — PROCESS ERROR"
    echo "role: ${ROLE}"
    echo "exit_code: ${GEMINI_EXIT}"
    echo "elapsed: ${ELAPSED}s"
    echo ""
    echo "## Raw stdout"
    cat "$SUBAGENT_TMP_STDOUT" 2>/dev/null || echo "(no output)"
  } > "$OUTPUT_FILE"
  rm -f "$SUBAGENT_TMP_STDOUT"
  exit 3
fi

# ---------------------------------------------------------------------------
# Parse STATUS from subagent output
# ---------------------------------------------------------------------------
SUBAGENT_OUTPUT="$(cat "$SUBAGENT_TMP_STDOUT")"

# Find last STATUS: line (case-insensitive search for robustness)
STATUS_LINE="$(echo "$SUBAGENT_OUTPUT" | grep -iE '^STATUS:' | tail -1 || true)"

if [[ -z "$STATUS_LINE" ]]; then
  # No STATUS line found — treat as BLOCKED
  STATUS_LINE="STATUS: BLOCKED — subagent did not emit a STATUS line"
fi

STATUS_CODE="$(echo "$STATUS_LINE" | sed 's/STATUS:[[:space:]]*//' | awk '{print $1}' | tr '[:lower:]' '[:upper:]')"

# ---------------------------------------------------------------------------
# Write structured output file
# ---------------------------------------------------------------------------
{
  echo "# Subagent Result"
  echo ""
  echo "| Field   | Value |"
  echo "|---------|-------|"
  echo "| Role    | ${ROLE} |"
  echo "| Model   | ${MODEL} |"
  echo "| Elapsed | ${ELAPSED}s |"
  echo "| Status  | ${STATUS_LINE} |"
  echo ""
  echo "## Full Output"
  echo ""
  echo "$SUBAGENT_OUTPUT"
} > "$OUTPUT_FILE"

# ---------------------------------------------------------------------------
# Print summary & exit
# ---------------------------------------------------------------------------
echo "" >&2
echo "✅ Subagent completed in ${ELAPSED}s" >&2
echo "📊 ${STATUS_LINE}" >&2
echo "📝 Full output: ${OUTPUT_FILE}" >&2

# Print raw output to stdout (so orchestrator can capture it)
echo "$SUBAGENT_OUTPUT"

# Cleanup temp files
rm -f "$SUBAGENT_TMP_STDOUT"

# Exit code based on status
case "$STATUS_CODE" in
  DONE|DONE_WITH_CONCERNS) exit 0 ;;
  NEEDS_CONTEXT|BLOCKED)   exit 1 ;;
  *)
    echo "⚠️  Unrecognised status code: '${STATUS_CODE}' — treating as BLOCKED" >&2
    exit 1
    ;;
esac
