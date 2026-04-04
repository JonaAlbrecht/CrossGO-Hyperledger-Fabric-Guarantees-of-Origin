'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const fs = require('fs');
const path = require('path');

/**
 * Workload: oracle:GetGridData — reads oracle grid generation records.
 * Uses oracle IDs from oracle-ids.json (populated by seed round).
 */
class GetGridDataWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
        this.oracleIds = [];
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        const idsFile = path.join(__dirname, 'oracle-ids.json');
        if (fs.existsSync(idsFile)) {
            this.oracleIds = JSON.parse(fs.readFileSync(idsFile, 'utf8'));
        }
        if (this.oracleIds.length === 0) {
            this.oracleIds = ['oracle_0000000000000001'];
        }
    }

    async submitTransaction() {
        this.txIndex++;
        const oracleId = this.oracleIds[(this.txIndex - 1) % this.oracleIds.length];

        const args = {
            contractId: 'golifecycle',
            contractFunction: 'oracle:GetGridData',
            contractArguments: [oracleId],
            readOnly: true,
            timeout: 30
        };

        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new GetGridDataWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
