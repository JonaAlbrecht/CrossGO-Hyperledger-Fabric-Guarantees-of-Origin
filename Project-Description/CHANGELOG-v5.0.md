# Changelog — v5.0 (Interoperability & Multi-Carrier Release)

**Date:** 2026-04-05
**Chaincode:** `golifecycle` v5.0
**Fabric:** 2.5.12 | CA 1.5.17 | CouchDB 3.3.3

---

## Summary

v5.0 addresses the **interoperability** and **extensibility** gaps from `ARCHITECTURE_CRITIQUE.md`. Six ADRs (012–016) introduce CEN-EN 16325 standard alignment, API versioning, dynamic organization onboarding, a biogas energy carrier, and event-driven CQRS support. These changes follow agency-blueprint principles: Separation of Concerns (admin contract), Explicit > Implicit (version negotiation), and the Scout Rule (standardised data models).

---

## ADR-012 — CEN-EN 16325 Data Model Alignment

**Problem:** GO data structures used custom ad-hoc fields. The critique (§4.2.2) identified that `ElectricityGO` and `GreenHydrogenGO` did not conform to CEN-EN 16325 (EU GO standard) or IEC 62325 (energy market communication).

**Solution:**
- Added CEN-EN 16325 aligned fields to public GO structs:
  - `CountryOfOrigin` — ISO 3166-1 alpha-2 code (e.g. "DE", "NL")
  - `GridConnectionPoint` — EIC code of the grid connection
  - `SupportScheme` — "none", "FIT", "FIP", "quota" etc.
  - `EnergySource` — EN 16325 source code (e.g. "F01010100" = solar PV)
  - `ProductionPeriodStart` / `ProductionPeriodEnd` — UNIX timestamps
- All new fields use `omitempty` JSON tags for backward compatibility
- Applied to `ElectricityGO`, `GreenHydrogenGO`, and the new `BiogasGO`

**Files changed:** `electricity_go.go`, `hydrogen_go.go`, `biogas_go.go`

---

## ADR-013 — API Versioning

**Problem:** The critique (§4.2.5) noted no versioning strategy. The v2→v3 transition introduced breaking changes with no version negotiation or deprecation policy.

**Solution:**
- New `AdminContract` with `GetVersion()` function
- Returns `VersionInfo` struct: version ("5.0.0"), chaincode ID, supported API levels, breaking change flag
- Clients call `admin:GetVersion` before invoking domain functions to verify compatibility
- API levels follow `<contract>/v1` naming convention for future version routing
- Contract registered in `main.go` as the 7th contract namespace

**Files changed:** `admin.go` (new), `main.go`

---

## ADR-014 — Dynamic Organization Onboarding

**Problem:** The critique (§4.2.4) identified that adding organizations required manual crypto generation, channel config updates, collection config changes, and chaincode upgrades — a multi-day process.

**Solution:**
- `RegisterOrganization` function in `AdminContract` (issuer-only)
- Records `RegisteredOrganization` on-chain: MSP, display name, type, carriers, country, registration time
- `GetOrganization` query to retrieve registration records
- On-chain registration provides the application-layer record; the Fabric channel config update (adding MSP) remains an out-of-band admin operation
- Emits `ORG_REGISTERED` lifecycle event for off-chain indexers
- Path toward OU-based organization model: producers can share a common MSP with distinguishing OU attributes

**Files changed:** `admin.go` (new), `main.go`

---

## ADR-015 — Biogas Energy Carrier

**Problem:** The critique (§4.1.1) noted the multi-carrier design was extensible but only electricity and hydrogen were implemented. RED III mandates GO schemes for all renewable carriers including biogas.

**Solution:**
- New `BiogasGO` public struct and `BiogasGOPrivateDetails` private struct
- New `CancellationStatementBiogas` for biogas cancellation records
- New `BiogasContract` with `CreateBiogasGO` and `CancelBiogasGO` functions
- Biogas-specific attributes: `VolumeNm3`, `EnergyContentMWh`, `BiogasProductionMethod`, `FeedstockType`
- ID prefixes: `bGO_`, `bCancel_` with range queries `bGO~`
- Includes all v4.0 patterns: Status field, QuantityCommitment, tombstone cancellation
- Query functions: `GetCurrentBGOsList`, `GetCurrentBGOsListPaginated`, `ReadPublicBGO`
- Registered as 8th contract namespace in `main.go`

