#!/bin/bash
# Cross-channel conversion test (Phase 0-3)
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

echo "====== End-to-end Cross-Channel Conversion Test ======"
echo ""

# ---- Phase 0: Get or create an electricity GO ----
echo "--- Phase 0: Finding active electricity GOs ---"
set_peer_env eproducer1 eproducer1MSP 9051
ACTIVE_GOID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null \
  | python3 -c "import sys,json; gos=json.load(sys.stdin); print(gos[0]['AssetID'] if gos else '')" 2>/dev/null || true)

if [ -z "$ACTIVE_GOID" ]; then
  echo "No active electricity GOs found. Creating a new one first..."
  set_peer_env eissuer eissuerMSP 7051
  CREATE_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -C "$E_CHANNEL" -n "$CC_NAME" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --tls --cafile "$ORDERER_CA" \
    -c '{"function":"electricity:CreateElectricityGO","Args":["eproducer1MSP","device_eproducer1_001","100","MWh","2026-01-01T00:00:00Z","2026-01-01T01:00:00Z","10.5","solar","DE","GridPointE1","SchemeE1",""]}' 2>&1)
  echo "$CREATE_OUT"
  sleep 3
  
  set_peer_env eproducer1 eproducer1MSP 9051
  ACTIVE_GOID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
    -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null \
    | python3 -c "import sys,json; gos=json.load(sys.stdin); print(gos[0]['AssetID'] if gos else '')" 2>/dev/null || true)
  
  if [ -z "$ACTIVE_GOID" ]; then
    echo "[FAIL] Could not create electricity GO"
    exit 1
  fi
fi

echo "Using electricity GO: $ACTIVE_GOID"

# ---- Phase 1: LockGOForConversion (electricity-de) ----
echo ""
echo "--- Phase 1: LockGOForConversion (electricity-de) ---"
set_peer_env eproducer1 eproducer1MSP 9051

# Build transient data for LockGOForConversion
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
" 2>/dev/null)

PHASE1_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$E_CHANNEL" -n "$CC_NAME" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls --cafile "$ORDERER_CA" \
  -c '{"function":"conversion:LockGOForConversion","Args":[]}' \
  --transient "{\"LockForConversion\":\"$LOCK_INPUT\"}" 2>&1)

echo "$PHASE1_OUT"
sleep 4

# Query for locks and get the newest one
set_peer_env eissuer eissuerMSP 7051
LOCK_ID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"conversion:ListConversionLocks","Args":["100",""]}' 2>/dev/null \
  | python3 -c "import sys,json; locks=json.load(sys.stdin); print(locks[-1]['LockID'] if locks else '')" 2>/dev/null || true)

if [ -z "$LOCK_ID" ]; then
  echo "[FAIL] Phase 1 failed - no lock ID found"
  exit 1
fi

echo "Lock ID: $LOCK_ID"
if echo "$PHASE1_OUT" | python3 -c "import sys; exit(0 if 'Chaincode invoke successful' in sys.stdin.read() else 1)"; then
  echo "[PASS] Phase 1 LockGOForConversion succeeded"
else
  echo "[FAIL] Phase 1 LockGOForConversion failed"
  exit 1
fi

# Get lock details
set_peer_env eissuer eissuerMSP 7051
LOCK_JSON=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c "{\"function\":\"conversion:GetConversionLock\",\"Args\":[\"$LOCK_ID\"]}" 2>/dev/null || echo "{}")

# ---- Phase 2: MintFromConversion (hydrogen-de) ----
echo ""
echo "--- Phase 2: MintFromConversion (hydrogen-de) ---"
set_peer_env hproducer1 hproducer1MSP 11051

