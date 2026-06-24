#!/bin/bash
set -e
cd "$(dirname "$0")"
export PATH=$PATH:/usr/local/go/bin

echo "Building..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o tt-mcp .

echo "Deploying..."
sudo killall tt-mcp 2>/dev/null || true
sudo cp tt-mcp /opt/tt-mcp/
sudo cp config.json /opt/tt-mcp/

# Fix permissions
MCP_USER="mcp-reader"
for dir in /www/wwwlogs /var/log/nginx /home/*/logs; do
  [ -d "$dir" ] && sudo setfacl -R -m u:$MCP_USER:rx "$dir" 2>/dev/null || sudo chmod -R o+r "$dir" 2>/dev/null || true
done

echo "Done! $(ls -lh tt-mcp | awk '{print $5}')"
