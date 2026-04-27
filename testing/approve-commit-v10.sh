#!/bin/bash
# Approve and commit golifecycle v10.1 to both channels (after BOM fix)

REPO_DIR="/root/hlf-go/repo"
PKG_ID="golifecycle_10.1:6222da555adb1aed28961e8bc66a80daae8857686dec6d1e8d7e903df94852a9"
E_COLL="$REPO_DIR/collections/collection-config-electricity-de.json"
H_COLL="$REPO_DIR/collections/collection-config-hydrogen-de.json"

export PATH="$REPO_DIR/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$REPO_DIR/network"
ORDERER_CA="$REPO_DIR/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$REPO_DIR/network/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$REPO_DIR/network/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "=== Approving for electricity-de channel ==="
set_peer_env eissuer eissuerMSP 7051
peer lifecycle chaincode approveformyorg \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID electricity-de --name golifecycle --version 10.1 \
  --package-id "$PKG_ID" --sequence 1 \
  --collections-config "$E_COLL" \
  --tls --cafile "$ORDERER_CA"
echo "✓ eissuer"

set_peer_env eproducer1 eproducer1MSP 9051
peer lifecycle chaincode approveformyorg \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID electricity-de --name golifecycle --version 10.1 \
  --package-id "$PKG_ID" --sequence 1 \
  --collections-config "$E_COLL" \
  --tls --cafile "$ORDERER_CA"
echo "✓ eproducer1"

set_peer_env ebuyer1 ebuyer1MSP 13051
peer lifecycle chaincode approveformyorg \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID electricity-de --name golifecycle --version 10.1 \
  --package-id "$PKG_ID" --sequence 1 \
  --collections-config "$E_COLL" \
  --tls --cafile "$ORDERER_CA"
echo "✓ ebuyer1"

echo ""
echo "=== Committing to electricity-de ==="
set_peer_env eissuer eissuerMSP 7051
peer lifecycle chaincode commit \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID electricity-de --name golifecycle --version 10.1 \
  --sequence 1 --collections-config "$E_COLL" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$REPO_DIR/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$REPO_DIR/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt" \
  --tls --cafile "$ORDERER_CA"
echo "✓ electricity-de committed"

echo ""
echo "=== Approving for hydrogen-de channel ==="
set_peer_env hissuer hissuerMSP 8051
peer lifecycle chaincode approveformyorg \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID hydrogen-de --name golifecycle --version 10.1 \
  --package-id "$PKG_ID" --sequence 1 \
  --collections-config "$H_COLL" \
  --tls --cafile "$ORDERER_CA"
echo "✓ hissuer"

set_peer_env hproducer1 hproducer1MSP 11051
peer lifecycle chaincode approveformyorg \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID hydrogen-de --name golifecycle --version 10.1 \
  --package-id "$PKG_ID" --sequence 1 \
  --collections-config "$H_COLL" \
  --tls --cafile "$ORDERER_CA"
echo "✓ hproducer1"

set_peer_env hbuyer1 hbuyer1MSP 14051
peer lifecycle chaincode approveformyorg \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID hydrogen-de --name golifecycle --version 10.1 \
  --package-id "$PKG_ID" --sequence 1 \
  --collections-config "$H_COLL" \
  --tls --cafile "$ORDERER_CA"
echo "✓ hbuyer1"

echo ""
echo "=== Committing to hydrogen-de ==="
set_peer_env hissuer hissuerMSP 8051
peer lifecycle chaincode commit \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  --channelID hydrogen-de --name golifecycle --version 10.1 \
  --sequence 1 --collections-config "$H_COLL" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$REPO_DIR/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$REPO_DIR/network/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$REPO_DIR/network/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt" \
  --tls --cafile "$ORDERER_CA"
echo "✓ hydrogen-de committed"

echo ""
echo "=== Verifying Deployments ==="
echo "Checking electricity-de:"
set_peer_env eissuer eissuerMSP 7051
peer lifecycle chaincode querycommitted --channelID electricity-de --name golifecycle

echo ""
echo "Checking hydrogen-de:"
set_peer_env hissuer hissuerMSP 8051
peer lifecycle chaincode querycommitted --channelID hydrogen-de --name golifecycle

echo ""
echo "====== Deployment Complete ======"
