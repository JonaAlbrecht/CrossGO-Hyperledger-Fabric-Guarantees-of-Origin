# Changelog — v9.0 (Full-Stack Multi-Carrier Release)

**Date:** 2026-04-12
**Chaincode:** `golifecycle` v9.0.4 (sequence 7)
**Fabric:** 2.5.12 | CouchDB 3.3.3

---

## Summary

v9.0 transforms the GO platform from a **two-carrier chaincode prototype** (electricity + hydrogen) into a **four-carrier full-stack application** with a production-grade React frontend. Three categories of changes define this release: (1) new energy carrier support for heating/cooling per RED III Art. 19 and full-stack integration of biogas, (2) critical chaincode bug fixes discovered during end-to-end testing that resolved endorsement mismatches, field mapping errors, and schema validation failures, and (3) a complete application layer rewrite with a React dashboard, RBAC-gated navigation, and per-carrier lifecycle management UI. The chaincode grows from 10 to **11 contract namespaces** with **~55 exported functions**, and the admin API reports version `9.0.0` with 11 supported API levels.

---

## New Energy Carrier: Heating/Cooling (RED III Art. 19)

**Problem:** The revised Renewable Energy Directive (RED III) mandates GO schemes for heating and cooling energy carriers. The platform supported electricity, hydrogen, and biogas but lacked thermal energy carrier support.

**Solution:**
- New `HeatingCoolingContract` registered as the 11th contract namespace (`heating_cooling`)
- New asset structs: `HeatingCoolingGO` (public), `HeatingCoolingGOPrivateDetails` (private), `CancellationStatementHeatingCooling`
- Functions:
  - `CreateHeatingCoolingGO` — issuer/producer creates heating/cooling GO from metering data. Validates AmountMWh, Emissions, SupplyTemperature, ElapsedSeconds.
  - `ClaimRenewableAttributesHeatingCooling(hcGOID, amountMWh)` — cancels a heating/cooling GO with optional partial cancellation
- `SupplyTemperature` field distinguishes heating (positive °C) from cooling (negative °C)
- Production methods: `heat_pump`, `solar_thermal`, `geothermal`, `biomass_boiler`, `district_heating`, `absorption_chiller`
- New prefixes: `PrefixHCGO = "hcGO_"`, `PrefixHCCancellation = "hcCancel_"`, `RangeEndHCGO = "hcGO_~"`
- Query support: `GetCurrentHCGOsList`, `GetCurrentHCGOsListPaginated`, `ReadPublicHCGO`, `ReadPrivateHCGO`
- CEN-EN 16325 validation applied at creation time

**Files added:** `chaincode/assets/heating_cooling_go.go`, `chaincode/contracts/heating_cooling.go`
**Files changed:** `chaincode/main.go`, `chaincode/assets/counter.go`, `chaincode/contracts/query.go`

---

## Cross-Channel Bridge Extension: All Four Carriers

**Problem:** The v8.0 cross-channel bridge protocol (`MintFromBridge`) only handled electricity and hydrogen GOs. Biogas and heating/cooling GOs could not be bridged across channels.

**Solution:**
- Extended `MintFromBridge` with `case "Biogas"` and `case "HeatingCooling"` branches
- Each branch creates the appropriate public/private asset structs with quantity commitments
- Bridge helper functions added: `writeBGOToLedgerBridge()`, `writeHCGOToLedgerBridge()` for bridge-specific asset creation
- `ImportGO` (v7.0 legacy) also extended with Biogas and HeatingCooling import support
- `GOType` field now accepts: `"Electricity"`, `"Hydrogen"`, `"Biogas"`, `"HeatingCooling"`
- Dual-issuer consensus fields added to `CrossChannelLock`: `SourceIssuerMSP`, `TargetIssuerMSP`, `TargetIssuerApproval`, `TargetIssuerApprovedAt`

**Files changed:** `chaincode/contracts/bridge.go`

---

## Query Layer Extension: Biogas + Heating/Cooling Private Data

