#!/bin/bash
# deploy-v7.sh — Upgrade the GO lifecycle chaincode to v7.0
# This installs the new package for v7.0 (sequence 3) with all v6+v7 features:
# - ADR-017: Random commitment salts
# - ADR-018: CEN-EN 16325 field validation
# - ADR-019: State-based endorsement policies
# - ADR-022: Deprecated unpaginated queries
# - ADR-024: Cross-registry bridge contract
# - ADR-027: IoT smart meter attestation
# - ADR-029: External data oracle (ENTSO-E)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(dirname "$NETWORK_DIR")"
CHANNEL_NAME="goplatformchannel"
CHAINCODE_NAME="golifecycle"
CHAINCODE_VERSION="7.0"
CHAINCODE_SEQUENCE=3
CHAINCODE_PATH="$REPO_DIR/chaincode"
COLLECTIONS_CONFIG="$REPO_DIR/collections/collection-config.json"

export PATH="${REPO_DIR}/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$NETWORK_DIR"

ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

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
echo "  GO Platform Chaincode — Upgrade to v7.0"
echo "=============================================="
echo ""

# ─── Step 0: Vendor Go modules ───────────────────────────────────────────
info "Vendoring Go modules..."
cd "$CHAINCODE_PATH"
GO111MODULE=on go mod tidy
GO111MODULE=on go mod vendor
cd "$REPO_DIR"
log "Go modules vendored"

# ─── Step 1: Package v7.0 ────────────────────────────────────────────────
info "Packaging chaincode v${CHAINCODE_VERSION}..."
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode package "${CHAINCODE_NAME}_v7.tar.gz" \
  --path "$CHAINCODE_PATH" \
  --lang golang \
  --label "${CHAINCODE_NAME}_${CHAINCODE_VERSION}"
log "Packaged: ${CHAINCODE_NAME}_v7.tar.gz"

# ─── Step 2: Install on each peer ────────────────────────────────────────
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  info "Installing on ${org}..."
  set_peer_env "$org" "$msp" "$port"
  peer lifecycle chaincode install "${CHAINCODE_NAME}_v7.tar.gz"
  log "Installed on ${org}"
done

# ─── Step 3: Get package ID ──────────────────────────────────────────────
set_peer_env issuer1 issuer1MSP 7051
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | \
  jq -r ".installed_chaincodes[] | select(.label==\"${CHAINCODE_NAME}_${CHAINCODE_VERSION}\") | .package_id")
echo "  Package ID: $PACKAGE_ID"

if [ -z "$PACKAGE_ID" ]; then
  fail "Failed to get package ID — installation may have failed"
fi

# ─── Step 4: Approve for each org ────────────────────────────────────────
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  info "Approving for ${org}..."
  set_peer_env "$org" "$msp" "$port"
  peer lifecycle chaincode approveformyorg \
    -o localhost:7050 \
    --ordererTLSHostnameOverride localhost \
    --channelID "$CHANNEL_NAME" \
    --name "$CHAINCODE_NAME" \
    --version "$CHAINCODE_VERSION" \
    --package-id "$PACKAGE_ID" \
    --sequence "$CHAINCODE_SEQUENCE" \
    --collections-config "$COLLECTIONS_CONFIG" \
    --tls --cafile "$ORDERER_CA"
  log "Approved for ${org}"
done

# ─── Step 5: Check readiness ─────────────────────────────────────────────
info "Checking commit readiness..."
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode checkcommitreadiness \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --output json
log "All orgs approved"

# ─── Step 6: Commit ──────────────────────────────────────────────────────
PEER_CONN_ARGS=""
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  PEER_CONN_ARGS="$PEER_CONN_ARGS --peerAddresses localhost:${port}"
  PEER_CONN_ARGS="$PEER_CONN_ARGS --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
done

info "Committing chaincode v${CHAINCODE_VERSION}..."
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode commit \
  -o localhost:7050 \
  --ordererTLSHostnameOverride localhost \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN_ARGS
log "Chaincode committed"

# ─── Step 7: Verify ──────────────────────────────────────────────────────
info "Verifying deployment..."
peer lifecycle chaincode querycommitted \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --output json
log "Chaincode v${CHAINCODE_VERSION} deployed (sequence ${CHAINCODE_SEQUENCE})"

# ─── Step 8: Verify version ──────────────────────────────────────────────
info "Querying GetVersion..."
peer chaincode query \
  -C "$CHANNEL_NAME" \
  -n "$CHAINCODE_NAME" \
  -c '{"function":"admin:GetVersion","Args":[]}'

echo ""
echo "=============================================="
echo "  Chaincode v7.0 Deployed Successfully"
echo "=============================================="
echo ""
echo "New contracts registered: bridge, oracle"
echo "New functions available:"
echo "  bridge:ExportGO, bridge:ImportGO, bridge:ConfirmExport"
echo "  bridge:GetBridgeTransfer, bridge:ListBridgeTransfersPaginated"
echo "  oracle:PublishGridData, oracle:GetGridData, oracle:ListGridDataPaginated"
echo "  oracle:CrossReferenceGO"
echo "  device:VerifyDeviceReading, device:SubmitSignedReading"
echo ""
echo "Deprecated functions (ADR-022):"
echo "  query:GetCurrentEGOsList → use GetCurrentEGOsListPaginated"
echo "  query:GetCurrentHGOsList → use GetCurrentHGOsListPaginated"
echo "  query:GetCurrentBGOsList → use GetCurrentBGOsListPaginated"
echo "  device:ListDevices → use ListDevicesPaginated"
