'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * Workload: biogas:CreateBiogasGO — creates bGOs via transient data.
 * Run as eproducer1 identity (producer role with biogas device).
 */
class CreateBiogasGOWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const volumeNm3 = 100 + Math.floor(Math.random() * 50);
        const energyMWh = volumeNm3 * 0.01;

        const bGOData = {
            VolumeNm3: volumeNm3,
            EnergyContentMWh: energyMWh,
            Emissions: volumeNm3 * 2,
            BiogasProductionMethod: 'anaerobic_digestion',
            FeedstockType: 'agricultural_waste',
            ElapsedSeconds: 3600
        };

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'biogas:CreateBiogasGO',
            contractArguments: [],
            timeout: 60,
            transientMap: { bGO: JSON.stringify(bGOData) }
        };
        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new CreateBiogasGOWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
