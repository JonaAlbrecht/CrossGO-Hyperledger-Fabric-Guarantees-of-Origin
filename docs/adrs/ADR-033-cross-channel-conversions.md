# ADR-033: Cross-Channel Carrier Conversions

**Status:** Accepted  
**Date:** 2026-04-25  
**Design Cycle:** v10.1  
**Supersedes:** In-channel conversions (v9.0)  
**Related:** ADR-030 (Multi-Channel Topology), ADR-031 (Cross-Channel Bridge), ADR-032 (Tri-Party Endorsement)

## Context

The v9.0 architecture operates on a single channel (`goplatformchannel`) where all four energy carriers (electricity, hydrogen, biogas, heating/cooling) coexist. Conversions between carriers (e.g., electricity → hydrogen via electrolysis) happen within the same channel using the `ConversionContractV10`.

However, ADR-030 established that **channel-per-carrier-per-region** topology offers significant benefits:
- **Storage reduction**: 13.5× improvement at EU scale
- **Structural privacy**: Producers can't access other carriers' data
- **Independent scaling**: Each carrier has its own ordering pipeline

The current single-channel conversion design conflicts with the multi-channel vision from v8.0. If we adopt separate channels per carrier, conversions must become **cross-channel operations** similar to cross-border bridge transfers.

### Current State (v9.0)
```
Channel: goplatformchannel
┌─────────────────────────────────────┐
│  eGO_123  →  ConvertEtoH  →  hGO_456 │
│  (same channel, same ledger)         │
└─────────────────────────────────────┘
```

### Desired State (v10.1)
```
Channel: electricity-de          Channel: hydrogen-de
┌─────────────────────┐          ┌─────────────────────┐
│  eGO_123 (active)   │          │                     │
│       ↓             │          │                     │
│  eGO_123 (LOCKED)   │ ------→  │  hGO_456 (active)   │
│       ↓             │  relay   │                     │
│  eGO_123 (CONSUMED) │          │                     │
└─────────────────────┘          └─────────────────────┘
```

## Decision

Extend the multi-channel architecture (ADR-030) to **all four energy carriers** and implement **cross-channel conversions** using a lock-mint-finalize protocol similar to the bridge transfer pattern (ADR-031).

### Channel Topology

Create **four carrier-specific channels** for the German region:

| Channel | Carrier | Joined by |
|---------|---------|-----------|
| `electricity-de` | Electricity GOs | Issuer, EProducer, Buyer |
| `hydrogen-de` | Hydrogen GOs | Issuer, HProducer, Buyer |
| `biogas-de` | Biogas GOs | Issuer, BiogasProducer, Buyer |
| `heating-de` | Heating/Cooling GOs | Issuer, HeatingProducer, Buyer |

**Organization membership**:
- **Issuer** (`issuer1MSP`): Joins **ALL** channels. Acts as trust anchor and cross-channel relay.
- **Producers**: Join only their carrier-specific channel(s). For example:
  - `eproducer1MSP` joins only `electricity-de`
  - `hproducer1MSP` joins only `hydrogen-de`
  - Multi-carrier producers (e.g., a biogas plant producing both biogas and electricity) join multiple channels
- **Buyer** (`buyer1MSP`): Joins **ALL** channels (needs to purchase and consume GOs of any type).

### Cross-Channel Conversion Protocol

Conversions follow a **3-phase lock-mint-finalize protocol** with **tri-party endorsement** (owner + source issuer + dest issuer):

#### Phase 1: Lock Source GO (Source Channel)

**Channel**: Source carrier channel (e.g., `electricity-de`)  
**Function**: `conversion:LockGOForConversion`  
**Endorsement**: Source issuer + GO owner (tri-party)

```json
{
  "goAssetID": "eGO_123",
  "destinationChannel": "hydrogen-de",
  "destinationCarrier": "hydrogen",
  "conversionEfficiency": 0.65,
  "ownerMSP": "eproducer1MSP"
}
```

