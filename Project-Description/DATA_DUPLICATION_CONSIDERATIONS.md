# Data Duplication and Storage Overhead Considerations

## Hyperledger Fabric GO Platform vs. Centralised Architecture

> **Context:** This analysis examines the compute and storage overhead introduced by
> operating the Guarantee of Origin platform on Hyperledger Fabric compared to the
> status quo: **27 separate national GO registries** (e.g., Grexel/NECS, Pronovo,
> UBA/HKNR), each operating its own database infrastructure. The comparison
> baseline is therefore not a single centralised database but a federation of 27
> independent systems — one per EU member state — as mandated by the Renewable
> Energy Directive. The analysis uses the **v10.5.4 `golifecycle` chaincode** as the
> reference implementation and combines (a) the Caliper read benchmarks and
> peer-CLI single-phase write benchmarks collected on the original 4-org single-VM
> deployment with (b) the **dual-channel cross-channel conversion benchmark**
> (electricity-de + hydrogen-de, 6 peers, 4 Raft orderers) reported in Section 7
> of the main paper.

---

## 1. Data Replication Model

### 1.1 Fabric Replication Architecture

In HLF, data is replicated at three distinct levels:

| Layer | Replication Factor | Contents |
|---|---|---|
| **Block ledger** | N (every peer) | Immutable transaction log — all endorsed transactions for the channel |
| **Public world state (CouchDB)** | N (every peer) | Current key-value state for all public assets (eGO, hGO, bGO, devices, org roles) |
| **Private data collections** | K (collection members only) | Confidential GO details, hashes disseminated to all peers |
| **Private data hashes** | N (every peer) | SHA-256 hash of every private data entry, stored on the block ledger |

Where **N = number of peers** (4 in the current deployment) and **K = collection membership count** (typically 1–2: the owning org + issuer).

### 1.2 Status Quo: 27 National Registry Architectures

The current European GO system is **not** a single centralised database. It consists of 27 independent national registries (one per EU member state), each operated by a national competent body. Each registry already maintains its own infrastructure:

| Layer | Replication Factor (per registry) | Contents |
|---|---|---|
| **Primary database** | 1 | Source of truth for that country's GO records |
| **Read replicas** (optional) | 1–2 | Query offloading; eventual consistency |
| **Backup** | 1 | Point-in-time recovery (not live) |

Total live copies per registry: **1–3** (typically 1 primary + 1 replica).

**Across the EU, the status quo already involves 27 separate database deployments**, each storing its national GO data in isolation. Cross-border transfers are handled via the AIB Hub — an asynchronous message-passing system — not through shared database access. This means the fair baseline for an EU-wide comparison is 27 independent databases (each ~1–3 replicas), not a single centralised system.

### 1.3 Replication Factor Comparison

| Data Type | Fabric (4-org) | Fabric (27 EU orgs, single channel) | Status Quo (27 national registries) |
|---|---|---|---|
| Public GO records | 4× | 27× | 1× per country (27 isolated silos) |
| Private GO details | 2× (producer + issuer) | 2× | 1× per country |
| Private data hashes | 4× | 27× | N/A |
| Transaction history | 4× | 27× | 1× per country (audit log) |
| **Effective replication per asset** | **~4.5×** | **~27.5×** | **~1.5×** |

**Critical distinction:** The status quo already maintains 27 separate databases — Fabric does **not** multiply the number of infrastructure deployments from 1 to 27. What Fabric changes is that **each peer stores a full copy of all data on its channel**, including data originating from other countries. In the status quo, Germany's registry only stores German GOs; in a single-channel Fabric deployment, Germany's peer would store all 27 countries' GOs.

### 1.4 The N-Participant Full Replication Problem

The core data overhead concern with Fabric is not the number of organisations (which already mirrors the status quo) but that **every peer on a channel stores a complete copy of the entire ledger and world state for that channel**:

| What each peer stores | Size class | Scales with |
|---|---|---|
| **Full block ledger** (every committed transaction) | Large | Total transaction volume on the channel |
| **Full public world state** (CouchDB: all current public assets) | Large | Total number of live assets on the channel |
| **All private data hashes** (SHA-256 of every private entry) | Small | Total private data entries on the channel |
| **Private data** (raw, only for own collections) | Medium | Own organisation's private data only |

