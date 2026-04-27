#!/bin/bash
# Create Electricity GOs with verbose output

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

echo "=== Creating 3 test Electricity GOs with full output ==="

for i in {1..3}; do
  echo ""
  echo "--- Creating GO $i ---"
  
  # Add backlog entry
  echo "Step 1: Adding backlog..."
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"device:AddToBacklogElectricity","Args":[]}' \
    --transient '{"eBacklog":"{\"AmountMWh\":500,\"Emissions\":250,\"ElectricityProductionMethod\":\"Solar\",\"ElapsedSeconds\":3600}"}' \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA"
  
  echo ""
  sleep 1
  
  # Create GO from backlog
  echo "Step 2: Creating GO..."
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"device:CreateElectricityGO","Args":[]}' \
    --transient '{"eGO":"{\"AmountMWh\":500,\"Emissions\":250,\"ElectricityProductionMethod\":\"Solar\",\"ElapsedSeconds\":3600}"}' \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA"
  
  echo ""
  sleep 1
done

echo ""
echo "=== Querying created GOs ==="
peer chaincode query \
  -C electricity-de -n golifecycle \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls

echo ""
echo "=== GO creation complete ==="
