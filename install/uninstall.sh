#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────
#  My PaaS — Uninstaller
#  Usage:  bash uninstall.sh [--purge]
#
#  --purge   Also remove Docker volumes (ALL DATA WILL BE LOST)
# ─────────────────────────────────────────────────────────────────
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BOLD='\033[1m'; NC='\033[0m'

log()  { echo -e "${GREEN}[my-paas]${NC} $*"; }
warn() { echo -e "${YELLOW}[my-paas]${NC} $*"; }
err()  { echo -e "${RED}[my-paas]${NC} $*" >&2; }

PURGE=false
[[ "${1:-}" == "--purge" ]] && PURGE=true

if [[ $EUID -ne 0 ]]; then
  err "This script must be run as root (use sudo)."
  exit 1
fi

echo ""
echo -e "${RED}${BOLD}  ⚠  My PaaS Uninstaller${NC}"
echo ""
if [[ "$PURGE" == true ]]; then
  warn "PURGE mode: Docker volumes and all data will be PERMANENTLY DELETED."
fi

read -rp "Are you sure you want to uninstall My PaaS? [y/N] " confirm
if [[ "$confirm" != [yY] ]]; then
  echo "Aborted."
  exit 0
fi

# ── Stop Swarm stack ────────────────────────────────────────────
if docker stack ls 2>/dev/null | grep -q mypaas; then
  log "Removing Swarm stack 'mypaas'..."
  docker stack rm mypaas
  sleep 5
fi

# ── Stop Compose ────────────────────────────────────────────────
for dir in /opt/mypaas .; do
  if [[ -f "$dir/deploy/docker-compose.yml" ]]; then
    log "Stopping Compose services in $dir..."
    docker compose -f "$dir/deploy/docker-compose.yml" down 2>/dev/null || true
    break
  fi
done

# ── Remove systemd service ─────────────────────────────────────
if [[ -f /etc/systemd/system/mypaas.service ]]; then
  log "Removing systemd service..."
  systemctl stop mypaas.service 2>/dev/null || true
  systemctl disable mypaas.service 2>/dev/null || true
  rm -f /etc/systemd/system/mypaas.service
  systemctl daemon-reload
fi

# ── Remove overlay network ──────────────────────────────────────
if docker network ls --format '{{.Name}}' 2>/dev/null | grep -q '^mypaas-network$'; then
  log "Removing mypaas-network..."
  docker network rm mypaas-network 2>/dev/null || true
fi

# ── Purge volumes ───────────────────────────────────────────────
if [[ "$PURGE" == true ]]; then
  log "Removing Docker volumes..."
  for vol in mypaas-data mypaas-builds traefik-certs \
             postgres-data redis-data prometheus-data grafana-data; do
    full="${vol}"
    # Try both with and without stack prefix
    docker volume rm "$full" 2>/dev/null || true
    docker volume rm "mypaas_${full}" 2>/dev/null || true
    docker volume rm "deploy_${full}" 2>/dev/null || true
  done
  log "Volumes removed."
fi

# ── Remove images ───────────────────────────────────────────────
log "Removing My PaaS Docker images..."
docker rmi deploy-mypaas:latest 2>/dev/null || true
docker rmi mypaas:enterprise 2>/dev/null || true

# ── Clean up install dir ────────────────────────────────────────
if [[ -d /opt/mypaas ]]; then
  read -rp "Remove /opt/mypaas source code? [y/N] " rm_src
  if [[ "$rm_src" == [yY] ]]; then
    rm -rf /opt/mypaas
    log "Source code removed."
  fi
fi

echo ""
log "My PaaS has been uninstalled."
if [[ "$PURGE" != true ]]; then
  warn "Docker volumes were preserved. Run with --purge to remove all data."
fi
echo ""
