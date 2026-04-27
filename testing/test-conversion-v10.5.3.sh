#!/bin/bash
# Full 3-phase cross-channel conversion test for v10.5.3
set -euo pipefail
REPO=/root/hlf-go/repo
ND=$REPO/network
export PATH=$REPO/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$ND

CC=golifecycle
E_CH=electricity-de
H_CH=hydrogen-de
ORDERER_CA=$ND/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

EISSUER_TLS=$ND/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
EPRODUCER1_TLS=$ND/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
HISSUER_TLS=$ND/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt
HPRODUCER1_TLS=$ND/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt
EBUYER1_TLS=$ND/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt
HBUYER1_TLS=$ND/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt

E_PEERS="--peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS"
H_PEERS="--peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS --peerAddresses localhost:14051 --tlsRootCertFiles $HBUYER1_TLS"

set_env() {
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID=$2
  export CORE_PEER_TLS_ROOTCERT_FILE=$ND/organizations/peerOrganizations/${1}.go-platform.com/peers/peer0.${1}.go-platform.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/${1}.go-platform.com/users/Admin@${1}.go-platform.com/msp
  export CORE_PEER_ADDRESS=localhost:${3}
}

PASS=0; FAIL=0
check() {
  local label=$1; local out=$2
  if echo "$out" | grep -q "invoke successful"; then
    echo "[PASS] $label"; ((PASS++))
  else
    echo "[FAIL] $label"; echo "$out" | grep -i 'error\|Error\|failed' | head -3; ((FAIL++))
  fi
}
check_query() {
  local label=$1; local out=$2
  if echo "$out" | grep -qvE '^Error|^$'; then
    echo "[PASS] $label"; ((PASS++))
  else
    echo "[FAIL] $label"; echo "$out" | head -3; ((FAIL++))
  fi
}

# =====================================================================
# PHASE 1: LockGOForConversion (electricity-de, as eproducer1MSP)
# =====================================================================
echo ""
echo "====== PHASE 1: LockGOForConversion (electricity-de) ======"
set_env eissuer eissuerMSP 7051
GO_ID=$(peer chaincode query -C $E_CH -n $CC -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null \
  | python3 -c "import sys,json; gos=json.load(sys.stdin); print(next(g['AssetID'] for g in gos if g['Status']=='active'))")
echo "  Using GO: $GO_ID"

LOCK_IN=$(python3 -c "
import json,base64
d={
  'GOAssetID':'$GO_ID',
  'DestinationChannel':'$H_CH',
  'DestinationCarrier':'hydrogen',
  'ConversionMethod':'electrolysis',
  'ConversionEfficiency':0.65,
  'OwnerMSP':'eproducer1MSP',
  'DestinationOwnerMSP':'hproducer1MSP'
}
print(base64.b64encode(json.dumps(d).encode()).decode())
")

set_env eproducer1 eproducer1MSP 9051
P1_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C $E_CH -n $CC --tls --cafile $ORDERER_CA \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  -c '{"function":"conversion:LockGOForConversion","Args":[]}' \
  --transient "{\"LockForConversion\":\"$LOCK_IN\"}" 2>&1)
check "Phase 1: LockGOForConversion" "$P1_OUT"
sleep 4

# Query the lock to get TxID (stored in lock by v10.5.3)
set_env eissuer eissuerMSP 7051
LOCK_ID=$(peer chaincode query -C $E_CH -n $CC \
  -c '{"function":"conversion:ListConversionLocks","Args":["10",""]}' 2>/dev/null \
  | python3 -c "
import sys,json
locks=json.load(sys.stdin)
locked=[l for l in locks if l.get('Status')=='locked']
print(locked[0]['LockID'] if locked else '')
")
echo "  Lock ID: $LOCK_ID"

LOCK_JSON=$(peer chaincode query -C $E_CH -n $CC \
  -c "{\"function\":\"conversion:GetConversionLock\",\"Args\":[\"$LOCK_ID\"]}" 2>/dev/null)
echo "  Lock record:"
echo "$LOCK_JSON" | python3 -m json.tool 2>/dev/null

LOCK_HASH=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['LockReceiptHash'])")
TXID=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['TxID'])")
EISSUER_MSP=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['SourceIssuerMSP'])")

