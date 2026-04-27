#!/bin/bash
# Fix gossip configuration for all peers
# This script updates the base.yaml and peer docker-compose files to enable gossip debug logging
# and fix bootstrap peer configuration

REPO_DIR="/root/hlf-go/repo/network"

echo "========================================="
echo "FIXING GOSSIP CONFIGURATION"
echo "========================================="
echo ""

# 1. Update base.yaml to enable gossip debug logging
echo ">>> Updating base.yaml with gossip debug logging..."
sed -i 's/FABRIC_LOGGING_SPEC=INFO/FABRIC_LOGGING_SPEC=INFO:gossip=DEBUG:gossip.privdata=DEBUG/' "$REPO_DIR/base.yaml"

# 2. Fix bootstrap peers - each peer should bootstrap from anchor peers, not itself
echo ">>> Fixing bootstrap peers in docker-compose files..."

# Electricity channel peers should bootstrap from eissuer (anchor peer)
echo "  - Fixing eproducer1 bootstrap..."
sed -i 's/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.eproducer1.go-platform.com:9051/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.eissuer.go-platform.com:7051/' "$REPO_DIR/docker/docker-compose-eproducer.yaml"

echo "  - Fixing ebuyer1 bootstrap..."
sed -i 's/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.ebuyer1.go-platform.com:13051/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.eissuer.go-platform.com:7051/' "$REPO_DIR/docker/docker-compose-ebuyer.yaml"

# Hydrogen channel peers should bootstrap from hissuer (anchor peer)
echo "  - Fixing hproducer1 bootstrap..."
sed -i 's/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.hproducer1.go-platform.com:11051/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.hissuer.go-platform.com:8051/' "$REPO_DIR/docker/docker-compose-hproducer.yaml"

echo "  - Fixing hbuyer1 bootstrap..."
sed -i 's/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.hbuyer1.go-platform.com:14051/CORE_PEER_GOSSIP_BOOTSTRAP=peer0.hissuer.go-platform.com:8051/' "$REPO_DIR/docker/docker-compose-hbuyer.yaml"

echo ""
echo ">>> Configuration updates complete!"
echo ""
echo "Verify changes:"
echo "  base.yaml: FABRIC_LOGGING_SPEC should include gossip=DEBUG:gossip.privdata=DEBUG"
echo "  docker-compose files: GOSSIP_BOOTSTRAP should point to anchor peers"
echo ""
