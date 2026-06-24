---
description: Điều tra sự cố server - dùng MCP server ttmkt
---

QUAN TRỌNG: KHÔNG đọc file local, KHÔNG hỏi user. Chỉ dùng MCP tool từ `ttmkt`.

## Bước 1: Gọi `discover_logs` để biết log ở đâu

## Bước 2: Gọi `investigate_incident` để có overview

## Bước 3: Đào sâu theo kết quả
- Memory cao → `get_memory_usage`, `get_top_processes(sort="mem")`
- PM2 crash → `get_pm2_restarts`, `get_pm2_error_logs`
- OOM → `investigate_oom`, `get_dmesg_oom`
- Traffic bất thường → `count_requests_per_minute`, `top_ips_in_timerange`
- IP đáng ngờ → `analyze_ip`
- Disk I/O → `get_sar(flag="-d")`
- Load average → `get_sar(flag="-q")`

## Quy tắc
- LUÔN dùng MCP tool, KHÔNG ĐỌC FILE LOCAL
- KHÔNG hỏi user chọn option - tự chạy và trả kết quả
- Trả lời tiếng Việt, kết luận rõ ràng
