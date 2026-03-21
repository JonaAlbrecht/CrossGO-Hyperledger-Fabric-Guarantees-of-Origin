# version1 — Hyperledger Fabric Network Implementation

This is the **main implementation** of the blockchain-based GO conversion issuance system built on Hyperledger Fabric 2.x.

## Architecture Overview

### Network Topology

| Organization | MSP ID | Role | Peer | Port | CA Port |
|-------------|--------|------|------|------|---------|
| **Buyer** | `buyerMSP` | Energy consumer/purchaser | `b-peer0` | 7051 | 7054 |
| **E-Producer** | `eproducerMSP` | Electricity producer (solar) | `e-peer0` | 9051 | 8054 |
| **Issuer** | `issuerMSP` | Issuing body / registry operator | `i-peer0` | 11051 | 10054 |
| **Orderer** | `OrdererMSP` | Raft ordering service (4 nodes) | orderer1–4 | 7050–10050 | 9054 |
| **H-Producer** | `hproducerMSP` | Hydrogen producer (PEM electrolyser) | `h-peer0` | 13051 | 11054 |

- **Consensus**: Raft (etcdraft) with 4 orderer nodes
- **Channel**: Single channel (`mychannel`)
- **State DB**: CouchDB per peer
- **TLS**: Enabled throughout

### Special IoT Identities
- **SmartMeter** — Automated electricity metering device (X.509 attributes: `electricitytrustedDevice=true`, `maxEfficiency=50`, `technologyType=solar`, `emissionIntensity=41`)
- **OutputMeter** — Automated hydrogen output meter (X.509 attributes: `hydrogentrustedDevice=true`, `maxOutput=10000`, `conversionEfficiency=0.8`, `technologyType=PEMelectrolyser`)

## Folder Structure

### `artifacts/`
Network artifacts (channel config, chaincode, private data collections). See [artifacts/DOCS.md](artifacts/DOCS.md).

### `setup1/`
Deployment configuration for all 5 organization VMs, lifecycle scripts, and client applications. See [setup1/DOCS.md](setup1/DOCS.md).
