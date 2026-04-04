# Changelog — v7.0 (Market Integration Release)

**Date:** 2026-04-04
**Chaincode:** `golifecycle` v7.0
**Fabric:** 2.5.12 | CA 1.5.17 | CouchDB 3.3.3

---

## Summary

v7.0 extends the GO platform from a national registry to a **cross-border market integration layer**. Three ADRs (024, 027, 029) add cross-registry bridge transfers, IoT device attestation with ECDSA signature verification, and external data oracle integration for ENTSO-E grid generation records. The chaincode grows from 8 to **10 contract namespaces** with **~50 exported functions**, and the admin API reports version `7.0.0` with 9 supported API levels.

---

## ADR-024 — Cross-Registry Bridge Transfers

**Problem:** The v5.0 critique (§4.2.2) noted that GO transfers are confined to a single Fabric channel. European GO trading requires interoperability between national registries (e.g., a German eGO consumed by a Dutch buyer). No mechanism existed for importing or exporting GOs across registry boundaries.

**Solution:**
- New `BridgeContract` registered as the 9th contract namespace (`bridge`)
- `BridgeTransfer` struct with fields: `TransferID`, `Direction` (export/import), `GOAssetID`, `ExternalRegistry`, `ExternalID`, `GOType` (eGO/hGO/bGO), `Status`, `InitiatedBy`, `AmountMWh`, `CountryOfOrigin`, `EnergySource`, timestamps
- Functions:
  - `ExportGO(goAssetID, externalRegistry, externalID)` — locks the GO on the local registry, creates a `pending` bridge record. Requires issuer role.
  - `ConfirmExport(transferID)` — transitions bridge record to `confirmed` after the receiving registry acknowledges. Requires issuer role.
  - `ImportGO(externalRegistry, externalID, goType, amountMWh, countryOfOrigin, energySource)` — creates a local GO asset from an external registry's export confirmation. Requires issuer role.
  - `GetBridgeTransfer(transferID)` — point read of a bridge record
  - `ListBridgeTransfersPaginated(pageSize, bookmark)` — cursor-based pagination over bridge records
- Key prefix: `bridge_` with range queries `bridge_` → `bridge~`
- CEN-EN 16325 validation applied to all bridge records: country code, energy source code
- Lifecycle events emitted: `BRIDGE_EXPORT_INITIATED`, `BRIDGE_EXPORT_CONFIRMED`, `BRIDGE_IMPORT_COMPLETED`

**Files changed:** `chaincode/contracts/bridge.go` (new), `chaincode/main.go`

---

## ADR-027 — IoT Device Attestation

**Problem:** The v5.0 critique (§3.1.2) identified that device identity was enforced only by X.509 certificate attributes. There was no mechanism for devices to cryptographically sign their readings, and no on-chain verification of measurement data integrity beyond the TLS transport layer.

**Solution:**
- New `PublicKeyPEM` field on the `Device` struct (optional, `metadata:",optional"`)
- New functions on `DeviceManagementContract`:
  - `VerifyDeviceReading(deviceID, readingJSON, signatureBase64)` — verifies an ECDSA P-256 signature against the device's registered public key. Returns verification status without state changes.
  - `SubmitSignedReading(deviceID, readingJSON, signatureBase64)` — verifies the signature and, if valid, stores a `DeviceReading` record on the ledger with the verified flag
- Signature verification flow:
  1. Loads device record, extracts `PublicKeyPEM`
  2. Parses the PEM-encoded ECDSA P-256 public key
  3. Computes SHA-256 hash of the reading JSON
  4. Verifies the ASN.1 DER signature using `ecdsa.VerifyASN1`
- The `metadata:",optional"` tag ensures the `PublicKeyPEM` field is excluded from the JSON schema's `required` array, maintaining backward compatibility with pre-v7.0 devices that lack public keys

**Files changed:** `chaincode/assets/device.go`, `chaincode/contracts/device_mgmt.go`

---

## ADR-029 — External Data Oracle

**Problem:** GO registries must cross-reference issued GOs against actual grid generation data (ENTSO-E Transparency Platform) to detect over-issuance. No on-chain mechanism existed for publishing or querying external grid generation data, requiring off-chain manual verification.

