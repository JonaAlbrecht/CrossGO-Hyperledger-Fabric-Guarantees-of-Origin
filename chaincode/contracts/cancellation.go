package contracts

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// CancellationContract groups GO cancellation (claim renewable attributes) and verification functions.
type CancellationContract struct {
	contractapi.Contract
}

// ClaimRenewableAttributesElectricity cancels electricity GOs to claim their renewable attributes.
// GOs are cancelled fully until the target amount is met; the last GO may be split.
// Transient key: "ClaimRenewables" containing EGOList ("+"-separated), Collection, Cancelamount.
//
// Bug fix #8: Remainder GO preserves original CreationDateTime.
func (c *CancellationContract) ClaimRenewableAttributesElectricity(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleConsumer); err != nil {
		return fmt.Errorf("only producers and consumers can cancel eGOs: %v", err)
	}

	type claimInput struct {
		EGOList      string      `json:"EGOList"`
		Collection   string      `json:"Collection"`
		Cancelamount json.Number `json:"Cancelamount"`
	}

	var input claimInput
	if err := util.UnmarshalTransient(ctx, "ClaimRenewables", &input); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("Collection", input.Collection); err != nil {
		return err
	}

	cancelAmount, err := input.Cancelamount.Float64()
	if err != nil {
		return fmt.Errorf("error converting Cancelamount: %v", err)
	}
	if err := util.ValidatePositive(map[string]float64{"Cancelamount": cancelAmount}); err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	eGOList := strings.Split(input.EGOList, "+")
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}
	timecheck := now - ExpiryPeriod

	var claimedAmount float64
	suffixCounter := 0

	for i := 0; claimedAmount < cancelAmount; i++ {
		if i >= len(eGOList) {
			return fmt.Errorf("insufficient eGOs: claimed %.4f of %.4f MWh", claimedAmount, cancelAmount)
		}

		eGOJSON, err := ctx.GetStub().GetPrivateData(input.Collection, eGOList[i])
		if err != nil {
			return fmt.Errorf("failed to read eGO %s: %v", eGOList[i], err)
		}
		if eGOJSON == nil {
			return fmt.Errorf("eGO %s does not exist in collection %s", eGOList[i], input.Collection)
		}

		var eGOPrivate assets.ElectricityGOPrivateDetails
		if err := json.Unmarshal(eGOJSON, &eGOPrivate); err != nil {
			return fmt.Errorf("failed to unmarshal eGO: %v", err)
		}

		if timecheck > eGOPrivate.CreationDateTime {
			return fmt.Errorf("eGO %s is expired", eGOPrivate.AssetID)
		}

		// ADR-001: transaction-ID-derived deterministic ID
		eCancelKey, err := assets.GenerateID(ctx, assets.PrefixECancellation, suffixCounter)
		if err != nil {
			return fmt.Errorf("error generating cancellation key: %v", err)
		}
		suffixCounter++

		claimedAmount += eGOPrivate.AmountMWh

		if claimedAmount <= cancelAmount {
			// Cancel entire eGO
			statement := assets.CancellationStatementElectricity{
				ECancellationkey:            eCancelKey,
				CancellationTime:            now,
				OwnerID:                     eGOPrivate.OwnerID,
				AmountMWh:                   eGOPrivate.AmountMWh,
				Emissions:                   eGOPrivate.Emissions,
				ElectricityProductionMethod: eGOPrivate.ElectricityProductionMethod,
				ConsumptionDeclarations:     eGOPrivate.ConsumptionDeclarations,
			}
			stmtBytes, err := json.Marshal(statement)
			if err != nil {
				return fmt.Errorf("failed to marshal cancellation statement: %v", err)
			}

			if err := util.DeleteEGOFromLedger(ctx, eGOList[i], input.Collection); err != nil {
				return err
			}
			if err := ctx.GetStub().PutPrivateData(input.Collection, eCancelKey, stmtBytes); err != nil {
				return fmt.Errorf("failed to write cancellation statement: %v", err)
			}
		} else {
			// Split: cancel needed portion, remainder stays as new eGO
			excessAmount := claimedAmount - cancelAmount
			ratio := excessAmount / eGOPrivate.AmountMWh
			cancelledAmount := eGOPrivate.AmountMWh - excessAmount
			cancelledEmissions := (1 - ratio) * eGOPrivate.Emissions

			declarations := make([]string, len(eGOPrivate.ConsumptionDeclarations))
			copy(declarations, eGOPrivate.ConsumptionDeclarations)
			declarations = append(declarations, "split")

			statement := assets.CancellationStatementElectricity{
				ECancellationkey:            eCancelKey,
				CancellationTime:            now,
				OwnerID:                     eGOPrivate.OwnerID,
				AmountMWh:                   cancelledAmount,
				Emissions:                   cancelledEmissions,
				ElectricityProductionMethod: eGOPrivate.ElectricityProductionMethod,
				ConsumptionDeclarations:     declarations,
			}
			stmtBytes, err := json.Marshal(statement)
			if err != nil {
				return fmt.Errorf("failed to marshal cancellation statement: %v", err)
			}

			// ADR-001: transaction-ID-derived deterministic ID for remainder
			remainderID, err := assets.GenerateID(ctx, assets.PrefixEGO, suffixCounter)
			if err != nil {
				return fmt.Errorf("error generating remainder eGO ID: %v", err)
			}
			suffixCounter++

			// Bug fix #8: preserve original creation timestamp on remainder
			remainderPub := &assets.ElectricityGO{
				AssetID:          remainderID,
				CreationDateTime: eGOPrivate.CreationDateTime,
				GOType:           "Electricity",
			}
			remainderPriv := &assets.ElectricityGOPrivateDetails{
				AssetID:                     remainderID,
				OwnerID:                     clientMSP,
				CreationDateTime:            eGOPrivate.CreationDateTime,
				AmountMWh:                   excessAmount,
				Emissions:                   ratio * eGOPrivate.Emissions,
				ElectricityProductionMethod: eGOPrivate.ElectricityProductionMethod,
				ConsumptionDeclarations:     declarations,
				DeviceID:                    eGOPrivate.DeviceID,
			}

			// Delete original, write cancellation statement, write remainder
			if err := util.DeleteEGOFromLedger(ctx, eGOList[i], input.Collection); err != nil {
				return err
			}
			if err := ctx.GetStub().PutPrivateData(input.Collection, eCancelKey, stmtBytes); err != nil {
				return fmt.Errorf("failed to write cancellation statement: %v", err)
			}
			if err := util.WriteEGOToLedger(ctx, remainderPub, remainderPriv, input.Collection); err != nil {
				return fmt.Errorf("error writing remainder eGO: %v", err)
			}
		}
	}
	return nil
}

