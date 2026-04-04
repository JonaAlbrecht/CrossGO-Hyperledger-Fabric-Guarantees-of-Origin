# Architecture Redesign: GO Lifecycle Platform v3

## 1. Overview

This document describes the comprehensive architectural overhaul of the Hyperledger Fabric–based Guarantee of Origin (GO) issuance, transfer and conversion system. The redesign transforms the project from a fixed-org monolithic prototype into a **tiered, multi-carrier, object-oriented platform** with a full-stack frontend.

### Design Principles
- **Tiered network**: Role-based organizations (Issuer, Producer, Consumer) instead of hardcoded org names
- **Multi-carrier**: Extensible beyond electricity/hydrogen — support any energy carrier via a common GO interface
- **Object-oriented chaincode**: Separate asset types (Device, GO, Certificate) into dedicated files with shared base types
- **Contention-free writes**: Deterministic hash-based IDs derived from transaction IDs — no shared state during ID generation (v3.0)
- **Performance-validated**: Architecture verified by Hyperledger Caliper benchmarks; 100% write success rate at 50 TPS (v3.0)
- **Bug-free**: Fix all 12 identified bugs from the current monolithic chaincode
- **Full-stack**: TypeScript frontend using the Fabric Gateway client API

---

## 2. Repository Structure (New)

