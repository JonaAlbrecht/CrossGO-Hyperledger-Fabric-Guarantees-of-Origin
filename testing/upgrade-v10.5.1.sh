#!/bin/bash
# Rebuild and upgrade to v10.5.1 (fixed pointer types)
set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network

REPO_DIR=/root/hlf-go/repo
NETWORK_DIR=$REPO_DIR/network
CC_NAME=golifecycle
CC_VERSION=10.5.1
CC_SRC="$REPO_DIR/chaincode"
E_CHANNEL=electricity-de
H_CHANNEL=hydrogen-de
E_COLL="$REPO_DIR/collections/collection-config-electricity-de.json"
H_COLL="$REPO_DIR/collections/collection-config-hydrogen-de.json"
ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "Building chaincode v10.5.1 (fixed pointer types)..."
cd "$REPO_DIR"
peer lifecycle chaincode package "${CC_NAME}_${CC_VERSION}.tar.gz" \
  --path "$CC_SRC" \
  --lang golang \
  --label "${CC_NAME}_${CC_VERSION}"
echo "Package created: ${CC_NAME}_${CC_VERSION}.tar.gz"

echo ""
echo "Installing on all 6 peers..."
for org_entry in "eissuer:eissuerMSP:7051" "eproducer1:eproducer1MSP:9051" "ebuyer1:ebuyer1MSP:13051" "hissuer:hissuerMSP:8051" "hproducer1:hproducer1MSP:11051" "hbuyer1:hbuyer1MSP:14051"; do
  IFS=':' read -r org msp port <<< "$org_entry"
  set_peer_env "$org" "$msp" "$port"
  echo "  Installing on peer0.$org..."
  peer lifecycle chaincode install "${CC_NAME}_${CC_VERSION}.tar.gz" 2>&1 | tail -n 2 || echo "    (already installed or error)"
done

# Get package ID
set_peer_env eissuer eissuerMSP 7051
PKG_ID=$(peer lifecycle chaincode queryinstalled --output json | python3 -c "import sys,json; d=json.load(sys.stdin); print([x['package_id'] for x in d.get('installed_chaincodes',[]) if x['label']=='${CC_NAME}_${CC_VERSION}'][0])" 2>/dev/null || echo "")
echo "Package ID: $PKG_ID"

if [ -z "$PKG_ID" ]; then
  echo "ERROR: Failed to get package ID"
  exit 1
fi

echo ""
echo "Approving for sequence 6 on both channels..."
E_ORGS=("eissuer:eissuerMSP:7051" "eproducer1:eproducer1MSP:9051" "ebuyer1:ebuyer1MSP:13051")
H_ORGS=("hissuer:hissuerMSP:8051" "hproducer1:hproducer1MSP:11051" "hbuyer1:hbuyer1MSP:14051")

for entry in "${E_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  echo "  Approving $msp for $E_CHANNEL seq 6..."
  peer lifecycle chaincode approveformyorg \
    -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --channelID "$E_CHANNEL" --name "$CC_NAME" \
    --version "$CC_VERSION" --package-id "$PKG_ID" \
    --sequence 6 \
    --collections-config "$E_COLL" \
    --tls --cafile "$ORDERER_CA" 2>&1 | tail -n 1
done

for entry in "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  echo "  Approving $msp for $H_CHANNEL seq 6..."
  peer lifecycle chaincode approveformyorg \
    -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --channelID "$H_CHANNEL" --name "$CC_NAME" \
    --version "$CC_VERSION" --package-id "$PKG_ID" \
    --sequence 6 \
    --collections-config "$H_COLL" \
    --tls --cafile "$ORDERER_CA" 2>&1 | tail -n 1
done

echo ""
echo "Committing on both channels..."
set_peer_env eissuer eissuerMSP 7051
peer lifecycle chaincode commit \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$E_CHANNEL" --name "$CC_NAME" --version "$CC_VERSION" \
  --sequence 6 \
  --collections-config "$E_COLL" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$NETWORK_DIR/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$NETWORK_DIR/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$NETWORK_DIR/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt" \
  --tls --cafile "$ORDERER_CA" 2>&1 | tail -n 3

set_peer_env hissuer hissuerMSP 8051
peer lifecycle chaincode commit \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$H_CHANNEL" --name "$CC_NAME" --version "$CC_VERSION" \
  --sequence 6 \
  --collections-config "$H_COLL" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$NETWORK_DIR/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$NETWORK_DIR/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$NETWORK_DIR/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt" \
  --tls --cafile "$ORDERER_CA" 2>&1 | tail -n 3

echo ""
echo "Upgrade complete! Chaincode v10.5.1 is now active."
