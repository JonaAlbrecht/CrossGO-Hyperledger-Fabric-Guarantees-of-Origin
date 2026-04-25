// Bridge routes — cross-channel GO transfer with tri-party endorsement (v10.1)
// v10.1: Updated to support owner consent requirement for cross-channel bridge transfers
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { logger } from '../middleware/logger';

function safeParse(data: Uint8Array, fallback: any = []) {
    const str = new TextDecoder().decode(data).trim();
    return str ? JSON.parse(str) : fallback;
}

const router = Router();
router.use(authenticate);

async function getFabricConn(req: Request) {
    const { mspId, orgName, userName } = req.user!;
    const { certPath, keyPath } = getCryptoPath(orgName, userName);
    return connectToFabric(mspId, certPath, keyPath);
}

// POST /api/bridge/verify — verify a cross-channel bridge transfer proof
router.post('/verify', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { bridgeTransactionID } = req.body;
        if (!bridgeTransactionID) {
            res.status(400).json({ error: 'bridgeTransactionID is required' });
            return;
        }

        // Get bridge contract  
        const bridgeContract = conn.contract('BridgeContract');
        const result = await bridgeContract.evaluate('VerifyBridgeTransfer', bridgeTransactionID);

        res.json({ verified: true, data: safeParse(result, null) });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`VerifyBridgeTransfer failed: ${message}`);
        res.status(500).json({ error: message, verified: false });
    } finally {
        conn.close();
    }
});

