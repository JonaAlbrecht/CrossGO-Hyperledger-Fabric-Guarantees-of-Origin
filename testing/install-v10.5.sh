#!/bin/bash
# Install chaincode v10.5 on all 6 peers
set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network

NETWORK_DIR=/root/hlf-go/repo/network
CC_PKG=/root/hlf-go/repo/golifecycle_10.5.tar.gz

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "Installing golifecycle_10.5 on all 6 peers..."
for org_entry in "eissuer:eissuerMSP:7051" "eproducer1:eproducer1MSP:9051" "ebuyer1:ebuyer1MSP:13051" "hissuer:hissuerMSP:8051" "hproducer1:hproducer1MSP:11051" "hbuyer1:hbuyer1MSP:14051"; do
  IFS=':' read -r org msp port <<< "$org_entry"
  set_peer_env "$org" "$msp" "$port"
  echo "  Installing on peer0.$org..."
  peer lifecycle chaincode install "$CC_PKG" 2>&1 || echo "    (already installed or error)"
done

echo "Installation complete!"
