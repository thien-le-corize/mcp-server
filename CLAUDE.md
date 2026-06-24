# CLAUDE.md

## MCP Server

Server `corize` được kết nối qua MCP. Khi điều tra server:

- LUÔN dùng MCP tool từ `corize` - KHÔNG BAO GIỜ đọc file local
- KHÔNG BAO GIỜ hỏi user chọn A/B/C - tự chạy tool và trả kết quả
- Nếu tool lỗi, thử tool khác, KHÔNG hỏi user

## Quy trình bắt buộc

1. Gọi `discover_logs` trước tiên nếu chưa biết log ở đâu
2. Dùng kết quả `discover_logs` để biết đường dẫn đúng
3. Sau đó gọi các tool phân tích (top_ips, count_requests, etc.)

## Tools có sẵn từ MCP ttmkt

- get_memory_usage, get_disk_usage, get_top_processes
- get_disk_io, get_sar
- get_pm2_status, get_pm2_logs, get_pm2_error_logs, get_pm2_restarts
- get_docker_ps, get_docker_logs
- get_nginx_errors, get_nginx_access
- count_requests_in_timerange, count_requests_per_site, count_requests_per_minute
- top_ips_in_timerange, top_urls_in_timerange
- analyze_ip, grep_requests
- investigate_oom, get_dmesg_oom
- get_journal, get_app_log
- discover_logs, investigate_incident
