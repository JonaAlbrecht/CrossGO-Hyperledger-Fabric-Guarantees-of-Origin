package main

import (
	"log"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/contracts"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	issuanceContract := new(contracts.IssuanceContract)
	issuanceContract.Name = "issuance"
	issuanceContract.TransactionContextHandler = new(contractapi.TransactionContext)

	transferContract := new(contracts.TransferContract)
	transferContract.Name = "transfer"
	transferContract.TransactionContextHandler = new(contractapi.TransactionContext)

	// v10.1: ADR-033 cross-channel conversion (lock-mint-finalize protocol)
	// All carrier-to-carrier conversions span channels per ADR-030, so one contract covers all.
	conversionContract := new(contracts.ConversionCrossChannelContract)
	conversionContract.Name = "conversion"
	conversionContract.TransactionContextHandler = new(contractapi.TransactionContext)

	// v10.1: Dedicated backlog contract for UI integration
	backlogContract := new(contracts.BacklogContract)
	backlogContract.Name = "backlog"
	backlogContract.TransactionContextHandler = new(contractapi.TransactionContext)

	cancellationContract := new(contracts.CancellationContract)
	cancellationContract.Name = "cancellation"
	cancellationContract.TransactionContextHandler = new(contractapi.TransactionContext)

	queryContract := new(contracts.QueryContract)
	queryContract.Name = "query"
	queryContract.TransactionContextHandler = new(contractapi.TransactionContext)

	deviceContract := new(contracts.DeviceContract)
	deviceContract.Name = "device"
	deviceContract.TransactionContextHandler = new(contractapi.TransactionContext)

	adminContract := new(contracts.AdminContract)
	adminContract.Name = "admin"
	adminContract.TransactionContextHandler = new(contractapi.TransactionContext)

	// v10.0: Single bridge contract for all country-to-country transfers (already carrier-agnostic)
	bridgeContract := new(contracts.BridgeContract)
	bridgeContract.Name = "bridge"
	bridgeContract.TransactionContextHandler = new(contractapi.TransactionContext)

	oracleContract := new(contracts.OracleContract)
	oracleContract.Name = "oracle"
	oracleContract.TransactionContextHandler = new(contractapi.TransactionContext)

	// v10.0: Biogas and HeatingCooling contracts removed — functionality merged into
	// IssuanceContract (create), CancellationContract (cancel), and TransferContract (transfer)
	// v10.1: Backlog management extracted to dedicated BacklogContract with query functions for UI

	chaincode, err := contractapi.NewChaincode(
		issuanceContract,
		transferContract,
		conversionContract,
		backlogContract,
		cancellationContract,
		queryContract,
		deviceContract,
		adminContract,
		bridgeContract,
		oracleContract,
	)
	if err != nil {
		log.Panicf("Error creating GO lifecycle chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting GO lifecycle chaincode: %v", err)
	}
}
