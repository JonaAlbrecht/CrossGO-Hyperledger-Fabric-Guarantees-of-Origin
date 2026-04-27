#!/bin/bash
# upgrade-v10.5-and-test-conversion.sh
# Run on the Hetzner server at /root/hlf-go/repo/
# Upgrades chaincode to v10.5 (DestinationOwnerMSP fix + compilation fixes)
# Then tests the full 3-phase cross-channel conversion protocol
set -euo pipefail

REPO_DIR=/root/hlf-go/repo
NETWORK_DIR=$REPO_DIR/network
CC_SRC=$REPO_DIR/chaincode
export PATH=$REPO_DIR/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$NETWORK_DIR

E_CHANNEL=electricity-de
H_CHANNEL=hydrogen-de
ORDERER_CA=$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem
CC_NAME=golifecycle
CC_VERSION=10.5
# Query current sequence and increment
echo "Querying current committed sequence..."
set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

set_peer_env eissuer eissuerMSP 7051
CURRENT_SEQ=$(peer lifecycle chaincode querycommitted -C "$E_CHANNEL" -n "$CC_NAME" --output json 2>/dev/null | jq -r '.sequence // 0')
CC_SEQUENCE=$((CURRENT_SEQ + 1))
echo "Current sequence: $CURRENT_SEQ → upgrading to sequence: $CC_SEQUENCE"

E_ORGS=("eissuer:eissuerMSP:7051" "eproducer1:eproducer1MSP:9051" "ebuyer1:ebuyer1MSP:13051")
H_ORGS=("hissuer:hissuerMSP:8051" "hproducer1:hproducer1MSP:11051" "hbuyer1:hbuyer1MSP:14051")
ALL_ORGS=("${E_ORGS[@]}" "${H_ORGS[@]}")

E_COLL=$REPO_DIR/collections/collection-config-electricity-de.json
H_COLL=$REPO_DIR/collections/collection-config-hydrogen-de.json

# ============================================================
# STEP 1: Build and package chaincode
# ============================================================
echo ""
echo "====== [1/5] Building chaincode v${CC_VERSION} ======"
cd "$CC_SRC"
go build ./... 2>&1
echo "Build OK"

cd "$REPO_DIR"
peer lifecycle chaincode package "${CC_NAME}_${CC_VERSION}.tar.gz" \
  --path "$CC_SRC" \
  --lang golang \
  --label "${CC_NAME}_${CC_VERSION}"
echo "Package created: ${CC_NAME}_${CC_VERSION}.tar.gz"

# ============================================================
# STEP 2: Install on all 6 peers
# ============================================================
echo ""
echo "====== [2/5] Installing on all 6 peers ======"
for entry in "${ALL_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  echo "  Installing on peer0.${org}..."
  peer lifecycle chaincode install "${CC_NAME}_${CC_VERSION}.tar.gz" 2>&1 | tail -3
done

# ============================================================
# STEP 3: Approve on both channels
# ============================================================
echo ""
echo "====== [3/5] Approving on both channels ======"

set_peer_env eissuer eissuerMSP 7051
PKG_ID=$(peer lifecycle chaincode queryinstalled --output json | jq -r ".installed_chaincodes[] | select(.label==\"${CC_NAME}_${CC_VERSION}\") | .package_id")
echo "Package ID: $PKG_ID"

# Approve for electricity-de (all 3 E orgs)
for entry in "${E_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  echo "  Approving $msp for $E_CHANNEL seq $CC_SEQUENCE..."
  peer lifecycle chaincode approveformyorg \
    -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --channelID "$E_CHANNEL" --name "$CC_NAME" \
    --version "$CC_VERSION" --package-id "$PKG_ID" \
    --sequence "$CC_SEQUENCE" \
    --collections-config "$E_COLL" \
    --tls --cafile "$ORDERER_CA"
done

# Approve for hydrogen-de (all 3 H orgs)
for entry in "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  echo "  Approving $msp for $H_CHANNEL seq $CC_SEQUENCE..."
  peer lifecycle chaincode approveformyorg \
    -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --channelID "$H_CHANNEL" --name "$CC_NAME" \
    --version "$CC_VERSION" --package-id "$PKG_ID" \
    --sequence "$CC_SEQUENCE" \
    --collections-config "$H_COLL" \
    --tls --cafile "$ORDERER_CA"
done

# ============================================================
# STEP 4: Commit on both channels
# ============================================================
echo ""
echo "====== [4/5] Committing on both channels ======"

# Build electricity-de peer conn args
E_PEER_ARGS=""
for entry in "${E_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  E_PEER_ARGS="$E_PEER_ARGS --peerAddresses localhost:${port}"
  E_PEER_ARGS="$E_PEER_ARGS --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