```
HLF-GOconversionissuance-JA-MA/
├── README.md
├── ARCHITECTURE.md                    # This file
├── Project-Description/
│
├── chaincode/                         # ── NEW: Restructured Go chaincode ──
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                        # Entrypoint: registers all contracts
│   ├── assets/                        # Asset type definitions
│   │   ├── base.go                    # BaseGO interface & common fields
│   │   ├── device.go                  # Device (SmartMeter, OutputMeter) asset
│   │   ├── electricity_go.go          # ElectricityGO + private details
│   │   ├── hydrogen_go.go             # HydrogenGO + private details + backlog
│   │   ├── certificate.go             # CancellationStatement, ConsumptionDeclaration
│   │   └── counter.go                 # Persistent on-chain counter (replaces in-memory)
│   ├── contracts/                     # Smart contract logic grouped by domain
│   │   ├── issuance.go                # CreateElectricityGO, CreateHydrogenGO
│   │   ├── transfer.go                # TransferGO, TransferGOByAmount
│   │   ├── conversion.go              # Hydrogen backlog, IssuehGO (e→h conversion)
│   │   ├── cancellation.go            # ClaimRenewableAttributes, VerifyCancellation
│   │   ├── query.go                   # All read/query functions
│   │   └── device_mgmt.go             # RegisterDevice, RevokeDevice, ListDevices
│   ├── access/                        # Access control helpers
│   │   ├── roles.go                   # Role-based checks (IsIssuer, IsProducer, etc.)
│   │   ├── abac.go                    # ABAC attribute helpers (getOrgAttr, device attrs)
│   │   └── collections.go             # Collection name resolution by org role
│   └── util/                          # Shared utilities
│       ├── iterator.go                # Generic iterator constructors
│       ├── validate.go                # Input validation helpers
│       └── split.go                   # Proportional splitting algorithm
│
├── collections/                       # ── NEW: Private data collection configs ──
│   └── collection-config.json         # Dynamic role-based collections
│
├── network/                           # ── NEW: Consolidated network setup ──
│   ├── configtx.yaml                  # Channel configuration (tiered roles)
│   ├── docker/                        # Docker Compose files
│   │   ├── docker-compose-orderer.yaml
│   │   ├── docker-compose-issuer.yaml
│   │   ├── docker-compose-producer.yaml   # Template for any producer
│   │   ├── docker-compose-consumer.yaml   # Template for any consumer
│   │   ├── docker-compose-ca.yaml
│   │   └── docker-compose-couchdb.yaml
│   ├── organizations/                 # Crypto material generation
│   │   ├── fabric-ca/                 # CA configs per org-type
│   │   └── cryptogen/                 # Alternative: cryptogen configs
│   ├── scripts/                       # Operational scripts
│   │   ├── network-up.sh
│   │   ├── network-down.sh
│   │   ├── create-channel.sh
│   │   ├── deploy-chaincode.sh
│   │   ├── add-org.sh                 # ── NEW: Dynamic org onboarding ──
│   │   └── set-anchor-peers.sh
│   └── base.yaml                      # Base peer/orderer config
│
├── application/                       # ── NEW: Full-stack frontend ──
│   ├── backend/                       # Node.js/Express REST API
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   ├── src/
│   │   │   ├── index.ts               # Express server entrypoint
│   │   │   ├── fabric/                # Fabric Gateway connection
│   │   │   │   ├── gateway.ts         # gRPC + Gateway setup
│   │   │   │   ├── identity.ts        # Wallet & identity management
│   │   │   │   └── contracts.ts       # Contract accessor helpers
│   │   │   ├── routes/                # REST API routes
│   │   │   │   ├── auth.ts            # Authentication (JWT + Fabric identity)
│   │   │   │   ├── devices.ts         # Device registration/management
│   │   │   │   ├── guarantees.ts      # GO CRUD operations
│   │   │   │   ├── transfers.ts       # GO transfer operations
│   │   │   │   ├── conversions.ts     # Conversion (e→h) operations
│   │   │   │   ├── cancellations.ts   # Cancellation/claim operations
│   │   │   │   └── queries.ts         # Read-only query endpoints
│   │   │   ├── middleware/
│   │   │   │   ├── auth.ts            # JWT verification middleware
│   │   │   │   └── error.ts           # Error handling middleware
│   │   │   └── types/                 # Shared TypeScript types
│   │   │       └── index.ts
│   │   └── tests/
│   └── frontend/                      # React SPA
│       ├── package.json
│       ├── tsconfig.json
│       ├── vite.config.ts
│       ├── index.html
│       └── src/
│           ├── main.tsx
│           ├── App.tsx
│           ├── api/                    # API client (axios/fetch)
│           │   └── client.ts
│           ├── components/
│           │   ├── Layout.tsx
│           │   ├── Dashboard/
│           │   │   ├── IssuerDashboard.tsx
│           │   │   ├── ProducerDashboard.tsx
│           │   │   └── ConsumerDashboard.tsx
│           │   ├── Devices/
│           │   │   ├── DeviceList.tsx
│           │   │   └── RegisterDevice.tsx
│           │   ├── Guarantees/
│           │   │   ├── GOList.tsx
│           │   │   ├── GODetail.tsx
│           │   │   ├── CreateGO.tsx
│           │   │   └── TransferGO.tsx
│           │   ├── Conversion/
│           │   │   ├── BacklogView.tsx
│           │   │   └── ConvertGO.tsx
│           │   ├── Certificates/
│           │   │   ├── CancellationList.tsx
│           │   │   └── CancelGO.tsx
│           │   └── common/
│           │       ├── Table.tsx
│           │       ├── Modal.tsx
│           │       └── StatusBadge.tsx
│           ├── hooks/
│           │   └── useAuth.ts
│           ├── pages/
│           │   ├── Login.tsx
│           │   ├── Dashboard.tsx
│           │   ├── Devices.tsx
│           │   ├── Guarantees.tsx
│           │   ├── Transfers.tsx
│           │   └── Certificates.tsx
│           └── styles/
│               └── globals.css
│
├── testing/                           # Retained: Caliper benchmarks
│   ├── caliper-config-GO/
│   └── Test-results/
│
└── version1/                          # PRESERVED: Original prototype (read-only reference)
```

---

## 3. Tiered Network Architecture

### 3.1 Organization Roles

Instead of hardcoded organizations (buyer, eproducer, hproducer, issuer), the new network uses **role-based** organizations. Any number of organizations can join the network under one of three tiers:

| Tier | Role | MSP Convention | Description |
|------|------|---------------|-------------|
| **Tier 1** | Issuer | `issuer<N>MSP` | Registry operator. Oversees device registration, GO validation, auditing. At least one issuer is required. |
| **Tier 2** | Producer | `producer<N>MSP` | Energy producer. Can register multiple device types (electricity meters, hydrogen output meters, etc.). Each producer can produce GOs for **any carrier** they have registered devices for. |
| **Tier 3** | Consumer | `consumer<N>MSP` | Energy buyer/consumer. Can receive GOs via transfer, cancel GOs to claim renewable attributes. |

