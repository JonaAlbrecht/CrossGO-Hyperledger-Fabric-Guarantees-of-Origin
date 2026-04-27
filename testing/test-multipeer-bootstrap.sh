#!/bin/bash
# Bootstrap with multi-peer endorsement
set -e

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem
EISSUER_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
EPRODUCER1_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
EBUYER1_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt

set_peer_env() {
  local org=$1
  local port=$2
  export CORE_PEER_LOCALMSPID="${org}MSP"
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp
  export CORE_PEER_ADDRESS=localhost:${port}
}

echo "===== ELECTRICITY CHANNEL BOOTSTRAP ====="
echo ""
echo "[E1] Bootstrapping eissuer as issuer (multi-peer endorsement)..."
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS \
  -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}'

sleep 3

echo "[E2] Registering eproducer1 as producer..."
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}'

sleep 3

echo "[E3] Registering ebuyer1 as buyer..."
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS \
  -c '{"function":"device:RegisterOrgRole","Args":["ebuyer1MSP","buyer"]}'

echo ""
echo "===== BOOTSTRAP COMPLETE ====="
echo "All roles registered on electricity-de"
