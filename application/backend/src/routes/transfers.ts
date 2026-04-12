// Transfer routes — transfer GOs between organizations (v9: all 4 energy carriers)
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getTransferContract } from '../fabric/contracts';
import { logger } from '../middleware/logger';

const router = Router();
router.use(authenticate);

async function getFabricConn(req: Request) {
    const { mspId, orgName, userName } = req.user!;
    const { certPath, keyPath } = getCryptoPath(orgName, userName);
    return connectToFabric(mspId, certPath, keyPath);
}

// POST /api/transfers — transfer a single GO by asset ID
router.post('/', requireRole('producer', 'buyer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { goAssetID, recipientMSP } = req.body;
        if (!goAssetID || !recipientMSP) {
            res.status(400).json({ error: 'goAssetID and recipientMSP are required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            AssetID: goAssetID,
            RecipientMSP: recipientMSP,
        }));

        await getTransferContract(conn).submit('TransferEGO', {
            transientData: { TransferInput: transientData },
        });

        res.json({ message: 'GO transferred successfully' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`TransferEGO failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/transfers/electricity-by-amount — transfer electricity GO by MWh amount
router.post('/electricity-by-amount', requireRole('producer', 'buyer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { recipientMSP, amountMWh } = req.body;
        if (!recipientMSP || !amountMWh) {
            res.status(400).json({ error: 'recipientMSP and amountMWh are required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            RecipientMSP: recipientMSP,
            AmountMWh: amountMWh,
        }));

        await getTransferContract(conn).submit('TransferEGOByAmount', {
            transientData: { TransferInput: transientData },
        });

        res.json({ message: `${amountMWh} MWh electricity transferred to ${recipientMSP}` });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`TransferEGOByAmount failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/transfers/hydrogen-by-amount — transfer hydrogen GO by kg amount
router.post('/hydrogen-by-amount', requireRole('producer', 'buyer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { recipientMSP, kilos } = req.body;
        if (!recipientMSP || !kilos) {
            res.status(400).json({ error: 'recipientMSP and kilos are required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            RecipientMSP: recipientMSP,
            Kilos: kilos,
        }));

        await getTransferContract(conn).submit('TransferHGOByAmount', {
            transientData: { TransferInput: transientData },
        });

        res.json({ message: `${kilos} kg hydrogen transferred to ${recipientMSP}` });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`TransferHGOByAmount failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/transfers/biogas-by-amount — transfer biogas GO by cubic meters (v9)
router.post('/biogas-by-amount', requireRole('producer', 'buyer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { recipientMSP, cubicMeters } = req.body;
        if (!recipientMSP || !cubicMeters) {
            res.status(400).json({ error: 'recipientMSP and cubicMeters are required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            RecipientMSP: recipientMSP,
            CubicMeters: cubicMeters,
        }));

        await getTransferContract(conn).submit('TransferBiogasGOByAmount', {
            transientData: { TransferInput: transientData },
        });

        res.json({ message: `${cubicMeters} m³ biogas transferred to ${recipientMSP}` });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`TransferBiogasGOByAmount failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/transfers/heating-cooling-by-amount — transfer heating/cooling GO by MWh (v9)
router.post('/heating-cooling-by-amount', requireRole('producer', 'buyer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { recipientMSP, amountMWh } = req.body;
        if (!recipientMSP || !amountMWh) {
            res.status(400).json({ error: 'recipientMSP and amountMWh are required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            RecipientMSP: recipientMSP,
            AmountMWh: amountMWh,
        }));

        await getTransferContract(conn).submit('TransferHeatingCoolingGOByAmount', {
            transientData: { TransferInput: transientData },
        });

        res.json({ message: `${amountMWh} MWh heating/cooling transferred to ${recipientMSP}` });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`TransferHeatingCoolingGOByAmount failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
