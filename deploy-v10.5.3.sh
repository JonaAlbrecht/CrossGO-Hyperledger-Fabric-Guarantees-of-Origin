#!/bin/bash
set -euo pipefail
REPO=/root/hlf-go/repo
ND=$REPO/network
export PATH=$REPO/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$ND

CC=golifecycle
VER=10.5.3
SEQ=1
E_CH=electricity-de
H_CH=hydrogen-de
ORDERER_CA=$ND/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem
E_COLL=$REPO/collections/collection-config-electricity-de.json
H_COLL=$REPO/collections/collection-config-hydrogen-de.json

set_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID=$msp
  export CORE_PEER_TLS_ROOTCERT_FILE=$ND/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=$ND/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp
  export CORE_PEER_ADDRESS=localhost:${port}
}

E_ORGS=("eissuer:eissuerMSP:7051" "eproducer1:eproducer1MSP:9051" "ebuyer1:ebuyer1MSP:13051")
H_ORGS=("hissuer:hissuerMSP:8051" "hproducer1:hproducer1MSP:11051" "hbuyer1:hbuyer1MSP:14051")

EISSUER_TLS=$ND/organizations/peerOrganizations/eissuer.go-platform.com/peers/peer0.eissuer.go-platform.com/tls/ca.crt
EPRODUCER1_TLS=$ND/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
EBUYER1_TLS=$ND/organizations/peerOrganizations/ebuyer1.go-platform.com/peers/peer0.ebuyer1.go-platform.com/tls/ca.crt
HISSUER_TLS=$ND/organizations/peerOrganizations/hissuer.go-platform.com/peers/peer0.hissuer.go-platform.com/tls/ca.crt
HPRODUCER1_TLS=$ND/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt
HBUYER1_TLS=$ND/organizations/peerOrganizations/hbuyer1.go-platform.com/peers/peer0.hbuyer1.go-platform.com/tls/ca.crt

E_PEERS="--peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS --peerAddresses localhost:13051 --tlsRootCertFiles $EBUYER1_TLS"
H_PEERS="--peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS --peerAddresses localhost:14051 --tlsRootCertFiles $HBUYER1_TLS"

invoke_e() {
  set_env eissuer eissuerMSP 7051
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -C $E_CH -n $CC --tls --cafile $ORDERER_CA $E_PEERS "$@" 2>&1
  sleep 2
}
invoke_h() {
  set_env hissuer hissuerMSP 8051
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -C $H_CH -n $CC --tls --cafile $ORDERER_CA $H_PEERS "$@" 2>&1
  sleep 2
}
query_e() { set_env eissuer eissuerMSP 7051; peer chaincode query -C $E_CH -n $CC "$@" 2>&1; }
query_h() { set_env hissuer hissuerMSP 8051; peer chaincode query -C $H_CH -n $CC "$@" 2>&1; }

echo "=== [1/5] Installing golifecycle_${VER}.tar.gz on all 6 peers ==="
for entry in "${E_ORGS[@]}" "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_env $org $msp $port
  peer lifecycle chaincode install $REPO/golifecycle_${VER}.tar.gz 2>&1 | grep -E 'Installed|already|Error' || true
  echo "  $msp: installed"
done

echo "=== [2/5] Approving on both channels (seq=$SEQ) ==="
set_env eissuer eissuerMSP 7051
PKG=$(peer lifecycle chaincode queryinstalled --output json | python3 -c "
import sys,json
ccs=json.load(sys.stdin)['installed_chaincodes']
label='${CC}_${VER}'
match=[c['package_id'] for c in ccs if label in c['package_id']]
print(match[0])
")
echo "  Package ID: $PKG"

for entry in "${E_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_env $org $msp $port
  peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --channelID $E_CH --name $CC --version $VER --package-id "$PKG" --sequence $SEQ \
    --collections-config $E_COLL --tls --cafile $ORDERER_CA 2>&1 | tail -1
  echo "  $msp approved for $E_CH"
done

for entry in "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_env $org $msp $port
  peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride localhost \
    --channelID $H_CH --name $CC --version $VER --package-id "$PKG" --sequence $SEQ \
    --collections-config $H_COLL --tls --cafile $ORDERER_CA 2>&1 | tail -1
  echo "  $msp approved for $H_CH"
done

echo "=== [3/5] Committing on both channels ==="
set_env eissuer eissuerMSP 7051
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --channelID $E_CH --name $CC --version $VER --sequence $SEQ \
  --collections-config $E_COLL --tls --cafile $ORDERER_CA $E_PEERS 2>&1 | tail -2
echo "  Committed to $E_CH"

set_env hissuer hissuerMSP 8051
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --channelID $H_CH --name $CC --version $VER --sequence $SEQ \
  --collections-config $H_COLL --tls --cafile $ORDERER_CA $H_PEERS 2>&1 | tail -2
echo "  Committed to $H_CH"

echo "  Waiting 8s for chaincode containers to start..."
sleep 8

echo "=== [4/5] Initializing ledger ==="
invoke_e -c '{"function":"device:InitLedger","Args":["eissuerMSP"]}' | tail -1
echo "[E1] eissuerMSP bootstrapped as issuer"
invoke_e -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}' | tail -1
echo "[E2] eproducer1MSP = producer"
invoke_e -c '{"function":"device:RegisterOrgRole","Args":["ebuyer1MSP","buyer"]}' | tail -1
echo "[E3] ebuyer1MSP = buyer"

