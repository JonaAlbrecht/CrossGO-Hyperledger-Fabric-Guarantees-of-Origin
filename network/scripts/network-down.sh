#!/bin/bash
# network-down.sh — Tear down the GO Platform network
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"

echo "====== Stopping GO Platform Network ======"

docker compose -f "$NETWORK_DIR/docker/docker-compose-consumer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-producer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-issuer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-orderer.yaml" down --volumes --remove-orphans 2>/dev/null || true
docker compose -f "$NETWORK_DIR/docker/docker-compose-ca.yaml" down --volumes --remove-orphans 2>/dev/null || true

# Remove chaincode containers and images
docker rm -f $(docker ps -aq --filter "name=dev-peer*") 2>/dev/null || true
docker rmi -f $(docker images -q "dev-peer*") 2>/dev/null || true

echo "====== GO Platform Network Stopped ======"
