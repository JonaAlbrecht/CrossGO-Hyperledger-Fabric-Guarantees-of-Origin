'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: query:GetCurrentEGOsListPaginated — paginated eGO query.
 * Round arguments: pageSize (default 50)
 */
class GetCurrentEGOsListPaginatedWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.pageSize = 50;
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
        this.pageSize = roundArguments.pageSize || 50;
    }

    async submitTransaction() {
        this.txIndex++;
        const args = {
            contractId: 'golifecycle',
            contractFunction: 'query:GetCurrentEGOsListPaginated',
            contractArguments: [String(this.pageSize), ''],
            readOnly: true,
            timeout: 30
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new GetCurrentEGOsListPaginatedWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
