"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class ListOracleDataPaginated extends WorkloadModuleBase {
  async submitTransaction() {
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "oracle:ListOracleDataPaginated",
      contractArguments: ["electricity", "10", ""],
      readOnly: true
    });
  }
}
module.exports.createWorkloadModule = () => new ListOracleDataPaginated();