### 3.2 Channel Policy

```yaml
# Endorsement: Issuer + at least one other org
Endorsement:
  Type: Signature
  Rule: "OutOf(2, 'issuer1MSP.peer', 'producer1MSP.peer', 'consumer1MSP.peer')"
```

For operations that modify GOs:
- **Issuance**: Requires endorsement from the producer's peer AND the issuer's peer
- **Transfer**: Requires endorsement from the sender's peer AND one other org (e.g., issuer)
- **Cancellation**: Requires endorsement from the cancelling org's peer AND the issuer

### 3.3 Private Data Collections (Dynamic)

Collections follow the naming convention `privateDetails-{OrgMSPID}`:

```json
[
  {
    "name": "publicGOcollection",
    "policy": "OR(<all-org-members>)",
    "requiredPeerCount": 0,
    "maxPeerCount": 3,
    "blockToLive": 0,
    "memberOnlyRead": true,
    "memberOnlyWrite": true
  },
  {
    "name": "privateDetails-issuer1MSP",
    "policy": "OR('issuer1MSP.member')",
    "requiredPeerCount": 0,
    "maxPeerCount": 1,
    "blockToLive": 0,
    "memberOnlyRead": true,
    "memberOnlyWrite": false,
    "endorsementPolicy": {
      "signaturePolicy": "OR('issuer1MSP.member')"
    }
  },
  {
    "name": "privateDetails-producer1MSP",
    "policy": "OR('producer1MSP.member', 'issuer1MSP.member')",
    "requiredPeerCount": 0,
    "maxPeerCount": 2,
    "blockToLive": 0,
    "memberOnlyRead": true,
    "memberOnlyWrite": false,
    "endorsementPolicy": {
      "signaturePolicy": "OR('producer1MSP.member', 'issuer1MSP.member')"
    }
  }
]
```

When a new org joins the network, a new `privateDetails-{newOrgMSP}` collection is added via chaincode upgrade.

### 3.4 Device Registration (New Feature)

Devices (SmartMeters, OutputMeters) are now **on-chain assets**:

```go
type Device struct {
    DeviceID       string   `json:"deviceID"`
    DeviceType     string   `json:"deviceType"`     // "SmartMeter", "OutputMeter"
    OwnerOrgMSP    string   `json:"ownerOrgMSP"`
    EnergyCarriers []string `json:"energyCarriers"`  // ["electricity"], ["hydrogen"], or both
    Status         string   `json:"status"`          // "active", "revoked", "suspended"
    RegisteredBy   string   `json:"registeredBy"`    // Issuer identity
    RegisteredAt   int64    `json:"registeredAt"`
    Attributes     map[string]string `json:"attributes"` // maxEfficiency, emissionIntensity, etc.
}
```

This replaces the X.509 ABAC attributes for device validation. Benefits:
- Devices can be registered/revoked at runtime without re-enrolling certificates
- Multiple devices per organization
- Flexible attribute management

---

## 4. Chaincode Architecture (Object-Oriented)

### 4.1 Asset Hierarchy

```
BaseGO (interface)
├── ElectricityGO           — electricity guarantee
├── HydrogenGO              — hydrogen guarantee  
└── (extensible: BiogasGO, HeatGO, ...)

Device                      — metering device registration
├── SmartMeter              — electricity production meter
└── OutputMeter             — hydrogen/other output meter

Certificate (interface)
├── CancellationStatement   — proof of GO cancellation
└── ConsumptionDeclaration  — record of energy input consumption
```

### 4.2 Base GO Interface

```go
// assets/base.go
type BaseGOPublic struct {
    AssetID          string `json:"assetID"`
    CreationDateTime int64  `json:"creationDateTime"`
    GOType           string `json:"goType"`           // "electricity", "hydrogen", etc.
    ProducerOrgMSP   string `json:"producerOrgMSP"`
}

type BaseGOPrivate struct {
    AssetID                 string   `json:"assetID"`
    OwnerID                 string   `json:"ownerID"`
    CreationDateTime        int64    `json:"creationDateTime"`
    Quantity                float64  `json:"quantity"`         // MWh for electricity, kg for hydrogen
    QuantityUnit            string   `json:"quantityUnit"`     // "MWh", "kg"
    Emissions               float64  `json:"emissions"`
    ProductionMethod        string   `json:"productionMethod"`
    ConsumptionDeclarations []string `json:"consumptionDeclarations"`
    DeviceID                string   `json:"deviceID"`         // Which device produced this
}
```

