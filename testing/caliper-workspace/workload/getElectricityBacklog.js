"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class GetElectricityBacklog extends WorkloadModuleBase {
  async submitTransaction() {
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "backlog:GetElectricityBacklog",
      contractArguments: [],
      readOnly: true
    });
  }
}
module.exports.createWorkloadModule = () => new GetElectricityBacklog();
