#!/bin/bash
set -euo pipefail
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
export CORE_PEER_TLS_ENABLED=true
ND=/root/hlf-go/repo/network
ORDERER_CA=$ND/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem
export CORE_PEER_LOCALMSPID=issuer1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=$ND/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/issuer1.go-platform.com/users/Admin@issuer1.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

ISSUER_TLS=$ND/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
EPROD_TLS=$ND/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
HPROD_TLS=$ND/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt

echo '=== InitLedger ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$ISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  -c '{"function":"device:InitLedger","Args":["issuer1MSP"]}' --waitForEvent
echo "InitLedger exit: $?"

sleep 2

echo '=== RegisterOrgRole eproducer1MSP => producer ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$ISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}' --waitForEvent
sleep 1

echo '=== RegisterOrgRole hproducer1MSP => producer ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$ISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}' --waitForEvent
sleep 1

echo '=== RegisterOrgRole buyer1MSP => consumer ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C goplatformchannel -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$ISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  -c '{"function":"device:RegisterOrgRole","Args":["buyer1MSP","consumer"]}' --waitForEvent
sleep 1

echo '=== Verify: ListDevices ==='
peer chaincode query -C goplatformchannel -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'

echo '=== Done ==='
