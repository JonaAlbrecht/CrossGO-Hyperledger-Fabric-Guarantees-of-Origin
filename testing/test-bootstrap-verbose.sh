#!/bin/bash
# Verbose bootstrap test
set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

set_peer_env() {
  local org=$1
  local port=$2
  export CORE_PEER_LOCALMSPID="${org}MSP"
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp
  export CORE_PEER_ADDRESS=localhost:${port}
}

echo "========================================="
echo "STEP 1: Bootstrap eissuer via Init Ledger"
echo "========================================="
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}'

echo ""
echo "Waiting 5 seconds for state to commit..."
sleep 5

echo ""
echo "========================================="
echo "STEP 2: Query eissuer role to verify write"
echo "========================================="
set_peer_env eissuer 7051
peer chaincode query -C electricity-de -n golifecycle \
  -c '{"function":"query:GetState","Args":["orgRole_eissuerMSP"]}' || echo "Query failed!"

echo ""
echo "========================================="
echo "STEP 3: Try to register producer role"
echo "========================================="
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}'

echo ""
echo "========================================="
echo "TEST COMPLETE"
echo "========================================="
