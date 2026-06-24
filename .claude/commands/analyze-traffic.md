---
description: Phân tích traffic nginx theo khoảng thời gian - dùng MCP server ttmkt
args:
  - name: timerange
    description: "Khoảng thời gian cần phân tích, VD: 15:40-15:51 ngày 23/6/2026"
---

Dùng MCP server `ttmkt` để phân tích traffic nginx trong khoảng $ARGUMENTS.timerange:

1. `count_requests_in_timerange` - Tổng request
2. `count_requests_per_minute` - Request theo phút (tìm spike)
3. `top_ips_in_timerange(limit=20)` - Top IP
4. `top_urls_in_timerange(limit=30)` - Top URL

Nếu phát hiện IP bất thường (>100 request), tự động gọi `analyze_ip` cho IP đó.

Trả kết quả tiếng Việt, kết luận: traffic bình thường hay bất thường, IP nào đáng ngờ.
