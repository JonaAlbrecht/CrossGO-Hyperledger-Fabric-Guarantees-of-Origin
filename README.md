# Blockchain-Based Guarantee of Origin Conversion Issuance

A Hyperledger Fabric prototype for **cross-carrier interoperability** between Guarantee of Origin (GO) schemes, enabling verifiable data sharing and attribute transfer during energy carrier conversion (electricity → hydrogen).

## Overview

This project implements a permissioned blockchain network where electricity Guarantees of Origin can be converted into hydrogen Guarantees of Origin during energy carrier conversion processes. The system ensures that renewable energy attributes (emissions, production method, production time) are verifiably inherited across energy carriers — a key requirement for decarbonising multi-carrier energy systems.

Built as part of a Design Science Research project investigating:

> *How to design a system architecture of interoperable GO schemes that allows verifiable cross-domain data sharing and asset transfer to enable energy carrier conversion processes in multi-carrier energy systems?*

## Architecture

### Network Topology

| Organization | Role | Description |
|-------------|------|-------------|
| **Issuer** | Registry operator & trust anchor | Manages GO lifecycle, registers metering devices, audits all collections |
| **E-Producer** | Electricity producer | Solar energy facility; receives auto-generated electricity GOs |
| **H-Producer** | Hydrogen producer | PEM electrolyser facility; converts electricity GOs into hydrogen GOs |
| **Buyer** | Energy consumer | Purchases and claims renewable attributes from GOs |
| **Orderer** | Consensus service | 4-node Raft ordering service |

### Key Features

- **Smart contract** (~2,350 lines Go) implementing full GO lifecycle: creation, transfer, conversion, cancellation, and verification
- **Attribute-Based Access Control (ABAC)** via X.509 certificate attributes for metering devices and trusted users
- **Private Data Collections** for confidential GO attributes (amount, emissions, production method) with public metadata on-chain
- **Automated IoT metering** via SmartMeter (electricity) and OutputMeter (hydrogen) Docker containers with cron-triggered chaincode invocations
- **GO splitting** for partial transfers and cancellations with proportional attribute allocation
- **Cross-carrier attribute inheritance** during conversion: hydrogen GOs inherit emissions, production methods, and consumption declarations from cancelled electricity GOs
- **Performance benchmarking** via Hyperledger Caliper

## Repository Structure

```
├── network/                      # Hyperledger Fabric network
│   ├── configtx.yaml             #   Channel/profile definitions (electricity-de, hydrogen-de)
│   ├── crypto-config.yaml        #   cryptogen org topology (6 orgs + orderer)
│   ├── docker/                   #   docker-compose files (orderer + 6 peer stacks + CA)
│   └── scripts/                  #   network-up / network-down / deploy-chaincode
├── chaincode/                    # `golifecycle` Go chaincode (v10.5.4)
│   ├── contracts/                #   GO lifecycle, conversion, cross-channel bridge
│   ├── access/                   #   ABAC policies
│   └── main.go
├── collections/                  # Private data collection configs (per channel)
├── application/
│   ├── backend/                  # Node.js / Express REST API (Fabric Gateway)
│   └── frontend/                 # React + Vite SPA
├── testing/                      # Hyperledger Caliper benchmarks + result logs
├── docs/adrs/                    # Architecture Decision Records (ADR-030, 031, 033)
├── ARCHITECTURE.md               # Full architecture (current: v10.5.4)
└── version1/                     # Preserved original prototype (read-only reference)
```

## Getting Started

### Prerequisites

- Ubuntu 22.04 (or WSL2 / a Linux VM)
- Docker Engine + Docker Compose v2 (`docker compose ...`)
- Go 1.22.1+
- Node.js 18+ and npm
- `jq`
- Hyperledger Fabric 2.5+ binaries (`peer`, `orderer`, `cryptogen`, `configtxgen`, `configtxlator`, `osnadmin`)

Install the Fabric binaries into `./fabric-bin/` at the repository root (the scripts add `fabric-bin/bin` to `PATH`):

```bash
curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh \
  | bash -s -- --fabric-version 2.5.9 binary
mv fabric-samples/bin   fabric-bin/bin
mv fabric-samples/config fabric-bin/config
```

### 1. Bring up the Fabric network

The current topology (per ADR-030) is **two carrier-specific channels** sharing a 4-node Raft orderer cluster:

| Channel | Member orgs | Issuer |
|---------|-------------|--------|
| `electricity-de` | `eissuerMSP`, `eproducer1MSP`, `ebuyer1MSP` | `eissuer` |
| `hydrogen-de`    | `hissuerMSP`, `hproducer1MSP`, `hbuyer1MSP` | `hissuer` |

Bring everything up from the `network/scripts/` directory:

