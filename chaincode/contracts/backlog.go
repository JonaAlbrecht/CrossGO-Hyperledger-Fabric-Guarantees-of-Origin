package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ================================================================================
// v10.0: Universal Backlog Contract
// ================================================================================
// This contract manages backlog accumulation for ALL energy carriers.
// Each carrier maintains a backlog that accumulates metering data before GO
// issuance or conversion. Backlog data is propagated to the UI via query functions.

// BacklogContract handles backlog accumulation and querying for all energy carriers.
type BacklogContract struct {
	contractapi.Contract
}

// ============================================================================
// Backlog Addition Functions — Each Carrier Type
// ============================================================================

// AddToBacklogElectricity records electricity production data to the backlog.
// Transient key: "eBacklog" containing AmountMWh, Emissions, ElectricityProductionMethod, ElapsedSeconds.
func (c *BacklogContract) AddToBacklogElectricity(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can add to electricity backlog: %v", err)
	}

	type backlogInput struct {
		AmountMWh                   json.Number `json:"AmountMWh"`
		Emissions                   json.Number `json:"Emissions"`
		ElectricityProductionMethod string      `json:"ElectricityProductionMethod"`
		ElapsedSeconds              json.Number `json:"ElapsedSeconds"`
	}

	var input backlogInput
	if err := util.UnmarshalTransient(ctx, "eBacklog", &input); err != nil {
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
	if err := util.ValidateNonEmpty("ElectricityProductionMethod", input.ElectricityProductionMethod); err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	collection := access.GetCollectionForOrg(clientMSP)

	backlogKey := assets.BacklogKeyElectricity + "_" + clientMSP
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	existingJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return fmt.Errorf("error reading existing backlog: %v", err)
	}

	var backlogPrivate assets.ElectricityBacklogPrivateDetails
	if existingJSON != nil {
		if err := json.Unmarshal(existingJSON, &backlogPrivate); err != nil {
			return fmt.Errorf("error unmarshaling existing backlog: %v", err)
		}
		backlogPrivate.AccumulatedMWh += amountMWh
		backlogPrivate.AccumulatedEmissions += emissions
		backlogPrivate.LastMeteringTimestamp = now
	} else {
		backlogPrivate = assets.ElectricityBacklogPrivateDetails{
			BacklogKey:                  backlogKey,
			OwnerMSP:                    clientMSP,
			AccumulatedMWh:              amountMWh,
			AccumulatedEmissions:        emissions,
			ElectricityProductionMethod: input.ElectricityProductionMethod,
			FirstMeteringTimestamp:      now - int64(elapsedSeconds),
			LastMeteringTimestamp:       now,
		}
	}

	backlogPublic := assets.CarrierBacklog{
		BacklogKey:  backlogKey,
		CarrierType: "Electricity",
		OwnerMSP:    clientMSP,
	}
	pubBytes, err := json.Marshal(backlogPublic)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog public data: %v", err)
	}
	if err := ctx.GetStub().PutState(backlogKey, pubBytes); err != nil {
		return fmt.Errorf("failed to put backlog in public state: %v", err)
	}

	privBytes, err := json.Marshal(backlogPrivate)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog private data: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, backlogKey, privBytes); err != nil {
		return fmt.Errorf("failed to put backlog private data: %v", err)
	}

	return nil
}

