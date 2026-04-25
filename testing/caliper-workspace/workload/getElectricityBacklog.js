"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class GetElectricityBacklog extends WorkloadModuleBase {
  async submitTransaction() {
    // Generate a test DeviceID
    const deviceId = `DEV-ELEC-${Math.floor(Math.random() * 1000).toString().padStart(3, '0')}`;
    
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "backlog:GetElectricityBacklog",
      contractArguments: [deviceId],
      readOnly: true,
      invokerIdentity: "eproducer1MSP_eproducer1_admin"
    });
  }
}
module.exports.createWorkloadModule = () => new GetElectricityBacklog();
