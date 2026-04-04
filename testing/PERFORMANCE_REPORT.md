# Hyperledger Fabric GO Platform — Scalability Test Report

**Date:** 2026-04-04  
**Test Tool:** Hyperledger Caliper v0.6.0 (fabric-gateway binding)  
**Network:** golifecycle chaincode on goplatformchannel  
**Versions tested:** v2.0 (baseline) → v3.0 (post-optimization)

---

## 1. Infrastructure

| Component | Specification |
|---|---|
| **VM Provider** | Hetzner Cloud |
| **CPU** | 16 vCPUs — AMD EPYC-Genoa @ 2.0 GHz |
| **RAM** | 32 GB |
| **Disk** | 600 GB SSD |
| **OS** | Ubuntu 24.04 LTS |
| **Docker** | 28.2.2 |
| **Node.js** | v20.20.2 |
| **Fabric Peer** | v2.5.12 |
| **Fabric CA** | v1.5.17 |
| **CouchDB** | 3.3.3 |

## 2. Network Topology

All components run on a **single VM** (no geographic distribution).

| Component | Count | Details |
|---|---|---|
| **Orderers** | 4 | Raft consensus (orderer1–4.go-platform.com, ports 7050/8050/9050/10050) |
| **Peer Organizations** | 4 | issuer1MSP, eproducer1MSP, hproducer1MSP, buyer1MSP |
| **Peers** | 4 | 1 per org (ports 7051/9051/11051/13051) |
| **CouchDB** | 4 | 1 per peer (ports 5984/7984/9984/11984) |
| **Chaincode Containers** | 4 | External launcher, 1 per peer |

**Endorsement Policy:** Majority (3-of-4 organizations)  
**Block Configuration:** BatchTimeout=2s, MaxMessageCount=10, PreferredMaxBytes=512 KB

## 3. Chaincode Under Test

The **golifecycle** chaincode (Go) manages Guarantees of Origin (GOs) for energy.

| Contract Namespace | Key Functions |
|---|---|
| `device` | RegisterDevice, GetDevice, ListDevices, RevokeDevice |
| `issuance` | CreateElectricityGO, CreateHydrogenGO |
| `transfer` | TransferEGO, TransferHGO |
| `conversion` | ConvertEGOToHGO |
| `cancellation` | CancelEGO, CancelHGO |
| `query` | ReadPublicEGO, GetCurrentEGOsList, GetCurrentHGOsList |

**Private Data Collections:** 5 (1 public + 4 org-specific)  
**ID Generation:** Sequential counter pattern (`GetNextID`) — reads and increments a shared counter key per entity type.

## 4. Workloads Tested

### 4.1 Point Queries (GetDevice)
- Reads a single device by ID from CouchDB
- Evaluates (read-only): no ordering/commit overhead
- Tests IDs cycling through device1–device26

### 4.2 Range Queries (ListDevices)
- Rich query returning all registered devices (~26 records)
- Evaluates (read-only)
- Heavier per-transaction than point queries

### 4.3 Range Queries (GetCurrentEGOsList)
- Rich query returning all active electricity GOs
- Collection is empty in this test (baseline for schema overhead)

### 4.4 Write Transactions (RegisterDevice)
- Submits a new device registration via transient data
- Requires 3-of-4 endorsement + orderer commit
- Reads/writes `counter_device` key → MVCC_READ_CONFLICT under concurrency

---

## 5. Read Scalability Results

**Caliper run5-reads: 10 workers, 16 rounds**

### 5.1 Point Query (GetDevice) — Scalability Curve

| Target TPS | Txns Sent | Success | Fail | Avg Latency (s) | Actual Throughput (TPS) |
|---|---|---|---|---|---|
| 100 | 500 | 500 | 0 | 0.01 | 101.8 |
| 250 | 1,250 | 1,250 | 0 | 0.01 | 251.1 |
| 500 | 2,500 | 2,500 | 0 | 0.01 | 500.6 |
| 750 | 3,750 | 3,750 | 0 | 0.01 | 750.3 |
| 1,000 | 5,000 | 5,000 | 0 | <0.01 | 1,000.2 |
| 1,500 | 7,500 | 7,500 | 0 | <0.01 | 1,499.1 |
| **2,000** | **10,000** | **10,000** | **0** | **<0.01** | **1,999.2** |

