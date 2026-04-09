#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────
#  My PaaS — Add a worker node to the Swarm cluster
#
#  Run on the WORKER machine:
#    curl -fsSL .../add-worker.sh | bash -s -- --manager-ip <IP> --token <TOKEN>
# ─────────────────────────────────────────────────────────────────
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
log()  { echo -e "${GREEN}[my-paas]${NC} $*"; }
warn() { echo -e "${YELLOW}[my-paas]${NC} $*"; }
err()  { echo -e "${RED}[my-paas]${NC} $*" >&2; }

MANAGER_IP=""
TOKEN=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --manager-ip) MANAGER_IP="$2"; shift 2 ;;
    --token)      TOKEN="$2";      shift 2 ;;
    -h|--help)
      echo "Usage: add-worker.sh --manager-ip <IP> --token <TOKEN>"
      echo ""
      echo "Get the join token from the manager:"
      echo "  docker swarm join-token worker -q"
      exit 0
      ;;
    *) err "Unknown option: $1"; exit 1 ;;
  esac
done

if [[ -z "$MANAGER_IP" || -z "$TOKEN" ]]; then
  err "Both --manager-ip and --token are required."
  err "Run: add-worker.sh --manager-ip <MANAGER_IP> --token <SWARM_TOKEN>"
  exit 1
fi

if [[ $EUID -ne 0 ]]; then
  err "Run as root (sudo)."
  exit 1
fi

# ── Install Docker if needed ────────────────────────────────────
if ! command -v docker &>/dev/null; then
  log "Installing Docker..."
  . /etc/os-release

  apt-get update -y
  apt-get install -y ca-certificates curl gnupg lsb-release

  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL "https://download.docker.com/linux/${ID}/gpg" \
    | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg

  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
    https://download.docker.com/linux/${ID} \
    $(lsb_release -cs) stable" \
    | tee /etc/apt/sources.list.d/docker.list > /dev/null

  apt-get update -y
  apt-get install -y docker-ce docker-ce-cli containerd.io \
                     docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
  log "Docker installed."
fi

# ── Configure daemon ────────────────────────────────────────────
if [[ ! -f /etc/docker/daemon.json ]] || ! grep -q "mypaas-registry" /etc/docker/daemon.json 2>/dev/null || grep -q "live-restore" /etc/docker/daemon.json 2>/dev/null; then
  log "Configuring Docker daemon..."
  mkdir -p /etc/docker
  cat > /etc/docker/daemon.json <<'EOF'
{
  "insecure-registries": ["mypaas-registry:5000", "127.0.0.1:5000", "localhost:5000"],
  "log-driver": "json-file",
  "log-opts": { "max-size": "10m", "max-file": "3" }
}
EOF
  systemctl restart docker
fi

# ── Open firewall ports ─────────────────────────────────────────
if command -v ufw &>/dev/null; then
  log "Configuring firewall..."
  ufw allow 22/tcp   2>/dev/null || true
  ufw allow 2377/tcp 2>/dev/null || true
  ufw allow 7946     2>/dev/null || true
  ufw allow 4789/udp 2>/dev/null || true
  ufw reload 2>/dev/null || true
fi

# ── Join Swarm ──────────────────────────────────────────────────
local_state=$(docker info --format '{{.Swarm.LocalNodeState}}' 2>/dev/null)
node_id=$(docker info --format '{{.Swarm.NodeID}}' 2>/dev/null)

if [[ "$local_state" == "active" && -n "$node_id" ]]; then
  warn "This node is already part of a Swarm."
  echo "  Node ID: $node_id"
else
  # Leave any stale swarm state first
  docker swarm leave --force 2>/dev/null || true

  log "Joining Swarm cluster at $MANAGER_IP..."
  if docker swarm join --token "$TOKEN" "${MANAGER_IP}:2377" 2>&1; then
    log "Successfully joined the Swarm!"
  else
    # Docker may report a timeout but still join in the background
    warn "Join command returned non-zero. Waiting for background join..."
    local tries=0
    while [[ $tries -lt 15 ]]; do
      sleep 2
      local state
      state=$(docker info --format '{{.Swarm.LocalNodeState}}' 2>/dev/null)
      if [[ "$state" == "active" ]]; then
        log "Successfully joined the Swarm! (background)"
        break
      fi
      tries=$((tries + 1))
    done
    if [[ $tries -ge 15 ]]; then
      err "Failed to join the Swarm after waiting 30s."
      err "Check network connectivity to $MANAGER_IP:2377"
      exit 1
    fi
  fi
fi

echo ""
log "Worker node is ready."
log "Verify from the manager:  docker node ls"
echo ""
