#!/bin/bash
# init_ledger_fixed.sh — v10.1: Initialize both channels with bootstrap issuer first
set -euo pipefail

REPO_DIR=/root/hlf-go/repo
NETWORK_DIR=$REPO_DIR/network
export PATH=$REPO_DIR/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$NETWORK_DIR

E_CHANNEL="electricity-de"
H_CHANNEL="hydrogen-de"
CC="golifecycle"
ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

invoke() {
  local org=$1 msp=$2 port=$3 channel=$4; shift 4
  set_peer_env "$org" "$msp" "$port"
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --tls --cafile "$ORDERER_CA" -C "$channel" -n "$CC" "$@" 2>&1 | grep -v '^Note:' || true
  sleep 1
}

echo "====== Initializing Ledger — electricity-de ======"
echo ""

# CRITICAL: Bootstrap issuer FIRST before any other operations
echo "[E0] Bootstrapping eissuer as initial issuer (device:InitLedger)..."
invoke eissuer eissuerMSP 7051 "$E_CHANNEL" -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}'

echo "[E1] Registering organizations on electricity-de (admin:RegisterOrganization)..."

REG=$(echo -n '{"DisplayName":"Electricity Issuer","OrgMSP":"eissuerMSP","OrgType":"issuer","EnergyCarriers":["electricity"],"Country":"DE"}' | base64 -w0)
invoke eissuer eissuerMSP 7051 "$E_CHANNEL" -c '{"function":"admin:RegisterOrganization","Args":[]}' --transient "{\"OrgRegistration\":\"$REG\"}"

REG=$(echo -n '{"DisplayName":"Electricity Producer 1","OrgMSP":"eproducer1MSP","OrgType":"producer","EnergyCarriers":["electricity"],"Country":"DE"}' | base64 -w0)
invoke eissuer eissuerMSP 7051 "$E_CHANNEL" -c '{"function":"admin:RegisterOrganization","Args":[]}' --transient "{\"OrgRegistration\":\"$REG\"}"

REG=$(echo -n '{"DisplayName":"Electricity Buyer 1","OrgMSP":"ebuyer1MSP","OrgType":"buyer","EnergyCarriers":["electricity"],"Country":"DE"}' | base64 -w0)
invoke eissuer eissuerMSP 7051 "$E_CHANNEL" -c '{"function":"admin:RegisterOrganization","Args":[]}' --transient "{\"OrgRegistration\":\"$REG\"}"

echo "[E2] Registering roles on electricity-de (admin:RegisterOrgRole)..."
invoke eissuer eissuerMSP 7051 "$E_CHANNEL" -c '{"function":"admin:RegisterOrgRole","Args":["eissuerMSP","issuer"]}'
invoke eissuer eissuerMSP 7051 "$E_CHANNEL" -c '{"function":"admin:RegisterOrgRole","Args":["eproducer1MSP","producer"]}'
invoke eissuer eissuerMSP 7051 "$E_CHANNEL" -c '{"function":"admin:RegisterOrgRole","Args":["ebuyer1MSP","buyer"]}'