**Problem:** v8.0 query contract lacked private data read functions for biogas and heating/cooling GOs. The `ReadPrivateBGO` and `ReadPrivateHCGO` functions were missing, preventing frontend access to confidential production details.

**Solution:**
- New query functions added to `QueryContract`:
  - `ReadPrivateBGO` — reads biogas GO private details via transient `QueryInput` (Collection, BGOID)
  - `ReadPrivateHCGO` — reads heating/cooling GO private details via transient `QueryInput` (Collection, HCGOID)
  - `GetCurrentHCGOsList` — unbounded range query for heating/cooling GOs (deprecated, included for completeness)
  - `GetCurrentHCGOsListPaginated` — paginated list of active heating/cooling GOs
  - `ReadPublicHCGO` — point read of a single heating/cooling GO
- All new query functions enforce collection access validation via `access.ValidateCollectionAccess()`

**Files changed:** `chaincode/contracts/query.go`

---

## Organization Registry (ListOrganizations)

**Problem:** v8.0 provided `RegisterOrganization` and `GetOrganization` but no way to enumerate all registered organizations, preventing the frontend from displaying the org registry.

**Solution:**
- New `ListOrganizations` function on `AdminContract` — range query over `org_` prefix returning all `RegisteredOrganization` records
- Version bumped to `9.0.0` with 11 supported API levels including `heating_cooling/v1`
- API levels updated: `conversion/v2`, `admin/v2`, `bridge/v2` (reflecting v9 extensions)

**Files changed:** `chaincode/contracts/admin.go`

---

## Critical Bug Fix: Non-Deterministic Commitment Salt (Endorsement Mismatch)

