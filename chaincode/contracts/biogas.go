package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// BiogasContract groups biogas GO issuance and cancellation functions.
// ADR-015: Extends the multi-carrier model with biogas support per RED III.
type BiogasContract struct {
	contractapi.Contract
}

// CreateBiogasGO creates a new biogas guarantee of origin from metering data.
// Transient key: "bGO" containing VolumeNm3, EnergyContentMWh, Emissions,
// BiogasProductionMethod, FeedstockType, ElapsedSeconds.
func (c *BiogasContract) CreateBiogasGO(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can create biogas GOs: %v", err)
	}

	type bGOTransientInput struct {
		VolumeNm3             json.Number `json:"VolumeNm3"`
		EnergyContentMWh      json.Number `json:"EnergyContentMWh"`
		Emissions             json.Number `json:"Emissions"`
		BiogasProductionMethod string      `json:"BiogasProductionMethod"`
		FeedstockType         string      `json:"FeedstockType"`
		ElapsedSeconds        json.Number `json:"ElapsedSeconds"`
	}

	var input bGOTransientInput
	if err := util.UnmarshalTransient(ctx, "bGO", &input); err != nil {
		return err
	}

	volumeNm3, err := input.VolumeNm3.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert VolumeNm3: %v", err)
	}
	energyMWh, err := input.EnergyContentMWh.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert EnergyContentMWh: %v", err)
	}
	emissions, err := input.Emissions.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert Emissions: %v", err)
	}
	elapsedSeconds, err := input.ElapsedSeconds.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert ElapsedSeconds: %v", err)
	}

	if err := util.ValidatePositive(map[string]float64{
		"VolumeNm3":        volumeNm3,
		"EnergyContentMWh": energyMWh,
		"ElapsedSeconds":   elapsedSeconds,
	}); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("BiogasProductionMethod", input.BiogasProductionMethod); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("FeedstockType", input.FeedstockType); err != nil {
		return err
	}

	// NOTE: Device attribute validation (maxOutput) deferred until Fabric CA setup.

	bGOID, err := assets.GenerateID(ctx, assets.PrefixBGO, 0)
	if err != nil {
		return fmt.Errorf("error generating bGO ID: %v", err)
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
	commitment, salt, err := assets.GenerateCommitment(ctx, volumeNm3)
	if err != nil {
		return fmt.Errorf("error generating quantity commitment: %v", err)
	}

	pub := &assets.BiogasGO{
		AssetID:               bGOID,
		CreationDateTime:      creationTime,
		GOType:                "Biogas",
		Status:                assets.GOStatusActive,
		QuantityCommitment:    commitment,
		CountryOfOrigin:       "DE",
		EnergySource:          input.BiogasProductionMethod,
		SupportScheme:         "none",
		ProductionPeriodStart: creationTime - int64(elapsedSeconds),
		ProductionPeriodEnd:   creationTime,
	}

	priv := &assets.BiogasGOPrivateDetails{
		AssetID:                bGOID,
		OwnerID:                clientMSP,
		CreationDateTime:       creationTime,
		VolumeNm3:             volumeNm3,
		EnergyContentMWh:      energyMWh,
		Emissions:             emissions,
		BiogasProductionMethod: input.BiogasProductionMethod,
		FeedstockType:         input.FeedstockType,
		ConsumptionDeclarations: []string{"none"},
		CommitmentSalt:         salt,
	}

	collection := access.GetCollectionForOrg(clientMSP)
	if err := writeBGOToLedger(ctx, pub, priv, collection); err != nil {
		return err
	}

	// ADR-016: Emit lifecycle event
	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: util.EventGOCreated,
		AssetID:   bGOID,
		GOType:    "Biogas",
		Initiator: clientMSP,
		Timestamp: creationTime,
	})
}

