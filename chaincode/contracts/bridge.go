package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// BridgeContract implements the cross-registry bridge protocol (ADR-024, v7.0).
// Enables export of GOs to external registries (e.g., AIB hub) and import of
// foreign GOs into this platform. This follows the AIB hub adapter pattern.
type BridgeContract struct {
	contractapi.Contract
}

// BridgeTransfer represents a GO that is being exported to or imported from an external registry.
type BridgeTransfer struct {
	TransferID       string `json:"transferId"`
	Direction        string `json:"direction"`        // "export" or "import"
	GOAssetID        string `json:"goAssetId"`        // local asset ID
	ExternalRegistry string `json:"externalRegistry"` // e.g. "AIB-hub", "NECS"
	ExternalID       string `json:"externalId"`       // ID in the external registry
	GOType           string `json:"goType"`           // "Electricity", "Hydrogen", "Biogas"
	Status           string `json:"status"`           // "pending", "confirmed", "failed", "cancelled"
	InitiatedBy      string `json:"initiatedBy"`      // MSP ID
	InitiatedAt      int64  `json:"initiatedAt"`
	ConfirmedAt      int64  `json:"confirmedAt,omitempty"`
	AmountMWh        float64 `json:"amountMWh,omitempty"`
	CountryOfOrigin  string `json:"countryOfOrigin,omitempty"`
	EnergySource     string `json:"energySource,omitempty"`
}

// Bridge transfer status constants.
const (
	BridgeStatusPending   = "pending"
	BridgeStatusConfirmed = "confirmed"
	BridgeStatusFailed    = "failed"
	BridgeStatusCancelled = "cancelled"
)

// Bridge transfer direction constants.
const (
	BridgeDirectionExport = "export"
	BridgeDirectionImport = "import"
)

// Bridge ID prefix and range.
const (
	PrefixBridge  = "bridge_"
	RangeEndBridge = "bridge_~"
)

// ExportGO initiates an export of a local GO to an external registry.
// The GO is locked (status set to "transferred") and a bridge transfer record is created.
// Only issuers can initiate cross-registry exports.
// Transient key: "BridgeExport" containing GOAssetID, ExternalRegistry, ExternalID.
func (c *BridgeContract) ExportGO(ctx contractapi.TransactionContextInterface) (*BridgeTransfer, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can export GOs: %v", err)
	}

	type exportInput struct {
		GOAssetID        string `json:"GOAssetID"`
		ExternalRegistry string `json:"ExternalRegistry"`
		ExternalID       string `json:"ExternalID"`
	}

	var input exportInput
	if err := util.UnmarshalTransient(ctx, "BridgeExport", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("GOAssetID", input.GOAssetID); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("ExternalRegistry", input.ExternalRegistry); err != nil {
		return nil, err
	}

	// Verify the GO exists and is active
	goJSON, err := ctx.GetStub().GetState(input.GOAssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read GO: %v", err)
	}
	if goJSON == nil {
		return nil, fmt.Errorf("GO %s does not exist", input.GOAssetID)
	}

	var goData map[string]interface{}
	if err := json.Unmarshal(goJSON, &goData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GO: %v", err)
	}
	status, _ := goData["Status"].(string)
	if status != assets.GOStatusActive && status != "" {
		return nil, fmt.Errorf("GO %s is not active (status: %s)", input.GOAssetID, status)
	}

	// Lock the GO by setting status to transferred
	goData["Status"] = assets.GOStatusTransferred
	updatedJSON, err := json.Marshal(goData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated GO: %v", err)
	}
	if err := ctx.GetStub().PutState(input.GOAssetID, updatedJSON); err != nil {
		return nil, fmt.Errorf("failed to update GO status: %v", err)
	}

	// Create bridge transfer record
	bridgeID, err := assets.GenerateID(ctx, PrefixBridge, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating bridge ID: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	goType, _ := goData["GOType"].(string)
	country, _ := goData["CountryOfOrigin"].(string)
	energy, _ := goData["EnergySource"].(string)

	transfer := &BridgeTransfer{
		TransferID:       bridgeID,
		Direction:        BridgeDirectionExport,
		GOAssetID:        input.GOAssetID,
		ExternalRegistry: input.ExternalRegistry,
		ExternalID:       input.ExternalID,
		GOType:           goType,
		Status:           BridgeStatusPending,
		InitiatedBy:      issuerMSP,
		InitiatedAt:      now,
		CountryOfOrigin:  country,
		EnergySource:     energy,
	}

	transferBytes, err := json.Marshal(transfer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bridge transfer: %v", err)
	}
	if err := ctx.GetStub().PutState(bridgeID, transferBytes); err != nil {
		return nil, fmt.Errorf("failed to write bridge transfer: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "GO_EXPORTED",
		AssetID:   bridgeID,
		GOType:    goType,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"goAssetId":        input.GOAssetID,
			"externalRegistry": input.ExternalRegistry,
			"externalId":       input.ExternalID,
		},
	})

	return transfer, nil
}