### 4.3 Contract Registration (main.go)

```go
func main() {
    issuanceContract := new(contracts.IssuanceContract)
    issuanceContract.Name = "issuance"

    transferContract := new(contracts.TransferContract)
    transferContract.Name = "transfer"

    conversionContract := new(contracts.ConversionContract)
    conversionContract.Name = "conversion"

    cancellationContract := new(contracts.CancellationContract)
    cancellationContract.Name = "cancellation"

    queryContract := new(contracts.QueryContract)
    queryContract.Name = "query"

    deviceContract := new(contracts.DeviceContract)
    deviceContract.Name = "device"

    chaincode, err := contractapi.NewChaincode(
        issuanceContract,
        transferContract,
        conversionContract,
        cancellationContract,
        queryContract,
        deviceContract,
    )
    // ...
}
```

Clients invoke functions as `issuance:CreateElectricityGO`, `transfer:TransferGO`, etc.

### 4.4 ID Generation (v3.0 — Hash-Based Deterministic IDs)

v2.0 used on-chain sequential counters (`GetNextID`) that created a hot-key serialization bottleneck (MVCC_READ_CONFLICT). v3.0 replaced this with deterministic hash-based IDs:

```go
// assets/counter.go — v3.0
func GenerateID(ctx contractapi.TransactionContextInterface, prefix string, suffix int) (string, error) {
    txID := ctx.GetStub().GetTxID()
    raw := txID + "_" + strconv.Itoa(suffix)
    hash := sha256.Sum256([]byte(raw))
    return prefix + hex.EncodeToString(hash[:8]), nil
}
```

Each transaction ID is guaranteed unique by Fabric. The `suffix` parameter disambiguates multiple IDs within one transaction (e.g., cancellation creates a certificate + remainder GO). This approach:
- **Eliminates shared state**: No `GetState`/`PutState` on a shared counter key
- **Enables parallel writes**: Every transaction's read-write set is independent
- **Is deterministic**: All endorsing peers compute the same ID from the same txID
- **Is collision-resistant**: SHA-256 with 8-byte (64-bit) output gives negligible collision probability for practical volumes

Prefix constants (`PrefixDevice = "device_"`, `PrefixEGO = "eGO_"`, etc.) and range-end constants (`RangeEndDevice = "device_~"`) ensure consistent key formatting and efficient range queries.

The legacy `GetNextID()` function is retained but marked DEPRECATED for backward compatibility.

**Impact:** Write throughput improved from 0.8 TPS to 50.5 TPS (63× improvement) with 100% success rate.

### 4.5 Role-Based Access Control

```go
// access/roles.go
func GetOrgRole(ctx contractapi.TransactionContextInterface) (string, error) {
    mspID, err := ctx.GetClientIdentity().GetMSPID()
    if err != nil {
        return "", err
    }
    // Determine role from on-chain org registry or MSP naming convention
    role, err := ctx.GetStub().GetState("orgRole_" + mspID)
    if err != nil || role == nil {
        return "", fmt.Errorf("organization %s is not registered", mspID)
    }
    return string(role), nil
}

func RequireRole(ctx contractapi.TransactionContextInterface, requiredRole string) error {
    role, err := GetOrgRole(ctx)
    if err != nil {
        return err
    }
    if role != requiredRole {
        return fmt.Errorf("access denied: requires role %s, got %s", requiredRole, role)
    }
    return nil
}

func GetCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {
    mspID, err := ctx.GetClientIdentity().GetMSPID()
    if err != nil {
        return "", err
    }
    return "privateDetails-" + mspID, nil
}
```

### 4.6 Bug Fixes Summary

