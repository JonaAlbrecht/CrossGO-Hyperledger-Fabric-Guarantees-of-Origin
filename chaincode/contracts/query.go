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

// DefaultPageSize is the default number of records per page when no PageSize is specified.
const DefaultPageSize = 50

// MaxPageSize prevents excessively large result sets (ADR-006).
const MaxPageSize = 200

// PaginatedResult wraps a paginated query response with a bookmark for the next page.
type PaginatedResult struct {
	Records  interface{} `json:"records"`
	Bookmark string      `json:"bookmark"`
	Count    int32       `json:"count"`
}

// QueryContract groups all read/query functions.
type QueryContract struct {
	contractapi.Contract
}

// GetCurrentEGOsList returns all active electricity GOs from the public world state.
// ADR-007: Filters out cancelled/transferred GOs (tombstone pattern).
// DEPRECATED (ADR-022, v6.0): Use GetCurrentEGOsListPaginated instead.
// This function performs unbounded range scans that degrade CouchDB performance
// as GO count grows. It will be removed in v8.0.
func (c *QueryContract) GetCurrentEGOsList(ctx contractapi.TransactionContextInterface) ([]*assets.ElectricityGO, error) {
	log.Println("WARNING: GetCurrentEGOsList is deprecated (ADR-022). Use GetCurrentEGOsListPaginated.")
	resultsIterator, err := ctx.GetStub().GetStateByRange("eGO", "eGO~")
	if err != nil {
		return nil, fmt.Errorf("error getting eGO state range: %v", err)
	}
	defer resultsIterator.Close()
	all, err := util.ConstructEGOsFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	// ADR-007: only return active GOs
	var active []*assets.ElectricityGO
	for _, ego := range all {
		if ego.Status == assets.GOStatusActive || ego.Status == "" {
			active = append(active, ego)
		}
	}
	return active, nil
}

// GetCurrentEGOsListPaginated returns a paginated list of active electricity GOs.
// ADR-006: Accepts pageSize and bookmark for cursor-based pagination.
func (c *QueryContract) GetCurrentEGOsListPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (*PaginatedResult, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination("eGO", "eGO~", pageSize, bookmark)
	if err != nil {
		return nil, fmt.Errorf("error getting paginated eGO range: %v", err)
	}
	defer resultsIterator.Close()
	eGOs, err := util.ConstructEGOsFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	// ADR-007: filter tombstoned
	var active []*assets.ElectricityGO
	for _, ego := range eGOs {
		if ego.Status == assets.GOStatusActive || ego.Status == "" {
			active = append(active, ego)
		}
	}
	return &PaginatedResult{
		Records:  active,
		Bookmark: metadata.GetBookmark(),
		Count:    metadata.GetFetchedRecordsCount(),
	}, nil
}

// GetCurrentHGOsList returns all active hydrogen GOs from the public world state.
// ADR-007: Filters out cancelled/transferred GOs.
// DEPRECATED (ADR-022, v6.0): Use GetCurrentHGOsListPaginated instead.
func (c *QueryContract) GetCurrentHGOsList(ctx contractapi.TransactionContextInterface) ([]*assets.GreenHydrogenGO, error) {
	log.Println("WARNING: GetCurrentHGOsList is deprecated (ADR-022). Use GetCurrentHGOsListPaginated.")
	resultsIterator, err := ctx.GetStub().GetStateByRange("hGO", "hGO~")
	if err != nil {
		return nil, fmt.Errorf("error getting hGO state range: %v", err)
	}
	defer resultsIterator.Close()
	all, err := util.ConstructHGOsFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	var active []*assets.GreenHydrogenGO
	for _, hgo := range all {
		if hgo.Status == assets.GOStatusActive || hgo.Status == "" {
			active = append(active, hgo)
		}
	}
	return active, nil
}

// GetCurrentHGOsListPaginated returns a paginated list of active hydrogen GOs.
// ADR-006: Accepts pageSize and bookmark for cursor-based pagination.
func (c *QueryContract) GetCurrentHGOsListPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (*PaginatedResult, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination("hGO", "hGO~", pageSize, bookmark)
	if err != nil {
		return nil, fmt.Errorf("error getting paginated hGO range: %v", err)
	}
	defer resultsIterator.Close()
	hGOs, err := util.ConstructHGOsFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	var active []*assets.GreenHydrogenGO
	for _, hgo := range hGOs {
		if hgo.Status == assets.GOStatusActive || hgo.Status == "" {
			active = append(active, hgo)
		}
	}
	return &PaginatedResult{
		Records:  active,
		Bookmark: metadata.GetBookmark(),
		Count:    metadata.GetFetchedRecordsCount(),
	}, nil
}