echo "[E3] Registering electricity smart meter device..."
EPRODUCER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt"
EISSUER_TLS="$NETWORK_DIR/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt"
set_peer_env eproducer1 eproducer1MSP 9051
DEV=$(echo -n '{"deviceType":"SmartMeter","ownerOrgMSP":"eproducer1MSP","energyCarriers":["electricity"],"attributes":{"vendor":"SiemensMeter","model":"SM-7000","serialNumber":"SM-001","calibrationDate":"2025-01-01"}}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile "$ORDERER_CA" -C "$E_CHANNEL" -n "$CC" \
  --peerAddresses localhost:9051 --tlsRootCertFiles "$EPRODUCER1_TLS" \
  --peerAddresses localhost:7051 --tlsRootCertFiles "$EISSUER_TLS" \
  -c '{"function":"device:RegisterDevice","Args":[]}' \
  --transient "{\"Device\":\"$DEV\"}" 2>&1 | grep -v '^Note:' || true
sleep 2

echo "[E4] Adding electricity backlog..."
set_peer_env eproducer1 eproducer1MSP 9051
EBACKLOG=$(echo -n '{"AmountMWh":500,"Emissions":25.5,"ElectricityProductionMethod":"solar_pv","ElapsedSeconds":3600}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile "$ORDERER_CA" -C "$E_CHANNEL" -n "$CC" \
  -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' \
  --transient "{\"eBacklog\":\"$EBACKLOG\"}" 2>&1 | grep -v '^Note:' || true
sleep 2

echo ""
echo "====== Initializing Ledger — hydrogen-de ======"
echo ""

# CRITICAL: Bootstrap issuer FIRST
echo "[H0] Bootstrapping hissuer as initial issuer (device:InitLedger)..."
invoke hissuer hissuerMSP 8051 "$H_CHANNEL" -c '{"function":"device:InitLedger","Args":["hissuerMSP"]}'

echo "[H1] Registering organizations on hydrogen-de (admin:RegisterOrganization)..."
REG=$(echo -n '{"DisplayName":"Hydrogen Issuer","OrgMSP":"hissuerMSP","OrgType":"issuer","EnergyCarriers":["hydrogen"],"Country":"DE"}' | base64 -w0)
invoke hissuer hissuerMSP 8051 "$H_CHANNEL" -c '{"function":"admin:RegisterOrganization","Args":[]}' --transient "{\"OrgRegistration\":\"$REG\"}"

REG=$(echo -n '{"DisplayName":"Hydrogen Producer 1","OrgMSP":"hproducer1MSP","OrgType":"producer","EnergyCarriers":["hydrogen"],"Country":"DE"}' | base64 -w0)
invoke hissuer hissuerMSP 8051 "$H_CHANNEL" -c '{"function":"admin:RegisterOrganization","Args":[]}' --transient "{\"OrgRegistration\":\"$REG\"}"

REG=$(echo -n '{"DisplayName":"Hydrogen Buyer 1","OrgMSP":"hbuyer1MSP","OrgType":"buyer","EnergyCarriers":["hydrogen"],"Country":"DE"}' | base64 -w0)
invoke hissuer hissuerMSP 8051 "$H_CHANNEL" -c '{"function":"admin:RegisterOrganization","Args":[]}' --transient "{\"OrgRegistration\":\"$REG\"}"

echo "[H2] Registering roles on hydrogen-de (admin:RegisterOrgRole)..."
invoke hissuer hissuerMSP 8051 "$H_CHANNEL" -c '{"function":"admin:RegisterOrgRole","Args":["hissuerMSP","issuer"]}'
invoke hissuer hissuerMSP 8051 "$H_CHANNEL" -c '{"function":"admin:RegisterOrgRole","Args":["hproducer1MSP","producer"]}'
invoke hissuer hissuerMSP 8051 "$H_CHANNEL" -c '{"function":"admin:RegisterOrgRole","Args":["hbuyer1MSP","buyer"]}'

echo "[H3] Registering hydrogen production meter device..."
HPRODUCER1_TLS="$NETWORK_DIR/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt"
HISSUER_TLS="$NETWORK_DIR/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt"
set_peer_env hproducer1 hproducer1MSP 11051
DEV=$(echo -n '{"deviceType":"OutputMeter","ownerOrgMSP":"hproducer1MSP","energyCarriers":["hydrogen"],"attributes":{"vendor":"NEL Hydrogen","model":"HM-2000","serialNumber":"HM-001","calibrationDate":"2025-01-01","electrolyzerType":"PEM","nominalCapacityKgPerDay":"200"}}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile "$ORDERER_CA" -C "$H_CHANNEL" -n "$CC" \
  --peerAddresses localhost:11051 --tlsRootCertFiles "$HPRODUCER1_TLS" \
  --peerAddresses localhost:8051 --tlsRootCertFiles "$HISSUER_TLS" \
  -c '{"function":"device:RegisterDevice","Args":[]}' \
  --transient "{\"Device\":\"$DEV\"}" 2>&1 | grep -v '^Note:' || true
sleep 2

echo "[H4] Adding hydrogen backlog..."
set_peer_env hproducer1 hproducer1MSP 11051
HBACKLOG=$(echo -n '{"Kilosproduced":1000,"EmissionsHydrogen":5.0,"UsedMWh":55.0,"HydrogenProductionMethod":"pem_electrolysis","ElapsedSeconds":3600}' | base64 -w0)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --tls --cafile "$ORDERER_CA" -C "$H_CHANNEL" -n "$CC" \
  -c '{"function":"backlog:AddToBacklogHydrogen","Args":[]}' \
  --transient "{\"hBacklog\":\"$HBACKLOG\"}" 2>&1 | grep -v '^Note:' || true
sleep 2

echo ""
echo "====== Ledger Initialized ======"
echo "Both channels ready for testing!"
