'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: issuance:CreateElectricityGO — creates eGOs via transient data.
 * Run as eproducer1 identity (producer role).
 */
class CreateElectricityGOWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const amountMWh = 40 + Math.floor(Math.random() * 10);
        const emissions = amountMWh * 50;

        const eGOData = {
            AmountMWh: amountMWh,
            Emissions: emissions,
            ElapsedSeconds: 3600,
            ElectricityProductionMethod: 'solar'
        };

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'issuance:CreateElectricityGO',
            contractArguments: [],
            timeout: 60,
            transientMap: { eGO: JSON.stringify(eGOData) }
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new CreateElectricityGOWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
