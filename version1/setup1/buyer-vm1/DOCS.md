# buyer-vm1 — Energy Buyer Organization

This folder configures the **Buyer** — an energy consumer/purchaser in the GO network.

## Role
The Buyer represents an energy consumer that:
- Purchases renewable energy GOs (both electricity and hydrogen)
- Receives transferred eGOs and hGOs from producers
- Claims renewable attributes by cancelling GOs (proving renewable energy consumption)
- Has the most complete client implementation, including an experimental **Node.js application**

## Components
- **`b-peer0`** (port 7051) — Buyer's endorsing peer
- **2 CouchDB instances** — State database
- **Fabric CA** (port 7054) — Issues identities for buyer users
- **`buyer-client`** container — CLI for chaincode operations

## Files
| File | Purpose |
|------|---------|
| `.env` | Image versions and compose project name |
| `base.yaml` | Base Docker service definitions (peer, CouchDB) |
| `docker-compose.yaml` | Docker Compose for peer, CouchDBs, CLI |
| `joinChannel.sh` | Joins `b-peer0` to `mychannel` and updates anchor peer |
| `installAndApproveChaincode.sh` | Chaincode lifecycle: package → install → approve |

## Client Scripts (`buyer-client/`)

### `bash/` — Shell Script Client
Full set of CLI scripts for all buyer operations:
| Script | Purpose |
|--------|---------|
| `TransferEGO.sh` / `TransferEGObyAmount.sh` | Transfer eGOs |
| `ClaimRenewableattributesElectricity.sh` / `ClaimRenewableattributesHydrogen.sh` | Cancel GOs for renewable claims |
| `GetCurrentEGOList.sh` | List all electricity GOs |
| `QueryPrivateEGOsbyAmountMWh.sh` | Find eGOs by MWh amount |
| `ReadPublicEGO.sh` / `ReadPublicHGO.sh` | Read public GO metadata |
| `ReadPrivateElectricityGO.sh` / `ReadPrivateHydrogenGO.sh` | Read private GO details |
| `ReadCancelStatement*.sh` / `ReadConsumptionDeclaration*.sh` | Read all statement types |
| `newGatewaypkg.js` | Experimental Fabric Gateway SDK connection test |

### `nodejs/` — Node.js Application (Experimental/Incomplete)
An attempt to build a Node.js frontend/API for the buyer:
| File | Purpose |
|------|---------|
| `app.js` | Main application entry point (Express-like) |
| `modules/gateway.js` | Fabric Gateway SDK connection module |
| `buyer-wallet/` | Identity wallet for the buyer |
| `ccpandcerts/` | Connection profiles and TLS certificates |
| `Dockerfile` / `docker-compose.yaml` | Containerization config |
| `package.json` | Node.js dependencies |

> **Note**: The Node.js application was not fully completed — the Fabric discovery service connection could not be established (see thesis limitations section).