// AddToBacklogHydrogen records hydrogen production data to the backlog.
// Transient key: "hBacklog" containing Kilosproduced, EmissionsHydrogen, UsedMWh, HydrogenProductionMethod, ElapsedSeconds.
func (c *BacklogContract) AddToBacklogHydrogen(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can add to hydrogen backlog: %v", err)
	}

	type backlogInput struct {
		Kilosproduced            json.Number `json:"Kilosproduced"`
		EmissionsHydrogen        json.Number `json:"EmissionsHydrogen"`
		UsedMWh                  json.Number `json:"UsedMWh"`
		HydrogenProductionMethod string      `json:"HydrogenProductionMethod"`
		ElapsedSeconds           json.Number `json:"ElapsedSeconds"`
	}

	var input backlogInput
	if err := util.UnmarshalTransient(ctx, "hBacklog", &input); err != nil {
		return err
	}

	kilos, err := input.Kilosproduced.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert Kilosproduced: %v", err)
	}
	emissionsH, err := input.EmissionsHydrogen.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert EmissionsHydrogen: %v", err)
	}
	usedMWh, err := input.UsedMWh.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert UsedMWh: %v", err)
	}
	elapsedSeconds, err := input.ElapsedSeconds.Float64()
	if err != nil {
		return fmt.Errorf("failed to convert ElapsedSeconds: %v", err)
	}

	if err := util.ValidatePositive(map[string]float64{
		"Kilosproduced":  kilos,
		"UsedMWh":        usedMWh,
		"ElapsedSeconds": elapsedSeconds,
	}); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("HydrogenProductionMethod", input.HydrogenProductionMethod); err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	collection := access.GetCollectionForOrg(clientMSP)

	backlogKey := assets.BacklogKeyHydrogen + "_" + clientMSP
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	existingJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return fmt.Errorf("error reading existing backlog: %v", err)
	}

	var backlogPrivate assets.HydrogenBacklogPrivateDetails
	if existingJSON != nil {
		if err := json.Unmarshal(existingJSON, &backlogPrivate); err != nil {
			return fmt.Errorf("error unmarshaling existing backlog: %v", err)
		}
		backlogPrivate.AccumulatedKilosProduced += kilos
		backlogPrivate.AccumulatedEmissions += emissionsH
		backlogPrivate.AccumulatedInputMWh += usedMWh
		backlogPrivate.LastMeteringTimestamp = now
	} else {
		backlogPrivate = assets.HydrogenBacklogPrivateDetails{
			BacklogKey:               backlogKey,
			OwnerMSP:                 clientMSP,
			AccumulatedKilosProduced: kilos,
			AccumulatedEmissions:     emissionsH,
			HydrogenProductionMethod: input.HydrogenProductionMethod,
			AccumulatedInputMWh:      usedMWh,
			FirstMeteringTimestamp:   now - int64(elapsedSeconds),
			LastMeteringTimestamp:    now,
		}
	}

	backlogPublic := assets.CarrierBacklog{
		BacklogKey:  backlogKey,
		CarrierType: "Hydrogen",
		OwnerMSP:    clientMSP,
	}
	pubBytes, err := json.Marshal(backlogPublic)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog public data: %v", err)
	}
	if err := ctx.GetStub().PutState(backlogKey, pubBytes); err != nil {
		return fmt.Errorf("failed to put backlog in public state: %v", err)
	}

	privBytes, err := json.Marshal(backlogPrivate)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog private data: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, backlogKey, privBytes); err != nil {
		return fmt.Errorf("failed to put backlog private data: %v", err)
	}

	return nil
}

// AddToBacklogBiogas records biogas production data to the backlog.
// Transient key: "bBacklog" containing VolumeNm3, EnergyContentMWh, Emissions, BiogasProductionMethod, FeedstockType, ElapsedSeconds.
func (c *BacklogContract) AddToBacklogBiogas(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can add to biogas backlog: %v", err)
	}

	type backlogInput struct {
		VolumeNm3              json.Number `json:"VolumeNm3"`
		EnergyContentMWh       json.Number `json:"EnergyContentMWh"`
		Emissions              json.Number `json:"Emissions"`
		BiogasProductionMethod string      `json:"BiogasProductionMethod"`
		FeedstockType          string      `json:"FeedstockType"`
		ElapsedSeconds         json.Number `json:"ElapsedSeconds"`
	}

	var input backlogInput
	if err := util.UnmarshalTransient(ctx, "bBacklog", &input); err != nil {
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

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	collection := access.GetCollectionForOrg(clientMSP)

	backlogKey := assets.BacklogKeyBiogas + "_" + clientMSP
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	existingJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return fmt.Errorf("error reading existing backlog: %v", err)
	}

	var backlogPrivate assets.BiogasBacklogPrivateDetails
	if existingJSON != nil {
		if err := json.Unmarshal(existingJSON, &backlogPrivate); err != nil {
			return fmt.Errorf("error unmarshaling existing backlog: %v", err)
		}
		backlogPrivate.AccumulatedVolumeNm3 += volumeNm3
		backlogPrivate.AccumulatedEnergyContentMWh += energyMWh
		backlogPrivate.AccumulatedEmissions += emissions
		backlogPrivate.LastMeteringTimestamp = now
	} else {
		backlogPrivate = assets.BiogasBacklogPrivateDetails{
			BacklogKey:                  backlogKey,
			OwnerMSP:                    clientMSP,
			AccumulatedVolumeNm3:        volumeNm3,
			AccumulatedEnergyContentMWh: energyMWh,
			AccumulatedEmissions:        emissions,
			BiogasProductionMethod:      input.BiogasProductionMethod,
			FeedstockType:               input.FeedstockType,
			FirstMeteringTimestamp:      now - int64(elapsedSeconds),
			LastMeteringTimestamp:       now,
		}
	}

	backlogPublic := assets.CarrierBacklog{
		BacklogKey:  backlogKey,
		CarrierType: "Biogas",
		OwnerMSP:    clientMSP,
	}
	pubBytes, err := json.Marshal(backlogPublic)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog public data: %v", err)
	}
	if err := ctx.GetStub().PutState(backlogKey, pubBytes); err != nil {
		return fmt.Errorf("failed to put backlog in public state: %v", err)
	}

	privBytes, err := json.Marshal(backlogPrivate)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog private data: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, backlogKey, privBytes); err != nil {
		return fmt.Errorf("failed to put backlog private data: %v", err)
	}

	return nil
}

