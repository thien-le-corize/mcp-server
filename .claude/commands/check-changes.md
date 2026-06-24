---
description: Kiểm tra thay đổi hệ thống (apt upgrade, service restart, package update) - dùng MCP server ttmkt
args:
  - name: timerange
    description: "Khoảng thời gian, VD: 15:55-16:01 ngày 23/6/2026"
---

Dùng MCP server `ttmkt` để kiểm tra hệ thống có thay đổi gì trong khoảng $ARGUMENTS.timerange.

## Bước 1: Kiểm tra journal hệ thống

Gọi `get_journal` với:
- `since`: thời gian bắt đầu (format: "2026-06-23 15:55:00")
- `until`: thời gian kết thúc (format: "2026-06-23 16:01:00")
- `grep`: "apt|upgrade|restart|php|systemctl|stop|start|dpkg|yum|dnf"

## Bước 2: Kiểm tra service cụ thể (nếu cần)

- `get_journal(unit="php8.0-fpm", since="...", until="...")`
- `get_journal(unit="nginx", since="...", until="...")`
- `get_journal(unit="mysql", since="...", until="...")`

## Bước 3: Kiểm tra apt/dpkg log

- `get_app_log(path="/var/log/apt/history.log", lines=50)`
- `get_app_log(path="/var/log/dpkg.log", lines=50)`

## Kết luận cần trả lời

- Có apt upgrade/update chạy không?
- Có package nào được cài/gỡ/update không?
- Có service nào bị restart không? (nginx, php-fpm, mysql, docker...)
- Có systemctl stop/start/restart nào không?
- Có unattended-upgrades tự chạy không?

Trả kết quả tiếng Việt.
