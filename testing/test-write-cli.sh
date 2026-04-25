#!/bin/bash
set -e

export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

echo "========================================"
echo "Test 1: PublishOracleData (eissuer)"
echo "========================================"
export CORE_PEER_LOCALMSPID="eissuerMSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/users/Admin@eissuer.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

ORACLE_DATA='{"CarrierType":"electricity","Zone":"DE-LU","PeriodStart":1700000000,"PeriodEnd":1700003600,"ProductionMethod":"solar_pv","EnergyUnit":"MWh","Quantity":500,"EmissionFactor":0,"DataSource":"ENTSO-E-TP","Attributes":{}}'

/root/hlf-go/repo/fabric-bin/bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile "$ORDERER_CA" \
  -C electricity-de -n golifecycle \
  -c '{"function":"oracle:PublishOracleData","Args":[]}' \
  --transient "{\"OracleData\":\"$(echo -n "$ORACLE_DATA" | base64 -w0)\"}"

echo -e "\n========================================"
echo "Test 2: GetCurrentEGOsList (verify)"
echo "========================================"

/root/hlf-go/repo/fabric-bin/bin/peer chaincode query \
  -C electricity-de -n golifecycle \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}'

echo -e "\nAll tests completed successfully!"
