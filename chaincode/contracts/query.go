package contracts

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// QueryContract groups all read/query functions.
type QueryContract struct {
	contractapi.Contract
}

// GetCurrentEGOsList returns all electricity GOs from the public world state.
func (c *QueryContract) GetCurrentEGOsList(ctx contractapi.TransactionContextInterface) ([]*assets.ElectricityGO, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("eGO0", "eGO999999999")
	if err != nil {
		return nil, fmt.Errorf("error getting eGO state range: %v", err)
	}
	defer resultsIterator.Close()
	return util.ConstructEGOsFromIterator(resultsIterator)
}

// GetCurrentHGOsList returns all hydrogen GOs from the public world state.
func (c *QueryContract) GetCurrentHGOsList(ctx contractapi.TransactionContextInterface) ([]*assets.GreenHydrogenGO, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("hGO0", "hGO999999999")
	if err != nil {
		return nil, fmt.Errorf("error getting hGO state range: %v", err)
	}
	defer resultsIterator.Close()
	return util.ConstructHGOsFromIterator(resultsIterator)
}

// ReadPublicEGO reads the public (world-state) data for an electricity GO by ID.
func (c *QueryContract) ReadPublicEGO(ctx contractapi.TransactionContextInterface, eGOID string) (*assets.ElectricityGO, error) {
	eGOJSON, err := ctx.GetStub().GetState(eGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read eGO: %v", err)
	}
	if eGOJSON == nil {
		return nil, fmt.Errorf("eGO %s does not exist", eGOID)
	}
	var eGO assets.ElectricityGO
	if err := json.Unmarshal(eGOJSON, &eGO); err != nil {
		return nil, fmt.Errorf("failed to unmarshal eGO: %v", err)
	}
	return &eGO, nil
}

// ReadPublicHGO reads the public (world-state) data for a hydrogen GO by ID.
func (c *QueryContract) ReadPublicHGO(ctx contractapi.TransactionContextInterface, hGOID string) (*assets.GreenHydrogenGO, error) {
	hGOJSON, err := ctx.GetStub().GetState(hGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read hGO: %v", err)
	}
	if hGOJSON == nil {
		return nil, fmt.Errorf("hGO %s does not exist", hGOID)
	}
	var hGO assets.GreenHydrogenGO
	if err := json.Unmarshal(hGOJSON, &hGO); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hGO: %v", err)
	}
	return &hGO, nil
}

// ReadPrivateEGO reads the private details of an electricity GO from a specified collection.
// Transient key: "QueryInput" containing Collection, EGOID.
// Bug fix #12: validates collection access.
func (c *QueryContract) ReadPrivateEGO(ctx contractapi.TransactionContextInterface) (*assets.ElectricityGOPrivateDetails, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		EGOID      string `json:"EGOID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	eGOJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.EGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read eGO private data: %v", err)
	}
	if eGOJSON == nil {
		log.Printf("Private details for eGO %s do not exist in collection %s", input.EGOID, input.Collection)
		return nil, nil
	}

	var eGOPrivate assets.ElectricityGOPrivateDetails
	if err := json.Unmarshal(eGOJSON, &eGOPrivate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal eGO private data: %v", err)
	}
	return &eGOPrivate, nil
}

// ReadPrivateHGO reads the private details of a hydrogen GO from a specified collection.
// Transient key: "QueryInput" containing Collection, HGOID.
func (c *QueryContract) ReadPrivateHGO(ctx contractapi.TransactionContextInterface) (*assets.GreenHydrogenGOPrivateDetails, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		HGOID      string `json:"HGOID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	hGOJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.HGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read hGO private data: %v", err)
	}
	if hGOJSON == nil {
		log.Printf("Private details for hGO %s do not exist in collection %s", input.HGOID, input.Collection)
		return nil, nil
	}

	var hGOPrivate assets.GreenHydrogenGOPrivateDetails
	if err := json.Unmarshal(hGOJSON, &hGOPrivate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hGO private data: %v", err)
	}
	return &hGOPrivate, nil
}

// ReadCancellationStatementElectricity reads an electricity cancellation statement.
// Transient key: "QueryInput" containing Collection, ECancelID.
func (c *QueryContract) ReadCancellationStatementElectricity(ctx contractapi.TransactionContextInterface) (*assets.CancellationStatementElectricity, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		ECancelID  string `json:"eCancelID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	cancelJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.ECancelID)
	if err != nil {
		return nil, fmt.Errorf("failed to read cancellation statement: %v", err)
	}
	if cancelJSON == nil {
		log.Printf("Cancellation statement %s does not exist in collection %s", input.ECancelID, input.Collection)
		return nil, nil
	}

	var statement assets.CancellationStatementElectricity
	if err := json.Unmarshal(cancelJSON, &statement); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cancellation statement: %v", err)
	}
	return &statement, nil
}

// ReadCancellationStatementHydrogen reads a hydrogen cancellation statement.
// Transient key: "QueryInput" containing Collection, HCancelID.
func (c *QueryContract) ReadCancellationStatementHydrogen(ctx contractapi.TransactionContextInterface) (*assets.CancellationStatementHydrogen, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		HCancelID  string `json:"hCancelID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	cancelJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.HCancelID)
	if err != nil {
		return nil, fmt.Errorf("failed to read cancellation statement: %v", err)
	}
	if cancelJSON == nil {
		log.Printf("Hydrogen cancellation statement %s does not exist in collection %s", input.HCancelID, input.Collection)
		return nil, nil
	}

	var statement assets.CancellationStatementHydrogen
	if err := json.Unmarshal(cancelJSON, &statement); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cancellation statement: %v", err)
	}
	return &statement, nil
}

