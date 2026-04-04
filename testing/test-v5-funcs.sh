#!/bin/bash
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
export CORE_PEER_TLS_ENABLED=true
ND=/root/hlf-go/repo/network
export CORE_PEER_LOCALMSPID=issuer1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=$ND/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/issuer1.go-platform.com/users/Admin@issuer1.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

echo '=== admin:GetVersion ==='
peer chaincode query -C goplatformchannel -n golifecycle \
  -c '{"function":"admin:GetVersion","Args":[]}'

echo '=== device:ListDevices ==='
peer chaincode query -C goplatformchannel -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'

echo '=== query:GetCurrentEGOsList ==='
peer chaincode query -C goplatformchannel -n golifecycle \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}'

echo '=== Done ==='
