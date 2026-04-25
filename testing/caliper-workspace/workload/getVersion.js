"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class GetVersion extends WorkloadModuleBase {
  async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
    await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
  }
  async submitTransaction() {
    const request = {
      contractId: "golifecycle",
      contractFunction: "admin:GetVersion",
      contractArguments: [],
      readOnly: true
    };
    await this.sutAdapter.sendRequests(request);
  }
}
module.exports.createWorkloadModule = () => new GetVersion();