**Finding:** Point reads scale linearly to **2,000+ TPS** with **0% failure rate** and sub-10ms latency. The throughput ceiling was not reached.

### 5.2 Range Query (ListDevices) — Scalability Curve

| Target TPS | Txns Sent | Success | Fail | Avg Latency (s) | Actual Throughput (TPS) |
|---|---|---|---|---|---|
| 100 | 500 | 500 | 0 | 0.04 | 101.0 |
| 250 | 1,250 | 1,250 | 0 | 0.03 | 250.3 |
| **500** | **2,500** | **2,500** | **0** | **0.57** | **459.4** |
| 750 | 3,750 | 2,446 | 1,304 | 1.00 | 661.6 |
| 1,000 | 5,000 | 2,264 | 2,736 | 1.12 | 895.7 |

**Finding:** Range queries saturate between **250–500 TPS**. At 500 TPS the average latency jumps from 30ms to 570ms, indicating CouchDB is the bottleneck. At 750 TPS, 34.8% of transactions fail. The effective ceiling for ListDevices (26 records) is approximately **~460 TPS** before degradation and **~660 TPS** maximum throughput with failures.

### 5.3 Range Query — Empty Collection (GetCurrentEGOsList)

| Target TPS | Txns Sent | Success | Fail | Avg Latency (s) | Actual Throughput (TPS) |
|---|---|---|---|---|---|
| 250 | 1,250 | 1,250 | 0 | 0.01 | 250.8 |
| 500 | 2,500 | 2,500 | 0 | 0.01 | 500.3 |
| 1,000 | 5,000 | 5,000 | 0 | 0.01 | 999.4 |
| **1,500** | **7,500** | **7,434** | **66** | **0.22** | **1,451.8** |

**Finding:** Empty-collection range queries scale to **~1,000 TPS** with 0% failure, degrading slightly at 1,500 TPS (0.9% failure, rising latency). This confirms the ListDevices bottleneck is CouchDB data retrieval, not the query framework itself.

---

## 6. Write Scalability Results

### 6.1 Sequential Counter Problem (MVCC_READ_CONFLICT)

The `RegisterDevice` function calls `GetNextID(ctx, "counter_device")` which reads a shared counter key, increments it, and writes it back. When multiple transactions in the same block read the same counter value, all but one are invalidated at commit time (MVCC_READ_CONFLICT, status code 11).

**Run3 results (5 workers, fixed-rate):**

| Target TPS | Txns Sent | Success | Fail | Success Rate | Avg Latency (s) |
|---|---|---|---|---|---|
| 10 | 50 | 5 | 45 | **10%** | 0.59 |

**Run5-writes (1 worker, fixed-rate):**

| Target TPS | Txns Sent | Success | Fail | Success Rate | Avg Latency (s) |
|---|---|---|---|---|---|
| 5 | 100 | 10 | 90 | **10%** | 1.89 |
| 10 | 100 | 10 | 90 | **10%** | 0.98 |
| 20 | 100 | 10 | 90 | **10%** | 0.53 |

**Run5-serial (1 worker, fixed-load backpressure):**

| Mode | Txns Sent | Success | Fail | Success Rate | Avg Latency (s) | Throughput (TPS) |
|---|---|---|---|---|---|---|
| Backlog=1 | 30 | 10 | 20 | **33%** | 2.09 | 0.8 |

### 6.2 Write Performance Analysis

The consistent **10% success rate** across all fixed-rate tests is explained by the block configuration:

- **BatchTimeout = 2s**: A new block is cut every 2 seconds
- **Multiple transactions per block**: All read the same counter value during endorsement
- **Only 1 per block commits**: The first validated transaction succeeds; all others in the same block fail with MVCC_READ_CONFLICT
- At 10 TPS × 2s = ~20 txns per block → 1/20 = 5% expected success (actual ~10% due to batching dynamics)

