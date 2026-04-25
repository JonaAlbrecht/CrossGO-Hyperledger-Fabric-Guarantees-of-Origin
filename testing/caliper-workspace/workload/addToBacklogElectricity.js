"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class AddToBacklogElectricity extends WorkloadModuleBase {
  async submitTransaction() {
    const amount = 10 + Math.floor(Math.random() * 90);
    const payload = JSON.stringify({
      AmountMWh: amount,
      Emissions: amount * 0.05,
      ElectricityProductionMethod: "wind_onshore",
      ElapsedSeconds: 900
    });
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "backlog:AddToBacklogElectricity",
      contractArguments: [],
      transientData: { eBacklog: Buffer.from(payload).toString("base64") },
      readOnly: false,
      invokerIdentity: "eproducer1MSP_eproducer1_admin"
    });
  }
}
module.exports.createWorkloadModule = () => new AddToBacklogElectricity();