**Actions**:
1. Verify GO ownership (owner MSP must match GO's private data owner)
2. Change GO status to `LOCKED_CONVERSION`
3. Create conversion lock record with SHA-256 hash:
   ```
   hash = SHA-256(lockID || goAssetID || sourceChannel || destinationChannel || 
                  destinationCarrier || conversionEfficiency || ownerMSP || txID)
   ```
4. Set tri-party state-based endorsement policy (source issuer + owner)
5. Emit `CONVERSION_LOCK_CREATED` lifecycle event

**Lock Record Structure**:
```go
type ConversionLock struct {
    LockID              string   // conversion_lock_<timestamp>_<suffix>
    GOAssetID           string   // eGO_123
    SourceChannel       string   // electricity-de
    SourceCarrier       string   // electricity
    DestinationChannel  string   // hydrogen-de
    DestinationCarrier  string   // hydrogen
    ConversionMethod    string   // electrolysis, fuel_cell, biogas_chp, heat_pump
    ConversionEfficiency float64 // 0.65 (65% efficient)
    OwnerMSP            string   // eproducer1MSP
    SourceIssuerMSP     string   // issuer1MSP
    LockReceiptHash     string   // SHA-256 cryptographic proof
    CreatedAt           int64    // Unix timestamp
    Status              string   // locked, approved, consumed, expired
}
```

#### Phase 2: Mint Destination GO (Destination Channel)

**Channel**: Destination carrier channel (e.g., `hydrogen-de`)  
**Function**: `conversion:MintFromConversion`  
**Endorsement**: Destination issuer only (lock receipt hash proves source consent)

**Input** (relayed by issuer from source channel):
```json
{
  "lockID": "conversion_lock_1745500800_001",
  "goAssetID": "eGO_123",
  "sourceChannel": "electricity-de",
  "sourceCarrier": "electricity",
  "destinationCarrier": "hydrogen",
  "conversionMethod": "electrolysis",
  "conversionEfficiency": 0.65,
  "sourceAmount": 100.0,
  "sourceAmountUnit": "MWh",
  "ownerMSP": "eproducer1MSP",
  "sourceIssuerMSP": "issuer1MSP",
  "lockReceiptHash": "9a3f2c1...",
  "txID": "source_txid"
}
```

**Actions**:
1. Verify lock receipt hash matches input (proves authenticity)
2. Check mint hasn't already occurred (idempotency: `conversion_mint_receipt_<lockReceiptHash>`)
3. Calculate destination amount:
   ```
   destAmount = sourceAmount * conversionEfficiency
   // Example: 100 MWh * 0.65 = 65 MWh → ~3,250 kg H2 (at 20 kWh/kg)
   ```
4. Create destination GO with:
   - Status: `ACTIVE`
   - Owner: Same MSP as source GO owner (`eproducer1MSP`)
   - Production method: Conversion from source carrier (e.g., `electrolysis_from_renewable_electricity`)
   - Emissions: Inherited from source GO (proportional to conversion efficiency)
   - Consumption declarations: Track source GO ID
5. Write mint receipt to prevent double-minting
6. Emit `CONVERSION_MINT_CREATED` lifecycle event

**Destination GO Structure** (hydrogen example):
```json
{
  "assetID": "hGO_456",
  "ownerID": "eproducer1MSP",
  "kilosproduced": 3250.0,
  "emissions": 32.5,
  "hydrogenProductionMethod": "electrolysis_from_renewable_electricity",
  "consumptionDeclarations": ["eGO_123"],
  "conversionLockID": "conversion_lock_1745500800_001",
  "creationDateTime": 1745500850
}
```

#### Phase 3: Finalize Lock (Source Channel)

**Channel**: Source carrier channel (e.g., `electricity-de`)  
**Function**: `conversion:FinalizeLock`  
**Endorsement**: Source issuer + GO owner (tri-party)

```json
{
  "lockID": "conversion_lock_1745500800_001",
  "mintedAssetID": "hGO_456",
  "destinationChannel": "hydrogen-de",
  "ownerMSP": "eproducer1MSP"
}
```

**Actions**:
1. Verify lock exists and status is `approved`
2. Change locked GO status to `CONSUMED`
3. Update lock status to `consumed`
4. Emit `CONVERSION_FINALIZED` lifecycle event

### Tri-Party Endorsement

Similar to bridge transfers (ADR-032), conversion locks require **three-party consent**:

1. **GO Owner**: Consent to convert their GO (producer who owns the source GO)
2. **Source Channel Issuer**: Authority approval to lock GO on source channel
3. **Destination Channel Issuer**: Authority approval to mint GO on destination channel

**Why tri-party?**
- Prevents issuers from unilaterally converting GOs without owner consent
- Ensures regulatory compliance (both source and dest jurisdictions/carriers agree)
- Cryptographic proof via lock receipt hash

**Implementation**:
- Phases 1 and 3 use state-based endorsement: `AddOrgs(RoleTypePeer, sourceIssuerMSP, ownerMSP)`
- Phase 2 uses lock receipt hash verification (already proves owner + source issuer consent)

### Supported Conversion Paths

| Source → Destination | Conversion Method | Typical Efficiency |
|---------------------|-------------------|-------------------|
| Electricity → Hydrogen | Electrolysis | 60–75% |
| Hydrogen → Electricity | Fuel Cell | 40–60% |
| Electricity → Heating/Cooling | Heat Pump | 300–400% (COP) |
| Biogas → Electricity | Biogas CHP | 35–45% |
| Biogas → Heating/Cooling | Biogas Boiler | 85–95% |
| Biogas → Hydrogen | Biogas Reforming | 70–80% |

**Note**: Heat pump efficiency (COP) can exceed 100% because it moves heat rather than generating it. The conversion protocol uses the declared efficiency factor directly without validation (issuer responsibility).

## Consequences

### Positive

✅ **Storage reduction**: Each producer peer stores only its carrier channel(s). At EU scale (27 regions × 4 carriers = 108 channels), an electricity-only producer stores ~1/108th of the total ledger.

✅ **Structural privacy**: Hydrogen producers physically cannot access electricity channel data (no block events, no world state queries, no gossip messages).

✅ **Independent scaling**: Each carrier has its own ordering pipeline. Electricity GO issuance doesn't contend with hydrogen conversions for ordering slots.

✅ **Regulatory alignment**: Per-carrier channels map to certification body boundaries (e.g., TÜV SÜD for electricity, DIN CERTCO for hydrogen).

✅ **Subsidy isolation**: Carrier-specific subsidy schemes (e.g., CfD for electricity, H2Global for hydrogen) are enforced at channel level, preventing cross-contamination.

✅ **Security**: Tri-party endorsement prevents unilateral conversions. Lock receipt hash provides cryptographic proof of consent across channels.

### Negative

❌ **Deployment complexity**: Chaincode must be deployed to **4 channels** instead of 1. Each channel needs:
- Separate chaincode approval and commit
- Carrier-specific collection configs
- Channel-specific role initialization

❌ **Cross-channel coordination**: Conversions require **3 transactions** (lock, mint, finalize) across 2 channels instead of 1 transaction on 1 channel.

❌ **Issuer relay responsibility**: The issuer must relay lock receipts between channels. If the issuer is offline, conversions cannot proceed (Phase 2 blocked).

❌ **Orderer resource overhead**: 4 channels = 4× orderer state machines. At EU scale (108 channels), orderer capacity planning becomes critical.

❌ **Backlog management complexity**: The current backlog design (v9.0) stores accumulated production data per producer. With cross-channel conversions, backlog data must be read from the **destination channel** during lock phase. This requires:
  - Cross-channel queries (not natively supported in Fabric)
  - OR: Move backlog management out of chaincode to off-chain database
  - OR: Remove backlog entirely and require 1:1 conversion (1 eGO → 1 hGO with exact efficiency calculation)

### Neutral

➡️ **Sample setup unchanged**: In the PoC deployment, there's only one issuer org (`issuer1MSP`) that joins all 4 channels. Multi-issuer scenarios (e.g., separate German and Dutch issuers) would require additional endorsement orchestration.

➡️ **Conversion efficiency**: The protocol accepts any efficiency factor declared by the producer. No on-chain validation of thermodynamic feasibility (e.g., could declare 10% or 1000% efficiency). This is acceptable if we trust the issuer to validate conversion methods during device certification.

## Implementation Plan

### Phase 1: Network Infrastructure (v10.1)

1. **Update `configtx.yaml`** to define 4 carrier-specific channels:
   - `electricity-de`, `hydrogen-de`, `biogas-de`, `heating-de`

2. **Update `network-up-v10.sh`** to create 4 genesis blocks and join orgs selectively

3. **Create 4 collection configs**:
   - `collection-config-electricity.json` (Issuer, EProducer, Buyer)
   - `collection-config-hydrogen.json` (Issuer, HProducer, Buyer)
   - `collection-config-biogas.json` (Issuer, BiogasProducer, Buyer)
   - `collection-config-heating.json` (Issuer, HeatingProducer, Buyer)

4. **Update `deploy-v10.sh`** to deploy chaincode to all 4 channels with carrier-specific configs

### Phase 2: Chaincode Implementation

1. **Create `conversion_crosschannel.go` contract** with functions:
   - `LockGOForConversion` (Phase 1: lock source GO)
   - `MintFromConversion` (Phase 2: mint dest GO)
   - `FinalizeLock` (Phase 3: finalize source lock)
   - `GetConversionLock` (query lock record)
   - `ListConversionLocks` (list all locks)

2. **Add `ConversionLock` struct** to `assets/types.go`

3. **Implement tri-party endorsement** using `statebased` package (similar to bridge.go)

4. **Update `util/helpers.go`** with conversion-specific helpers:
   - `CalculateDestinationAmount(sourceAmount, efficiency, sourceUnit, destUnit)`
   - `ValidateConversionPath(sourceCarrier, destCarrier, method)`

### Phase 3: Backend API

1. **Create `application/backend/src/routes/conversion-crosschannel.ts`** with routes:
   - `POST /api/conversion/lock` — Lock source GO for conversion
   - `POST /api/conversion/mint` — Mint destination GO
   - `POST /api/conversion/finalize` — Finalize conversion
   - `GET /api/conversion/locks` — List conversion locks
   - `GET /api/conversion/locks/:lockID` — Get specific lock

2. **Modify Fabric Gateway connection helper** to support multi-channel operations:
   ```typescript
   async function connectToChannel(channelName: string, identity: Identity) {
     const gateway = connect({...});
     const network = gateway.getNetwork(channelName);
     return network.getContract('golifecycle');
   }
   ```

3. **Implement issuer relay service** for Phase 2 (read lock from source channel, submit mint to dest channel)

### Phase 4: Frontend UI

1. **Modify `ConversionsPage.tsx`**:
   - Add "Source Channel" and "Destination Channel" dropdowns
   - Show conversion efficiency input
   - Add 3-phase progress indicator (lock → mint → finalize)
   - Display tri-party endorsement warning

2. **Add conversion lock status badges** (locked, approved, consumed, expired)

3. **Create "Approve Conversion" interface** for destination issuer

### Phase 5: Testing & Documentation

1. **Integration tests** for cross-channel conversion flow
2. **Update deployment guide** with 4-channel setup instructions
3. **Document backlog deprecation** (or cross-channel backlog design if kept)

## Alternatives Considered

### Alternative 1: Keep Single Channel, Add Carrier Tags

**Approach**: Stay on one channel but add `carrier` tags to all GOs. Use access control to filter queries.

**Rejected because**:
- Does not solve storage explosion (all peers still replicate all carriers)
- Privacy leakage through block events (carriers can see each other's transaction metadata)
- No independent scaling per carrier

### Alternative 2: Cross-Channel Backlog Queries

**Approach**: During lock phase, read backlog data from destination channel via cross-channel query.

**Rejected because**:
- Fabric does not natively support cross-channel queries from chaincode
- Would require custom relay service or Fabric Private Data Collections for cross-channel communication
- Adds significant complexity for marginal value (backlog is optional)

### Alternative 3: Embed Full Source GO in Lock Receipt

**Approach**: Include all source GO data in the lock receipt hash instead of just metadata.

**Rejected because**:
- Increases lock receipt size (block bloat)
- Redundant data (source GO already exists on source channel)
- Doesn't improve security (hash is already cryptographically binding)

## Related ADRs

- **ADR-030**: Multi-Channel Topology — Establishes channel-per-carrier pattern
- **ADR-031**: Cross-Channel Bridge — Lock-mint-finalize protocol for cross-border transfers
- **ADR-032**: Tri-Party Endorsement — Owner + issuer consent requirement for locks
- **ADR-024**: Lifecycle Events — Event schema for conversion lock/mint/finalize

## References

- RED III Article 19: Guarantee of Origin schemes per energy carrier
- AIB Hub EECS Rules: Conversion rules for multi-energy carriers
- IEA Hydrogen Report 2023: Electrolysis efficiency benchmarks (60–75%)
- IPCC AR6 WGIII: Heat pump COP ranges (300–400%)

## Appendix A: Lock Receipt Hash Calculation

```go
func GenerateConversionLockReceipt(lock *ConversionLock, txID string) (string, error) {
    data := fmt.Sprintf("%s||%s||%s||%s||%s||%.6f||%s||%s",
        lock.LockID,
        lock.GOAssetID,
        lock.SourceChannel,
        lock.DestinationChannel,
        lock.DestinationCarrier,
        lock.ConversionEfficiency,
        lock.OwnerMSP,
        txID,
    )
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:]), nil
}
```

## Appendix B: Sample Deployment Topology

```
┌──────────────────────────────────────────────────────────────┐
│                 Shared Raft Orderer Cluster                  │
│  orderer1.go-platform.com:7050                               │
│  orderer2.go-platform.com:8050                               │
│  orderer3.go-platform.com:9050                               │
│  orderer4.go-platform.com:10050                              │
└──────────────────────────────────────────────────────────────┘
           ↓                ↓                ↓                ↓
    electricity-de    hydrogen-de      biogas-de       heating-de
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ issuer1      │  │ issuer1      │  │ issuer1      │  │ issuer1      │
│ eproducer1   │  │ hproducer1   │  │ bgproducer1  │  │ htproducer1  │
│ buyer1       │  │ buyer1       │  │ buyer1       │  │ buyer1       │
└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
```

**Cross-carrier producer example**:
- A biogas plant that produces biogas AND converts biogas → electricity would join:
  - `biogas-de` (to issue biogas GOs)
  - `electricity-de` (to mint electricity GOs from conversion)
  
  This is implemented by having the producer org's peers join both channels.

## Appendix C: Migration from v9.0

**Breaking changes**:
1. Single channel (`goplatformchannel`) → 4 carrier-specific channels
2. In-channel conversions → Cross-channel 3-phase protocol
3. Backlog-based conversions → Lock-mint-finalize with explicit efficiency

**Migration steps**:
1. Export all active GOs from v9.0 channel (public + private data)
2. Split GOs by carrier type (electricity, hydrogen, biogas, heating/cooling)
3. Import into corresponding v10.1 carrier channels
4. Update all GO statuses to `ACTIVE`
5. Discard backlog data (or migrate to off-chain database if needed)

**Data loss**: Conversion history prior to v10.1 will not be preserved in the lock-mint-finalize format. Historical conversions can be archived off-chain.

---

**Approval**: [Pending signature from issuer1MSP, eproducer1MSP, hproducer1MSP, buyer1MSP]
