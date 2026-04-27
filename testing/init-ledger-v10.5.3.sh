#!/bin/bash
set -euo pipefail
REPO=/root/hlf-go/repo
ND=$REPO/network
export PATH=$REPO/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$ND

CC=golifecycle
E_CH=electricity-de
H_CH=hydrogen-de
ORDERER_CA=$ND/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

EISSUER_TLS=$ND/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
EPRODUCER1_TLS=$ND/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
EBUYER1_TLS=$ND/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt
HISSUER_TLS=$ND/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt
HPRODUCER1_TLS=$ND/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt
HBUYER1_TLS=$ND/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt

E_PEERS="--peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS"
H_PEERS="--peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS --peerAddresses localhost:14051 --tlsRootCertFiles $HBUYER1_TLS"

set_env() {
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID=$2
  export CORE_PEER_TLS_ROOTCERT_FILE=$ND/organizations/peerOrganizations/${1}.go-platform.com/peers/peer0.${1}.go-platform.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/${1}.go-platform.com/users/Admin@${1}.go-platform.com/msp
  export CORE_PEER_ADDRESS=localhost:${3}
}

invoke_e() {
  set_env eissuer eissuerMSP 7051
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -C $E_CH -n $CC --tls --cafile $ORDERER_CA $E_PEERS "$@" 2>&1 | grep -E 'successful|Error' || true
  sleep 2
}
invoke_h() {
  set_env hissuer hissuerMSP 8051
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -C $H_CH -n $CC --tls --cafile $ORDERER_CA $H_PEERS "$@" 2>&1 | grep -E 'successful|Error' || true
  sleep 2
}

b64() { echo -n "$1" | base64 -w0; }

echo "=== Initializing electricity-de ledger ==="
invoke_e -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}'
echo "[E1] eissuerMSP = issuer"
invoke_e -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}'
echo "[E2] eproducer1MSP = producer"
invoke_e -c '{"function":"device:RegisterOrgRole","Args":["ebuyer1MSP","buyer"]}'
echo "[E3] ebuyer1MSP = buyer"

EDEV=$(b64 '{"deviceType":"SmartMeter","ownerOrgMSP":"eproducer1MSP","energyCarriers":["electricity"],"attributes":{"location":"Berlin","installedCapacityKW":"500"}}')
invoke_e -c '{"function":"device:RegisterDevice","Args":[]}' --transient "{\"Device\":\"$EDEV\"}"
echo "[E4] electricity SmartMeter registered"

EBACKLOG=$(b64 '{"AmountMWh":500,"Emissions":25.0,"ElectricityProductionMethod":"solar_pv","ElapsedSeconds":3600}')
set_env eproducer1 eproducer1MSP 9051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost -C $E_CH -n $CC --tls --cafile $ORDERER_CA \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' --transient "{\"eBacklog\":\"$EBACKLOG\"}" 2>&1 | grep -E 'successful|Error' || true
sleep 2
echo "[E5] electricity backlog: 500 MWh"

echo ""
echo "=== Initializing hydrogen-de ledger ==="
invoke_h -c '{"function":"device:InitLedger","Args":["hissuerMSP"]}'
echo "[H1] hissuerMSP = issuer"
invoke_h -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}'
echo "[H2] hproducer1MSP = producer"
invoke_h -c '{"function":"device:RegisterOrgRole","Args":["hbuyer1MSP","buyer"]}'
echo "[H3] hbuyer1MSP = buyer"

HDEV=$(b64 '{"deviceType":"OutputMeter","ownerOrgMSP":"hproducer1MSP","energyCarriers":["hydrogen"],"attributes":{"location":"Hamburg","capacity":"50t/year"}}')
invoke_h -c '{"function":"device:RegisterDevice","Args":[]}' --transient "{\"Device\":\"$HDEV\"}"
echo "[H4] hydrogen OutputMeter registered"

HBACKLOG=$(b64 '{"Kilosproduced":1000,"EmissionsHydrogen":5.0,"UsedMWh":50.0,"HydrogenProductionMethod":"electrolysis","ElapsedSeconds":3600}')
set_env hproducer1 hproducer1MSP 11051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost -C $H_CH -n $CC --tls --cafile $ORDERER_CA \
  --peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS \
  --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS \
  -c '{"function":"backlog:AddToBacklogHydrogen","Args":[]}' --transient "{\"hBacklog\":\"$HBACKLOG\"}" 2>&1 | grep -E 'successful|Error' || true
sleep 2
echo "[H5] hydrogen backlog: 1000 kg (50 MWh input)"

echo ""
echo "=== Creating 3 electricity GOs ==="
for i in 1 2 3; do
  EGO=$(b64 '{"AmountMWh":100,"Emissions":5.0,"ElectricityProductionMethod":"solar_pv","ElapsedSeconds":3600}')
  set_env eproducer1 eproducer1MSP 9051
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost -C $E_CH -n $CC --tls --cafile $ORDERER_CA \
    --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
    --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
    -c '{"function":"issuance:CreateElectricityGO","Args":[]}' --transient "{\"eGO\":\"$EGO\"}" 2>&1 | grep -E 'successful|Error' || true
  sleep 2
  echo "  GO #$i created"
done

echo ""
echo "=== Verifying ledger state ==="
set_env eissuer eissuerMSP 7051
peer chaincode query -C $E_CH -n $CC -c '{"function":"query:GetCurrentEGOsList","Args":[]}' 2>/dev/null | python3 -c "
import sys,json
gos=json.load(sys.stdin)
print(f'Electricity GOs on ledger: {len(gos)}')
for g in gos: print(f'  {g[\"AssetID\"]}  status={g[\"Status\"]}')
"
echo ""
echo "=== Ledger initialization COMPLETE ==="
