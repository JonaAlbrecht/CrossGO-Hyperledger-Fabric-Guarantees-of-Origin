#!/bin/bash
# Conversion Operation Scalability Test
# Tests the 3-phase cross-channel conversion protocol with multi-threading

REPO_DIR="/root/hlf-go/repo"
export PATH="$REPO_DIR/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$REPO_DIR/network"
ORDERER_CA="$REPO_DIR/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

# TLS certs
EISSUER_TLS="$REPO_DIR/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt"
EPRODUCER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
EBUYER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt"
HISSUER_TLS="$REPO_DIR/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt"
HPRODUCER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt"
HBUYER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt"

# Results file
RESULTS_FILE="/tmp/conversion-scalability-results.txt"
> "$RESULTS_FILE"

# Set eproducer1 environment
set_eproducer1_env() {
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="eproducer1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE="$EPRODUCER1_TLS"
  export CORE_PEER_MSPCONFIGPATH="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/users/Admin@eproducer1.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:9051"
}

# Set hproducer1 environment
set_hproducer1_env() {
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="hproducer1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE="$HPRODUCER1_TLS"
  export CORE_PEER_MSPCONFIGPATH="$REPO_DIR/network/organizations/peerOrganizations/hproducer1.go-platform.com/users/Admin@hproducer1.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:11051"
}

echo "======================================"
echo "Conversion Scalability Test Suite"
echo "Testing 3-phase cross-channel protocol"
echo "======================================"
echo ""

# First, query available GOs
set_eproducer1_env
echo "=== Querying available Electricity GOs ==="
GO_LIST=$(peer chaincode query \
  -C electricity-de -n golifecycle \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls 2>&1)

echo "$GO_LIST" | head -10
echo ""

# Extract first GO ID for testing (use AssetID, not GOAssetID)
GO_ID=$(echo "$GO_LIST" | grep -o '"AssetID":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$GO_ID" ]; then
  echo "❌ No GOs found! Creating test GO first..."
  
  # Create a test GO
  BACKLOG_JSON='{"AmountMWh":500,"Emissions":250,"ElectricityProductionMethod":"Solar","ElapsedSeconds":3600}'
  BACKLOG_B64=$(echo -n "$BACKLOG_JSON" | base64 | tr -d '\n')
  
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' \
    --transient "{\"eBacklog\":\"$BACKLOG_B64\"}" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA" 2>&1 | grep -E 'committed|invoke'
  
  sleep 2
  
  GO_JSON='{"AmountMWh":500,"Emissions":250,"ElectricityProductionMethod":"Solar","ElapsedSeconds":3600}'
  GO_B64=$(echo -n "$GO_JSON" | base64 | tr -d '\n')
  
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"issuance:CreateElectricityGO","Args":[]}' \
    --transient "{\"eGO\":\"$GO_B64\"}" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA" 2>&1 | grep -E 'committed|invoke'
  
  sleep 2
  
  # Query again
  GO_LIST=$(peer chaincode query \
    -C electricity-de -n golifecycle \
    -c '{"function":"query:GetCurrentEGOsList","Args":[]}' \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --tls 2>&1)
  
  GO_ID=$(echo "$GO_LIST" | grep -o '"AssetID":"[^"]*"' | head -1 | cut -d'"' -f4)
fi

if [ -z "$GO_ID" ]; then
  echo "❌ Failed to find or create GO. Exiting."
  exit 1
fi

echo "✓ Using GO: $GO_ID"
echo ""

# ============================================================
# TEST 1: Phase 1 - LockGOForConversion (Single-threaded)
# ============================================================
echo "======================================"
echo "TEST 1: LockGOForConversion (Phase 1)"
echo "Single-threaded baseline"
echo "======================================"

TEST1_COUNT=10
TEST1_SUCCESS=0
TEST1_START=$(date +%s%N)

set_eproducer1_env

