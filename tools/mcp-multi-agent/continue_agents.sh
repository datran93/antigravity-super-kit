#!/bin/bash

# Core directory for Multi-Agent model (fixed)
SCRIPT_DIR="/Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent"
WORKER_SCRIPT="$SCRIPT_DIR/worker.py"
PYTHON_ENV="$SCRIPT_DIR/.venv/bin/python"

WORKSPACE=$(pwd)
LOG_DIR="$WORKSPACE/.agent_logs"
mkdir -p "$LOG_DIR"

echo "🔋 Resuming background functions (cleaning existing processes)..."
pkill -f "worker.py"
pkill -f "dashboard.py"
sleep 1

# Note the --resume flag being passed to all agents

echo "♻️ Resuming [Planner Agent] with previous memory..."
PLANNER_INST="You are a strict PLANNER and ARCHITECT.
CRITICAL RULES:
1. DO NOT WRITE CODE directly. Do not implement features yourself.
2. YOUR JOB: Analyze requirements using tools (view_file, list_dir), design the architecture, and break the mission into atomic tasks.
3. DELEGATION: Send tasks to the 'coder' role ONE AT A TIME using the 'publish_message' tool. Wait for 'reviewer' to approve a task before sending the next one.
4. COORDINATION: Always call 'read_messages(receiver_role=\"planner\")' periodically to see if 'reviewer' has finished.
5. EXPLORE FIRST: Do not hallucinate files or context.
6. NEVER ask the user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role planner --instruction "$PLANNER_INST" --resume > "$LOG_DIR/planner.log" 2>&1 &
sleep 10

echo "♻️ Resuming [Coder Agent]..."
CODER_INST="You are a strict CODER.
YOUR JOB:
1. Wait for tasks from 'planner' by calling 'read_messages(receiver_role=\"coder\")'.
2. EXPLORE: Use tools (list_dir, view_file) to understand existing code before writing. Do not hallucinate.
3. Implement the requested code in the workspace.
4. Once done, call 'publish_message' to notify 'tester' to verify your work.
5. If 'tester' reports errors, fix them. IMPORTANT: If you fail to fix the issue after 3 attempts, escalate back to 'planner' for help. Prevent infinite loops!
DO NOT plan the architecture. Just implement. NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role coder --instruction "$CODER_INST" --resume > "$LOG_DIR/coder.log" 2>&1 &
sleep 10

echo "♻️ Resuming [Tester Agent]..."
TESTER_INST="You are a strict TESTER/QA.
YOUR JOB:
1. Wait for notifications from 'coder' by calling 'read_messages(receiver_role=\"tester\")'.
2. Verify the logic, run tests, or check for syntax errors using CLI commands.
3. If errors found: call 'publish_message' to 'coder' with detailed feedback. IMPORTANT: If 'coder' fails to fix issues after 3 attempts, inform 'coder' to escalate.
4. If logic is correct: call 'publish_message' to 'reviewer' to approve the code.
NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role tester --instruction "$TESTER_INST" --resume > "$LOG_DIR/tester.log" 2>&1 &
sleep 10

echo "♻️ Resuming [Reviewer Agent]..."
REVIEWER_INST="You are a strict REVIEWER/AUDITOR.
YOUR JOB:
1. Wait for approval requests from 'tester' by calling 'read_messages(receiver_role=\"reviewer\")'.
2. Perform a final audit of the code quality and security.
3. If rejected: call 'publish_message' to 'coder' with reasons. Be pragmatic.
4. If approved: call 'publish_message' to 'planner' to mark the task as COMPLETED.
Prevent infinite loops by accepting reasonable work instead of demanding perfection. NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role reviewer --instruction "$REVIEWER_INST" --resume > "$LOG_DIR/reviewer.log" 2>&1 &

echo "🚀 Starting Web Dashboard..."
export MULTI_AGENT_DB_PATH="$LOG_DIR/multi_agent_bus.db"
cd "$SCRIPT_DIR" || exit
nohup "$PYTHON_ENV" dashboard.py > "$LOG_DIR/dashboard.log" 2>&1 &
cd "$WORKSPACE" || exit

echo "---------------------------------------------------------"
echo "✅ ALL AGENTS HAVE RESUMED WITH CONTEXT AT: $WORKSPACE!"
echo "📊 Monitoring Dashboard: http://localhost:6060"
echo "---------------------------------------------------------"
