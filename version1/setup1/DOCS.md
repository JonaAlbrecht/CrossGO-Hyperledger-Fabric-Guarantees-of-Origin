# setup1 — Network Deployment & Organization VMs

This folder contains the deployment configuration for all 5 organizations and the network lifecycle scripts.

## Network Lifecycle Scripts

| Script | Purpose |
|--------|---------|
| `network-up.sh` | Full network bring-up: CAs → crypto material → orderers → channel creation → Docker images → peers → channel join → anchor peers → chaincode deploy |
| `network-upsplit1.sh` | Part 1 only: CA startup + crypto material generation |
| `network-upsplit2.sh` | Part 2 only: Orderers → channel → Docker images → peers → channel join → anchor peers → chaincode deploy |
| `network-down.sh` | Full teardown: removes channels, stops all containers, deletes crypto material, images, and artifacts |
| `channel-down.sh` | Placeholder (empty) |

## Organization Folders

Each organization folder follows a consistent structure:

```
org-vm/
├── .env                              # Fabric image version, CA version, compose project name
├── base.yaml                         # Base Docker service definitions (peer, CouchDB)
├── docker-compose.yaml               # Docker Compose for peer(s), CouchDB(s), CLI container
├── docker-compose-two-peers.yaml     # Alternative config with 2 peers per org
├── create-cryptomaterial-{org}/      # Fabric CA enrollment scripts + generated crypto
├── joinChannel.sh                    # Peer joins mychannel + sets anchor peer
├── installAndApproveChaincode.sh     # Chaincode lifecycle: package → install → approve
├── {org}-client/                     # Shell scripts to invoke chaincode functions
└── log.txt                           # Deployment logs
```

### `orderer-vm4/` — Ordering Service
- 4 Raft orderer nodes (`orderer.GOnetwork.com` ports 7050/7053, `orderer2` 8050/8053, `orderer3` 9050/9053, `orderer4` 10050/10053)
- Own Fabric CA (port 9054)
- Crypto material generated via `create-cryptomaterial-orderer/`

### `issuer-vm3/` — Issuing Body (Central Authority)
- Peer `i-peer0` (port 11051) with 2 CouchDB instances
- Fabric CA (port 10054) — **trust anchor** for the network
- **Registers all metering devices** (SmartMeter, OutputMeter) with production attributes
- Additional containers:
  - **SmartMeter** (`SmartMeter-config/`): Docker container simulating IoT electricity meter. Runs cron job every minute executing `execute.sh` — generates random solar production data (45–49 MWh) and invokes `CreateElectricityGO`
  - **OutputMeter** (`OutputMeter-config/`): Docker container simulating hydrogen output meter. Runs cron job every minute executing `execute-hproducer.sh` — generates random hydrogen production data (~150 kg/min) and invokes `AddHydrogentoBacklog`
- **`issuer-client/`**: Scripts for chaincode commit, querying all GO types, reading public/private data
- **`createChannel.sh`**: Fetches genesis block and joins the channel
- **`deployChaincode.sh`**: Full chaincode lifecycle on the issuer peer
- **`test.sh`**: Quick test script

### `eproducer-vm2/` — Electricity Producer
- Peer `e-peer0` (port 9051) with 2 CouchDB instances
- Fabric CA (port 8054)
- **`eproducer-client/`**: Scripts for:
  - `TransferEGO.sh` / `TransferEGObyAmount.sh` — Transfer eGOs
  - `ClaimRenewableattributesElectricity.sh` — Cancel eGOs for renewable claims
  - `ReadPubliceGO.sh` / `ReadPrivateElectricityGO.sh` — Read GO data
  - `QueryPrivateeGOsbyAmountMWh.sh` — Find eGOs by amount
  - `ReadCancelStatement*.sh` / `ReadConsumptionDeclaration*.sh` — Read statements

### `hproducer-vm5/` — Hydrogen Producer
- Peer `h-peer0` (port 13051) with 2 CouchDB instances
- Fabric CA (port 11054)
- **`hproducer-client/`**: Scripts for:
  - `IssueHGO.sh` — **Core conversion**: cancel eGOs → issue hydrogen GO
  - `TransferHGO.sh` / `TransferHGObyAmount.sh` — Transfer hGOs
  - `ClaimRenewableattributesHydrogen.sh` — Cancel hGOs for renewable claims
  - `QueryHydrogenBacklog.sh` — Check hydrogen production backlog
  - All read/query operations for both electricity and hydrogen GOs

### `buyer-vm1/` — Energy Buyer
- Peer `b-peer0` (port 7051) with 2 CouchDB instances
- Fabric CA (port 7054)
- **`buyer-client/`**: Contains both:
  - `bash/` — Shell scripts for all buyer operations (transfer, claim, read, query)
  - `nodejs/` — **Node.js application** (incomplete/experimental):
    - `app.js` — Express-like application for buyer interaction
    - `modules/gateway.js` — Fabric Gateway SDK connection module
    - `buyer-wallet/` — Identity wallet for the buyer
    - `ccpandcerts/` — Connection profiles and TLS certificates
