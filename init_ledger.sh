#!/bin/bash
set -euo pipefail

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
NETWORK_DIR=/root/hlf-go/repo/network

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=issuer1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=$NETWORK_DIR/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=$NETWORK_DIR/organizations/peerOrganizations/issuer1.go-platform.com/users/Admin@issuer1.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

ORDERER_CA=$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

PEER_CONN="--peerAddresses localhost:7051 --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt --peerAddresses localhost:11051 --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt --peerAddresses localhost:13051 --tlsRootCertFiles $NETWORK_DIR/organizations/peerOrganizations/buyer1.go-platform.com/peers/peer0.buyer1.go-platform.com/tls/ca.crt"

echo "=== InitLedger ==="
peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN \
  -c '{"function":"device:InitLedger","Args":["issuer1MSP"]}'

sleep 2

echo "=== RegisterOrgRole eproducer1MSP ==="
peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}'

sleep 1

echo "=== RegisterOrgRole hproducer1MSP ==="
peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}'

sleep 1

echo "=== RegisterOrgRole buyer1MSP ==="
peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle \
  --tls --cafile "$ORDERER_CA" \
  $PEER_CONN \
  -c '{"function":"device:RegisterOrgRole","Args":["buyer1MSP","consumer"]}'

echo ""
echo "=== Ledger Initialized ==="
