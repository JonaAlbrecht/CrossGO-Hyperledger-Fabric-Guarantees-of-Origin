'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: oracle:PublishGridData — publishes ENTSO-E grid generation records.
 * Run as issuer1 identity (only issuers can publish oracle data).
 */
class PublishGridDataWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const now = Math.floor(Date.now() / 1000);
        const periodStart = now - 3600;
        const periodEnd = now;

        const gridData = {
            BiddingZone: 'DE-LU',
            PeriodStart: periodStart,
            PeriodEnd: periodEnd,
            EnergySource: 'F01010100', // Solar PV
            GenerationMW: 5000 + Math.floor(Math.random() * 2000),
            EmissionFactor: 20 + Math.floor(Math.random() * 10),
            DataSource: 'ENTSO-E-TP'
        };

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'oracle:PublishGridData',
            contractArguments: [],
            invokerMspId: 'issuer1MSP',
            invokerIdentity: 'issuer1-admin',
            timeout: 60,
            transientMap: { GridData: JSON.stringify(gridData) }
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new PublishGridDataWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