This means:
- **4 peers** → 4 full copies of the ledger and state
- **27 peers** (one per EU country) → 27 full copies
- **1,000 peers** (if every market participant ran a peer) → **1,000 full copies**

The replication factor scales **linearly with the number of peers on a channel**. Unlike the status quo — where each national registry only stores its own country's data — a naïve single-channel Fabric deployment forces every participant to store *everyone's* data. This is the fundamental scalability constraint that necessitates channel sharding (see Section 6.2).

---

## 2. Per-Asset Storage Overhead

### 2.1 Electricity GO Storage Breakdown

The following estimates are based on the v10.5.4 `ElectricityGO` and `ElectricityGOPrivateDetails` structs, including CEN-EN 16325 fields:

| Component | Centralised (PostgreSQL) | Fabric (per peer) | Overhead |
|---|---|---|---|
| **Public record** (12 fields) | ~320 bytes (row) | ~450 bytes (JSON in CouchDB) | 1.4× |
| **Private record** (9 fields) | included above | ~280 bytes (JSON in CouchDB) | — |
| **Private data hash** | N/A | 64 bytes (SHA-256 on ledger) | — |
| **Transaction envelope** | N/A | ~2,500 bytes (proposal + endorsements + signatures) | — |
| **Block header** | N/A | ~120 bytes (amortised per tx across block) | — |
| **CouchDB indexes** | ~100 bytes (B-tree) | ~200 bytes (JSON views) | 2× |
| **CouchDB revision metadata** | N/A | ~80 bytes (_id, _rev, ~attachments) | — |
| **Total per GO** | **~420 bytes** | **~3,694 bytes** | **8.8×** |

### 2.2 Per-Asset Storage by Entity Type

| Entity Type | Centralised | Fabric (per peer) | Overhead Factor |
|---|---|---|---|
| Electricity GO (eGO) | ~420 bytes | ~3,700 bytes | 8.8× |
| Hydrogen GO (hGO) | ~450 bytes | ~3,900 bytes | 8.7× |
| Biogas GO (bGO) | ~480 bytes | ~4,100 bytes | 8.5× |
| Metering Device | ~300 bytes | ~2,800 bytes | 9.3× |
| Cancellation Statement | ~350 bytes | ~3,200 bytes | 9.1× |
| Organization Registration | ~200 bytes | ~2,600 bytes | 13.0× |

### 2.3 Total Per-Asset Across All Peers

For a 4-peer deployment, total storage consumed per GO across the entire network:

| Entity | Status quo (1 national registry, 1 replica) | Fabric (4 peers) | Network Overhead |
|---|---|---|---|
| Electricity GO | 840 bytes | **14,800 bytes** (4 × 3,700) | **17.6×** |
| Hydrogen GO | 900 bytes | 15,600 bytes | 17.3× |
| Biogas GO | 960 bytes | 16,400 bytes | 17.1× |

For larger deployments, the overhead grows linearly with peer count:

| Number of peers | eGO total network storage | vs. single registry |
|---|---|---|
| 4 (pilot) | 14.8 KB | 17.6× |
| 27 (EU, single channel) | 99.9 KB | 119× |
| 100 (large market) | 370 KB | 440× |
| 1,000 (every participant as peer) | 3.7 MB | 4,400× |

This linear scaling is the reason why **not every market participant should run a full peer** — most participants should interact via gateway peers or light clients, and the network should be partitioned via channel sharding.

---

## 3. Storage Projections at Scale

### 3.1 Assumptions

Based on European GO market volumes (AIB statistics 2023):
- **Annual GO issuance:** ~700 million MWh → ~700 million GOs (at 1 MWh per GO)
- **Active devices:** ~50,000 registered metering points
- **Organizations:** ~500 producing companies across EU-27
- **Transfers per GO:** ~2.5 average (producer → trader → consumer)
- **Cancellations:** ~600 million per year

