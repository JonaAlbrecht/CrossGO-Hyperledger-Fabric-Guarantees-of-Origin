// Shared TypeScript types mirroring chaincode Go structs

// ---- Electricity GOs ----
export interface ElectricityGO {
    AssetID: string;
    CreationDateTime: number;
    GOType: string;
}

export interface ElectricityGOPrivateDetails {
    AssetID: string;
    OwnerID: string;
    AmountMWh: number;
    Emissions: number;
    ElectricityProductionMethod: string;
    ConsumptionDeclarations: ConsumptionDeclarationElectricity[];
    DeviceID: string;
}

// ---- Hydrogen GOs ----
export interface GreenHydrogenGO {
    AssetID: string;
    CreationDateTime: number;
    GOType: string;
}

export interface GreenHydrogenGOPrivateDetails {
    AssetID: string;
    OwnerID: string;
    Kilosproduced: number;
    EmissionsHydrogen: number;
    HydrogenProductionMethod: string;
    InputEmissions: number;
    UsedMWh: number;
    ElectricityProductionMethod: string[];
    ConsumptionDeclarations: ConsumptionDeclarationHydrogen[];
    DeviceID: string;
}

// ---- Certificates ----
export interface CancellationStatementElectricity {
    AssetID: string;
    AmountMWh: number;
    Emissions: number;
    ElectricityProductionMethod: string;
    DateTime: number;
    BeneficiaryID: string;
}

export interface CancellationStatementHydrogen {
    AssetID: string;
    Kilosproduced: number;
    EmissionsHydrogen: number;
    HydrogenProductionMethod: string;
    InputEmissions: number;
    UsedMWh: number;
    ElectricityProductionMethod: string[];
    DateTime: number;
    BeneficiaryID: string;
}

export interface ConsumptionDeclarationElectricity {
    AssetID: string;
    AmountMWh: number;
    Emissions: number;
    ElectricityProductionMethod: string;
    DateTime: number;
    ConsumerID: string;
}

export interface ConsumptionDeclarationHydrogen {
    AssetID: string;
    Kilosproduced: number;
    EmissionsHydrogen: number;
    HydrogenProductionMethod: string;
    InputEmissions: number;
    UsedMWh: number;
    ElectricityProductionMethod: string[];
    DateTime: number;
    ConsumerID: string;
}

// ---- Devices ----
export interface Device {
    DeviceID: string;
    DeviceType: string;
    OwnerOrgMSP: string;
    EnergyCarriers: string[];
    Status: string;
    RegisteredBy: string;
    RegisteredAt: number;
    Attributes: Record<string, string>;
}

// ---- Auth ----
export interface UserIdentity {
    mspId: string;
    orgName: string;
    userName: string;
    role: string;
}

export interface JWTPayload {
    mspId: string;
    orgName: string;
    userName: string;
    role: string;
    iat?: number;
    exp?: number;
}

// ---- API request/response helpers ----
export interface ApiError {
    error: string;
    details?: string;
}

export interface TransferRequest {
    goAssetID: string;
    recipientMSP: string;
}

export interface TransferByAmountRequest {
    recipientMSP: string;
    amountMWh: number;
}

export interface BacklogRequest {
    backlogID: string;
    kilosHydrogen: number;
    hydrogenProductionMethod: string;
    mwhElectricity: number;
}

export interface CancellationRequest {
    goAssetID: string;
    amountMWh?: number;
    kilos?: number;
}