// VerifyQuantityCommitment verifies that a disclosed quantity and salt match the on-chain commitment.
// ADR-009: Enables selective disclosure — a verifier can confirm a producer's claims
// without requiring private data collection access.
func (c *QueryContract) VerifyQuantityCommitment(ctx contractapi.TransactionContextInterface, goID string, quantity float64, salt string) (bool, error) {
	// Try eGO first
	goJSON, err := ctx.GetStub().GetState(goID)
	if err != nil {
		return false, fmt.Errorf("failed to read GO: %v", err)
	}
	if goJSON == nil {
		return false, fmt.Errorf("GO %s does not exist", goID)
	}

	// Parse as generic map to extract QuantityCommitment
	var goData map[string]interface{}
	if err := json.Unmarshal(goJSON, &goData); err != nil {
		return false, fmt.Errorf("failed to unmarshal GO: %v", err)
	}
	commitment, ok := goData["QuantityCommitment"].(string)
	if !ok || commitment == "" {
		return false, fmt.Errorf("GO %s has no quantity commitment", goID)
	}

	return assets.VerifyCommitment(quantity, salt, commitment), nil
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

// GetCurrentBGOsList returns all active biogas GOs from the public world state.
// ADR-015: Biogas carrier support. ADR-007: Filters tombstoned GOs.
// DEPRECATED (ADR-022, v6.0): Use GetCurrentBGOsListPaginated instead.
func (c *QueryContract) GetCurrentBGOsList(ctx contractapi.TransactionContextInterface) ([]*assets.BiogasGO, error) {
	log.Println("WARNING: GetCurrentBGOsList is deprecated (ADR-022). Use GetCurrentBGOsListPaginated.")
	resultsIterator, err := ctx.GetStub().GetStateByRange("bGO", "bGO~")
	if err != nil {
		return nil, fmt.Errorf("error getting bGO state range: %v", err)
	}
	defer resultsIterator.Close()

	var active []*assets.BiogasGO
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var bgo assets.BiogasGO
		if err := json.Unmarshal(queryResult.Value, &bgo); err != nil {
			return nil, err
		}
		if bgo.Status == assets.GOStatusActive || bgo.Status == "" {
			active = append(active, &bgo)
		}
	}
	return active, nil
}

// GetCurrentBGOsListPaginated returns a paginated list of active biogas GOs.
// ADR-006 + ADR-015.
func (c *QueryContract) GetCurrentBGOsListPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (*PaginatedResult, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination("bGO", "bGO~", pageSize, bookmark)
	if err != nil {
		return nil, fmt.Errorf("error getting paginated bGO range: %v", err)
	}
	defer resultsIterator.Close()

	var active []*assets.BiogasGO
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var bgo assets.BiogasGO
		if err := json.Unmarshal(queryResult.Value, &bgo); err != nil {
			return nil, err
		}
		if bgo.Status == assets.GOStatusActive || bgo.Status == "" {
			active = append(active, &bgo)
		}
	}
	return &PaginatedResult{
		Records:  active,
		Bookmark: metadata.GetBookmark(),
		Count:    metadata.GetFetchedRecordsCount(),
	}, nil
}

// ReadPublicBGO reads a single biogas GO from public world state.
// ADR-015: Biogas carrier support.
func (c *QueryContract) ReadPublicBGO(ctx contractapi.TransactionContextInterface, bGOID string) (*assets.BiogasGO, error) {
	bgoJSON, err := ctx.GetStub().GetState(bGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read bGO %s: %v", bGOID, err)
	}
	if bgoJSON == nil {
		return nil, fmt.Errorf("bGO %s does not exist", bGOID)
	}
	var bgo assets.BiogasGO
	if err := json.Unmarshal(bgoJSON, &bgo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bGO: %v", err)
	}
	return &bgo, nil
}

// ============================================================================
// v9.0 Heating/Cooling GO query functions
// ============================================================================

