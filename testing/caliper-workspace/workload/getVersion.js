'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: admin:GetVersion (read-only, no args)
 */
class GetVersionWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const args = {
            contractId: 'golifecycle',
            contractFunction: 'admin:GetVersion',
            contractArguments: [],
            readOnly: true,
            timeout: 30
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new GetVersionWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
