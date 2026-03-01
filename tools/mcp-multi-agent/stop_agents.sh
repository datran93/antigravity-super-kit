#!/bin/bash
WORKSPACE=$(pwd)
echo "🛑 Đang dập tắt toàn bộ các Agents đang chạy ngầm trên dự án: $WORKSPACE..."
pkill -f "worker.py --workspace $WORKSPACE"
echo "✅ Đã tắt xong."
