---
description: Tắt ngay lập tức tất cả hệ thống và Subagents đang cày cuốc.
---
# Stop Auto Agent Workflow

Dùng khi muốn bảo trì, update code, hoặc dừng ngốn CPU. Cắt điện hệ thống. Các tin nhắn hay code còn dang dở chưa chạy xong sẽ dừng ngay lập tức. Sau đó gọi `continue-auto-agent` để hồi sinh chúng.

// turbo-all
1. Tắt nguồn ngay lập tức (diệt process):
`bash /Users/datran/LearnDev/antigravity-kit/tools/mcp-multi-agent/stop_agents.sh`