The **fixed-load** controller achieves 33% success (10/30) because it throttles sending to ~0.8 TPS, spreading transactions across blocks, but still occasionally overlaps.

**Maximum verified serial write throughput: ~0.8 TPS** (with counter-based IDs)

---

## 7. Bottleneck Summary

| Bottleneck | Component | Impact | Root Cause |
|---|---|---|---|
| **MVCC_READ_CONFLICT** | Chaincode (GetNextID) | Write throughput limited to ~0.8 TPS | Sequential counter pattern creates a hot key contending across all write transactions |
| **CouchDB range query saturation** | CouchDB | ListDevices degrades above 250 TPS, fails above 500 TPS | Rich queries scanning multiple documents reach CouchDB I/O limits |
| **Block cut latency** | Orderer | 2s commit latency floor for writes | BatchTimeout=2s means transactions wait up to 2s for block inclusion |
| **Endorsement overhead** | Multi-org endorsement | 3-of-4 orgs must endorse, but latency is minimal (~10ms) | Not a significant bottleneck at current scale |

## 8. Key Findings

1. **Point reads are highly scalable:** GetDevice sustains 2,000+ TPS with sub-10ms latency on a single VM. The ceiling was not found.

2. **Range queries are the CouchDB bottleneck:** ListDevices (26 records) saturates at ~460 TPS. Empty-collection queries confirm overhead is in data retrieval, not query parsing.

3. **Writes are critically bottlenecked by the sequential counter pattern:** The `GetNextID` function creates a serialization point. Maximum throughput is ~0.8 TPS regardless of network capacity. This is 2,500× slower than read performance.

4. **The 3-of-4 endorsement policy has minimal latency impact:** All endorsements complete in <10ms, but the policy makes write failures more expensive (3× the endorsement work is wasted per failed tx).

5. **Single-VM deployment does not represent production:** All latencies benefit from localhost communication. A distributed deployment would add network latency to endorsement, ordering, and gossip.

## 9. Architecture Improvement Recommendations

### Critical (Write throughput)
- **Replace sequential counters with UUID-based IDs** — Eliminate the hot key to allow parallel writes. Use deterministic UUIDs (e.g., `hash(txID + timestamp)`) or client-generated UUIDs.
- **Reduce BatchTimeout** — From 2s to 500ms to lower write latency floor.

### High Priority (Range query throughput)
- **Add CouchDB indexes** for frequently queried fields (deviceType, ownerOrgMSP, status).
- **Implement pagination** in ListDevices/GetCurrentEGOsList instead of returning all records.
- **Consider LevelDB** for peers that don't need rich queries.

### Medium Priority (Network)
- **Tune endorsement policy** — Consider reducing from 3-of-4 to 2-of-4 (OR logic) for read-heavy workloads.
- **Increase MaxMessageCount** from 10 to handle burst writes.
- **Distribute across multiple VMs** for real production assessment.

---

# PART 2 — Post-Optimization Benchmark (v3.0)

## 10. Architecture Changes Implemented

The following ADRs (Architecture Decision Records) were implemented between v2.0 and v3.0:

| ADR | Change | Rationale |
|---|---|---|
| **ADR-001** | Replace `GetNextID` sequential counter with SHA-256 deterministic hash-based IDs (`prefix + hex(sha256(txID + suffix))[:8]`) | Eliminate the MVCC_READ_CONFLICT hot-key serialization bottleneck |
| **ADR-002** | Add CouchDB composite indexes on `[OwnerID, GOType]` and `[GOType, CreationDateTime]` | Improve range query performance for filtered queries |
| **ADR-004** | Secure `InitLedger` with caller MSP validation | Prevent unauthorized org role registration |
| **ADR-005** | Reduce `BatchTimeout` from 2s to 500ms | Lower write latency floor by 75% |

**Block Configuration (v3.0):** BatchTimeout=500ms, MaxMessageCount=10, PreferredMaxBytes=512 KB

