// Auth routes — login (identity-based) and token issuance
import { Router, Request, Response } from 'express';
import { signToken } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getDeviceContract } from '../fabric/contracts';
import { logger } from '../middleware/logger';

const router = Router();

// Maps org names to MSP IDs and roles
const ORG_MAP: Record<string, { mspId: string; role: string }> = {
    issuer1: { mspId: 'issuer1MSP', role: 'issuer' },
    producer1: { mspId: 'producer1MSP', role: 'producer' },
    consumer1: { mspId: 'consumer1MSP', role: 'consumer' },
};

// POST /api/auth/login — authenticate with org name + user name
router.post('/login', async (req: Request, res: Response) => {
    try {
        const { orgName, userName } = req.body;
        if (!orgName || !userName) {
            res.status(400).json({ error: 'orgName and userName are required' });
            return;
        }

        const org = ORG_MAP[orgName];
        if (!org) {
            res.status(400).json({ error: `Unknown org: ${orgName}. Valid: ${Object.keys(ORG_MAP).join(', ')}` });
            return;
        }

        // Verify the user's crypto materials exist and can connect
        const { certPath, keyPath } = getCryptoPath(orgName, userName);
        const conn = await connectToFabric(org.mspId, certPath, keyPath);

        // Quick connectivity check — call a read-only function
        const deviceContract = getDeviceContract(conn);
        await deviceContract.evaluateTransaction('ListDevices');
        conn.close();

        const token = signToken({
            mspId: org.mspId,
            orgName,
            userName,
            role: org.role,
        });

        logger.info(`User ${userName}@${orgName} authenticated`);
        res.json({ token, role: org.role, mspId: org.mspId });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`Login failed: ${message}`);
        res.status(401).json({ error: 'Authentication failed', details: message });
    }
});

export default router;
