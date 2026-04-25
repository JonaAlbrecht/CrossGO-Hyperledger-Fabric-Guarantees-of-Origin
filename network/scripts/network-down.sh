#!/bin/bash
# network-down.sh — v10.1: Tear down both channels
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
cd "$NETWORK_DIR/docker"

for f in docker-compose-orderer.yaml docker-compose-eissuer.yaml docker-compose-hissuer.yaml \
          docker-compose-eproducer.yaml docker-compose-hproducer.yaml \
          docker-compose-ebuyer.yaml docker-compose-hbuyer.yaml; do
  [ -f "$f" ] && docker compose -f "$f" down --volumes --remove-orphans 2>/dev/null || true
done

# Remove chaincode containers and images
docker ps -a --format "{{.Names}}" | grep "^dev-peer" | xargs -r docker rm -f
docker images --format "{{.Repository}}:{{.Tag}}" | grep "^dev-peer" | xargs -r docker rmi -f

echo "[OK] Network down"
