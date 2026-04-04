#!/bin/bash
export PATH=/root/hlf-go/repo/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=/root/hlf-go/repo/network
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=issuer1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/users/Admin@issuer1.go-platform.com/msp
export CORE_PEER_ADDRESS=localhost:7051

ORDERER_CA=/root/hlf-go/repo/network/organizations/ordererOrganizations/orderer.go-platform.com/orderers/orderer1.go-platform.com/tls/ca.crt

echo "=== Step 1: InitLedger with waitForEvent ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:InitLedger","Args":["issuer1MSP"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
echo "InitLedger exit code: $?"

echo "=== Step 2: Check CouchDB for orgRole ==="
curl -s http://admin:adminpw@localhost:5984/goplatformchannel_golifecycle/_all_docs

echo ""
echo "=== Step 3: RegisterOrgRole for eproducer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["eproducer1MSP","producer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
echo "RegisterOrgRole eproducer1MSP exit code: $?"

echo "=== Step 4: RegisterOrgRole for hproducer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["hproducer1MSP","producer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
echo "RegisterOrgRole hproducer1MSP exit code: $?"

echo "=== Step 5: RegisterOrgRole for buyer1MSP ==="
peer chaincode invoke -C goplatformchannel -n golifecycle \
  -c '{"function":"device:RegisterOrgRole","Args":["buyer1MSP","consumer"]}' \
  --tls --cafile "$ORDERER_CA" \
  -o localhost:7050 --ordererTLSHostnameOverride localhost \
  --waitForEvent --waitForEventTimeout 30s \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles /root/hlf-go/repo/network/organizations/peerOrganizations/issuer1.go-platform.com/peers/peer0.issuer1.go-platform.com/tls/ca.crt
echo "RegisterOrgRole buyer1MSP exit code: $?"

echo "=== Step 6: Final CouchDB check ==="
curl -s http://admin:adminpw@localhost:5984/goplatformchannel_golifecycle/_all_docs

echo ""
echo "=== Done ==="
