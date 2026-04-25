#!/bin/bash
# deploy-chaincode.sh — v10.1: deploy golifecycle to electricity-de and hydrogen-de
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(dirname "$NETWORK_DIR")"
CC_NAME="golifecycle"
CC_VERSION="10.1"
CC_SEQUENCE=1
CC_PATH="$REPO_DIR/chaincode"
E_CHANNEL="electricity-de"
H_CHANNEL="hydrogen-de"
E_COLL="$REPO_DIR/collections/collection-config-electricity-de.json"
H_COLL="$REPO_DIR/collections/collection-config-hydrogen-de.json"

export PATH="${REPO_DIR}/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$NETWORK_DIR"
ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

E_ORGS=("eissuer:eissuerMSP:7051" "eproducer1:eproducer1MSP:9051" "ebuyer1:ebuyer1MSP:13051")
H_ORGS=("hissuer:hissuerMSP:8051" "hproducer1:hproducer1MSP:11051" "hbuyer1:hbuyer1MSP:14051")

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "====== Deploying $CC_NAME v$CC_VERSION to both channels ======"

# 1. Package
echo "[1/5] Packaging chaincode..."
set_peer_env eissuer eissuerMSP 7051
peer lifecycle chaincode package "${CC_NAME}.tar.gz" \
  --path "$CC_PATH" --lang golang --label "${CC_NAME}_${CC_VERSION}"
echo "  Packaged: ${CC_NAME}.tar.gz"

# 2. Install on ALL org peers (needed even if org only uses one channel, for endorsement)
echo "[2/5] Installing on all peers..."
for entry in "${E_ORGS[@]}" "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  peer lifecycle chaincode install "${CC_NAME}.tar.gz"
  echo "  Installed on $org"
done

# 3. Get package ID
set_peer_env eissuer eissuerMSP 7051
PKG_ID=$(peer lifecycle chaincode queryinstalled --output json \
  | jq -r ".installed_chaincodes[] | select(.label==\"${CC_NAME}_${CC_VERSION}\") | .package_id")
echo "Package ID: $PKG_ID"

approve_for_channel() {
  local channel=$1 coll=$2
  shift 2
  local orgs=("$@")
  echo "  Approving for channel $channel..."
  for entry in "${orgs[@]}"; do
    IFS=':' read -r org msp port <<< "$entry"
    set_peer_env "$org" "$msp" "$port"
    peer lifecycle chaincode approveformyorg \
      -o localhost:7050 --ordererTLSHostnameOverride localhost \
      --channelID "$channel" --name "$CC_NAME" --version "$CC_VERSION" \
      --package-id "$PKG_ID" --sequence "$CC_SEQUENCE" \
      --collections-config "$coll" \
      --tls --cafile "$ORDERER_CA"
    echo "    Approved: $org"
  done
}

commit_for_channel() {
  local channel=$1 coll=$2
  shift 2
  local orgs=("$@")
  # Build --peerAddresses args
  local peer_args=""
  for entry in "${orgs[@]}"; do
    IFS=':' read -r org msp port <<< "$entry"
    local tls_cert="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
    peer_args="$peer_args --peerAddresses localhost:${port} --tlsRootCertFiles ${tls_cert}"
  done
  set_peer_env $(echo "${orgs[0]}" | cut -d: -f1) $(echo "${orgs[0]}" | cut -d: -f2) $(echo "${orgs[0]}" | cut -d: -f3)
  # shellcheck disable=SC2086
  peer lifecycle chaincode commit \
    -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --channelID "$channel" --name "$CC_NAME" --version "$CC_VERSION" \
    --sequence "$CC_SEQUENCE" --collections-config "$coll" \
    --tls --cafile "$ORDERER_CA" $peer_args
  echo "  Committed on $channel"
}

# 4. Approve
echo "[4/5] Approving..."
approve_for_channel "$E_CHANNEL" "$E_COLL" "${E_ORGS[@]}"
approve_for_channel "$H_CHANNEL" "$H_COLL" "${H_ORGS[@]}"

# 5. Commit
echo "[5/5] Committing..."
commit_for_channel "$E_CHANNEL" "$E_COLL" "${E_ORGS[@]}"
commit_for_channel "$H_CHANNEL" "$H_COLL" "${H_ORGS[@]}"

echo ""
echo "====== Chaincode deployed to both channels ======"
echo "Next: bash init_ledger.sh"
