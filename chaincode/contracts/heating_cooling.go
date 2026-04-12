package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// HeatingCoolingContract groups heating/cooling GO issuance and cancellation functions.
// v9.0: Extends the multi-carrier model with heating and cooling support per RED III Art. 19.
type HeatingCoolingContract struct {
	contractapi.Contract
}

// CreateHeatingCoolingGO creates a new heating/cooling guarantee of origin from metering data.
// Transient key: "hcGO" containing AmountMWh, Emissions, HeatingCoolingProductionMethod,
// SupplyTemperature, ElapsedSeconds.
func (c *HeatingCoolingContract) CreateHeatingCoolingGO(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can create heating/cooling GOs: %v", err)
	}

	type hcGOTransientInput struct {
		AmountMWh                      json.Number `json:"AmountMWh"`
		Emissions                      json.Number `json:"Emissions"`
		HeatingCoolingProductionMethod string      `json:"HeatingCoolingProductionMethod"`
		SupplyTemperature              json.Number `json:"SupplyTemperature"`
		ElapsedSeconds                 json.Number `json:"ElapsedSeconds"`
	}

	var input hcGOTransientInput
	if err := util.UnmarshalTransient(ctx, "hcGO", &input); err != nil {
		return err
	}

	amountMWh, err := input.AmountMWh.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert AmountMWh: %v", err)
	}
	emissions, err := input.Emissions.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert Emissions: %v", err)
	}
	supplyTemp, err := input.SupplyTemperature.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert SupplyTemperature: %v", err)
	}
	elapsedSeconds, err := input.ElapsedSeconds.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert ElapsedSeconds: %v", err)
	}

	if err := util.ValidatePositive(map[string]float64{
		"AmountMWh":      amountMWh,
		"ElapsedSeconds": elapsedSeconds,
	}); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("HeatingCoolingProductionMethod", input.HeatingCoolingProductionMethod); err != nil {
		return err
	}

	// NOTE: Device attribute validation (maxOutput) deferred until Fabric CA setup.

	hcGOID, err := assets.GenerateID(ctx, assets.PrefixHCGO, 0)
	if err != nil {
		return fmt.Errorf("error generating hcGO ID: %v", err)
	}

	creationTime, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	// ADR-009: Generate quantity commitment for selective disclosure
	commitment, salt, err := assets.GenerateCommitment(ctx, amountMWh)
	if err != nil {
		return fmt.Errorf("error generating quantity commitment: %v", err)
	}

	pub := &assets.HeatingCoolingGO{
		AssetID:               hcGOID,
		CreationDateTime:      creationTime,
		GOType:                "HeatingCooling",
		Status:                assets.GOStatusActive,
		QuantityCommitment:    commitment,
		CountryOfOrigin:       "DE",
		EnergySource:          input.HeatingCoolingProductionMethod,
		SupportScheme:         "none",
		ProductionPeriodStart: creationTime - int64(elapsedSeconds),
		ProductionPeriodEnd:   creationTime,
	}

	priv := &assets.HeatingCoolingGOPrivateDetails{
		AssetID:                        hcGOID,
		OwnerID:                        clientMSP,
		CreationDateTime:               creationTime,
		AmountMWh:                      amountMWh,
		Emissions:                      emissions,
		HeatingCoolingProductionMethod: input.HeatingCoolingProductionMethod,
		SupplyTemperature:              supplyTemp,
		ConsumptionDeclarations:        []string{"none"},
		CommitmentSalt:                 salt,
	}

	collection := access.GetCollectionForOrg(clientMSP)
	if err := writeHCGOToLedger(ctx, pub, priv, collection); err != nil {
		return err
	}

	// ADR-016: Emit lifecycle event
	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: util.EventGOCreated,
		AssetID:   hcGOID,
		GOType:    "HeatingCooling",
		Initiator: clientMSP,
		Timestamp: creationTime,
	})
}

