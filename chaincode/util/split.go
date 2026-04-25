package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TransferConsumptionDeclarations copies all consumption declaration private data
// from the sender's collection to the receiver's collection and deletes the originals
// from the sender. "split" and "none" markers are skipped.
func TransferConsumptionDeclarations(
	ctx contractapi.TransactionContextInterface,
	declarations []string,
	senderCollection, receiverCollection string,
	deleteFromSender bool,
) error {
	for _, declKey := range declarations {
		if !strings.HasPrefix(declKey, "eCon") && !strings.HasPrefix(declKey, "hCon") {
			continue
		}
		declJSON, err := ctx.GetStub().GetPrivateData(senderCollection, declKey)
		if err != nil {
			return fmt.Errorf("error reading consumption declaration %s: %v", declKey, err)
		}
		if declJSON == nil {
			continue
		}
		err = ctx.GetStub().PutPrivateData(receiverCollection, declKey, declJSON)
		if err != nil {
			return fmt.Errorf("error writing consumption declaration %s to receiver: %v", declKey, err)
		}
		if deleteFromSender {
			err = ctx.GetStub().DelPrivateData(senderCollection, declKey)
			if err != nil {
				return fmt.Errorf("error deleting consumption declaration %s from sender: %v", declKey, err)
			}
		}
	}
	return nil
}

// SplitElectricityGO performs a proportional split of an ElectricityGO.
// Given an eGO and the amount to take, it returns:
//   - takenPrivate: the portion transferred/cancelled (with takenAmount MWh)
//   - remainderPrivate: the portion left with the original owner
//   - remainderPublic: the new public GO for the remainder (with a new ID)
//
// Bug fix #8: The remainder preserves the original CreationDateTime.
func SplitElectricityGO(
	ctx contractapi.TransactionContextInterface,
	original *assets.ElectricityGOPrivateDetails,
	takenAmount float64,
	newOwnerID string,
) (taken *assets.ElectricityGOPrivateDetails, remainderPrivate *assets.ElectricityGOPrivateDetails, remainderPublic *assets.ElectricityGO, err error) {
	excessAmount := original.AmountMWh - takenAmount
	ratio := excessAmount / original.AmountMWh
	takenEmissions := (1 - ratio) * original.Emissions
	excessEmissions := ratio * original.Emissions

	declarations := make([]string, len(original.ConsumptionDeclarations))
	copy(declarations, original.ConsumptionDeclarations)
	declarations = append(declarations, "split")

	taken = &assets.ElectricityGOPrivateDetails{
		AssetID:                     original.AssetID,
		OwnerID:                     newOwnerID,
		CreationDateTime:            original.CreationDateTime,
		AmountMWh:                   takenAmount,
		Emissions:                   takenEmissions,
		ElectricityProductionMethod: original.ElectricityProductionMethod,
		ConsumptionDeclarations:     declarations,
		DeviceID:                    original.DeviceID,
	}

	// ADR-001: transaction-ID-derived deterministic ID (no shared counter)
	remainderID, err := assets.GenerateID(ctx, assets.PrefixEGO, 0)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating new eGO ID for remainder: %v", err)
	}

	remainderPublic = &assets.ElectricityGO{
		AssetID:          remainderID,
		CreationDateTime: original.CreationDateTime, // Bug fix #8: preserve original timestamp
		GOType:           "Electricity",
		Status:           assets.GOStatusActive, // ADR-007
	}

	remainderPrivate = &assets.ElectricityGOPrivateDetails{
		AssetID:                     remainderID,
		OwnerID:                     original.OwnerID,
		CreationDateTime:            original.CreationDateTime, // Bug fix #8: preserve original timestamp
		AmountMWh:                   excessAmount,
		Emissions:                   excessEmissions,
		ElectricityProductionMethod: original.ElectricityProductionMethod,
		ConsumptionDeclarations:     declarations,
		DeviceID:                    original.DeviceID,
	}

	return taken, remainderPrivate, remainderPublic, nil
}

