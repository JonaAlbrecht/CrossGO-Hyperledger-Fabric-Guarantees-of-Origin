#!/bin/bash
# Re-initialize ledger after sequence 2 upgrade

set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem
ORDERER_ENDPOINT=localhost:7050

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
echo "INITIALIZING LEDGER (POST-UPGRADE)"
echo "========================================="
echo ""

# ELECTRICITY-DE CHANNEL
echo ">>> Initializing ledger on electricity-de..."
set_peer_env eissuer 7051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}'

sleep 2

echo ">>> Registering eissuerMSP as issuer on electricity-de..."
set_peer_env eissuer 7051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:RegisterOrgRole","Args":["eissuerMSP","issuer"]}'

sleep 2

echo ">>> Registering eproducer1MSP as producer on electricity-de..."
set_peer_env eissuer 7051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}'

sleep 2

echo ">>> Registering ebuyer1MSP as buyer on electricity-de..."
set_peer_env eissuer 7051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:RegisterOrgRole","Args":["ebuyer1MSP","buyer"]}'

sleep 2

echo ""
echo ">>> Initializing ledger on hydrogen-de..."
set_peer_env hissuer 8051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:InitLedger","Args":["hissuerMSP"]}'

sleep 2

echo ">>> Registering hissuerMSP as issuer on hydrogen-de..."
set_peer_env hissuer 8051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:RegisterOrgRole","Args":["hissuerMSP","issuer"]}'

sleep 2

echo ">>> Registering hproducer1MSP as producer on hydrogen-de..."
set_peer_env hissuer 8051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}'

sleep 2

echo ">>> Registering hbuyer1MSP as buyer on hydrogen-de..."
set_peer_env hissuer 8051
peer chaincode invoke -o $ORDERER_ENDPOINT --ordererTLSHostnameOverride orderer1.go-platform.com \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt \
  -c '{"function":"device:RegisterOrgRole","Args":["hbuyer1MSP","buyer"]}'

echo ""
echo "========================================="
echo "LEDGER INITIALIZATION COMPLETE"
echo "========================================="
