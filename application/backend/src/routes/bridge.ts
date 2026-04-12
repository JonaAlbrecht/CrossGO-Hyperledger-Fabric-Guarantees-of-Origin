// Bridge routes — cross-channel GO transfer with dual-issuer consensus (v9)
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

export default router;
