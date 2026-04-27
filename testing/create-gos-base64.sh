#!/bin/bash
# Create Electricity GOs with properly base64-encoded transient data

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

# TLS certs for 3-org SBE endorsement
EISSUER_TLS="$REPO_DIR/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt"
EPRODUCER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
EBUYER1_TLS="$REPO_DIR/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt"

echo "=== Creating 3 test Electricity GOs ==="

for i in {1..3}; do
  echo ""
  echo "--- Creating GO $i ---"
  
  # Prepare transient data
  BACKLOG_JSON='{"AmountMWh":500,"Emissions":250,"ElectricityProductionMethod":"Solar","ElapsedSeconds":3600}'
  BACKLOG_B64=$(echo -n "$BACKLOG_JSON" | base64 | tr -d '\n')
  
  # Add backlog entry
  echo "Step 1: Adding backlog..."
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' \
    --transient "{\"eBacklog\":\"$BACKLOG_B64\"}" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA" 2>&1 | grep -E 'committed|invoke|Error'
  
  sleep 1
  
  # Create GO from backlog
  GO_JSON='{"AmountMWh":500,"Emissions":250,"ElectricityProductionMethod":"Solar","ElapsedSeconds":3600}'
  GO_B64=$(echo -n "$GO_JSON" | base64 | tr -d '\n')
  
  echo "Step 2: Creating GO..."
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"issuance:CreateElectricityGO","Args":[]}' \
    --transient "{\"eGO\":\"$GO_B64\"}" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA" 2>&1 | grep -E 'committed|invoke|Error'
  
  sleep 1
done

echo ""
echo "=== Querying created GOs ==="
peer chaincode query \
  -C electricity-de -n golifecycle \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls | head -20

echo ""
echo "=== GO creation complete ==="