done

# Build hydrogen-de peer conn args
H_PEER_ARGS=""
for entry in "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  H_PEER_ARGS="$H_PEER_ARGS --peerAddresses localhost:${port}"
  H_PEER_ARGS="$H_PEER_ARGS --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
done

set_peer_env eissuer eissuerMSP 7051
echo "  Committing on $E_CHANNEL..."
peer lifecycle chaincode commit \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --channelID "$E_CHANNEL" --name "$CC_NAME" \
  --version "$CC_VERSION" --sequence "$CC_SEQUENCE" \
  --collections-config "$E_COLL" \
  --tls --cafile "$ORDERER_CA" \
  $E_PEER_ARGS

set_peer_env hissuer hissuerMSP 8051
echo "  Committing on $H_CHANNEL..."
peer lifecycle chaincode commit \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --channelID "$H_CHANNEL" --name "$CC_NAME" \
  --version "$CC_VERSION" --sequence "$CC_SEQUENCE" \
  --collections-config "$H_COLL" \
  --tls --cafile "$ORDERER_CA" \
  $H_PEER_ARGS

echo "  Verifying..."
set_peer_env eissuer eissuerMSP 7051
peer lifecycle chaincode querycommitted -C "$E_CHANNEL" -n "$CC_NAME" --output json | jq '{version,sequence}'
set_peer_env hissuer hissuerMSP 8051
peer lifecycle chaincode querycommitted -C "$H_CHANNEL" -n "$CC_NAME" --output json | jq '{version,sequence}'

sleep 3

# ============================================================
# STEP 5: End-to-end 3-phase conversion test
# ============================================================
echo ""
echo "====== [5/5] End-to-end Cross-Channel Conversion Test ======"

EISSUER_TLS="$NETWORK_DIR/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt"
EPRODUCER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
EBUYER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt"
HISSUER_TLS="$NETWORK_DIR/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt"
HPRODUCER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt"
HBUYER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt"

# ---- Phase 0: Find an active electricity GO ----
echo ""
echo "--- Phase 0: Finding active electricity GOs ---"
set_peer_env eissuer eissuerMSP 7051
ACTIVE_GOID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null \
  | python3 -c "import sys,json; gos=json.load(sys.stdin); active=[g['AssetID'] for g in gos if g.get('Status')=='active']; print(active[0] if active else '')" 2>/dev/null || true)

if [ -z "$ACTIVE_GOID" ]; then
  echo "No active electricity GOs found. Creating a new one first..."
  # Add backlog
  set_peer_env eproducer1 eproducer1MSP 9051
  EBACKLOG=$(echo -n '{"AmountMWh":200,"Emissions":10.0,"ElectricityProductionMethod":"solar_pv","ElapsedSeconds":3600}' | base64 -w0)
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -C "$E_CHANNEL" -n "$CC_NAME" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --tls --cafile "$ORDERER_CA" \
    -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' \
    --transient "{\"eBacklog\":\"$EBACKLOG\"}"
  sleep 3

  # Create electricity GO
  EGO_IN=$(echo -n '{"AmountMWh":150,"Emissions":7.5,"ElectricityProductionMethod":"solar_pv","ElapsedSeconds":3600}' | base64 -w0)
  set_peer_env eproducer1 eproducer1MSP 9051
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -C "$E_CHANNEL" -n "$CC_NAME" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --tls --cafile "$ORDERER_CA" \
    -c '{"function":"issuance:CreateElectricityGO","Args":[]}' \
    --transient "{\"eGO\":\"$EGO_IN\"}"
  sleep 3

  set_peer_env eissuer eissuerMSP 7051
  ACTIVE_GOID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
    -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null \
    | python3 -c "import sys,json; gos=json.load(sys.stdin); active=[g['AssetID'] for g in gos if g.get('Status')=='active']; print(active[0] if active else '')" 2>/dev/null || true)
fi

echo "Using GO: $ACTIVE_GOID"
if [ -z "$ACTIVE_GOID" ]; then
  echo "ERROR: No active electricity GO available. Aborting."
  exit 1
fi

# ---- Phase 1: Lock GO for Conversion (electricity-de) ----
echo ""
echo "--- Phase 1: LockGOForConversion (electricity-de) ---"
set_peer_env eproducer1 eproducer1MSP 9051

LOCK_INPUT=$(echo -n "{\"GOAssetID\":\"${ACTIVE_GOID}\",\"DestinationChannel\":\"${H_CHANNEL}\",\"DestinationCarrier\":\"hydrogen\",\"ConversionMethod\":\"electrolysis\",\"ConversionEfficiency\":0.65,\"OwnerMSP\":\"eproducer1MSP\",\"DestinationOwnerMSP\":\"hproducer1MSP\"}" | base64 -w0)

