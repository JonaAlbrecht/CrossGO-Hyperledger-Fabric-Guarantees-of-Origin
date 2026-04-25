"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");

class GetConversionLocks extends WorkloadModuleBase {
  async submitTransaction() {
    const targetCarrier = Math.random() > 0.5 ? "hydrogen" : "biogas";
    
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "conversion:GetConversionLocks",
      contractArguments: [targetCarrier],
      readOnly: true,
      invokerIdentity: "_hissuerMSP_hissuer_admin"
    });
  }
}

module.exports.createWorkloadModule = () => new GetConversionLocks();
