"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class GetCurrentEGOsList extends WorkloadModuleBase {
  async submitTransaction() {
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "query:GetCurrentEGOsList",
      contractArguments: [],
      readOnly: true
    });
  }
}
module.exports.createWorkloadModule = () => new GetCurrentEGOsList();
