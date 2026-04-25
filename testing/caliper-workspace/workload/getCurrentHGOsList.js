"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class GetCurrentHGOsList extends WorkloadModuleBase {
  async submitTransaction() {
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "query:GetCurrentHGOsList",
      contractArguments: [],
      readOnly: true
    });
  }
}
module.exports.createWorkloadModule = () => new GetCurrentHGOsList();
