'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: device:ListDevicesPaginated — paginated device list.
 * Round arguments: pageSize (default 10)
 */
class ListDevicesPaginatedWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.pageSize = 10;
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
        this.pageSize = roundArguments.pageSize || 10;
    }

    async submitTransaction() {
        this.txIndex++;
        const args = {
            contractId: 'golifecycle',
            contractFunction: 'device:ListDevicesPaginated',
            contractArguments: [String(this.pageSize), ''],
            readOnly: true,
            timeout: 30
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new ListDevicesPaginatedWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
