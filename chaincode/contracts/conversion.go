package contracts

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ConversionContract groups the hydrogen backlog and electricity→hydrogen conversion functions.
type ConversionContract struct {
	contractapi.Contract
}

// AddHydrogenToBacklog records hydrogen production data that will later be matched with
// electricity GOs during the IssuehGO conversion step.
// Transient key: "hGObacklog" containing Kilosproduced, EmissionsHydrogen, UsedMWh, HydrogenProductionMethod, ElapsedSeconds.
func (c *ConversionContract) AddHydrogenToBacklog(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can add to hydrogen backlog: %v", err)
	}
	if err := access.AssertAttribute(ctx, "hydrogentrustedDevice", "true"); err != nil {
		return fmt.Errorf("submitting sensor not authorized: not a trusted hydrogen OutputMeter: %v", err)
	}

	type backlogInput struct {
		Kilosproduced            json.Number `json:"Kilosproduced"`
		EmissionsHydrogen        json.Number `json:"EmissionsHydrogen"`
		UsedMWh                  json.Number `json:"UsedMWh"`
		HydrogenProductionMethod string      `json:"HydrogenProductionMethod"`
		ElapsedSeconds           json.Number `json:"ElapsedSeconds"`
	}

	var input backlogInput
	if err := util.UnmarshalTransient(ctx, "hGObacklog", &input); err != nil {
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

	// Validate against device attributes
	maxOutputStr, err := access.GetAttribute(ctx, "maxOutput")
	if err != nil {
		return fmt.Errorf("error getting maxOutput: %v", err)
	}
	maxOutputInt, err := strconv.Atoi(maxOutputStr)
	if err != nil {
		return fmt.Errorf("maxOutput could not be converted: %v", err)
	}
	impliedOutput := kilos / elapsedSeconds
	if float64(maxOutputInt) < impliedOutput {
		return fmt.Errorf("backlog rejected — output rate is suspiciously high")
	}

	kwhperkiloStr, err := access.GetAttribute(ctx, "kwhperkilo")
	if err != nil {
		return fmt.Errorf("error getting kwhperkilo: %v", err)
	}
	kwhperkiloInt, err := strconv.Atoi(kwhperkiloStr)
	if err != nil {
		return fmt.Errorf("kwhperkilo could not be converted: %v", err)
	}
	impliedKwhperkilo := usedMWh / kilos
	if float64(kwhperkiloInt) < impliedKwhperkilo {
		return fmt.Errorf("backlog rejected — kwh per kilo is suspiciously high")
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	collection := access.GetCollectionForOrg(clientMSP)

	// Check if a backlog already exists and accumulate
	backlogKey := "hydrogenbacklog"
	existingJSON, err := ctx.GetStub().GetPrivateData(collection, backlogKey)
	if err != nil {
		return fmt.Errorf("error reading existing backlog: %v", err)
	}

	var backlogPrivate assets.GreenHydrogenGOBacklogPrivateDetails
	if existingJSON != nil {
		if err := json.Unmarshal(existingJSON, &backlogPrivate); err != nil {
			return fmt.Errorf("error unmarshaling existing backlog: %v", err)
		}
		backlogPrivate.Kilosproduced += kilos
		backlogPrivate.EmissionsHydrogen += emissionsH
		backlogPrivate.UsedMWh += usedMWh
	} else {
		backlogPrivate = assets.GreenHydrogenGOBacklogPrivateDetails{
			Backlogkey:               backlogKey,
			OwnerID:                  clientMSP,
			Kilosproduced:            kilos,
			EmissionsHydrogen:        emissionsH,
			HydrogenProductionMethod: input.HydrogenProductionMethod,
			UsedMWh:                  usedMWh,
		}
	}

	// Write public backlog marker
	backlogPublic := assets.GreenHydrogenGOBacklog{
		Backlogkey: backlogKey,
		GOType:     "Hydrogen",
	}
	pubBytes, err := json.Marshal(backlogPublic)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog public data: %v", err)
	}
	if err := ctx.GetStub().PutState(backlogKey, pubBytes); err != nil {
		return fmt.Errorf("failed to put backlog in public state: %v", err)
	}

	// Write private backlog
	privBytes, err := json.Marshal(backlogPrivate)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog private data: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, backlogKey, privBytes); err != nil {
		return fmt.Errorf("failed to put backlog private data: %v", err)
	}

	return nil
}

