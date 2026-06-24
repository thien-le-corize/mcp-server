---
description: Grep & đếm request nginx theo thời gian, website, phút, IP, URL - dùng MCP server ttmkt
args:
  - name: query
    description: "Mô tả cần tìm. VD: đếm request 15:40-15:51 ngày 23/6, hoặc: tìm request elementor lúc 15:50"
---

QUAN TRỌNG: KHÔNG đọc file local, KHÔNG hỏi user. Chỉ dùng MCP tool từ `ttmkt`.

Bước 1: Gọi `discover_logs` để biết log_base_path và log file pattern.
Bước 2: Dùng tool phù hợp với kết quả discover_logs.

Dựa vào yêu cầu "$ARGUMENTS.query", chọn tool phù hợp:

## Đếm tổng request trong khoảng thời gian
→ `count_requests_in_timerange(date, start_hour, start_min, end_hour, end_min, log_suffix)`

## Đếm theo từng website
→ `count_requests_per_site(date, start_hour, start_min, end_hour, end_min, log_suffix)`

## Đếm theo từng phút
→ `count_requests_per_minute(date, start_hour, start_min, end_hour, end_min, log_suffix)`

## Top IP trong khoảng đó
→ `top_ips_in_timerange(date, start_hour, start_min, end_hour, end_min, limit, log_suffix)`

## Top URL trong khoảng đó
→ `top_urls_in_timerange(date, start_hour, start_min, end_hour, end_min, limit, log_suffix)`

## Tìm request theo pattern (elementor, admin-ajax, wp-json...)
→ `grep_requests(pattern, date, start_hour, start_min, end_hour, end_min, log_suffix)`

## Phân tích 1 IP cụ thể
→ `analyze_ip(ip, date, log_suffix)`

## Tham số

- `date`: format trong nginx log, VD: `23/Jun/2026`
- `log_suffix`: tên file suffix, VD: `2026-06-23` (cho file access.log-2026-06-23)
- `start_hour`, `end_hour`: giờ 2 chữ số, VD: `15`
- `start_min`, `end_min`: phút 2 chữ số, VD: `40`, `51`

## Quy tắc

- Nếu user chỉ nói "ngày 23/6" → date=`23/Jun/2026`, log_suffix=`2026-06-23`
- Nếu không nói ngày → dùng log hôm nay, log_suffix để trống
- Luôn chạy nhiều tool cùng lúc nếu user hỏi chung (VD: "phân tích traffic" → chạy cả count, per_minute, top_ips)
- Khi thấy IP có >100 request → tự động `analyze_ip`
- Trả kết quả tiếng Việt, format bảng dễ đọc
