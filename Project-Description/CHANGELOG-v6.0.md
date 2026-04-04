# Changelog — v6.0 (Production Hardening Release)

**Date:** 2026-04-04
**Chaincode:** `golifecycle` v6.0
**Fabric:** 2.5.12 | CA 1.5.17 | CouchDB 3.3.3

---

## Summary

v6.0 addresses the **production readiness** gaps from `ARCHITECTURE_CRITIQUE_v5.md`. Six ADRs (017–022) harden the system for deployment: cryptographically secure commitment salts, CEN-EN 16325 field validation, state-based endorsement for per-key owner controls, data archiving support, deprecation of unsafe query endpoints, and an explicit deprecation policy. These changes transform the chaincode from a functional prototype to a production-hardened contract layer.

---

## ADR-017 — Cryptographically Secure Commitment Salts

**Problem:** The v5.0 critique (§3.2.3) identified that `GenerateCommitment` used a deterministic SHA-256 hash of the quantity only. Without a random salt, the commitment was trivially reversible by brute-force enumeration of plausible quantity values.

**Solution:**
- Replaced deterministic hashing with `crypto/rand` 128-bit random salt generation
- The commitment is now `SHA-256(quantity || random_salt)` (non-invertible without the salt)
- Salt is stored in the private data collection alongside the quantity
- Backward compatible: existing commitments remain valid; new commitments use the salted scheme

**Files changed:** `chaincode/assets/counter.go`

---

## ADR-018 — CEN-EN 16325 Field Validation

**Problem:** v5.0 (ADR-012) added CEN-EN 16325 fields to GO structs but accepted any string value. Invalid country codes, malformed EECS energy source codes, or impossible production periods could be recorded on-chain.

**Solution:**
- New `chaincode/util/validate_cen.go` validation module with functions:
  - `ValidateCountryOfOrigin` — allowlist of 30 EU/EEA + CH/GB ISO 3166-1 alpha-2 codes
  - `ValidateSupportScheme` — enum validation (`none`, `FIT`, `FIP`, `quota`, `tax`, `loan`, `other`)
  - `ValidateEnergySource` — regex `^F\d{8}$` matching EECS energy source code format
  - `ValidateGridConnectionPoint` — regex `^[A-Za-z0-9]{16}$` matching 16-character EIC codes
  - `ValidateProductionPeriod` — ensures `PeriodEnd > PeriodStart` and duration ≤ 1 year
  - `ValidateCENFields` — composite function invoking all validators
- Validation runs on every GO creation (electricity, hydrogen, biogas)
- Invalid inputs return descriptive error messages identifying the specific validation failure

**Files changed:** `chaincode/util/validate_cen.go` (new)

---

## ADR-019 — State-Based Endorsement

**Problem:** The v5.0 critique (§2.2.2) noted that the majority endorsement policy (3-of-4) applies uniformly to all state keys, meaning any 3 organizations can modify any GO record — even one belonging to a specific producer.

**Solution:**
- After writing eGO/hGO/bGO public keys to the ledger, the issuance contract sets per-key endorsement policies
- Uses `fabric-contract-api-go/statebased` package:
  - `statebased.NewStateEP(nil)` creates a blank policy
  - Adds the client's MSP (`clientMSP`) as `RoleTypePeer` — the asset owner
  - Adds `issuer1MSP` as `RoleTypePeer` — the registry authority
- Result: only the owner org + issuer can modify each GO key, regardless of the global endorsement policy
- Transfers and cancellations must be endorsed by the current owner

**Files changed:** `chaincode/contracts/issuance.go`

---

## ADR-020 — Data Archiving Support

**Problem:** The v5.0 critique (§1.2.4) and `DATA_DUPLICATION_CONSIDERATIONS.md` identified that per-peer storage grows unboundedly. CouchDB lacks built-in data lifecycle management — cancelled or expired GOs accumulate forever.

**Solution:**
- Designed as an architectural pattern (documented in ADR) rather than chaincode code: cancelled/expired GOs are candidates for off-chain archival to Elasticsearch/S3 with only a hash-pointer retained on-chain
- The tombstone soft-delete pattern (ADR-007) already marks cancelled GOs as `Status: cancelled`, providing the filtering predicate for archival scripts
- CouchDB compaction tuning recommendations documented for reducing revision history overhead

**Files changed:** Architecture documentation only (no chaincode changes)

---

## ADR-021 — Deprecation Policy

**Problem:** No formal process existed for deprecating or removing chaincode functions. ADR-022 required deprecating unpaginated queries, but there was no defined deprecation lifecycle.

**Solution:**
- Established a three-phase deprecation lifecycle:
  1. **Deprecated** — function returns results but logs a deprecation warning in the response
  2. **Sunset** — function returns an error directing callers to the replacement
  3. **Removed** — function removed from chaincode, endpoint returns "unknown function"
- Each deprecated function includes metadata: deprecation version, sunset version, replacement function name
- Minimum 2 minor versions between deprecation and sunset

**Files changed:** Architecture documentation (policy), applied in ADR-022

---

## ADR-022 — Query Endpoint Deprecation

**Problem:** The v5.0 critique (§1.2.3) identified unpaginated query endpoints (`ListDevices`, `GetCurrentEGOsList`, `GetCurrentHGOsList`, `GetCurrentBGOsList`) as latent scalability risks. At production scale, these endpoints would trigger full CouchDB scans returning megabytes of data.

**Solution:**
- Added deprecation warnings to:
  - `device:ListDevices` → replacement: `device:ListDevicesPaginated`
  - `query:GetCurrentEGOsList` → replacement: `query:GetCurrentEGOsListPaginated`
  - `query:GetCurrentHGOsList` → replacement: `query:GetCurrentHGOsListPaginated`
  - `query:GetCurrentBGOsList` → replacement: `query:GetCurrentBGOsListPaginated`
- Each function now returns a `DeprecationNotice` field in the response: version deprecated, replacement path, sunset version
- Functions remain operational for backward compatibility; sunset planned for v8.0

**Files changed:** `chaincode/contracts/device_mgmt.go`, `chaincode/contracts/query.go`

---

## Migration Notes

- **No breaking changes.** All v5.0 APIs remain functional.
- **CEN-EN 16325 validation:** Existing GO records are not retroactively validated. New GOs must pass validation. If external systems submit `EnergySource` fields that don't match `^F\d{8}$` format, they will be rejected.
- **State-based endorsement:** Applies only to newly created GOs. Existing GOs retain the global endorsement policy. A one-time migration script could apply per-key policies to existing assets.
- **Deprecation warnings:** Client applications should begin migrating to paginated query endpoints. Unpaginated endpoints will return errors in v8.0.

---

## Deployment

```bash
# Package and install
peer lifecycle chaincode package golifecycle_6.0.tar.gz --path ./chaincode --lang golang --label golifecycle_6.0
peer lifecycle chaincode install golifecycle_6.0.tar.gz

# Approve and commit (all 4 orgs)
peer lifecycle chaincode approveformyorg --channelID goplatformchannel --name golifecycle --version 6.0 --sequence <N>
peer lifecycle chaincode commit --channelID goplatformchannel --name golifecycle --version 6.0 --sequence <N>
```

---

*Changelog compiled 2026-04-04. Covers ADR-017 through ADR-022.*
