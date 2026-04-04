#!/bin/bash
# boot-network.sh — Full boot sequence with orderer symlink fix
set -euo pipefail

REPO_DIR=/root/hlf-go/repo
NETWORK_DIR=$REPO_DIR/network
CHANNEL_NAME=goplatformchannel

export PATH=$REPO_DIR/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$NETWORK_DIR

ORGS=(
  "issuer1:issuer1MSP:7051"
  "eproducer1:eproducer1MSP:9051"
  "hproducer1:hproducer1MSP:11051"
  "buyer1:buyer1MSP:13051"
)

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "=============================================="
echo "  GO Platform Network — Starting"
echo "=============================================="

# Step 0: Clean
echo "[->] Cleaning previous state..."
bash "$NETWORK_DIR/scripts/network-down.sh" 2>/dev/null || true
rm -rf "$NETWORK_DIR/organizations" "$NETWORK_DIR/channel-artifacts"
echo "[OK] Clean"

# Step 1: Generate crypto material
echo "[->] Generating crypto material..."
cryptogen generate --config="$NETWORK_DIR/crypto-config.yaml" --output="$NETWORK_DIR/organizations"
echo "[OK] Crypto material generated"

# Step 2: Fix orderer paths - cryptogen creates ordererX.orderer.go-platform.com, configs expect ordererX.go-platform.com
echo "[->] Creating orderer directory copies..."
ORDERER_DIR="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/orderers"
for i in 1 2 3 4; do
  cp -a "${ORDERER_DIR}/orderer${i}.orderer.go-platform.com" "${ORDERER_DIR}/orderer${i}.go-platform.com"
done
echo "[OK] Orderer directory copies created"

# Step 3: Generate genesis block
echo "[->] Generating channel genesis block..."
mkdir -p "$NETWORK_DIR/channel-artifacts"
configtxgen -profile GOPlatformChannel \
  -outputBlock "$NETWORK_DIR/channel-artifacts/${CHANNEL_NAME}.block" \
  -channelID "$CHANNEL_NAME"
echo "[OK] Genesis block created"

# Step 4: Start orderer cluster
echo "[->] Starting orderer cluster (4-node Raft)..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-orderer.yaml" up -d
sleep 3
echo "[OK] Orderers started"

# Step 5: Start all peers
echo "[->] Starting peers and CouchDB instances..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-issuer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-eproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-hproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-buyer.yaml" up -d
sleep 5
echo "[OK] All peers started"

# Step 6: Join orderers to channel
echo "[->] Joining orderers to channel..."
for i in 1 2 3 4; do
  ADMIN_PORT=$((9442 + i))
  ORDERER_TLS="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/orderers/orderer${i}.go-platform.com/tls"
  osnadmin channel join \
    --channelID "$CHANNEL_NAME" \
    --config-block "$NETWORK_DIR/channel-artifacts/${CHANNEL_NAME}.block" \
    -o "localhost:${ADMIN_PORT}" \
    --ca-file "$ORDERER_TLS/ca.crt" \
    --client-cert "$ORDERER_TLS/server.crt" \
    --client-key "$ORDERER_TLS/server.key"
  echo "  orderer${i} joined channel"
done

# Step 7: Join peers to channel
echo "[->] Joining peers to channel..."
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  peer channel join -b "$NETWORK_DIR/channel-artifacts/${CHANNEL_NAME}.block"
  echo "  peer0.${org} joined channel"
done

# Step 8: Set anchor peers
echo "[->] Setting anchor peers..."
ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"
for entry in "${ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  peer channel fetch config "$NETWORK_DIR/channel-artifacts/config_block.pb" \
    -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -c "$CHANNEL_NAME" --tls --cafile "$ORDERER_CA" 2>/dev/null
  configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/config_block.pb" \
    --type common.Block 2>/dev/null \
    | jq '.data.data[0].payload.data.config' > "$NETWORK_DIR/channel-artifacts/${org}_config.json"
  jq --arg msp "$msp" --arg host "peer0.${org}.go-platform.com" --argjson port "$port" \
    '.channel_group.groups.Application.groups[$msp].values += {
      "AnchorPeers": {
        "mod_policy": "Admins",
        "value": {"anchor_peers": [{"host": $host, "port": $port}]},
        "version": "0"
      }
    }' "$NETWORK_DIR/channel-artifacts/${org}_config.json" \
    > "$NETWORK_DIR/channel-artifacts/${org}_modified_config.json"
  configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${org}_config.json" \
    --type common.Config --output "$NETWORK_DIR/channel-artifacts/${org}_config.pb" 2>/dev/null
  configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${org}_modified_config.json" \
    --type common.Config --output "$NETWORK_DIR/channel-artifacts/${org}_modified_config.pb" 2>/dev/null
  if configtxlator compute_update --channel_id "$CHANNEL_NAME" \
    --original "$NETWORK_DIR/channel-artifacts/${org}_config.pb" \
    --updated "$NETWORK_DIR/channel-artifacts/${org}_modified_config.pb" \
    --output "$NETWORK_DIR/channel-artifacts/${org}_anchor_update.pb" 2>/dev/null; then
    configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/${org}_anchor_update.pb" \
      --type common.ConfigUpdate 2>/dev/null \
      | jq '{"payload":{"header":{"channel_header":{"channel_id":"'"$CHANNEL_NAME"'","type":2}},"data":{"config_update":.}}}' \
      | configtxlator proto_encode --type common.Envelope \
        --output "$NETWORK_DIR/channel-artifacts/${org}_anchor_update_tx.pb" 2>/dev/null
    peer channel update -f "$NETWORK_DIR/channel-artifacts/${org}_anchor_update_tx.pb" \
      -c "$CHANNEL_NAME" -o localhost:7050 \
      --ordererTLSHostnameOverride localhost \
      --tls --cafile "$ORDERER_CA"
    echo "  Anchor peer set for $msp"
  else
    echo "  Anchor peer already set for $msp"
  fi
done

echo ""
echo "=============================================="
echo "  GO Platform Network — RUNNING"
echo "=============================================="
echo ""
echo "Channel: $CHANNEL_NAME"
echo "Peers: issuer1:7051, eproducer1:9051, hproducer1:11051, buyer1:13051"
echo "Orderers: 7050, 8050, 9050, 10050"
echo ""
echo "Next: deploy chaincode with ./network/scripts/deploy-chaincode.sh"