// SplitHydrogenGO performs a proportional split of a GreenHydrogenGO.
// Bug fix #8: The remainder preserves the original CreationDateTime.
// Bug fix #9: ConsumptionDeclarations are deep-copied (not shared between split halves).
func SplitHydrogenGO(
	ctx contractapi.TransactionContextInterface,
	original *assets.GreenHydrogenGOPrivateDetails,
	takenKilos float64,
	newOwnerID string,
) (taken *assets.GreenHydrogenGOPrivateDetails, remainderPrivate *assets.GreenHydrogenGOPrivateDetails, remainderPublic *assets.GreenHydrogenGO, err error) {
	excessKilos := original.Kilosproduced - takenKilos
	ratio := excessKilos / original.Kilosproduced
	takenEmissionsH := (1 - ratio) * original.EmissionsHydrogen
	excessEmissionsH := ratio * original.EmissionsHydrogen
	takenInputEmissions := (1 - ratio) * original.InputEmissions
	excessInputEmissions := ratio * original.InputEmissions
	takenUsedMWh := (1 - ratio) * original.UsedMWh
	excessUsedMWh := ratio * original.UsedMWh

	// Bug fix #9: deep-copy consumption declarations
	takenDeclarations := make([]string, len(original.ConsumptionDeclarations))
	copy(takenDeclarations, original.ConsumptionDeclarations)
	takenDeclarations = append(takenDeclarations, "split")

	remainderDeclarations := make([]string, len(original.ConsumptionDeclarations))
	copy(remainderDeclarations, original.ConsumptionDeclarations)
	remainderDeclarations = append(remainderDeclarations, "split")

	// Deep-copy production method slices
	takenMethods := make([]string, len(original.ElectricityProductionMethod))
	copy(takenMethods, original.ElectricityProductionMethod)
	remainderMethods := make([]string, len(original.ElectricityProductionMethod))
	copy(remainderMethods, original.ElectricityProductionMethod)

	taken = &assets.GreenHydrogenGOPrivateDetails{
		AssetID:                     original.AssetID,
		OwnerID:                     newOwnerID,
		CreationDateTime:            original.CreationDateTime,
		Kilosproduced:               takenKilos,
		EmissionsHydrogen:           takenEmissionsH,
		HydrogenProductionMethod:    original.HydrogenProductionMethod,
		InputEmissions:              takenInputEmissions,
		UsedMWh:                     takenUsedMWh,
		ElectricityProductionMethod: takenMethods,
		ConsumptionDeclarations:     takenDeclarations,
		DeviceID:                    original.DeviceID,
	}

	// ADR-001: transaction-ID-derived deterministic ID (no shared counter)
	remainderID, err := assets.GenerateID(ctx, assets.PrefixHGO, 0)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating new hGO ID for remainder: %v", err)
	}

	remainderPublic = &assets.GreenHydrogenGO{
		AssetID:          remainderID,
		CreationDateTime: original.CreationDateTime,
		GOType:           "Hydrogen",
		Status:           assets.GOStatusActive, // ADR-007
	}

	remainderPrivate = &assets.GreenHydrogenGOPrivateDetails{
		AssetID:                     remainderID,
		OwnerID:                     original.OwnerID,
		CreationDateTime:            original.CreationDateTime,
		Kilosproduced:               excessKilos,
		EmissionsHydrogen:           excessEmissionsH,
		HydrogenProductionMethod:    original.HydrogenProductionMethod,
		InputEmissions:              excessInputEmissions,
		UsedMWh:                     excessUsedMWh,
		ElectricityProductionMethod: remainderMethods,
		ConsumptionDeclarations:     remainderDeclarations,
		DeviceID:                    original.DeviceID,
	}

	return taken, remainderPrivate, remainderPublic, nil
}

