// Conversion routes — hydrogen backlog and issuance from electricity GOs
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getConversionContract } from '../fabric/contracts';
import { logger } from '../middleware/logger';

const router = Router();
router.use(authenticate);

async function getFabricConn(req: Request) {
    const { mspId, orgName, userName } = req.user!;
    const { certPath, keyPath } = getCryptoPath(orgName, userName);
    return connectToFabric(mspId, certPath, keyPath);
}

// POST /api/conversions/backlog — add hydrogen production to backlog
router.post('/backlog', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { kilosHydrogen, hydrogenProductionMethod, mwhElectricity } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            Kilosproduced: kilosHydrogen,
            HydrogenProductionMethod: hydrogenProductionMethod,
            UsedMWh: mwhElectricity,
        }));

        await getConversionContract(conn).submit('AddHydrogenToBacklog', {
            transientData: { backlog: transientData },
        });

        res.status(201).json({ message: 'Added to hydrogen backlog' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`AddHydrogenToBacklog failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/conversions/issue — issue hydrogen GO from backlog (consumes electricity GOs)
router.post('/issue', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        await getConversionContract(conn).submitAsync('IssuehGO');
        res.status(201).json({ message: 'Hydrogen GO issued from backlog' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`IssuehGO failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/conversions/backlog — query current hydrogen backlog
router.get('/backlog', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getConversionContract(conn).evaluate('QueryHydrogenBacklog');
        res.json(JSON.parse(new TextDecoder().decode(result)));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`QueryHydrogenBacklog failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