// ReadConsumptionDeclarationElectricity reads an electricity consumption declaration.
// Transient key: "QueryInput" containing Collection, EConsumpID.
func (c *QueryContract) ReadConsumptionDeclarationElectricity(ctx contractapi.TransactionContextInterface) (*assets.ConsumptionDeclarationElectricity, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		EConsumpID string `json:"eConsumpID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	declJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.EConsumpID)
	if err != nil {
		return nil, fmt.Errorf("failed to read consumption declaration: %v", err)
	}
	if declJSON == nil {
		log.Printf("Consumption declaration %s does not exist in collection %s", input.EConsumpID, input.Collection)
		return nil, nil
	}

	var decl assets.ConsumptionDeclarationElectricity
	if err := json.Unmarshal(declJSON, &decl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consumption declaration: %v", err)
	}
	return &decl, nil
}

// ReadConsumptionDeclarationHydrogen reads a hydrogen consumption declaration.
// Transient key: "QueryInput" containing Collection, HConsumpID.
func (c *QueryContract) ReadConsumptionDeclarationHydrogen(ctx contractapi.TransactionContextInterface) (*assets.ConsumptionDeclarationHydrogen, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		HConsumpID string `json:"hConsumpID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	declJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.HConsumpID)
	if err != nil {
		return nil, fmt.Errorf("failed to read consumption declaration: %v", err)
	}
	if declJSON == nil {
		log.Printf("Hydrogen consumption declaration %s does not exist in collection %s", input.HConsumpID, input.Collection)
		return nil, nil
	}

	var decl assets.ConsumptionDeclarationHydrogen
	if err := json.Unmarshal(declJSON, &decl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consumption declaration: %v", err)
	}
	return &decl, nil
}

// QueryPrivateEGOsByAmountMWh returns a sorted list of eGOs from a collection,
// enough to cover the requested MWh amount. Used by conversion and transfer functions.
// Bug fix #5: no longer uses the buggy forward-iteration remove() pattern.
func (c *QueryContract) QueryPrivateEGOsByAmountMWh(ctx contractapi.TransactionContextInterface) ([]*assets.ElectricityGOPrivateDetails, error) {
	type queryInput struct {
		Collection string      `json:"Collection"`
		RequestedMWh json.Number `json:"RequestedMWh"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	requestedMWh, err := input.RequestedMWh.Float64()
	if err != nil {
		return nil, fmt.Errorf("error converting RequestedMWh: %v", err)
	}

	// Read all eGOs from the collection's private data
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(input.Collection, "eGO0", "eGO999999999")
	if err != nil {
		return nil, fmt.Errorf("error querying private eGOs: %v", err)
	}
	defer resultsIterator.Close()

	var allEGOs []*assets.ElectricityGOPrivateDetails
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var eGO assets.ElectricityGOPrivateDetails
		if err := json.Unmarshal(queryResult.Value, &eGO); err != nil {
			return nil, err
		}
		allEGOs = append(allEGOs, &eGO)
	}

	// Sort by AmountMWh (ascending) for efficient packing
	// Using a simple selection — enough GOs to cover the requested amount
	var selected []*assets.ElectricityGOPrivateDetails
	var totalMWh float64
	for _, eGO := range allEGOs {
		if totalMWh >= requestedMWh {
			break
		}
		selected = append(selected, eGO)
		totalMWh += eGO.AmountMWh
	}

	if totalMWh < requestedMWh {
		return nil, fmt.Errorf("insufficient eGOs: only %.4f MWh available, %.4f requested", totalMWh, requestedMWh)
	}

	return selected, nil
}

// QueryPrivateHGOsByAmount returns a list of hGOs from a collection,
// enough to cover the requested kilogram amount.
func (c *QueryContract) QueryPrivateHGOsByAmount(ctx contractapi.TransactionContextInterface) ([]*assets.GreenHydrogenGOPrivateDetails, error) {
	type queryInput struct {
		Collection    string      `json:"Collection"`
		RequestedKilos json.Number `json:"RequestedKilos"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	requestedKilos, err := input.RequestedKilos.Float64()
	if err != nil {
		return nil, fmt.Errorf("error converting RequestedKilos: %v", err)
	}

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(input.Collection, "hGO0", "hGO999999999")
	if err != nil {
		return nil, fmt.Errorf("error querying private hGOs: %v", err)
	}
	defer resultsIterator.Close()

	var allHGOs []*assets.GreenHydrogenGOPrivateDetails
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var hGO assets.GreenHydrogenGOPrivateDetails
		if err := json.Unmarshal(queryResult.Value, &hGO); err != nil {
			return nil, err
		}
		allHGOs = append(allHGOs, &hGO)
	}

	var selected []*assets.GreenHydrogenGOPrivateDetails
	var totalKilos float64
	for _, hGO := range allHGOs {
		if totalKilos >= requestedKilos {
			break
		}
		selected = append(selected, hGO)
		totalKilos += hGO.Kilosproduced
	}

	if totalKilos < requestedKilos {
		return nil, fmt.Errorf("insufficient hGOs: only %.4f kg available, %.4f requested", totalKilos, requestedKilos)
	}

	return selected, nil
}