// ClaimRenewableAttributesHydrogen cancels hydrogen GOs to claim their renewable attributes.
// Transient key: "ClaimHydrogen" containing HGOList ("+"-separated), collection, Cancelamount.
//
// Bug fix #9: ConsumptionDeclarations deep-copied on split.
// Bug fix #10: ConsumptionDeclarationHydrogen uses int64 for DateTime.
func (c *CancellationContract) ClaimRenewableAttributesHydrogen(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleConsumer); err != nil {
		return fmt.Errorf("only producers and consumers can cancel hGOs: %v", err)
	}

	type claimInput struct {
		HGOList      string      `json:"HGOList"`
		Collection   string      `json:"collection"`
		Cancelamount json.Number `json:"Cancelamount"`
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	var input claimInput
	if err := util.UnmarshalTransient(ctx, "ClaimHydrogen", &input); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("collection", input.Collection); err != nil {
		return err
	}

	cancelAmount, err := input.Cancelamount.Float64()
	if err != nil {
		return fmt.Errorf("error converting Cancelamount: %v", err)
	}
	if err := util.ValidatePositive(map[string]float64{"Cancelamount": cancelAmount}); err != nil {
		return err
	}

	hGOList := strings.Split(input.HGOList, "+")
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	var claimedKilos float64
	suffixCounter := 0

	for i := 0; claimedKilos < cancelAmount; i++ {
		if i >= len(hGOList) {
			return fmt.Errorf("insufficient hGOs: claimed %.4f of %.4f kg", claimedKilos, cancelAmount)
		}

		hGOJSON, err := ctx.GetStub().GetPrivateData(input.Collection, hGOList[i])
		if err != nil {
			return fmt.Errorf("failed to read hGO %s: %v", hGOList[i], err)
		}
		if hGOJSON == nil {
			return fmt.Errorf("hGO %s does not exist in collection %s", hGOList[i], input.Collection)
		}

		var hGOPrivate assets.GreenHydrogenGOPrivateDetails
		if err := json.Unmarshal(hGOJSON, &hGOPrivate); err != nil {
			return fmt.Errorf("failed to unmarshal hGO: %v", err)
		}

		// ADR-001: transaction-ID-derived deterministic ID
		hCancelKey, err := assets.GenerateID(ctx, assets.PrefixHCancellation, suffixCounter)
		if err != nil {
			return fmt.Errorf("error generating cancellation key: %v", err)
		}
		suffixCounter++

		claimedKilos += hGOPrivate.Kilosproduced

		if claimedKilos <= cancelAmount {
			// Cancel entire hGO
			statement := assets.CancellationStatementHydrogen{
				HCancellationkey:            hCancelKey,
				CancellationTime:            now,
				OwnerID:                     hGOPrivate.OwnerID,
				Kilosproduced:               hGOPrivate.Kilosproduced,
				EmissionsHydrogen:           hGOPrivate.EmissionsHydrogen,
				HydrogenProductionMethod:    hGOPrivate.HydrogenProductionMethod,
				InputEmissions:              hGOPrivate.InputEmissions,
				ElectricityProductionMethod: hGOPrivate.ElectricityProductionMethod,
				UsedMWh:                     hGOPrivate.UsedMWh,
				ConsumptionDeclarations:     hGOPrivate.ConsumptionDeclarations,
			}
			stmtBytes, err := json.Marshal(statement)
			if err != nil {
				return fmt.Errorf("failed to marshal hGO cancellation: %v", err)
			}

			if err := util.DeleteHGOFromLedger(ctx, hGOList[i], input.Collection); err != nil {
				return err
			}
			if err := ctx.GetStub().PutPrivateData(input.Collection, hCancelKey, stmtBytes); err != nil {
				return fmt.Errorf("failed to write hGO cancellation: %v", err)
			}

			// ADR-001: transaction-ID-derived deterministic ID
			hConsumptionKey, err := assets.GenerateID(ctx, assets.PrefixHConsumption, suffixCounter)
			if err != nil {
				return fmt.Errorf("error generating consumption ID: %v", err)
			}
			suffixCounter++
			consumption := assets.ConsumptionDeclarationHydrogen{
				Consumptionkey:           hConsumptionKey,
				CancelledGOID:            hGOPrivate.AssetID,
				ConsumptionDateTime:      now,
				Kilosproduced:            hGOPrivate.Kilosproduced,
				EmissionsHydrogen:        hGOPrivate.EmissionsHydrogen,
				HydrogenProductionMethod: hGOPrivate.HydrogenProductionMethod,
				ConsumptionDeclarations:  hGOPrivate.ConsumptionDeclarations,
			}
			consBytes, err := json.Marshal(consumption)
			if err != nil {
				return fmt.Errorf("failed to marshal hGO consumption declaration: %v", err)
			}
			if err := ctx.GetStub().PutPrivateData(input.Collection, hConsumptionKey, consBytes); err != nil {
				return fmt.Errorf("failed to write hGO consumption declaration: %v", err)
			}
		} else {
			// Split: cancel needed portion, remainder stays
			excessKilos := claimedKilos - cancelAmount
			ratio := excessKilos / hGOPrivate.Kilosproduced
			cancelledKilos := hGOPrivate.Kilosproduced - excessKilos

			// Bug fix #9: deep-copy declarations
			cancelDeclarations := make([]string, len(hGOPrivate.ConsumptionDeclarations))
			copy(cancelDeclarations, hGOPrivate.ConsumptionDeclarations)
			cancelDeclarations = append(cancelDeclarations, "split")

			statement := assets.CancellationStatementHydrogen{
				HCancellationkey:            hCancelKey,
				CancellationTime:            now,
				OwnerID:                     hGOPrivate.OwnerID,
				Kilosproduced:               cancelledKilos,
				EmissionsHydrogen:           (1 - ratio) * hGOPrivate.EmissionsHydrogen,
				HydrogenProductionMethod:    hGOPrivate.HydrogenProductionMethod,
				InputEmissions:              (1 - ratio) * hGOPrivate.InputEmissions,
				ElectricityProductionMethod: hGOPrivate.ElectricityProductionMethod,
				UsedMWh:                     (1 - ratio) * hGOPrivate.UsedMWh,
				ConsumptionDeclarations:     cancelDeclarations,
			}
			stmtBytes, err := json.Marshal(statement)
			if err != nil {
				return fmt.Errorf("failed to marshal hGO cancellation: %v", err)
			}

			// ADR-001: transaction-ID-derived deterministic ID for remainder
			remainderID, err := assets.GenerateID(ctx, assets.PrefixHGO, suffixCounter)
			if err != nil {
				return fmt.Errorf("error generating remainder hGO ID: %v", err)
			}
			suffixCounter++

			// Bug fix #9: deep-copy for remainder too
			remainderDeclarations := make([]string, len(hGOPrivate.ConsumptionDeclarations))
			copy(remainderDeclarations, hGOPrivate.ConsumptionDeclarations)
			remainderDeclarations = append(remainderDeclarations, "split")

			remainderMethods := make([]string, len(hGOPrivate.ElectricityProductionMethod))
			copy(remainderMethods, hGOPrivate.ElectricityProductionMethod)

			remainderPub := &assets.GreenHydrogenGO{
				AssetID:          remainderID,
				CreationDateTime: hGOPrivate.CreationDateTime,
				GOType:           "Hydrogen",
			}
			remainderPriv := &assets.GreenHydrogenGOPrivateDetails{
				AssetID:                     remainderID,
				OwnerID:                     clientMSP,
				CreationDateTime:            hGOPrivate.CreationDateTime,
				Kilosproduced:               excessKilos,
				EmissionsHydrogen:           ratio * hGOPrivate.EmissionsHydrogen,
				HydrogenProductionMethod:    hGOPrivate.HydrogenProductionMethod,
				InputEmissions:              ratio * hGOPrivate.InputEmissions,
				UsedMWh:                     ratio * hGOPrivate.UsedMWh,
				ElectricityProductionMethod: remainderMethods,
				ConsumptionDeclarations:     remainderDeclarations,
				DeviceID:                    hGOPrivate.DeviceID,
			}

			if err := util.DeleteHGOFromLedger(ctx, hGOList[i], input.Collection); err != nil {
				return err
			}
			if err := ctx.GetStub().PutPrivateData(input.Collection, hCancelKey, stmtBytes); err != nil {
				return fmt.Errorf("failed to write hGO cancellation: %v", err)
			}
			if err := util.WriteHGOToLedger(ctx, remainderPub, remainderPriv, input.Collection); err != nil {
				return fmt.Errorf("error writing remainder hGO: %v", err)
			}
		}
	}
	return nil
}

