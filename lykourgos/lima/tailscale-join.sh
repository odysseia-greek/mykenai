#!/bin/bash
# Join a Lima VM to Tailscale so it's reachable via Magic DNS
#
# Usage:
#   ./tailscale-join.sh                        # reads key from gopass
#   ./tailscale-join.sh tskey-auth-...         # explicit key
#   TAILSCALE_AUTHKEY=tskey-auth-... ./tailscale-join.sh
#
# After this script, Traefik is reachable at:
#   http://lima-byzantion.<tailnet>.ts.net:8080/

set -e

LIMA_VM="byzantion"
GOPASS_PATH="odysseia/tailscale/auth"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Resolve auth key: explicit arg > env var > gopass
AUTH_KEY="${1:-${TAILSCALE_AUTHKEY:-}}"

if [ -z "$AUTH_KEY" ]; then
  if command -v gopass >/dev/null 2>&1; then
    echo "Reading auth key from gopass (${GOPASS_PATH})..."
    AUTH_KEY=$(gopass show -o "${GOPASS_PATH}" 2>/dev/null || true)
  fi
fi

if [ -z "$AUTH_KEY" ]; then
  echo -e "${RED}Error: no auth key found${NC}"
  echo ""
  echo "Tried (in order):"
  echo "  1. First argument"
  echo "  2. TAILSCALE_AUTHKEY env var"
  echo "  3. gopass show ${GOPASS_PATH}"
  echo ""
  echo "Usage:"
  echo "  $0 tskey-auth-..."
  echo "  TAILSCALE_AUTHKEY=tskey-auth-... $0"
  exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Tailscale setup for ${LIMA_VM}${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check the VM is running
if ! limactl list --format '{{.Name}} {{.Status}}' 2>/dev/null | grep -q "^${LIMA_VM} Running"; then
  echo -e "${RED}Error: Lima VM '${LIMA_VM}' is not running${NC}"
  echo "Start it with: limactl start ${LIMA_VM}"
  exit 1
fi

echo "Checking if Tailscale is already installed..."
if limactl shell "$LIMA_VM" which tailscale >/dev/null 2>&1; then
  echo -e "${YELLOW}Tailscale already installed, skipping install${NC}"
else
  echo "Installing Tailscale..."
  limactl shell "$LIMA_VM" sudo sh -c 'curl -fsSL https://tailscale.com/install.sh | sh'
  echo -e "${GREEN}✓ Tailscale installed${NC}"
fi

echo ""
echo "Enabling tailscaled service..."
limactl shell "$LIMA_VM" sudo systemctl enable --now tailscaled

echo ""
echo "Joining tailnet..."
# --hostname forces the Magic DNS name to match what IngressRoutes expect
# --accept-routes picks up any subnet routes advertised by other nodes (e.g. your laptop)
limactl shell "$LIMA_VM" sudo tailscale up \
  --auth-key="${AUTH_KEY}" \
  --hostname="lima-byzantion" \
  --accept-routes

echo ""
echo "Waiting for Tailscale to come up..."
for i in {1..20}; do
  TS_IP=$(limactl shell "$LIMA_VM" tailscale ip -4 2>/dev/null || true)
  if [ -n "$TS_IP" ]; then
    break
  fi
  sleep 2
done

if [ -z "$TS_IP" ]; then
  echo -e "${RED}Tailscale did not get an IP within timeout${NC}"
  exit 1
fi

TS_NAME=$(limactl shell "$LIMA_VM" tailscale status --json 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['Self']['DNSName'].rstrip('.'))" 2>/dev/null || echo "lima-byzantion")

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Done!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "  Tailscale IP : ${TS_IP}"
echo "  Magic DNS    : ${TS_NAME}"
echo ""
echo "Access from your phone:"
echo "  http://${TS_NAME}:8080/"
echo "  http://${TS_NAME}:8080/homeros/v1"
