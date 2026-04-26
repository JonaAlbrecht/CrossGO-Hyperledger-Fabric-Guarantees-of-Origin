"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");

class LockGOForConversion extends WorkloadModuleBase {
  constructor() {
    super();
    this.goIds = [];
    this.idx = 0;
  }

  async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
    await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

    // Query all active electricity GOs from the ledger
    let result;
    try {
      result = await this.sutAdapter.sendRequests({
        contractId: "golifecycle",
        contractFunction: "query:GetCurrentEGOsList",
        contractArguments: [],
        readOnly: true,
        channel: "electricity-de",
        invokerIdentity: "eissuerMSP_eissuer_admin"
      });
      const gos = JSON.parse(result.GetResult ? result.GetResult() : result);
      this.goIds = gos
        .filter(go => go.Status === "active")
        .map(go => go.AssetID);
    } catch (e) {
      this.goIds = [];
    }

    // Partition IDs across workers so each worker gets a distinct slice
    if (this.goIds.length > 0) {
      const perWorker = Math.ceil(this.goIds.length / totalWorkers);
      const start = workerIndex * perWorker;
      this.goIds = this.goIds.slice(start, start + perWorker);
    }

    if (this.goIds.length === 0) {
      throw new Error(
        `Worker ${workerIndex}: no active electricity GOs found — run CreateElectricityGO round first`
      );
    }
  }

  async submitTransaction() {
    const goId = this.goIds[this.idx % this.goIds.length];
    this.idx++;

    // Transient key: "LockForConversion" (matches chaincode util.UnmarshalTransient call)
    const payload = JSON.stringify({
      GOAssetID: goId,
      DestinationChannel: "hydrogen-de",
      DestinationCarrier: "hydrogen",
      ConversionMethod: "electrolysis",
      ConversionEfficiency: 0.65,
      OwnerMSP: "eproducer1MSP",
      DestinationOwnerMSP: "hproducer1MSP"
    });

    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "conversion:LockGOForConversion",
      contractArguments: [],
      transientData: { LockForConversion: Buffer.from(payload).toString("base64") },
      readOnly: false,
      channel: "electricity-de",
      invokerIdentity: "eproducer1MSP_eproducer1_admin"
    });
  }
}

module.exports.createWorkloadModule = () => new LockGOForConversion();
