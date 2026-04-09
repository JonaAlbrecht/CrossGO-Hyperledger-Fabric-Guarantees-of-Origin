#!/bin/bash
# deploy-v8.sh — Deploy GO lifecycle chaincode v8.0 to both carrier channels (ADR-030/031)
#
# Deploys the same golifecycle chaincode to TWO channels with channel-specific
# collection configs:
#   electricity-de → collection-config-electricity.json (Issuer, EProducer, Buyer)
#   hydrogen-de    → collection-config-hydrogen.json    (Issuer, HProducer, Buyer)
#
# The chaincode binary is installed once on each peer. Approval and commit
# happen separately per channel with the appropriate endorsement subset.
#
# Usage: ./deploy-v8.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(dirname "$NETWORK_DIR")"

CHAINCODE_NAME="golifecycle"
CHAINCODE_VERSION="8.0"
CHAINCODE_SEQUENCE=1
CHAINCODE_PATH="$REPO_DIR/chaincode"

CHANNEL_ELEC="electricity-de"
CHANNEL_H2="hydrogen-de"

COLLECTIONS_ELEC="$REPO_DIR/collections/collection-config-electricity.json"
COLLECTIONS_H2="$REPO_DIR/collections/collection-config-hydrogen.json"

export PATH="${REPO_DIR}/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$NETWORK_DIR"

ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

# All orgs — for chaincode installation (every peer needs the binary)
ALL_ORGS=(
  "issuer1:issuer1MSP:7051"
  "eproducer1:eproducer1MSP:9051"
  "hproducer1:hproducer1MSP:11051"
  "buyer1:buyer1MSP:13051"
)

# Channel-specific org lists — for approval and commit
ORGS_ELEC=(
  "issuer1:issuer1MSP:7051"
  "eproducer1:eproducer1MSP:9051"
  "buyer1:buyer1MSP:13051"
)

ORGS_H2=(
  "issuer1:issuer1MSP:7051"
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
  local org_list=("$@")
  local args=""
  for entry in "${org_list[@]}"; do
    IFS=':' read -r org msp port <<< "$entry"
    args="$args --peerAddresses localhost:${port}"
    args="$args --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  done
  echo "$args"
}

deploy_to_channel() {
  local channel_name=$1
  local collections_config=$2
  shift 2
  local org_list=("$@")

  echo ""
  echo "====== Deploying to channel: $channel_name ======"

  # Approve for each org on this channel
  for entry in "${org_list[@]}"; do
    IFS=':' read -r org msp port <<< "$entry"
    info "Approving for ${org} on ${channel_name}..."
    set_peer_env "$org" "$msp" "$port"
    peer lifecycle chaincode approveformyorg \
      -o localhost:7050 \
      --ordererTLSHostnameOverride localhost \
      --channelID "$channel_name" \
      --name "$CHAINCODE_NAME" \
      --version "$CHAINCODE_VERSION" \
      --package-id "$PACKAGE_ID" \
      --sequence "$CHAINCODE_SEQUENCE" \
      --collections-config "$collections_config" \
      --tls --cafile "$ORDERER_CA"
    log "  ${org} approved on ${channel_name}"
  done

  # Check commit readiness
  info "Checking commit readiness on ${channel_name}..."
  set_peer_env issuer1 issuer1MSP 7051
  peer lifecycle chaincode checkcommitreadiness \
    --channelID "$channel_name" \
    --name "$CHAINCODE_NAME" \
    --version "$CHAINCODE_VERSION" \
    --sequence "$CHAINCODE_SEQUENCE" \
    --collections-config "$collections_config" \
    --output json

  # Build peer connection args for this channel's orgs
  PEER_CONN_ARGS=$(build_peer_conn_args "${org_list[@]}")

  # Commit
  info "Committing chaincode on ${channel_name}..."
  set_peer_env issuer1 issuer1MSP 7051
  peer lifecycle chaincode commit \
    -o localhost:7050 \
    --ordererTLSHostnameOverride localhost \
    --channelID "$channel_name" \
    --name "$CHAINCODE_NAME" \
    --version "$CHAINCODE_VERSION" \
    --sequence "$CHAINCODE_SEQUENCE" \
    --collections-config "$collections_config" \
    --tls --cafile "$ORDERER_CA" \
    $PEER_CONN_ARGS
  log "Chaincode committed on ${channel_name}"

  # Verify
  peer lifecycle chaincode querycommitted --channelID "$channel_name" --name "$CHAINCODE_NAME" --output json
}

