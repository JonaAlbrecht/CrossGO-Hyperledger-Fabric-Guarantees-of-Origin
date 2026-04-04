'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload module for listing all electricity GOs (range query).
 * Calls query:GetCurrentEGOsList with no arguments.
 */
class GetCurrentEGOsListWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'query:GetCurrentEGOsList',
            contractArguments: [],
            readOnly: true,
            timeout: 30
        };

        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new GetCurrentEGOsListWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
