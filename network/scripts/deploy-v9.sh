#!/bin/bash
# deploy-v9.sh — Deploy GO lifecycle chaincode v9.0.0
#
# Single channel (goplatformchannel) with unified collection config (all 4 orgs).
# 4 energy carriers: electricity, hydrogen, biogas, heating/cooling
# 11 chaincode contracts including new HeatingCoolingContract
# Dual-issuer bridge consensus (ADR-031 v2)
# Role rename: consumer → buyer
#
# Usage: ./deploy-v9.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(dirname "$NETWORK_DIR")"

CHAINCODE_NAME="golifecycle"
CHAINCODE_VERSION="9.0.0"
CHAINCODE_SEQUENCE=1
CHAINCODE_PATH="$REPO_DIR/chaincode"
CHANNEL_NAME="goplatformchannel"

COLLECTIONS_CONFIG="$REPO_DIR/collections/collection-config.json"

export PATH="${REPO_DIR}/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$NETWORK_DIR"

ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

ALL_ORGS=(
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

build_peer_conn_args() {
  local args=""
  for entry in "$@"; do
    IFS=':' read -r org msp port <<< "$entry"
    args="$args --peerAddresses localhost:${port}"
    args="$args --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  done
  echo "$args"
}

echo "=============================================="
echo "  Deploying Chaincode: $CHAINCODE_NAME v$CHAINCODE_VERSION"
echo "  Single Channel: $CHANNEL_NAME"
echo "  11 contracts, 4 energy carriers"
echo "=============================================="

# ─── Step 1: Package chaincode ───────────────────────────────────────────────
info "Packaging chaincode..."
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode package "${CHAINCODE_NAME}.tar.gz" \
  --path "$CHAINCODE_PATH" \
  --lang golang \
  --label "${CHAINCODE_NAME}_${CHAINCODE_VERSION}"
log "Packaged: ${CHAINCODE_NAME}.tar.gz"

# ─── Step 2: Install on ALL peers ────────────────────────────────────────────
for entry in "${ALL_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  info "Installing on ${org}..."
  set_peer_env "$org" "$msp" "$port"
  peer lifecycle chaincode install "${CHAINCODE_NAME}.tar.gz"
  log "  Installed on ${org}"
done

# ─── Step 3: Get package ID ─────────────────────────────────────────────────
set_peer_env issuer1 issuer1MSP 7051
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json \
  | jq -r ".installed_chaincodes[] | select(.label==\"${CHAINCODE_NAME}_${CHAINCODE_VERSION}\") | .package_id")
echo "Package ID: $PACKAGE_ID"
[[ -z "$PACKAGE_ID" ]] && fail "No package ID found — install may have failed"

# ─── Step 4: Approve for each org ────────────────────────────────────────────
for entry in "${ALL_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  info "Approving for ${org}..."
  set_peer_env "$org" "$msp" "$port"
  peer lifecycle chaincode approveformyorg \
    -o localhost:7050 \
    --ordererTLSHostnameOverride orderer1.go-platform.com \
    --channelID "$CHANNEL_NAME" \
    --name "$CHAINCODE_NAME" \
    --version "$CHAINCODE_VERSION" \
    --package-id "$PACKAGE_ID" \
    --sequence "$CHAINCODE_SEQUENCE" \
    --collections-config "$COLLECTIONS_CONFIG" \
    --tls --cafile "$ORDERER_CA"
  log "  ${org} approved"
done

# ─── Step 5: Check commit readiness ─────────────────────────────────────────
info "Checking commit readiness..."
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode checkcommitreadiness \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --output json

# ─── Step 6: Commit ─────────────────────────────────────────────────────────
PEER_CONN_ARGS=$(build_peer_conn_args "${ALL_ORGS[@]}")

info "Committing chaincode..."
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode commit \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN_ARGS
log "Chaincode committed"

# Verify
peer lifecycle chaincode querycommitted \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --output json

# ─── Step 7: Initialize roles ───────────────────────────────────────────────
info "Initializing ledger and roles..."
set_peer_env issuer1 issuer1MSP 7051

peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN_ARGS \
  -c '{"function":"device:InitLedger","Args":["issuer1MSP"]}'
sleep 2

# Register org roles (v9: buyer instead of consumer)
for role_entry in "eproducer1MSP:producer" "hproducer1MSP:producer" "buyer1MSP:buyer"; do
  IFS=':' read -r rmsp rrole <<< "$role_entry"
  info "Registering ${rmsp} as ${rrole}..."
  peer chaincode invoke \
    -o localhost:7050 \
    --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" \
    --tls --cafile "$ORDERER_CA" \
    $PEER_CONN_ARGS \
    -c "{\"function\":\"device:RegisterOrgRole\",\"Args\":[\"${rmsp}\",\"${rrole}\"]}"
  sleep 1
done
log "Roles initialized"

# ─── Step 8: Verify version ─────────────────────────────────────────────────
info "Verifying chaincode version..."
peer chaincode query \
  -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" \
  -c '{"function":"admin:GetVersion","Args":[]}'

echo ""
echo "=============================================="
echo "  Chaincode v${CHAINCODE_VERSION} Deployed"
echo "=============================================="
echo ""
echo "Channel:    $CHANNEL_NAME"
echo "Collection: collection-config.json (all 4 orgs)"
echo ""
echo "Orgs:"
echo "  issuer1MSP      → issuer  (German Issuing Authority)"
echo "  eproducer1MSP   → producer (Alpha WindFarm GmbH)"
echo "  hproducer1MSP   → producer (Beta Electrolyser B.V.)"
echo "  buyer1MSP       → buyer   (Gamma-Town EnergySupplier Ltd)"
echo ""
echo "Contracts (11):"
echo "  admin, device, query, transfer, cancellation,"
echo "  conversion, bridge, biogas, hydrogen, heating_cooling, electricity"
echo ""
echo "Energy carriers: electricity, hydrogen, biogas, heating_cooling"