init_channel_roles() {
  local channel_name=$1
  shift
  # First entries are role definitions (msp:role), last entries are org definitions (org:msp:port)
  # We use a separator "---" to split them
  local role_entries=()
  local channel_orgs=()
  local separator_found=false
  for arg in "$@"; do
    if [[ "$arg" == "---" ]]; then
      separator_found=true
      continue
    fi
    if $separator_found; then
      channel_orgs+=("$arg")
    else
      role_entries+=("$arg")
    fi
  done

  local PEER_CONN_ARGS=$(build_peer_conn_args "${channel_orgs[@]}")

  info "Initializing roles on ${channel_name}..."
  set_peer_env issuer1 issuer1MSP 7051
  peer chaincode invoke \
    -o localhost:7050 \
    --ordererTLSHostnameOverride localhost \
    -C "$channel_name" -n "$CHAINCODE_NAME" \
    --tls --cafile "$ORDERER_CA" \
    $PEER_CONN_ARGS \
    -c '{"function":"device:InitLedger","Args":["issuer1MSP"]}'
  sleep 2

  for role_entry in "${role_entries[@]}"; do
    IFS=':' read -r rmsp rrole <<< "$role_entry"
    peer chaincode invoke \
      -o localhost:7050 \
      --ordererTLSHostnameOverride localhost \
      -C "$channel_name" -n "$CHAINCODE_NAME" \
      --tls --cafile "$ORDERER_CA" \
      $PEER_CONN_ARGS \
      -c "{\"function\":\"device:RegisterOrgRole\",\"Args\":[\"${rmsp}\",\"${rrole}\"]}"
    sleep 1
  done
  log "Roles initialized on ${channel_name}"
}

echo "=============================================="
echo "  Deploying Chaincode: $CHAINCODE_NAME v$CHAINCODE_VERSION"
echo "  Multi-Channel (ADR-030/031)"
echo "=============================================="

# ─── Step 1: Package chaincode (once) ────────────────────────────────────────
info "Packaging chaincode..."
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode package "${CHAINCODE_NAME}.tar.gz" \
  --path "$CHAINCODE_PATH" \
  --lang golang \
  --label "${CHAINCODE_NAME}_${CHAINCODE_VERSION}"
log "Packaged: ${CHAINCODE_NAME}.tar.gz"

# ─── Step 2: Install on ALL peers (binary needed on every peer) ──────────────
for entry in "${ALL_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  info "Installing on ${org}..."
  set_peer_env "$org" "$msp" "$port"
  peer lifecycle chaincode install "${CHAINCODE_NAME}.tar.gz"
  log "  Installed on ${org}"
done

# ─── Step 3: Get package ID ──────────────────────────────────────────────────
set_peer_env issuer1 issuer1MSP 7051
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r ".installed_chaincodes[] | select(.label==\"${CHAINCODE_NAME}_${CHAINCODE_VERSION}\") | .package_id")
echo "Package ID: $PACKAGE_ID"

# ─── Step 4: Deploy to electricity-de channel ────────────────────────────────
deploy_to_channel "$CHANNEL_ELEC" "$COLLECTIONS_ELEC" "${ORGS_ELEC[@]}"

# ─── Step 5: Deploy to hydrogen-de channel ───────────────────────────────────
deploy_to_channel "$CHANNEL_H2" "$COLLECTIONS_H2" "${ORGS_H2[@]}"

# ─── Step 6: Initialize roles per channel ─────────────────────────────────────
# electricity-de roles
ELEC_ROLES=("eproducer1MSP:producer" "buyer1MSP:consumer")
init_channel_roles "$CHANNEL_ELEC" "${ELEC_ROLES[@]}" "---" "${ORGS_ELEC[@]}"

# hydrogen-de roles
H2_ROLES=("hproducer1MSP:producer" "buyer1MSP:consumer")
init_channel_roles "$CHANNEL_H2" "${H2_ROLES[@]}" "---" "${ORGS_H2[@]}"

echo ""
echo "=============================================="
echo "  Chaincode v8.0 Deployed — Multi-Channel"
echo "=============================================="
echo ""
echo "Channel: $CHANNEL_ELEC"
echo "  Collection config: collection-config-electricity.json"
echo "  Orgs: issuer1MSP (issuer), eproducer1MSP (producer), buyer1MSP (consumer)"
echo ""
echo "Channel: $CHANNEL_H2"
echo "  Collection config: collection-config-hydrogen.json"
echo "  Orgs: issuer1MSP (issuer), hproducer1MSP (producer), buyer1MSP (consumer)"
echo ""
echo "Bridge protocol (ADR-031):"
echo "  Phase 1: bridge:LockGO       (source channel)"
echo "  Phase 2: bridge:MintFromBridge (destination channel)"
echo "  Phase 3: bridge:FinalizeLock  (source channel)"
echo ""
echo "Test (electricity):"
echo "  peer chaincode query -C $CHANNEL_ELEC -n $CHAINCODE_NAME -c '{\"function\":\"admin:GetVersion\",\"Args\":[]}'"
echo "Test (hydrogen):"
echo "  peer chaincode query -C $CHANNEL_H2 -n $CHAINCODE_NAME -c '{\"function\":\"admin:GetVersion\",\"Args\":[]}'"