// POST /api/bridge/initiate — initiate a cross-channel bridge transfer (issuer only)
router.post('/initiate', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { goAssetID, targetChannel, targetIssuerMSP } = req.body;
        if (!goAssetID || !targetChannel || !targetIssuerMSP) {
            res.status(400).json({ error: 'goAssetID, targetChannel, and targetIssuerMSP are required' });
            return;
        }

        const bridgeContract = conn.contract('BridgeContract');
        const transientData = Buffer.from(JSON.stringify({
            AssetID: goAssetID,
            TargetChannel: targetChannel,
            TargetIssuerMSP: targetIssuerMSP,
        }));

        await bridgeContract.submit('InitiateBridgeTransfer', {
            transientData: { BridgeInput: transientData },
        });

        res.status(201).json({ message: 'Bridge transfer initiated — awaiting dual-issuer consensus' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`InitiateBridgeTransfer failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/bridge/approve — approve a pending bridge transfer (target channel issuer)
router.post('/approve', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { bridgeTransactionID } = req.body;
        if (!bridgeTransactionID) {
            res.status(400).json({ error: 'bridgeTransactionID is required' });
            return;
        }

        const bridgeContract = conn.contract('BridgeContract');
        await bridgeContract.submit('ApproveBridgeTransfer', bridgeTransactionID);

        res.json({ message: 'Bridge transfer approved — dual-issuer consensus achieved' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ApproveBridgeTransfer failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// ============================================================================
// v10.1 Tri-Party Endorsement Routes (Owner + Source Issuer + Dest Issuer)
// ============================================================================

// POST /api/bridge/lock — Phase 1: Lock a GO on source channel (requires owner + issuer endorsement)
router.post('/lock', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { goAssetID, destinationChannel, ownerMSP } = req.body;
        if (!goAssetID || !destinationChannel || !ownerMSP) {
            res.status(400).json({ error: 'goAssetID, destinationChannel, and ownerMSP are required' });
            return;
        }

        const { mspId } = req.user!;

        const bridgeContract = conn.contract('bridge');
        const transientData = Buffer.from(JSON.stringify({
            GOAssetID: goAssetID,
            DestinationChannel: destinationChannel,
            OwnerMSP: ownerMSP,
        }));

        // v10.1: Tri-party endorsement — requires both owner and issuer to sign
        const result = await bridgeContract.submit('LockGO', {
            transientData: { BridgeLock: transientData },
            endorsingOrganizations: [ownerMSP, mspId], // Owner + source issuer
        });

        const lock = safeParse(result, null);
        logger.info(`GO ${goAssetID} locked for bridge transfer by ${mspId} with owner consent from ${ownerMSP}`);
        res.status(201).json({ 
            message: 'GO locked for cross-channel bridge — awaiting destination issuer approval',
            lock,
        });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`LockGO failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/bridge/mint — Phase 2: Mint GO on destination channel (issuer only)
router.post('/mint', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { 
            sourceChannel, sourceLockID, sourceGOAssetID, lockReceiptHash, 
            ownerMSP, goType, amountMWh, countryOfOrigin, energySource 
        } = req.body;

        if (!sourceChannel || !sourceLockID || !sourceGOAssetID || !lockReceiptHash || !ownerMSP || !goType) {
            res.status(400).json({ 
                error: 'sourceChannel, sourceLockID, sourceGOAssetID, lockReceiptHash, ownerMSP, and goType are required' 
            });
            return;
        }

        const bridgeContract = conn.contract('bridge');
        const transientData = Buffer.from(JSON.stringify({
            SourceChannel: sourceChannel,
            SourceLockID: sourceLockID,
            SourceGOAssetID: sourceGOAssetID,
            LockReceiptHash: lockReceiptHash,
            OwnerMSP: ownerMSP, // v10.1: Owner MSP for cross-channel verification
            GOType: goType,
            AmountMWh: amountMWh || 0,
            CountryOfOrigin: countryOfOrigin || '',
            EnergySource: energySource || '',
        }));

        const result = await bridgeContract.submit('MintFromBridge', {
            transientData: { BridgeMint: transientData },
        });

        const mint = safeParse(result, null);
        logger.info(`Minted GO on destination channel from lock ${sourceLockID} with owner ${ownerMSP}`);
        res.status(201).json({ 
            message: 'GO minted on destination channel — awaiting finalization on source channel',
            mint,
        });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`MintFromBridge failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/bridge/finalize — Phase 3: Finalize lock on source channel (requires owner + issuer endorsement)
router.post('/finalize', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { lockID, mintedAssetID, ownerMSP } = req.body;
        if (!lockID || !mintedAssetID || !ownerMSP) {
            res.status(400).json({ error: 'lockID, mintedAssetID, and ownerMSP are required' });
            return;
        }

        const { mspId } = req.user!;

        const bridgeContract = conn.contract('bridge');
        
        // v10.1: Tri-party endorsement — requires both owner and issuer to sign finalization
        await bridgeContract.submit('FinalizeLock', {
            arguments: [lockID, mintedAssetID],
            endorsingOrganizations: [ownerMSP, mspId], // Owner + source issuer
        });

        logger.info(`Bridge lock ${lockID} finalized by ${mspId} with owner consent from ${ownerMSP}`);
        res.json({ message: 'Bridge transfer finalized — GO successfully bridged across channels' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`FinalizeLock failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/bridge/locks — List cross-channel locks (paginated)
router.get('/locks', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const pageSize = parseInt(req.query.pageSize as string) || 50;
        const bookmark = (req.query.bookmark as string) || '';

        const bridgeContract = conn.contract('bridge');
        const result = await bridgeContract.evaluate('ListLocksPaginated', pageSize.toString(), bookmark);
        const locks = safeParse(result, { records: [], bookmark: '', count: 0 });

        res.json(locks);
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ListLocksPaginated failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/bridge/locks/:lockID — Get a specific lock receipt
router.get('/locks/:lockID', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { lockID } = req.params;

        const bridgeContract = conn.contract('bridge');
        const result = await bridgeContract.evaluate('GetLockReceipt', lockID);
        const lock = safeParse(result, null);

        if (!lock) {
            res.status(404).json({ error: `Lock ${lockID} not found` });
            return;
        }

        res.json(lock);
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`GetLockReceipt failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