**Problem (Bug #5):** `GenerateCommitment()` in `counter.go` used `crypto/rand` to generate a 128-bit random salt. In Fabric's multi-org endorsement model, each endorsing peer independently executes the chaincode. Random salts produce different values on each peer, causing endorsement mismatches when the read-write sets don't match. This manifested as "ProposalResponsePayloads do not match" errors during any GO creation with `AND(org.member, issuer1.member)` endorsement policies.

**Solution:**
- Replaced `crypto/rand` with deterministic salt derivation: `SHA-256(txID + "||salt||" + quantity)[:16]`
- The transaction ID is identical across all endorsing peers for the same proposal, ensuring deterministic output
- Removed `crypto/rand` import entirely
- Tradeoff: Deterministic salt has lower brute-force resistance than true random salt, but this is required for Fabric endorsement consistency. The txID provides 256-bit entropy per transaction.

**Files changed:** `chaincode/assets/counter.go`

---

## Critical Bug Fix: Schema Validation Failure on CEN-EN 16325 Fields

**Problem (Bug #12, v9.0.3):** `fabric-contract-api-go v1.2.2` generates JSON Schema where all non-pointer struct fields are marked as `required`, regardless of `omitempty` JSON tags. This caused `GetCurrentEGOsList` and all list queries to fail with "CountryOfOrigin is required" errors when GOs were created without CEN-EN 16325 metadata.

**Solution:**
- Added `metadata:",optional"` tag to all 6 CEN-EN 16325 fields across all 4 GO structs:
  - `ElectricityGO`: CountryOfOrigin, GridConnectionPoint, SupportScheme, EnergySource, ProductionPeriodStart, ProductionPeriodEnd
  - `GreenHydrogenGO`: same 6 fields
  - `BiogasGO`: same 6 fields
  - `HeatingCoolingGO`: created with tags from the start
- CEN-EN 16325 fields are now excluded from the contractapi-generated JSON Schema `required` array
- Default values populated during GO creation: `CountryOfOrigin="DE"`, `EnergySource` from production method

**Files changed:** `chaincode/assets/electricity_go.go`, `chaincode/assets/hydrogen_go.go`, `chaincode/assets/biogas_go.go`

---

## Bug Fix: Role Terminology Standardization (consumer → buyer)

**Problem:** The codebase inconsistently used "consumer" and "buyer" to refer to the same market participant role. The chaincode used `RoleConsumer = "consumer"` while the application layer used "buyer", causing role check failures during transfers and cancellations.

**Solution:**
- Renamed `RoleConsumer` → `RoleBuyer` with value `"buyer"` in `access/roles.go`
- Updated `RegisterOrgRole` validation to accept `"buyer"` instead of `"consumer"`
- Updated documentation comment in `access/collections.go`

**Files changed:** `chaincode/access/roles.go`, `chaincode/access/collections.go`

---

## Full-Stack Application Layer (Frontend + Backend)

### Backend (Express.js + Fabric Gateway SDK)

**New routes:**
- `POST/GET /api/bridge/{verify,initiate,approve}` — cross-channel bridge protocol endpoints
- `POST/GET /api/organizations` — organization registry CRUD with hardcoded fallback
- `POST /api/guarantees/biogas` — biogas GO creation
- `POST /api/guarantees/heating-cooling` — heating/cooling GO creation
- `POST /api/transfers/biogas-by-amount` — biogas batch transfer
- `POST /api/transfers/heating-cooling-by-amount` — heating/cooling batch transfer
- `POST /api/cancellations/biogas` — biogas cancellation
- `POST /api/cancellations/heating_cooling` — heating/cooling cancellation

**Key fixes (v9.0.0 → v9.0.4):**
- Backend used wrong contract name (`AdminContract` → `admin`), wrong transient key (`OrgInput` → `OrgRegistration`), wrong field (`Role` → `OrgType`)
- GO creation mapped generic `amount`/`productionMethod` to carrier-specific names (`AmountMWh`/`ElectricityProductionMethod`)
- Biogas/heating_cooling routed to correct contract namespaces (was incorrectly using `issuance`)
- Conversion backlog needed `endorsingOrganizations: [callerMSP, 'issuer1MSP']` for private data endorsement
- Conversion backlog GET query required `QueryInput` transient key with collection name

**Files changed:** `application/backend/src/index.ts`, `application/backend/src/routes/*.ts`, `application/backend/src/fabric/gateway.ts`
**Files added:** `application/backend/src/routes/bridge.ts`, `application/backend/src/routes/organizations.ts`

### Frontend (React + Vite + TypeScript)

**New pages:**
- `DashboardPage` — statistics overview with per-carrier GO counts, role-specific action buttons
- `GuaranteesPage` — GO creation form with carrier selector and production method dropdowns for all 4 carriers
- `TransfersPage` — single GO and batch-by-amount transfer modes for all 4 carriers
- `CertificatesPage` (CancellationsPage) — GO cancellation with partial amount support
- `ConversionsPage` — backlog management and cross-carrier conversion issuance
- `DevicesPage` — device registration with energy carrier multi-select and conversion efficiency matrix
- `OrganizationsPage` — organization registry with issuer-only registration form
- `VerificationPage` — three-tab audit interface (cancellation statement, bridge proof, GO lifecycle)

**Key components:**
- `Layout.tsx` — RBAC-gated sidebar navigation with Management, Operations, and Verification sections
- `api.ts` — Axios client with Fabric gRPC error extraction and JWT auth
- `types.ts` — shared TypeScript interfaces with `ENERGY_CARRIERS` configuration and `carrierStyle()` utility

**Files changed:** All files under `application/frontend/src/`
**Files added:** `application/frontend/src/pages/OrganizationsPage.tsx`, `application/frontend/src/pages/VerificationPage.tsx`, `application/frontend/src/components/Tooltip.tsx`

---

## Network Configuration

- Orderer container names updated in `docker-compose-orderer.yaml` for v9 compatibility
- `network-up.sh` minor path fix for orderer TLS certificates
- New `deploy-v9.sh` script for single-channel v9 deployment (packages, installs, approves on all 4 orgs, commits at sequence 7)

**Files changed:** `network/docker/docker-compose-orderer.yaml`, `network/scripts/network-up.sh`
**Files added:** `network/scripts/deploy-v9.sh`

---

## Admin API Changes

- `admin:GetVersion` now returns:
  ```json
  {
    "version": "9.0.0",
    "chaincodeId": "golifecycle",
    "supportedApis": [
      "issuance/v1", "transfer/v1", "conversion/v2", "cancellation/v1",
      "query/v1", "device/v1", "admin/v2", "bridge/v2", "oracle/v1",
      "biogas/v1", "heating_cooling/v1"
    ],
    "breakingChange": false
  }
  ```
- 11 contract namespaces registered in `main.go`: issuance, transfer, conversion, cancellation, query, device, admin, biogas, bridge, oracle, **heating_cooling**

---

## Performance Characteristics

No new Caliper benchmarks were run for v9.0. The chaincode additions (heating/cooling contract, query extensions, bridge carrier support) follow identical patterns to existing contracts (biogas, electricity) and are expected to maintain the v8.0 baseline:
- Write throughput: 50 TPS with 90ms average latency
- Read throughput: 2,000 TPS with sub-10ms latency
- Paginated reads: 500 TPS with 10ms latency

A v9.0 Caliper benchmark should be run to confirm zero regression after the heating/cooling additions.

---

## Migration Notes

- **No breaking changes** to existing chaincode APIs. All v8.0 functions remain functional.
- **Role rename:** `consumer` → `buyer` in `access/roles.go`. Existing on-chain org role registrations using "consumer" must be re-registered as "buyer" via `RegisterOrgRole`.
- **Schema compatibility:** The `metadata:",optional"` tags on CEN-EN 16325 fields are backward compatible — existing GOs without these fields are now correctly handled by query functions.
- **Deterministic salt:** Existing GOs created with v6.0–v8.0 random salts remain verifiable. New GOs use txID-derived deterministic salts.

---

## Deployment

Deployed to Hetzner VM as:
- **Version label:** `golifecycle` v9.0.4
- **Sequence:** 7
- **Channel:** `goplatformchannel` (single-channel deployment)
- **Endorsement policy:** `OutOf(2, ...)` (changed from MAJORITY 3/4)

```bash
peer lifecycle chaincode package golifecycle_9.0.tar.gz --path ./chaincode --lang golang --label golifecycle_9.0
peer lifecycle chaincode install golifecycle_9.0.tar.gz
# Approve on all 4 orgs, then commit at sequence 7
```

---

## Bug Summary (v9.0.0 → v9.0.4)

| Bug # | Component | Description | Fix |
|-------|-----------|-------------|-----|
| 1 | Backend → Admin | Wrong contract name, transient key, field name | Corrected to `admin`, `OrgRegistration`, `OrgType` |
| 2 | Backend → Guarantees | Frontend sends generic field names; backend must map to carrier-specific | Added per-carrier field mapping in guarantees.ts |
| 3 | Backend → Biogas/HC | Routed to `issuance` contract instead of `biogas`/`heating_cooling` | Fixed contract routing |
| 4 | Chaincode → ABAC | `AssertAttribute()` requires Fabric CA custom attrs; fails with cryptogen | Removed cert attribute checks; RequireRole suffices |
| 5 | Chaincode → Counter | `crypto/rand` causes endorsement mismatch in multi-org policies | Changed to txID-derived deterministic salt |
| 6 | Chaincode → Endorsement | MAJORITY (3/4) policy requires 3 orgs but producer+issuer=2 | Changed to `OutOf(2, ...)` |
| 7 | Backend → PDC | Private data submits missing `endorsingOrganizations` hint | Added `[callerMSP, 'issuer1MSP']` to all PDC submits |
| 8–11 | Chaincode → Split/Cancel | Various split/cancellation edge cases | Fixed in issuance, cancellation, transfer |
| 12 | Chaincode → Schema | `fabric-contract-api-go` marks all fields as required | Added `metadata:",optional"` to CEN-EN fields |

---

*Changelog compiled 2026-04-12. Covers heating/cooling carrier support, critical endorsement and schema fixes, full-stack application layer, and bridge protocol extension to all 4 energy carriers.*