PHASE1_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$E_CHANNEL" -n "$CC_NAME" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls --cafile "$ORDERER_CA" \
  -c '{"function":"conversion:LockGOForConversion","Args":[]}' \
  --transient "{\"LockForConversion\":\"$LOCK_INPUT\"}" 2>&1)

echo "$PHASE1_OUT"
if ! echo "$PHASE1_OUT" | grep -q "Chaincode invoke successful"; then
  echo "ERROR: Phase 1 failed"
  exit 1
fi
sleep 4

# Get the lock ID from the ledger
set_peer_env eissuer eissuerMSP 7051
LOCK_ID=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"conversion:ListConversionLocks","Args":["10",""]}' 2>/dev/null \
  | python3 -c "import sys,json; locks=json.load(sys.stdin); active=[l for l in locks if l.get('Status')=='locked' and l.get('GOAssetID')=='${ACTIVE_GOID}']; print(active[0]['LockID'] if active else '')" 2>/dev/null || true)

echo "Lock ID: $LOCK_ID"
if [ -z "$LOCK_ID" ]; then
  echo "ERROR: Could not find lock for $ACTIVE_GOID"
  exit 1
fi

# Get the full lock details
LOCK_JSON=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c "{\"function\":\"conversion:GetConversionLock\",\"Args\":[\"$LOCK_ID\"]}" 2>/dev/null)
echo "Lock details:"
echo "$LOCK_JSON" | python3 -m json.tool 2>/dev/null || echo "$LOCK_JSON"

LOCK_HASH=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('LockReceiptHash',''))" 2>/dev/null || true)
echo "LockReceiptHash: $LOCK_HASH"

# Extract TxID from the newest block on electricity-de
echo "Extracting TxID from latest block on $E_CHANNEL..."
peer channel fetch newest /tmp/elec_latest.block \
  -C "$E_CHANNEL" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile "$ORDERER_CA" 2>/dev/null

TXID=$(configtxlator proto_decode --input /tmp/elec_latest.block --type common.Block 2>/dev/null \
  | python3 -c "
import sys, json
block = json.load(sys.stdin)
data = block.get('data',{}).get('data',[])
if data:
    ch = data[0].get('payload',{}).get('header',{}).get('channel_header',{})
    print(ch.get('tx_id',''))
else:
    print('')
" 2>/dev/null || true)

echo "TxID from latest block: $TXID"

# Verify hash matches (compute locally)
SOURCE_DATA=$(python3 -c "
import hashlib, sys
lock_id = '$LOCK_ID'
go_id = '$ACTIVE_GOID'
src_ch = '$E_CHANNEL'
dst_ch = '$H_CHANNEL'
dst_carrier = 'hydrogen'
eff = 0.65
owner_msp = 'eproducer1MSP'
tx_id = '$TXID'
data = f'{lock_id}||{go_id}||{src_ch}||{dst_ch}||{dst_carrier}||{eff:.6f}||{owner_msp}||{tx_id}'
print('Hash input:', data, file=sys.stderr)
print(hashlib.sha256(data.encode()).hexdigest())
" 2>/tmp/hash_debug.txt || true)
echo "Computed hash: $SOURCE_DATA"
echo "Stored hash:   $LOCK_HASH"
cat /tmp/hash_debug.txt

if [ "$SOURCE_DATA" != "$LOCK_HASH" ]; then
  echo "WARNING: Hash mismatch — TxID might be wrong or block parse failed"
  echo "Proceeding anyway with extracted TxID..."
fi

# ---- Get source GO private details for the receipt ----
echo ""
echo "--- Reading source GO private details ---"
set_peer_env eproducer1 eproducer1MSP 9051
QUERY_IN=$(echo -n "{\"Collection\":\"privateDetails-eproducer1MSP\",\"EGOID\":\"$ACTIVE_GOID\"}" | base64 -w0)
PRIVATE_GO=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c '{"function":"query:ReadPrivateEGO","Args":[]}' \
  --transient "{\"QueryInput\":\"$QUERY_IN\"}" 2>/dev/null || echo '{}')
echo "Private GO: $PRIVATE_GO" | head -5

SOURCE_AMOUNT=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('AmountMWh',100))" 2>/dev/null || echo "100")
SOURCE_EMISSIONS=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Emissions',5.0))" 2>/dev/null || echo "5.0")
SOURCE_PROD_METHOD=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('ElectricityProductionMethod','solar_pv'))" 2>/dev/null || echo "solar_pv")
SOURCE_DEVICE=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('DeviceID',''))" 2>/dev/null || echo "")
SOURCE_CREATION=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('CreationDateTime',1745000000))" 2>/dev/null || echo "1745000000")
SOURCE_CONSUMP=$(echo "$PRIVATE_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(json.dumps(d.get('ConsumptionDeclarations',[])))" 2>/dev/null || echo "[]")

