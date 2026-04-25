"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class AddToBacklogHydrogen extends WorkloadModuleBase {
  async submitTransaction() {
    const kilos = 10 + Math.floor(Math.random() * 90);
    const payload = JSON.stringify({
      Kilosproduced: kilos,
      EmissionsHydrogen: kilos * 0.001,
      UsedMWh: kilos * 0.055,
      HydrogenProductionMethod: "pem_electrolysis",
      ElapsedSeconds: 900
    });
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "backlog:AddToBacklogHydrogen",
      contractArguments: [],
      transientData: { hBacklog: Buffer.from(payload).toString("base64") },
      readOnly: false
    });
  }
}
module.exports.createWorkloadModule = () => new AddToBacklogHydrogen();
