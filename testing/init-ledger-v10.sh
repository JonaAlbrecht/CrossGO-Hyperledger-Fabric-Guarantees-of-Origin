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

echo '=== admin:InitializeRoles (electricity-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C electricity-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER_TLS" \
  -c '{"function":"admin:InitializeRoles","Args":[]}' --waitForEvent
echo "InitializeRoles (elec) exit: $?"
sleep 2

echo '=== admin:RegisterOrganization eissuerMSP (electricity-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C electricity-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER_TLS" \
  -c '{"function":"admin:RegisterOrganization","Args":["eissuerMSP","issuer","eissuer.go-platform.com"]}' --waitForEvent
sleep 1

echo '=== admin:RegisterOrganization eproducer1MSP (electricity-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C electricity-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER_TLS" \
  -c '{"function":"admin:RegisterOrganization","Args":["eproducer1MSP","producer","eproducer1.go-platform.com"]}' --waitForEvent
sleep 1

echo '=== admin:RegisterOrganization ebuyer1MSP (electricity-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C electricity-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPROD_TLS" \
  --peerAddresses localhost:13051 --tlsRootCertFiles "$EBUYER_TLS" \
  -c '{"function":"admin:RegisterOrganization","Args":["ebuyer1MSP","consumer","ebuyer1.go-platform.com"]}' --waitForEvent
sleep 1

echo '=== admin:ListOrganizations (electricity-de) ==='
peer chaincode query -C electricity-de -n golifecycle \
  -c '{"function":"admin:ListOrganizations","Args":[]}'

echo ""
echo "============================================"
echo "Initializing hydrogen-de channel"
echo "============================================"

# Set peer context to hissuer for hydrogen-de
export CORE_PEER_LOCALMSPID=hissuerMSP
export CORE_PEER_TLS_ROOTCERT_FILE=$HISSUER_TLS
export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/hissuer.go-platform.com/users/Admin@hissuer.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:8051

echo '=== admin:InitializeRoles (hydrogen-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C hydrogen-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER_TLS" \
  -c '{"function":"admin:InitializeRoles","Args":[]}' --waitForEvent
echo "InitializeRoles (h2) exit: $?"
sleep 2

echo '=== admin:RegisterOrganization hissuerMSP (hydrogen-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C hydrogen-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER_TLS" \
  -c '{"function":"admin:RegisterOrganization","Args":["hissuerMSP","issuer","hissuer.go-platform.com"]}' --waitForEvent
sleep 1

echo '=== admin:RegisterOrganization hproducer1MSP (hydrogen-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C hydrogen-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER_TLS" \
  -c '{"function":"admin:RegisterOrganization","Args":["hproducer1MSP","producer","hproducer1.go-platform.com"]}' --waitForEvent
sleep 1

echo '=== admin:RegisterOrganization hbuyer1MSP (hydrogen-de) ==='
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  -C hydrogen-de -n golifecycle --tls --cafile "$ORDERER_CA" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPROD_TLS" \
  --peerAddresses localhost:14051 --tlsRootCertFiles "$HBUYER_TLS" \
  -c '{"function":"admin:RegisterOrganization","Args":["hbuyer1MSP","consumer","hbuyer1.go-platform.com"]}' --waitForEvent
sleep 1

echo '=== admin:ListOrganizations (hydrogen-de) ==='
peer chaincode query -C hydrogen-de -n golifecycle \
  -c '{"function":"admin:ListOrganizations","Args":[]}'

echo ""
echo "============================================"
echo "Initialization complete"
echo "============================================"
echo "Electricity channel organizations:"
peer chaincode query -C electricity-de -n golifecycle \
  -c '{"function":"admin:ListOrganizations","Args":[]}'
echo ""
echo "Hydrogen channel organizations:"
export CORE_PEER_LOCALMSPID=hissuerMSP
export CORE_PEER_ADDRESS=localhost:8051
peer chaincode query -C hydrogen-de -n golifecycle \
  -c '{"function":"admin:ListOrganizations","Args":[]}'

echo ""
echo "=== Done ===" 