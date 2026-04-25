#!/bin/bash
# Test write operations using docker exec from inside peer containers
# This approach resolves the hostname issue by running commands from within the Docker network

set -e

CHANNEL_NAME="electricity-de"
CC_NAME="golifecycle"
ORDERER_ADDR="orderer1.go-platform.com:7050"
ORDERER_CA="/organizations/ordererOrganizations/go-platform.com/tlsca/tlsca.go-platform.com-cert.pem"

echo "==============================================="
echo "Testing Write Operations via Docker Exec"
echo "==============================================="
echo ""

# Test 1: PublishOracleData (requires issuer role) - execute from eissuer peer
echo "Test 1: PublishOracleData (issuer role)"
echo "---------------------------------------"

ORACLE_DATA=$(cat <<'EOF'
{
  "CarrierType": "electricity",
  "Zone": "DE-North",
  "PeriodStart": "2026-04-25T12:00:00Z",
  "PeriodEnd": "2026-04-25T13:00:00Z",
  "ProductionMethod": "wind_onshore",
  "EmissionFactor": 0.05,
  "PricePerMWh": 45.5
}
EOF
)

ORACLE_B64=$(echo -n "$ORACLE_DATA" | base64 -w0)

docker exec peer0.eissuer.go-platform.com peer chaincode invoke \
  -o ${ORDERER_ADDR} \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile ${ORDERER_CA} \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"oracle:PublishOracleData","Args":[]}' \
  --transient "{\"oracleData\":\"${ORACLE_B64}\"}" \
  --peerAddresses peer0.eissuer.go-platform.com:7051 \
  --tlsRootCertFiles /organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt

echo "✅ PublishOracleData successful (executed by eissuer)"
echo ""

# Test 2: AddToBacklogElectricity (requires producer role) - execute from eproducer1 peer
echo "Test 2: AddToBacklogElectricity (producer role)"
echo "-----------------------------------------------"

BACKLOG_DATA=$(cat <<'EOF'
{
  "AmountMWh": 150.5,
  "Emissions": 7.5,
  "ElectricityProductionMethod": "wind_onshore",
  "ElapsedSeconds": 900
}
EOF
)

BACKLOG_B64=$(echo -n "$BACKLOG_DATA" | base64 -w0)

docker exec peer0.eproducer1.go-platform.com peer chaincode invoke \
  -o ${ORDERER_ADDR} \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile ${ORDERER_CA} \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' \
  --transient "{\"eBacklog\":\"${BACKLOG_B64}\"}" \
  --peerAddresses peer0.eproducer1.go-platform.com:8051 \
  --tlsRootCertFiles /organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt

echo "✅ AddToBacklogElectricity successful (executed by eproducer1)"
echo ""

# Test 3: CreateElectricityGO (requires producer role + SBE endorsement)
echo "Test 3: CreateElectricityGO (producer role, SBE)"
echo "------------------------------------------------"

EGO_DATA=$(cat <<'EOF'
{
  "AmountMWh": 200.0,
  "Emissions": 10.0,
  "ElectricityProductionMethod": "solar_pv",
  "ElapsedSeconds": 3600
}
EOF
)

EGO_B64=$(echo -n "$EGO_DATA" | base64 -w0)

docker exec peer0.eproducer1.go-platform.com peer chaincode invoke \
  -o ${ORDERER_ADDR} \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile ${ORDERER_CA} \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"issuance:CreateElectricityGO","Args":[]}' \
  --transient "{\"eGO\":\"${EGO_B64}\"}" \
  --peerAddresses peer0.eissuer.go-platform.com:7051 \
  --tlsRootCertFiles /organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  --peerAddresses peer0.eproducer1.go-platform.com:8051 \
  --tlsRootCertFiles /organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt \
  --peerAddresses peer0.ebuyer1.go-platform.com:9051 \
  --tlsRootCertFiles /organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt

echo "✅ CreateElectricityGO successful (executed by eproducer1)"
echo ""

# Test 4: Query operations to verify writes
echo "Test 4: Verification Queries"
echo "----------------------------"

echo "4.1 - GetCurrentEGOsList:"
docker exec peer0.eissuer.go-platform.com peer chaincode query \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"issuance:GetCurrentEGOsList","Args":[]}'

echo ""
echo "4.2 - GetElectricityBacklog for DEV-ELEC-123:"
docker exec peer0.eproducer1.go-platform.com peer chaincode query \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"backlog:GetElectricityBacklog","Args":["DEV-ELEC-123"]}'

echo ""
echo "==============================================="
echo "✅ All write operation tests completed!"
echo "==============================================="
