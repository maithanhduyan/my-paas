#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────
#  My PaaS — One-line installer for Ubuntu / Debian
#
#  Usage:
#    curl -fsSL https://raw.githubusercontent.com/<repo>/main/install/install.sh | bash
#    # or
#    bash install.sh [--domain example.com] [--email admin@example.com] [--enterprise] [--swarm]
#
#  What it does:
#    1. Installs Docker Engine + Compose plugin
#    2. Configures insecure local registry
#    3. Optionally initialises Docker Swarm
#    4. Clones the repo (or uses the local copy)
#    5. Builds the Docker image
#    6. Deploys via docker compose (standalone) or docker stack (swarm)
#    7. Opens firewall ports
#    8. Prints access URL + default credentials
# ─────────────────────────────────────────────────────────────────
set -euo pipefail

# ── Colours ──────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

log()  { echo -e "${GREEN}[my-paas]${NC} $*"; }
warn() { echo -e "${YELLOW}[my-paas]${NC} $*"; }
err()  { echo -e "${RED}[my-paas]${NC} $*" >&2; }
banner() {
  echo ""
  echo -e "${CYAN}${BOLD}"
  echo "  ╔══════════════════════════════════════════╗"
  echo "  ║           My PaaS  Installer             ║"
  echo "  ║      Self-hosted PaaS on Docker          ║"
  echo "  ╚══════════════════════════════════════════╝"
  echo -e "${NC}"
}

# ── Defaults ─────────────────────────────────────────────────────
DOMAIN=""
ACME_EMAIL="admin@example.com"
ENTERPRISE=false
SWARM=false
INSTALL_DIR="/opt/mypaas"
DATA_DIR="/data"
REPO_URL="https://github.com/your-org/my-paas.git"
BRANCH="main"
SKIP_DOCKER=false

# ── Parse arguments ──────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --domain)      DOMAIN="$2";      shift 2 ;;
    --email)       ACME_EMAIL="$2";  shift 2 ;;
    --enterprise)  ENTERPRISE=true;  shift   ;;
    --swarm)       SWARM=true;       shift   ;;
    --branch)      BRANCH="$2";     shift 2 ;;
    --skip-docker) SKIP_DOCKER=true; shift   ;;
    --help|-h)
      echo "Usage: install.sh [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --domain DOMAIN     Set the domain for Traefik (e.g. mypaas.example.com)"
      echo "  --email  EMAIL      Email for Let's Encrypt certificates"
      echo "  --enterprise        Deploy enterprise stack (PostgreSQL + Redis + Prometheus + Grafana)"
      echo "  --swarm             Initialise Docker Swarm and deploy as a stack"
      echo "  --branch BRANCH     Git branch to clone (default: main)"
      echo "  --skip-docker       Skip Docker installation (if already installed)"
      echo "  -h, --help          Show this help"
      exit 0
      ;;
    *) err "Unknown option: $1"; exit 1 ;;
  esac
done

# ── Pre-flight checks ───────────────────────────────────────────
banner

if [[ $EUID -ne 0 ]]; then
  err "This script must be run as root (use sudo)."
  exit 1
fi

# Detect distro
if [[ ! -f /etc/os-release ]]; then
  err "Cannot detect OS. Only Ubuntu/Debian are supported."
  exit 1
fi

. /etc/os-release
case "$ID" in
  ubuntu|debian) log "Detected: $PRETTY_NAME" ;;
  *)
    err "Unsupported OS: $ID. Only Ubuntu/Debian are supported."
    exit 1
    ;;
esac

# ── 1. Install Docker ───────────────────────────────────────────
install_docker() {
  if command -v docker &>/dev/null && [[ "$SKIP_DOCKER" == true ]]; then
    log "Docker already installed, skipping."
    return
  fi

  if command -v docker &>/dev/null; then
    local ver
    ver=$(docker version --format '{{.Server.Version}}' 2>/dev/null || echo "unknown")
    log "Docker already installed (v${ver})."
  else
    log "Installing Docker Engine..."

    # Remove old packages
    apt-get remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true

    # Install prerequisites
    apt-get update -y
    apt-get install -y ca-certificates curl gnupg lsb-release

    # Add Docker GPG key
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL "https://download.docker.com/linux/${ID}/gpg" \
      | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg

    # Add Docker repository
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
      https://download.docker.com/linux/${ID} \
      $(lsb_release -cs) stable" \
      | tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update -y
    apt-get install -y docker-ce docker-ce-cli containerd.io \
                       docker-buildx-plugin docker-compose-plugin

    systemctl enable --now docker
    log "Docker installed successfully."
  fi

  # Verify
  docker version --format 'Docker {{.Server.Version}}' 2>/dev/null \
    || { err "Docker installation failed."; exit 1; }
  docker compose version 2>/dev/null \
    || { err "Docker Compose plugin not found."; exit 1; }
}

