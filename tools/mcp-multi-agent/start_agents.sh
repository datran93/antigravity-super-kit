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

if [ -z "$1" ]; then
  echo "⚠️ Error: No request provided for Agent."
  echo "💡 Usage: $0 \"<Your Request>\""
  exit 1
fi

TASK="$1"

echo "🧹 Cleaning up existing Agents in this workspace (if any)..."
pkill -f "worker.py"
sleep 1

WORKSPACE_DB="$LOG_DIR/multi_agent_bus.db"
if [ -f "$WORKSPACE_DB" ]; then
    echo "🗑 Resetting Agent memory to start a NEW project..."
    rm "$WORKSPACE_DB"
fi

echo "🚀 Summoning [Planner Agent] for project: $WORKSPACE..."
PLANNER_INST="You are a strict PLANNER and ARCHITECT.
CRITICAL RULES:
1. DO NOT WRITE CODE directly. Do not implement features yourself.
2. YOUR JOB: Analyze requirements, design the architecture, and break the mission into atomic tasks.
3. DELEGATION: Send tasks to the 'coder' role using the 'publish_message' tool.
4. COORDINATION: Always call 'read_messages(receiver_role=\"planner\")' periodically to see if 'reviewer' has finished.
5. NEVER ask the user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role planner --instruction "$PLANNER_INST" --task "$TASK" > "$LOG_DIR/planner.log" 2>&1 &
sleep 10

echo "🚀 Summoning [Coder Agent]..."
CODER_INST="You are a strict CODER.
YOUR JOB:
1. Wait for tasks from 'planner' by calling 'read_messages(receiver_role=\"coder\")'.
2. Implement the requested code/files in the workspace.
3. Once done, call 'publish_message' to notify 'tester' to verify your work.
4. If 'tester' reports errors, fix them and notify again.
DO NOT plan the architecture. Just implement. NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role coder --instruction "$CODER_INST" > "$LOG_DIR/coder.log" 2>&1 &
sleep 10

echo "🚀 Summoning [Tester Agent]..."
TESTER_INST="You are a strict TESTER/QA.
YOUR JOB:
1. Wait for notifications from 'coder' by calling 'read_messages(receiver_role=\"tester\")'.
2. Verify the logic, run tests, or check for syntax errors.
3. If errors found: call 'publish_message' to 'coder' with detailed feedback.
4. If logic is correct: call 'publish_message' to 'reviewer' to approve the code.
NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role tester --instruction "$TESTER_INST" > "$LOG_DIR/tester.log" 2>&1 &
sleep 10

echo "🚀 Summoning [Reviewer Agent]..."
REVIEWER_INST="You are a strict REVIEWER/AUDITOR.
YOUR JOB:
1. Wait for approval requests from 'tester' by calling 'read_messages(receiver_role=\"reviewer\")'.
2. Perform a final audit of the code quality and security.
3. If rejected: call 'publish_message' to 'coder' with reasons.
4. If approved: call 'publish_message' to 'planner' to mark the task as COMPLETED.
NEVER ask user for confirmation."
nohup "$PYTHON_ENV" -u "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role reviewer --instruction "$REVIEWER_INST" > "$LOG_DIR/reviewer.log" 2>&1 &

echo "---------------------------------------------------------"
echo "✅ ALL 4 AGENTS HAVE BEEN RESET AND STARTED IN: $WORKSPACE!"
echo "📊 Monitoring Dashboard: http://localhost:6060"
echo "---------------------------------------------------------"
