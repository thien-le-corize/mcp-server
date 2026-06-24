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
echo "Done! $(ls -lh tt-mcp | awk '{print $5}')"
