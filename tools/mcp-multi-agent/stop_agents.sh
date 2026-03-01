#!/bin/bash
WORKSPACE=$(pwd)
echo "🛑 Stopping all background Agents and Dashboard running on project: $WORKSPACE..."
pkill -f "worker.py --workspace $WORKSPACE"
pkill -f "dashboard.py"
echo "✅ Agents and Dashboard stopped."
