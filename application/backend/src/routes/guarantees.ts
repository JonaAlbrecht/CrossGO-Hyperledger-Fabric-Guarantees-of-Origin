// Guarantee of Origin issuance routes (v9: all 4 energy carriers)
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getIssuanceContract, getQueryContract } from '../fabric/contracts';
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

// POST /api/guarantees/electricity — create electricity GO (producer only)
router.post('/electricity', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { amount, amountMWh, emissions, elapsedSeconds,
                productionMethod, electricityProductionMethod } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            AmountMWh: amountMWh ?? amount,
            Emissions: emissions,
            ElapsedSeconds: elapsedSeconds,
            ElectricityProductionMethod: electricityProductionMethod ?? productionMethod,
        }));

        // Private data collection requires endorsement from both producer and issuer
        const endorsingOrgs = [req.user!.mspId, 'issuer1MSP'];

        await getIssuanceContract(conn).submit('CreateElectricityGO', {
            transientData: { eGO: transientData },
            endorsingOrganizations: endorsingOrgs,
        });

        res.status(201).json({ message: 'Electricity Guarantee of Origin created' });
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`CreateElectricityGO failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/guarantees/hydrogen — create hydrogen GO directly (producer only)
router.post('/hydrogen', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { amount, kilosProduced, emissions, emissionsHydrogen,
                productionMethod, hydrogenProductionMethod,
                elapsedSeconds, usedMWh } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            Kilosproduced: kilosProduced ?? amount,
            EmissionsHydrogen: emissionsHydrogen ?? emissions,
            HydrogenProductionMethod: hydrogenProductionMethod ?? productionMethod,
            UsedMWh: usedMWh ?? 0,
            ElapsedSeconds: elapsedSeconds,
        }));

        const endorsingOrgs = [req.user!.mspId, 'issuer1MSP'];

        await getIssuanceContract(conn).submit('CreateHydrogenGO', {
            transientData: { hGO: transientData },
            endorsingOrganizations: endorsingOrgs,
        });

        res.status(201).json({ message: 'Hydrogen Guarantee of Origin created' });
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`CreateHydrogenGO failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/guarantees/biogas — create biogas GO (producer only) — v9
router.post('/biogas', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { amount, cubicMeters, emissions, elapsedSeconds,
                productionMethod, biogasProductionMethod,
                energyContentMWh, feedstockType } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            VolumeNm3: cubicMeters ?? amount,
            EnergyContentMWh: energyContentMWh ?? 0,
            Emissions: emissions,
            ElapsedSeconds: elapsedSeconds,
            BiogasProductionMethod: biogasProductionMethod ?? productionMethod,
            FeedstockType: feedstockType ?? 'organic_waste',
        }));

        const endorsingOrgs = [req.user!.mspId, 'issuer1MSP'];

        await conn.contract('biogas').submit('CreateBiogasGO', {
            transientData: { bGO: transientData },
            endorsingOrganizations: endorsingOrgs,
        });

        res.status(201).json({ message: 'Biogas Guarantee of Origin created' });
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`CreateBiogasGO failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/guarantees/heating-cooling — create heating/cooling GO (producer only) — v9
router.post('/heating-cooling', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { amount, amountMWh, emissions, elapsedSeconds,
                productionMethod, heatingCoolingProductionMethod,
                supplyTemperature } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            AmountMWh: amountMWh ?? amount,
            Emissions: emissions,
            HeatingCoolingProductionMethod: heatingCoolingProductionMethod ?? productionMethod,
            SupplyTemperature: supplyTemperature ?? 60,
            ElapsedSeconds: elapsedSeconds,
        }));

        const endorsingOrgs = [req.user!.mspId, 'issuer1MSP'];

        await conn.contract('heating_cooling').submit('CreateHeatingCoolingGO', {
            transientData: { hcGO: transientData },
            endorsingOrganizations: endorsingOrgs,
        });

        res.status(201).json({ message: 'Heating/Cooling Guarantee of Origin created' });
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`CreateHeatingCoolingGO failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/guarantees — list all GOs for current org (public data)
router.get('/', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const queryContract = getQueryContract(conn);
        const [egoResult, hgoResult] = await Promise.all([
            queryContract.evaluate('GetCurrentEGOsList'),
            queryContract.evaluate('GetCurrentHGOsList'),
        ]);
        // v9: also query biogas and heating/cooling — graceful fallback if chaincode functions don't exist yet
        let biogasGOs: unknown[] = [];
        let heatingCoolingGOs: unknown[] = [];
        try {
            const bgResult = await queryContract.evaluate('GetCurrentBGOsList');
            biogasGOs = safeParse(bgResult, []);
        } catch { /* chaincode not yet supporting biogas queries */ }
        try {
            const hcResult = await queryContract.evaluate('GetCurrentHCGOsList');
            heatingCoolingGOs = safeParse(hcResult, []);
        } catch { /* chaincode not yet supporting heating/cooling queries */ }
        res.json({
            electricityGOs: safeParse(egoResult, []),
            hydrogenGOs: safeParse(hgoResult, []),
            biogasGOs,
            heatingCoolingGOs,
        });
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`List GOs failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/guarantees/:id — read public GO data
router.get('/:id', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const queryContract = getQueryContract(conn);
        // Try each GO type by prefix
        const id = req.params.id;
        let result: Uint8Array;
        const fnMap: Record<string, string> = {
            eGO: 'ReadPublicEGO', hGO: 'ReadPublicHGO',
            bGO: 'ReadPublicBGO', hcGO: 'ReadPublicHCGO',
        };
        const prefix = Object.keys(fnMap).find(p => id.startsWith(p));
        const fn = prefix ? fnMap[prefix] : 'ReadPublicEGO';
        result = await queryContract.evaluate(fn, { arguments: [id] });
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`ReadGO failed: ${message}`);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/guarantees/:id/private — read private GO details (collection-aware)
router.get('/:id/private', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const queryContract = getQueryContract(conn);
        let result: Uint8Array;
        try {
            result = await queryContract.evaluate('ReadPrivateEGO', { arguments: [req.params.id] });
        } catch {
            result = await queryContract.evaluate('ReadPrivateHGO', { arguments: [req.params.id] });
        }
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`ReadPrivateGO failed: ${message}`);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/guarantees/:id/history — full lifecycle history of a GO (v9: audit trail)
router.get('/:id/history', async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const queryContract = getQueryContract(conn);
        let result: Uint8Array;
        try {
            result = await queryContract.evaluate('GetGOHistory', { arguments: [req.params.id] });
        } catch {
            // Fallback: return the public data if history not yet implemented in chaincode
            try {
                result = await queryContract.evaluate('ReadPublicEGO', { arguments: [req.params.id] });
            } catch {
                result = await queryContract.evaluate('ReadPublicHGO', { arguments: [req.params.id] });
            }
        }
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`GetGOHistory failed: ${message}`);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
