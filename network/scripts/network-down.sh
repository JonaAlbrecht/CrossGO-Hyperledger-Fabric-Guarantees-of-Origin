#!/bin/bash
# network-down.sh — Tear down the GO Platform network
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"

echo "====== Stopping GO Platform Network ======"

docker compose -f "$NETWORK_DIR/docker/docker-compose-buyer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-hproducer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-eproducer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-issuer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-orderer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-ca.yaml" down --volumes --remove-orphans 2>/dev/null || true

# Remove chaincode containers and images
docker rm -f $(docker ps -aq --filter "name=dev-peer*") 2>/dev/null || true
docker rmi -f $(docker images -q "dev-peer*") 2>/dev/null || true

# Remove generated artifacts
sudo rm -rf "$NETWORK_DIR/organizations" "$NETWORK_DIR/channel-artifacts"

echo "====== GO Platform Network Stopped ======"
