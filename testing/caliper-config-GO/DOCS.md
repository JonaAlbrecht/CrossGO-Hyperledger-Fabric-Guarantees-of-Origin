# caliper-config-GO — Hyperledger Caliper Performance Benchmarks

Configuration for benchmarking the GO network using **Hyperledger Caliper 0.6.0**.

## Files

### `bench-config.yaml`
Defines 3 benchmark testing rounds:

| Round | Transactions | Target TPS | Workload Module |
|-------|-------------|------------|-----------------|
| Read Public eGOs | 100 | 40 | `ReadPubliceGO.js` |
| Read Private eGOs | 100 | 40 | `ReadPrivateeGO.js` |
| Get current eGO list | 100 | 40 | `GetcurrenteGOlist.js` |

### `network-config.yaml`
Caliper network topology configuration:
- 4 organizations (buyer, eproducer, hproducer, issuer)
- 8 peers total (2 per org)
- 4 Raft orderers
- CouchDB state database
- 5 local workers
- Single-host deployment

### `docker-compose.yaml`
Docker configuration for running Caliper containers alongside the Fabric network.

### `Functions/` — Workload Modules (JavaScript)
| Module | Description |
|--------|-------------|
| `createElectricityGO.js` | Generates random eGO creation transactions with simulated solar production data (45–49 MWh, 50 CO₂/MWh emission intensity) |
| `ReadPubliceGO.js` | Reads public eGO1 data via transient map |
| `ReadPrivateeGO.js` | Reads private eGO1 details from `privateDetails-buyerMSP` collection |
| `GetcurrenteGOlist.js` | Range query returning all eGOs from eGO1 to eGO999 |

### `connection_profiles/`
- `ccp-buyer.yaml` — Connection profile for the buyer organization used by Caliper workers

## Running Benchmarks
```bash
npx caliper launch manager \
  --caliper-workspace ./ \
  --caliper-networkconfig network-config.yaml \
  --caliper-benchconfig bench-config.yaml \
  --caliper-flow-only-test
```
