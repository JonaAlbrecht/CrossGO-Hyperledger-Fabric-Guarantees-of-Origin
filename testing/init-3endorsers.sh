#!/bin/bash
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=issuer1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/users/Admin@issuer1.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/orderers/orderer1.go-platform.com/tls/ca.crt
ISSUER_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
EPROD_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/eproducer1.go-platform.com/peers/peer0.eproducer1.go-platform.com/tls/ca.crt
HPROD_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/hproducer1.go-platform.com/peers/peer0.hproducer1.go-platform.com/tls/ca.crt
BUYER_TLS=/root/hlf-go/repo/network/organizations/peerOrganizations/buyer1.go-platform.com/peers/peer0.buyer1.go-platform.com/tls/ca.crt

# Need 3 of 4 endorsing peers (majority endorsement policy)
ENDORSE_PEERS="--peerAddresses localhost:7051 --tlsRootCertFiles $ISSUER_TLS --peerAddresses localhost:9051 --tlsRootCertFiles $EPROD_TLS --peerAddresses localhost:11051 --tlsRootCertFiles $HPROD_TLS"

echo "=== Step 1: InitLedger (3 endorsers) ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:InitLedger","Args":["issuer1MSP"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  $ENDORSE_PEERS
echo "InitLedger exit code: $?"

echo "=== Step 2: RegisterOrgRole eproducer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  $ENDORSE_PEERS
echo "exit code: $?"

echo "=== Step 3: RegisterOrgRole hproducer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  $ENDORSE_PEERS
echo "exit code: $?"

echo "=== Step 4: RegisterOrgRole buyer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["buyer1MSP","consumer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  $ENDORSE_PEERS
echo "exit code: $?"

echo "=== Step 5: Verify CouchDB ==="
curl -s http://admin:adminpw@localhost:5984/goplatformchannel_golifecycle/_all_docs

echo ""
echo "=== Step 6: Test RegisterDevice ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterDevice","Args":[]}' \
  --transient "{\"Device\":\"$(echo '{"deviceType":"SmartMeter","ownerOrgMSP":"eproducer1MSP","energyCarriers":["electricity"],"attributes":{"maxEfficiency":"100","emissionIntensity":"50","technologyType":"solar"}}' | base64 -w0)\"}" \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  $ENDORSE_PEERS
echo "RegisterDevice exit code: $?"

echo "=== Step 7: List Devices ==="
peer chaincode query -C goplatformchannel -n golifecycle \
  -c '{"function":"device:ListDevices","Args":[]}'

echo ""
echo "=== Done ==="
