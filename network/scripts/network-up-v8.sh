#!/bin/bash
# network-up-v8.sh — Bring up the GO Platform v8.0 multi-channel network (ADR-030)
#
# Creates TWO carrier-specific channels with selective org membership:
#   electricity-de: Issuer, EProducer, Buyer (3 orgs)
#   hydrogen-de:    Issuer, HProducer, Buyer (3 orgs)
#
# Shared Raft orderer cluster serves both channels.
# Issuer and Buyer span both channels; producers join only their carrier channel.
#
# Usage: ./network-up-v8.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NETWORK_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(dirname "$NETWORK_DIR")"

CHANNEL_ELEC="electricity-de"
CHANNEL_H2="hydrogen-de"

# Fabric binaries
export PATH="${REPO_DIR}/fabric-bin/bin:$PATH"
export FABRIC_CFG_PATH="$NETWORK_DIR"

# Org definitions: name mspID peerPort
ORGS=(
  "issuer1:issuer1MSP:7051"
  "eproducer1:eproducer1MSP:9051"
  "hproducer1:hproducer1MSP:11051"
  "buyer1:buyer1MSP:13051"
)

# Channel membership — selective per ADR-030
ORGS_ELEC=(
  "issuer1:issuer1MSP:7051"
  "eproducer1:eproducer1MSP:9051"
  "buyer1:buyer1MSP:13051"
)

ORGS_H2=(
  "issuer1:issuer1MSP:7051"
  "hproducer1:hproducer1MSP:11051"
  "buyer1:buyer1MSP:13051"
)

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${CYAN}[→]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
fail() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

set_peer_env() {
  local org=$1 msp=$2 port=$3
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID="$msp"
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/peers/peer0.${org}.go-platform.com/tls/ca.crt"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/${org}.go-platform.com/users/Admin@${org}.go-platform.com/msp"
  export CORE_PEER_ADDRESS="localhost:${port}"
}

join_orderers_to_channel() {
  local channel_name=$1
  local block_path=$2
  info "Joining orderers to channel: $channel_name"
  for i in 1 2 3 4; do
    ADMIN_PORT=$((9442 + i))
    ORDERER_TLS="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/orderers/orderer${i}.go-platform.com/tls"
    osnadmin channel join \
      --channelID "$channel_name" \
      --config-block "$block_path" \
      -o "localhost:${ADMIN_PORT}" \
      --ca-file "$ORDERER_TLS/ca.crt" \
      --client-cert "$ORDERER_TLS/server.crt" \
      --client-key "$ORDERER_TLS/server.key"
    log "  orderer${i} joined $channel_name"
  done
}

join_peers_to_channel() {
  local channel_name=$1
  local block_path=$2
  shift 2
  local org_list=("$@")
  info "Joining peers to channel: $channel_name"
  for entry in "${org_list[@]}"; do
    IFS=':' read -r org msp port <<< "$entry"
    set_peer_env "$org" "$msp" "$port"
    peer channel join -b "$block_path"
    log "  peer0.${org} joined $channel_name"
  done
}

set_anchor_peers() {
  local channel_name=$1
  shift
  local org_list=("$@")
  local ORDERER_CA="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/msp/tlscacerts/tlsca.orderer.go-platform.com-cert.pem"

  info "Setting anchor peers for channel: $channel_name"
  for entry in "${org_list[@]}"; do
    IFS=':' read -r org msp port <<< "$entry"
    set_peer_env "$org" "$msp" "$port"

    # Fetch current config
    peer channel fetch config "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_config_block.pb" \
      -o localhost:7050 --ordererTLSHostnameOverride localhost \
      -c "$channel_name" --tls --cafile "$ORDERER_CA" 2>/dev/null

    # Decode to JSON
    configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_config_block.pb" \
      --type common.Block 2>/dev/null \
      | jq '.data.data[0].payload.data.config' > "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_config.json"

    # Modify — add anchor peer
    jq --arg msp "$msp" --arg host "peer0.${org}.go-platform.com" --argjson port "$port" \
      '.channel_group.groups.Application.groups[$msp].values += {
        "AnchorPeers": {
          "mod_policy": "Admins",
          "value": {"anchor_peers": [{"host": $host, "port": $port}]},
          "version": "0"
        }
      }' "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_config.json" \
      > "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_modified_config.json"

    # Encode both
    configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_config.json" \
      --type common.Config --output "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_config.pb" 2>/dev/null
    configtxlator proto_encode --input "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_modified_config.json" \
      --type common.Config --output "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_modified_config.pb" 2>/dev/null

    # Compute update delta
    if configtxlator compute_update --channel_id "$channel_name" \
      --original "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_config.pb" \
      --updated "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_modified_config.pb" \
      --output "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_anchor_update.pb" 2>/dev/null; then

      # Wrap in envelope
      configtxlator proto_decode --input "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_anchor_update.pb" \
        --type common.ConfigUpdate 2>/dev/null \
        | jq '{"payload":{"header":{"channel_header":{"channel_id":"'"$channel_name"'","type":2}},"data":{"config_update":.}}}' \
        | configtxlator proto_encode --type common.Envelope \
          --output "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_anchor_update_tx.pb" 2>/dev/null

      # Submit
      peer channel update -f "$NETWORK_DIR/channel-artifacts/${channel_name}_${org}_anchor_update_tx.pb" \
        -c "$channel_name" -o localhost:7050 \
        --ordererTLSHostnameOverride localhost \
        --tls --cafile "$ORDERER_CA"
      log "  Anchor peer set for $msp on $channel_name"
    else
      log "  Anchor peer already set for $msp on $channel_name (no update needed)"
    fi
  done
}

