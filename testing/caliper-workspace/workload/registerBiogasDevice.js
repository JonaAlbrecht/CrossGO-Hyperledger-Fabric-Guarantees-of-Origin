'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: device:RegisterDevice — registers biogas devices.
 * Run as issuer1 identity (issuer-only operation).
 */
class RegisterBiogasDeviceWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const deviceData = {
            deviceType: this.txIndex % 2 === 0 ? 'SmartMeter' : 'OutputMeter',
            ownerOrgMSP: 'eproducer1MSP',
            energyCarriers: ['biogas'],
            attributes: {
                maxEfficiency: '100',
                emissionIntensity: '200',
                technologyType: 'anaerobic_digestion'
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
    return new RegisterBiogasDeviceWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
