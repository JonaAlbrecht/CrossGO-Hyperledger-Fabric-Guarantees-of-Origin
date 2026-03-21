// Cancellation routes — claim renewable attributes / cancel GOs
import { Router, Request, Response } from 'express';
import { authenticate } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getCancellationContract, getQueryContract } from '../fabric/contracts';
import { logger } from '../middleware/logger';

const router = Router();
router.use(authenticate);

async function getFabricConn(req: Request) {
    const { mspId, orgName, userName } = req.user!;
    const { certPath, keyPath } = getCryptoPath(orgName, userName);
    return connectToFabric(mspId, certPath, keyPath);
}

// POST /api/cancellations/electricity — cancel electricity GO (claim renewable attributes)
router.post('/electricity', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { goAssetID, amountMWh } = req.body;
        if (!goAssetID) {
            res.status(400).json({ error: 'goAssetID is required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            AssetID: goAssetID,
            AmountMWh: amountMWh,
        }));

        await getCancellationContract(conn).submit('ClaimRenewableAttributesElectricity', {
            transientData: { cancel: transientData },
        });

        res.status(201).json({ message: 'Electricity GO cancelled — cancellation statement created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ClaimRenewableAttributesElectricity failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/cancellations/hydrogen — cancel hydrogen GO (claim renewable attributes)
router.post('/hydrogen', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { goAssetID, kilos } = req.body;
        if (!goAssetID) {
            res.status(400).json({ error: 'goAssetID is required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            AssetID: goAssetID,
            Kilos: kilos,
        }));

        await getCancellationContract(conn).submit('ClaimRenewableAttributesHydrogen', {
            transientData: { cancel: transientData },
        });

        res.status(201).json({ message: 'Hydrogen GO cancelled — cancellation statement created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ClaimRenewableAttributesHydrogen failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/cancellations — list cancellation statements
router.get('/', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const queryContract = getQueryContract(conn);
        const { type } = req.query;
        let result: Uint8Array;

        if (type === 'hydrogen') {
            result = await queryContract.evaluate('ReadCancellationStatementHydrogen', req.query.id as string ?? '');
        } else {
            result = await queryContract.evaluate('ReadCancellationStatementElectricity', req.query.id as string ?? '');
        }
        res.json(JSON.parse(new TextDecoder().decode(result)));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ReadCancellationStatement failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/cancellations/verify — verify a cancellation statement
router.post('/verify', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { cancellationStatementID } = req.body;
        if (!cancellationStatementID) {
            res.status(400).json({ error: 'cancellationStatementID is required' });
            return;
        }

        const result = await getCancellationContract(conn).evaluate(
            'VerifyCancellationStatement',
            cancellationStatementID
        );

        res.json({ verified: true, data: JSON.parse(new TextDecoder().decode(result)) });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`VerifyCancellationStatement failed: ${message}`);
        res.status(500).json({ error: message, verified: false });
    } finally {
        conn.close();
    }
});

export default router;
