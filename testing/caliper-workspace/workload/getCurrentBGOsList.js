'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: query:GetCurrentBGOsList — range query for biogas GOs.
 */
class GetCurrentBGOsListWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const args = {
            contractId: 'golifecycle',
            contractFunction: 'query:GetCurrentBGOsList',
            contractArguments: [],
            readOnly: true,
            timeout: 30
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new GetCurrentBGOsListWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
