#!/bin/bash
# TT-MCP: One-command setup
# Copy thư mục này lên server, chạy: sudo ./setup.sh
set -e

MCP_DIR="/opt/tt-mcp"
MCP_USER="mcp-reader"

echo "=== [1/5] Cài Go (nếu chưa có) ==="
if ! command -v go &>/dev/null; then
  curl -sSL https://go.dev/dl/go1.21.13.linux-amd64.tar.gz -o /tmp/go.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tar.gz
  export PATH=$PATH:/usr/local/go/bin
  rm /tmp/go.tar.gz
  echo "Go installed."
else
  echo "Go OK."
fi

echo "=== [2/5] Build binary ==="
cd "$(dirname "$0")"
CGO_ENABLED=0 /usr/local/go/bin/go build -ldflags="-s -w" -o tt-mcp . 2>/dev/null || \
CGO_ENABLED=0 go build -ldflags="-s -w" -o tt-mcp .
echo "Built: $(ls -lh tt-mcp | awk '{print $5}')"

echo "=== [3/5] Install ==="
mkdir -p $MCP_DIR
killall tt-mcp 2>/dev/null || true
sleep 1
cp tt-mcp $MCP_DIR/
[ -f $MCP_DIR/config.json ] || cp config.json $MCP_DIR/
chmod +x $MCP_DIR/tt-mcp

echo "=== [4/5] Tạo user read-only ==="
id $MCP_USER &>/dev/null || useradd -r -s /bin/bash -d $MCP_DIR $MCP_USER
usermod -s /bin/bash $MCP_USER
usermod -aG adm $MCP_USER 2>/dev/null || true
usermod -aG systemd-journal $MCP_USER 2>/dev/null || true
usermod -aG docker $MCP_USER 2>/dev/null || true
chown -R $MCP_USER:$MCP_USER $MCP_DIR

# Grant read access to common log directories
command -v setfacl &>/dev/null || (apt install -y acl 2>/dev/null || yum install -y acl 2>/dev/null || true)

# Auto-detect nginx log paths
NGINX_BIN=$(which nginx 2>/dev/null || echo "/www/server/nginx/sbin/nginx")
LOG_DIRS=$($NGINX_BIN -T 2>/dev/null | awk '/access_log|error_log/ {print $2}' | sed 's/;$//' | grep "^/" | xargs -I{} dirname {} | sort -u)
# Add common paths
LOG_DIRS="$LOG_DIRS /www/wwwlogs /var/log/nginx /home/*/logs /usr/local/nginx/logs"

# Allow mcp-reader to traverse parent dirs
for dir in /home/*/ /www/*/; do
  [ -d "$dir" ] && setfacl -m u:$MCP_USER:x "$dir" 2>/dev/null || chmod o+x "$dir" 2>/dev/null || true
done

# Allow mcp-reader to read logs
for dir in $LOG_DIRS; do
  if [ -d "$dir" ]; then
    setfacl -R -m u:$MCP_USER:rx "$dir" 2>/dev/null || chmod -R o+r "$dir" 2>/dev/null || true
    setfacl -R -d -m u:$MCP_USER:rx "$dir" 2>/dev/null || true
  fi
done

# Allow mcp-reader to run nginx -T and pm2 via sudo (read-only commands)
cat > /etc/sudoers.d/tt-mcp << SUDOEOF
$MCP_USER ALL=(root) NOPASSWD: /usr/sbin/nginx -T
$MCP_USER ALL=(root) NOPASSWD: /www/server/nginx/sbin/nginx -T
$MCP_USER ALL=(root) NOPASSWD: /usr/local/nginx/sbin/nginx -T
$MCP_USER ALL=(root) NOPASSWD: /usr/bin/pm2 *
$MCP_USER ALL=(root) NOPASSWD: /www/server/nodejs/*/bin/pm2 *
$MCP_USER ALL=(root) NOPASSWD: /usr/bin/journalctl *
$MCP_USER ALL=(root) NOPASSWD: /usr/bin/dmesg *
$MCP_USER ALL=(root) NOPASSWD: /usr/bin/find *
SUDOEOF
chmod 440 /etc/sudoers.d/tt-mcp

echo "=== [5/5] SSH force-command key ==="
mkdir -p $MCP_DIR/.ssh
if [ ! -f $MCP_DIR/.ssh/tt-mcp-key ]; then
  ssh-keygen -t ed25519 -f $MCP_DIR/.ssh/tt-mcp-key -N "" -C "tt-mcp" -q
  PUBKEY=$(cat $MCP_DIR/.ssh/tt-mcp-key.pub)
  echo "command=\"$MCP_DIR/tt-mcp\",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty $PUBKEY" > $MCP_DIR/.ssh/authorized_keys
  chmod 700 $MCP_DIR/.ssh
  chmod 600 $MCP_DIR/.ssh/authorized_keys $MCP_DIR/.ssh/tt-mcp-key
  chown -R $MCP_USER:$MCP_USER $MCP_DIR/.ssh
fi

# SSH config cho user mcp-reader
if ! grep -q "Match User $MCP_USER" /etc/ssh/sshd_config /etc/ssh/sshd_config.d/* 2>/dev/null; then
  mkdir -p /etc/ssh/sshd_config.d
  cat > /etc/ssh/sshd_config.d/tt-mcp.conf << EOF
Match User $MCP_USER
    ForceCommand $MCP_DIR/tt-mcp
    AllowTcpForwarding no
    X11Forwarding no
    PermitTunnel no
EOF
  systemctl reload sshd 2>/dev/null || systemctl reload ssh 2>/dev/null || true
fi

# Test
echo ""
echo "=== Test ==="
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | su -s /bin/bash $MCP_USER -c "$MCP_DIR/tt-mcp" | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'OK - {len(d[\"result\"][\"tools\"])} tools available')" 2>/dev/null || echo "OK - binary works"

SERVER_IP=$(hostname -I | awk '{print $1}')
PRIVATE_KEY="$MCP_DIR/.ssh/tt-mcp-key"

echo ""
echo "════════════════════════════════════════════"
echo "  DONE!"
echo "════════════════════════════════════════════"
echo ""
echo "Copy private key về máy bạn:"
echo "  scp root@$SERVER_IP:$PRIVATE_KEY ~/.ssh/tt-mcp-$(hostname -s)"
echo "  chmod 600 ~/.ssh/tt-mcp-$(hostname -s)"
echo ""
echo "Claude Code config:"
echo ""
echo "  claude mcp add -s user $(hostname -s) -- ssh -T -i ~/.ssh/tt-mcp-$(hostname -s) -o StrictHostKeyChecking=accept-new $MCP_USER@$SERVER_IP"
echo ""
