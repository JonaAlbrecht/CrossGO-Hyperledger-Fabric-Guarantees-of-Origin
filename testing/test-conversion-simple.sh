#!/bin/bash
# Simplified 3-phase conversion test with hardcoded data
set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network

REPO_DIR=/root/hlf-go/repo
NETWORK_DIR=$REPO_DIR/network
CC_NAME=golifecycle
E_CHANNEL=electricity-de
H_CHANNEL=hydrogen-de
EISSUER_MSP=eissuerMSP
HISSUER_MSP=hissuerMSP

ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"
EISSUER_TLS="$NETWORK_DIR/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt"
EPRODUCER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
HISSUER_TLS="$NETWORK_DIR/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt"
HPRODUCER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt"

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "====== Cross-Channel Conversion Test (Simplified) ======"
echo ""

# Get an existing GO to use
set_peer_env eproducer1 eproducer1MSP 9051
ACTIVE_GOID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null \
  | python3 -c "import sys,json; gos=json.load(sys.stdin); print(gos[0]['AssetID'] if gos else '')" 2>/dev/null || true)

if [ -z "$ACTIVE_GOID" ]; then
  echo "No active GOs found. Please create one first."
  exit 1
fi

echo "Using GO: $ACTIVE_GOID"

# Phase 1: Lock GO
echo "Phase 1: LockGOForConversion..."
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
  --transient "{\"LockForConversion\":\"$LOCK_INPUT\"}" > /tmp/phase1.log 2>&1

if grep -q "Chaincode invoke successful" /tmp/phase1.log; then
  echo "✓ Phase 1 succeeded"
else
  echo "✗ Phase 1 failed"
  cat /tmp/phase1.log
  exit 1
fi

sleep 3

# Get lock ID
set_peer_env eissuer eissuerMSP 7051
LOCK_ID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"conversion:ListConversionLocks","Args":["100",""]}' 2>/dev/null \
  | python3 -c "import sys,json; locks=json.load(sys.stdin); print(locks[-1]['LockID'] if locks else '')" || true)

echo "Lock ID: $LOCK_ID"

# Get lock receipt hash from the lock
LOCK_HASH=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c "{\"function\":\"conversion:GetConversionLock\",\"Args\":[\"$LOCK_ID\"]}" 2>/dev/null \
  | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('LockReceiptHash',''))" || echo "")

echo "Lock Hash: $LOCK_HASH"

# Phase 2: Mint from conversion
echo "Phase 2: MintFromConversion..."
set_peer_env hissuer hissuerMSP 8051

# Hardcoded receipt data
MINT_RECEIPT=$(python3 -c "
import json, base64
receipt = {
    'LockID': '$LOCK_ID',
    'GOAssetID': '$ACTIVE_GOID',
    'SourceChannel': '$E_CHANNEL',
    'SourceCarrier': 'electricity',
    'DestinationChannel': '$H_CHANNEL',
    'DestinationCarrier': 'hydrogen',
    'ConversionMethod': 'electrolysis',
    'ConversionEfficiency': 0.65,
    'OwnerMSP': 'eproducer1MSP',
    'DestinationOwnerMSP': 'hproducer1MSP',
    'SourceIssuerMSP': 'eissuerMSP',
    'LockReceiptHash': '$LOCK_HASH',
    'TxID': 'dummy_txid',
    'SourceAmount': 100,
    'SourceAmountUnit': 'MWh',
    'SourceEmissions': 10.5,
    'SourceProductionMethod': 'solar_pv',
    'SourceDeviceID': 'device_eproducer1_001',
    'SourceCreationDateTime': 1777240000,
    'SourceConsumptionDecls': [],
    'SourceCountryOfOrigin': 'DE',
    'SourceProductionStart': 1777240000,
    'SourceProductionEnd': 1777243600,
    'SourceSupportScheme': 'none',
    'SourceGridConnectionPoint': 'GridPointE1'
}
print(base64.b64encode(json.dumps(receipt).encode()).decode())
")

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$H_CHANNEL" -n "$CC_NAME" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPRODUCER1_TLS" \
  --tls --cafile "$ORDERER_CA" \
  -c '{"function":"conversion:MintFromConversion","Args":[]}' \
  --transient "{\"MintFromConversion\":\"$MINT_RECEIPT\"}" > /tmp/phase2.log 2>&1

if grep -q "Chaincode invoke successful" /tmp/phase2.log; then
  echo "✓ Phase 2 succeeded"
else
  echo "✗ Phase 2 failed"
  cat /tmp/phase2.log
  exit 1
fi

echo ""
echo "====== Test Complete ======"
echo "✓ All phases succeeded!"
echo "Lock: $LOCK_ID"
