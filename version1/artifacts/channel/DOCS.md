# channel — Channel Configuration & Genesis Block Artifacts

This folder contains the Hyperledger Fabric channel configuration for `mychannel`.

## Files

### `configtx.yaml`
The core channel configuration defining:
- **5 Organizations**: Buyer (`buyerMSP`), E-Producer (`eproducerMSP`), Issuer (`issuerMSP`), H-Producer (`hproducerMSP`), Orderer (`OrdererMSP`)
- **Raft Consensus**: 4 orderer nodes with etcdraft
- **Batch Configuration**: 2s timeout, max 10 messages, 99MB absolute max bytes
- **Capabilities**: Application V2_5
- **Channel Profiles**:
  - `GOnetworkChannel` — Genesis block profile with all 4 application orgs
  - `ChannelGO` — Application channel profile

### `create-artifacts.sh`
Generates channel artifacts and bootstraps the ordering service:
1. Runs `configtxgen` to create the genesis block for `mychannel`
2. Enrolls orderer admin identities
3. Joins each orderer node to the channel via `osnadmin channel join`

### `mychannel28.block`
Pre-generated genesis block for the channel (iteration 28 of the config).

### `config/`
Fabric configuration templates:
- `core.yaml` — Peer configuration defaults
- `orderer.yaml` — Orderer configuration defaults

### Anchor Peer Directories
- `buyerAnchor/` — Anchor peer update TX for buyer org
- `eproducerAnchor/` — Anchor peer update TX for e-producer org
- `hproducerAnchor/` — Anchor peer update TX for h-producer org
- `issuerAnchor/` — Anchor peer update TX for issuer org
