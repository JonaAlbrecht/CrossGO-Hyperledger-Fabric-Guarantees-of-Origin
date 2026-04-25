"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
class PublishOracleData extends WorkloadModuleBase {
  async submitTransaction() {
    const now = Math.floor(Date.now() / 1000);
    const payload = JSON.stringify({
      CarrierType: "electricity",
      Zone: "DE-LU",
      PeriodStart: now - 3600,
      PeriodEnd: now,
      ProductionMethod: "solar_pv",
      EnergyUnit: "MWh",
      Quantity: 100 + Math.random() * 900,
      EmissionFactor: 0,
      DataSource: "ENTSO-E-TP",
      Attributes: {}
    });
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "oracle:PublishOracleData",
      contractArguments: [],
      transientData: { OracleData: Buffer.from(payload).toString("base64") },
      readOnly: false,
      invokerIdentity: "eissuer_admin"
    });
  }
}
module.exports.createWorkloadModule = () => new PublishOracleData();