// GetCurrentHCGOsList returns all active heating/cooling GOs from the public world state.
// DEPRECATED: Use GetCurrentHCGOsListPaginated instead.
func (c *QueryContract) GetCurrentHCGOsList(ctx contractapi.TransactionContextInterface) ([]*assets.HeatingCoolingGO, error) {
	log.Println("WARNING: GetCurrentHCGOsList is deprecated. Use GetCurrentHCGOsListPaginated.")
	resultsIterator, err := ctx.GetStub().GetStateByRange(assets.PrefixHCGO, assets.RangeEndHCGO)
	if err != nil {
		return nil, fmt.Errorf("error getting hcGO state range: %v", err)
	}
	defer resultsIterator.Close()

	var active []*assets.HeatingCoolingGO
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var hcgo assets.HeatingCoolingGO
		if err := json.Unmarshal(queryResult.Value, &hcgo); err != nil {
			return nil, err
		}
		if hcgo.Status == assets.GOStatusActive || hcgo.Status == "" {
			active = append(active, &hcgo)
		}
	}
	return active, nil
}

// GetCurrentHCGOsListPaginated returns a paginated list of active heating/cooling GOs.
func (c *QueryContract) GetCurrentHCGOsListPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (*PaginatedResult, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination(assets.PrefixHCGO, assets.RangeEndHCGO, pageSize, bookmark)
	if err != nil {
		return nil, fmt.Errorf("error getting paginated hcGO range: %v", err)
	}
	defer resultsIterator.Close()

	var active []*assets.HeatingCoolingGO
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var hcgo assets.HeatingCoolingGO
		if err := json.Unmarshal(queryResult.Value, &hcgo); err != nil {
			return nil, err
		}
		if hcgo.Status == assets.GOStatusActive || hcgo.Status == "" {
			active = append(active, &hcgo)
		}
	}
	return &PaginatedResult{
		Records:  active,
		Bookmark: metadata.GetBookmark(),
		Count:    metadata.GetFetchedRecordsCount(),
	}, nil
}

// ReadPublicHCGO reads a single heating/cooling GO from public world state.
func (c *QueryContract) ReadPublicHCGO(ctx contractapi.TransactionContextInterface, hcGOID string) (*assets.HeatingCoolingGO, error) {
	hcgoJSON, err := ctx.GetStub().GetState(hcGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read hcGO %s: %v", hcGOID, err)
	}
	if hcgoJSON == nil {
		return nil, fmt.Errorf("hcGO %s does not exist", hcGOID)
	}
	var hcgo assets.HeatingCoolingGO
	if err := json.Unmarshal(hcgoJSON, &hcgo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hcGO: %v", err)
	}
	return &hcgo, nil
}

// ReadPrivateBGO reads the private details of a biogas GO from a specified collection.
// Transient key: "QueryInput" containing Collection, BGOID.
func (c *QueryContract) ReadPrivateBGO(ctx contractapi.TransactionContextInterface) (*assets.BiogasGOPrivateDetails, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		BGOID      string `json:"BGOID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}
	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	bgoJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.BGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read bGO private data: %v", err)
	}
	if bgoJSON == nil {
		log.Printf("Private details for bGO %s do not exist in collection %s", input.BGOID, input.Collection)
		return nil, nil
	}

	var bgoPrivate assets.BiogasGOPrivateDetails
	if err := json.Unmarshal(bgoJSON, &bgoPrivate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bGO private data: %v", err)
	}
	return &bgoPrivate, nil
}

// ReadPrivateHCGO reads the private details of a heating/cooling GO from a specified collection.
// Transient key: "QueryInput" containing Collection, HCGOID.
func (c *QueryContract) ReadPrivateHCGO(ctx contractapi.TransactionContextInterface) (*assets.HeatingCoolingGOPrivateDetails, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
		HCGOID     string `json:"HCGOID"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}
	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	hcgoJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.HCGOID)
	if err != nil {
		return nil, fmt.Errorf("failed to read hcGO private data: %v", err)
	}
	if hcgoJSON == nil {
		log.Printf("Private details for hcGO %s do not exist in collection %s", input.HCGOID, input.Collection)
		return nil, nil
	}

	var hcgoPrivate assets.HeatingCoolingGOPrivateDetails
	if err := json.Unmarshal(hcgoJSON, &hcgoPrivate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hcGO private data: %v", err)
	}
	return &hcgoPrivate, nil
}
