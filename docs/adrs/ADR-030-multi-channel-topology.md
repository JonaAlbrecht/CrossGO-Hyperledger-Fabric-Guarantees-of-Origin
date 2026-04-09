# ADR-030: Multi-Channel Topology — Channel per Carrier per Region

**Status:** Accepted  
**Date:** 2026-04-09  
**Design Cycle:** v8.0  
**Supersedes:** Single-channel architecture (v1.0–v7.0)

## Context

The v7.0 architecture operates on a single Fabric channel (`goplatformchannel`) where all four organizations (issuer, electricity producer, hydrogen producer, buyer) share the complete ledger. At EU production scale (27 member states, 700M+ GOs/year), this creates three fundamental problems:

1. **Storage explosion**: Every peer replicates the entire ledger. With 27 national peers on one channel, each peer stores ~9.8 TB/year — a 352× overhead vs. the current centralized status quo (~28 GB per national registry). See `DATA_DUPLICATION_CONSIDERATIONS.md` for the full analysis.

2. **Cross-carrier privacy leakage**: Electricity producers see hydrogen transaction metadata (and vice versa) through shared block events and world state, even with private data collections. The `BRIDGE_EXPORT_INITIATED` events from ADR-024 expose cross-border trading patterns to all channel subscribers.

3. **Write contention**: All carrier types compete for the same ordering pipeline. At peak solar production hours, electricity GO issuance contends with hydrogen GO transfers for ordering slots in the same BatchTimeout window.

## Decision

Adopt a **channel-per-carrier-per-region** topology where each energy carrier in each jurisdiction operates on a dedicated Fabric channel:

- `electricity-de` — German electricity GOs
- `hydrogen-de` — German hydrogen GOs
- Future: `biogas-de`, `electricity-nl`, `hydrogen-nl`, etc.

**Selective organization membership per channel:**
- **Issuer** (`issuer1MSP`): Joins ALL channels. Acts as trust anchor and cross-channel relay.
- **Electricity Producer** (`eproducer1MSP`): Joins only `electricity-de`.
- **Hydrogen Producer** (`hproducer1MSP`): Joins only `hydrogen-de`.
- **Buyer** (`buyer1MSP`): Joins ALL channels (needs to purchase GOs of any type).

**Shared Raft orderer cluster**: All 4 orderers serve both channels. Orderer nodes are carrier-agnostic — they order transactions without accessing application data.

## Consequences

### Positive

- **Storage reduction**: Each peer stores only its channels' data. Projected reduction from ~9.8 TB/peer/year to ~725 GB/peer/year (13.5× improvement at EU scale). See `DATA_DUPLICATION_CONSIDERATIONS.md` §4.
- **Structural privacy**: HProducer physically cannot access the `electricity-de` channel. No block events, no world state queries, no gossip messages. This is stronger than any application-level access control.
- **Independent scaling**: Each channel has its own ordering pipeline. Electricity and hydrogen GOs no longer compete for ordering slots. Adding a new carrier = adding a new channel, zero impact on existing channels.
- **Regulatory alignment**: Per-jurisdiction channels naturally map to national issuing body authority boundaries (EU RED III Article 19).

### Negative

- **Deployment complexity**: Chaincode must be installed, approved, and committed on each channel separately. Collection configs are channel-specific (`collection-config-electricity.json` vs `collection-config-hydrogen.json`).
- **Cross-channel operations require a relay**: The issuer must relay data between channels for conversion (electricity → hydrogen). See ADR-031 for the cross-channel bridge protocol.
- **Orderer resource overhead**: Each channel maintains independent Raft consensus. 2 channels = 2× orderer state machines. At 27+ channels (EU scale), orderer capacity planning becomes non-trivial.

### Neutral

- **Same chaincode binary**: The `golifecycle` chaincode is deployed identically to both channels. No code changes needed for channel awareness — each channel's chaincode instance operates independently on its own world state.
- **Collection configs differ only in member policies**: The collection structure (1 public + N private) is unchanged; only the `policy` fields reference different org MSPs per channel.

## Related

- **ADR-031**: Cross-channel bridge protocol (LockGO → MintFromBridge → FinalizeLock)
- **ADR-024**: Legacy single-channel bridge (retained for backward compatibility and external registry interoperability)
- `DATA_DUPLICATION_CONSIDERATIONS.md`: Full storage overhead analysis motivating this decision
- `ARCHITECTURE_CRITIQUE_v7.md` §1.2.3, §2.2.2, §4.2.2: Identified the core limitations