| # | Bug | Fix |
|---|-----|-----|
| 1 | In-memory counters reset on restart → ID collisions | On-chain persistent counters (§4.4) |
| 2 | Race condition in counter (read after unlock) | Fabric deterministic execution, single PutState per tx |
| 3 | IssuehGO overwrites accumulated emissions at end | Track emissions in running total; don't overwrite |
| 4 | IssuehGO skips final eGO (no else branch) | Add else clause for `AmountMWh < backlog.UsedMWh` |
| 5 | QueryPrivateeGOs remove() loop skips elements | Iterate backwards or use copy-on-write slice |
| 6 | No bounds checking → panics | Validate all inputs at function entry |
| 7 | `organization` attr vs `GetMSPID()` inconsistency | Use `GetMSPID()` exclusively for collection names |
| 8 | Remainder eGO CreationDateTime reset | Preserve original CreationDateTime on splits |
| 9 | ConsumptionDeclarations duplicated on hGO splits | Deep-copy declarations, only assign relevant ones |
| 10 | ConsumptionDeclarationHydrogen.DateTime is string | Unify to `int64` across all types |
| 11 | VerifyCancellationStatement hash never matches | Fix hash comparison to use correct key format |
| 12 | No access control on read functions | Add role-based read checks where appropriate |

---

## 5. Frontend Architecture

### 5.1 Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Frontend | React 18 + TypeScript | Single-page application |
| Styling | Tailwind CSS | Utility-first CSS |
| Build | Vite | Fast development server |
| Backend | Express.js + TypeScript | REST API layer |
| Fabric SDK | `@hyperledger/fabric-gateway` | Blockchain interaction |
| Auth | JWT tokens | Session management |

### 5.2 Backend API Routes

```
POST   /api/auth/login              # Authenticate with Fabric identity
POST   /api/auth/enroll             # Enroll new user via Fabric CA

POST   /api/devices                 # Register a new device
GET    /api/devices                 # List devices for current org
GET    /api/devices/:id             # Get device details
PUT    /api/devices/:id/revoke      # Revoke a device

POST   /api/guarantees/electricity  # Create electricity GO (from SmartMeter)
POST   /api/guarantees/hydrogen     # Create hydrogen GO (from OutputMeter)
GET    /api/guarantees              # List GOs for current org
GET    /api/guarantees/:id          # Get GO details (public)
GET    /api/guarantees/:id/private  # Get GO private details

POST   /api/transfers               # Transfer GO to another org
POST   /api/transfers/by-amount     # Transfer partial GO by amount

POST   /api/conversions/backlog     # Add to hydrogen backlog
POST   /api/conversions/issue       # Issue hydrogen GO from backlog

POST   /api/cancellations/electricity  # Cancel electricity GO (claim attributes)
POST   /api/cancellations/hydrogen     # Cancel hydrogen GO (claim attributes)
GET    /api/cancellations              # List cancellation statements
POST   /api/cancellations/verify       # Verify cancellation statement

GET    /api/queries/ego-list        # Get all current electricity GOs
GET    /api/queries/hgo-list        # Get all current hydrogen GOs
```

### 5.3 Frontend Pages

| Page | Role | Description |
|------|------|-------------|
| **Login** | All | Identity selection / enrollment |
| **Dashboard** | All | Role-specific overview with key metrics |
| **Devices** | Issuer, Producer | Register/manage metering devices |
| **My GOs** | Producer, Consumer | View owned GOs, initiate transfers |
| **Transfer** | Producer, Consumer | Transfer GOs between orgs |
| **Conversion** | Producer | Manage hydrogen backlog, issue converted GOs |
| **Certificates** | All | View cancellation statements and consumption declarations |
| **Audit** | Issuer | Cross-org GO tracking and verification |

### 5.4 Gateway Connection Pattern

```typescript
// backend/src/fabric/gateway.ts
import { connect, Contract, Identity, Signer, signers } from '@hyperledger/fabric-gateway';
import * as grpc from '@grpc/grpc-js';

export async function connectGateway(orgMSP: string, certPath: string, keyPath: string) {
    const tlsRootCert = await fs.readFile(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
    const client = new grpc.Client(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': peerHostOverride,
    });

    const credentials = await fs.readFile(certPath);
    const identity: Identity = { mspId: orgMSP, credentials };

    const privateKeyPem = await fs.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    const signer: Signer = signers.newPrivateKeySigner(privateKey);

    return connect({ client, identity, signer, hash: hash.sha256 });
}
```

