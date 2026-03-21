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

	// Get a new ID for the remainder
	nextID, err := assets.GetNextID(ctx, assets.CounterKeyEGO)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting new eGO ID for remainder: %v", err)
	}
	remainderID := fmt.Sprintf("eGO%d", nextID)

	remainderPublic = &assets.ElectricityGO{
		AssetID:          remainderID,
		CreationDateTime: original.CreationDateTime, // Bug fix #8: preserve original timestamp
		GOType:           "Electricity",
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

	nextID, err := assets.GetNextID(ctx, assets.CounterKeyHGO)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting new hGO ID for remainder: %v", err)
	}
	remainderID := fmt.Sprintf("hGO%d", nextID)

	remainderPublic = &assets.GreenHydrogenGO{
		AssetID:          remainderID,
		CreationDateTime: original.CreationDateTime,
		GOType:           "Hydrogen",
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

// DeleteEGOFromLedger removes both public and private parts of an electricity GO.
func DeleteEGOFromLedger(ctx contractapi.TransactionContextInterface, assetID, collection string) error {
	err := ctx.GetStub().DelState(assetID)
	if err != nil {
		return fmt.Errorf("error deleting eGO %s from public state: %v", assetID, err)
	}
	err = ctx.GetStub().DelPrivateData(collection, assetID)
	if err != nil {
		return fmt.Errorf("error deleting eGO %s from private collection: %v", assetID, err)
	}
	return nil
}

// DeleteHGOFromLedger removes both public and private parts of a hydrogen GO.
func DeleteHGOFromLedger(ctx contractapi.TransactionContextInterface, assetID, collection string) error {
	err := ctx.GetStub().DelState(assetID)
	if err != nil {
		return fmt.Errorf("error deleting hGO %s from public state: %v", assetID, err)
	}
	err = ctx.GetStub().DelPrivateData(collection, assetID)
	if err != nil {
		return fmt.Errorf("error deleting hGO %s from private collection: %v", assetID, err)
	}
	return nil
}