// ImportGO records an imported GO from an external registry.
// Only issuers can import GOs. The imported GO is created as a new asset.
// Transient key: "BridgeImport" containing ExternalRegistry, ExternalID, GOType,
// AmountMWh, CountryOfOrigin, EnergySource.
func (c *BridgeContract) ImportGO(ctx contractapi.TransactionContextInterface) (*BridgeTransfer, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can import GOs: %v", err)
	}

	type importInput struct {
		ExternalRegistry string  `json:"ExternalRegistry"`
		ExternalID       string  `json:"ExternalID"`
		GOType           string  `json:"GOType"`
		AmountMWh        float64 `json:"AmountMWh"`
		CountryOfOrigin  string  `json:"CountryOfOrigin"`
		EnergySource     string  `json:"EnergySource"`
	}

	var input importInput
	if err := util.UnmarshalTransient(ctx, "BridgeImport", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("ExternalRegistry", input.ExternalRegistry); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("ExternalID", input.ExternalID); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("GOType", input.GOType); err != nil {
		return nil, err
	}

	// ADR-018: Validate CEN-EN 16325 fields
	if err := util.ValidateCENFields(input.CountryOfOrigin, "", "", input.EnergySource, 0, 0); err != nil {
		return nil, fmt.Errorf("CEN-EN 16325 validation failed: %v", err)
	}

	bridgeID, err := assets.GenerateID(ctx, PrefixBridge, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating bridge ID: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	// Create a local GO asset for the import
	var localAssetID string
	switch input.GOType {
	case "Electricity":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixEGO, 1)
		if err != nil {
			return nil, err
		}
		commitment, salt, err := assets.GenerateCommitment(ctx, input.AmountMWh)
		if err != nil {
			return nil, err
		}
		pub := &assets.ElectricityGO{
			AssetID:            localAssetID,
			CreationDateTime:   now,
			GOType:             "Electricity",
			Status:             assets.GOStatusActive,
			QuantityCommitment: commitment,
			CountryOfOrigin:    input.CountryOfOrigin,
			EnergySource:       input.EnergySource,
		}
		priv := &assets.ElectricityGOPrivateDetails{
			AssetID:          localAssetID,
			OwnerID:          issuerMSP,
			CreationDateTime: now,
			AmountMWh:        input.AmountMWh,
			CommitmentSalt:   salt,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := util.WriteEGOToLedger(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	case "Hydrogen":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixHGO, 1)
		if err != nil {
			return nil, err
		}
		pub := &assets.GreenHydrogenGO{
			AssetID:          localAssetID,
			CreationDateTime: now,
			GOType:           "Hydrogen",
			Status:           assets.GOStatusActive,
		}
		priv := &assets.GreenHydrogenGOPrivateDetails{
			AssetID:          localAssetID,
			OwnerID:          issuerMSP,
			CreationDateTime: now,
			Kilosproduced:    input.AmountMWh, // reuse field for quantity
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := util.WriteHGOToLedger(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported GO type for import: %s", input.GOType)
	}

	transfer := &BridgeTransfer{
		TransferID:       bridgeID,
		Direction:        BridgeDirectionImport,
		GOAssetID:        localAssetID,
		ExternalRegistry: input.ExternalRegistry,
		ExternalID:       input.ExternalID,
		GOType:           input.GOType,
		Status:           BridgeStatusConfirmed,
		InitiatedBy:      issuerMSP,
		InitiatedAt:      now,
		ConfirmedAt:      now,
		AmountMWh:        input.AmountMWh,
		CountryOfOrigin:  input.CountryOfOrigin,
		EnergySource:     input.EnergySource,
	}

	transferBytes, err := json.Marshal(transfer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bridge transfer: %v", err)
	}
	if err := ctx.GetStub().PutState(bridgeID, transferBytes); err != nil {
		return nil, fmt.Errorf("failed to write bridge transfer: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "GO_IMPORTED",
		AssetID:   bridgeID,
		GOType:    input.GOType,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"localAssetId":     localAssetID,
			"externalRegistry": input.ExternalRegistry,
			"externalId":       input.ExternalID,
		},
	})

	return transfer, nil
}

// ConfirmExport confirms that an export has been accepted by the external registry.
// Only issuers can confirm exports.
func (c *BridgeContract) ConfirmExport(ctx contractapi.TransactionContextInterface, transferID string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can confirm exports: %v", err)
	}

	transferBytes, err := ctx.GetStub().GetState(transferID)
	if err != nil {
		return fmt.Errorf("failed to read bridge transfer: %v", err)
	}
	if transferBytes == nil {
		return fmt.Errorf("bridge transfer %s does not exist", transferID)
	}

	var transfer BridgeTransfer
	if err := json.Unmarshal(transferBytes, &transfer); err != nil {
		return fmt.Errorf("failed to unmarshal transfer: %v", err)
	}
	if transfer.Status != BridgeStatusPending {
		return fmt.Errorf("transfer %s is not pending (status: %s)", transferID, transfer.Status)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	transfer.Status = BridgeStatusConfirmed
	transfer.ConfirmedAt = now

	updatedBytes, err := json.Marshal(transfer)
	if err != nil {
		return fmt.Errorf("failed to marshal updated transfer: %v", err)
	}
	return ctx.GetStub().PutState(transferID, updatedBytes)
}

// GetBridgeTransfer reads a bridge transfer record by ID.
func (c *BridgeContract) GetBridgeTransfer(ctx contractapi.TransactionContextInterface, transferID string) (*BridgeTransfer, error) {
	transferBytes, err := ctx.GetStub().GetState(transferID)
	if err != nil {
		return nil, fmt.Errorf("failed to read bridge transfer: %v", err)
	}
	if transferBytes == nil {
		return nil, fmt.Errorf("bridge transfer %s does not exist", transferID)
	}

	var transfer BridgeTransfer
	if err := json.Unmarshal(transferBytes, &transfer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transfer: %v", err)
	}
	return &transfer, nil
}

// ListBridgeTransfersPaginated returns paginated bridge transfer records.
func (c *BridgeContract) ListBridgeTransfersPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination(PrefixBridge, RangeEndBridge, pageSize, bookmark)
	if err != nil {
		return "", fmt.Errorf("error querying bridge transfers: %v", err)
	}
	defer resultsIterator.Close()

	var transfers []*BridgeTransfer
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		var transfer BridgeTransfer
		if err := json.Unmarshal(queryResult.Value, &transfer); err != nil {
			return "", err
		}
		transfers = append(transfers, &transfer)
	}

	result := struct {
		Transfers []*BridgeTransfer `json:"transfers"`
		Bookmark  string            `json:"bookmark"`
		Count     int32             `json:"count"`
	}{
		Transfers: transfers,
		Bookmark:  metadata.GetBookmark(),
		Count:     metadata.GetFetchedRecordsCount(),
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}
	return string(resultBytes), nil
}