// AddToBacklogHeatingCooling records heating/cooling production data to the backlog.
// Transient key: "hcBacklog" containing AmountMWh, Emissions, HeatingCoolingProductionMethod, SupplyTemperature, ElapsedSeconds.
func (c *BacklogContract) AddToBacklogHeatingCooling(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can add to heating/cooling backlog: %v", err)
	}

	type backlogInput struct {
		AmountMWh                      json.Number `json:"AmountMWh"`
		Emissions                      json.Number `json:"Emissions"`
		HeatingCoolingProductionMethod string      `json:"HeatingCoolingProductionMethod"`
		SupplyTemperature              json.Number `json:"SupplyTemperature"`
		ElapsedSeconds                 json.Number `json:"ElapsedSeconds"`
	}

	var input backlogInput
	if err := util.UnmarshalTransient(ctx, "hcBacklog", &input); err != nil {
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

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	collection := access.GetCollectionForOrg(clientMSP)

	backlogKey := assets.BacklogKeyHeatingCooling + "_" + clientMSP
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	existingJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return fmt.Errorf("error reading existing backlog: %v", err)
	}

	var backlogPrivate assets.HeatingCoolingBacklogPrivateDetails
	if existingJSON != nil {
		if err := json.Unmarshal(existingJSON, &backlogPrivate); err != nil {
			return fmt.Errorf("error unmarshaling existing backlog: %v", err)
		}
		// Weighted average for temperature
		totalMWh := backlogPrivate.AccumulatedAmountMWh + amountMWh
		backlogPrivate.AverageSupplyTemperature = (backlogPrivate.AverageSupplyTemperature*backlogPrivate.AccumulatedAmountMWh + supplyTemp*amountMWh) / totalMWh
		backlogPrivate.AccumulatedAmountMWh = totalMWh
		backlogPrivate.AccumulatedEmissions += emissions
		backlogPrivate.LastMeteringTimestamp = now
	} else {
		backlogPrivate = assets.HeatingCoolingBacklogPrivateDetails{
			BacklogKey:                     backlogKey,
			OwnerMSP:                       clientMSP,
			AccumulatedAmountMWh:           amountMWh,
			AccumulatedEmissions:           emissions,
			HeatingCoolingProductionMethod: input.HeatingCoolingProductionMethod,
			AverageSupplyTemperature:       supplyTemp,
			FirstMeteringTimestamp:         now - int64(elapsedSeconds),
			LastMeteringTimestamp:          now,
		}
	}

	backlogPublic := assets.CarrierBacklog{
		BacklogKey:  backlogKey,
		CarrierType: "HeatingCooling",
		OwnerMSP:    clientMSP,
	}
	pubBytes, err := json.Marshal(backlogPublic)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog public data: %v", err)
	}
	if err := ctx.GetStub().PutState(backlogKey, pubBytes); err != nil {
		return fmt.Errorf("failed to put backlog in public state: %v", err)
	}

	privBytes, err := json.Marshal(backlogPrivate)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog private data: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, backlogKey, privBytes); err != nil {
		return fmt.Errorf("failed to put backlog private data: %v", err)
	}

	return nil
}

// ============================================================================
// Backlog Query Functions — For UI Integration
// ============================================================================

// GetElectricityBacklog returns the electricity backlog for the calling organization.
// Returns both public and private backlog data for UI display.
func (c *BacklogContract) GetElectricityBacklog(ctx contractapi.TransactionContextInterface) (*assets.ElectricityBacklogPrivateDetails, error) {
	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}
	collection := access.GetCollectionForOrg(clientMSP)
	backlogKey := assets.BacklogKeyElectricity + "_" + clientMSP

	backlogJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return nil, fmt.Errorf("error reading electricity backlog: %v", err)
	}
	if backlogJSON == nil {
		// Return zero backlog if none exists
		return &assets.ElectricityBacklogPrivateDetails{
			BacklogKey:         backlogKey,
			OwnerMSP:           clientMSP,
			AccumulatedMWh:     0,
			AccumulatedEmissions: 0,
		}, nil
	}

	var backlog assets.ElectricityBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return nil, fmt.Errorf("error unmarshaling electricity backlog: %v", err)
	}

	return &backlog, nil
}

