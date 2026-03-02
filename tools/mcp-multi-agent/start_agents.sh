#!/bin/bash

# Core directory for Multi-Agent model (fixed)
SCRIPT_DIR="/Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent"
WORKER_SCRIPT="$SCRIPT_DIR/worker.py"
PYTHON_ENV="$SCRIPT_DIR/.venv/bin/python"
DB_PATH="$SCRIPT_DIR/multi_agent_bus.db"

# Current workspace
WORKSPACE=$(pwd)
LOG_DIR="$WORKSPACE/.agent_logs"

mkdir -p "$LOG_DIR"

ENGINE="kilocode"
TASK=""

# Parse options
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -e|--engine) ENGINE="$2"; shift ;;
        -m|--model) MODEL="$2"; shift ;;
        *) TASK="$1" ;;
    esac
    shift
done

# Pass model argument if provided
MODEL_ARG=""
if [ -n "$MODEL" ]; then
    MODEL_ARG="--model $MODEL"
fi

if [ -z "$TASK" ]; then
  echo "⚠️ Error: No request provided for Agent."
  echo "💡 Usage: $0 [-e|--engine <engine>] \"<Your Request>\""
  exit 1
fi

echo "🧹 Cleaning up existing Agents and Dashboard in this workspace (if any)..."
pkill -f "worker.py --workspace $WORKSPACE"
pkill -f "dashboard.py"
sleep 1

WORKSPACE_DB="$LOG_DIR/multi_agent_bus.db"
if [ -f "$WORKSPACE_DB" ]; then
    echo "🗑 Resetting Agent memory to start a NEW project..."
    rm "$WORKSPACE_DB"
fi

echo "🚀 Summoning [Planner Agent] for project: $WORKSPACE..."
PLANNER_INST="You are a strict PLANNER and ARCHITECT. You are the leader of a static 4-agent team: planner, coder, reviewer, tester.
CRITICAL RULES:
1. STATE MANAGEMENT: Maintain a single source of truth at '.agent_logs/STATE.md'. It must track Architecture, Task Status, and Current Assignments.
2. NO SPAWNING: DO NOT use 'delegate_to_subagent'. Use 'publish_message' to assign tasks to 'coder'.
3. ASSIGNMENT: To start a task, update STATE.md then 'publish_message(workspace_path=\"$WORKSPACE\", topic=\"task\", sender_role=\"planner\", receiver_role=\"coder\", content=\"Task #X assigned. Check STATE.md\")'.
4. COMPLETION: When 'tester' notifies you of completion, mark the task as done in STATE.md. If 'tester' reports bugs, assign a fix back to 'coder'.
5. EXPLORE FIRST: Do not hallucinate files or context.
NEVER ask the user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role planner --instruction "$PLANNER_INST" --task "$TASK" --engine "$ENGINE" $MODEL_ARG > "$LOG_DIR/planner.log" 2>&1 &
sleep 10

echo "🚀 Summoning [Coder Agent]..."
CODER_INST="You are a strict CODER. Your role name is 'coder'. You are part of a 4-agent team (planner, coder, reviewer, tester).
YOUR JOB:
1. POLLING: Periodically call 'read_messages(workspace_path=\"$WORKSPACE\", receiver_role=\"coder\")' to wait for tasks from 'planner' or rework from 'reviewer'.
2. READ STATE: Read '.agent_logs/STATE.md' to understand the architecture and your specific task.
3. IMPLEMENT: Write CLEAN, TESTABLE code (SOLID, DRY). Use dependency injection. Ensure logic is isolated.
4. HANDOFF: Once done, update STATE.md then call 'publish_message(workspace_path=\"$WORKSPACE\", topic=\"review\", sender_role=\"coder\", receiver_role=\"reviewer\", content=\"Task ready for review\")'.
5. FAIL-FAST: If blocked, update STATE.md and notify 'planner'.
NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role coder --instruction "$CODER_INST" --engine "$ENGINE" $MODEL_ARG > "$LOG_DIR/coder.log" 2>&1 &
sleep 10

echo "🚀 Summoning [Reviewer Agent]..."
REVIEWER_INST="You are a strict REVIEWER/AUDITOR. You are part of a 4-agent team (planner, coder, reviewer, tester).
YOUR JOB:
1. Wait for notifications from 'coder' via 'read_messages(workspace_path=\"$WORKSPACE\", receiver_role=\"reviewer\")'.
2. AUDIT: Read '.agent_logs/STATE.md' and the changed code. Perform a complete audit of quality and security.
3. DECISION:
   - IF APPROVED: Update STATE.md and 'publish_message(workspace_path=\"$WORKSPACE\", topic=\"test\", sender_role=\"reviewer\", receiver_role=\"tester\", content=\"Review passed. Proceed to test\")'.
   - IF REJECTED: Log feedback in STATE.md and 'publish_message(workspace_path=\"$WORKSPACE\", topic=\"rework\", sender_role=\"reviewer\", receiver_role=\"coder\", content=\"Review failed. See feedback in STATE.md\")'.
Prevent infinite loops. NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role reviewer --instruction "$REVIEWER_INST" --engine "$ENGINE" $MODEL_ARG > "$LOG_DIR/reviewer.log" 2>&1 &
sleep 10

echo "🚀 Summoning [Tester Agent]..."
TESTER_INST="You are a strict TESTER/QA. You are part of a 4-agent team (planner, coder, reviewer, tester).
YOUR JOB:
1. Wait for audit approval from 'reviewer' via 'read_messages(workspace_path=\"$WORKSPACE\", receiver_role=\"tester\")'.
2. TEST & COVERAGE: Read STATE.md. Write unit tests for ALL functions implemented by coder. Aim for 100% logic coverage. Ensure tests are independent.
3. RESULT:
   - IF PASS: Mark as COMPLETED in STATE.md and 'publish_message(workspace_path=\"$WORKSPACE\", topic=\"completion\", sender_role=\"tester\", receiver_role=\"planner\", content=\"Task COMPLETED successfully\")'.
   - IF FAIL: Write bug details to STATE.md and 'publish_message(workspace_path=\"$WORKSPACE\", topic=\"bug\", sender_role=\"tester\", receiver_role=\"planner\", content=\"Bugs found. See STATE.md for details\")'.
NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role tester --instruction "$TESTER_INST" --engine "$ENGINE" $MODEL_ARG > "$LOG_DIR/tester.log" 2>&1 &

echo "🚀 Starting Web Dashboard..."
export MULTI_AGENT_DB_PATH="$WORKSPACE_DB"
cd "$SCRIPT_DIR" || exit
nohup "$PYTHON_ENV" dashboard.py > "$LOG_DIR/dashboard.log" 2>&1 &
cd "$WORKSPACE" || exit

echo "---------------------------------------------------------"
echo "✅ ALL 4 AGENTS HAVE BEEN RESET AND STARTED IN: $WORKSPACE!"
echo "📊 Monitoring Dashboard: http://localhost:6060"
echo "---------------------------------------------------------"
