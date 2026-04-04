'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: admin:RegisterOrganization — registers orgs via transient data.
 * Run as issuer1 identity.
 */
class RegisterOrganizationWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const orgData = {
            DisplayName: `TestOrg_${Date.now()}_${this.txIndex}`,
            OrgMSP: `testOrg${this.txIndex}MSP`,
            OrgType: this.txIndex % 2 === 0 ? 'producer' : 'consumer',
            EnergyCarriers: ['electricity'],
            Country: 'DE'
        };

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'admin:RegisterOrganization',
            contractArguments: [],
            timeout: 60,
            transientMap: { OrgRegistration: JSON.stringify(orgData) }
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new RegisterOrganizationWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