# ── 2. Configure Docker daemon ──────────────────────────────────
configure_docker() {
  local daemon_file="/etc/docker/daemon.json"
  local needs_restart=false

  # Merge insecure-registries if not already present, or fix live-restore for Swarm
  if [[ -f "$daemon_file" ]]; then
    if ! grep -q "mypaas-registry:5000" "$daemon_file" 2>/dev/null; then
      needs_restart=true
    fi
    if grep -q "live-restore" "$daemon_file" 2>/dev/null; then
      needs_restart=true
    fi
  else
    needs_restart=true
  fi

  if [[ "$needs_restart" == true ]]; then
    log "Configuring Docker daemon (insecure local registry)..."
    mkdir -p /etc/docker
    cat > "$daemon_file" <<'DAEMON'
{
  "insecure-registries": ["mypaas-registry:5000", "127.0.0.1:5000", "localhost:5000"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
DAEMON
    systemctl restart docker
    log "Docker daemon configured."
  else
    log "Docker daemon already configured."
  fi
}

# ── 3. Initialise Swarm (optional) ──────────────────────────────
init_swarm() {
  if [[ "$SWARM" != true ]]; then
    return
  fi

  if docker info --format '{{.Swarm.LocalNodeState}}' 2>/dev/null | grep -q "active"; then
    log "Docker Swarm already active."
  else
    log "Initialising Docker Swarm..."

    # Try to detect the best IP for the advertise address
    local advertise_addr=""
    # Prefer the primary non-loopback IPv4
    advertise_addr=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{print $7; exit}' || true)
    if [[ -z "$advertise_addr" ]]; then
      advertise_addr=$(hostname -I | awk '{print $1}')
    fi

    docker swarm init --advertise-addr "$advertise_addr" 2>/dev/null || true
    log "Swarm initialised. Advertise address: $advertise_addr"

    local token
    token=$(docker swarm join-token worker -q 2>/dev/null)
    log "Worker join token: $token"
    log "Add workers with:  docker swarm join --token $token $advertise_addr:2377"
  fi

  # Create overlay network
  if ! docker network ls --format '{{.Name}}' | grep -q '^mypaas-network$'; then
    docker network create --driver overlay --attachable mypaas-network
    log "Created overlay network: mypaas-network"
  fi
}

# ── 4. Get source code ──────────────────────────────────────────
get_source() {
  # If running from inside the repo already, use it
  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  local repo_root
  repo_root="$(dirname "$script_dir")"

  if [[ -f "$repo_root/deploy/Dockerfile" && -f "$repo_root/deploy/docker-stack.yml" ]]; then
    log "Using local source at: $repo_root"
    INSTALL_DIR="$repo_root"
    return
  fi

  # Otherwise clone
  if [[ -d "$INSTALL_DIR/.git" ]]; then
    log "Updating existing installation at $INSTALL_DIR..."
    cd "$INSTALL_DIR"
    git fetch origin "$BRANCH"
    git reset --hard "origin/$BRANCH"
  else
    log "Cloning My PaaS to $INSTALL_DIR..."
    apt-get install -y git 2>/dev/null || true
    git clone --branch "$BRANCH" --depth 1 "$REPO_URL" "$INSTALL_DIR"
  fi

  cd "$INSTALL_DIR"
  log "Source ready at $INSTALL_DIR"
}

# ── 5. Build Docker image ───────────────────────────────────────
build_image() {
  cd "$INSTALL_DIR"

  log "Building My PaaS Docker image..."
  docker compose -f deploy/docker-compose.yml build

  if [[ "$ENTERPRISE" == true ]]; then
    log "Tagging enterprise image..."
    docker tag deploy-mypaas:latest mypaas:enterprise
  fi

  log "Image built: deploy-mypaas:latest"
}

# ── 6. Generate environment file ────────────────────────────────
generate_env() {
  local env_file="$INSTALL_DIR/.env"

  if [[ -f "$env_file" ]]; then
    log "Environment file exists, keeping current values."
    return
  fi

  log "Generating environment file..."

  # Generate secure random secrets
  local jwt_secret
  jwt_secret=$(openssl rand -hex 32 2>/dev/null || head -c 64 /dev/urandom | base64 | tr -d '=/+' | head -c 64)
  local encryption_key
  encryption_key=$(openssl rand -hex 16 2>/dev/null || head -c 32 /dev/urandom | base64 | tr -d '=/+' | head -c 32)
  local pg_password
  pg_password=$(openssl rand -hex 16 2>/dev/null || head -c 32 /dev/urandom | base64 | tr -d '=/+' | head -c 32)
  local grafana_password
  grafana_password=$(openssl rand -base64 12 2>/dev/null || echo "admin")

  cat > "$env_file" <<EOF
# My PaaS configuration — generated $(date -Iseconds)
MYPAAS_DOMAIN=${DOMAIN}
ACME_EMAIL=${ACME_EMAIL}

# Secrets (auto-generated, keep safe)
MYPAAS_SECRET=${jwt_secret}
MYPAAS_ENCRYPTION_KEY=${encryption_key}

# Enterprise (PostgreSQL + Redis)
POSTGRES_PASSWORD=${pg_password}
GRAFANA_PASSWORD=${grafana_password}
EOF

  chmod 600 "$env_file"
  log "Environment file created: $env_file"
}

# ── 7. Create data directories ──────────────────────────────────
prepare_data() {
  mkdir -p "$DATA_DIR/builds" "$DATA_DIR/backups"
  log "Data directories ready: $DATA_DIR"
}

# ── 8. Deploy ───────────────────────────────────────────────────
deploy() {
  cd "$INSTALL_DIR"

  # Export env vars
  if [[ -f .env ]]; then
    set -a; source .env; set +a
  fi

  if [[ "$SWARM" == true ]]; then
    deploy_swarm
  else
    deploy_compose
  fi
}

deploy_compose() {
  log "Deploying with Docker Compose (standalone mode)..."

  # Stop existing if running
  docker compose -f deploy/docker-compose.yml down 2>/dev/null || true

  docker compose -f deploy/docker-compose.yml up -d

  log "Compose deployment complete."
}

deploy_swarm() {
  local stack_file="deploy/docker-stack.yml"
  if [[ "$ENTERPRISE" == true ]]; then
    stack_file="deploy/docker-stack.enterprise.yml"
  fi

  log "Deploying with Docker Swarm ($stack_file)..."

  # Remove old stack if exists
  docker stack rm mypaas 2>/dev/null || true
  # Wait for network cleanup
  sleep 5

  # Ensure network exists
  docker network create --driver overlay --attachable mypaas-network 2>/dev/null || true

  docker stack deploy -c "$stack_file" mypaas

  log "Swarm stack deployed."
}

# ── 9. Firewall ─────────────────────────────────────────────────
configure_firewall() {
  if ! command -v ufw &>/dev/null; then
    log "UFW not found, skipping firewall configuration."
    return
  fi

  log "Configuring firewall rules (UFW)..."

  # Only add rules — do NOT enable UFW automatically.
  # Docker manages iptables directly; force-enabling UFW can break
  # Swarm overlay networking and published ports.

  # Essential ports
  ufw allow 22/tcp   comment 'SSH'           2>/dev/null || true
  ufw allow 80/tcp   comment 'HTTP'          2>/dev/null || true
  ufw allow 443/tcp  comment 'HTTPS'         2>/dev/null || true
  ufw allow 8080/tcp comment 'My PaaS API'   2>/dev/null || true

  # Swarm ports (if applicable)
  if [[ "$SWARM" == true ]]; then
    ufw allow 2377/tcp comment 'Swarm mgmt'  2>/dev/null || true
    ufw allow 7946     comment 'Swarm gossip' 2>/dev/null || true
    ufw allow 4789/udp comment 'VXLAN overlay' 2>/dev/null || true
  fi

  ufw reload 2>/dev/null || true
  log "Firewall configured."
}

# ── 10. Wait for healthy ────────────────────────────────────────
wait_healthy() {
  log "Waiting for My PaaS to start..."
  local max_wait=90
  local waited=0

  while [[ $waited -lt $max_wait ]]; do
    # Try curl first
    if curl -sf --connect-timeout 3 --max-time 5 http://localhost:8080/api/health &>/dev/null; then
      echo ""
      log "My PaaS is healthy!"
      return 0
    fi

    # Fallback: check Docker container/service health status
    if [[ "$SWARM" == true ]]; then
      local task_state
      task_state=$(docker service ps mypaas_mypaas --format '{{.CurrentState}}' --filter 'desired-state=running' 2>/dev/null | head -1)
      if [[ "$task_state" == Running* ]]; then
        # Check container health via docker inspect
        local cid
        cid=$(docker ps -q --filter "label=com.docker.swarm.service.name=mypaas_mypaas" 2>/dev/null | head -1)
        if [[ -n "$cid" ]]; then
          local health
          health=$(docker inspect --format='{{.State.Health.Status}}' "$cid" 2>/dev/null)
          if [[ "$health" == "healthy" ]]; then
            echo ""
            log "My PaaS is healthy! (verified via container health check)"
            return 0
          fi
        fi
      fi
    fi

    sleep 2
    waited=$((waited + 2))
    printf "."
  done

  echo ""
  warn "My PaaS did not become healthy within ${max_wait}s."
  warn "Check logs with:  docker logs mypaas-server"
  if [[ "$SWARM" == true ]]; then
    warn "  or:  docker service logs mypaas_mypaas"
  fi
  return 1
}

# ── 11. Create systemd service (for auto-restart on boot) ───────
create_systemd_service() {
  # Only for compose mode — Swarm handles auto-restart natively
  if [[ "$SWARM" == true ]]; then
    return
  fi

  local service_file="/etc/systemd/system/mypaas.service"
  if [[ -f "$service_file" ]]; then
    log "Systemd service already exists."
    return
  fi

  log "Creating systemd service for auto-start on boot..."

  cat > "$service_file" <<EOF
[Unit]
Description=My PaaS Platform
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=${INSTALL_DIR}
ExecStart=/usr/bin/docker compose -f deploy/docker-compose.yml up -d
ExecStop=/usr/bin/docker compose -f deploy/docker-compose.yml down
TimeoutStartSec=120

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable mypaas.service
  log "Systemd service created and enabled."
}

# ── Print summary ───────────────────────────────────────────────
print_summary() {
  local ip
  ip=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{print $7; exit}' || hostname -I | awk '{print $1}')

  echo ""
  echo -e "${GREEN}${BOLD}════════════════════════════════════════════════════${NC}"
  echo -e "${GREEN}${BOLD}  My PaaS installed successfully!${NC}"
  echo -e "${GREEN}${BOLD}════════════════════════════════════════════════════${NC}"
  echo ""
  echo -e "  ${BOLD}Dashboard:${NC}   http://${ip}:8080"
  if [[ -n "$DOMAIN" ]]; then
    echo -e "  ${BOLD}Domain:${NC}      https://${DOMAIN}"
  fi
  echo -e "  ${BOLD}Credentials:${NC} admin / admin"
  echo -e "  ${BOLD}Mode:${NC}        $(if [[ "$SWARM" == true ]]; then echo "Swarm cluster"; else echo "Standalone"; fi)"
  if [[ "$ENTERPRISE" == true ]]; then
    echo -e "  ${BOLD}Edition:${NC}     Enterprise (PostgreSQL + Redis + Prometheus + Grafana)"
  else
    echo -e "  ${BOLD}Edition:${NC}     Community (SQLite)"
  fi
  echo ""
  echo -e "  ${BOLD}Install dir:${NC} ${INSTALL_DIR}"
  echo -e "  ${BOLD}Data dir:${NC}    ${DATA_DIR}"
  echo -e "  ${BOLD}Env file:${NC}    ${INSTALL_DIR}/.env"
  echo ""
  echo -e "  ${CYAN}Useful commands:${NC}"
  if [[ "$SWARM" == true ]]; then
    echo "    docker service ls                          # List services"
    echo "    docker service logs mypaas_mypaas          # View logs"
    echo "    docker stack rm mypaas                     # Stop everything"
    echo "    docker stack deploy -c deploy/docker-stack.yml mypaas  # Redeploy"
  else
    echo "    docker compose -f deploy/docker-compose.yml logs -f    # View logs"
    echo "    docker compose -f deploy/docker-compose.yml restart    # Restart"
    echo "    docker compose -f deploy/docker-compose.yml down       # Stop"
    echo "    systemctl restart mypaas                               # Restart via systemd"
  fi
  echo ""
  echo -e "  ${YELLOW}⚠  Change the default password immediately after first login!${NC}"
  echo ""
}

# ── Main ─────────────────────────────────────────────────────────
main() {
  log "Starting My PaaS installation..."
  log "Options: domain=${DOMAIN:-<none>} email=${ACME_EMAIL} enterprise=${ENTERPRISE} swarm=${SWARM}"
  echo ""

  install_docker
  configure_docker
  init_swarm
  get_source
  build_image
  generate_env
  prepare_data
  deploy
  create_systemd_service

  if wait_healthy; then
    configure_firewall
    print_summary
  else
    configure_firewall
    echo ""
    warn "Installation completed but the service is not healthy yet."
    warn "It may still be starting up. Try: curl http://localhost:8080/api/health"
    print_summary
  fi
}

main "$@"
