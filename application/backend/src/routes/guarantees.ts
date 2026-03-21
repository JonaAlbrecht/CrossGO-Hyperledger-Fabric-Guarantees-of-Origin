// Guarantee of Origin issuance routes — create electricity and hydrogen GOs
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getIssuanceContract, getQueryContract } from '../fabric/contracts';
import { logger } from '../middleware/logger';

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
        const { amountMWh, emissions, elapsedSeconds, electricityProductionMethod } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            AmountMWh: amountMWh,
            Emissions: emissions,
            ElapsedSeconds: elapsedSeconds,
            ElectricityProductionMethod: electricityProductionMethod,
        }));

        await getIssuanceContract(conn).submit('CreateElectricityGO', {
            transientData: { eGO: transientData },
        });

        res.status(201).json({ message: 'Electricity GO created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
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
        const { kilosProduced, emissionsHydrogen, hydrogenProductionMethod,
                inputEmissions, usedMWh, electricityProductionMethods } = req.body;
        const transientData = Buffer.from(JSON.stringify({
            Kilosproduced: kilosProduced,
            EmissionsHydrogen: emissionsHydrogen,
            HydrogenProductionMethod: hydrogenProductionMethod,
            InputEmissions: inputEmissions,
            UsedMWh: usedMWh,
            ElectricityProductionMethod: electricityProductionMethods ?? [],
        }));

        await getIssuanceContract(conn).submit('CreateHydrogenGO', {
            transientData: { hGO: transientData },
        });

        res.status(201).json({ message: 'Hydrogen GO created' });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`CreateHydrogenGO failed: ${message}`);
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
        res.json({
            electricityGOs: JSON.parse(new TextDecoder().decode(egoResult)),
            hydrogenGOs: JSON.parse(new TextDecoder().decode(hgoResult)),
        });
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
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
        // Try electricity first, fall back to hydrogen
        let result: Uint8Array;
        try {
            result = await queryContract.evaluate('ReadPublicEGO', req.params.id);
        } catch {
            result = await queryContract.evaluate('ReadPublicHGO', req.params.id);
        }
        res.json(JSON.parse(new TextDecoder().decode(result)));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
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
            result = await queryContract.evaluate('ReadPrivateEGO', req.params.id);
        } catch {
            result = await queryContract.evaluate('ReadPrivateHGO', req.params.id);
        }
        res.json(JSON.parse(new TextDecoder().decode(result)));
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        logger.error(`ReadPrivateGO failed: ${message}`);
        res.status(404).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