**Files changed:** `biogas_go.go` (new), `counter.go` (+prefixes), `biogas.go` (new contract), `query.go` (+biogas queries), `main.go`

---

## ADR-016 — Event-Driven CQRS

**Problem:** The critique (§5, scale engineering) identified that complex queries require a CQRS/off-chain indexing strategy. CouchDB rich queries are insufficient for production analytics workloads.

**Solution:**
- New `util/events.go` with `EmitLifecycleEvent` helper
- Standardized `LifecycleEvent` struct: EventType, AssetID, GOType, Initiator, TxID, Timestamp, Details
- Seven event types: `GO_CREATED`, `GO_TRANSFERRED`, `GO_CANCELLED`, `GO_CONVERTED`, `GO_SPLIT`, `DEVICE_REGISTERED`, `DEVICE_REVOKED`
- Events emitted from issuance (eGO, hGO, bGO creation) and biogas cancellation
- Off-chain listeners consume these Fabric chaincode events to build query-optimised read models (e.g. PostgreSQL, Elasticsearch)
- Single event per transaction (Fabric constraint); compound operations use Details map

**Files changed:** `events.go` (new), `issuance.go`, `biogas.go`, `admin.go`

---

## New Files (v5.0)

| File | Purpose |
|------|---------|
| `chaincode/assets/biogas_go.go` | BiogasGO, BiogasGOPrivateDetails, CancellationStatementBiogas structs |
| `chaincode/contracts/biogas.go` | BiogasContract: CreateBiogasGO, CancelBiogasGO |
| `chaincode/contracts/admin.go` | AdminContract: GetVersion, RegisterOrganization, GetOrganization |
| `chaincode/util/events.go` | LifecycleEvent struct, EmitLifecycleEvent helper |

## Modified Files (v5.0)

| File | Changes |
|------|---------|
| `chaincode/assets/electricity_go.go` | +CEN-EN 16325 fields (CountryOfOrigin, etc.) |
| `chaincode/assets/hydrogen_go.go` | +CEN-EN 16325 fields |
| `chaincode/assets/counter.go` | +PrefixBGO, +PrefixBCancellation, +RangeEndBGO |
| `chaincode/contracts/query.go` | +GetCurrentBGOsList, +GetCurrentBGOsListPaginated, +ReadPublicBGO |
| `chaincode/contracts/issuance.go` | +Event emission (GO_CREATED) |
| `chaincode/main.go` | +adminContract, +biogasContract registration |

---

## Contract Registry (v5.0)

| # | Namespace | Functions |
|---|-----------|-----------|
| 1 | `issuance` | CreateElectricityGO, CreateHydrogenGO |
| 2 | `transfer` | TransferEGO, TransferEGOByAmount, TransferHGOByAmount |
| 3 | `conversion` | AddHydrogenToBacklog, IssuehGO |
| 4 | `cancellation` | ClaimRenewableAttributesElectricity, ClaimRenewableAttributesHydrogen, VerifyCancellationStatement |
| 5 | `query` | 18 functions (point reads, paginated lists, commitment verification) |
| 6 | `device` | RegisterDevice, GetDevice, ListDevices, ListDevicesPaginated, RevokeDevice, SuspendDevice, ReactivateDevice, RegisterOrgRole, InitLedger |
| 7 | `admin` | GetVersion, RegisterOrganization, GetOrganization |
| 8 | `biogas` | CreateBiogasGO, CancelBiogasGO |

---

## Backward Compatibility

- **CEN-EN 16325 fields**: All use `omitempty` — existing GOs without these fields unmarshal correctly.
- **Admin contract**: Additive change. Existing clients that don't call `admin:*` functions are unaffected.
- **Biogas contract**: Additive change. New contract namespace doesn't affect existing electricity/hydrogen flows.
- **Lifecycle events**: Additive change. If no off-chain listener is running, events are simply not consumed.
- **Version negotiation**: Optional. Clients that don't call `GetVersion` continue working as before.
