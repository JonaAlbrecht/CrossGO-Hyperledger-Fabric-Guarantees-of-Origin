#!/bin/bash
# Simple wrapper to query GOs without escaping issues

REPO_DIR="/root/hlf-go/repo"
export PATH="$REPO_DIR/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$REPO_DIR/network"

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="eproducer1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
export CORE_PEER_MSPCONFIGPATH="$REPO_DIR/network/organizations/peerOrganizations/eproducer1.go-platform.com/users/Admin@eproducer1.go-platform.com/msp"
export CORE_PEER_ADDRESS="localhost:9051"

peer chaincode query \
  -C electricity-de -n golifecycle \
  -c '{"function":"query:GetCurrentEGOsList","Args":[]}' \
  --peerAddresses localhost:9051 \
  --tlsRootCertFiles "$CORE_PEER_TLS_ROOTCERT_FILE" \
  --tls
