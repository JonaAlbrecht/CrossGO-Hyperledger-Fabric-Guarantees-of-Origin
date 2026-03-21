# issuer-vm3 — Issuing Body / Registry Operator

This folder configures the **Issuer organization** — the central authority and trust anchor of the GO network.

## Role
The Issuer acts as:
- **Registry operator** managing GO lifecycle
- **Trust anchor** — its CA registers all metering devices (SmartMeter, OutputMeter) with audited production attributes
- **Auditor** — has read access to all private data collections

## Components

### Peer Infrastructure
- **`i-peer0`** (port 11051) — Issuer's endorsing peer
- **2 CouchDB instances** — State database for the peer
- **Fabric CA** (port 10054) — Issues identities for issuer, metering devices, and trusted users
- **`issuer-client`** container — CLI for chaincode operations

### IoT Metering Devices

#### SmartMeter (`SmartMeter-config/`)
Simulates an electricity production metering device:
- Docker container based on `hyperledger/fabric-tools` with cron
- **Cron job**: Runs `execute.sh` every minute
- **`execute.sh`**: Generates random solar electricity data (45–49 MWh, 50 CO₂/MWh emission intensity, 60s elapsed) and invokes `CreateElectricityGO` with transient data
- Enrolled with X.509 attributes: `electricitytrustedDevice=true`, `maxEfficiency=50`, `technologyType=solar`, `emissionIntensity=41`

#### OutputMeter (`OutputMeter-config/`)
Simulates a hydrogen production output meter:
- Docker container based on `hyperledger/fabric-tools` with cron
- **Cron job**: Runs `execute-hproducer.sh` every minute
- **`execute-hproducer.sh`**: Generates random hydrogen production data (~150 kg/min, 4 CO₂/kg, 50 kWh/kg) and invokes `AddHydrogentoBacklog` with transient data
- Enrolled with X.509 attributes: `hydrogentrustedDevice=true`, `maxOutput=10000`, `conversionEfficiency=0.8`, `technologyType=PEMelectrolyser`, `emissionIntensity=20`

## Scripts
| Script | Purpose |
|--------|---------|
| `createChannel.sh` | Fetches genesis block and joins issuer peer to `mychannel` |
| `deployChaincode.sh` | Chaincode lifecycle: package → install → query installed → approve |
| `deployChaincode-fullscript.sh` | Same with absolute paths for automated deployment |
| `test.sh` | Quick test invocations |

## Client Scripts (`issuer-client/`)
| Script | Purpose |
|--------|---------|
| `commitChaincode.sh` | Commits chaincode definition (requires all 4 peer endorsements) and invokes init |
| `GetCurrenteGOList.sh` | Lists all current electricity GOs |
| `QueryHydrogenBacklog.sh` | Reads hydrogen production backlog |
| `ReadPublicEGO.sh` / `ReadPublicHGO.sh` | Read public GO metadata |
| `ReadPrivateElectricityGO.sh` / `ReadPrivateHydrogenGO.sh` | Read private GO details |