// writeHCGOToLedger writes both the public and private parts of a heating/cooling GO.
func writeHCGOToLedger(ctx contractapi.TransactionContextInterface, pub *assets.HeatingCoolingGO, priv *assets.HeatingCoolingGOPrivateDetails, collection string) error {
	pubBytes, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal hcGO public data: %v", err)
	}
	if err := ctx.GetStub().PutState(pub.AssetID, pubBytes); err != nil {
		return fmt.Errorf("failed to put hcGO in public state: %v", err)
	}
	privBytes, err := json.Marshal(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal hcGO private data: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, priv.AssetID, privBytes); err != nil {
		return fmt.Errorf("failed to put hcGO private data: %v", err)
	}
	return nil
}

// CancelHeatingCoolingGO cancels a single heating/cooling GO and creates a cancellation statement.
// Transient key: "ClaimHeatingCooling" containing HCGOID, Collection.
func (c *HeatingCoolingContract) CancelHeatingCoolingGO(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleBuyer); err != nil {
		return fmt.Errorf("only producers and buyers can cancel hcGOs: %v", err)
	}

	type cancelInput struct {
		HCGOID     string `json:"HCGOID"`
		Collection string `json:"Collection"`
	}

	var input cancelInput
	if err := util.UnmarshalTransient(ctx, "ClaimHeatingCooling", &input); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("HCGOID", input.HCGOID); err != nil {
		return err
	}

	hcGOJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.HCGOID)
	if err != nil {
		return fmt.Errorf("failed to read hcGO %s: %v", input.HCGOID, err)
	}
	if hcGOJSON == nil {
		return fmt.Errorf("hcGO %s does not exist in collection %s", input.HCGOID, input.Collection)
	}

	var hcGOPrivate assets.HeatingCoolingGOPrivateDetails
	if err := json.Unmarshal(hcGOJSON, &hcGOPrivate); err != nil {
		return fmt.Errorf("failed to unmarshal hcGO: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	cancelKey, err := assets.GenerateID(ctx, assets.PrefixHCCancellation, 0)
	if err != nil {
		return fmt.Errorf("error generating cancellation key: %v", err)
	}

	statement := assets.CancellationStatementHeatingCooling{
		HCCancellationKey:              cancelKey,
		CancellationTime:               now,
		OwnerID:                        hcGOPrivate.OwnerID,
		AmountMWh:                      hcGOPrivate.AmountMWh,
		Emissions:                      hcGOPrivate.Emissions,
		HeatingCoolingProductionMethod: hcGOPrivate.HeatingCoolingProductionMethod,
	}
	stmtBytes, err := json.Marshal(statement)
	if err != nil {
		return fmt.Errorf("failed to marshal heating/cooling cancellation: %v", err)
	}

	// ADR-007: Tombstone — mark as cancelled instead of deleting
	hcGOPubJSON, err := ctx.GetStub().GetState(input.HCGOID)
	if err != nil {
		return fmt.Errorf("error reading hcGO public state: %v", err)
	}
	if hcGOPubJSON != nil {
		var hcGOPub assets.HeatingCoolingGO
		if err := json.Unmarshal(hcGOPubJSON, &hcGOPub); err != nil {
			return fmt.Errorf("error unmarshalling hcGO public: %v", err)
		}
		hcGOPub.Status = assets.GOStatusCancelled
		updatedBytes, err := json.Marshal(hcGOPub)
		if err != nil {
			return fmt.Errorf("error marshalling tombstoned hcGO: %v", err)
		}
		if err := ctx.GetStub().PutState(input.HCGOID, updatedBytes); err != nil {
			return fmt.Errorf("error writing tombstoned hcGO: %v", err)
		}
	}

	if err := ctx.GetStub().PutPrivateData(input.Collection, cancelKey, stmtBytes); err != nil {
		return fmt.Errorf("failed to write heating/cooling cancellation: %v", err)
	}

	// ADR-016: Emit lifecycle event
	clientMSP, _ := access.GetClientMSPID(ctx)
	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: util.EventGOCancelled,
		AssetID:   input.HCGOID,
		GOType:    "HeatingCooling",
		Initiator: clientMSP,
		Timestamp: now,
	})
}