For a national pilot (e.g., Germany only):
- **Annual GO issuance:** ~50 million GOs
- **Devices:** ~5,000
- **Organizations:** ~50

### 3.2 Storage Volume Projections — National Pilot (1 Year)

| Data Category | Centralised | Fabric (4 peers) | Fabric (10 peers) |
|---|---|---|---|
| **GO records (50M)** | 21 GB | 185 GB | 370 GB |
| **Transfer transactions (125M)** | 12.5 GB | 312 GB | 625 GB |
| **Cancellation records (40M)** | 14 GB | 128 GB | 256 GB |
| **Device registrations (5K)** | 1.5 MB | 14 MB | 28 MB |
| **Block headers + metadata** | N/A | 15 GB | 30 GB |
| **CouchDB indexes** | 5 GB | 40 GB | 80 GB |
| **CouchDB compaction overhead** | N/A | 30 GB | 60 GB |
| **Total (1 year)** | **~53 GB** | **~710 GB** | **~1.4 TB** |

**Fabric requires ~13× more storage than a centralised system for a national pilot.**

### 3.3 Storage Volume Projections — EU-Wide (1 Year)

| Data Category | Status quo (27 registries, aggregate) | Fabric (27 peers, single channel) | Fabric (27 channels, ~5 peers each) |
|---|---|---|---|
| **GO records (700M)** | ~294 GB (distributed across 27 DBs) | ~70 TB (each peer stores all) | ~5.2 TB |
| **Transfer transactions (1.75B)** | ~175 GB | ~118 TB | ~8.7 TB |
| **Cancellation records (600M)** | ~210 GB | ~52 TB | ~3.9 TB |
| **Block storage** | N/A | ~5 TB | ~370 GB |
| **CouchDB indexes + compaction** | ~70 GB | ~19 TB | ~1.4 TB |
| **Total (1 year)** | **~750 GB** | **~264 TB** | **~19.6 TB** |

**Single-channel EU deployment:** 264 TB across the network — each of the 27 peers stores ~9.8 TB. This is **352× the aggregate status quo** or ~9.8 TB per peer vs. ~28 GB per national registry today.

**Sharded EU deployment (recommended):** 19.6 TB across the network — each peer stores only its jurisdiction's data (~725 GB per peer). This is **~26× the aggregate status quo**, or **~26× per peer** compared to ~28 GB per national registry today.

> **Key insight:** The status quo already distributes ~750 GB across 27 independent systems (~28 GB each). Fabric's overhead per organisation comes from (a) the per-asset storage overhead of blockchain structures (8.8×, see Section 2.1) and (b) the replication factor within each channel (~5× with sharding). The total per-organisation overhead with sharding is therefore ~725 GB vs. ~28 GB — roughly 26×, not 350×.

---

## 4. Compute Overhead

### 4.1 Transaction Processing Cost

| Processing Step | Centralised | Fabric | Overhead |
|---|---|---|---|
| **Input validation** | 1× (server) | 3× (3-of-4 endorsing peers each execute chaincode) | 3× |
| **State read** | 1× (DB query) | 3× (each endorsing peer queries CouchDB) | 3× |
| **State write** | 1× (DB write) | 3× endorsement + 1× orderer + 4× commit | 8× |
| **Cryptographic operations** | 1× (TLS) | 3× endorsement signatures + 1× orderer sign + 4× block validation | 8× |
| **Consensus** | N/A | Raft: 3 round-trips per block (leader → followers) | N/A |
| **Private data dissemination** | N/A | Gossip to collection members + hash to all | 2× |

### 4.2 CPU Utilisation Model

Based on the Caliper benchmarks (16 vCPU VM):
- At **50 TPS writes** (RegisterDevice): ~40% CPU across all 12 Docker containers
- At **2,000 TPS reads** (GetVersion): ~25% CPU
- At **500 TPS range queries** (ListDevices): ~60% CPU

Equivalent centralised system estimate for the same throughput:
- 50 TPS writes: ~5% CPU on a 4-vCPU PostgreSQL server
- 2,000 TPS reads: ~10% CPU
- 500 TPS range queries: ~15% CPU

