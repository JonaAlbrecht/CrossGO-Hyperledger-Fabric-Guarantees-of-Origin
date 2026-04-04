# GO Platform v7.0 — Architecture Diagrams

> Architecture diagrams for the Hyperledger Fabric Guarantee of Origin platform.
> Updated for v7.0 (cross-registry bridge, IoT device attestation, external data oracle).
> Diagrams use [Mermaid](https://mermaid.js.org/) syntax and render natively on GitHub.

---

## 1. System Architecture Overview

Shows the full three-tier stack: React frontend → Express backend → Hyperledger Fabric network with tiered organizations.

```mermaid
graph TB
    subgraph Frontend["🖥️ React Frontend · Vite + Tailwind CSS"]
        direction TB
        LP[LoginPage]
        DP[DashboardPage]
        DEV[DevicesPage]
        GP[GuaranteesPage]
        TP[TransfersPage]
        CP[ConversionsPage]
        CTP[CertificatesPage]
    end

    subgraph Backend["⚙️ Express.js Backend · TypeScript"]
        direction TB
        AUTH[Auth Middleware<br/>JWT + RBAC]
        subgraph Routes["REST API Routes"]
            R1[/api/auth]
            R2[/api/devices]
            R3[/api/guarantees]
            R4[/api/transfers]
            R5[/api/conversions]
            R6[/api/cancellations]
            R7[/api/queries]
        end
        GW["Fabric Gateway<br/>@hyperledger/fabric-gateway<br/>gRPC + TLS"]
    end

    subgraph Fabric["🔗 Hyperledger Fabric 2.x Network"]
        direction TB
        subgraph Orderers["Raft Orderer Cluster"]
            O1[orderer1<br/>:7050]
            O2[orderer2<br/>:8050]
            O3[orderer3<br/>:9050]
            O4[orderer4<br/>:10050]
        end
        subgraph Channel["GOPlatformChannel"]
            subgraph Issuer["🏛️ Tier 1 — Issuer"]
                IP[peer0.issuer1<br/>:7051]
                ICA[CA issuer<br/>:7054]
                IDB[(CouchDB)]
            end
            subgraph Producer["🏭 Tier 2 — Producer"]
                PP[peer0.producer1<br/>:9051]
                PCA[CA producer<br/>:8054]
                PDB[(CouchDB)]
            end
            subgraph Consumer["🏠 Tier 3 — Consumer"]
                COP[peer0.consumer1<br/>:11051]
                CCA[CA consumer<br/>:9054]
                CDB[(CouchDB)]
            end
        end
        subgraph Chaincode["📜 Go Chaincode · 10 Named Contracts"]
            CC1[issuance<br/>CreateElectricityGO<br/>CreateHydrogenGO]
            CC2[transfer<br/>TransferEGO<br/>TransferByAmount]
            CC3[conversion<br/>AddHydrogenToBacklog<br/>IssuehGO]
            CC4[cancellation<br/>ClaimRenewableAttributes<br/>VerifyCancellation]
            CC5[query<br/>GetGOsList<br/>ReadPublic / ReadPrivate]
            CC6[device<br/>RegisterDevice<br/>VerifyDeviceReading]
            CC7[admin<br/>GetVersion<br/>RegisterOrganization]
            CC8[biogas<br/>CreateBiogasGO<br/>CancelBiogasGO]
            CC9[bridge<br/>ExportGO / ImportGO<br/>ConfirmExport]
            CC10[oracle<br/>PublishGridData<br/>CrossReferenceGO]
        end
        subgraph PDC["🔒 Private Data Collections"]
            PUB[publicGOcollection<br/>All orgs]
            PRV1[privateDetails-issuer1MSP]
            PRV2[privateDetails-producer1MSP]
            PRV3[privateDetails-consumer1MSP]
        end
    end

    Frontend -->|HTTP/JSON| AUTH
    AUTH --> Routes
    Routes --> GW
    GW -->|gRPC + TLS| IP
    GW -->|gRPC + TLS| PP
    GW -->|gRPC + TLS| COP
    IP --> Chaincode
    PP --> Chaincode
    COP --> Chaincode
    Chaincode --> PDC
    IP --> IDB
    PP --> PDB
    COP --> CDB
    O1 --- O2
    O2 --- O3
    O3 --- O4

    classDef frontend fill:#dbeafe,stroke:#2563eb,color:#1e3a5f
    classDef backend fill:#fef3c7,stroke:#d97706,color:#78350f
    classDef fabric fill:#dcfce7,stroke:#16a34a,color:#14532d
    classDef chaincode fill:#f3e8ff,stroke:#7c3aed,color:#3b0764
    classDef pdc fill:#fce7f3,stroke:#db2777,color:#831843

    class LP,DP,DEV,GP,TP,CP,CTP frontend
    class AUTH,R1,R2,R3,R4,R5,R6,R7,GW backend
    class IP,PP,COP,ICA,PCA,CCA,IDB,PDB,CDB,O1,O2,O3,O4 fabric
    class CC1,CC2,CC3,CC4,CC5,CC6,CC7,CC8,CC9,CC10 chaincode
    class PUB,PRV1,PRV2,PRV3 pdc
```

---

## 2. Chaincode Internal Architecture — Package Dependencies

Shows how the 6 named contracts in `contracts/` depend on shared packages: `assets/`, `access/`, and `util/`.

```mermaid
graph LR
    subgraph main["main.go"]
        ENTRY["contractapi.NewChaincode()"]
    end

    subgraph contracts["contracts/"]
        IC["IssuanceContract<br/>━━━━━━━━━━━━━━━<br/>CreateElectricityGO()<br/>CreateHydrogenGO()"]
        TC["TransferContract<br/>━━━━━━━━━━━━━━━<br/>TransferEGO()<br/>TransferEGOByAmount()<br/>TransferHGOByAmount()"]
        CC["ConversionContract<br/>━━━━━━━━━━━━━━━<br/>AddHydrogenToBacklog()<br/>IssuehGO()<br/>QueryHydrogenBacklog()"]
        XC["CancellationContract<br/>━━━━━━━━━━━━━━━<br/>ClaimRenewableAttrsE()<br/>ClaimRenewableAttrsH()<br/>VerifyCancellation()<br/>SetGOEndorsementPolicy()"]
        QC["QueryContract<br/>━━━━━━━━━━━━━━━<br/>GetCurrentEGOsList()<br/>GetCurrentHGOsList()<br/>ReadPublic/Private()"]
        DC["DeviceContract<br/>━━━━━━━━━━━━━━━<br/>RegisterDevice()<br/>RevokeDevice()<br/>SuspendDevice()<br/>InitLedger()"]
    end

    subgraph assets["assets/"]
        EGO["ElectricityGO<br/>ElectricityGOPrivateDetails"]
        HGO["GreenHydrogenGO<br/>GreenHydrogenGOPrivateDetails<br/>GreenHydrogenGOBacklog"]
        CERT["CancellationStatementE/H<br/>ConsumptionDeclarationE/H"]
        DEV["Device<br/>SmartMeter │ OutputMeter"]
        CTR["Counter<br/>GenerateID() — SHA-256 hash"]
    end

    subgraph access["access/"]
        ROLES["roles.go<br/>━━━━━━━━━<br/>GetOrgRole()<br/>RequireRole()<br/>IsIssuer/IsProducer()"]
        ABAC["abac.go<br/>━━━━━━━━━<br/>GetAttribute()<br/>AssertAttribute()<br/>GetClientMSPID()"]
        COLL["collections.go<br/>━━━━━━━━━<br/>GetOwnCollection()<br/>GetCollectionForOrg()<br/>ValidateCollectionAccess()"]
    end

    subgraph util["util/"]
        SPLIT["split.go<br/>SplitElectricityGO()<br/>SplitHydrogenGO()<br/>Write/DeleteLedger()"]
        VALID["validate.go<br/>UnmarshalTransient()<br/>ValidatePositive()<br/>GetTimestamp()"]
        ITER["iterator.go<br/>ConstructEGOs()<br/>ConstructHGOs()"]
    end

    ENTRY --> IC
    ENTRY --> TC
    ENTRY --> CC
    ENTRY --> XC
    ENTRY --> QC
    ENTRY --> DC

    IC --> EGO
    IC --> CTR
    IC --> ROLES
    IC --> ABAC
    IC --> VALID

    TC --> EGO
    TC --> HGO
    TC --> SPLIT
    TC --> COLL

    CC --> HGO
    CC --> EGO
    CC --> CTR
    CC --> ITER

    XC --> CERT
    XC --> SPLIT
    XC --> CTR

    QC --> EGO
    QC --> HGO
    QC --> CERT
    QC --> COLL
    QC --> ITER

    DC --> DEV
    DC --> ROLES
    DC --> CTR

    classDef contractStyle fill:#f3e8ff,stroke:#7c3aed,color:#3b0764
    classDef assetStyle fill:#dbeafe,stroke:#2563eb,color:#1e3a5f
    classDef accessStyle fill:#fef3c7,stroke:#d97706,color:#78350f
    classDef utilStyle fill:#dcfce7,stroke:#16a34a,color:#14532d
    classDef mainStyle fill:#fce7f3,stroke:#db2777,color:#831843

    class IC,TC,CC,XC,QC,DC contractStyle
    class EGO,HGO,CERT,DEV,CTR assetStyle
    class ROLES,ABAC,COLL accessStyle
    class SPLIT,VALID,ITER utilStyle
    class ENTRY mainStyle
```

---

## 3. GO Lifecycle Sequence — Issuance → Transfer → Conversion → Cancellation → Verification

Shows the end-to-end flow of a Guarantee of Origin through the system, from metered data to verified certificate.

```mermaid
sequenceDiagram
    participant SM as SmartMeter<br/>(Device)
    participant Prod as Producer Peer
    participant Iss as Issuer Peer
    participant Ord as Raft Orderers
    participant PDC as Private Data<br/>Collections
    participant Cons as Consumer Peer

    Note over SM,Cons: ① ISSUANCE — Producer creates GO from metered data
    SM->>Prod: Electricity reading (transient)
    Prod->>Iss: Endorse: issuance:CreateElectricityGO
    Iss-->>Prod: Endorsement (validates device attrs)
    Prod->>Ord: Submit transaction
    Ord->>Prod: Block committed
    Prod->>PDC: Store private details in<br/>privateDetails-producer1MSP
    Note right of PDC: Public: AssetID, DateTime, GOType<br/>Private: Amount, Emissions, Method

    Note over SM,Cons: ② TRANSFER — Producer sends GO to Consumer
    Prod->>Cons: Endorse: transfer:TransferEGOByAmount
    Cons-->>Prod: Endorsement
    Prod->>Ord: Submit (with split if partial)
    Ord->>Cons: Block committed
    PDC-->>PDC: Move private data to<br/>privateDetails-consumer1MSP

    Note over SM,Cons: ③ CONVERSION — Producer converts electricity → hydrogen
    Prod->>Prod: conversion:AddHydrogenToBacklog
    Prod->>Iss: Endorse: conversion:IssuehGO
    Iss-->>Prod: Endorsement
    Note right of Prod: Consumes eGOs, creates<br/>ConsumptionDeclarations,<br/>mints new hGO

    Note over SM,Cons: ④ CANCELLATION — Consumer claims renewable attributes
    Cons->>Iss: Endorse: cancellation:ClaimRenewableAttributesElectricity
    Iss-->>Cons: Endorsement
    Cons->>Ord: Submit
    Note right of Cons: GO deleted, CancellationStatement<br/>created (immutable certificate)

    Note over SM,Cons: ⑤ VERIFICATION — Anyone verifies certificate
    Cons->>Iss: Evaluate: cancellation:VerifyCancellationStatement
    Iss-->>Cons: Verified ✓
```

---

## 4. Network Topology

```mermaid
graph TB
    subgraph OrdererCluster["Raft Consensus Cluster"]
        O1["orderer1.go-platform.com<br/>:7050"]
        O2["orderer2.go-platform.com<br/>:8050"]
        O3["orderer3.go-platform.com<br/>:9050"]
        O4["orderer4.go-platform.com<br/>:10050"]
        O1 <--> O2
        O2 <--> O3
        O3 <--> O4
        O4 <--> O1
    end

    subgraph CAs["Fabric Certificate Authorities"]
        CA1["ca-issuer :7054"]
        CA2["ca-producer :8054"]
        CA3["ca-consumer :9054"]
        CA4["ca-orderer :10054"]
    end

    subgraph IssuerOrg["Issuer Organization (Tier 1)"]
        IP1["peer0.issuer1<br/>:7051 + CouchDB :5984"]
    end

    subgraph ProducerOrg["Producer Organization (Tier 2)"]
        PP1["peer0.producer1<br/>:9051 + CouchDB :7984"]
    end

    subgraph ConsumerOrg["Consumer Organization (Tier 3)"]
        CP1["peer0.consumer1<br/>:11051 + CouchDB :9984"]
    end

    CA1 -.->|enrolls| IP1
    CA2 -.->|enrolls| PP1
    CA3 -.->|enrolls| CP1
    CA4 -.->|enrolls| OrdererCluster

    IP1 -->|gossip| PP1
    PP1 -->|gossip| CP1
    IP1 -->|gossip| CP1

    IP1 --> OrdererCluster
    PP1 --> OrdererCluster
    CP1 --> OrdererCluster

    classDef orderer fill:#fed7aa,stroke:#c2410c
    classDef ca fill:#e0e7ff,stroke:#4338ca
    classDef peer fill:#bbf7d0,stroke:#15803d

    class O1,O2,O3,O4 orderer
    class CA1,CA2,CA3,CA4 ca
    class IP1,PP1,CP1 peer
```

---

## 5. Data Architecture — Public vs. Private State

```mermaid
graph LR
    subgraph WorldState["World State (Public Ledger)"]
        WS1["eGO_a1b2c3d4... → {AssetID, CreationDateTime, GOType}"]
        WS2["hGO_f9e8d7c6... → {AssetID, CreationDateTime, GOType}"]
        WS3["device_000499e6... → {DeviceID, Type, OwnerOrg, Status}"]
        WS4["eCancel_11223344... → {statement data}"]
        WS5["orgRole_issuer1MSP → issuer"]
    end

    subgraph PDC1["privateDetails-producer1MSP"]
        PD1["eGO_a1b2c3d4... → {OwnerID, AmountMWh,<br/>Emissions, ProductionMethod, DeviceID}"]
        PD2["hGO_f9e8d7c6... → {OwnerID, Kilos,<br/>Emissions, InputData, DeviceID}"]
    end

    subgraph PDC2["privateDetails-consumer1MSP"]
        PD3["eGO_b3c4d5e6... → {OwnerID, AmountMWh, ...}"]
    end

    subgraph PDC3["publicGOcollection"]
        PD4["Shared metadata accessible<br/>to all channel members"]
    end

    subgraph Indexes["CouchDB Composite Indexes"]
        IDX1["indexOwner: [OwnerID, GOType]"]
        IDX2["indexGOType: [GOType, CreationDateTime]"]
    end

    WS1 -.->|private details in| PDC1
    WS2 -.->|private details in| PDC1
    PDC1 -.->|accelerated by| Indexes

    classDef public fill:#dcfce7,stroke:#16a34a
    classDef private fill:#fce7f3,stroke:#db2777
    classDef shared fill:#dbeafe,stroke:#2563eb
    classDef index fill:#fef3c7,stroke:#d97706

    class WS1,WS2,WS3,WS4,WS5 public
    class PD1,PD2,PD3 private
    class PD4 shared
    class IDX1,IDX2 index
```

---

## 6. v3.0 ID Generation Flow

Shows how v3.0 generates contention-free IDs from the transaction ID, eliminating the MVCC_READ_CONFLICT bottleneck.

```mermaid
sequenceDiagram
    participant Client
    participant Peer as Endorsing Peer
    participant CC as Chaincode
    participant Ledger as World State

    Note over Client,Ledger: v2.0 — Sequential Counter (MVCC_READ_CONFLICT)
    Client->>Peer: Submit RegisterDevice
    Peer->>CC: Invoke
    CC->>Ledger: GetState("counter_device") → 42
    CC->>Ledger: PutState("counter_device", 43)
    CC->>Ledger: PutState("device43", {...})
    Note right of Ledger: ❌ If another tx in same block<br/>also read counter=42 → MVCC conflict

    Note over Client,Ledger: v3.0 — Hash-Based ID (No shared state)
    Client->>Peer: Submit RegisterDevice
    Peer->>CC: Invoke (txID = abc123...)
    CC->>CC: SHA-256("abc123..._0")[:8] → "a1b2c3d4e5f6a7b8"
    CC->>Ledger: PutState("device_a1b2c3d4e5f6a7b8", {...})
    Note right of Ledger: ✅ No shared state read —<br/>parallel writes succeed
```

---

## 7. v4.0 GO Lifecycle State Machine (ADR-007 Tombstone)

Shows the lifecycle states introduced by the tombstone pattern.

```mermaid
stateDiagram-v2
    [*] --> Active: CreateGO / IssuehGO
    Active --> Cancelled: ClaimRenewableAttributes
    Active --> Transferred: TransferGO
    Active --> Split: Split (partial cancel/transfer)
    Split --> Active: Remainder GO (new ID)
    Split --> Cancelled: Taken portion (cancel)
    Split --> Transferred: Taken portion (transfer)

    Cancelled --> [*]: Audit record preserved
    Transferred --> [*]: Audit record preserved

    note right of Active
        Status = "active"
        Visible in list queries
    end note

    note right of Cancelled
        Status = "cancelled"
        Hidden from list queries
        Public state retained (tombstone)
    end note

    note right of Transferred
        Status = "transferred"
        Hidden from list queries
        Public state retained (tombstone)
    end note
```

---

## 8. v4.0 Selective Disclosure (ADR-009)

Shows the hash commitment scheme enabling quantity verification without collection access.

```mermaid
sequenceDiagram
    participant Producer
    participant Chaincode
    participant Ledger as World State
    participant Verifier

    Note over Producer,Verifier: Issuance — Commitment Generation
    Producer->>Chaincode: CreateElectricityGO(AmountMWh=100)
    Chaincode->>Chaincode: salt = SHA-256(txID + "_commitment_salt")[:16]
    Chaincode->>Chaincode: commitment = SHA-256("100" || salt)
    Chaincode->>Ledger: PutState(eGO_xxx, {QuantityCommitment: commitment})
    Chaincode->>Ledger: PutPrivateData(coll, eGO_xxx, {AmountMWh: 100, CommitmentSalt: salt})

    Note over Producer,Verifier: Verification — Selective Disclosure
    Producer->>Verifier: Disclose: AmountMWh=100, salt=abc123
    Verifier->>Chaincode: VerifyQuantityCommitment(eGO_xxx, 100, "abc123")
    Chaincode->>Ledger: GetState(eGO_xxx) → {QuantityCommitment: commitment}
    Chaincode->>Chaincode: SHA-256("100" || "abc123") == commitment?
    Chaincode->>Verifier: true ✅
```

---

## 9. v4.0 Paginated Query Flow (ADR-006)

```mermaid
sequenceDiagram
    participant Client
    participant Chaincode
    participant CouchDB

    Client->>Chaincode: GetCurrentEGOsListPaginated(pageSize=50, bookmark="")
    Chaincode->>CouchDB: GetStateByRangeWithPagination("eGO", "eGO~", 50, "")
    CouchDB-->>Chaincode: 50 records + bookmark="g1AAAABle..."
    Chaincode->>Chaincode: Filter: Status != "cancelled" && != "transferred"
    Chaincode-->>Client: {records: [...], bookmark: "g1AAAABle...", count: 50}

    Client->>Chaincode: GetCurrentEGOsListPaginated(50, "g1AAAABle...")
    Chaincode->>CouchDB: GetStateByRangeWithPagination("eGO", "eGO~", 50, "g1AAAABle...")
    CouchDB-->>Chaincode: 30 records + bookmark=""
    Chaincode-->>Client: {records: [...], bookmark: "", count: 30}
    Note over Client: bookmark="" → no more pages
```

---

## 10. v7.0 Contract Architecture (10 Namespaces)

```mermaid
graph TB
    subgraph Chaincode["golifecycle v7.0"]
        direction TB
        subgraph Domain["Domain Contracts"]
            ISS[issuance<br/>CreateElectricityGO<br/>CreateHydrogenGO]
            TRN[transfer<br/>TransferEGO<br/>TransferEGO/HGOByAmount]
            CNV[conversion<br/>AddHydrogenToBacklog<br/>IssuehGO]
            CAN[cancellation<br/>ClaimRenewableAttributes<br/>VerifyCancellation]
            BIO[biogas<br/>CreateBiogasGO<br/>CancelBiogasGO]
        end

        subgraph MarketIntegration["Market Integration Contracts (v7.0)"]
            BRG[bridge<br/>ExportGO / ImportGO<br/>ConfirmExport<br/>ListBridgeTransfersPaginated]
            ORC[oracle<br/>PublishGridData<br/>CrossReferenceGO<br/>ListGridDataPaginated]
        end

        subgraph Infrastructure["Infrastructure Contracts"]
            QRY[query<br/>18+ functions<br/>Point reads + Paginated lists<br/>Deprecation warnings (v6.0)]
            DEV[device<br/>Register/Revoke/List<br/>+Paginated<br/>VerifyDeviceReading (v7.0)<br/>SubmitSignedReading (v7.0)]
            ADM[admin<br/>GetVersion (7.0.0)<br/>RegisterOrganization<br/>9 supported APIs]
        end
    end

    subgraph Shared["Shared Packages"]
        ASS[assets/<br/>ElectricityGO, HydrogenGO<br/>BiogasGO, Device, Certificates<br/>BridgeTransfer, GridGenerationRecord]
        ACC[access/<br/>RBAC, ABAC<br/>Collections]
        UTL[util/<br/>Split, Validate, Events<br/>Iterator, ValidateCEN (v6.0)]
    end

    Domain --> ASS
    Domain --> UTL
    Domain --> ACC
    MarketIntegration --> ASS
    MarketIntegration --> UTL
    MarketIntegration --> ACC
    Infrastructure --> ASS
    Infrastructure --> ACC

    classDef domain fill:#dbeafe,stroke:#2563eb
    classDef market fill:#dcfce7,stroke:#16a34a
    classDef infra fill:#fef3c7,stroke:#d97706
    classDef shared fill:#f3e8ff,stroke:#7c3aed

    class ISS,TRN,CNV,CAN,BIO domain
    class BRG,ORC market
    class QRY,DEV,ADM infra
    class ASS,ACC,UTL shared
```

---

## 11. v5.0 CQRS Event Flow (ADR-016)

```mermaid
graph LR
    subgraph Fabric["Hyperledger Fabric"]
        CC[Chaincode<br/>EmitLifecycleEvent]
        BL[Block Listener<br/>Peer Event Service]
    end

    subgraph OffChain["Off-Chain Indexer"]
        EL[Event Listener<br/>SDK EventService]
        EP[Event Processor]
        DB[(Read Model<br/>PostgreSQL / ES)]
    end

    subgraph Clients["Client Applications"]
        API[REST API]
        UI[Frontend SPA]
    end

    CC -->|SetEvent| BL
    BL -->|gRPC stream| EL
    EL --> EP
    EP -->|INSERT/UPDATE| DB
    API -->|SELECT| DB
    UI --> API

    classDef fabric fill:#bbf7d0,stroke:#15803d
    classDef offchain fill:#fed7aa,stroke:#c2410c
    classDef client fill:#dbeafe,stroke:#2563eb

    class CC,BL fabric
    class EL,EP,DB offchain
    class API,UI client
```

---

## 12. v6.0 Selective Disclosure with Secure Salts (ADR-009 + ADR-017)

Shows the improved hash commitment scheme using `crypto/rand` instead of `txID`-derived salts.

```mermaid
sequenceDiagram
    participant Producer
    participant Chaincode
    participant CryptoRand as crypto/rand
    participant Ledger as World State
    participant PDC as Private Data
    participant Verifier

    Note over Producer,Verifier: Issuance — Secure Commitment (v6.0)
    Producer->>Chaincode: CreateElectricityGO(AmountMWh=100)
    Chaincode->>CryptoRand: Read 16 random bytes
    CryptoRand-->>Chaincode: salt = 0xA7F3...9B21 (128 bits)
    Chaincode->>Chaincode: commitment = SHA-256("100" || salt)
    Chaincode->>Ledger: PutState(eGO_xxx, {QuantityCommitment: commitment})
    Chaincode->>PDC: PutPrivateData(coll, eGO_xxx, {AmountMWh: 100, CommitmentSalt: salt})
    Note right of PDC: Salt is NOT derivable from txID<br/>Search space: 2^128 per guess

    Note over Producer,Verifier: Verification — Selective Disclosure
    Producer->>Verifier: Disclose: AmountMWh=100, salt=A7F3...9B21
    Verifier->>Chaincode: VerifyQuantityCommitment(eGO_xxx, 100, "A7F3...9B21")
    Chaincode->>Ledger: GetState(eGO_xxx) → {QuantityCommitment: commitment}
    Chaincode->>Chaincode: SHA-256("100" || "A7F3...9B21") == commitment?
    Chaincode->>Verifier: true ✅
```

---

## 13. v7.0 Cross-Registry Bridge Flow (ADR-024)

Shows the three-phase bridge protocol for GO export and import between registries.

```mermaid
sequenceDiagram
    participant Producer as Producer Org
    participant CC as BridgeContract
    participant Ledger as World State
    participant ExtReg as External Registry<br/>(e.g. AIB Hub)
    participant Importer as Importing Org

    Note over Producer,Importer: ① EXPORT — Producer initiates cross-registry transfer
    Producer->>CC: bridge:ExportGO(transient: BridgeExport)
    CC->>Ledger: GetState(eGO_xxx) → status="active"
    CC->>Ledger: PutState(eGO_xxx, {status: "transferred"})
    CC->>Ledger: PutState(bridge_abc123, {direction: "export", status: "pending"})
    CC-->>Producer: BridgeTransfer{transferId: "abc123", status: "pending"}

    Note over Producer,Importer: ② CONFIRM — External registry acknowledges receipt
    Producer->>ExtReg: Off-chain: send GO data + bridge transfer ID
    ExtReg-->>Producer: Acknowledgment received
    Producer->>CC: bridge:ConfirmExport("abc123")
    CC->>Ledger: PutState(bridge_abc123, {status: "confirmed", confirmedAt: now})

    Note over Producer,Importer: ③ IMPORT — Receiving org creates local GO from external source
    Importer->>CC: bridge:ImportGO(transient: BridgeImport)
    CC->>Ledger: PutState(eGO_newId, {new GO asset from import data})
    CC->>Ledger: PutState(bridge_def456, {direction: "import", status: "confirmed"})
    CC-->>Importer: BridgeTransfer{transferId: "def456", status: "confirmed"}
```

---

## 14. v7.0 IoT Device Attestation Flow (ADR-027)

Shows how physical smart meters cryptographically attest their readings on-chain.

```mermaid
sequenceDiagram
    participant HSM as Meter HSM<br/>(tamper-resistant)
    participant Meter as Smart Meter
    participant Prod as Producer Peer
    participant CC as DeviceContract
    participant Ledger as World State

    Note over HSM,Ledger: Setup — Register device with public key
    Prod->>CC: device:RegisterDevice(transient: {PublicKeyPEM: "-----BEGIN PUBLIC KEY-----..."})
    CC->>Ledger: PutState(device_xxx, {PublicKeyPEM: "...", status: "active"})

    Note over HSM,Ledger: Runtime — Meter signs and submits reading
    Meter->>HSM: Sign(readingJSON)
    HSM-->>Meter: ECDSA P-256 signature (ASN.1 DER)
    Meter->>Prod: Submit reading + signature
    Prod->>CC: device:SubmitSignedReading(deviceID, readingJSON, signatureBase64)
    CC->>Ledger: GetState(device_xxx) → {PublicKeyPEM: "..."}
    CC->>CC: Parse PEM → ECDSA public key
    CC->>CC: ecdsa.VerifyASN1(pubKey, SHA-256(readingJSON), signature)
    CC->>Ledger: PutState(reading_xxx, {verified reading data})
    CC-->>Prod: Success ✅
    Note right of CC: Signature verification ensures<br/>reading originated from registered meter,<br/>not fabricated by producer
```

---

## 15. v7.0 Oracle Cross-Reference Flow (ADR-029)

Shows how GO production claims are validated against external grid generation data.

```mermaid
sequenceDiagram
    participant ENTSO as ENTSO-E<br/>Transparency Platform
    participant Issuer as Issuer Org
    participant CC as OracleContract
    participant Ledger as World State
    participant Auditor

    Note over ENTSO,Auditor: ① Issuer publishes grid generation data
    ENTSO-->>Issuer: API response: generation data for DE-LU zone
    Issuer->>CC: oracle:PublishGridData(transient: GridData)
    CC->>CC: Validate EECS energy source code (regex: ^F\d{8}$)
    CC->>Ledger: PutState(oracle_rec123, {BiddingZone: "DE-LU", EnergySource: "F01010100", GenerationMW: 2000, ...})

    Note over ENTSO,Auditor: ② Cross-reference a GO against oracle data
    Auditor->>CC: oracle:CrossReferenceGO(eGO_xxx, oracle_rec123)
    CC->>Ledger: GetState(eGO_xxx) → {ProductionPeriodStart, ProductionPeriodEnd, EnergySource}
    CC->>Ledger: GetState(oracle_rec123) → {PeriodStart, PeriodEnd, EnergySource}
    CC->>CC: Check: GO period overlaps oracle period?
    CC->>CC: Check: Energy source codes match?
    CC-->>Auditor: true ✅ (GO production claim consistent with grid data)
```

---

## 16. v7.0 Verifiability Stack — End-to-End Provenance

Shows how v7.0 completes the full provenance chain from physical meter to cross-border transfer.

```mermaid
graph TB
    subgraph Physical["Physical Layer"]
        M[Smart Meter<br/>with HSM]
        SIG[ECDSA P-256<br/>Signature]
    end

    subgraph Attestation["Attestation Layer (v7.0)"]
        VDR[VerifyDeviceReading<br/>Signature verification]
        SSR[SubmitSignedReading<br/>Store verified reading]
    end

    subgraph Issuance["Issuance Layer (v3.0–v6.0)"]
        CEN[CEN-EN 16325<br/>Field Validation]
        COMMIT[Quantity Commitment<br/>SHA-256 + crypto/rand salt]
        SBE[State-Based<br/>Endorsement]
        TS[Timestamp<br/>Drift Guard]
    end

    subgraph Lifecycle["Lifecycle Layer (v4.0–v5.0)"]
        TOMB[Tombstone<br/>Status Tracking]
        CQRS[CQRS Events<br/>GO_CREATED, etc.]
    end

    subgraph CrossRef["Cross-Reference Layer (v7.0)"]
        ORC[Oracle Grid Data<br/>ENTSO-E records]
        XREF[CrossReferenceGO<br/>Period + source match]
    end

    subgraph Bridge["Cross-Registry Layer (v7.0)"]
        EXP[ExportGO<br/>Lock + bridge record]
        CONF[ConfirmExport<br/>Destination ack]
        IMP[ImportGO<br/>Create from external]
    end

    M --> SIG
    SIG --> VDR
    VDR --> SSR
    SSR --> CEN
    CEN --> COMMIT
    COMMIT --> SBE
    SBE --> TS
    TS --> TOMB
    TOMB --> CQRS
    CEN --> XREF
    ORC --> XREF
    TOMB --> EXP
    EXP --> CONF
    CONF --> IMP

    classDef physical fill:#fce7f3,stroke:#db2777
    classDef attest fill:#fed7aa,stroke:#c2410c
    classDef issue fill:#dbeafe,stroke:#2563eb
    classDef lifecycle fill:#dcfce7,stroke:#16a34a
    classDef crossref fill:#fef3c7,stroke:#d97706
    classDef bridge fill:#f3e8ff,stroke:#7c3aed

    class M,SIG physical
    class VDR,SSR attest
    class CEN,COMMIT,SBE,TS issue
    class TOMB,CQRS lifecycle
    class ORC,XREF crossref
    class EXP,CONF,IMP bridge
```
