// Conversion routes — generalized per-carrier backlog and issuance (v9)
import { Router, Request, Response } from 'express';
import { authenticate, requireRole } from '../middleware/auth';
import { connectToFabric, getCryptoPath } from '../fabric/gateway';
import { getConversionContract } from '../fabric/contracts';
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

// POST /api/conversions/backlog — add production to carrier-specific backlog (v9: generalized)
router.post('/backlog', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { targetCarrier, sourceCarrier, productionMethod, ...amounts } = req.body;

        // v9: generalized backlog — still routes to hydrogen-specific chaincode for now
        // When chaincode is extended, dispatch by targetCarrier
        if (targetCarrier === 'hydrogen' || !targetCarrier) {
            // Legacy: hydrogen from electricity
            const kilos = amounts.kilosHydrogen ?? amounts.outputAmount;
            const mwh = amounts.sourceAmount_amountMWh ?? amounts.mwhElectricity;
            const emissions = amounts.emissionsHydrogen ?? amounts.emissions ?? 0;
            const elapsed = amounts.elapsedSeconds ?? 1;
            const transientData = Buffer.from(JSON.stringify({
                Kilosproduced: kilos,
                EmissionsHydrogen: emissions,
                HydrogenProductionMethod: productionMethod ?? amounts.hydrogenProductionMethod ?? 'electrolysis',
                UsedMWh: mwh,
                ElapsedSeconds: elapsed,
            }));

            const endorsingOrgs = [req.user!.mspId, 'issuer1MSP'];

            await getConversionContract(conn).submit('AddHydrogenToBacklog', {
                transientData: { hGObacklog: transientData },
                endorsingOrganizations: endorsingOrgs,
            });

            res.status(201).json({ message: `Added to ${targetCarrier ?? 'hydrogen'} backlog` });
        } else {
            // v9 future carriers — generic backlog submission
            const transientData = Buffer.from(JSON.stringify({
                TargetCarrier: targetCarrier,
                SourceCarrier: sourceCarrier,
                ProductionMethod: productionMethod,
                ...amounts,
            }));

            await getConversionContract(conn).submit('AddToBacklog', {
                transientData: { backlogEntry: transientData },
            });

            res.status(201).json({ message: `Added to ${targetCarrier} backlog` });
        }
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`AddToBacklog failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// POST /api/conversions/issue — issue GO from backlog (v9: target carrier aware)
router.post('/issue', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        const { targetCarrier } = req.body ?? {};

        if (targetCarrier === 'hydrogen' || !targetCarrier) {
            const endorsingOrgs = [req.user!.mspId, 'issuer1MSP'];
            await getConversionContract(conn).submit('IssuehGO', {
                endorsingOrganizations: endorsingOrgs,
            });
            res.status(201).json({ message: 'Hydrogen GO issued from backlog' });
        } else {
            // v9 generic issuance
            await getConversionContract(conn).submit('IssueGOFromBacklog', targetCarrier);
            res.status(201).json({ message: `${targetCarrier} GO issued from backlog` });
        }
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`IssueFromBacklog failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

// GET /api/conversions/backlog — query current backlog (optionally filtered by carrier)
router.get('/backlog', requireRole('producer'), async (req: Request, res: Response) => {
    const conn = await getFabricConn(req);
    try {
        // QueryHydrogenBacklog requires transient key "QueryInput" with Collection name
        const collection = `privateDetails-${req.user!.mspId}`;
        const transientData = Buffer.from(JSON.stringify({ Collection: collection }));
        const result = await getConversionContract(conn).evaluate('QueryHydrogenBacklog', {
            transientData: { QueryInput: transientData },
        });
        res.json(safeParse(result, null));
    } catch (err: unknown) {
        const message = fabricError(err);
        logger.error(`QueryBacklog failed: ${message}`);
        res.status(500).json({ error: message });
    } finally {
        conn.close();
    }
});

export default router;