// WriteEGOToLedger writes both the public and private parts of an electricity GO.
func WriteEGOToLedger(ctx contractapi.TransactionContextInterface, pub *assets.ElectricityGO, priv *assets.ElectricityGOPrivateDetails, collection string) error {
	pubBytes, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal eGO public data: %v", err)
	}
	err = ctx.GetStub().PutState(pub.AssetID, pubBytes)
	if err != nil {
		return fmt.Errorf("failed to put eGO in public state: %v", err)
	}
	privBytes, err := json.Marshal(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal eGO private data: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(collection, priv.AssetID, privBytes)
	if err != nil {
		return fmt.Errorf("failed to put eGO private data: %v", err)
	}
	return nil
}

// WriteHGOToLedger writes both the public and private parts of a hydrogen GO.
func WriteHGOToLedger(ctx contractapi.TransactionContextInterface, pub *assets.GreenHydrogenGO, priv *assets.GreenHydrogenGOPrivateDetails, collection string) error {
	pubBytes, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal hGO public data: %v", err)
	}
	err = ctx.GetStub().PutState(pub.AssetID, pubBytes)
	if err != nil {
		return fmt.Errorf("failed to put hGO in public state: %v", err)
	}
	privBytes, err := json.Marshal(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal hGO private data: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(collection, priv.AssetID, privBytes)
	if err != nil {
		return fmt.Errorf("failed to put hGO private data: %v", err)
	}
	return nil
}

// WriteBGOToLedger writes both the public and private parts of a biogas GO.
func WriteBGOToLedger(ctx contractapi.TransactionContextInterface, pub *assets.BiogasGO, priv *assets.BiogasGOPrivateDetails, collection string) error {
	pubBytes, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal bGO public data: %v", err)
	}
	err = ctx.GetStub().PutState(pub.AssetID, pubBytes)
	if err != nil {
		return fmt.Errorf("failed to put bGO in public state: %v", err)
	}
	privBytes, err := json.Marshal(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal bGO private data: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(collection, priv.AssetID, privBytes)
	if err != nil {
		return fmt.Errorf("failed to put bGO private data: %v", err)
	}
	return nil
}

// WriteHCGOToLedger writes both the public and private parts of a heating/cooling GO.
func WriteHCGOToLedger(ctx contractapi.TransactionContextInterface, pub *assets.HeatingCoolingGO, priv *assets.HeatingCoolingGOPrivateDetails, collection string) error {
	pubBytes, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal hcGO public data: %v", err)
	}
	err = ctx.GetStub().PutState(pub.AssetID, pubBytes)
	if err != nil {
		return fmt.Errorf("failed to put hcGO in public state: %v", err)
	}
	privBytes, err := json.Marshal(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal hcGO private data: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(collection, priv.AssetID, privBytes)
	if err != nil {
		return fmt.Errorf("failed to put hcGO private data: %v", err)
	}
	return nil
}

