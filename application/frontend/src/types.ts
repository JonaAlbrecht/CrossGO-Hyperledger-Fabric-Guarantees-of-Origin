// Shared types mirroring the backend — used across all frontend components

export interface UserSession {
    token: string;
    mspId: string;
    orgName: string;
    userName: string;
    role: 'issuer' | 'producer' | 'consumer';
}

export interface ElectricityGO {
    AssetID: string;
    CreationDateTime: number;
    GOType: string;
}

export interface ElectricityGOPrivate {
    AssetID: string;
    OwnerID: string;
    AmountMWh: number;
    Emissions: number;
    ElectricityProductionMethod: string;
    DeviceID: string;
}

export interface HydrogenGO {
    AssetID: string;
    CreationDateTime: number;
    GOType: string;
}

export interface HydrogenGOPrivate {
    AssetID: string;
    OwnerID: string;
    Kilosproduced: number;
    EmissionsHydrogen: number;
    HydrogenProductionMethod: string;
    InputEmissions: number;
    UsedMWh: number;
    ElectricityProductionMethod: string[];
    DeviceID: string;
}

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

export interface CancellationStatement {
    AssetID: string;
    DateTime: number;
    BeneficiaryID: string;
    AmountMWh?: number;
    Kilosproduced?: number;
    Emissions?: number;
    EmissionsHydrogen?: number;
}
