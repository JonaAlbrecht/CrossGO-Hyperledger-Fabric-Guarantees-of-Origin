// Cancellation routes — cancel GOs and create Cancellation Statements (v9: all 4 energy carriers)
import { Router, Request, Response } from 'express';
import { authenticate } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getCancellationContract, getQueryContract } from '../fabric/contracts';
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

// POST /api/cancellations/electricity — cancel electricity GO
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
            transientData: { ClaimRenewables: transientData },
        });

        res.status(201).json({ message: 'Electricity GO cancelled — Cancellation Statement created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ClaimRenewableAttributesElectricity failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/cancellations/hydrogen — cancel hydrogen GO
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
            transientData: { ClaimHydrogen: transientData },
        });

        res.status(201).json({ message: 'Hydrogen GO cancelled — Cancellation Statement created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ClaimRenewableAttributesHydrogen failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/cancellations/biogas — cancel biogas GO (v9)
router.post('/biogas', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { goAssetID, cubicMeters } = req.body;
        if (!goAssetID) {
            res.status(400).json({ error: 'goAssetID is required' });
            return;
        }

        const transientData = Buffer.from(JSON.stringify({
            AssetID: goAssetID,
            CubicMeters: cubicMeters,
        }));

        await getCancellationContract(conn).submit('ClaimRenewableAttributesBiogas', {
            transientData: { ClaimBiogas: transientData },
        });

        res.status(201).json({ message: 'Biogas GO cancelled — Cancellation Statement created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ClaimRenewableAttributesBiogas failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/cancellations/heating_cooling — cancel heating/cooling GO (v9)
router.post('/heating_cooling', async (req: Request, res: Response) => {
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

        await getCancellationContract(conn).submit('ClaimRenewableAttributesHeatingCooling', {
            transientData: { ClaimHeatingCooling: transientData },
        });

        res.status(201).json({ message: 'Heating/Cooling GO cancelled — Cancellation Statement created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ClaimRenewableAttributesHeatingCooling failed: ${message}`);
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
            result = await queryContract.evaluate('ReadCancellationStatementHydrogen', { arguments: [req.query.id as string ?? ''] });
        } else {
            result = await queryContract.evaluate('ReadCancellationStatementElectricity', { arguments: [req.query.id as string ?? ''] });
        }
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ReadCancellationStatement failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/cancellations/verify — verify a Cancellation Statement
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

        res.json({ verified: true, data: safeParse(result, null) });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`VerifyCancellationStatement failed: ${message}`);
        res.status(500).json({ error: message, verified: false });
    } finally {
        conn.close();
    }
});

export default router;
