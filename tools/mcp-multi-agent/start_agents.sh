#!/bin/bash

# Thư mục lõi chứa mã nguồn của mô hình Multi-Agent (Cố định, không đổi)
SCRIPT_DIR="/Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent"
WORKER_SCRIPT="$SCRIPT_DIR/worker.py"
PYTHON_ENV="$SCRIPT_DIR/.venv/bin/python"
DB_PATH="$SCRIPT_DIR/multi_agent_bus.db"

# Workspace hiện tại
WORKSPACE=$(pwd)
LOG_DIR="$WORKSPACE/.agent_logs"

mkdir -p "$LOG_DIR"

if [ -z "$1" ]; then
  echo "⚠️ Lỗi: Bạn chưa cung cấp yêu cầu cho Agent."
  echo "💡 Cách dùng: $0 \"<Yêu cầu của bạn>\""
  exit 1
fi

TASK="$1"

echo "🧹 Đang dọn dẹp các Agent cũ đang chạy trong workspace này (nếu có)..."
pkill -f "worker.py"
sleep 1

if [ -f "$DB_PATH" ]; then
    echo "🗑 Xoá lịch sử bộ nhớ Agent cũ để bắt đầu dự án MỚI TINH..."
    rm "$DB_PATH"
fi

echo "🚀 Đang triệu hồi [Planner Agent] vào dự án: $WORKSPACE..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role planner --instruction "You are a strict PLANNER in a Multi-Agent architecture. You have access to tools. Do your job thoroughly without asking the user. Once done, output a final summary." --task "$TASK" > "$LOG_DIR/planner.log" 2>&1 &

echo "🚀 Đang triệu hồi [Coder Agent]..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role coder --instruction "You are a strict CODER in a Multi-Agent architecture. Bạn sẽ nhận task từ Planner qua message. Đọc task, tạo file hoặc chỉnh sửa code. Code xong gửi publish_message cho tester để nhờ kiểm thử. Sửa code nếu Tester báo lỗi." > "$LOG_DIR/coder.log" 2>&1 &

echo "🚀 Đang triệu hồi [Tester Agent]..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role tester --instruction "You are a strict TESTER in a Multi-Agent architecture. Đợi tin nhắn từ Coder xác nhận đã code xong. Nhiệm vụ của bạn là check file, chạy unit test hoặc kiểm duyệt cú pháp. Báo lỗi chi tiết lại cho Coder nếu sai, báo pass cho Reviewer nếu đúng." > "$LOG_DIR/tester.log" 2>&1 &

echo "🚀 Đang triệu hồi [Reviewer Agent]..."
nohup "$PYTHON_ENV" "$WORKER_SCRIPT" --workspace "$WORKSPACE" --role reviewer --instruction "You are a strict REVIEWER in a Multi-Agent architecture. Đợi thông báo pass từ Tester, hãy thẩm định chất lượng Code. Đạt tiêu chuẩn thì nhắn cho Planner hoàn thành." > "$LOG_DIR/reviewer.log" 2>&1 &

echo "---------------------------------------------------------"
echo "✅ TẤT CẢ 4 AGENTS ĐÃ ĐƯỢC RESET VÀ KHỞI ĐỘNG VÀO: $WORKSPACE!"
echo "📊 Dashboard theo dõi: http://localhost:6060"
echo "---------------------------------------------------------"
