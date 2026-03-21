# Mychaincode — GO Network Smart Contract

This is the core **Hyperledger Fabric chaincode** (smart contract) written in Go, implementing the full lifecycle of Guarantees of Origin for electricity and hydrogen.

## Entry Point
- **`main.go`** — Initializes the `SmartContract` struct via Fabric's `contractapi.NewChaincode()` and starts the chaincode process.

## Core Logic: `GOnetwork/conversion.go` (~2,350 lines)

### Data Model

#### Public State (on-chain, visible to all channel members)
| Struct | Key Pattern | Fields |
|--------|------------|--------|
| `ElectricityGO` | `eGO1`, `eGO2`... | AssetID, CreationDateTime, GOType ("Electricity") |
| `GreenHydrogenGO` | `hGO1`, `hGO2`... | AssetID, CreationDateTime, GOType ("Hydrogen") |
| `GreenHydrogenGObacklog` | `hydrogenbacklog` | Backlogkey, GOType |

#### Private State (org-specific collections)
| Struct | Collection | Key Fields |
|--------|-----------|------------|
| `ElectricityGOprivatedetails` | `privateDetails-{orgMSP}` | AssetID, OwnerID, AmountMWh, Emissions, ElectricityProductionMethod, ConsumptionDeclarations |
| `GreenHydrogenGOprivatedetails` | `privateDetails-{orgMSP}` | AssetID, OwnerID, Kilosproduced, EmissionsHydrogen, HydrogenProductionMethod, InputEmissions, UsedMWh, ElectricityProductionMethod[] |
| `CancellationstatementElectricity` | `privateDetails-{orgMSP}` | eCancellationkey, CancellationTime, OwnerID, AmountMWh, Emissions |
| `CancellationstatementHydrogen` | `privateDetails-{orgMSP}` | hCancellationkey, CancellationTime, OwnerID, Kilosproduced, EmissionsHydrogen |
| `ConsumptionDeclarationElectricity` | `privateDetails-{orgMSP}` | Consumptionkey, CancelledGOID, ConsumptionDateTime, AmountMWh |
| `ConsumptionDeclarationHydrogen` | `privateDetails-{orgMSP}` | Consumptionkey, CancelledGOID, ConsumptionDateTime, Kilosproduced |

### Chaincode Functions

#### Creation & Metering
| Function | Access Control | Description |
|----------|---------------|-------------|
| `CreateElectricityGO` | `electricitytrustedDevice=true` (SmartMeter) | Creates an eGO from metering data. Validates efficiency, emission intensity, and technology type against X.509 certificate attributes. |
| `AddHydrogentoBacklog` | `hydrogentrustedDevice=true` (OutputMeter) | Accumulates hydrogen output data into a running backlog for later hGO issuance. |

#### Transfer
| Function | Access Control | Description |
|----------|---------------|-------------|
| `TransfereGO` | `TrustedUser=true` | Transfers a single eGO between private data collections. |
| `TransfereGObyAmount` | `TrustedUser=true` | Transfers eGOs by MWh amount; handles partial splitting with proportional emission/amount allocation. |
| `TansferhGObyAmount` | `TrustedUser=true` | Transfers hGOs by kilos amount with partial splitting. |

#### Conversion (Core Business Logic)
| Function | Access Control | Description |
|----------|---------------|-------------|
| `IssuehGO` | `hydrogentrustedUser=true` | **Key conversion function**: Cancels consumed eGOs, transcribes their attributes (emissions, production methods) onto a new hydrogen GO, and creates consumption declarations. Implements the cross-carrier attribute inheritance model. |

#### Cancellation & Claims
| Function | Access Control | Description |
|----------|---------------|-------------|
| `ClaimRenewableattributesElectricity` | Any member | Cancels eGOs and issues cancellation statements; supports partial cancellation. |
| `ClaimRenewableattributesHydrogen` | Any member | Cancels hGOs and issues hydrogen cancellation statements. |

#### Query & Read
| Function | Description |
|----------|-------------|
| `ReadPubliceGO` / `ReadPublichGO` | Read public GO metadata |
| `ReadPrivatefromCollectioneGO` / `ReadPrivatefromCollectionhGO` | Read private GO details from a specific collection |
| `GetcurrenteGOsList` / `GetcurrenthGOsList` | Range query on all existing GO keys |
| `QueryPrivateeGOsbyAmountMWh` / `QueryPrivatehGOsbyAmount` | Find GOs meeting a needed amount, sorted by earliest expiry |
| `QueryHydrogenBacklog` | Read the hydrogen production backlog |
| `ReadCancelstatement*` / `ReadConsumptionDeclaration*` | Read cancellation/consumption statements for electricity or hydrogen |
| `VerifyCancellationStatement` | Hash-based verification of cancellation statements against on-chain private data hashes |

### Key Design Decisions
- **GO Expiry**: eGOs expire after 3,600s (1 hour), configurable to 900s (15 min). 300s safety margin for transfers.
- **Partial Splitting**: When transferring/cancelling by amount, the last GO is split proportionally (amount, emissions). A "split" marker is added to consumption declarations.
- **Attribute Inheritance**: During conversion (`IssuehGO`), hydrogen GOs inherit input emissions, electricity production methods, and consumption declarations from the cancelled eGOs.
- **Thread-Safe Counters**: Mutex-protected counters for eGO, hGO, cancellation, and consumption declaration IDs.
- **ABAC**: Attribute-Based Access Control via X.509 certificate attributes for metering devices and trusted users.
