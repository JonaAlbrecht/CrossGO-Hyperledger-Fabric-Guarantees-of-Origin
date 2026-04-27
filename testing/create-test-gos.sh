#!/bin/bash
# Create test Electricity GOs for conversion testing

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

echo "=== Creating 5 test Electricity GOs ==="

for i in {1..5}; do
  echo ""
  echo "--- Creating GO $i ---"
  
  # Add backlog entry
  BACKLOG_DATA=$(cat <<EOF
{
  "AmountMWh": $((100 + i * 10)),
  "Emissions": $((50 + i * 5)),
  "ElectricityProductionMethod": "Solar",
  "ElapsedSeconds": $((3600 + i * 600))
}
EOF
)
  
  echo "Adding backlog entry..."
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"device:AddToBacklogElectricity","Args":[]}' \
    --transient "{\"eBacklog\":\"$BACKLOG_DATA\"}" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA" 2>&1 | grep -E 'committed|Chaincode invoke'
  
  sleep 1
  
  # Create GO from backlog
  GO_DATA=$(cat <<EOF
{
  "AmountMWh": $((100 + i * 10)),
  "Emissions": $((50 + i * 5)),
  "ElectricityProductionMethod": "Solar",
  "ElapsedSeconds": $((3600 + i * 600))
}
EOF
)
  
  echo "Creating Electricity GO..."
  peer chaincode invoke \
    -o localhost:7050 --ordererTLSHostnameOverride orderer1.go-platform.com \
    -C electricity-de -n golifecycle \
    -c '{"function":"device:CreateElectricityGO","Args":[]}' \
    --transient "{\"eGO\":\"$GO_DATA\"}" \
    --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
    --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
    --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER1_TLS" \
    --tls --cafile "$ORDERER_CA" 2>&1 | grep -E 'committed|Chaincode invoke'
  
  sleep 1
done

echo ""
echo "=== Querying created GOs ==="
peer chaincode query \
  -C electricity-de -n golifecycle \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --tls 2>&1 | head -20

echo ""
echo "=== Test GO creation complete ==="
