#!/bin/bash

SCRIPT_DIR="/Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent"
WORKER_SCRIPT="$SCRIPT_DIR/worker.py"
PYTHON_ENV="$SCRIPT_DIR/.venv/bin/python"

WORKSPACE=$(pwd)
LOG_DIR="$WORKSPACE/.agent_logs"
mkdir -p "$LOG_DIR"

echo "🔋 Đang khôi phục các chức năng ngầm..."
pkill -f "worker.py"
sleep 1

# Note the --resume flag being passed to all agents

echo "♻️ Khôi phục [Planner Agent] với trí nhớ cũ..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role planner --instruction "You are a strict PLANNER in a Multi-Agent architecture. You have access to tools. Do your job thoroughly without asking the user. Once done, output a final summary." --resume > "$LOG_DIR/planner.log" 2>&1 &

echo "♻️ Khôi phục [Coder Agent]..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role coder --instruction "You are a strict CODER in a Multi-Agent architecture. Bạn sẽ nhận task từ Planner qua message. Đọc task, tạo file hoặc chỉnh sửa code. Code xong gửi publish_message cho tester để nhờ kiểm thử. Sửa code nếu Tester báo lỗi." --resume > "$LOG_DIR/coder.log" 2>&1 &

echo "♻️ Khôi phục [Tester Agent]..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role tester --instruction "You are a strict TESTER in a Multi-Agent architecture. Đợi tin nhắn từ Coder xác nhận đã code xong. Nhiệm vụ của bạn là check file, chạy unit test hoặc kiểm duyệt cú pháp. Báo lỗi chi tiết lại cho Coder nếu sai, báo pass cho Reviewer nếu đúng." --resume > "$LOG_DIR/tester.log" 2>&1 &

echo "♻️ Khôi phục [Reviewer Agent]..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role reviewer --instruction "You are a strict REVIEWER in a Multi-Agent architecture. Đợi thông báo pass từ Tester, hãy thẩm định chất lượng Code. Đạt tiêu chuẩn thì nhắn cho Planner hoàn thành." --resume > "$LOG_DIR/reviewer.log" 2>&1 &

echo "---------------------------------------------------------"
echo "✅ TẤT CẢ AGENTS ĐÃ SỐNG DẬY VÀ NHỚ LẠI BỐI CẢNH CŨ TẠI: $WORKSPACE!"
echo "📊 Dashboard theo dõi: http://localhost:6060"
echo "---------------------------------------------------------"
