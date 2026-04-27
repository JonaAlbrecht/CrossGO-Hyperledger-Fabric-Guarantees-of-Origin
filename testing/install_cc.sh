#!/bin/bash
set -euo pipefail
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
cd /root/hlf-go/repo

ORGS=("issuer1:issuer1MSP:7051" "eproducer1:eproducer1MSP:9051" "hproducer1:hproducer1MSP:11051" "buyer1:buyer1MSP:13051")

for entry in "${ORGS[@]}"; do
  IFS=":" read -r org msp port <<< "$entry"
  echo "=== Installing on $org ==="
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
  peer lifecycle chaincode install golifecycle.tar.gz
done

echo ""
echo "=== Querying installed chaincode ==="
export CORE_PEER_LOCALMSPID="issuer1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE="/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt"
export CORE_PEER_MSPCONFIGPATH="/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/users/Admin@issuer1.go-platform.com/msp"
export CORE_PEER_ADDRESS="localhost:7051"
peer lifecycle chaincode queryinstalled