// SplitBiogasGO performs a proportional split of a BiogasGO.
func SplitBiogasGO(
	ctx contractapi.TransactionContextInterface,
	original *assets.BiogasGOPrivateDetails,
	takenVolumeNm3 float64,
	newOwnerID string,
) (taken *assets.BiogasGOPrivateDetails, remainderPrivate *assets.BiogasGOPrivateDetails, remainderPublic *assets.BiogasGO, err error) {
	excessVolumeNm3 := original.VolumeNm3 - takenVolumeNm3
	ratio := excessVolumeNm3 / original.VolumeNm3
	takenEnergyMWh := (1 - ratio) * original.EnergyContentMWh
	excessEnergyMWh := ratio * original.EnergyContentMWh
	takenEmissions := (1 - ratio) * original.Emissions
	excessEmissions := ratio * original.Emissions

	declarations := make([]string, len(original.ConsumptionDeclarations))
	copy(declarations, original.ConsumptionDeclarations)
	declarations = append(declarations, "split")

	taken = &assets.BiogasGOPrivateDetails{
		AssetID:                original.AssetID,
		OwnerID:                newOwnerID,
		CreationDateTime:       original.CreationDateTime,
		VolumeNm3:              takenVolumeNm3,
		EnergyContentMWh:       takenEnergyMWh,
		Emissions:              takenEmissions,
		BiogasProductionMethod: original.BiogasProductionMethod,
		FeedstockType:          original.FeedstockType,
		ConsumptionDeclarations: declarations,
		DeviceID:               original.DeviceID,
	}

	remainderID, err := assets.GenerateID(ctx, assets.PrefixBGO, 0)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating new bGO ID for remainder: %v", err)
	}

	remainderPublic = &assets.BiogasGO{
		AssetID:          remainderID,
		CreationDateTime: original.CreationDateTime,
		GOType:           "Biogas",
		Status:           assets.GOStatusActive,
	}

	remainderPrivate = &assets.BiogasGOPrivateDetails{
		AssetID:                remainderID,
		OwnerID:                original.OwnerID,
		CreationDateTime:       original.CreationDateTime,
		VolumeNm3:              excessVolumeNm3,
		EnergyContentMWh:       excessEnergyMWh,
		Emissions:              excessEmissions,
		BiogasProductionMethod: original.BiogasProductionMethod,
		FeedstockType:          original.FeedstockType,
		ConsumptionDeclarations: declarations,
		DeviceID:               original.DeviceID,
	}

	return taken, remainderPrivate, remainderPublic, nil
}

// SplitHeatingCoolingGO performs a proportional split of a HeatingCoolingGO.
func SplitHeatingCoolingGO(
	ctx contractapi.TransactionContextInterface,
	original *assets.HeatingCoolingGOPrivateDetails,
	takenAmountMWh float64,
	newOwnerID string,
) (taken *assets.HeatingCoolingGOPrivateDetails, remainderPrivate *assets.HeatingCoolingGOPrivateDetails, remainderPublic *assets.HeatingCoolingGO, err error) {
	excessAmountMWh := original.AmountMWh - takenAmountMWh
	ratio := excessAmountMWh / original.AmountMWh
	takenEmissions := (1 - ratio) * original.Emissions
	excessEmissions := ratio * original.Emissions

	declarations := make([]string, len(original.ConsumptionDeclarations))
	copy(declarations, original.ConsumptionDeclarations)
	declarations = append(declarations, "split")

	taken = &assets.HeatingCoolingGOPrivateDetails{
		AssetID:                        original.AssetID,
		OwnerID:                        newOwnerID,
		CreationDateTime:               original.CreationDateTime,
		AmountMWh:                      takenAmountMWh,
		Emissions:                      takenEmissions,
		HeatingCoolingProductionMethod: original.HeatingCoolingProductionMethod,
		SupplyTemperature:              original.SupplyTemperature,
		ConsumptionDeclarations:        declarations,
		DeviceID:                       original.DeviceID,
	}

	remainderID, err := assets.GenerateID(ctx, assets.PrefixHCGO, 0)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating new hcGO ID for remainder: %v", err)
	}

	remainderPublic = &assets.HeatingCoolingGO{
		AssetID:          remainderID,
		CreationDateTime: original.CreationDateTime,
		GOType:           "HeatingCooling",
		Status:           assets.GOStatusActive,
	}

	remainderPrivate = &assets.HeatingCoolingGOPrivateDetails{
		AssetID:                        remainderID,
		OwnerID:                        original.OwnerID,
		CreationDateTime:               original.CreationDateTime,
		AmountMWh:                      excessAmountMWh,
		Emissions:                      excessEmissions,
		HeatingCoolingProductionMethod: original.HeatingCoolingProductionMethod,
		SupplyTemperature:              original.SupplyTemperature,
		ConsumptionDeclarations:        declarations,
		DeviceID:                       original.DeviceID,
	}

	return taken, remainderPrivate, remainderPublic, nil
}