# Build transient data for MintFromConversion
SOURCE_AMOUNT=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceAmount',100))" 2>/dev/null || echo "100")
SOURCE_EMISSIONS=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceEmissions',10.5))" 2>/dev/null || echo "10.5")
SOURCE_PROD_METHOD=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceProductionMethod','solar'))" 2>/dev/null || echo "solar")
SOURCE_DEVICE=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceDeviceID','device_eproducer1_001'))" 2>/dev/null || echo "device_eproducer1_001")
SOURCE_CREATION=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceCreationDateTime',0))" 2>/dev/null || echo "0")
SOURCE_CONSUMP=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(json.dumps(d.get('SourceConsumptionDecls',[])))" 2>/dev/null || echo "[]")
SOURCE_COUNTRY=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceCountryOfOrigin','DE'))" 2>/dev/null || echo "DE")
SOURCE_PROD_START=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceProductionStart',0))" 2>/dev/null || echo "0")
SOURCE_PROD_END=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceProductionEnd',0))" 2>/dev/null || echo "0")
SOURCE_SUPPORT=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceSupportScheme',''))" 2>/dev/null || echo "")
SOURCE_GRID=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceGridConnectionPoint',''))" 2>/dev/null || echo "")
LOCK_HASH=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('LockReceiptHash',''))" 2>/dev/null || echo "")
TXID=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('TxID',''))" 2>/dev/null || echo "")

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
    'SourceIssuerMSP': '$EISSUER_MSP',
    'LockReceiptHash': '$LOCK_HASH',
    'TxID': '$TXID',
    'SourceAmount': $SOURCE_AMOUNT,
    'SourceAmountUnit': 'MWh',
    'SourceEmissions': $SOURCE_EMISSIONS,
    'SourceProductionMethod': '$SOURCE_PROD_METHOD',
    'SourceDeviceID': '$SOURCE_DEVICE',
    'SourceCreationDateTime': $SOURCE_CREATION,
    'SourceConsumptionDecls': $SOURCE_CONSUMP,
    'SourceCountryOfOrigin': '$SOURCE_COUNTRY',
    'SourceProductionStart': $SOURCE_PROD_START,
    'SourceProductionEnd': $SOURCE_PROD_END,
    'SourceSupportScheme': '$SOURCE_SUPPORT',
    'SourceGridConnectionPoint': '$SOURCE_GRID'
}
print(base64.b64encode(json.dumps(receipt).encode()).decode())
")

PHASE2_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$H_CHANNEL" -n "$CC_NAME" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPRODUCER1_TLS" \
  --tls --cafile "$ORDERER_CA" \
  -c '{"function":"conversion:MintFromConversion","Args":[]}' \
  --transient "{\"MintFromConversion\":\"$MINT_RECEIPT\"}" 2>&1)

echo "$PHASE2_OUT"
sleep 4

# Get the newly minted hydrogen GO ID
set_peer_env hissuer hissuerMSP 8051
MINTED_HGO=$(peer chaincode query -C "$H_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"query:GetCurrentHGOsList","Args":[]}' 2>/dev/null \
  | python3 -c "import sys,json; gos=json.load(sys.stdin); print(gos[-1]['AssetID'] if gos else '')" 2>/dev/null || true)
echo "Minted hydrogen GO: $MINTED_HGO"

if echo "$PHASE2_OUT" | python3 -c "import sys; exit(0 if 'Chaincode invoke successful' in sys.stdin.read() else 1)"; then
  echo "[PASS] Phase 2 MintFromConversion succeeded"
else
  echo "[FAIL] Phase 2 MintFromConversion failed"
  exit 1
fi

# ---- Phase 3: FinalizeLock (electricity-de) ----
echo ""
echo "--- Phase 3: FinalizeLock (electricity-de) ---"
set_peer_env eproducer1 eproducer1MSP 9051

PHASE3_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$E_CHANNEL" -n "$CC_NAME" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls --cafile "$ORDERER_CA" \
  -c "{\"function\":\"conversion:FinalizeLock\",\"Args\":[\"$LOCK_ID\"]}" 2>&1)

echo "$PHASE3_OUT"
sleep 3

if echo "$PHASE3_OUT" | python3 -c "import sys; exit(0 if 'Chaincode invoke successful' in sys.stdin.read() else 1)"; then
  echo "[PASS] Phase 3 FinalizeLock succeeded"
else
  echo "[FAIL] Phase 3 FinalizeLock failed"
fi

echo ""
echo "====== Test Complete ======"
echo "Source GO (electricity): $ACTIVE_GOID"
echo "Conversion Lock: $LOCK_ID"
echo "Minted GO (hydrogen): $MINTED_HGO"
