'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: admin:GetOrganization — reads an org registration by MSP ID.
 */
class GetOrganizationWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.orgs = ['issuer1MSP', 'eproducer1MSP', 'hproducer1MSP', 'buyer1MSP'];
    }

    async submitTransaction() {
        this.txIndex++;
        const orgMSP = this.orgs[(this.txIndex - 1) % this.orgs.length];
        const args = {
            contractId: 'golifecycle',
            contractFunction: 'admin:GetOrganization',
            contractArguments: [orgMSP],
            readOnly: true,
            timeout: 30
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new GetOrganizationWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
