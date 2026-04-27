#!/bin/bash
# Full ledger initialization with multi-peer endorsement
set -euo pipefail

export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/fabric-bin/config

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

# TLS certs for electricity channel
EISSUER_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
EPRODUCER1_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
EBUYER1_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt

# TLS certs for hydrogen channel
HISSUER_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt
HPRODUCER1_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt
HBUYER1_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt

set_peer_env() {
  local org=$1
  local port=$2
  export CORE_PEER_LOCALMSPID="${org}MSP"
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp
  export CORE_PEER_ADDRESS=localhost:${port}
}

echo "====================================="
echo "ELECTRICITY CHANNEL INITIALIZATION"
echo "====================================="
echo ""

echo "[E1] Bootstrapping eissuer as issuer..."
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS \
  -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}' 2>&1 | grep -v '^Note:'
sleep 2

echo "[E2] Registering eproducer1 as producer..."
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}' 2>&1 | grep -v '^Note:'
sleep 2

echo "[E3] Registering ebuyer1 as buyer..."
set_peer_env eissuer 7051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS \
  -c '{"function":"device:RegisterOrgRole","Args":["ebuyer1MSP","buyer"]}' 2>&1 | grep -v '^Note:'
sleep 2

echo "[E4] Registering electricity smart meter device..."
set_peer_env eissuer 7051
DEV=$(echo -n '{"deviceType":"SmartMeter","ownerOrgMSP":"eproducer1MSP","energyCarriers":["electricity"],"attributes":{"vendor":"SiemensMeter","model":"SM-7000","serialNumber":"SM-001","calibrationDate":"2025-01-01"}}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  -c '{"function":"device:RegisterDevice","Args":[]}' \
  --transient "{\"Device\":\"$DEV\"}" 2>&1 | grep -v '^Note:'
sleep 2

echo "[E5] Adding electricity backlog..."
set_peer_env eproducer1 9051
EBACKLOG=$(echo -n '{"AmountMWh":500,"Emissions":25.5,"ElectricityProductionMethod":"solar_pv","ElapsedSeconds":3600}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C electricity-de -n golifecycle \
  -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' \
  --transient "{\"eBacklog\":\"$EBACKLOG\"}" 2>&1 | grep -v '^Note:'
sleep 2

echo ""
echo "====================================="
echo "HYDROGEN CHANNEL INITIALIZATION"
echo "====================================="
echo ""

echo "[H1] Bootstrapping hissuer as issuer..."
set_peer_env hissuer 8051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS \
  --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS \
  --peerAddresses localhost:14051 --tlsRootCertFiles $HBUYER1_TLS \
  -c '{"function":"device:InitLedger","Args":["hissuerMSP"]}' 2>&1 | grep -v '^Note:'
sleep 2

echo "[H2] Registering hproducer1 as producer..."
set_peer_env hissuer 8051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS \
  --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}' 2>&1 | grep -v '^Note:'
sleep 2

echo "[H3] Registering hbuyer1 as buyer..."
set_peer_env hissuer 8051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS \
  --peerAddresses localhost:14051 --tlsRootCertFiles $HBUYER1_TLS \
  -c '{"function":"device:RegisterOrgRole","Args":["hbuyer1MSP","buyer"]}' 2>&1 | grep -v '^Note:'
sleep 2

echo "[H4] Registering hydrogen production meter device..."
set_peer_env hissuer 8051
DEV=$(echo -n '{"deviceType":"OutputMeter","ownerOrgMSP":"hproducer1MSP","energyCarriers":["hydrogen"],"attributes":{"vendor":"NEL Hydrogen","model":"HM-2000","serialNumber":"HM-001","calibrationDate":"2025-01-01","electrolyzerType":"PEM","nominalCapacityKgPerDay":"200"}}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  --peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS \
  --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS \
  -c '{"function":"device:RegisterDevice","Args":[]}' \
  --transient "{\"Device\":\"$DEV\"}" 2>&1 | grep -v '^Note:'
sleep 2

echo "[H5] Adding hydrogen backlog..."
set_peer_env hproducer1 11051
HBACKLOG=$(echo -n '{"Kilosproduced":1000,"EmissionsHydrogen":5.0,"UsedMWh":55.0,"HydrogenProductionMethod":"pem_electrolysis","ElapsedSeconds":3600}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile $ORDERER_CA -C hydrogen-de -n golifecycle \
  -c '{"function":"backlog:AddToBacklogHydrogen","Args":[]}' \
  --transient "{\"hBacklog\":\"$HBACKLOG\"}" 2>&1 | grep -v '^Note:'
sleep 2

echo ""
echo "====================================="
echo "LEDGER INITIALIZATION COMPLETE"
echo "====================================="
echo "Both channels ready with:"
echo "  - Organizations registered with roles"
echo "  - Smart meter devices registered"
echo "  - Initial backlog created"
echo ""
echo "Next: Create GOs with create-gos-base64.sh"