```bash
cd network/scripts

# Tear down any previous state (volumes, chaincode containers/images)
./network-down.sh

# 1a. Crypto material (cryptogen) + orderer-path symlinks
cryptogen generate \
  --config=../crypto-config.yaml \
  --output=../organizations
ORDERER_BASE=../organizations/ordererOrganizations/orderer.go-platform.com/orderers
for i in 1 2 3 4; do
  ln -sfn "$ORDERER_BASE/orderer${i}.orderer.go-platform.com" \
          "$ORDERER_BASE/orderer${i}.go-platform.com"
done

# 1b. Genesis blocks for both channels
mkdir -p ../channel-artifacts
configtxgen -profile ElectricityDEChannel \
  -outputBlock ../channel-artifacts/electricity-de.block -channelID electricity-de
configtxgen -profile HydrogenDEChannel \
  -outputBlock ../channel-artifacts/hydrogen-de.block   -channelID hydrogen-de

# 1c. Start orderers + all 6 peer stacks (each with its own CouchDB)
docker compose -f ../docker/docker-compose-orderer.yaml   up -d
docker compose -f ../docker/docker-compose-eissuer.yaml   up -d
docker compose -f ../docker/docker-compose-hissuer.yaml   up -d
docker compose -f ../docker/docker-compose-eproducer.yaml up -d
docker compose -f ../docker/docker-compose-hproducer.yaml up -d
docker compose -f ../docker/docker-compose-ebuyer.yaml    up -d
docker compose -f ../docker/docker-compose-hbuyer.yaml    up -d
```

Then join orderers and peers to their channels (the `osnadmin channel join` and `peer channel join` calls from `network-up-v8.sh` work as-is once the 6-org docker stacks are running — see that script for the exact loops; only the `ORGS_ELEC` / `ORGS_H2` lists must be updated to the v10 MSP IDs above).

**Verify the network is up:**

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" \
  | grep -E "orderer|peer0|couchdb"
```

You should see 4 orderers, 6 peers, and 6 CouchDB containers, all healthy.

### 2. Deploy the chaincode

The `golifecycle` chaincode (current version `v10.5.4`) is deployed to **both** channels with the appropriate private-data collection config:

```bash
# Still in network/scripts
./deploy-chaincode.sh
```

This packages, installs, approves and commits `golifecycle` on `electricity-de` (using `collections/collection-config-electricity-de.json`) and `hydrogen-de` (using `collections/collection-config-hydrogen-de.json`).

### 3. Start the backend (REST API)

```bash
cd application/backend
cp .env.example .env       # then edit endpoints / TLS paths if needed
npm install
npm run dev                # ts-node-dev on http://localhost:3001
```

Health check:

```bash
curl http://localhost:3001/api/health
# { "status": "ok", "service": "go-platform-backend", "version": "9.0.0" }
```

The backend exposes `/api/auth`, `/api/devices`, `/api/guarantees`, `/api/transfers`, `/api/conversions`, `/api/cancellations`, `/api/queries`, `/api/bridge`, `/api/organizations`. It connects to the peers via the Fabric Gateway (gRPC) using crypto material under `network/organizations/`.

### 4. Start the frontend (UI)

In a second terminal:

```bash
cd application/frontend
npm install
npm run dev                # Vite on http://localhost:5173
```

Open <http://localhost:5173> in your browser. The SPA talks to the backend at `http://localhost:3001` (configurable via `VITE_API_URL` if you change the backend port).

Login with one of the seeded role identities (Issuer / Producer / Buyer for either carrier) — JWT is issued by the backend and stored client-side; subsequent calls are signed with the matching Fabric identity from the wallet on the backend.

### 5. Tear down

```bash
cd network/scripts
./network-down.sh          # stops all containers, volumes and chaincode images
```

Stop the backend and frontend with `Ctrl+C` in their terminals.

### 6. Optional — Performance benchmarks

```bash
cd testing/caliper-config-GO
npx caliper launch manager \
  --caliper-bind-sut fabric:2.5 \
  --caliper-networkconfig networks/networkConfig.yaml \
  --caliper-benchconfig benchmarks/myAssetBenchmark.yaml \
  --caliper-workspace .
```

See [`testing/`](testing/) for prior result logs and `ARCHITECTURE.md` §11 for the latest v10.5.4 numbers.

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Blockchain platform | Hyperledger Fabric 2.x |
| Smart contract | Go (`contractapi`) |
| Consensus | Raft (etcdraft, 4 nodes) |
| State database | CouchDB |
| Identity management | Fabric CA with X.509 certificates |
| Performance testing | Hyperledger Caliper 0.6.0 |
| Client application | Bash scripts / Node.js (experimental) |
| Deployment | Docker Compose |

