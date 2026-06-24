# TT-MCP

Read-only MCP server for incident investigation. Pure Go, zero dependencies, ~5MB RAM.

## How It Works

```
┌──────────────────────────────────────────────────────┐
│  Máy Mac (local)                                      │
│                                                       │
│  1. Bạn gõ: "kiểm tra memory trên corize"            │
│                    │                                  │
│                    ▼                                  │
│  2. Claude Code (AI) đọc ~/.claude.json               │
│     → thấy MCP server "corize" config:               │
│        command: ssh                                   │
│        args: [-T, -i, ~/.ssh/tt-mcp-key, ...]        │
│     → spawn SSH process                              │
│                    │                                  │
│  3. SSH gửi JSON qua stdin:                           │
│     {"method":"tools/call","params":{"name":          │
│      "get_memory_usage"}}                            │
│                    │                                  │
└────────────────────┼─────────────────────────────────┘
                     │ SSH tunnel (encrypted)
                     ▼
┌──────────────────────────────────────────────────────┐
│  Server (46.137.253.133)                              │
│                                                       │
│  4. sshd nhận connection từ key tt-mcp-key            │
│     → authorized_keys có force-command                │
│     → BẮT BUỘC chạy /opt/tt-mcp/tt-mcp              │
│     → KHÔNG THỂ chạy lệnh khác                      │
│                                                       │
│  5. Binary tt-mcp đọc JSON từ stdin                   │
│     → parse: tool = "get_memory_usage"               │
│     → chạy: free -h                                  │
│     → trả JSON qua stdout                           │
│                                                       │
└────────────────────┼─────────────────────────────────┘
                     │ stdout (JSON response)
                     ▼
┌──────────────────────────────────────────────────────┐
│  6. Claude Code nhận kết quả                          │
│     → phân tích: "RAM 3.8GB, dùng 932MB"            │
│     → trả lời bạn bằng tiếng Việt                   │
└──────────────────────────────────────────────────────┘
```

### Tại sao Claude có quyền SSH?

Claude Code đọc file `~/.claude.json` trên máy Mac của bạn. File này chứa:

```json
{
  "mcpServers": {
    "corize": {
      "command": "ssh",
      "args": ["-T", "-i", "~/.ssh/tt-mcp-key", "mcp-reader@46.137.253.133"]
    }
  }
}
```

Khi cần gọi tool MCP, Claude Code **spawn process SSH** trên máy bạn (giống bạn gõ `ssh ...` trong terminal). SSH key nằm trên máy bạn (`~/.ssh/tt-mcp-key`).

### Flow khi dùng `/investigate`

```
/investigate web corize bị down 15:40-15:50 ngày 23/6

Claude đọc skill investigate.md → biết quy trình 9 phase
│
├─ Phase 1: gọi discover_logs → biết log ở /home/*/logs/nginx/
├─ Phase 2: gọi get_sar → load=25.65, blocked=42
├─ Phase 3: gọi get_top_processes → mysql 880MB
├─ Phase 4: gọi top_ips_in_timerange → IP 185.177.72.51 = 504 req
├─ Phase 5: gọi count_requests_per_minute → spike lúc 15:48
├─ Phase 6: gọi analyze_ip → bot, UA bất thường
├─ Phase 7: đánh giá Spam Score
├─ Phase 8: gọi get_journal(grep="oom|killed") → memory pressure
└─ Phase 9: correlation → viết report + lưu file .md
```

## Architecture

```
Claude Code (máy local)
    │
    │ SSH (force-command, read-only key)
    ▼
Server (chạy /opt/tt-mcp/tt-mcp)
    │
    │ Read-only commands
    ▼
Linux Resources (logs, memory, disk, processes)
```

## Tools (26)

| Tool | Description |
|------|-------------|
| **System** | |
| `get_memory_usage` | Memory & swap |
| `get_disk_usage` | Disk space |
| `get_top_processes` | Top processes by CPU/MEM |
| `get_disk_io` | Disk I/O from sar |
| `get_sar` | Any sar flag: -q (load), -r (mem), -u (cpu), -d (disk), -S (swap) |
| `discover_logs` | Auto-detect all nginx log paths on server |
| **PM2** | |
| `get_pm2_status` | PM2 process list (JSON) |
| `get_pm2_logs` | PM2 stdout logs |
| `get_pm2_error_logs` | PM2 error logs |
| `get_pm2_restarts` | PM2 restart history |
| `get_pm2_report` | Full PM2 diagnostic report |
| **Docker** | |
| `get_docker_ps` | Docker containers |
| `get_docker_logs` | Docker logs (with since/until) |
| **Nginx** | |
| `get_nginx_errors` | Nginx error log |
| `get_nginx_access` | Nginx access log |
| **Traffic Analysis** | |
| `count_requests_in_timerange` | Count requests in time range |
| `count_requests_per_site` | Requests per website |
| `count_requests_per_minute` | Requests per minute (find spikes) |
| `top_ips_in_timerange` | Top IPs |
| `top_urls_in_timerange` | Top URLs |
| `analyze_ip` | Full IP analysis (sites, URLs, UA, status, timeline) |
| `grep_requests` | Custom grep on access logs |
| **OOM** | |
| `investigate_oom` | OOM killer investigation |
| `get_dmesg_oom` | Kernel OOM messages |
| **Other** | |
| `get_journal` | Systemd journal (with unit, since, until, grep) |
| `get_app_log` | Read allowed log files |
| `investigate_incident` | Quick overview |

