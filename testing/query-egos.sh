#!/bin/bash
# Simple query script to check GO status
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=eproducer1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/users/Admin@eproducer1.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:9051

echo "Querying electricity GOs..."
peer chaincode query -C electricity-de -n golifecycle -c '{"function":"query:GetCurrentEGOsList","Args":[]}'
