#!/bin/bash
# deploy-chaincode.sh — Package, install, approve, and commit the GO lifecycle chaincode
# Usage: ./deploy-chaincode.sh
set -euo pipefail

CHANNEL_NAME="goplatformchannel"
CHAINCODE_NAME="golifecycle"
CHAINCODE_VERSION="2.0"
CHAINCODE_SEQUENCE=1
CHAINCODE_PATH="../../chaincode"
COLLECTIONS_CONFIG="../../collections/collection-config.json"

echo "====== Deploying Chaincode: $CHAINCODE_NAME v$CHAINCODE_VERSION ======"

# 1. Package chaincode
echo "--- Packaging Chaincode ---"
peer lifecycle chaincode package "${CHAINCODE_NAME}.tar.gz" \
  --path "$CHAINCODE_PATH" \
  --lang golang \
  --label "${CHAINCODE_NAME}_${CHAINCODE_VERSION}"

# 2. Install on each org's peer
echo "--- Installing on Issuer Peer ---"
# Set environment for issuer1
export CORE_PEER_ADDRESS=peer0.issuer1.go-platform.com:7051
export CORE_PEER_LOCALMSPID=issuer1MSP
# TODO: Set CORE_PEER_TLS_ROOTCERT_FILE, CORE_PEER_MSPCONFIGPATH
peer lifecycle chaincode install "${CHAINCODE_NAME}.tar.gz"

echo "--- Installing on Producer Peer ---"
export CORE_PEER_ADDRESS=peer0.producer1.go-platform.com:9051
export CORE_PEER_LOCALMSPID=producer1MSP
peer lifecycle chaincode install "${CHAINCODE_NAME}.tar.gz"

echo "--- Installing on Consumer Peer ---"
export CORE_PEER_ADDRESS=peer0.consumer1.go-platform.com:11051
export CORE_PEER_LOCALMSPID=consumer1MSP
peer lifecycle chaincode install "${CHAINCODE_NAME}.tar.gz"

# 3. Get package ID
echo "--- Getting Package ID ---"
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r ".installed_chaincodes[] | select(.label==\"${CHAINCODE_NAME}_${CHAINCODE_VERSION}\") | .package_id")
echo "Package ID: $PACKAGE_ID"

# 4. Approve for each org
echo "--- Approving for Issuer ---"
export CORE_PEER_ADDRESS=peer0.issuer1.go-platform.com:7051
export CORE_PEER_LOCALMSPID=issuer1MSP
peer lifecycle chaincode approveformyorg \
  -o orderer1.go-platform.com:7050 \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --package-id "$PACKAGE_ID" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --tls --cafile /path/to/orderer/ca.crt

echo "--- Approving for Producer ---"
export CORE_PEER_ADDRESS=peer0.producer1.go-platform.com:9051
export CORE_PEER_LOCALMSPID=producer1MSP
peer lifecycle chaincode approveformyorg \
  -o orderer1.go-platform.com:7050 \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --package-id "$PACKAGE_ID" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --tls --cafile /path/to/orderer/ca.crt

echo "--- Approving for Consumer ---"
export CORE_PEER_ADDRESS=peer0.consumer1.go-platform.com:11051
export CORE_PEER_LOCALMSPID=consumer1MSP
peer lifecycle chaincode approveformyorg \
  -o orderer1.go-platform.com:7050 \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --package-id "$PACKAGE_ID" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --tls --cafile /path/to/orderer/ca.crt

# 5. Check commit readiness
echo "--- Checking Commit Readiness ---"
peer lifecycle chaincode checkcommitreadiness \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --output json

# 6. Commit
echo "--- Committing Chaincode ---"
peer lifecycle chaincode commit \
  -o orderer1.go-platform.com:7050 \
  --channelID "$CHANNEL_NAME" \
  --name "$CHAINCODE_NAME" \
  --version "$CHAINCODE_VERSION" \
  --sequence "$CHAINCODE_SEQUENCE" \
  --collections-config "$COLLECTIONS_CONFIG" \
  --tls --cafile /path/to/orderer/ca.crt \
  --peerAddresses peer0.issuer1.go-platform.com:7051 \
  --peerAddresses peer0.producer1.go-platform.com:9051 \
  --peerAddresses peer0.consumer1.go-platform.com:11051

echo "====== Chaincode Deployed Successfully ======"
echo "Invoke with: peer chaincode invoke ... -C $CHANNEL_NAME -n $CHAINCODE_NAME -c '{\"function\":\"issuance:CreateElectricityGO\",\"Args\":[]}'"