**Solution:**
- New `OracleContract` registered as the 10th contract namespace (`oracle`)
- `GridGenerationRecord` struct: `RecordID`, `BiddingZone`, `PeriodStart`, `PeriodEnd`, `EnergySource` (EECS code), `GenerationMW`, `EmissionFactor`, `DataSource`, `PublishedBy`, `PublishedAt`
- Functions:
  - `PublishGridData` — issuer publishes ENTSO-E grid generation records via transient data. Validates EECS energy source code format. Requires issuer role.
  - `GetGridData(recordID)` — point read of a grid generation record
  - `ListGridDataPaginated(pageSize, bookmark)` — cursor-based pagination over oracle records
  - `CrossReferenceGO(goAssetID)` — validates a GO's `ProductionPeriodStart`, `ProductionPeriodEnd`, and `EnergySource` against available oracle records for the same bidding zone and time period. Returns `matching_records_found` or `no_matching_data` status with the matched records.
- Key prefix: `oracle_` with range queries `oracle_` → `oracle~`
- CEN-EN 16325 validation applied: energy source code format
- Lifecycle events emitted: `ORACLE_DATA_PUBLISHED`, `ORACLE_CROSS_REFERENCE`

**Files changed:** `chaincode/contracts/oracle.go` (new), `chaincode/main.go`

---

## Admin API Changes

- `admin:GetVersion` now returns:
  ```json
  {
    "version": "7.0.0",
    "chaincodeId": "golifecycle",
    "supportedAPIs": [
      "issuance/v1", "query/v1", "transfer/v1", "cancellation/v1",
      "device/v1", "admin/v1", "biogas/v1", "bridge/v1", "oracle/v1"
    ]
  }
  ```
- 10 contract namespaces registered in `main.go`: issuance, transfer, conversion, cancellation, query, device, admin, biogas, **bridge**, **oracle**

---

## Bug Fixes

- **Device struct schema validation fix:** The `PublicKeyPEM` field was added with only `json:"publicKeyPEM,omitempty"`. The `fabric-contract-api-go` schema generator marks all string fields as `required` by default, causing `ListDevices` to fail with "publicKeyPEM is required" for pre-v7.0 devices. Fixed by adding `metadata:",optional"` tag, which instructs the schema generator to exclude the field from the `required` array. (Commit 8d35eb3)

---

## Performance Characteristics

Caliper v0.6.0 benchmark (28 rounds, 10 workers) on v7.0 deployed at sequence 5:

| New Function | Type | Peak Tested TPS | Success Rate | Avg Latency |
|---|---|---|---|---|
| `oracle:GetGridData` | Read | 500 | 100% | 10ms |
| `oracle:PublishGridData` | Write | Serial (10 txns) | 100% | 110ms |

All existing v5.0 read and write benchmarks maintain baseline performance — no regressions detected. Full results in `testing/PERFORMANCE_REPORT_v7.md`.

---

## Migration Notes

- **No breaking changes.** All v5.0/v6.0 APIs remain functional.
- **Device PublicKeyPEM:** Existing devices do not need public keys. The field is optional and only used by `VerifyDeviceReading` / `SubmitSignedReading`. Devices without public keys cannot use signing features.
- **Oracle seeding:** The `CrossReferenceGO` function requires oracle data to be published before it can validate GOs. An external process should periodically publish ENTSO-E data via `PublishGridData`.
- **Bridge transfers:** Require out-of-band coordination with the external registry. The on-chain bridge record tracks the transfer lifecycle; actual data exchange occurs outside the chaincode.

---

## Deployment

Deployed to Hetzner VM as:
- **Version label:** `golifecycle_7.0.1` (bumped from 7.0 to 7.0.1 to disambiguate package IDs)
- **Sequence:** 5
- **Package ID:** `golifecycle_7.0.1:63b3a0ad56c01e63a4bec6068792402c4480cc210b8b728feb1049f9c5d9acad`

```bash
peer lifecycle chaincode package golifecycle_7.0.1.tar.gz --path ./chaincode --lang golang --label golifecycle_7.0.1
peer lifecycle chaincode install golifecycle_7.0.1.tar.gz
# Approve on all 4 orgs, then commit at sequence 5
```

---

*Changelog compiled 2026-04-04. Covers ADR-024, ADR-027, ADR-029.*
