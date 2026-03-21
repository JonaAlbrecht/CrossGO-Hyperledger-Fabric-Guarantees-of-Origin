# Testing

This folder contains the **Hyperledger Caliper** performance benchmarking setup and test results for the GO network.

## Contents

### `caliper-config-GO/`
Performance testing configuration using Hyperledger Caliper 0.6.0:

- **`bench-config.yaml`** — Benchmark configuration defining 3 test rounds:
  | Round | Transactions | TPS | Workload |
  |-------|-------------|-----|----------|
  | Read Public eGOs | 100 | 40 | `ReadPubliceGO.js` |
  | Read Private eGOs | 100 | 40 | `ReadPrivateeGO.js` |
  | Get current eGO list | 100 | 40 | `GetcurrenteGOlist.js` |

- **`network-config.yaml`** — Caliper network topology: 4 orgs, 8 peers (2 per org), 4 orderers (Raft), CouchDB state, single-host deployment with 5 local workers

- **`docker-compose.yaml`** — Docker config for Caliper containers

- **`Functions/`** — Caliper workload modules (JavaScript):
  - `createElectricityGO.js` — Generates random eGO creation transactions (45–49 MWh, solar, 50 CO₂/MWh)
  - `ReadPubliceGO.js` — Reads public eGO data via transient map
  - `ReadPrivateeGO.js` — Reads private eGO from `privateDetails-buyerMSP`
  - `GetcurrenteGOlist.js` — Range query eGO1–eGO999

- **`connection_profiles/`** — Caliper connection profile for buyer org (`ccp-buyer.yaml`)

### `Test-results/`
Raw test output from the benchmarking runs:
- `caliper third cycle.html` — Caliper HTML report from 3rd design cycle
- `eproducer-client second cycle.txt` — eproducer client logs (2nd cycle)
- `hproducer-client second cycle.txt` — hproducer client logs (2nd cycle)
- `more eproducer-client second cycle.txt` — Extended eproducer logs
- `more hproducer-client second cycle.txt` — Extended hproducer logs
- `OutputMeter second cycle.txt` — OutputMeter timing/output logs