echo "  LockReceiptHash: $LOCK_HASH"
echo "  TxID:            $TXID"

check_query "GetConversionLock query" "$LOCK_JSON"

# =====================================================================
# PHASE 2: MintFromConversion (hydrogen-de, as hissuerMSP)
# =====================================================================
echo ""
echo "====== PHASE 2: MintFromConversion (hydrogen-de) ======"

# Read source GO private data
set_env eproducer1 eproducer1MSP 9051
QUERY_IN=$(python3 -c "
import json,base64
d={'Collection':'privateDetails-eproducer1MSP','EGOID':'$GO_ID'}
print(base64.b64encode(json.dumps(d).encode()).decode())
")
PRIVATE_GO=$(peer chaincode query -C $E_CH -n $CC \
  -c '{"function":"query:ReadPrivateEGO","Args":[]}' \
  --transient "{\"QueryInput\":\"$QUERY_IN\"}" 2>/dev/null || echo '{}')

SRC_AMOUNT=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('AmountMWh',100))" 2>/dev/null || echo 100)
SRC_EMISSIONS=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Emissions',5.0))" 2>/dev/null || echo 5.0)
SRC_METHOD=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('ElectricityProductionMethod','solar_pv'))" 2>/dev/null || echo solar_pv)
SRC_DEVICE=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('DeviceID',''))" 2>/dev/null || echo "")
SRC_CREATED=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('CreationDateTime',1777000000))" 2>/dev/null || echo 1777000000)

set_env eissuer eissuerMSP 7051
PUBLIC_GO=$(peer chaincode query -C $E_CH -n $CC \
  -c "{\"function\":\"query:ReadPublicEGO\",\"Args\":[\"$GO_ID\"]}" 2>/dev/null || echo '{}')
SRC_COUNTRY=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('CountryOfOrigin','DE'))" 2>/dev/null || echo DE)
SRC_START=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('ProductionPeriodStart',1777000000))" 2>/dev/null || echo 1777000000)
SRC_END=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('ProductionPeriodEnd',1777001000))" 2>/dev/null || echo 1777001000)
SRC_SUPPORT=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SupportScheme','none'))" 2>/dev/null || echo none)
SRC_GRID=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('GridConnectionPoint','10YDE-VE-------2'))" 2>/dev/null || echo 10YDE-VE-------2)

MINT_IN=$(python3 -c "
import json,base64
r={
  'LockID':'$LOCK_ID',
  'GOAssetID':'$GO_ID',
  'SourceChannel':'$E_CH',
  'SourceCarrier':'electricity',
  'DestinationChannel':'$H_CH',
  'DestinationCarrier':'hydrogen',
  'ConversionMethod':'electrolysis',
  'ConversionEfficiency':0.65,
  'OwnerMSP':'eproducer1MSP',
  'DestinationOwnerMSP':'hproducer1MSP',
  'SourceIssuerMSP':'$EISSUER_MSP',
  'LockReceiptHash':'$LOCK_HASH',
  'TxID':'$TXID',
  'SourceAmount':$SRC_AMOUNT,
  'SourceAmountUnit':'MWh',
  'SourceEmissions':$SRC_EMISSIONS,
  'SourceProductionMethod':'$SRC_METHOD',
  'SourceDeviceID':'$SRC_DEVICE',
  'SourceCreationDateTime':$SRC_CREATED,
  'SourceConsumptionDecls':[],
  'SourceCountryOfOrigin':'$SRC_COUNTRY',
  'SourceProductionStart':$SRC_START,
  'SourceProductionEnd':$SRC_END,
  'SourceSupportScheme':'$SRC_SUPPORT',
  'SourceGridConnectionPoint':'$SRC_GRID'
}
print(base64.b64encode(json.dumps(r).encode()).decode())
")

set_env hissuer hissuerMSP 8051
P2_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C $H_CH -n $CC --tls --cafile $ORDERER_CA \
  --peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS \
  --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS \
  -c '{"function":"conversion:MintFromConversion","Args":[]}' \
  --transient "{\"MintFromConversion\":\"$MINT_IN\"}" 2>&1)
