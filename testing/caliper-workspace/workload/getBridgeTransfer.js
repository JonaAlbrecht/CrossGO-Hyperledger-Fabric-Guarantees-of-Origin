'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const fs = require('fs');
const path = require('path');

/**
 * Workload: bridge:GetBridgeTransfer — reads bridge transfer records.
 * Uses bridge IDs from bridge-ids.json (populated by seed round).
 * Fallback: uses generated prefixed IDs.
 */
class GetBridgeTransferWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.bridgeIds = [];
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        const idsFile = path.join(__dirname, 'bridge-ids.json');
        if (fs.existsSync(idsFile)) {
            this.bridgeIds = JSON.parse(fs.readFileSync(idsFile, 'utf8'));
        }
        if (this.bridgeIds.length === 0) {
            this.bridgeIds = ['bridge_0000000000000001'];
        }
    }

    async submitTransaction() {
        this.txIndex++;
        const bridgeId = this.bridgeIds[(this.txIndex - 1) % this.bridgeIds.length];

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'bridge:GetBridgeTransfer',
            contractArguments: [bridgeId],
            readOnly: true,
            timeout: 30
        };

        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new GetBridgeTransferWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
