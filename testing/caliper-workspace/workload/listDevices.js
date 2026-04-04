'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload module for listing all devices (range query).
 * Calls device:ListDevices with no arguments.
 */
class ListDevicesWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'device:ListDevices',
            contractArguments: [],
            readOnly: true,
            timeout: 30
        };

        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new ListDevicesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
