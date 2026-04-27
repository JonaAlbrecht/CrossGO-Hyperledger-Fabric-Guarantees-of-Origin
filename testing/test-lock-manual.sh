#!/bin/bash
# Test lock operation manually

REPO_DIR="/root/hlf-go/repo"
export PATH="$REPO_DIR/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$REPO_DIR/network"
ORDERER_CA="$REPO_DIR/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

# eproducer1 environment
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="eproducer1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
export CORE_PEER_MSPCONFIGPATH="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/users/Admin@eproducer1.go-platform.com/msp"
export CORE_PEER_ADDRESS="localhost:9051"

# TLS certs
EISSUER_TLS="$REPO_DIR/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt"
EPRODUCER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
EBUYER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt"

GO_ID="eGO_39f2865b4e629367"

LOCK_JSON="{\"GOAssetID\":\"$GO_ID\",\"DestinationChannel\":\"hydrogen-de\",\"DestinationCarrier\":\"H2\",\"ConversionMethod\":\"Electrolysis\",\"ConversionEfficiency\":0.75,\"OwnerMSP\":\"eproducer1MSP\"}"
LOCK_B64=$(echo -n "$LOCK_JSON" | base64 | tr -d '\n')

echo "Testing LockGOForConversion with GO: $GO_ID"
echo "Lock JSON: $LOCK_JSON"
echo "Lock B64: $LOCK_B64"
echo ""

peer chaincode invoke \
  -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
  -C electricity-de -n golifecycle \
  -c '{"function":"conversion:LockGOForConversion","Args":[]}' \
  --transient "{\"LockForConversion\":\"$LOCK_B64\"}" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
  --tls --cafile "$ORDERER_CA"
