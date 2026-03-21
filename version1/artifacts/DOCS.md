# artifacts — Channel Configuration, Chaincode & Private Data

This folder contains the core blockchain artifacts for the GO network.

## Contents

### `channel/`
Channel and genesis block configuration:
- **`configtx.yaml`** — Channel configuration defining all 5 organizations (Buyer, E-Producer, Issuer, H-Producer, Orderer), their MSP paths and anchor peers. Configures:
  - Raft consensus with 4 orderer nodes
  - Batch settings: 2s timeout, max 10 messages, 99MB absolute max
  - Application capability V2_5
  - Channel profiles: `GOnetworkChannel` (consortium) and `ChannelGO` (channel)
- **`create-artifacts.sh`** — Generates the channel genesis block using `configtxgen` and enrolls orderer nodes via `osnadmin channel join`
- **`mychannel28.block`** — Pre-generated genesis block for `mychannel`
- **`config/`** — Fabric configuration templates (core.yaml, orderer.yaml)
- **`*Anchor/`** — Anchor peer update transaction artifacts per organization

### `Mychaincode/`
The Go smart contract (chaincode) — the **heart of the system**:
- **`main.go`** — Entry point; initializes `SmartContract` via Fabric's `contractapi`
- **`go.mod` / `go.sum`** — Go module dependencies
- **`GOnetwork/conversion.go`** — ~2,350 lines of core chaincode implementing all GO lifecycle operations. See [Mychaincode/DOCS.md](Mychaincode/DOCS.md) for detailed function reference.

### `private-data-collections/`
- **`collection-config.json`** — Defines 4 private data collections:

  | Collection | Access | Purpose |
  |-----------|--------|---------|
  | `publicGOcollection` | All 4 orgs | Shared public GO metadata |
  | `privateDetails-eproducerMSP` | eproducer + issuer | Electricity producer's private GO attributes |
  | `privateDetails-hproducerMSP` | hproducer + issuer | Hydrogen producer's private GO attributes |
  | `privateDetails-buyerMSP` | buyer + issuer | Buyer's private GO attributes |

  The **Issuer** always has read access to all collections for auditing purposes.
