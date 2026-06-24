# TT-MCP

Read-only MCP server for incident investigation. Pure Go, zero dependencies, ~5MB RAM.

## Tools

| Tool | Description |
|------|-------------|
| `get_memory_usage` | Memory & swap |
| `get_disk_usage` | Disk space |
| `get_disk_io` | Disk I/O from sar |
| `get_top_processes` | Top processes by CPU/MEM |
| `get_pm2_status` | PM2 process list |
| `get_pm2_logs` | PM2 stdout logs |
| `get_pm2_error_logs` | PM2 error logs |
| `get_pm2_restarts` | PM2 restart history |
| `get_docker_ps` | Docker containers |
| `get_docker_logs` | Docker logs |
| `get_nginx_errors` | Nginx error log |
| `get_nginx_access` | Nginx access log |
| `count_requests_in_timerange` | Count requests in time range |
| `count_requests_per_site` | Requests per website |
| `count_requests_per_minute` | Requests per minute |
| `top_ips_in_timerange` | Top IPs |
| `top_urls_in_timerange` | Top URLs |
| `analyze_ip` | Full IP analysis (sites, URLs, UA, timeline) |
| `grep_requests` | Custom grep on access logs |
| `investigate_oom` | OOM killer investigation |
| `get_dmesg_oom` | Kernel OOM messages |
| `get_journal` | Systemd journal |
| `get_app_log` | Read allowed log files |
| `investigate_incident` | Quick overview |

## Install

```bash
scp -r . root@your-server:/tmp/tt-mcp
ssh root@your-server "cd /tmp/tt-mcp && sudo ./setup.sh"
```

## Security

- Dedicated `mcp-reader` user (no shell)
- SSH force-command (key can only run tt-mcp binary)
- Binary only executes read-only commands
- No shell access, no write, no delete, no restart
