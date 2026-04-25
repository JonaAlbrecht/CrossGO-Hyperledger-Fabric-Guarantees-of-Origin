#!/bin/bash
# Initialize ledger for v10 topology (electricity-de + hydrogen-de channels)
set -euo pipefail
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
export CORE_PEER_TLS_ENABLED=true
ND=/root/hlf-go/repo/network
ORDERER_CA=$ND/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

# TLS cert paths for all orgs
EISSUER_TLS=$ND/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
HISSUER_TLS=$ND/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt
EPROD_TLS=$ND/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
HPROD_TLS=$ND/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt
EBUYER_TLS=$ND/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt
HBUYER_TLS=$ND/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt

echo "============================================"
echo "Initializing electricity-de channel"
echo "============================================"

# Set peer context to eissuer for electricity-de
export CORE_PEER_LOCALMSPID=eissuerMSP
export CORE_PEER_TLS_ROOTCERT_FILE=$EISSUER_TLS
export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/eissuer.go-platform.com/users/Admin@eissuer.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

echo '=== device:InitLedger (electricity-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C electricity-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER_TLS" \
  -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}' --waitForEvent
echo "InitLedger (elec) exit: $?"
sleep 2

echo '=== device:RegisterOrgRole eproducer1MSP => producer (electricity-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C electricity-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER_TLS" \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}' --waitForEvent
sleep 1

echo '=== device:RegisterOrgRole ebuyer1MSP => buyer (electricity-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C electricity-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER_TLS" \
  -c '{"function":"device:RegisterOrgRole","Args":["ebuyer1MSP","buyer"]}' --waitForEvent
sleep 1

echo '=== device:ListDevices (electricity-de) ==='
peer chaincode query -C electricity-de -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'

echo ""
echo "============================================"
echo "Initializing hydrogen-de channel"
echo "============================================"

# Set peer context to hissuer for hydrogen-de
export CORE_PEER_LOCALMSPID=hissuerMSP
export CORE_PEER_TLS_ROOTCERT_FILE=$HISSUER_TLS
export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/hissuer.go-platform.com/users/Admin@hissuer.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:8051

echo '=== device:InitLedger (hydrogen-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C hydrogen-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER_TLS" \
  -c '{"function":"device:InitLedger","Args":["hissuerMSP"]}' --waitForEvent
echo "InitLedger (h2) exit: $?"
sleep 2

echo '=== device:RegisterOrgRole hproducer1MSP => producer (hydrogen-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C hydrogen-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER_TLS" \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}' --waitForEvent
sleep 1

echo '=== device:RegisterOrgRole hbuyer1MSP => buyer (hydrogen-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C hydrogen-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER_TLS" \
  -c '{"function":"device:RegisterOrgRole","Args":["hbuyer1MSP","buyer"]}' --waitForEvent
sleep 1

echo '=== device:ListDevices (hydrogen-de) ==='
peer chaincode query -C hydrogen-de -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'

echo ""
echo "============================================"
echo "Initialization complete"
echo "============================================"
echo "Electricity channel devices:"
peer chaincode query -C electricity-de -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'
echo ""
echo "Hydrogen channel devices:"
export CORE_PEER_LOCALMSPID=hissuerMSP
export CORE_PEER_ADDRESS=localhost:8051
peer chaincode query -C hydrogen-de -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'

echo ""
echo "=== Done ===" 