echo "$P2_OUT"
check "Phase 2: MintFromConversion" "$P2_OUT"
sleep 4

# Get minted HGO id
set_env hissuer hissuerMSP 8051
MINTED_HGO=$(peer chaincode query -C $H_CH -n $CC \
  -c '{"function":"query:GetCurrentHGOsList","Args":[]}' 2>/dev/null \
  | python3 -c "import sys,json; gos=json.load(sys.stdin); print(gos[-1]['AssetID'] if gos else 'unknown')" 2>/dev/null || echo "unknown")
echo "  Minted hydrogen GO: $MINTED_HGO"

# =====================================================================
# PHASE 3: FinalizeLock (electricity-de)
# =====================================================================
echo ""
echo "====== PHASE 3: FinalizeLock (electricity-de) ======"

FINALIZE_IN=$(python3 -c "
import json,base64
d={
  'LockID':'$LOCK_ID',
  'MintedAssetID':'$MINTED_HGO',
  'DestinationChannel':'$H_CH',
  'OwnerMSP':'eproducer1MSP'
}
print(base64.b64encode(json.dumps(d).encode()).decode())
")

set_env eproducer1 eproducer1MSP 9051
P3_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C $E_CH -n $CC --tls --cafile $ORDERER_CA \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  -c '{"function":"conversion:FinalizeLock","Args":[]}' \
  --transient "{\"FinalizeLock\":\"$FINALIZE_IN\"}" 2>&1)
echo "$P3_OUT"
check "Phase 3: FinalizeLock" "$P3_OUT"
sleep 3

# =====================================================================
# FINAL STATE VERIFICATION
# =====================================================================
echo ""
echo "====== Final State Verification ======"
set_env eissuer eissuerMSP 7051

FINAL_LOCK=$(peer chaincode query -C $E_CH -n $CC \
  -c "{\"function\":\"conversion:GetConversionLock\",\"Args\":[\"$LOCK_ID\"]}" 2>/dev/null || echo '{}')
FINAL_LOCK_STATUS=$(echo "$FINAL_LOCK" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Status','unknown'))" 2>/dev/null)
FINAL_MINTED=$(echo "$FINAL_LOCK" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('MintedAssetID',''))" 2>/dev/null)

FINAL_GO=$(peer chaincode query -C $E_CH -n $CC \
  -c "{\"function\":\"query:ReadPublicEGO\",\"Args\":[\"$GO_ID\"]}" 2>/dev/null || echo '{}')
FINAL_GO_STATUS=$(echo "$FINAL_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Status','unknown'))" 2>/dev/null)

set_env hissuer hissuerMSP 8051
H2_GOS=$(peer chaincode query -C $H_CH -n $CC \
  -c '{"function":"query:GetCurrentHGOsList","Args":[]}' 2>/dev/null | python3 -c "import sys,json; gos=json.load(sys.stdin); print(len(gos))" 2>/dev/null || echo 0)

echo "  Source GO ($GO_ID):   status = $FINAL_GO_STATUS"
echo "  Lock ($LOCK_ID):      status = $FINAL_LOCK_STATUS"
echo "  Minted H2 GO:         $FINAL_MINTED"
echo "  H2 GOs on hydrogen-de: $H2_GOS"

echo ""
echo "====== Conversion Test Results ======"
echo "  PASS: $PASS   FAIL: $FAIL"
if [ "$FAIL" -eq 0 ] && [ "$FINAL_LOCK_STATUS" = "consumed" ] && [ "$FINAL_GO_STATUS" = "consumed" ]; then
  echo "  ✅ Full 3-phase cross-channel conversion SUCCESSFUL"
else
  echo "  ⚠️  Some steps failed — see output above"
fi
