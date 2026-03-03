#!/bin/bash
WORKSPACE=$(pwd)
echo "🛑 Stopping all background Agents and Dashboard running on project: $WORKSPACE..."
pkill -f "worker.py --workspace $WORKSPACE"
pkill -f "dashboard.py"
echo "📝 Updating Agent statuses to OFFLINE..."
DB_PATH="$WORKSPACE/.agent_logs/multi_agent_bus.db"
if [ -f "$DB_PATH" ]; then
    sqlite3 "$DB_PATH" "UPDATE agent_status SET status = 'OFFLINE', current_task = '' ;"
fi
echo "🧹 Cleaning up resource locks..."
rm -f "$WORKSPACE/.agent_logs"/*.lock
rm -f "$WORKSPACE/.agent_logs"/*.db-shm
rm -f "$WORKSPACE/.agent_logs"/*.db-wal
echo "✅ Agents and Dashboard stopped."