for i in $(seq 1 $TEST1_COUNT); do
  LOCK_JSON="{\"GOAssetID\":\"$GO_ID\",\"DestinationChannel\":\"hydrogen-de\",\"DestinationCarrier\":\"H2\",\"ConversionMethod\":\"Electrolysis\",\"ConversionEfficiency\":0.75,\"OwnerMSP\":\"eproducer1MSP\"}"
  LOCK_B64=$(echo -n "$LOCK_JSON" | base64 | tr -d '\n')
  
  RESULT=$(peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"conversion:LockGOForConversion","Args":[]}' \
    --transient "{\"LockForConversion\":\"$LOCK_B64\"}" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA" 2>&1)
  
  if echo "$RESULT" | grep -q "status (VALID)"; then
    ((TEST1_SUCCESS++))
  fi
done

TEST1_END=$(date +%s%N)
TEST1_DURATION=$(echo "scale=3; ($TEST1_END - $TEST1_START) / 1000000000" | bc)
TEST1_TPS=$(echo "scale=2; $TEST1_SUCCESS / $TEST1_DURATION" | bc)

echo "Completed: $TEST1_SUCCESS/$TEST1_COUNT successful"
echo "Duration: ${TEST1_DURATION}s"
echo "TPS: $TEST1_TPS"
echo ""
echo "TEST1|LockGOForConversion|$TEST1_COUNT|$TEST1_SUCCESS|${TEST1_DURATION}|${TEST1_TPS}" >> "$RESULTS_FILE"

# Get the lock ID from the last operation
LOCK_ID=$(peer chaincode query \
  -C electricity-de -n golifecycle \
  -c '{"function":"conversion:ListConversionLocks","Args":["10",""]}' \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls 2>&1 | grep -o '"LockID":"[^"]*"' | head -1 | cut -d'"' -f4)

echo "✓ Lock ID: $LOCK_ID"
echo ""