**Fabric consumes approximately 4–8× more CPU per transaction than a centralised database**, primarily due to redundant chaincode execution during endorsement and block validation across multiple peers.

### 4.3 Network Bandwidth

| Operation | Centralised | Fabric (4-org) |
|---|---|---|
| Write transaction | ~1 KB (client → server) | ~8 KB (proposal to 3 endorsers + orderer + gossip to 4 peers) |
| Read query | ~0.5 KB | ~2 KB (via gateway + service discovery) |
| Block gossip | N/A | ~12 KB per block × 4 peers (continuous) |
| Private data gossip | N/A | ~2 KB per private entry × 2 collection members |

**Network overhead: 4–10× per transaction, plus continuous gossip background traffic.**

---

## 5. What Fabric Provides in Return

The overhead exists because Fabric delivers properties that a centralised architecture cannot:

| Property | Centralised | Fabric |
|---|---|---|
| **Trust model** | Single trusted operator | Multi-party consensus — no single point of trust |
| **Tamper evidence** | Database admin can modify records | Immutable blockchain ledger with hash chains |
| **Organisational data sovereignty** | All data controlled by operator | Private data collections allow each org to control its own data |
| **Multi-party endorsement** | Server validates unilaterally | 3-of-4 organisations must independently agree |
| **Auditability** | Audit logs can be modified | Transaction history is cryptographically sealed |
| **Censorship resistance** | Operator can block participants | No single org can prevent others from transacting |
| **Decentralised governance** | Operator sets all rules | Channel configuration requires multi-org consensus |

### 5.1 Cost-Benefit Assessment

| Deployment Scale | Storage Overhead (vs. status quo) | Compute Overhead | Justification |
|---|---|---|---|
| **Pilot (4 orgs, national)** | ~13× per registry | 4–8× | Acceptable for trust establishment and regulatory compliance demonstration |
| **Production (27 orgs, EU, single channel)** | ~350× aggregate / ~9.8 TB per peer | 8–12× | **Not viable** — requires channel sharding |
| **Production (27 orgs, EU, sharded)** | ~26× per registry (~725 GB/peer) | 4–8× | Viable with per-jurisdiction channels |
| **Hypothetical (1,000 peers, single channel)** | ~4,400× aggregate | 20–30× | **Not viable** — demonstrates why not every participant should run a full peer |

---

## 6. Mitigation Strategies

### 6.1 Implemented (v4.0–v10.5.4)

| Strategy | ADR | Impact |
|---|---|---|
| **Pagination** | ADR-006 | Reduces query response sizes from unbounded to max 200 records |
| **blockToLive** | ADR-010 | Private data purged after ~70 days, limiting CouchDB growth |
| **CQRS events** | ADR-016 | Off-chain read models reduce CouchDB query load |
| **Hash-based IDs** | ADR-001 | Eliminates shared counter key — no write amplification from retries |
| **Per-carrier channel split** (electricity-de / hydrogen-de) | ADR-025 + v10.x | Issuance, transfer, cancellation traffic for each carrier stays on its own ledger; conversion uses a 3-phase lock-mint-finalize protocol with a hash-bound `ConversionLock` receipt (v10.5.3+ stores `TxID` for hash reproducibility) |

### 6.2 Channel Sharding (ADR-025) — Primary Mitigation

The most critical mitigation for the N-participant replication problem is **multi-channel sharding by jurisdiction**. Instead of placing all participants on a single channel, the network is partitioned so that each country operates its own Fabric channel:

| Design Parameter | Single Channel (naïve) | Sharded (per-jurisdiction) |
|---|---|---|
| **Channels** | 1 | 27 (one per EU member state) + 1 cross-border |
| **Peers per channel** | 27 (all EU orgs) | ~3–5 (national body + auditor + 1–3 market participants) |
| **Data each peer stores** | All 700M EU-wide GOs | Only its country's GOs (~26M avg.) |
| **Storage per peer** | ~9.8 TB/year | ~725 GB/year |
| **Cross-border transfers** | Direct (same channel) | Routed via cross-border relay channel with atomic swaps |