---

## 6. Implementation Phases

### Phase 1: Chaincode Restructure
1. Create the `chaincode/` directory structure
2. Define asset types in `assets/`
3. Implement persistent counters
4. Implement role-based access control in `access/`
5. Migrate all 26 exported functions into domain contracts
6. Fix all 12 bugs
7. Unit tests for each contract

### Phase 2: Network Configuration
1. Create new `network/` with tiered configtx.yaml
2. Create Docker Compose templates for each org role
3. Implement `add-org.sh` for dynamic org onboarding
4. Update collection-config.json for role-based collections
5. Test network bring-up with 1 issuer + 1 producer + 1 consumer

### Phase 3: Frontend Application
1. Backend: Express + Fabric Gateway connection
2. Backend: REST API routes for all chaincode functions
3. Frontend: React app scaffold with Vite
4. Frontend: Login/auth flow with Fabric identity
5. Frontend: Role-specific dashboards
6. Frontend: GO lifecycle pages (create, transfer, convert, cancel)
7. Frontend: Device management pages
8. Frontend: Certificate viewing and verification

### Phase 4: Integration & Testing
1. End-to-end tests (frontend → backend → chaincode)
2. Update Caliper benchmarks for new chaincode
3. Documentation updates
4. CI/CD pipeline setup

---

## 7. Migration Notes

- The original `version1/` directory is **preserved intact** as a reference
- The new `chaincode/` replaces `version1/artifacts/Mychaincode/`
- The new `network/` replaces `version1/setup1/` and `version1/artifacts/channel/`
- The new `collections/` replaces `version1/artifacts/private-data-collections/`
- All existing DOCS.md files remain in `version1/`

---

## 8. v3.0 Architecture Changes

### 8.1 Performance Optimizations

| Change | Description | File(s) |
|--------|-------------|--------|
| **Hash-based IDs (ADR-001)** | Replace `GetNextID()` sequential counter with `GenerateID()` using SHA-256 of transaction ID | `assets/counter.go`, all contract files |
| **CouchDB Indexes (ADR-002)** | Composite indexes on `[OwnerID, GOType]` and `[GOType, CreationDateTime]` | `chaincode/META-INF/statedb/couchdb/indexes/` |
| **BatchTimeout (ADR-005)** | Reduced from 2s to 500ms for lower write latency | `network/configtx.yaml` |

### 8.2 Security Hardening

| Change | Description | File(s) |
|--------|-------------|--------|
| **Secure InitLedger (ADR-004)** | Caller MSP must match the `issuerMSP` argument — prevents cross-org privilege escalation | `contracts/device_mgmt.go` |

### 8.3 Data Architecture — ID Format Change

```
v2.0: eGO1, eGO2, eGO3, ...       (sequential, shared counter)
v3.0: eGO_a1b2c3d4e5f6a7b8, ...   (hash-based, contention-free)
```

Range queries use lexicographic bounds (e.g., `"eGO"` to `"eGO~"`) that capture both formats, enabling mixed-format operation during migration.

### 8.4 Block Configuration

| Parameter | v2.0 | v3.0 | Rationale |
|-----------|------|------|----------|
| BatchTimeout | 2s | **500ms** | Reduces write latency floor by 75% |
| MaxMessageCount | 10 | 10 | Unchanged |
| PreferredMaxBytes | 512 KB | 512 KB | Unchanged |

### 8.5 Benchmark Validation

All changes validated by Hyperledger Caliper v0.6.0 on a 16-vCPU Hetzner VM:

| Metric | v2.0 | v3.0 |
|--------|------|------|
| Write success rate | 10% | **100%** |
| Max write throughput | 0.8 TPS | **50.5 TPS** |
| Point read throughput | >2,000 TPS | >2,000 TPS |
| Write latency (serial) | 2.09s | **0.10s** |

Detailed results: see `testing/PERFORMANCE_REPORT.md`