# ============================================================
# TEST 2: Phase 2 - MintFromConversion (Single-threaded)
# ============================================================
if [ -n "$LOCK_ID" ]; then
  echo "======================================"
  echo "TEST 2: MintFromConversion (Phase 2)"
  echo "Single-threaded baseline"
  echo "======================================"
  
  TEST2_COUNT=10
  TEST2_SUCCESS=0
  TEST2_START=$(date +%s%N)
  
  set_hproducer1_env
  
  for i in $(seq 1 $TEST2_COUNT); do
    RESULT=$(peer chaincode invoke \
      -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
      -C hydrogen-de -n golifecycle \
      -c "{\"function\":\"conversion:MintFromConversion\",\"Args\":[\"$LOCK_ID\",\"electricity-de\"]}" \
      --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
      --peerAddresses localhost:11051 --tlsRootCertFiles "$HPRODUCER1_TLS" \
      --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER1_TLS" \
      --tls --cafile "$ORDERER_CA" 2>&1)
    
    if echo "$RESULT" | grep -q "status (VALID)"; then
      ((TEST2_SUCCESS++))
    fi
  done
  
  TEST2_END=$(date +%s%N)
  TEST2_DURATION=$(echo "scale=3; ($TEST2_END - $TEST2_START) / 1000000000" | bc)
  TEST2_TPS=$(echo "scale=2; $TEST2_SUCCESS / $TEST2_DURATION" | bc)
  
  echo "Completed: $TEST2_SUCCESS/$TEST2_COUNT successful"
  echo "Duration: ${TEST2_DURATION}s"
  echo "TPS: $TEST2_TPS"
  echo ""
  echo "TEST2|MintFromConversion|$TEST2_COUNT|$TEST2_SUCCESS|${TEST2_DURATION}|${TEST2_TPS}" >> "$RESULTS_FILE"
fi

# ============================================================
# TEST 3: Phase 3 - FinalizeLock (Single-threaded)
# ============================================================
if [ -n "$LOCK_ID" ]; then
  echo "======================================"
  echo "TEST 3: FinalizeLock (Phase 3)"
  echo "Single-threaded baseline"
  echo "======================================"
  
  TEST3_COUNT=10
  TEST3_SUCCESS=0
  TEST3_START=$(date +%s%N)
  
  set_eproducer1_env
  
  for i in $(seq 1 $TEST3_COUNT); do
    RESULT=$(peer chaincode invoke \
      -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
      -C electricity-de -n golifecycle \
      -c "{\"function\":\"conversion:FinalizeLock\",\"Args\":[\"$LOCK_ID\"]}" \
      --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
      --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
      --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
      --tls --cafile "$ORDERER_CA" 2>&1)
    
    if echo "$RESULT" | grep -q "status (VALID)"; then
      ((TEST3_SUCCESS++))
    fi
  done
  
  TEST3_END=$(date +%s%N)
  TEST3_DURATION=$(echo "scale=3; ($TEST3_END - $TEST3_START) / 1000000000" | bc)
  TEST3_TPS=$(echo "scale=2; $TEST3_SUCCESS / $TEST3_DURATION" | bc)
  
  echo "Completed: $TEST3_SUCCESS/$TEST3_COUNT successful"
  echo "Duration: ${TEST3_DURATION}s"
  echo "TPS: $TEST3_TPS"
  echo ""
  echo "TEST3|FinalizeLock|$TEST3_COUNT|$TEST3_SUCCESS|${TEST3_DURATION}|${TEST3_TPS}" >> "$RESULTS_FILE"
fi

# ============================================================
# TEST 4: GetConversionLock Query (Read operation)
# ============================================================
if [ -n "$LOCK_ID" ]; then
  echo "======================================"
  echo "TEST 4: GetConversionLock (Query)"
  echo "Single-threaded baseline"
  echo "======================================"
  
  TEST4_COUNT=100
  TEST4_SUCCESS=0
  TEST4_START=$(date +%s%N)
  
  set_eproducer1_env
  
  for i in $(seq 1 $TEST4_COUNT); do
    RESULT=$(peer chaincode query \
      -C electricity-de -n golifecycle \
      -c "{\"function\":\"conversion:GetConversionLock\",\"Args\":[\"$LOCK_ID\"]}" \
      --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
      --tls 2>&1)
    
    if [ $? -eq 0 ]; then
      ((TEST4_SUCCESS++))
    fi
  done
  
  TEST4_END=$(date +%s%N)
  TEST4_DURATION=$(echo "scale=3; ($TEST4_END - $TEST4_START) / 1000000000" | bc)
  TEST4_TPS=$(echo "scale=2; $TEST4_SUCCESS / $TEST4_DURATION" | bc)
  
  echo "Completed: $TEST4_SUCCESS/$TEST4_COUNT successful"
  echo "Duration: ${TEST4_DURATION}s"
  echo "TPS: $TEST4_TPS"
  echo ""
  echo "TEST4|GetConversionLock|$TEST4_COUNT|$TEST4_SUCCESS|${TEST4_DURATION}|${TEST4_TPS}" >> "$RESULTS_FILE"
fi

# ============================================================
# TEST 5: ListConversionLocks Query (Read operation)
# ============================================================
echo "======================================"
echo "TEST 5: ListConversionLocks (Query)"
echo "Single-threaded baseline"
echo "======================================"

TEST5_COUNT=100
TEST5_SUCCESS=0
TEST5_START=$(date +%s%N)

set_eproducer1_env

for i in $(seq 1 $TEST5_COUNT); do
  RESULT=$(peer chaincode query \
    -C electricity-de -n golifecycle \
    -c '{"function":"conversion:ListConversionLocks","Args":["10",""]}' \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --tls 2>&1)
  
  if [ $? -eq 0 ]; then
    ((TEST5_SUCCESS++))
  fi
done

TEST5_END=$(date +%s%N)
TEST5_DURATION=$(echo "scale=3; ($TEST5_END - $TEST5_START) / 1000000000" | bc)
TEST5_TPS=$(echo "scale=2; $TEST5_SUCCESS / $TEST5_DURATION" | bc)

echo "Completed: $TEST5_SUCCESS/$TEST5_COUNT successful"
echo "Duration: ${TEST5_DURATION}s"
echo "TPS: $TEST5_TPS"
echo ""
echo "TEST5|ListConversionLocks|$TEST5_COUNT|$TEST5_SUCCESS|${TEST5_DURATION}|${TEST5_TPS}" >> "$RESULTS_FILE"

# ============================================================
# SUMMARY
# ============================================================
echo ""
echo "======================================"
echo "SCALABILITY TEST SUMMARY"
echo "======================================"
cat "$RESULTS_FILE"
echo ""
echo "Results saved to: $RESULTS_FILE"
echo "======================================"