**How it works:**
1. Each national channel handles issuance, transfer, and cancellation for that country's GOs.
2. A **cross-border channel** connects all 27 national bodies and the AIB Hub equivalent. Cross-border transfers are executed as atomic two-phase operations: burn on source channel, mint on destination channel, with proof anchored on the cross-border channel.
3. Each organisation only joins the channels it needs — a German producer only needs the DE channel; a pan-European trader joins the cross-border channel plus relevant national channels.

**Result:** Replication factor drops from 27× to ~5× per channel, and each peer's storage drops from ~9.8 TB to ~725 GB per year.

### 6.3 Participant Tiers — Not Everyone Needs a Full Peer

The 1,000-participant scenario (every market actor running a full peer) is a design antipattern, not an inevitability. Fabric supports multiple participation modes:

| Tier | Role | Infrastructure | Sees full ledger? |
|---|---|---|---|
| **Full peer** | National competent body, auditor | Peer + CouchDB + orderer | Yes (for its channels) |
| **Endorsing org** | Large utility, TSO | Peer + CouchDB | Yes (for its channels) |
| **Gateway client** | Producer, trader, consumer | Client SDK only (connects to org's peer) | No — queries via peer |
| **Light client** | Small producer, end consumer | REST API to org's gateway | No |

In a well-designed deployment, full peers number in the **tens** (national bodies + major market participants), not the thousands. The vast majority of market participants interact through gateway clients or REST APIs, without replicating any ledger data.

### 6.4 Additional Planned Mitigations (post-v10.5)

| Strategy | Impact | Storage Reduction |
|---|---|---|
| **Off-chain data archival** | Move aged GO records to Elasticsearch/S3 | 3–5× |
| **CouchDB compaction tuning** | Reduce revision history overhead | 1.5–2× |
| **State pruning (LevelDB for non-query peers)** | Eliminate CouchDB JSON overhead | 2× for write-only peers |
| **Selective peer deployment** | Not every org needs a full peer (light clients) | Variable |

### 6.5 Architectural Alternatives Considered

| Alternative | Trade-off |
|---|---|
| **LevelDB instead of CouchDB** | Eliminates JSON overhead but loses rich queries — requires off-chain indexing for all queries |
| **Single-peer-per-org** (no redundancy) | Reduces replication but eliminates fault tolerance within org |
| **Prunable block store** | HLF does not natively support ledger pruning; requires custom external tooling |
| **Off-chain private data** | Store private data externally (e.g., IPFS) with only hashes on-chain — reduces per-peer storage but adds integration complexity |

---

## 7. Throughput–Replication Trade-off and Regulatory Fit

The v10.5.4 cross-channel conversion benchmark (Section 7 of the main paper) measured a per-phase chaincode throughput of **6.8–8.5 TPS** for `LockGOForConversion`, `MintFromConversion`, and `FinalizeLock`, and an end-to-end conversion rate of **~0.29 cycles/s** when the three phases are executed sequentially with the required inter-phase block-commit waits. This is roughly an order of magnitude below the single-phase write throughput of `oracle:PublishOracleData` (41.9 TPS) on the same hardware, and two orders of magnitude below the read throughput (40–50 TPS via Caliper).

This throughput drop is **not a regression but the direct cost of the multi-channel architecture** described in Section 6.2 — and it is the right cost to pay. The cross-channel conversion has to atomically (a) lock and consume an electricity GO on `electricity-de`, (b) mint a hydrogen GO on `hydrogen-de` whose receipt hash is bound to the lock created on the source channel, and (c) finalize the lock on the source channel once the destination mint is committed. Each phase is a separate Fabric transaction on a separate ledger with its own endorsement, ordering, and commit pipeline; the floor on end-to-end latency is therefore set by *two* block intervals plus the cost of carrying a tamper-evident receipt across channels, not by chaincode CPU time.

In exchange, the data-duplication picture inside each channel is dramatically better than the single-channel alternative. With per-carrier (and ultimately per-jurisdiction) channels, every peer only replicates the GOs of the carrier and country it actually participates in: an `eissuer`/`eproducer1`/`ebuyer1` peer never sees hydrogen GOs, and an `hissuer`/`hproducer1`/`hbuyer1` peer never sees electricity GOs. The replication factor inside each channel collapses from ≈27 (all EU orgs on one ledger) to ≈3–5 (the issuer plus the active market participants of that carrier in that country), and per-peer storage drops from the ~9.8 TB/year of the single-channel naïve case to ~725 GB/year (Section 3.3).

Crucially, this architecture is **isomorphic to the real-world regulatory landscape**. Under the Renewable Energy Directive each EU member state designates *one* competent issuing body per energy carrier — Pronovo for Swiss electricity, UBA / HKNR for German electricity, the dena Biogasregister for German biogas, and so on. There is no central pan-European registry; cross-border movements are reconciled asynchronously via the AIB Hub. Mapping each (country, carrier) pair to its own Fabric channel, with the national issuing body of that pair as the only org that holds a private data collection containing **all** GO details on that channel, mirrors this institutional structure exactly:

- The issuing body sees every record on its channel — which it must, because it is legally responsible for issuance, transfer validation, and cancellation auditing for that carrier in that jurisdiction.
- Producers, traders, and buyers on the same channel see only their own private collections; their counterparties' confidential data is protected by PDC membership rules even though they share the public ledger.
- Foreign issuing bodies and unrelated participants are not on the channel at all and therefore store nothing — the same data-minimisation property that today is achieved by running 27 unconnected national registries, but now with cryptographic atomicity, tamper-evident audit trails, and a well-defined cross-channel conversion protocol layered on top.

In other words, the slower conversion path is **the price of *not* asking 27 national bodies to mirror each other's data**, and the privileged read access of the issuer on each channel is **the on-chain reflection of an off-chain regulatory mandate that already exists**. A single-channel design would buy ~10–100× higher conversion throughput, but only by collapsing 27 sovereign registries into one shared ledger that no national competent body has the authority to operate alone — which is precisely why the status quo is federated rather than centralised in the first place.

---

## 8. Conclusion

A Hyperledger Fabric-based GO platform produces **approximately 9× more data per asset** than a traditional database (due to transaction envelopes, block headers, CouchDB JSON overhead, and private data hashes). In a 4-organisation national pilot, this translates to ~13× total storage overhead compared to a single national registry.

**At EU scale, the correct baseline for comparison is not a single database but the 27 separate national registries that already exist today.** The status quo already distributes ~750 GB of GO data across 27 isolated systems (~28 GB per registry). Fabric does not multiply the number of infrastructure deployments — it changes what is stored at each node.

The critical scalability concern is **full ledger replication**: every peer on a Fabric channel stores a complete copy of all data on that channel. On a single channel with N peers, total network storage scales as N × per-peer size. This means:
- **27 peers (single channel):** each peer stores ~9.8 TB/year → ~264 TB aggregate (352× status quo)
- **1,000 peers (single channel):** ~4,400× status quo — clearly not viable

The solution is **architectural, not optimisation-based:**

1. **Channel sharding by jurisdiction** (ADR-025): Partition into 27 national channels + 1 cross-border channel. Each peer only stores its jurisdiction's data. Per-peer storage drops to ~725 GB/year (~26× the equivalent national registry — a manageable premium for cryptographic trust, tamper-evidence, and multi-party endorsement).

2. **Participant tiering:** Only national bodies and major market participants run full peers (tens of nodes, not thousands). Producers, traders, and consumers interact via gateway clients or REST APIs without replicating ledger data.

With both mitigations, the EU-wide network stores ~19.6 TB across all peers (~26× the status quo's ~750 GB aggregate), while delivering multi-party trust, tamper-evident audit trails, and organisational data sovereignty that the current fragmented registry system cannot provide.

---

*Analysis date: 2026-04-27. Based on v10.5.4 `golifecycle` chaincode structure, Caliper read benchmarks, peer-CLI single-phase and 3-phase cross-channel conversion benchmarks (10 cycles, 100% success, 6.8–8.5 TPS per phase), AIB 2023 GO market statistics, and CouchDB storage characteristics.*
