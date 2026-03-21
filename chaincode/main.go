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

	conversionContract := new(contracts.ConversionContract)
	conversionContract.Name = "conversion"
	conversionContract.TransactionContextHandler = new(contractapi.TransactionContext)

	cancellationContract := new(contracts.CancellationContract)
	cancellationContract.Name = "cancellation"
	cancellationContract.TransactionContextHandler = new(contractapi.TransactionContext)

	queryContract := new(contracts.QueryContract)
	queryContract.Name = "query"
	queryContract.TransactionContextHandler = new(contractapi.TransactionContext)

	deviceContract := new(contracts.DeviceContract)
	deviceContract.Name = "device"
	deviceContract.TransactionContextHandler = new(contractapi.TransactionContext)

	chaincode, err := contractapi.NewChaincode(
		issuanceContract,
		transferContract,
		conversionContract,
		cancellationContract,
		queryContract,
		deviceContract,
	)
	if err != nil {
		log.Panicf("Error creating GO lifecycle chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting GO lifecycle chaincode: %v", err)
	}
}
