#!/bin/bash
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=issuer1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/users/Admin@issuer1.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/orderers/orderer1.go-platform.com/tls/ca.crt

echo "=== Query ListDevices ==="
peer chaincode query -C goplatformchannel -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'

echo "=== Invoke RegisterOrgRole for eproducer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost --waitForEvent

echo "=== Invoke RegisterOrgRole for hproducer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost --waitForEvent

echo "=== Invoke RegisterOrgRole for buyer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["buyer1MSP","consumer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost --waitForEvent

echo "=== Verify: RegisterDevice test ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterDevice","Args":[]}' \
  --transient '{"Device":"{\"deviceType\":\"SmartMeter\",\"ownerOrgMSP\":\"eproducer1MSP\",\"energyCarriers\":[\"electricity\"],\"attributes\":{\"maxEfficiency\":\"100\",\"emissionIntensity\":\"50\",\"technologyType\":\"solar\"}}"}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost --waitForEvent

echo "=== Done ==="
