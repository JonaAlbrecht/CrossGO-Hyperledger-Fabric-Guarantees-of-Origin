# Changelog — v4.0 (Hardening Release)

**Date:** 2026-04-05
**Chaincode:** `golifecycle` v4.0
**Fabric:** 2.5.12 | CA 1.5.17 | CouchDB 3.3.3

---

## Summary

v4.0 addresses the **scalability**, **privacy**, and **auditability** gaps identified in `ARCHITECTURE_CRITIQUE.md`. Five new ADRs (006–010) drive the changes, following agency-blueprint principles: Performance First (pagination), Explicit > Implicit (tombstone states), Security by Design (timestamp drift, commitment schemes, data retention).

---

## ADR-006 — Cursor-Based Pagination

**Problem:** `GetCurrentEGOsList`, `GetCurrentHGOsList`, and `ListDevices` loaded all records into memory. At 930+ records, `ListDevices` took 16 seconds and would OOM at scale.

**Solution:**
- Added `GetCurrentEGOsListPaginated(pageSize, bookmark)` and `GetCurrentHGOsListPaginated(pageSize, bookmark)` in `query.go`
- Added `ListDevicesPaginated(pageSize, bookmark)` in `device_mgmt.go`
- Default page size: 50, max: 200 (configurable constants)
- Returns `PaginatedResult` struct with `records`, `bookmark`, `count`
- Original unpaginated functions retained for backward compatibility, now filter Active GOs only

**Files changed:** `query.go`, `device_mgmt.go`

---

## ADR-007 — Tombstone Pattern (Soft Delete)

**Problem:** `DeleteEGOFromLedger` / `DeleteHGOFromLedger` called `DelState` + `DelPrivateData`, permanently removing records. This broke audit trails — a cancelled GO had no on-chain evidence it ever existed.

**Solution:**
- Added `Status` field to `ElectricityGO` and `GreenHydrogenGO` public structs: `"active"`, `"cancelled"`, `"transferred"`
- `DeleteEGOFromLedger` / `DeleteHGOFromLedger` now update Status to `"cancelled"` instead of calling `DelState`
- Added `MarkEGOTransferred` / `MarkHGOTransferred` for transfer tombstoning
- All GO creation paths (issuance, conversion, split remainder) set `Status = "active"`
- Query list functions filter out `cancelled`/`transferred` GOs
- New CouchDB composite index on `[Status, GOType]`
- Private data is retained for audit inspection

**Files changed:** `electricity_go.go`, `hydrogen_go.go`, `counter.go` (constants), `split.go`, `issuance.go`, `cancellation.go`, `conversion.go`, `query.go`, `indexStatus.json`

---

## ADR-008 — Timestamp Drift Guard

**Problem:** Proposal timestamps are controlled by the submitting client. A malicious client could backdate a GO creation to circumvent expiry checks.

**Solution:**
- `GetTimestamp()` in `validate.go` now compares the transaction proposal timestamp against `time.Now().Unix()`
- Rejects if |drift| > `MaxTimestampDrift` (300 seconds / 5 minutes)
- Constant defined in `counter.go` for easy tuning

**Files changed:** `validate.go`, `counter.go`

---

## ADR-009 — Hash Commitment for Selective Disclosure

**Problem:** Quantity data (AmountMWh, Kilosproduced) lives only in private data collections. Verifiers without collection access cannot validate any production claims.

**Solution:**
- Added `QuantityCommitment` (SHA-256 hash) to public GO structs
- Added `CommitmentSalt` to private detail structs
- `GenerateCommitment(ctx, quantity)` returns (commitment, salt) using `SHA-256(quantity || salt)` where salt is derived from txID
- `VerifyCommitment(quantity, salt, expectedCommitment)` for off-chain verification
- `VerifyQuantityCommitment` query function allows on-chain verification
- Producers can selectively disclose their quantity to specific verifiers by revealing the salt

**Files changed:** `electricity_go.go`, `hydrogen_go.go`, `counter.go`, `issuance.go`, `query.go`

---

## ADR-010 — CouchDB Hardening / blockToLive

**Problem:** All private data collections had `blockToLive: 0` (never purge), causing unbounded CouchDB growth. No data retention policy.

**Solution:**
- Set `blockToLive: 1000000` (~1M blocks) for all org-specific private data collections
- `publicGOcollection` retains `blockToLive: 0` (shared data must persist)
- At ~10 blocks/minute, this provides ~70 days of private data retention
- Tunable per deployment requirements

**Files changed:** `collection-config.json`

---

## Files Modified (v4.0)

| File | Changes |
|------|---------|
| `chaincode/assets/electricity_go.go` | +Status, +QuantityCommitment fields |
| `chaincode/assets/hydrogen_go.go` | +Status, +QuantityCommitment, +CommitmentSalt fields |
| `chaincode/assets/counter.go` | +GOStatus constants, +MaxTimestampDrift, +GenerateCommitment(), +VerifyCommitment() |
| `chaincode/util/validate.go` | Timestamp drift guard in GetTimestamp() |
| `chaincode/util/split.go` | Tombstone DeleteEGO/HGOFromLedger, +MarkEGO/HGOTransferred |
| `chaincode/contracts/query.go` | +PaginatedResult, +Paginated queries, +VerifyQuantityCommitment, Status filtering |
| `chaincode/contracts/device_mgmt.go` | +ListDevicesPaginated |
| `chaincode/contracts/issuance.go` | Status=active, commitment generation, event emission |
| `chaincode/contracts/cancellation.go` | Status=active on remainder GOs |
| `chaincode/contracts/conversion.go` | Status=active on all created GOs |
| `collections/collection-config.json` | blockToLive: 1000000 for private collections |
| `chaincode/META-INF/.../indexStatus.json` | New composite index [Status, GOType] |

---

## Backward Compatibility

- **Public GO struct**: New fields use `omitempty` JSON tags. Existing clients that don't read `Status` or `QuantityCommitment` are unaffected.
- **Query functions**: Original unpaginated `GetCurrentEGOsList` / `GetCurrentHGOsList` / `ListDevices` still work; they now filter tombstoned records (an improvement).
- **Tombstone pattern**: GOs previously deleted before v4.0 will not appear in filtered queries (they were already gone). New cancellations create tombstones.
- **blockToLive**: Affects only new private data writes. Existing data is unaffected until the block threshold is reached.
