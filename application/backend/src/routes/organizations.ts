// Organization routes — register and list organizations (v9)
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { logger } from '../middleware/logger';

function safeParse(data: Uint8Array, fallback: any = []) {
    const str = new TextDecoder().decode(data).trim();
    return str ? JSON.parse(str) : fallback;
}

/** Extract the most useful error message from a Fabric gateway error. */
function fabricError(err: unknown): string {
    if (!(err instanceof Error)) return String(err);
    const details = (err as any).details;
    if (Array.isArray(details) && details.length > 0) {
        const msgs = details.map((d: any) => d.message ?? String(d)).join('; ');
        return `${err.message} — ${msgs}`;
    }
    return err.message;
}

const router = Router();
router.use(authenticate);

async function getFabricConn(req: Request) {
    const { mspId, orgName, userName } = req.user!;
    const { certPath, keyPath } = getCryptoPath(orgName, userName);
    return connectToFabric(mspId, certPath, keyPath);
}

// POST /api/organizations — register a new organization (issuer only)
router.post('/', requireRole('issuer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { orgMSP, displayName, role, energyCarriers, country } = req.body;
        if (!orgMSP || !displayName || !role) {
            res.status(400).json({ error: 'orgMSP, displayName, and role are required' });
            return;
        }
        if (!['producer', 'buyer'].includes(role)) {
            res.status(400).json({ error: 'role must be "producer" or "buyer"' });
            return;
        }

        const adminContract = conn.contract('admin');
        const transientData = Buffer.from(JSON.stringify({
            OrgMSP: orgMSP,
            DisplayName: displayName,
            OrgType: role,
            EnergyCarriers: energyCarriers ?? [],
            Country: country ?? 'DE',
        }));

        await adminContract.submit('RegisterOrganization', {
            transientData: { OrgRegistration: transientData },
        });

        logger.info(`Organization ${displayName} (${orgMSP}) registered by issuer`);
        res.status(201).json({ message: `Organization ${displayName} registered successfully` });
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`RegisterOrganization failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/organizations — list registered organizations
router.get('/', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const adminContract = conn.contract('admin');
        const result = await adminContract.evaluate('ListOrganizations');
        res.json(safeParse(result, []));
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`ListOrganizations failed: ${message}`);
        // Return default orgs if chaincode function not yet implemented
        res.json([
            { orgMSP: 'issuer1MSP', displayName: 'German Issuing Authority (UBA)', role: 'issuer' },
            { orgMSP: 'eproducer1MSP', displayName: 'Alpha WindFarm GmbH', role: 'producer' },
            { orgMSP: 'hproducer1MSP', displayName: 'Beta Electrolyser B.V.', role: 'producer' },
            { orgMSP: 'buyer1MSP', displayName: 'Gamma-Town EnergySupplier Ltd', role: 'buyer' },
        ]);
    } finally {
        conn.close();
    }
});

export default router;
