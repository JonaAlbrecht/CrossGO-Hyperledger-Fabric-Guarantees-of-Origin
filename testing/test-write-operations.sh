#!/bin/bash
# Test write operations using peer CLI with role-based identities
# This script demonstrates chaincode functions that require specific org roles

set -e

CHANNEL_NAME="electricity-de"
CC_NAME="golifecycle"
ORDERER_CA="/root/hlf-go/repo/network/organizations/ordererOrganizations/go-platform.com/tlsca/tlsca.go-platform.com-cert.pem"
ORDERER_ADDR="orderer1.go-platform.com:7050"

echo "==============================================="
echo "Testing Write Operations with Role-Based Access"
echo "==============================================="
echo ""

# Test 1: PublishOracleData (requires issuer role)
echo "Test 1: PublishOracleData (issuer role)"
echo "---------------------------------------"
export CORE_PEER_LOCALMSPID="eissuerMSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/users/Admin@eissuer.go-platform.com/msp
export CORE_PEER_ADDRESS=peer0.eissuer.go-platform.com:7051

ORACLE_DATA=$(cat <<EOF
{
  "CarrierType": "electricity",
  "Zone": "DE-North",
  "PeriodStart": "2026-04-25T00:00:00Z",
  "PeriodEnd": "2026-04-25T01:00:00Z",
  "ProductionMethod": "wind_onshore",
  "EmissionFactor": 0.05,
  "PricePerMWh": 45.5
}
EOF
)

peer chaincode invoke \
  -o ${ORDERER_ADDR} \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile ${ORDERER_CA} \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"oracle:PublishOracleData","Args":[]}' \
  --transient "{\"oracleData\":\"$(echo -n "$ORACLE_DATA" | base64 -w0)\"}" \
  --peerAddresses peer0.eissuer.go-platform.com:7051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt

echo "✓ PublishOracleData successful"
echo ""

# Test 2: AddToBacklogElectricity (requires producer role)
echo "Test 2: AddToBacklogElectricity (producer role)"
echo "-----------------------------------------------"
export CORE_PEER_LOCALMSPID="eproducer1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/users/Admin@eproducer1.go-platform.com/msp
export CORE_PEER_ADDRESS=peer0.eproducer1.go-platform.com:8051

BACKLOG_DATA=$(cat <<EOF
{
  "AmountMWh": 150.5,
  "Emissions": 7.5,
  "ElectricityProductionMethod": "wind_onshore",
  "ElapsedSeconds": 900
}
EOF
)

peer chaincode invoke \
  -o ${ORDERER_ADDR} \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile ${ORDERER_CA} \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' \
  --transient "{\"eBacklog\":\"$(echo -n "$BACKLOG_DATA" | base64 -w0)\"}" \
  --peerAddresses peer0.eproducer1.go-platform.com:8051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt

echo "✓ AddToBacklogElectricity successful"
echo ""

# Test 3: CreateElectricityGO (requires producer role + SBE endorsement)
echo "Test 3: CreateElectricityGO (producer role, state-based endorsement)"
echo "--------------------------------------------------------------------"

EGO_DATA=$(cat <<EOF
{
  "AmountMWh": 200.0,
  "Emissions": 10.0,
  "ElectricityProductionMethod": "solar_pv",
  "ElapsedSeconds": 3600
}
EOF
)

peer chaincode invoke \
  -o ${ORDERER_ADDR} \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile ${ORDERER_CA} \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c '{"function":"issuance:CreateElectricityGO","Args":[]}' \
  --transient "{\"eGO\":\"$(echo -n "$EGO_DATA" | base64 -w0)\"}" \
  --peerAddresses peer0.eissuer.go-platform.com:7051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  --peerAddresses peer0.eproducer1.go-platform.com:8051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt \
  --peerAddresses peer0.ebuyer1.go-platform.com:9051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt

echo "✓ CreateElectricityGO successful"
echo ""

# Test 4: GetElectricityBacklog (read operation, any role)
echo "Test 4: GetElectricityBacklog (read)"
echo "------------------------------------"
DEVICE_ID="DEV-ELEC-123"

peer chaincode query \
  -C ${CHANNEL_NAME} \
  -n ${CC_NAME} \
  -c "{\"function\":\"backlog:GetElectricityBacklog\",\"Args\":[\"${DEVICE_ID}\"]}"

echo "✓ GetElectricityBacklog successful"
echo ""

echo "==============================================="
echo "All write operation tests completed successfully!"
echo "==============================================="