echo "=============================================="
echo "  GO Platform v8.0 Multi-Channel Network"
echo "  ADR-030: Channel-per-carrier topology"
echo "=============================================="
echo ""

# ─── Step 0: Clean previous state ───────────────────────────────────────────
info "Cleaning previous state..."
"$SCRIPT_DIR/network-down.sh" 2>/dev/null || true
sudo rm -rf "$NETWORK_DIR/organizations" "$NETWORK_DIR/channel-artifacts"
log "Clean"

# ─── Step 1: Generate crypto material ────────────────────────────────────────
info "Generating crypto material with cryptogen..."
cryptogen generate --config="$NETWORK_DIR/crypto-config.yaml" --output="$NETWORK_DIR/organizations"
log "Crypto material generated"

# Fix orderer crypto paths (cryptogen generates orderer{i}.orderer.go-platform.com, Docker expects orderer{i}.go-platform.com)
info "Fixing orderer crypto paths (symlinks)..."
ORDERER_BASE="$NETWORK_DIR/organizations/ordererOrganizations/orderer.go-platform.com/orderers"
for i in 1 2 3 4; do
  ln -sfn "$ORDERER_BASE/orderer${i}.orderer.go-platform.com" "$ORDERER_BASE/orderer${i}.go-platform.com"
done
log "Orderer crypto paths fixed"

# ─── Step 2: Generate genesis blocks for BOTH channels ───────────────────────
info "Generating genesis blocks for both channels..."
mkdir -p "$NETWORK_DIR/channel-artifacts"

configtxgen -profile ElectricityDEChannel \
  -outputBlock "$NETWORK_DIR/channel-artifacts/${CHANNEL_ELEC}.block" \
  -channelID "$CHANNEL_ELEC"
log "Genesis block: channel-artifacts/${CHANNEL_ELEC}.block"

configtxgen -profile HydrogenDEChannel \
  -outputBlock "$NETWORK_DIR/channel-artifacts/${CHANNEL_H2}.block" \
  -channelID "$CHANNEL_H2"
log "Genesis block: channel-artifacts/${CHANNEL_H2}.block"

# ─── Step 3: Start orderer cluster ──────────────────────────────────────────
info "Starting shared orderer cluster (4-node Raft)..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-orderer.yaml" up -d
sleep 3
log "Orderers started (shared across both channels)"

# ─── Step 4: Start all peers ────────────────────────────────────────────────
info "Starting peers and CouchDB instances..."
docker compose -f "$NETWORK_DIR/docker/docker-compose-issuer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-eproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-hproducer.yaml" up -d
docker compose -f "$NETWORK_DIR/docker/docker-compose-buyer.yaml" up -d
sleep 5
log "All peers started"

# ─── Step 5: Join orderers to BOTH channels ─────────────────────────────────
join_orderers_to_channel "$CHANNEL_ELEC" "$NETWORK_DIR/channel-artifacts/${CHANNEL_ELEC}.block"
join_orderers_to_channel "$CHANNEL_H2" "$NETWORK_DIR/channel-artifacts/${CHANNEL_H2}.block"

# ─── Step 6: Selective peer joins ────────────────────────────────────────────
# electricity-de: Issuer + EProducer + Buyer (NOT HProducer)
join_peers_to_channel "$CHANNEL_ELEC" "$NETWORK_DIR/channel-artifacts/${CHANNEL_ELEC}.block" "${ORGS_ELEC[@]}"

# hydrogen-de: Issuer + HProducer + Buyer (NOT EProducer)
join_peers_to_channel "$CHANNEL_H2" "$NETWORK_DIR/channel-artifacts/${CHANNEL_H2}.block" "${ORGS_H2[@]}"

# ─── Step 7: Set anchor peers per channel ────────────────────────────────────
# Anchor peers are already defined in configtx.yaml org definitions and baked
# into the genesis blocks. Dynamic updates are optional — skip on policy error.
set_anchor_peers "$CHANNEL_ELEC" "${ORGS_ELEC[@]}" || warn "Anchor peer update skipped for $CHANNEL_ELEC (already defined in genesis)"
set_anchor_peers "$CHANNEL_H2" "${ORGS_H2[@]}" || warn "Anchor peer update skipped for $CHANNEL_H2 (already defined in genesis)"

# ─── Done ────────────────────────────────────────────────────────────────────
echo ""
echo "=============================================="
echo "  GO Platform v8.0 Multi-Channel — RUNNING"
echo "=============================================="
echo ""
echo "Channels:"
echo "  ${CHANNEL_ELEC}:  Issuer, EProducer, Buyer"
echo "  ${CHANNEL_H2}:    Issuer, HProducer, Buyer"
echo ""
echo "Peers:"
echo "  Issuer:            peer0.issuer1.go-platform.com:7051    (BOTH channels)"
echo "  E-Producer:        peer0.eproducer1.go-platform.com:9051 (electricity-de only)"
echo "  H-Producer:        peer0.hproducer1.go-platform.com:11051 (hydrogen-de only)"
echo "  Buyer:             peer0.buyer1.go-platform.com:13051    (BOTH channels)"
echo ""
echo "Orderers (shared):   7050, 8050, 9050, 10050"
echo "CouchDB:             5984, 7984, 9984, 11984"
echo ""
echo "Next: ./deploy-v8.sh"
