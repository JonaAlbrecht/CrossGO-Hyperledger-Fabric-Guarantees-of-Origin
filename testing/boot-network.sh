#!/bin/bash
# boot-network.sh — v10.1: Two-channel (electricity-de, hydrogen-de) 6-org network
set -euo pipefail

REPO_DIR=/root/hlf-go/repo
NETWORK_DIR=$REPO_DIR/network
export PATH=$REPO_DIR/fabric-bin/bin:$PATH
export FABRIC_CFG_PATH=$NETWORK_DIR

E_CHANNEL=electricity-de
H_CHANNEL=hydrogen-de
ORDERER_CA=$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem

# Org: name  mspID          port   channel
E_ORGS=("eissuer:eissuerMSP:7051" "eproducer1:eproducer1MSP:9051" "ebuyer1:ebuyer1MSP:13051")
H_ORGS=("hissuer:hissuerMSP:8051" "hproducer1:hproducer1MSP:11051" "hbuyer1:hbuyer1MSP:14051")
ALL_ORGS=("${E_ORGS[@]}" "${H_ORGS[@]}")

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

echo "====== GO Platform v10.1 — Starting Two-Channel Network ======"

# Step 0: Clean
echo "[1/9] Cleaning previous state..."
bash "$NETWORK_DIR/scripts/network-down.sh" 2>/dev/null || true
rm -rf "$NETWORK_DIR/organizations" "$NETWORK_DIR/channel-artifacts"

# Step 1: Generate crypto
echo "[2/9] Generating crypto material (6 orgs)..."
cryptogen generate --config="$NETWORK_DIR/crypto-config.yaml" --output="$NETWORK_DIR/organizations"

# Step 2: Fix orderer paths
echo "[3/9] Fixing orderer paths..."
ORDERER_DIR="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/orderers"
for i in 1 2 3 4; do
  cp -a "${ORDERER_DIR}/orderer${i}.orderer.go-platform.com" "${ORDERER_DIR}/orderer${i}.go-platform.com"
done

# Step 3: Generate genesis blocks for both channels
echo "[4/9] Generating channel genesis blocks..."
mkdir -p "$NETWORK_DIR/channel-artifacts"
configtxgen -profile ElectricityDEChannel \
  -outputBlock "$NETWORK_DIR/channel-artifacts/${E_CHANNEL}.block" \
  -channelID "$E_CHANNEL"
configtxgen -profile HydrogenDEChannel \
  -outputBlock "$NETWORK_DIR/channel-artifacts/${H_CHANNEL}.block" \
  -channelID "$H_CHANNEL"

# Step 4: Start orderers
echo "[5/9] Starting orderers..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-orderer.yaml" up -d
sleep 4

# Step 5: Start all peers
echo "[6/9] Starting all peers..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-eissuer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-hissuer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-eproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-hproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-ebuyer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-hbuyer.yaml" up -d
sleep 6

# Step 6: Join orderers to both channels
echo "[7/9] Joining orderers to channels..."
for i in 1 2 3 4; do
  ADMIN_PORT=$((9442 + i))
  ORDERER_TLS="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/orderers/orderer${i}.go-platform.com/tls"
  for CH in "$E_CHANNEL" "$H_CHANNEL"; do
    osnadmin channel join \
      --channelID "$CH" \
      --config-block "$NETWORK_DIR/channel-artifacts/${CH}.block" \
      -o "localhost:${ADMIN_PORT}" \
      --ca-file "$ORDERER_TLS/ca.crt" \
      --client-cert "$ORDERER_TLS/server.crt" \
      --client-key "$ORDERER_TLS/server.key"
  done
  echo "  orderer${i} joined both channels"
done

# Step 7: Join peers to their respective channels
echo "[8/9] Joining peers to channels..."
for entry in "${E_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  peer channel join -b "$NETWORK_DIR/channel-artifacts/${E_CHANNEL}.block"
  echo "  peer0.${org} joined ${E_CHANNEL}"
done
for entry in "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_peer_env "$org" "$msp" "$port"
  peer channel join -b "$NETWORK_DIR/channel-artifacts/${H_CHANNEL}.block"
  echo "  peer0.${org} joined ${H_CHANNEL}"
done

# Step 8: Set anchor peers
echo "[9/9] Setting anchor peers..."
set_anchor() {
  local org=$1 msp=$2 port=$3 channel=$4
  set_peer_env "$org" "$msp" "$port"
  peer channel fetch config "$NETWORK_DIR/channel-artifacts/${org}_config.pb" \
    -o localhost:7050 --ordererTLSHostnameOverride localhost \
    -c "$channel" --tls --cafile "$ORDERER_CA" 2>/dev/null
  configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/${org}_config.pb" \
    --type common.Block 2>/dev/null \
    | jq ".data.data[0].payload.data.config" > "$NETWORK_DIR/channel-artifacts/${org}_config.json"
  jq --arg msp "$msp" --arg host "peer0.${org}.go-platform.com" --argjson port "$port" \
    ".channel_group.groups.Application.groups[\$msp].values += {
      \"AnchorPeers\": {
        \"mod_policy\": \"Admins\",
        \"value\": {\"anchor_peers\": [{\"host\": \$host, \"port\": \$port}]},
        \"version\": \"0\"
      }
    }" "$NETWORK_DIR/channel-artifacts/${org}_config.json" \
    > "$NETWORK_DIR/channel-artifacts/${org}_modified_config.json"
  configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${org}_config.json" \
    --type common.Config --output "$NETWORK_DIR/channel-artifacts/${org}_config.pb" 2>/dev/null
  configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${org}_modified_config.json" \
    --type common.Config --output "$NETWORK_DIR/channel-artifacts/${org}_modified_config.pb" 2>/dev/null
  if configtxlator compute_update --channel_id "$channel" \
    --original "$NETWORK_DIR/channel-artifacts/${org}_config.pb" \
    --updated "$NETWORK_DIR/channel-artifacts/${org}_modified_config.pb" \
    --output "$NETWORK_DIR/channel-artifacts/${org}_anchor_update.pb" 2>/dev/null; then
    configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/${org}_anchor_update.pb" \
      --type common.ConfigUpdate 2>/dev/null \
      | jq "{\"payload\":{\"header\":{\"channel_header\":{\"channel_id\":\"${channel}\",\"type\":2}},\"data\":{\"config_update\":.}}}" \
      | configtxlator proto_encode --type common.Envelope \
        --output "$NETWORK_DIR/channel-artifacts/${org}_anchor_update_tx.pb" 2>/dev/null
    peer channel update -f "$NETWORK_DIR/channel-artifacts/${org}_anchor_update_tx.pb" \
      -c "$channel" -o localhost:7050 --ordererTLSHostnameOverride localhost \
      --tls --cafile "$ORDERER_CA"
    echo "  Anchor peer set for $msp on $channel"
  fi
}

for entry in "${E_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_anchor "$org" "$msp" "$port" "$E_CHANNEL"
done
for entry in "${H_ORGS[@]}"; do
  IFS=':' read -r org msp port <<< "$entry"
  set_anchor "$org" "$msp" "$port" "$H_CHANNEL"
done

echo ""
echo "====== Network UP ======"
echo "electricity-de: eissuer:7051, eproducer1:9051, ebuyer1:13051"
echo "hydrogen-de:    hissuer:8051, hproducer1:11051, hbuyer1:14051"
echo "Orderers: 7050, 8050, 9050, 10050"
echo ""
echo "Next: bash network/scripts/deploy-chaincode.sh"