EDEV=$(python3 -c "
import json,base64
d={'DeviceType':'SmartMeter','OwnerOrgMSP':'eproducer1MSP','EnergyCarrier':'electricity','Attributes':{'location':'Berlin','installedCapacityKW':'500'}}
print(base64.b64encode(json.dumps(d).encode()).decode())
")
invoke_e -c '{"function":"device:RegisterDevice","Args":[]}' --transient "{\"Device\":\"$EDEV\"}" | tail -1
echo "[E4] electricity SmartMeter registered"

EBACKLOG=$(python3 -c "
import json,base64
d={'AmountMWh':500,'Emissions':25.0,'ElectricityProductionMethod':'solar_pv','ElapsedSeconds':3600}
print(base64.b64encode(json.dumps(d).encode()).decode())
")
set_env eproducer1 eproducer1MSP 9051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost -C $E_CH -n $CC --tls --cafile $ORDERER_CA \
  --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
  --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
  -c '{"function":"backlog:AddToBacklogElectricity","Args":[]}' --transient "{\"eBacklog\":\"$EBACKLOG\"}" 2>&1 | tail -1
sleep 2
echo "[E5] electricity backlog added (500 MWh)"

invoke_h -c '{"function":"device:InitLedger","Args":["hissuerMSP"]}' | tail -1
echo "[H1] hissuerMSP bootstrapped as issuer"
invoke_h -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}' | tail -1
echo "[H2] hproducer1MSP = producer"
invoke_h -c '{"function":"device:RegisterOrgRole","Args":["hbuyer1MSP","buyer"]}' | tail -1
echo "[H3] hbuyer1MSP = buyer"

HDEV=$(python3 -c "
import json,base64
d={'DeviceType':'OutputMeter','OwnerOrgMSP':'hproducer1MSP','EnergyCarrier':'hydrogen','Attributes':{'location':'Hamburg','capacity':'50t/year'}}
print(base64.b64encode(json.dumps(d).encode()).decode())
")
invoke_h -c '{"function":"device:RegisterDevice","Args":[]}' --transient "{\"Device\":\"$HDEV\"}" | tail -1
echo "[H4] hydrogen OutputMeter registered"

HBACKLOG=$(python3 -c "
import json,base64
d={'Kilosproduced':1000,'EmissionsHydrogen':5.0,'UsedMWh':50.0,'HydrogenProductionMethod':'electrolysis','ElapsedSeconds':3600}
print(base64.b64encode(json.dumps(d).encode()).decode())
")
set_env hproducer1 hproducer1MSP 11051
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost -C $H_CH -n $CC --tls --cafile $ORDERER_CA \
  --peerAddresses localhost:8051 --tlsRootCertFiles $HISSUER_TLS \
  --peerAddresses localhost:11051 --tlsRootCertFiles $HPRODUCER1_TLS \
  -c '{"function":"backlog:AddToBacklogHydrogen","Args":[]}' --transient "{\"hBacklog\":\"$HBACKLOG\"}" 2>&1 | tail -1
sleep 2
echo "[H5] hydrogen backlog added (1000 kg / 50 MWh input)"

echo "=== [5/5] Creating 3 electricity GOs ==="
for i in 1 2 3; do
  EGO=$(python3 -c "
import json,base64
d={'AmountMWh':100,'Emissions':5.0,'ElectricityProductionMethod':'solar_pv','ElapsedSeconds':3600}
print(base64.b64encode(json.dumps(d).encode()).decode())
")
  set_env eproducer1 eproducer1MSP 9051
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride localhost -C $E_CH -n $CC --tls --cafile $ORDERER_CA \
    --peerAddresses localhost:7051 --tlsRootCertFiles $EISSUER_TLS \
    --peerAddresses localhost:9051 --tlsRootCertFiles $EPRODUCER1_TLS \
    -c '{"function":"issuance:CreateElectricityGO","Args":[]}' --transient "{\"eGO\":\"$EGO\"}" 2>&1 | tail -1
  sleep 2
  echo "  GO #$i created"
done

echo ""
echo "=== Ledger Status ==="
query_e -c '{"function":"query:GetCurrentEGOsList","Args":[]}' | python3 -c "
import sys,json
gos=json.load(sys.stdin)
print(f'Electricity GOs: {len(gos)}')
for g in gos: print(f'  {g[\"AssetID\"]}  status={g[\"Status\"]}')
"
echo ""
echo "=== Deploy + Init COMPLETE. Run test-conversion-v10.5.3.sh next. ==="
