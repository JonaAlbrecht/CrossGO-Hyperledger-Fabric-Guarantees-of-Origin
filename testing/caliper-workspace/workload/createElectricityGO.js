"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
let counter = 0;
class CreateElectricityGO extends WorkloadModuleBase {
  async submitTransaction() {
    const amount = 50 + Math.floor(Math.random() * 200);
    const emissions = amount * 0.05;
    const payload = JSON.stringify({
      AmountMWh: amount,
      Emissions: emissions,
      ElectricityProductionMethod: "solar_pv",
      ElapsedSeconds: 3600
    });
    const transient = { eGO: Buffer.from(payload).toString("base64") };
    const request = {
      contractId: "golifecycle",
      contractFunction: "issuance:CreateElectricityGO",
      contractArguments: [],
      transientData: transient,
      readOnly: false
    };
    await this.sutAdapter.sendRequests(request);
    counter++;
  }
}
module.exports.createWorkloadModule = () => new CreateElectricityGO();
