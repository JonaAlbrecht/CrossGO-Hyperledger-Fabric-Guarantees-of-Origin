#!/bin/bash
# approve-commit-chaincode.sh — Approve and commit already-installed chaincode
set -euo pipefail

REPO_DIR=/root/hlf-go/repo
NETWORK_DIR=$REPO_DIR/network
CHANNEL_NAME=goplatformchannel
CHAINCODE_NAME=golifecycle
CHAINCODE_VERSION=2.0
CHAINCODE_SEQUENCE=1
COLLECTIONS_CONFIG=$REPO_DIR/collections/collection-config.json

export PATH=$REPO_DIR/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$NETWORK_DIR

ORDERER_CA=$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

ORGS=(
  "issuer1:issuer1MSP:7051"
  "eproducer1:eproducer1MSP:9051"
  "hproducer1:hproducer1MSP:11051"
  "buyer1:buyer1MSP:13051"
)

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

# Get package ID
set_peer_env issuer1 issuer1MSP 7051
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r ".installed_chaincodes[] | select(.label==\"${CHAINCODE_NAME}_${CHAINCODE_VERSION}\") | .package_id")
echo "Package ID: $PACKAGE_ID"

# Approve for each org
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  echo "--- Approving for ${org} ---"
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
done

# Check commit readiness
echo "--- Checking Commit Readiness ---"
set_peer_env issuer1 issuer1MSP 7051
peer lifecycle chaincode checkcommitreadiness \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --output json

# Build peer connection args
PEER_CONN_ARGS=""
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  PEER_CONN_ARGS="$PEER_CONN_ARGS --peerAddresses localhost:${port}"
  PEER_CONN_ARGS="$PEER_CONN_ARGS --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
done

# Commit
echo "--- Committing Chaincode ---"
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

# Verify
echo "--- Verifying ---"
peer lifecycle chaincode querycommitted --channelID "$CHANNEL_NAME" --name "$CHAINCODE_NAME" --output json

# Initialize ledger
echo "--- Initializing Ledger ---"
peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride localhost \
  -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN_ARGS \
  -c '{"function":"DeviceContract:InitLedger","Args":["issuer1MSP"]}'
sleep 2

for role_entry in "eproducer1MSP:producer" "hproducer1MSP:producer" "buyer1MSP:consumer"; do
  IFS=':' read -r rmsp rrole <<< "$role_entry"
  peer chaincode invoke \
    -o localhost:7050 \
    --ordererTLSHostnameOverride localhost \
    -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" \
    --tls --cafile "$ORDERER_CA" \
    $PEER_CONN_ARGS \
    -c "{\"function\":\"DeviceContract:RegisterOrgRole\",\"Args\":[\"${rmsp}\",\"${rrole}\"]}"
  sleep 1
done

echo ""
echo "====== Chaincode Deployed & Initialized ======"
