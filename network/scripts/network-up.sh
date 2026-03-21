#!/bin/bash
# network-up.sh — Bring up the GO Platform network
# 4 orgs: Issuer, Electricity Producer, Hydrogen Producer, Buyer
# 4-node Raft orderer cluster
# Usage: ./network-up.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(dirname "$NETWORK_DIR")"
CHANNEL_NAME="goplatformchannel"

# Fabric binaries
export PATH="${REPO_DIR}/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$NETWORK_DIR"

# Org definitions: name mspID peerPort
ORGS=(
  "issuer1:issuer1MSP:7051"
  "eproducer1:eproducer1MSP:9051"
  "hproducer1:hproducer1MSP:11051"
  "buyer1:buyer1MSP:13051"
)

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${CYAN}[→]${NC} $1"; }
fail() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "=============================================="
echo "  GO Platform Network — Starting"
echo "=============================================="
echo ""

# ─── Step 0: Clean previous state ───────────────────────────────────────────
info "Cleaning previous state..."
"$SCRIPT_DIR/network-down.sh" 2>/dev/null || true
sudo rm -rf "$NETWORK_DIR/organizations" "$NETWORK_DIR/channel-artifacts"
log "Clean"

# ─── Step 1: Generate crypto material ────────────────────────────────────────
info "Generating crypto material with cryptogen..."
cryptogen generate --config="$NETWORK_DIR/crypto-config.yaml" --output="$NETWORK_DIR/organizations"
log "Crypto material generated"

# ─── Step 2: Generate channel genesis block ──────────────────────────────────
info "Generating channel genesis block..."
mkdir -p "$NETWORK_DIR/channel-artifacts"
configtxgen -profile GOPlatformChannel \
  -outputBlock "$NETWORK_DIR/channel-artifacts/${CHANNEL_NAME}.block" \
  -channelID "$CHANNEL_NAME"
log "Genesis block: channel-artifacts/${CHANNEL_NAME}.block"

# ─── Step 3: Start orderer cluster ──────────────────────────────────────────
info "Starting orderer cluster (4-node Raft)..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-orderer.yaml" up -d
sleep 3
log "Orderers started"

# ─── Step 4: Start all peers ────────────────────────────────────────────────
info "Starting peers and CouchDB instances..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-issuer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-eproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-hproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-buyer.yaml" up -d
sleep 5
log "All peers started"

# ─── Step 5: Join orderers to channel ────────────────────────────────────────
info "Joining orderers to channel..."
for i in 1 2 3 4; do
  ADMIN_PORT=$((9442 + i))
  ORDERER_TLS="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/orderers/orderer${i}.go-platform.com/tls"
  osnadmin channel join \
    --channelID "$CHANNEL_NAME" \
    --config-block "$NETWORK_DIR/channel-artifacts/${CHANNEL_NAME}.block" \
    -o "localhost:${ADMIN_PORT}" \
    --ca-file "$ORDERER_TLS/ca.crt" \
    --client-cert "$ORDERER_TLS/server.crt" \
    --client-key "$ORDERER_TLS/server.key"
  log "  orderer${i} joined channel"
done

# ─── Step 6: Join peers to channel ──────────────────────────────────────────
info "Joining peers to channel..."
ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  peer channel join -b "$NETWORK_DIR/channel-artifacts/${CHANNEL_NAME}.block"
  log "  peer0.${org} joined channel"
done

# ─── Step 7: Set anchor peers ───────────────────────────────────────────────
info "Setting anchor peers..."
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"

  # Fetch current config
  peer channel fetch config "$NETWORK_DIR/channel-artifacts/config_block.pb" \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -c "$CHANNEL_NAME" --tls --cafile "$ORDERER_CA" 2>/dev/null

  # Decode to JSON
  configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/config_block.pb" \
    --type common.Block 2>/dev/null \
    | jq '.data.data[0].payload.data.config' > "$NETWORK_DIR/channel-artifacts/${org}_config.json"

  # Modify — add anchor peer
  jq --arg msp "$msp" --arg host "peer0.${org}.go-platform.com" --argjson port "$port" \
    '.channel_group.groups.Application.groups[$msp].values += {
      "AnchorPeers": {
        "mod_policy": "Admins",
        "value": {"anchor_peers": [{"host": $host, "port": $port}]},
        "version": "0"
      }
    }' "$NETWORK_DIR/channel-artifacts/${org}_config.json" \
    > "$NETWORK_DIR/channel-artifacts/${org}_modified_config.json"

  # Encode both
  configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${org}_config.json" \
    --type common.Config --output "$NETWORK_DIR/channel-artifacts/${org}_config.pb" 2>/dev/null
  configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${org}_modified_config.json" \
    --type common.Config --output "$NETWORK_DIR/channel-artifacts/${org}_modified_config.pb" 2>/dev/null

  # Compute update delta
  if configtxlator compute_update --channel_id "$CHANNEL_NAME" \
    --original "$NETWORK_DIR/channel-artifacts/${org}_config.pb" \
    --updated "$NETWORK_DIR/channel-artifacts/${org}_modified_config.pb" \
    --output "$NETWORK_DIR/channel-artifacts/${org}_anchor_update.pb" 2>/dev/null; then

    # Wrap in envelope
    configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/${org}_anchor_update.pb" \
      --type common.ConfigUpdate 2>/dev/null \
      | jq '{"payload":{"header":{"channel_header":{"channel_id":"'"$CHANNEL_NAME"'","type":2}},"data":{"config_update":.}}}' \
      | configtxlator proto_encode --type common.Envelope \
        --output "$NETWORK_DIR/channel-artifacts/${org}_anchor_update_tx.pb" 2>/dev/null

    # Submit
    peer channel update -f "$NETWORK_DIR/channel-artifacts/${org}_anchor_update_tx.pb" \
      -c "$CHANNEL_NAME" -o localhost:7050 \
      --ordererTLSHostnameOverride orderer1.go-platform.com \
      --tls --cafile "$ORDERER_CA"
    log "  Anchor peer set for $msp"
  else
    log "  Anchor peer already set for $msp (no update needed)"
  fi
done

# ─── Done ────────────────────────────────────────────────────────────────────
echo ""
echo "=============================================="
echo "  GO Platform Network — RUNNING"
echo "=============================================="
echo ""
echo "Channel: $CHANNEL_NAME"
echo ""
echo "Peers:"
echo "  Issuer:            peer0.issuer1.go-platform.com:7051"
echo "  E-Producer:        peer0.eproducer1.go-platform.com:9051"
echo "  H-Producer:        peer0.hproducer1.go-platform.com:11051"
echo "  Buyer:             peer0.buyer1.go-platform.com:13051"
echo ""
echo "Orderers:            7050, 8050, 9050, 10050"
echo "CouchDB:             5984, 7984, 9984, 11984"
echo ""
echo "Next: ./deploy-chaincode.sh"
