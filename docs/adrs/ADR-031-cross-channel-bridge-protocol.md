# ADR-031: Cross-Channel Bridge Protocol — 3-Phase Lock-Mint-Finalize

**Status:** Accepted  
**Date:** 2026-04-09  
**Design Cycle:** v8.0  
**Depends on:** ADR-030 (multi-channel topology)

## Context

With ADR-030's multi-channel topology, electricity and hydrogen GOs reside on separate channels. The core GO platform use case — **conversion issuance** (consuming electricity GOs to produce hydrogen GOs) — requires a mechanism to atomically transfer value between channels.

The v7.0 bridge contract (ADR-024) operates within a single channel and stores records about external registry transfers. It does not actually move assets between Fabric channels. For v8.0, a true cross-channel bridge protocol is needed where:

1. A GO is locked on the source channel (cannot be transferred, cancelled, or double-spent)
2. A corresponding GO is minted on the destination channel (inheriting provenance metadata)
3. The lock is finalized once the mint is confirmed (creating a permanent audit trail)

Hyperledger Fabric does not support native cross-channel atomic transactions. Each channel has an independent ledger and world state. The protocol must provide application-level atomicity through a trusted relay pattern.

## Decision

Implement a **3-phase cross-channel bridge protocol** with the issuer as a trusted relay:

### Phase 1: LockGO (source channel)

- Called on the source channel (e.g., `electricity-de`)
- The issuer invokes `bridge:LockGO` with the GO asset ID and destination channel
- The GO's status transitions from `active` → `locked` (new status, distinct from `transferred`)
- A `CrossChannelLock` record is created containing:
  - Lock ID, GO asset ID, source/destination channel names
  - **Lock receipt hash**: `SHA-256(lockID || goAssetID || sourceChannel || destinationChannel || txID)`
  - Timestamp, initiator MSP
- A `BRIDGE_GO_LOCKED` lifecycle event is emitted

### Phase 2: MintFromBridge (destination channel)

- Called on the destination channel (e.g., `hydrogen-de`)
- The issuer relays the lock receipt from Phase 1 to `bridge:MintFromBridge`
- The chaincode validates:
  - The lock receipt hash is non-empty
  - No previous mint exists for the same lock receipt hash (idempotency guard via `mint_receipt_<hash>` key)
  - CEN-EN 16325 fields are valid (ADR-018)
- A new GO is minted on the destination channel (e.g., a hydrogen GO)
- A `BridgeMint` record is created linking the new asset to the source lock
- A `BRIDGE_GO_MINTED` lifecycle event is emitted

### Phase 3: FinalizeLock (source channel)

- Called on the source channel (e.g., `electricity-de`)
- The issuer confirms the mint by invoking `bridge:FinalizeLock(lockID, mintedAssetID)`
- The lock status transitions from `locked` → `bridged`
- The GO status transitions from `locked` → `bridged`
- A `BRIDGE_LOCK_FINALIZED` lifecycle event is emitted

### Trust Model

The **issuer acts as the trusted relay** between channels. This is consistent with the issuer's existing role as the trust anchor for GO lifecycle operations. The issuer is the only organization with peers on both channels (ADR-030), making it the natural relay.

**Why the issuer relay is acceptable:**
- The issuer already has the authority to create, cancel, and convert GOs
- The issuer is a regulated national authority (e.g., UBA in Germany) with legal accountability
- Cross-channel relay adds no additional trust assumption beyond what already exists

**Lock receipt hash as cryptographic link:**
- The hash `SHA-256(lockID || goAssetID || sourceChannel || destinationChannel || txID)` provides a unique, tamper-evident link between the lock on the source channel and the mint on the destination channel
- An auditor can verify the correspondence by computing the hash from known parameters
- The idempotency guard (`mint_receipt_<hash>`) prevents double-minting even if the relay is invoked multiple times

### New GO Status Values

Two new lifecycle statuses are added to the tombstone pattern (ADR-007):
- `locked` — GO is locked for a pending cross-channel bridge transfer. Cannot be transferred, cancelled, or re-locked.
- `bridged` — GO has been successfully bridged to another channel. Terminal state (like `cancelled`). The GO remains on the source channel's ledger as an audit trail entry.

## Consequences

### Positive

- **Dual-ledger audit trail**: Both the source and destination channels maintain independent records of the bridge transfer. An auditor on the source channel sees the `CrossChannelLock` record; an auditor on the destination channel sees the `BridgeMint` record. Together they form a verifiable 6th verification layer (complementing the existing 5 layers from v7.0).
- **Idempotency**: The `mint_receipt_<hash>` guard key prevents double-minting even under network partitions or relay retries.
- **Backward compatibility**: The v7.0 `ExportGO`/`ImportGO`/`ConfirmExport` functions are retained for external registry interoperability (AIB hub pattern).

### Negative

- **Non-atomic**: The protocol is eventually consistent, not atomically consistent. Between Phase 1 (lock) and Phase 3 (finalize), the GO is in a `locked` state. If the relay fails between phases, manual intervention is needed. A future improvement could add timeout-based automatic unlock (HTLC pattern).
- **Issuer as single point of relay**: If the issuer's relay service is unavailable, cross-channel bridges stall. Mitigation: the lock remains valid and can be finalized when the relay recovers.
- **No cryptographic proof of foreign channel state**: The destination channel trusts the issuer's claim that a lock exists on the source channel. A stronger model would use SPV-style light client proofs, but Fabric does not natively support cross-channel state verification.

### Neutral

- **Same chaincode binary**: The `LockGO`, `MintFromBridge`, and `FinalizeLock` functions are part of the same `golifecycle` chaincode deployed to both channels. The chaincode detects its channel context via `ctx.GetStub().GetChannelID()`.

## Related

- **ADR-030**: Multi-channel topology enabling this protocol
- **ADR-024**: Legacy single-channel bridge (external registry pattern)
- **ADR-007**: Tombstone lifecycle pattern (extended with `locked` and `bridged` statuses)
- **ADR-009/017**: Quantity commitment scheme (applied to bridged mints)
- `v8_architecture.mmd`: Mermaid diagram of the 3-phase bridge protocol
