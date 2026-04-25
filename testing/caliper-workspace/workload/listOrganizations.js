"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class ListOrganizations extends WorkloadModuleBase {
  async submitTransaction() {
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "admin:ListOrganizations",
      contractArguments: [],
      readOnly: true
    });
  }
}
module.exports.createWorkloadModule = () => new ListOrganizations();
