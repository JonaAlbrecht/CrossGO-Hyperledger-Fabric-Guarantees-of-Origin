#!/bin/bash
# Manual Phase 1 test with verbose output
set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network

NETWORK_DIR=/root/hlf-go/repo/network
CC_NAME=golifecycle
E_CHANNEL=electricity-de
H_CHANNEL=hydrogen-de
ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"
EISSUER_TLS="$NETWORK_DIR/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt"
EPRODUCER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "Testing Phase 1: LockGOForConversion"
set_peer_env eproducer1 eproducer1MSP 9051

ACTIVE_GOID="eGO_186277a8af699282"
echo "Using GO: $ACTIVE_GOID"

echo ""
echo "Attempting to lock GO for conversion..."

# Build transient data
LOCK_INPUT=$(python3 -c "
import json, base64
lock_data = {
    'GOAssetID': '$ACTIVE_GOID',
    'DestinationChannel': '$H_CHANNEL',
    'DestinationCarrier': 'hydrogen',
    'ConversionMethod': 'electrolysis',
    'ConversionEfficiency': 0.65,
    'OwnerMSP': 'eproducer1MSP',
    'DestinationOwnerMSP': 'hproducer1MSP'
}
print(base64.b64encode(json.dumps(lock_data).encode()).decode())
")

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$E_CHANNEL" -n "$CC_NAME" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls --cafile "$ORDERER_CA" \
  -c '{"function":"conversion:LockGOForConversion","Args":[]}' \
  --transient "{\"LockForConversion\":\"$LOCK_INPUT\"}"