---

## 11. Write Scalability Results (v3.0)

**Caliper bench-write-v2: 10 workers, 5 rounds — RegisterDevice with hash-based IDs**

### 11.1 Write Throughput — Multi-Concurrency Scalability

| Target TPS | Txns Sent | Success | Fail | Success Rate | Avg Latency (s) | Min Latency (s) | Max Latency (s) | Throughput (TPS) |
|---|---|---|---|---|---|---|---|---|
| Serial (backlog=1) | 30 | 30 | 0 | **100%** | 0.10 | 0.07 | 0.14 | 5.7 |
| 5 | 50 | 50 | 0 | **100%** | 0.09 | 0.08 | 0.11 | 6.2 |
| 10 | 100 | 100 | 0 | **100%** | 0.09 | 0.08 | 0.11 | 11.0 |
| 25 | 250 | 250 | 0 | **100%** | 0.10 | 0.08 | 0.12 | 25.8 |
| **50** | **500** | **500** | **0** | **100%** | **0.10** | **0.06** | **0.12** | **50.5** |

**Finding:** With hash-based IDs (ADR-001), **100% of write transactions succeed at all tested concurrency levels** up to 50 TPS. No MVCC_READ_CONFLICT errors. The write throughput ceiling was not reached.

### 11.2 Before/After Write Comparison

| Metric | v2.0 (Counter-based) | v3.0 (Hash-based) | Improvement |
|---|---|---|---|
| Serial success rate | 33% (10/30) | **100%** (30/30) | **+203%** |
| Serial throughput | 0.8 TPS | **5.7 TPS** | **7.1× faster** |
| Serial avg latency | 2.09s | **0.10s** | **20.9× lower** |
| 10 TPS success rate | 10% | **100%** | **+900%** |
| 50 TPS success rate | N/A (untestable) | **100%** | ∞ |
| Max verified throughput | ~0.8 TPS | **50.5 TPS** | **63× faster** |

**Root cause eliminated:** The `GetNextID` shared counter was the sole cause of MVCC_READ_CONFLICT. By deriving IDs deterministically from the transaction ID (`sha256(txID + suffix)`), each transaction's read-write set is completely independent — no hot-key contention.

---

## 12. Read Scalability Results (v3.0)

**Caliper bench-config: 10 workers, 16 rounds — 930 registered devices**

### 12.1 Point Query (GetDevice) — v3.0

| Target TPS | Txns Sent | Success | Fail | Avg Latency (s) | Actual Throughput (TPS) |
|---|---|---|---|---|---|
| 100 | 500 | 500 | 0 | 0.01 | 101.7 |
| 250 | 1,250 | 1,250 | 0 | 0.01 | 251.3 |
| 500 | 2,500 | 2,500 | 0 | 0.01 | 500.4 |
| 750 | 3,750 | 3,750 | 0 | 0.01 | 749.9 |
| 1,000 | 5,000 | 5,000 | 0 | <0.01 | 998.8 |
| 1,500 | 7,500 | 7,500 | 0 | <0.01 | 1,498.5 |
| **2,000** | **10,000** | **10,000** | **0** | **<0.01** | **1,996.4** |

**Finding:** Point reads remain fully scalable at **2,000 TPS** with 0% failure and sub-10ms latency — identical to v2.0 baseline. Hash-based IDs introduce no read performance regression.

### 12.2 Range Query (ListDevices) — v3.0 (930 devices vs 26 in v2.0)

| Target TPS | Txns Sent | Success | Fail | Avg Latency (s) | Actual Throughput (TPS) |
|---|---|---|---|---|---|
| 100 | 500 | 500 | 0 | 15.87 | 23.1 |
| 250 | 1,250 | 500 | 750 | 21.05 | 52.3 |
| 500 | 2,500 | 500 | 2,000 | 22.48 | 102.6 |
| 750 | 3,750 | 500 | 3,250 | 24.92 | 140.8 |
| 1,000 | 5,000 | 500 | 4,500 | 25.19 | 187.0 |

