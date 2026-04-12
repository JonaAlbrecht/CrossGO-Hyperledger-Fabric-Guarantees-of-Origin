// Shared types mirroring the backend — used across all frontend components
// v9.0 — unified energy carrier model (electricity, hydrogen, biogas, heating & cooling)

export type EnergyCarrier = 'electricity' | 'hydrogen' | 'biogas' | 'heating_cooling';

export const ENERGY_CARRIERS: { value: EnergyCarrier; label: string; color: string; bgColor: string }[] = [
    { value: 'electricity', label: 'Electricity', color: 'text-yellow-700', bgColor: 'bg-yellow-100' },
    { value: 'hydrogen', label: 'Hydrogen', color: 'text-blue-700', bgColor: 'bg-blue-100' },
    { value: 'biogas', label: 'Biogas', color: 'text-green-700', bgColor: 'bg-green-100' },
    { value: 'heating_cooling', label: 'Heating & Cooling', color: 'text-orange-700', bgColor: 'bg-orange-100' },
];

export function carrierStyle(carrier: string) {
    return ENERGY_CARRIERS.find((c) => c.value === carrier) ?? ENERGY_CARRIERS[0];
}

export interface UserSession {
    token: string;
    mspId: string;
    orgName: string;
    userName: string;
    role: 'issuer' | 'producer' | 'consumer';
}

// Unified GO type — works for all energy carriers
export interface GuaranteeOfOrigin {
    AssetID: string;
    CreationDateTime: number;
    GOType: string; // energy carrier
    Status?: string;
}

export interface GOPrivateDetails {
    AssetID: string;
    OwnerID: string;
    Amount: number;       // MWh for electricity/biogas/heating_cooling, kg for hydrogen
    Emissions: number;
    ProductionMethod: string;
    DeviceID: string;
    InputEmissions?: number;
    InputAmount?: number;  // consumed input (e.g. MWh electricity for hydrogen)
    InputCarrier?: string; // source carrier for conversion
    ConsumptionDeclarations?: string[];
}

// Legacy types kept for backward compat with existing chaincode responses
export interface ElectricityGO {
    AssetID: string;
    CreationDateTime: number;
    GOType: string;
}

export interface HydrogenGO {
    AssetID: string;
    CreationDateTime: number;
    GOType: string;
}

export interface Device {
    deviceID: string;
    deviceType: string;
    ownerOrgMSP: string;
    energyCarriers: string[];
    status: string;
    registeredBy: string;
    registeredAt: number;
    attributes: Record<string, string>;
}

export interface CancellationStatement {
    AssetID: string;
    DateTime: number;
    BeneficiaryID: string;
    Amount?: number;
    Emissions?: number;
    EnergyCarrier?: string;
}

// Organization display names — v9 human-friendly names
export const ORG_DISPLAY: Record<string, string> = {
    issuer1: 'German Issuing Authority (UBA)',
    eproducer1: 'Alpha WindFarm GmbH',
    hproducer1: 'Beta Electrolyser B.V.',
    buyer1: 'Gamma-Town EnergySupplier Ltd',
    issuer1MSP: 'German Issuing Authority (UBA)',
    eproducer1MSP: 'Alpha WindFarm GmbH',
    hproducer1MSP: 'Beta Electrolyser B.V.',
    buyer1MSP: 'Gamma-Town EnergySupplier Ltd',
};

export function orgDisplayName(orgId: string): string {
    return ORG_DISPLAY[orgId] ?? orgId;
}
