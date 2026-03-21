# hproducer-vm5 — Hydrogen Producer Organization

This folder configures the **Hydrogen Producer (H-Producer)** — a PEM electrolyser facility that converts electricity into green hydrogen.

## Role
The H-Producer represents a hydrogen production facility that:
- Receives hydrogen production data via the **OutputMeter** (accumulates into a backlog)
- Receives transferred **Electricity GOs (eGOs)** from the E-Producer
- **Issues Hydrogen GOs (hGOs)** by converting eGOs — the core cross-carrier conversion operation
- Transfers hGOs to the Buyer
- Can claim renewable hydrogen attributes by cancelling hGOs

## Components
- **`h-peer0`** (port 13051) — H-Producer's endorsing peer
- **2 CouchDB instances** — State database
- **Fabric CA** (port 11054) — Issues identities for h-producer users
- **`hproducer-client`** container — CLI for chaincode operations

## Files
| File | Purpose |
|------|---------|
| `.env` | Image versions and compose project name |
| `base.yaml` | Base Docker service definitions (peer, CouchDB) |
| `docker-compose.yaml` | Docker Compose for peer, CouchDBs, CLI |
| `joinChannel.sh` | Joins `h-peer0` to `mychannel` and updates anchor peer |
| `installAndApproveChaincode.sh` | Chaincode lifecycle: package → install → approve |

## Client Scripts (`hproducer-client/`)
| Script | Purpose |
|--------|---------|
| **`IssueHGO.sh`** | **Core conversion**: Cancels consumed eGOs, creates consumption declarations, and issues a new hydrogen GO with inherited attributes |
| `TransferHGO.sh` | Transfer a single hGO to another org |
| `TransferHGObyAmount.sh` | Transfer hGOs by kilos amount (with splitting) |
| `ClaimRenewableattributesHydrogen.sh` | Cancel hGOs to claim renewable hydrogen attributes |
| `QueryHydrogenBacklog.sh` | Check hydrogen production backlog |
| `QueryPrivateEGOsbyAmountMWh.sh` / `QueryPrivatehGOsbyAmountMWh.sh` | Find GOs by amount |
| `ReadPublicEGO.sh` / `ReadPublicHGO.sh` | Read public GO metadata |
| `ReadPrivateElectricityGO.sh` / `ReadPrivateHydrogenGO.sh` | Read private GO details |
| `ReadCancelStatement*.sh` / `ReadConsumptionDeclaration*.sh` | Read all statement types |