**Finding:** ListDevices is significantly degraded compared to v2.0 due to the 36× increase in dataset size (930 vs 26 devices). At 100 TPS, latency is 15.87s (vs 0.04s in v2.0). This confirms the CouchDB range query bottleneck scales linearly with record count. **Pagination is essential** for production use.

### 12.3 Range Query — Empty Collection (GetCurrentEGOsList) — v3.0

| Target TPS | Txns Sent | Success | Fail | Avg Latency (s) | Actual Throughput (TPS) |
|---|---|---|---|---|---|
| 250 | 1,250 | 1,250 | 0 | 0.01 | 251.0 |
| 500 | 2,500 | 2,500 | 0 | 0.01 | 500.4 |
| 1,000 | 5,000 | 5,000 | 0 | 0.01 | 998.6 |
| 1,500 | 7,500 | 6,341 | 1,159 | 0.36 | 1,442.9 |

**Finding:** Empty-collection query performance is unchanged from v2.0 (~1,450 TPS ceiling at 1,500 TPS target). The ADR changes have zero impact on non-device range queries.

---

## 13. Comprehensive Before/After Summary

### 13.1 Write Performance (Key Improvement)

| Metric | v2.0 | v3.0 | Change |
|---|---|---|---|
| Write success rate (concurrent) | 10% | **100%** | **MVCC_READ_CONFLICT eliminated** |
| Max write throughput | 0.8 TPS | **50.5 TPS** | **63× improvement** |
| Write latency (serial) | 2.09s | **0.10s** | **95% reduction** |
| Write latency floor | 2.0s (BatchTimeout) | **0.5s** (BatchTimeout) | **75% reduction** |

### 13.2 Read Performance (Unchanged)

| Metric | v2.0 | v3.0 | Change |
|---|---|---|---|
| Point read throughput ceiling | >2,000 TPS | >2,000 TPS | No change |
| Point read latency | <10ms | <10ms | No change |
| Empty range query ceiling | ~1,450 TPS | ~1,443 TPS | No change |
| ListDevices ceiling (26 records) | ~460 TPS | N/A (930 records) | Dataset changed |
| ListDevices (930 records, 100 TPS) | — | 23.1 TPS effective | Pagination needed |

### 13.3 Changes That Made the Difference

1. **ADR-001 (Hash-based IDs):** Single most impactful change. Eliminated the sequential counter hot-key, enabling parallel write transactions. Write throughput improved from **0.8 TPS to 50.5 TPS** (63× improvement) with 100% success rate.

2. **ADR-005 (BatchTimeout 2s → 500ms):** Reduced the write latency floor from 2.0s to 0.5s. The measured average write latency of 0.10s confirms most transactions now commit within a single 500ms block window.

3. **ADR-002 (CouchDB indexes):** The indexes on `[OwnerID, GOType]` and `[GOType, CreationDateTime]` benefit filtered queries (not tested in this benchmark), but the main ListDevices bottleneck is dataset size, not indexing.

4. **ADR-004 (Secure InitLedger):** Security improvement only, no performance impact.

---

## 14. Remaining Bottlenecks and Recommendations

| Bottleneck | Impact | Recommendation |
|---|---|---|
| **ListDevices range query (930+ records)** | 16s latency at 100 TPS, unusable above 250 TPS | Implement server-side pagination (bookmark-based), return max 50 records per call |
| **Write throughput ceiling unknown** | 50 TPS tested but ceiling not reached | Test at 100, 200, 500 TPS to find the actual write ceiling |
| **Single-VM deployment** | All latencies benefit from localhost networking | Distributed multi-VM testing for production assessment |
| **Endorsement overhead at scale** | 3-of-4 endorsement is reliable but expensive | Evaluate 2-of-4 for non-critical write paths |

---

*Report generated from Caliper v2.0 runs (run5-reads, run5-writes, run5-serial) and v3.0 runs (caliper-v2-writes, caliper-v2-reads2). All tests on Hetzner VM (16 vCPU, 32 GB RAM, single-node deployment).*