// VerifyCancellationStatement verifies a cancellation statement's hash against on-chain data.
// Bug fix #11: uses correct collection and key format for hash comparison.
func (c *CancellationContract) VerifyCancellationStatement(ctx contractapi.TransactionContextInterface, assetID string, sellerCollection string) (bool, error) {
	transMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return false, fmt.Errorf("error getting transient: %v", err)
	}
	immutablePropertiesJSON, ok := transMap["CancelStatement"]
	if !ok {
		return false, fmt.Errorf("CancelStatement key not found in transient map")
	}

	// Bug fix #11: use the correct assetID as the key for the private data hash
	onChainHash, err := ctx.GetStub().GetPrivateDataHash(sellerCollection, assetID)
	if err != nil {
		return false, fmt.Errorf("failed to read private data hash from collection %s: %v", sellerCollection, err)
	}
	if onChainHash == nil {
		return false, fmt.Errorf("no private data hash found for asset %s in collection %s", assetID, sellerCollection)
	}

	hash := sha256.New()
	hash.Write(immutablePropertiesJSON)
	calculatedHash := hash.Sum(nil)

	if !bytes.Equal(onChainHash, calculatedHash) {
		return false, fmt.Errorf("hash mismatch: calculated %x does not match on-chain %x", calculatedHash, onChainHash)
	}

	// Note: the original check comparing hash to assetID is removed — asset IDs are sequential
	// counters (eGO1, eCancel1, etc.), not content hashes, so the comparison never matched.
	// This was bug #11.

	return true, nil
}

// SetGOEndorsementPolicy sets a key-level endorsement policy for an asset.
func SetGOEndorsementPolicy(ctx contractapi.TransactionContextInterface, assetID string, orgsToEndorse []string) error {
	endorsementPolicy, err := statebased.NewStateEP(nil)
	if err != nil {
		return err
	}
	err = endorsementPolicy.AddOrgs(statebased.RoleTypePeer, orgsToEndorse...)
	if err != nil {
		return fmt.Errorf("failed to add org to endorsement policy: %v", err)
	}
	policy, err := endorsementPolicy.Policy()
	if err != nil {
		return fmt.Errorf("failed to create endorsement policy bytes: %v", err)
	}
	err = ctx.GetStub().SetStateValidationParameter(assetID, policy)
	if err != nil {
		return fmt.Errorf("failed to set validation parameter: %v", err)
	}
	return nil
}
