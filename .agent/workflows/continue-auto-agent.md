---
description: KHÔI PHỤC VÀ TIẾP TỤC DỰ ÁN ĐANG CODE (Load lại trí nhớ từ DB, không cần Argument).
---
# Continue Auto Agent Workflow

Khi bạn nhỡ tay tắt hệ thống (`stop-auto-agent`) hoặc máy khởi động lại, dùng lệnh này để bật lại 4 Agent. Chúng sẽ TỰ ĐỘNG đọc đoạn chat cuối cùng trong DB và code tiếp chỗ đang dang dở. Không truyền Argument cũng được.

// turbo-all
1. Hồi sinh các dòng thời gian (ngâm vào `--resume` flag):
`bash /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/continue_agents.sh`

2. Mở Dashboard trên trình duyệt để theo dõi trực tiếp các Agent làm việc:
`open http://localhost:6060`
