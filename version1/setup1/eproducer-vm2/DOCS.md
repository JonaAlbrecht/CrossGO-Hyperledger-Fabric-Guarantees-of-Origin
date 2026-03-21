# eproducer-vm2 — Electricity Producer Organization

This folder configures the **Electricity Producer (E-Producer)** — a solar energy production facility in the GO network.

## Role
The E-Producer represents a renewable electricity generation facility that:
- Receives automatically created **Electricity GOs (eGOs)** from the SmartMeter
- Transfers eGOs to the Hydrogen Producer for conversion or to the Buyer
- Can claim renewable attributes by cancelling eGOs

## Components
- **`e-peer0`** (port 9051) — E-Producer's endorsing peer
- **2 CouchDB instances** — State database
- **Fabric CA** (port 8054) — Issues identities for e-producer users
- **`eproducer-client`** container — CLI for chaincode operations

## Files
| File | Purpose |
|------|---------|
| `.env` | Image versions and compose project name |
| `base.yaml` | Base Docker service definitions (peer, CouchDB) |
| `docker-compose.yaml` | Docker Compose for peer, CouchDBs, CLI |
| `joinChannel.sh` | Joins `e-peer0` to `mychannel` and updates anchor peer |
| `installAndApproveChaincode.sh` | Chaincode lifecycle: package → install → approve |

## Client Scripts (`eproducer-client/`)
| Script | Purpose |
|--------|---------|
| `TransferEGO.sh` | Transfer a single eGO to another org |
| `TransferEGObyAmount.sh` | Transfer eGOs by MWh amount (with splitting) |
| `ClaimRenewableattributesElectricity.sh` | Cancel eGOs to claim renewable attributes |
| `GetCurrenteGOList.sh` | List all current electricity GOs |
| `QueryPrivateeGOsbyAmountMWh.sh` | Find eGOs meeting a MWh threshold |
| `ReadPubliceGO.sh` | Read public eGO metadata |
| `ReadPrivateElectricityGO.sh` / `ReadPrivateHydrogenGO.sh` | Read private GO details |
| `ReadCancelStatementElectricity.sh` / `ReadCancelStatementHydrogen.sh` | Read cancellation statements |
| `ReadConsumptionDeclarationElectricity.sh` / `ReadConsumptionDeclarationHydrogen.sh` | Read consumption declarations |