// IssuehGO converts electricity GOs into a hydrogen GO using the backlog.
// This reads the hydrogen backlog, consumes electricity GOs proportionally,
// issues consumption declarations for each consumed eGO, and creates a new hGO.
// Transient key: "IssueInput" containing EGOList ("+"-separated IDs).
//
// Bug fixes applied: #3 (emissions tracking), #4 (final eGO handling), #5 (iteration order).
func (c *ConversionContract) IssuehGO(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can issue hydrogen GOs: %v", err)
	}
	if err := access.AssertAttribute(ctx, "hydrogentrustedUser", "true"); err != nil {
		return fmt.Errorf("submitting user not authorized to issue hydrogen GOs: %v", err)
	}

	type issueInput struct {
		EGOList string `json:"EGOList"`
	}

	var input issueInput
	if err := util.UnmarshalTransient(ctx, "IssueInput", &input); err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	collection := access.GetCollectionForOrg(clientMSP)

	// Read backlog
	backlogJSON, err := ctx.GetStub().GetPrivateData(collection, "hydrogenbacklog")
	if err != nil {
		return fmt.Errorf("error reading hydrogen backlog: %v", err)
	}
	if backlogJSON == nil {
		return fmt.Errorf("no hydrogen backlog found")
	}

	var backlog assets.GreenHydrogenGOBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return fmt.Errorf("error unmarshaling backlog: %v", err)
	}

	eGOList := strings.Split(input.EGOList, "+")
	impliedKwhPerKilo := backlog.UsedMWh / backlog.Kilosproduced

	// Build the hydrogen GO incrementally
	var hGOPrivate assets.GreenHydrogenGOPrivateDetails
	var toDelete []string
	suffixCounter := 0

	// Bug fix #3: track accumulated hydrogen emissions separately
	var accumulatedHEmissions float64

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}
	timecheck := now - ExpiryPeriod

	for _, currentID := range eGOList {
		inputGOJSON, err := ctx.GetStub().GetPrivateData(collection, currentID)
		if err != nil {
			return fmt.Errorf("error reading eGO %s: %v", currentID, err)
		}
		if inputGOJSON == nil {
			return fmt.Errorf("eGO %s does not exist", currentID)
		}

		var inputGO assets.ElectricityGOPrivateDetails
		if err := json.Unmarshal(inputGOJSON, &inputGO); err != nil {
			return fmt.Errorf("error unmarshaling eGO %s: %v", currentID, err)
		}

		if timecheck > inputGO.CreationDateTime {
			return fmt.Errorf("eGO %s is expired", inputGO.AssetID)
		}

		// Bug fix #4: handle both cases — eGO smaller than backlog AND eGO >= backlog
		if inputGO.AmountMWh < backlog.UsedMWh {
			// eGO is fully consumed
			emissionScalingFactor := inputGO.AmountMWh / backlog.UsedMWh

			hGOPrivate.Kilosproduced += inputGO.AmountMWh / impliedKwhPerKilo
			hGOPrivate.UsedMWh += inputGO.AmountMWh
			// Bug fix #3: accumulate hydrogen emissions proportionally
			accumulatedHEmissions += backlog.EmissionsHydrogen * emissionScalingFactor

			backlog.UsedMWh -= inputGO.AmountMWh
			backlog.Kilosproduced -= inputGO.AmountMWh / impliedKwhPerKilo
			backlog.EmissionsHydrogen -= backlog.EmissionsHydrogen * emissionScalingFactor

			hGOPrivate.InputEmissions += inputGO.Emissions
			hGOPrivate.ElectricityProductionMethod = append(hGOPrivate.ElectricityProductionMethod, inputGO.ElectricityProductionMethod)
			hGOPrivate.ConsumptionDeclarations = append(hGOPrivate.ConsumptionDeclarations, inputGO.ConsumptionDeclarations...)
			hGOPrivate.ConsumptionDeclarations = append(hGOPrivate.ConsumptionDeclarations, inputGO.AssetID)

			toDelete = append(toDelete, inputGO.AssetID)

			// ADR-001: transaction-ID-derived deterministic ID
			consumptionKey, err := assets.GenerateID(ctx, assets.PrefixEConsumption, suffixCounter)
			if err != nil {
				return fmt.Errorf("error generating consumption ID: %v", err)
			}
			suffixCounter++
			declaration := assets.ConsumptionDeclarationElectricity{
				Consumptionkey:              consumptionKey,
				CancelledGOID:               inputGO.AssetID,
				ConsumptionDateTime:         now,
				AmountMWh:                   inputGO.AmountMWh,
				Emissions:                   inputGO.Emissions,
				ElectricityProductionMethod: inputGO.ElectricityProductionMethod,
				ConsumptionDeclarations:     inputGO.ConsumptionDeclarations,
			}
			declBytes, err := json.Marshal(declaration)
			if err != nil {
				return fmt.Errorf("error marshaling consumption declaration: %v", err)
			}
			if err := ctx.GetStub().PutPrivateData(collection, consumptionKey, declBytes); err != nil {
				return fmt.Errorf("error writing consumption declaration: %v", err)
			}
		} else {
			// Bug fix #4: eGO amount >= remaining backlog — handle final eGO
			// Only consume what the backlog needs; remainder eGO stays
			neededMWh := backlog.UsedMWh
			ratio := neededMWh / inputGO.AmountMWh
			usedEmissions := ratio * inputGO.Emissions

			hGOPrivate.Kilosproduced += backlog.Kilosproduced
			hGOPrivate.UsedMWh += neededMWh
			accumulatedHEmissions += backlog.EmissionsHydrogen
			hGOPrivate.InputEmissions += usedEmissions
			hGOPrivate.ElectricityProductionMethod = append(hGOPrivate.ElectricityProductionMethod, inputGO.ElectricityProductionMethod)
			hGOPrivate.ConsumptionDeclarations = append(hGOPrivate.ConsumptionDeclarations, inputGO.ConsumptionDeclarations...)
			hGOPrivate.ConsumptionDeclarations = append(hGOPrivate.ConsumptionDeclarations, inputGO.AssetID)

			// If there's a remainder, create a new eGO for it
			remainderMWh := inputGO.AmountMWh - neededMWh
			if remainderMWh > 0.0001 { // floating point tolerance
				// ADR-001: transaction-ID-derived deterministic ID for remainder
				remainderEGOID, err := assets.GenerateID(ctx, assets.PrefixEGO, suffixCounter)
				if err != nil {
					return fmt.Errorf("error generating remainder eGO ID: %v", err)
				}
				suffixCounter++
				remainderPub := &assets.ElectricityGO{
					AssetID:          remainderEGOID,
					CreationDateTime: inputGO.CreationDateTime,
					GOType:           "Electricity",
				}
				remainderPriv := &assets.ElectricityGOPrivateDetails{
					AssetID:                     remainderEGOID,
					OwnerID:                     clientMSP,
					CreationDateTime:            inputGO.CreationDateTime,
					AmountMWh:                   remainderMWh,
					Emissions:                   inputGO.Emissions - usedEmissions,
					ElectricityProductionMethod: inputGO.ElectricityProductionMethod,
					ConsumptionDeclarations:     inputGO.ConsumptionDeclarations,
					DeviceID:                    inputGO.DeviceID,
				}
				if err := util.WriteEGOToLedger(ctx, remainderPub, remainderPriv, collection); err != nil {
					return fmt.Errorf("error writing remainder eGO: %v", err)
				}
			}

			// Delete original eGO
			toDelete = append(toDelete, inputGO.AssetID)

			// ADR-001: transaction-ID-derived deterministic ID
			consumptionKey, err := assets.GenerateID(ctx, assets.PrefixEConsumption, suffixCounter)
			if err != nil {
				return fmt.Errorf("error generating consumption ID: %v", err)
			}
			suffixCounter++
			declaration := assets.ConsumptionDeclarationElectricity{
				Consumptionkey:              consumptionKey,
				CancelledGOID:               inputGO.AssetID,
				ConsumptionDateTime:         now,
				AmountMWh:                   neededMWh,
				Emissions:                   usedEmissions,
				ElectricityProductionMethod: inputGO.ElectricityProductionMethod,
				ConsumptionDeclarations:     inputGO.ConsumptionDeclarations,
			}
			declBytes, err := json.Marshal(declaration)
			if err != nil {
				return fmt.Errorf("error marshaling consumption declaration: %v", err)
			}
			if err := ctx.GetStub().PutPrivateData(collection, consumptionKey, declBytes); err != nil {
				return fmt.Errorf("error writing consumption declaration: %v", err)
			}

			// Backlog is fully consumed
			backlog.UsedMWh = 0
			backlog.Kilosproduced = 0
			backlog.EmissionsHydrogen = 0
			break
		}
	}

	// ADR-001: transaction-ID-derived deterministic ID for the hydrogen GO
	hGOID, err := assets.GenerateID(ctx, assets.PrefixHGO, suffixCounter)
	if err != nil {
		return fmt.Errorf("error generating hGO ID: %v", err)
	}

	// Bug fix #3: use accumulated emissions, not the remaining backlog emissions
	hGOPrivate.EmissionsHydrogen = accumulatedHEmissions
	hGOPrivate.OwnerID = clientMSP
	hGOPrivate.AssetID = hGOID
	hGOPrivate.HydrogenProductionMethod = backlog.HydrogenProductionMethod
	hGOPrivate.CreationDateTime = now

	hGOPublic := &assets.GreenHydrogenGO{
		AssetID:          hGOID,
		CreationDateTime: now,
		GOType:           "Hydrogen",
	}

	// Delete consumed eGOs
	for _, id := range toDelete {
		if err := util.DeleteEGOFromLedger(ctx, id, collection); err != nil {
			return fmt.Errorf("error deleting consumed eGO %s: %v", id, err)
		}
	}

	// Write updated backlog remainder
	backlogRemainder := assets.GreenHydrogenGOBacklogPrivateDetails{
		Backlogkey:               "hydrogenbacklog",
		OwnerID:                  clientMSP,
		Kilosproduced:            backlog.Kilosproduced,
		EmissionsHydrogen:        backlog.EmissionsHydrogen,
		HydrogenProductionMethod: backlog.HydrogenProductionMethod,
		UsedMWh:                  backlog.UsedMWh,
	}
	backlogBytes, err := json.Marshal(backlogRemainder)
	if err != nil {
		return fmt.Errorf("failed to marshal backlog remainder: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(collection, "hydrogenbacklog", backlogBytes); err != nil {
		return fmt.Errorf("failed to write backlog remainder: %v", err)
	}

	// Write the new hydrogen GO
	if err := util.WriteHGOToLedger(ctx, hGOPublic, &hGOPrivate, collection); err != nil {
		return fmt.Errorf("error writing new hGO: %v", err)
	}

	return nil
}

// QueryHydrogenBacklog reads the current hydrogen backlog for the caller's organization.
// Transient key: "QueryInput" containing Collection.
func (c *ConversionContract) QueryHydrogenBacklog(ctx contractapi.TransactionContextInterface) (*assets.GreenHydrogenGOBacklogPrivateDetails, error) {
	type queryInput struct {
		Collection string `json:"Collection"`
	}

	var input queryInput
	if err := util.UnmarshalTransient(ctx, "QueryInput", &input); err != nil {
		return nil, err
	}

	if err := access.ValidateCollectionAccess(ctx, input.Collection); err != nil {
		return nil, err
	}

	backlogJSON, err := ctx.GetStub().GetPrivateData(input.Collection, "hydrogenbacklog")
	if err != nil {
		return nil, fmt.Errorf("error reading hydrogen backlog: %v", err)
	}
	if backlogJSON == nil {
		return nil, nil
	}

	var backlog assets.GreenHydrogenGOBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return nil, fmt.Errorf("error unmarshaling backlog: %v", err)
	}
	return &backlog, nil
}
