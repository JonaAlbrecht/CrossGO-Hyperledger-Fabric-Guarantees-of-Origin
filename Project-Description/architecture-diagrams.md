# GO Platform v3 — Architecture Diagrams

> Architecture diagrams for the Hyperledger Fabric Guarantee of Origin platform.
> Updated for v3.0 (hash-based IDs, CouchDB indexes, reduced BatchTimeout).
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
        subgraph Chaincode["📜 Go Chaincode · 6 Named Contracts"]
            CC1[issuance<br/>CreateElectricityGO<br/>CreateHydrogenGO]
            CC2[transfer<br/>TransferEGO<br/>TransferByAmount]
            CC3[conversion<br/>AddHydrogenToBacklog<br/>IssuehGO]
            CC4[cancellation<br/>ClaimRenewableAttributes<br/>VerifyCancellation]
            CC5[query<br/>GetGOsList<br/>ReadPublic / ReadPrivate]
            CC6[device<br/>RegisterDevice<br/>RevokeDevice]
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
    class CC1,CC2,CC3,CC4,CC5,CC6 chaincode
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
