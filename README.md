# Blockchain-Based Guarantee of Origin Conversion Issuance

**Master's Thesis — Jona Albrecht (2024)**

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
├── Project-Description/          # Academic documentation
│   ├── Literature-Review/        #   GO schemes, requirements analysis
│   ├── platform-agnostic-architecture/  #   Conceptual design (pre-implementation)
│   └── README-files/             #   Step-by-step setup guides
├── testing/                      # Performance benchmarking
│   ├── caliper-config-GO/        #   Hyperledger Caliper configuration & workloads
│   └── Test-results/             #   Benchmark outputs (HTML reports, logs)
└── version1/                     # Fabric network implementation
    ├── artifacts/
    │   ├── channel/              #     Channel config (configtx.yaml, genesis block)
    │   ├── Mychaincode/          #     Go chaincode (conversion.go — core logic)
    │   └── private-data-collections/  #  Collection access policies
    └── setup1/                   #   Deployment per organization
        ├── buyer-vm1/            #     Buyer peer, CA, bash & Node.js clients
        ├── eproducer-vm2/        #     Electricity producer peer, CA, client
        ├── hproducer-vm5/        #     Hydrogen producer peer, CA, client
        ├── issuer-vm3/           #     Issuer peer, CA, SmartMeter, OutputMeter
        └── orderer-vm4/          #     4 Raft orderer nodes, CA
```

> Each folder contains a `DOCS.md` file with detailed documentation of its contents.

## Getting Started

### Prerequisites

- Ubuntu 22.04 (or WSL2)
- Docker & Docker Compose
- Go 1.22.1+
- Node.js 18+
- Hyperledger Fabric binaries 2.x

### Setup Steps

1. **[Set up VM](Project-Description/README-files/Virtual-Machine-Setup.md)** *(optional)* — Provision a Google Cloud VM or use local WSL
2. **[Install prerequisites](Project-Description/README-files/Installing-Prerequisites.md)** — Docker, Go, Node.js, Fabric binaries
3. **[Bring up the network](Project-Description/README-files/Bringing-up-the-network.md)** — CAs, crypto material, orderers, peers, channel
4. **[Deploy chaincode](Project-Description/README-files/Deploying-and-commiting-the-Chaincode.md)** — Package, install, approve, commit
5. **[Execute network functions](Project-Description/README-files/Executing-network-commands.md)** — Create, transfer, convert, and cancel GOs
6. **[Run benchmarks](Project-Description/README-files/Connect-Caliper.md)** — Caliper performance testing

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

## License

This project was developed as part of a Master's thesis at the intersection of blockchain technology and energy systems research.