# Get public GO metadata
set_peer_env eissuer eissuerMSP 7051
PUBLIC_GO=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c "{\"function\":\"query:ReadPublicEGO\",\"Args\":[\"$ACTIVE_GOID\"]}" 2>/dev/null || echo '{}')
SOURCE_COUNTRY=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('CountryOfOrigin','DE'))" 2>/dev/null || echo "DE")
SOURCE_PROD_START=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('ProductionPeriodStart',1744000000))" 2>/dev/null || echo "1744000000")
SOURCE_PROD_END=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('ProductionPeriodEnd',1745000000))" 2>/dev/null || echo "1745000000")
SOURCE_SUPPORT=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SupportScheme','none'))" 2>/dev/null || echo "none")
SOURCE_GRID=$(echo "$PUBLIC_GO" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('GridConnectionPoint','10YDE-VE-------2'))" 2>/dev/null || echo "10YDE-VE-------2")

echo "Source GO: amount=$SOURCE_AMOUNT MWh, emissions=$SOURCE_EMISSIONS, method=$SOURCE_PROD_METHOD"

# ---- Phase 2: MintFromConversion (hydrogen-de) ----
echo ""
echo "--- Phase 2: MintFromConversion (hydrogen-de) ---"
set_peer_env hissuer hissuerMSP 8051

EISSUER_MSP=$(echo "$LOCK_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('SourceIssuerMSP','eissuerMSP'))" 2>/dev/null || echo "eissuerMSP")

MINT_RECEIPT=$(python3 -c "
import json, sys
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
print(json.dumps(receipt))
" | base64 -w0)

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

if echo "$PHASE2_OUT" | grep -q "Chaincode invoke successful"; then
  echo "[PASS] Phase 2 MintFromConversion succeeded"
else
  echo "[FAIL] Phase 2 MintFromConversion failed"
  # Continue to attempt Phase 3 anyway with placeholder
  MINTED_HGO="${MINTED_HGO:-hGO_unknown}"
fi

# ---- Phase 3: FinalizeLock (electricity-de) ----
echo ""
echo "--- Phase 3: FinalizeLock (electricity-de) ---"
set_peer_env eproducer1 eproducer1MSP 9051

FINALIZE_INPUT=$(echo -n "{\"LockID\":\"$LOCK_ID\",\"MintedAssetID\":\"$MINTED_HGO\",\"DestinationChannel\":\"$H_CHANNEL\",\"OwnerMSP\":\"eproducer1MSP\"}" | base64 -w0)

PHASE3_OUT=$(peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C "$E_CHANNEL" -n "$CC_NAME" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls --cafile "$ORDERER_CA" \
  -c '{"function":"conversion:FinalizeLock","Args":[]}' \
  --transient "{\"FinalizeLock\":\"$FINALIZE_INPUT\"}" 2>&1)

echo "$PHASE3_OUT"
sleep 3

# ---- Verify final state ----
echo ""
echo "--- Verifying final state ---"
set_peer_env eissuer eissuerMSP 7051
FINAL_LOCK=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c "{\"function\":\"conversion:GetConversionLock\",\"Args\":[\"$LOCK_ID\"]}" 2>/dev/null || echo '{}')
FINAL_STATUS=$(echo "$FINAL_LOCK" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Status','unknown'))" 2>/dev/null || echo "unknown")
echo "Lock final status: $FINAL_STATUS"

FINAL_GO_STATUS=$(peer chaincode query -C "$E_CHANNEL" -n "$CC_NAME" \
  -c "{\"function\":\"query:ReadPublicEGO\",\"Args\":[\"$ACTIVE_GOID\"]}" 2>/dev/null \
  | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Status','unknown'))" 2>/dev/null || echo "unknown")
echo "Source GO final status: $FINAL_GO_STATUS"

# Summary
echo ""
echo "====== Conversion Test Summary ======"
echo "Source GO:      $ACTIVE_GOID  → $FINAL_GO_STATUS"
echo "Lock:           $LOCK_ID      → $FINAL_STATUS"
echo "Minted H2 GO:   $MINTED_HGO"
echo ""
if [ "$FINAL_STATUS" = "consumed" ] && [ "$FINAL_GO_STATUS" = "consumed" ]; then
  echo "✅ 3-phase cross-channel conversion SUCCESSFUL"
else
  echo "⚠️  Conversion incomplete — check output above for errors"
fi
