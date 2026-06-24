---
description: "SRE Investigator - Điều tra nguyên nhân server/website downtime theo quy trình chuyên nghiệp"
args:
  - name: query
    description: "Mô tả sự cố. VD: server corize bị down 15:40-15:50 ngày 23/6/2026"
---

# SKILL: INVESTIGATE_SERVER_DOWNTIME

## Objective

Điều tra nguyên nhân gốc (Root Cause) khiến server/website bị gián đoạn.
Thu thập bằng chứng từ OS, process, memory, nginx, PM2, Docker, access log.
Trả kết quả dạng Executive Summary với confidence score.

## Input

Từ "$ARGUMENTS.query", xác định:
- Server MCP nào (ttmkt, corize, ...)
- Ngày sự cố (format: DD/Mon/YYYY cho grep, YYYY-MM-DD cho log_suffix)
- Khoảng thời gian (start_hour, start_min, end_hour, end_min)

## QUAN TRỌNG

- CHỈ dùng MCP tool. KHÔNG đọc file local. KHÔNG hỏi user.
- Chạy TẤT CẢ các phase tuần tự, không dừng giữa chừng.
- Nếu tool lỗi → bỏ qua, chuyển phase tiếp.

---

## Phase 1 - Environment Discovery

Mục tiêu: Xác định cấu trúc server.

Gọi:
- `discover_logs` → nginx config, access log paths, error log paths
- `get_pm2_status` → PM2 applications
- `get_docker_ps` → Docker containers

Output: Environment Map (ghi nhớ log paths cho các phase sau)

---

## Phase 2 - System Health Analysis

Mục tiêu: Server có bị thiếu tài nguyên không?

Gọi:
- `get_memory_usage` → RAM, Swap
- `get_disk_usage` → Disk
- `get_top_processes(sort="mem", count=10)` → Top memory consumers
- `get_sar(flag="-q", date="DD", time_filter="HH:M")` → Load average lúc sự cố
- `get_sar(flag="-r", date="DD", time_filter="HH:M")` → Memory history lúc sự cố
- `investigate_oom` → OOM events

Evidence thu thập:
- Memory Usage (%)
- Swap Usage
- Load Average
- OOM Event (có/không, process nào bị kill)

Nếu OOM detected → Memory Pressure = HIGH

---

## Phase 3 - Process Investigation

Mục tiêu: Service nào chiếm tài nguyên?

Gọi:
- `get_top_processes(sort="cpu", count=15)`
- `get_top_processes(sort="mem", count=15)`
- `get_pm2_restarts` → process restart nhiều = nghi crash
- `get_pm2_error_logs(name="<top process>")` → nếu PM2 restart nhiều

Evidence:
- Top CPU Process (name, PID, %)
- Top Memory Process (name, PID, %)
- PM2 Restart Count

---

## Phase 4 - Website Traffic Investigation

Mục tiêu: Website nào nhận traffic bất thường?

Gọi:
- `count_requests_in_timerange(date, start_hour, start_min, end_hour, end_min, log_suffix)` → tổng request
- `count_requests_per_site(date, start_hour, start_min, end_hour, end_min, log_suffix)` → request per website

Evidence:
- Total Requests in incident window
- Top Website By Traffic

---

## Phase 5 - Traffic Timeline Analysis

Mục tiêu: Thời điểm traffic tăng đột biến.

Gọi:
- `count_requests_per_minute(date, start_hour, start_min, end_hour, end_min, log_suffix)`

Evidence:
- Requests per minute timeline
- Peak Minute (phút nào nhiều nhất)
- Spike ratio (peak / average)

---

## Phase 6 - Top IP Analysis

Mục tiêu: IP nào tạo nhiều request nhất.

Gọi:
- `top_ips_in_timerange(date, start_hour, start_min, end_hour, end_min, limit=50, log_suffix)`

Evidence:
- Top 50 IPs with request count
- IP chiếm >10% tổng = suspicious

---

## Phase 7 - Suspicious IP Investigation

Mục tiêu: Xác định IP spam/bot.

Với mỗi IP nghi ngờ (>10% traffic hoặc >500 requests), gọi:
- `analyze_ip(ip, date, log_suffix)` → sites, URLs, status codes, user-agents, timeline

Tiêu chí đánh giá Spam Score:
- CRITICAL: >50% tổng request, single URL pattern, bot UA
- HIGH: >20% tổng request, repetitive pattern
- MEDIUM: >10% tổng request, unusual UA
- LOW: traffic cao nhưng pattern bình thường (Googlebot, monitoring)

Evidence:
- IP, Request Count, Top URLs, User-Agent, Spam Score

---

## Phase 8 - Error Analysis

Mục tiêu: Lỗi ứng dụng trong khoảng sự cố.

Gọi:
- `get_nginx_errors(lines=200)`
- `get_pm2_error_logs(name="<main app>", lines=200)`
- `get_journal(since="<start>", until="<end>", grep="error|fatal|crash|kill|restart|502|504|refused")`

Tìm keyword:
- 502, 504, Connection Refused, ECONNREFUSED
- Out Of Memory, Killed process, oom-killer
- Database Timeout, Redis Timeout
- Segmentation Fault, Crash, Restart

Evidence:
- Error Type
- Error Count
- Error Timeline

---

## Phase 9 - Correlation Analysis

Mục tiêu: Xây dựng chuỗi nguyên nhân (causal chain).

Sắp xếp tất cả evidence theo timeline:
```
HH:MM  Event
─────  ─────
15:42  Traffic spike (2000 → 50000 req/min)
15:43  Memory reaches 98%
15:44  OOM Killer kills node process
15:44  PM2 restarts app
15:45  Nginx returns 502
```

---

## Root Cause Classification

Phân loại vào 1 trong các nhóm:

**Resource Exhaustion**: Out Of Memory, Memory Leak, CPU Saturation, Disk Full
**Traffic Related**: Traffic Spike, Bot Attack, Spam Requests, DDoS
**Application Failure**: PM2 Crash, NodeJS Crash, Docker Crash
**Infrastructure**: Database Down, Redis Down, Network Issue
**Configuration**: Deploy Error, Nginx Misconfiguration, SSL Issue

---

## Output Format

Trả kết quả dạng:

```
══════════════════════════════════════
INCIDENT INVESTIGATION REPORT
══════════════════════════════════════

Incident Time: HH:MM → HH:MM DD/MM/YYYY
Root Cause: [phân loại]
Confidence: [0-100]%
Affected Service: [website/service name]

═══ EVIDENCE CHAIN ═══
1. [timestamp] [event]
2. [timestamp] [event]
...

═══ SUSPICIOUS IPs ═══
[IP] - [count] requests - Spam Score: [level]

═══ RECOMMENDED ACTIONS ═══
1. [Immediate] ...
2. [Short-term] ...
3. [Long-term] ...
```
