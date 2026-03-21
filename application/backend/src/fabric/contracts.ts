// Contract accessor helpers — one getter per named contract
import { Contract } from '@hyperledger/fabric-gateway';
import { FabricConnection } from './gateway';

export const CONTRACT_NAMES = {
    ISSUANCE: 'issuance',
    TRANSFER: 'transfer',
    CONVERSION: 'conversion',
    CANCELLATION: 'cancellation',
    QUERY: 'query',
    DEVICE: 'device',
} as const;

export function getIssuanceContract(conn: FabricConnection): Contract {
    return conn.contract(CONTRACT_NAMES.ISSUANCE);
}

export function getTransferContract(conn: FabricConnection): Contract {
    return conn.contract(CONTRACT_NAMES.TRANSFER);
}

export function getConversionContract(conn: FabricConnection): Contract {
    return conn.contract(CONTRACT_NAMES.CONVERSION);
}

export function getCancellationContract(conn: FabricConnection): Contract {
    return conn.contract(CONTRACT_NAMES.CANCELLATION);
}

export function getQueryContract(conn: FabricConnection): Contract {
    return conn.contract(CONTRACT_NAMES.QUERY);
}

export function getDeviceContract(conn: FabricConnection): Contract {
    return conn.contract(CONTRACT_NAMES.DEVICE);
}