## Install (Server mới)

```bash
git clone https://github.com/thien-le-corize/mcp-server.git /tmp/tt-mcp
cd /tmp/tt-mcp
sudo ./setup.sh
```

Setup tự động:
1. Cài Go (nếu chưa có)
2. Build binary (~2MB)
3. Install vào `/opt/tt-mcp/`
4. Tạo user `mcp-reader` (read-only)
5. Cài ACL + cấp quyền đọc log (auto-detect từ nginx -T)
6. Tạo SSH force-command key
7. Cấu hình sudoers cho nginx -T, pm2, journalctl, dmesg

## Update (Server đã cài)

```bash
cd /tmp/tt-mcp
git pull
sudo ./deploy.sh
```

## Kết nối Claude Code (Máy local)

### Bước 1: Copy key từ server

```bash
scp root@SERVER_IP:/opt/tt-mcp/.ssh/tt-mcp-key ~/.ssh/tt-mcp-servername
chmod 600 ~/.ssh/tt-mcp-servername
```

Hoặc `cat /opt/tt-mcp/.ssh/tt-mcp-key` trên server rồi paste vào file.

### Bước 2: Test SSH

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ssh -T -i ~/.ssh/tt-mcp-servername -o StrictHostKeyChecking=accept-new mcp-reader@SERVER_IP
```

Phải trả về JSON → OK.

### Bước 3: Thêm vào Claude Code

```bash
claude mcp add -s user servername -- ssh -T -i /path/to/.ssh/tt-mcp-servername -o StrictHostKeyChecking=accept-new mcp-reader@SERVER_IP
```

### Bước 4: Verify

```bash
claude mcp list
```

## Skills (Claude Commands)

Copy `.claude/commands/` vào `~/.claude/commands/`:

```bash
cp .claude/commands/*.md ~/.claude/commands/
```

| Skill | Cách dùng |
|-------|-----------|
| `/investigate` | Điều tra sự cố đầy đủ 9 phase |
| `/check-server` | Health check nhanh |
| `/grep-requests` | Đếm/tìm request nginx |
| `/analyze-traffic` | Phân tích traffic theo thời gian |
| `/check-changes` | Kiểm tra apt/service restart |

### Ví dụ

```
/investigate dùng mcp corize, web bị down 15:40-15:50 ngày 23/6/2026
/check-server trên ttmkt
/grep-requests top IP 15:40-15:51 ngày 23/6 trên corize
```

## Security

| Layer | Protection |
|-------|-----------|
| SSH Key | Force-command: key chỉ chạy được `/opt/tt-mcp/tt-mcp` |
| User | `mcp-reader`: không có shell login, chỉ đọc log |
| ACL | Chỉ `mcp-reader` được traverse `/home/*/`, đọc `/home/*/logs/` |
| Sudoers | Chỉ cho phép: `nginx -T`, `pm2`, `journalctl`, `dmesg`, `find` |
| Binary | Chỉ chạy read-only commands, không có shell exec tùy ý |

**Không thể:** delete file, modify config, restart service, chạy shell tùy ý.

## Permissions (Nếu log đọc không được)

```bash
# Cài ACL
sudo apt install acl -y

# Cho mcp-reader traverse /home/*/
sudo setfacl -m u:mcp-reader:x /home/*/

# Cho mcp-reader đọc log
sudo setfacl -R -m u:mcp-reader:rx /home/*/logs/

# Nếu không có ACL, dùng chmod
sudo chmod o+x /home/*/
sudo chmod -R o+r /home/*/logs/
```

## Auto-detect

Binary tự detect khi khởi động:
- **Nginx log paths**: từ `sudo nginx -T`
- **Log base path**: từ access_log paths trong nginx config
- **Node/PM2 path**: tìm trong `/www/server/nodejs/`, `/root/.nvm/`, `/usr/local/bin/`
- **Không cần sửa config.json thủ công**

## Troubleshooting

| Lỗi | Fix |
|-----|-----|
| `Permission denied` khi đọc log | `sudo setfacl -m u:mcp-reader:x /home/*/ && sudo setfacl -R -m u:mcp-reader:rx /home/*/logs/` |
| `pm2: command not found` | Sửa `"path"` trong `/opt/tt-mcp/config.json` thêm đường dẫn node |
| `Text file busy` khi deploy | `sudo killall tt-mcp` trước khi copy |
| Claude Code `/mcp` không thấy | `claude mcp add -s user ...` rồi restart Claude Code |
| `-32000` connection error | Thêm `-T` flag vào SSH args |
| `Pseudo-terminal` warning | Thêm `-T` flag vào SSH args |
| `sshd.service not found` | Server dùng `ssh` thay `sshd`, đã tự handle |
