#!/bin/bash
# network-up.sh — Bring up the GO Platform network
# Usage: ./network-up.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
CHANNEL_NAME="goplatformchannel"

echo "====== Starting GO Platform Network ======"

# 1. Start CAs
echo "--- Starting Certificate Authorities ---"
docker compose -f "$NETWORK_DIR/docker/docker-compose-ca.yaml" up -d
sleep 3

# 2. Generate crypto material (placeholder — implement with fabric-ca-client)
echo "--- Generating crypto material ---"
echo "TODO: Implement crypto material generation using fabric-ca-client"
echo "  - Enroll CA admins"
echo "  - Register and enroll peers, orderers, users"
echo "  - Generate MSP directories"

# 3. Start orderers
echo "--- Starting Orderer Cluster ---"
docker compose -f "$NETWORK_DIR/docker/docker-compose-orderer.yaml" up -d
sleep 5

# 4. Start peers
echo "--- Starting Issuer Peer ---"
docker compose -f "$NETWORK_DIR/docker/docker-compose-issuer.yaml" up -d

echo "--- Starting Producer Peer ---"
docker compose -f "$NETWORK_DIR/docker/docker-compose-producer.yaml" up -d

echo "--- Starting Consumer Peer ---"
docker compose -f "$NETWORK_DIR/docker/docker-compose-consumer.yaml" up -d
sleep 5

# 5. Create channel
echo "--- Creating Channel: $CHANNEL_NAME ---"
echo "TODO: Implement channel creation using osnadmin or configtxgen + peer channel create"
echo "  configtxgen -profile GOPlatformChannel -outputBlock $CHANNEL_NAME.block -channelID $CHANNEL_NAME"
echo "  osnadmin channel join --channelID $CHANNEL_NAME --config-block $CHANNEL_NAME.block -o orderer1.go-platform.com:9443"

# 6. Join peers to channel
echo "--- Joining Peers to Channel ---"
echo "TODO: peer channel join -b $CHANNEL_NAME.block for each org"

# 7. Set anchor peers
echo "--- Setting Anchor Peers ---"
echo "TODO: peer channel update for each org's anchor peer"

echo "====== GO Platform Network Started ======"
echo "Peers:"
echo "  Issuer:   peer0.issuer1.go-platform.com:7051"
echo "  Producer: peer0.producer1.go-platform.com:9051"
echo "  Consumer: peer0.consumer1.go-platform.com:11051"
echo "Orderers:   7050, 8050, 9050, 10050"
echo "CAs:        7054, 8054, 9054, 10054"
