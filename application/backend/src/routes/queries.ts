// Query routes — read-only queries against the ledger
import { Router, Request, Response } from 'express';
import { authenticate } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getQueryContract } from '../fabric/contracts';
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

// GET /api/queries/ego-list — all current electricity GOs
router.get('/ego-list', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getQueryContract(conn).evaluate('GetCurrentEGOsList');
        res.json(safeParse(result, []));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`GetCurrentEGOsList failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/queries/hgo-list — all current hydrogen GOs
router.get('/hgo-list', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getQueryContract(conn).evaluate('GetCurrentHGOsList');
        res.json(safeParse(result, []));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`GetCurrentHGOsList failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/queries/ego/:id — public electricity GO
router.get('/ego/:id', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getQueryContract(conn).evaluate('ReadPublicEGO', req.params.id);
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/queries/hgo/:id — public hydrogen GO
router.get('/hgo/:id', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getQueryContract(conn).evaluate('ReadPublicHGO', req.params.id);
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/queries/ego/:id/private — private electricity GO details
router.get('/ego/:id/private', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getQueryContract(conn).evaluate('ReadPrivateEGO', req.params.id);
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/queries/hgo/:id/private — private hydrogen GO details
router.get('/hgo/:id/private', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getQueryContract(conn).evaluate('ReadPrivateHGO', req.params.id);
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/queries/ego-by-amount?mwh=X — find electricity GOs meeting amount threshold
router.get('/ego-by-amount', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const mwh = req.query.mwh as string;
        if (!mwh) {
            res.status(400).json({ error: 'mwh query parameter is required' });
            return;
        }
        const result = await getQueryContract(conn).evaluate('QueryPrivateEGOsByAmountMWh', mwh);
        res.json(safeParse(result, []));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`QueryPrivateEGOsByAmountMWh failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/queries/hgo-by-amount?kilos=X — find hydrogen GOs meeting amount threshold
router.get('/hgo-by-amount', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const kilos = req.query.kilos as string;
        if (!kilos) {
            res.status(400).json({ error: 'kilos query parameter is required' });
            return;
        }
        const result = await getQueryContract(conn).evaluate('QueryPrivateHGOsByAmount', kilos);
        res.json(safeParse(result, []));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`QueryPrivateHGOsByAmount failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
