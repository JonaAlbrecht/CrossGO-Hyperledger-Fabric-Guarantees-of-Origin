"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class CreateHydrogenGO extends WorkloadModuleBase {
  async submitTransaction() {
    const kilos = 50 + Math.floor(Math.random() * 200);
    const payload = JSON.stringify({
      Kilosproduced: kilos,
      EmissionsHydrogen: kilos * 0.002,
      UsedMWh: kilos * 0.055,
      HydrogenProductionMethod: "pem_electrolysis",
      ElapsedSeconds: 3600
    });
    const request = {
      contractId: "golifecycle",
      contractFunction: "issuance:CreateHydrogenGO",
      contractArguments: [],
      transientData: { hGO: Buffer.from(payload).toString("base64") },
      readOnly: false
    };
    await this.sutAdapter.sendRequests(request);
  }
}
module.exports.createWorkloadModule = () => new CreateHydrogenGO();