// GetHydrogenBacklog returns the hydrogen backlog for the calling organization.
// Returns both public and private backlog data for UI display.
func (c *BacklogContract) GetHydrogenBacklog(ctx contractapi.TransactionContextInterface) (*assets.HydrogenBacklogPrivateDetails, error) {
	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}
	collection := access.GetCollectionForOrg(clientMSP)
	backlogKey := assets.BacklogKeyHydrogen + "_" + clientMSP

	backlogJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return nil, fmt.Errorf("error reading hydrogen backlog: %v", err)
	}
	if backlogJSON == nil {
		// Return zero backlog if none exists
		return &assets.HydrogenBacklogPrivateDetails{
			BacklogKey:               backlogKey,
			OwnerMSP:                 clientMSP,
			AccumulatedKilosProduced: 0,
			AccumulatedEmissions:     0,
			AccumulatedInputMWh:      0,
		}, nil
	}

	var backlog assets.HydrogenBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return nil, fmt.Errorf("error unmarshaling hydrogen backlog: %v", err)
	}

	return &backlog, nil
}

// GetBiogasBacklog returns the biogas backlog for the calling organization.
// Returns both public and private backlog data for UI display.
func (c *BacklogContract) GetBiogasBacklog(ctx contractapi.TransactionContextInterface) (*assets.BiogasBacklogPrivateDetails, error) {
	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}
	collection := access.GetCollectionForOrg(clientMSP)
	backlogKey := assets.BacklogKeyBiogas + "_" + clientMSP

	backlogJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return nil, fmt.Errorf("error reading biogas backlog: %v", err)
	}
	if backlogJSON == nil {
		// Return zero backlog if none exists
		return &assets.BiogasBacklogPrivateDetails{
			BacklogKey:                  backlogKey,
			OwnerMSP:                    clientMSP,
			AccumulatedVolumeNm3:        0,
			AccumulatedEnergyContentMWh: 0,
			AccumulatedEmissions:        0,
		}, nil
	}

	var backlog assets.BiogasBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return nil, fmt.Errorf("error unmarshaling biogas backlog: %v", err)
	}

	return &backlog, nil
}

// GetHeatingCoolingBacklog returns the heating/cooling backlog for the calling organization.
// Returns both public and private backlog data for UI display.
func (c *BacklogContract) GetHeatingCoolingBacklog(ctx contractapi.TransactionContextInterface) (*assets.HeatingCoolingBacklogPrivateDetails, error) {
	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}
	collection := access.GetCollectionForOrg(clientMSP)
	backlogKey := assets.BacklogKeyHeatingCooling + "_" + clientMSP

	backlogJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return nil, fmt.Errorf("error reading heating/cooling backlog: %v", err)
	}
	if backlogJSON == nil {
		// Return zero backlog if none exists
		return &assets.HeatingCoolingBacklogPrivateDetails{
			BacklogKey:               backlogKey,
			OwnerMSP:                 clientMSP,
			AccumulatedAmountMWh:     0,
			AccumulatedEmissions:     0,
			AverageSupplyTemperature: 0,
		}, nil
	}

	var backlog assets.HeatingCoolingBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return nil, fmt.Errorf("error unmarshaling heating/cooling backlog: %v", err)
	}

	return &backlog, nil
}

// GetAllBacklogs returns all backlog values for the calling organization.
// This is a convenience function for UI dashboards that need to display all carrier backlogs.
func (c *BacklogContract) GetAllBacklogs(ctx contractapi.TransactionContextInterface) (map[string]interface{}, error) {
	electricity, err := c.GetElectricityBacklog(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting electricity backlog: %v", err)
	}

	hydrogen, err := c.GetHydrogenBacklog(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting hydrogen backlog: %v", err)
	}

	biogas, err := c.GetBiogasBacklog(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting biogas backlog: %v", err)
	}

	heatingCooling, err := c.GetHeatingCoolingBacklog(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting heating/cooling backlog: %v", err)
	}

	result := map[string]interface{}{
		"Electricity":     electricity,
		"Hydrogen":        hydrogen,
		"Biogas":          biogas,
		"HeatingCooling":  heatingCooling,
	}

	return result, nil
}
