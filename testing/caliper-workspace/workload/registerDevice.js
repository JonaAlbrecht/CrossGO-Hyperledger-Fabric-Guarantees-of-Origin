'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload module for registering metering devices.
 * Uses the issuer identity (issuer1MSP has role=issuer).
 * Calls device:RegisterDevice with transient data.
 */
class RegisterDeviceWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;

        const deviceData = {
            deviceType: this.txIndex % 2 === 0 ? 'SmartMeter' : 'OutputMeter',
            ownerOrgMSP: 'eproducer1MSP',
            energyCarriers: ['electricity'],
            attributes: {
                maxEfficiency: '100',
                emissionIntensity: '50',
                technologyType: 'solar'
            }
        };

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'device:RegisterDevice',
            contractArguments: [],
            timeout: 60,
            transientMap: { Device: JSON.stringify(deviceData) }
        };

        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new RegisterDeviceWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
