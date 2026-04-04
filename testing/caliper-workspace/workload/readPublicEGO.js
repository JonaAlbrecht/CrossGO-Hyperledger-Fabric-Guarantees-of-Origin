'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload module for reading a public electricity GO by ID.
 * Calls query:ReadPublicEGO with eGO ID as a string argument.
 * Requires pre-seeded eGO data on the ledger.
 */
class ReadPublicEGOWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.maxEgoId = 1;
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
        this.maxEgoId = roundArguments.maxEgoId || 10;
    }

    async submitTransaction() {
        this.txIndex++;
        const eGOID = 'eGO' + (((this.txIndex - 1) % this.maxEgoId) + 1);

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'query:ReadPublicEGO',
            contractArguments: [eGOID],
            readOnly: true,
            timeout: 30
        };

        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new ReadPublicEGOWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
