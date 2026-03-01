---
description: Khởi chạy QUY TRÌNH MỚI CHO DỰ ÁN MỚI (Reset mọi trí nhớ cũ, xoá database) cho 4 Agent chạy.
---
# Start Auto Agent Workflow

Dùng khi bạn muốn Agent quên đi toàn bộ lịch sử (clear db), bắt đầu một dự án hoàn toàn mới tinh trong Workspace hiện tại. Nhớ truyền vào biến `$ARGUMENTS` để ra lệnh.

// turbo-all
1. Xóa trí nhớ cũ và bắt đầu quy trình:
`bash /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/start_agents.sh "$ARGUMENTS"`

2. Lời nhắc: Nếu dashboard chưa chạy, bạn cần mở tab Terminal mới và gọi lệnh sau. Sau đó mở `http://localhost:6060` trên cấu hình máy:
`python /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/dashboard.py $(pwd)/.agent_logs/multi_agent_bus.db`

3. Mở tab trình duyệt
`open http://localhost:6060`