// writeBGOToLedger writes both the public and private parts of a biogas GO.
func writeBGOToLedger(ctx contractapi.TransactionContextInterface, pub *assets.BiogasGO, priv *assets.BiogasGOPrivateDetails, collection string) error {
	pubBytes, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal bGO public data: %v", err)
	}
	if err := ctx.GetStub().PutState(pub.AssetID, pubBytes); err != nil {
		return fmt.Errorf("failed to put bGO in public state: %v", err)
	}
	privBytes, err := json.Marshal(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal bGO private data: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, priv.AssetID, privBytes); err != nil {
		return fmt.Errorf("failed to put bGO private data: %v", err)
	}
	return nil
}

// CancelBiogasGO cancels a single biogas GO and creates a cancellation statement.
// Transient key: "CancelBiogas" containing BGOID, Collection.
func (c *BiogasContract) CancelBiogasGO(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleBuyer); err != nil {
		return fmt.Errorf("only producers and buyers can cancel bGOs: %v", err)
	}

	type cancelInput struct {
		BGOID      string `json:"BGOID"`
		Collection string `json:"Collection"`
	}

	var input cancelInput
	if err := util.UnmarshalTransient(ctx, "CancelBiogas", &input); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("BGOID", input.BGOID); err != nil {
		return err
	}

	bGOJSON, err := ctx.GetStub().GetPrivateData(input.Collection, input.BGOID)
	if err != nil {
		return fmt.Errorf("failed to read bGO %s: %v", input.BGOID, err)
	}
	if bGOJSON == nil {
		return fmt.Errorf("bGO %s does not exist in collection %s", input.BGOID, input.Collection)
	}

	var bGOPrivate assets.BiogasGOPrivateDetails
	if err := json.Unmarshal(bGOJSON, &bGOPrivate); err != nil {
		return fmt.Errorf("failed to unmarshal bGO: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	cancelKey, err := assets.GenerateID(ctx, assets.PrefixBCancellation, 0)
	if err != nil {
		return fmt.Errorf("error generating cancellation key: %v", err)
	}

	statement := assets.CancellationStatementBiogas{
		BCancellationkey:       cancelKey,
		CancellationTime:       now,
		OwnerID:                bGOPrivate.OwnerID,
		VolumeNm3:             bGOPrivate.VolumeNm3,
		EnergyContentMWh:      bGOPrivate.EnergyContentMWh,
		Emissions:             bGOPrivate.Emissions,
		BiogasProductionMethod: bGOPrivate.BiogasProductionMethod,
		FeedstockType:         bGOPrivate.FeedstockType,
	}
	stmtBytes, err := json.Marshal(statement)
	if err != nil {
		return fmt.Errorf("failed to marshal biogas cancellation: %v", err)
	}

	// ADR-007: Tombstone — mark as cancelled instead of deleting
	bGOPubJSON, err := ctx.GetStub().GetState(input.BGOID)
	if err != nil {
		return fmt.Errorf("error reading bGO public state: %v", err)
	}
	if bGOPubJSON != nil {
		var bGOPub assets.BiogasGO
		if err := json.Unmarshal(bGOPubJSON, &bGOPub); err != nil {
			return fmt.Errorf("error unmarshalling bGO public: %v", err)
		}
		bGOPub.Status = assets.GOStatusCancelled
		updatedBytes, err := json.Marshal(bGOPub)
		if err != nil {
			return fmt.Errorf("error marshalling tombstoned bGO: %v", err)
		}
		if err := ctx.GetStub().PutState(input.BGOID, updatedBytes); err != nil {
			return fmt.Errorf("error writing tombstoned bGO: %v", err)
		}
	}

	if err := ctx.GetStub().PutPrivateData(input.Collection, cancelKey, stmtBytes); err != nil {
		return fmt.Errorf("failed to write biogas cancellation: %v", err)
	}

	// ADR-016: Emit lifecycle event
	clientMSP, _ := access.GetClientMSPID(ctx)
	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: util.EventGOCancelled,
		AssetID:   input.BGOID,
		GOType:    "Biogas",
		Initiator: clientMSP,
		Timestamp: now,
	})
}
