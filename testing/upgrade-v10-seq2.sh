#!/bin/bash
# Upgrade chaincode to sequence 2 with fixed collection configs
# This upgrades the chaincode definition with requiredPeerCount: 1 to fix private data dissemination

set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem
ORDERER_ENDPOINT=localhost:7050

E_COLL=/root/hlf-go/repo/collections/collection-config-electricity-de.json
H_COLL=/root/hlf-go/repo/collections/collection-config-hydrogen-de.json

PACKAGE_ID="golifecycle_10.1:6222da555adb1aed28961e8bc66a80daae8857686dec6d1e8d7e903df94852a9"

set_peer_env() {
  local org=$1
  local port=$2
  export CORE_PEER_LOCALMSPID="${org}MSP"
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp
  export CORE_PEER_ADDRESS=localhost:${port}
}

echo "========================================="
echo "UPGRADING CHAINCODE TO SEQUENCE 2"
echo "Fix: requiredPeerCount: 1 for private collections"
echo "========================================="
echo ""

# ELECTRICITY-DE CHANNEL - SEQUENCE 2
echo ">>> Approving chaincode for eissuer on electricity-de (sequence 2)..."
set_peer_env eissuer 7051
peer lifecycle chaincode approveformyorg -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID electricity-de --name golifecycle --version 10.1 \
  --package-id $PACKAGE_ID --sequence 2 --collections-config $E_COLL

echo ">>> Approving chaincode for eproducer1 on electricity-de (sequence 2)..."
set_peer_env eproducer1 9051
peer lifecycle chaincode approveformyorg -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID electricity-de --name golifecycle --version 10.1 \
  --package-id $PACKAGE_ID --sequence 2 --collections-config $E_COLL

echo ">>> Approving chaincode for ebuyer1 on electricity-de (sequence 2)..."
set_peer_env ebuyer1 13051
peer lifecycle chaincode approveformyorg -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID electricity-de --name golifecycle --version 10.1 \
  --package-id $PACKAGE_ID --sequence 2 --collections-config $E_COLL

echo ">>> Committing chaincode to electricity-de (sequence 2)..."
set_peer_env eissuer 7051
peer lifecycle chaincode commit -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID electricity-de --name golifecycle --version 10.1 \
  --sequence 2 --collections-config $E_COLL \
  --peerAddresses localhost:7051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  --peerAddresses localhost:9051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt \
  --peerAddresses localhost:13051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt

echo ">>> Verifying commit on electricity-de..."
peer lifecycle chaincode querycommitted --channelID electricity-de --name golifecycle

echo ""
echo "========================================="
echo "HYDROGEN-DE CHANNEL - SEQUENCE 2"
echo "========================================="
echo ""

# HYDROGEN-DE CHANNEL - SEQUENCE 2
echo ">>> Approving chaincode for hissuer on hydrogen-de (sequence 2)..."
set_peer_env hissuer 8051
peer lifecycle chaincode approveformyorg -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID hydrogen-de --name golifecycle --version 10.1 \
  --package-id $PACKAGE_ID --sequence 2 --collections-config $H_COLL

echo ">>> Approving chaincode for hproducer1 on hydrogen-de (sequence 2)..."
set_peer_env hproducer1 11051
peer lifecycle chaincode approveformyorg -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID hydrogen-de --name golifecycle --version 10.1 \
  --package-id $PACKAGE_ID --sequence 2 --collections-config $H_COLL

echo ">>> Approving chaincode for hbuyer1 on hydrogen-de (sequence 2)..."
set_peer_env hbuyer1 14051
peer lifecycle chaincode approveformyorg -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID hydrogen-de --name golifecycle --version 10.1 \
  --package-id $PACKAGE_ID --sequence 2 --collections-config $H_COLL

echo ">>> Committing chaincode to hydrogen-de (sequence 2)..."
set_peer_env hissuer 8051
peer lifecycle chaincode commit -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA --channelID hydrogen-de --name golifecycle --version 10.1 \
  --sequence 2 --collections-config $H_COLL \
  --peerAddresses localhost:8051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt \
  --peerAddresses localhost:11051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt \
  --peerAddresses localhost:14051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt

echo ">>> Verifying commit on hydrogen-de..."
peer lifecycle chaincode querycommitted --channelID hydrogen-de --name golifecycle

echo ""
echo "========================================="
echo "CHAINCODE UPGRADE TO SEQUENCE 2 COMPLETE"
echo "Collection configs updated with requiredPeerCount: 1"
echo "========================================="
