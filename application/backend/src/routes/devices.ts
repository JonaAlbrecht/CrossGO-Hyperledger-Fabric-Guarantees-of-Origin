// Device management routes — register, list, revoke, suspend, reactivate
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getDeviceContract } from '../fabric/contracts';
import { logger } from '../middleware/logger';

const router = Router();
router.use(authenticate);

// Helper: get a Fabric connection from the authenticated user
async function getFabricConn(req: Request) {
    const { mspId, orgName, userName } = req.user!;
    const { certPath, keyPath } = getCryptoPath(orgName, userName);
    return connectToFabric(mspId, certPath, keyPath);
}

// POST /api/devices — register a new device (issuer only)
router.post('/', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { deviceType, ownerOrgMSP, energyCarriers, attributes } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            deviceType,
            ownerOrgMSP,
            energyCarriers: energyCarriers ?? [],
            attributes: attributes ?? {},
        }));

        await getDeviceContract(conn).submit('RegisterDevice', {
            transientData: { device: transientData },
        });

        res.status(201).json({ message: 'Device registered' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`RegisterDevice failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/devices — list devices for current org
router.get('/', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getDeviceContract(conn).evaluate('ListDevices');
        res.json(JSON.parse(new TextDecoder().decode(result)));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ListDevices failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/devices/:id — get device by ID
router.get('/:id', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const result = await getDeviceContract(conn).evaluate('GetDevice', req.params.id);
        res.json(JSON.parse(new TextDecoder().decode(result)));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`GetDevice failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// PUT /api/devices/:id/revoke — revoke a device (issuer only)
router.put('/:id/revoke', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        await getDeviceContract(conn).submit('RevokeDevice', req.params.id);
        res.json({ message: 'Device revoked' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`RevokeDevice failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// PUT /api/devices/:id/suspend — suspend a device (issuer only)
router.put('/:id/suspend', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        await getDeviceContract(conn).submit('SuspendDevice', req.params.id);
        res.json({ message: 'Device suspended' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`SuspendDevice failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// PUT /api/devices/:id/reactivate — reactivate a suspended device (issuer only)
router.put('/:id/reactivate', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        await getDeviceContract(conn).submit('ReactivateDevice', req.params.id);
        res.json({ message: 'Device reactivated' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ReactivateDevice failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
