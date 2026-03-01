#!/bin/bash
WORKSPACE=$(pwd)
echo "🛑 Stopping all background Agents running on project: $WORKSPACE..."
pkill -f "worker.py --workspace $WORKSPACE"
echo "✅ Agents stopped."
