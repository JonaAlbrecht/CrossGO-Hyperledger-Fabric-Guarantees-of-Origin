// Package contracts implements the smart contract logic grouped by domain.
package contracts

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// IssuanceContract groups all GO creation functions.
type IssuanceContract struct {
	contractapi.Contract
}

// CreateElectricityGO creates a new electricity guarantee of origin from SmartMeter data.
// Transient key: "eGO" containing AmountMWh, Emissions, ElapsedSeconds, ElectricityProductionMethod.
func (c *IssuanceContract) CreateElectricityGO(ctx contractapi.TransactionContextInterface) error {
	// Access control: must be a producer with a trusted electricity device
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can create electricity GOs: %v", err)
	}
	if err := access.AssertAttribute(ctx, "electricitytrustedDevice", "true"); err != nil {
		return fmt.Errorf("submitting sensor not authorized: not a trusted electricity SmartMeter: %v", err)
	}

	type eGOTransientInput struct {
		AmountMWh                   json.Number `json:"AmountMWh"`
		Emissions                   json.Number `json:"Emissions"`
		ElapsedSeconds              json.Number `json:"ElapsedSeconds"`
		ElectricityProductionMethod string      `json:"ElectricityProductionMethod"`
	}

	var input eGOTransientInput
	if err := util.UnmarshalTransient(ctx, "eGO", &input); err != nil {
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

	// Bug fix #6: validate all inputs
	if err := util.ValidatePositive(map[string]float64{
		"AmountMWh":      amountMWh,
		"Emissions":      emissions,
		"ElapsedSeconds": elapsedSeconds,
	}); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("ElectricityProductionMethod", input.ElectricityProductionMethod); err != nil {
		return err
	}

	// Validate against device attributes (efficiency and emission intensity)
	maxEfficiencyStr, err := access.GetAttribute(ctx, "maxEfficiency")
	if err != nil {
		return fmt.Errorf("error getting maxEfficiency: %v", err)
	}
	maxEfficiencyInt, err := strconv.Atoi(maxEfficiencyStr)
	if err != nil {
		return fmt.Errorf("maxEfficiency could not be converted: %v", err)
	}
	impliedEfficiency := amountMWh / elapsedSeconds
	if float64(maxEfficiencyInt) < impliedEfficiency {
		return fmt.Errorf("GO rejected — efficiency is suspiciously high")
	}

	emissionIntensityStr, err := access.GetAttribute(ctx, "emissionIntensity")
	if err != nil {
		return fmt.Errorf("error getting emissionIntensity: %v", err)
	}
	emissionIntensityInt, err := strconv.Atoi(emissionIntensityStr)
	if err != nil {
		return fmt.Errorf("error converting emissionIntensity: %v", err)
	}
	impliedEmissionIntensity := (emissions / elapsedSeconds) * 3600
	if float64(emissionIntensityInt) > impliedEmissionIntensity {
		return fmt.Errorf("GO rejected — emissions are suspiciously low")
	}

	technologyType, err := access.GetAttribute(ctx, "technologyType")
	if err != nil {
		return fmt.Errorf("error getting technologyType: %v", err)
	}
	if technologyType != input.ElectricityProductionMethod {
		return fmt.Errorf("production method mismatch: expected %s, got %s", technologyType, input.ElectricityProductionMethod)
	}

	// Bug fix #1, #2: persistent on-chain counter
	nextID, err := assets.GetNextID(ctx, assets.CounterKeyEGO)
	if err != nil {
		return fmt.Errorf("error getting next eGO ID: %v", err)
	}
	eGOID := "eGO" + strconv.Itoa(nextID)

	creationTime, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	// Bug fix #7: use GetMSPID() for collection name, not "organization" attribute
	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	pub := &assets.ElectricityGO{
		AssetID:          eGOID,
		CreationDateTime: creationTime,
		GOType:           "Electricity",
	}

	priv := &assets.ElectricityGOPrivateDetails{
		AssetID:                     eGOID,
		OwnerID:                     clientMSP,
		CreationDateTime:            creationTime,
		AmountMWh:                   amountMWh,
		Emissions:                   emissions,
		ElectricityProductionMethod: input.ElectricityProductionMethod,
		ConsumptionDeclarations:     []string{"none"},
	}

	collection := access.GetCollectionForOrg(clientMSP)
	return util.WriteEGOToLedger(ctx, pub, priv, collection)
}

// CreateHydrogenGO creates a new hydrogen guarantee of origin (not from conversion).
// This is for direct hydrogen production metering — separate from the conversion flow.
// Transient key: "hGO" containing Kilosproduced, EmissionsHydrogen, UsedMWh, HydrogenProductionMethod, ElapsedSeconds.
func (c *IssuanceContract) CreateHydrogenGO(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can create hydrogen GOs: %v", err)
	}
	if err := access.AssertAttribute(ctx, "hydrogentrustedDevice", "true"); err != nil {
		return fmt.Errorf("submitting sensor not authorized: not a trusted hydrogen OutputMeter: %v", err)
	}

	type hGOTransientInput struct {
		Kilosproduced            json.Number `json:"Kilosproduced"`
		EmissionsHydrogen        json.Number `json:"EmissionsHydrogen"`
		UsedMWh                  json.Number `json:"UsedMWh"`
		HydrogenProductionMethod string      `json:"HydrogenProductionMethod"`
		ElapsedSeconds           json.Number `json:"ElapsedSeconds"`
	}

	var input hGOTransientInput
	if err := util.UnmarshalTransient(ctx, "hGO", &input); err != nil {
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
		return fmt.Errorf("GO rejected — output rate is suspiciously high")
	}

	nextID, err := assets.GetNextID(ctx, assets.CounterKeyHGO)
	if err != nil {
		return fmt.Errorf("error getting next hGO ID: %v", err)
	}
	hGOID := "hGO" + strconv.Itoa(nextID)

	creationTime, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	pub := &assets.GreenHydrogenGO{
		AssetID:          hGOID,
		CreationDateTime: creationTime,
		GOType:           "Hydrogen",
	}

	priv := &assets.GreenHydrogenGOPrivateDetails{
		AssetID:                     hGOID,
		OwnerID:                     clientMSP,
		CreationDateTime:            creationTime,
		Kilosproduced:               kilos,
		EmissionsHydrogen:           emissionsH,
		HydrogenProductionMethod:    input.HydrogenProductionMethod,
		InputEmissions:              0,
		UsedMWh:                     usedMWh,
		ElectricityProductionMethod: []string{},
		ConsumptionDeclarations:     []string{"none"},
	}

	collection := access.GetCollectionForOrg(clientMSP)
	return util.WriteHGOToLedger(ctx, pub, priv, collection)
}
