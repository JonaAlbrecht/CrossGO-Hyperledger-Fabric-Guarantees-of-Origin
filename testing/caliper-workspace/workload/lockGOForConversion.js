"use strict";
const { WorkloadModuleBase } = require("@hyperledger/caliper-core");

class LockGOForConversion extends WorkloadModuleBase {
  async submitTransaction() {
    // Generate a test GO ID
    const goId = `EGO-${Math.floor(Math.random() * 10000).toString().padStart(5, '0')}`;
    const amountMWh = 10 + Math.floor(Math.random() * 90);
    
    const payload = JSON.stringify({
      ElectricityGOID: goId,
      TargetCarrier: "hydrogen",
      ConversionRatio: 0.65,  // 65% efficiency
      AmountMWh: amountMWh
    });
    
    await this.sutAdapter.sendRequests({
      contractId: "golifecycle",
      contractFunction: "conversion:LockGOForConversion",
      contractArguments: [],
      transientData: { ConversionRequest: Buffer.from(payload).toString("base64") },
      readOnly: false,
      invokerIdentity: "_eproducer1MSP_eproducer1_admin"
    });
  }
}

module.exports.createWorkloadModule = () => new LockGOForConversion();
