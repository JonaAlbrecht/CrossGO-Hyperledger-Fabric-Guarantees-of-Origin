// Fabric Gateway connection using @hyperledger/fabric-gateway
import * as grpc from '@grpc/grpc-js';
import {
    connect,
    Contract,
    Gateway,
    Identity,
    Signer,
    signers,
    hash,
} from '@hyperledger/fabric-gateway';
import * as crypto from 'crypto';
import * as fs from 'fs/promises';
import * as path from 'path';
import { logger } from '../middleware/logger';

const CHANNEL_NAME = process.env.CHANNEL_NAME ?? 'goplatformchannel';
const CHAINCODE_NAME = process.env.CHAINCODE_NAME ?? 'golifecycle';

// Peer endpoints per org role
const PEER_ENDPOINTS: Record<string, { address: string; tlsCert: string; hostOverride: string }> = {
    issuer1MSP: {
        address: process.env.ISSUER_PEER_ENDPOINT ?? 'localhost:7051',
        tlsCert: process.env.ISSUER_TLS_CERT ?? '',
        hostOverride: 'peer0.issuer1.go-platform.com',
    },
    producer1MSP: {
        address: process.env.PRODUCER_PEER_ENDPOINT ?? 'localhost:9051',
        tlsCert: process.env.PRODUCER_TLS_CERT ?? '',
        hostOverride: 'peer0.producer1.go-platform.com',
    },
    consumer1MSP: {
        address: process.env.CONSUMER_PEER_ENDPOINT ?? 'localhost:11051',
        tlsCert: process.env.CONSUMER_TLS_CERT ?? '',
        hostOverride: 'peer0.consumer1.go-platform.com',
    },
};

export interface FabricConnection {
    gateway: Gateway;
    contract: (contractName: string) => Contract;
    close: () => void;
}

export async function connectToFabric(
    mspId: string,
    certPath: string,
    keyPath: string
): Promise<FabricConnection> {
    const peerConfig = PEER_ENDPOINTS[mspId];
    if (!peerConfig) {
        throw new Error(`Unknown MSP: ${mspId}. Expected one of: ${Object.keys(PEER_ENDPOINTS).join(', ')}`);
    }

    // Create gRPC client connection
    const tlsRootCert = await fs.readFile(peerConfig.tlsCert);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
    const grpcClient = new grpc.Client(peerConfig.address, tlsCredentials, {
        'grpc.ssl_target_name_override': peerConfig.hostOverride,
    });

    // Load user identity
    const certPem = await fs.readFile(certPath);
    const identity: Identity = { mspId, credentials: certPem };

    // Load user private key for signing
    const keyPem = await fs.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(keyPem);
    const signer: Signer = signers.newPrivateKeySigner(privateKey);

    const gateway = connect({
        client: grpcClient,
        identity,
        signer,
        hash: hash.sha256,
        evaluateOptions: () => ({ deadline: Date.now() + 5000 }),
        endorseOptions: () => ({ deadline: Date.now() + 15000 }),
        submitOptions: () => ({ deadline: Date.now() + 5000 }),
        commitStatusOptions: () => ({ deadline: Date.now() + 60000 }),
    });

    const network = gateway.getNetwork(CHANNEL_NAME);

    logger.info(`Connected to Fabric gateway as ${mspId}`);

    return {
        gateway,
        contract: (contractName: string) =>
            network.getContract(CHAINCODE_NAME, contractName),
        close: () => {
            gateway.close();
            grpcClient.close();
        },
    };
}

// Resolve crypto material paths for an org/user
export function getCryptoPath(orgName: string, userName: string): { certPath: string; keyPath: string } {
    const cryptoBase = process.env.CRYPTO_PATH ?? path.resolve(__dirname, '..', '..', '..', 'network', 'organizations');
    const orgDomain = `${orgName}.go-platform.com`;
    return {
        certPath: path.join(cryptoBase, 'peerOrganizations', orgDomain, 'users', `${userName}@${orgDomain}`, 'msp', 'signcerts', 'cert.pem'),
        keyPath: path.join(cryptoBase, 'peerOrganizations', orgDomain, 'users', `${userName}@${orgDomain}`, 'msp', 'keystore', 'priv_sk'),
    };
}