// DeleteEGOFromLedger marks an electricity GO as cancelled (tombstone) instead of deleting.
// ADR-007: Preserves audit trail by updating Status rather than calling DelState.
// Private data is retained for audit; only the public status changes.
func DeleteEGOFromLedger(ctx contractapi.TransactionContextInterface, assetID, collection string) error {
	egoJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("error reading eGO %s for tombstone: %v", assetID, err)
	}
	if egoJSON == nil {
		return fmt.Errorf("eGO %s does not exist in public state", assetID)
	}
	var ego assets.ElectricityGO
	if err := json.Unmarshal(egoJSON, &ego); err != nil {
		return fmt.Errorf("error unmarshalling eGO %s: %v", assetID, err)
	}
	ego.Status = assets.GOStatusCancelled
	updatedBytes, err := json.Marshal(ego)
	if err != nil {
		return fmt.Errorf("error marshalling tombstoned eGO %s: %v", assetID, err)
	}
	if err := ctx.GetStub().PutState(assetID, updatedBytes); err != nil {
		return fmt.Errorf("error writing tombstoned eGO %s: %v", assetID, err)
	}
	return nil
}

// DeleteHGOFromLedger marks a hydrogen GO as cancelled (tombstone) instead of deleting.
// ADR-007: Preserves audit trail by updating Status rather than calling DelState.
func DeleteHGOFromLedger(ctx contractapi.TransactionContextInterface, assetID, collection string) error {
	hgoJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("error reading hGO %s for tombstone: %v", assetID, err)
	}
	if hgoJSON == nil {
		return fmt.Errorf("hGO %s does not exist in public state", assetID)
	}
	var hgo assets.GreenHydrogenGO
	if err := json.Unmarshal(hgoJSON, &hgo); err != nil {
		return fmt.Errorf("error unmarshalling hGO %s: %v", assetID, err)
	}
	hgo.Status = assets.GOStatusCancelled
	updatedBytes, err := json.Marshal(hgo)
	if err != nil {
		return fmt.Errorf("error marshalling tombstoned hGO %s: %v", assetID, err)
	}
	if err := ctx.GetStub().PutState(assetID, updatedBytes); err != nil {
		return fmt.Errorf("error writing tombstoned hGO %s: %v", assetID, err)
	}
	return nil
}

// MarkEGOTransferred marks an electricity GO as transferred (tombstone for transfers).
// ADR-007: The original record is retained with Status="transferred" for auditability.
func MarkEGOTransferred(ctx contractapi.TransactionContextInterface, assetID string) error {
	egoJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("error reading eGO %s for transfer tombstone: %v", assetID, err)
	}
	if egoJSON == nil {
		return fmt.Errorf("eGO %s does not exist in public state", assetID)
	}
	var ego assets.ElectricityGO
	if err := json.Unmarshal(egoJSON, &ego); err != nil {
		return fmt.Errorf("error unmarshalling eGO %s: %v", assetID, err)
	}
	ego.Status = assets.GOStatusTransferred
	updatedBytes, err := json.Marshal(ego)
	if err != nil {
		return fmt.Errorf("error marshalling transferred eGO %s: %v", assetID, err)
	}
	return ctx.GetStub().PutState(assetID, updatedBytes)
}

// MarkHGOTransferred marks a hydrogen GO as transferred (tombstone for transfers).
func MarkHGOTransferred(ctx contractapi.TransactionContextInterface, assetID string) error {
	hgoJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("error reading hGO %s for transfer tombstone: %v", assetID, err)
	}
	if hgoJSON == nil {
		return fmt.Errorf("hGO %s does not exist in public state", assetID)
	}
	var hgo assets.GreenHydrogenGO
	if err := json.Unmarshal(hgoJSON, &hgo); err != nil {
		return fmt.Errorf("error unmarshalling hGO %s: %v", assetID, err)
	}
	hgo.Status = assets.GOStatusTransferred
	updatedBytes, err := json.Marshal(hgo)
	if err != nil {
		return fmt.Errorf("error marshalling transferred hGO %s: %v", assetID, err)
	}
	return ctx.GetStub().PutState(assetID, updatedBytes)
}
