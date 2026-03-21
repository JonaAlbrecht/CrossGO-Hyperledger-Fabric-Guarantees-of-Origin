package contracts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TransferContract groups all GO transfer functions.
type TransferContract struct {
	contractapi.Contract
}

// GO expiry period in seconds (1 hour). Can be changed to 900 for 15-minute time-granular GOs.
const ExpiryPeriod int64 = 3600

// SafetyMargin is the minimum time before expiry for a GO to still be eligible for transfer.
const SafetyMargin int64 = 300

// TransferEGO transfers a single electricity GO to another organization.
// Transient key: "TransferInput" containing EGOID, Recipient.
func (c *TransferContract) TransferEGO(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleConsumer); err != nil {
		return fmt.Errorf("only producers and consumers can transfer eGOs: %v", err)
	}

	type transferInput struct {
		EGOID     string `json:"EGOID"`
		Recipient string `json:"Recipient"`
	}

	var input transferInput
	if err := util.UnmarshalTransient(ctx, "TransferInput", &input); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("EGOID", input.EGOID); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("Recipient", input.Recipient); err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	senderCollection := access.GetCollectionForOrg(clientMSP)
	receiverCollection := access.GetCollectionForOrg(input.Recipient)

	// Read the eGO private details
	eGOJSON, err := ctx.GetStub().GetPrivateData(senderCollection, input.EGOID)
	if err != nil {
		return fmt.Errorf("failed to read eGO %s: %v", input.EGOID, err)
	}
	if eGOJSON == nil {
		return fmt.Errorf("eGO %s does not exist in your collection", input.EGOID)
	}

	var eGOPrivate assets.ElectricityGOPrivateDetails
	if err := json.Unmarshal(eGOJSON, &eGOPrivate); err != nil {
		return fmt.Errorf("failed to unmarshal eGO: %v", err)
	}

	// Expiry check
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}
	if now-ExpiryPeriod+SafetyMargin > eGOPrivate.CreationDateTime {
		return fmt.Errorf("eGO %s is no longer eligible for transfer (expired or near expiry)", input.EGOID)
	}

	// Transfer consumption declarations
	if err := util.TransferConsumptionDeclarations(ctx, eGOPrivate.ConsumptionDeclarations, senderCollection, receiverCollection, true); err != nil {
		return err
	}

	// Update owner and write to receiver
	eGOPrivate.OwnerID = input.Recipient
	updatedBytes, err := json.Marshal(eGOPrivate)
	if err != nil {
		return fmt.Errorf("failed to marshal updated eGO: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(receiverCollection, input.EGOID, updatedBytes); err != nil {
		return fmt.Errorf("failed to write eGO to receiver: %v", err)
	}

	// Delete from sender
	if err := ctx.GetStub().DelPrivateData(senderCollection, input.EGOID); err != nil {
		return fmt.Errorf("failed to delete eGO from sender: %v", err)
	}

	return nil
}

// TransferEGOByAmount transfers a specified MWh amount across one or more eGOs.
// GOs are consumed fully until the amount is met; the last GO may be split.
// Transient key: "TransferInput" containing EGOList ("+"-separated), Recipient, Neededamount.
func (c *TransferContract) TransferEGOByAmount(ctx contractapi.TransactionContextInterface) ([]string, error) {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleConsumer); err != nil {
		return nil, fmt.Errorf("only producers and consumers can transfer eGOs: %v", err)
	}

	type transferInput struct {
		EGOList      string      `json:"EGOList"`
		Recipient    string      `json:"Recipient"`
		Neededamount json.Number `json:"Neededamount"`
	}

	var input transferInput
	if err := util.UnmarshalTransient(ctx, "TransferInput", &input); err != nil {
		return nil, err
	}

	neededAmount, err := input.Neededamount.Float64()
	if err != nil {
		return nil, fmt.Errorf("error converting Neededamount: %v", err)
	}
	if err := util.ValidatePositive(map[string]float64{"Neededamount": neededAmount}); err != nil {
		return nil, err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}
	senderCollection := access.GetCollectionForOrg(clientMSP)
	receiverCollection := access.GetCollectionForOrg(input.Recipient)
	eGOList := strings.Split(input.EGOList, "+")

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	timecheck := now - ExpiryPeriod + SafetyMargin

	var transferredMWh float64
	var results []string

	for i := 0; transferredMWh < neededAmount; i++ {
		// Bug fix #6: bounds check
		if i >= len(eGOList) {
			return nil, fmt.Errorf("insufficient eGOs: transferred %.4f MWh of %.4f needed", transferredMWh, neededAmount)
		}

		currentID := eGOList[i]
		currentAssetJSON, err := ctx.GetStub().GetPrivateData(senderCollection, currentID)
		if err != nil {
			return nil, fmt.Errorf("error reading eGO %s: %v", currentID, err)
		}
		if currentAssetJSON == nil {
			return nil, fmt.Errorf("eGO %s does not exist in your collection", currentID)
		}

		var currentAsset assets.ElectricityGOPrivateDetails
		if err := json.Unmarshal(currentAssetJSON, &currentAsset); err != nil {
			return nil, fmt.Errorf("error unmarshaling eGO %s: %v", currentID, err)
		}

		// Expiry check
		if timecheck > currentAsset.CreationDateTime {
			return nil, fmt.Errorf("eGO %s is no longer eligible for transfer (expired or near expiry)", currentAsset.AssetID)
		}

		transferredMWh += currentAsset.AmountMWh

		if transferredMWh <= neededAmount {
			// Transfer entire GO
			if err := util.TransferConsumptionDeclarations(ctx, currentAsset.ConsumptionDeclarations, senderCollection, receiverCollection, true); err != nil {
				return nil, err
			}
			currentAsset.OwnerID = input.Recipient
			updatedBytes, err := json.Marshal(currentAsset)
			if err != nil {
				return nil, fmt.Errorf("error marshaling eGO: %v", err)
			}
			if err := ctx.GetStub().PutPrivateData(receiverCollection, currentAsset.AssetID, updatedBytes); err != nil {
				return nil, fmt.Errorf("error writing eGO to receiver: %v", err)
			}
			if err := ctx.GetStub().DelPrivateData(senderCollection, currentAsset.AssetID); err != nil {
				return nil, fmt.Errorf("error deleting eGO from sender: %v", err)
			}
			results = append(results, fmt.Sprintf("Transferred eGO %s (%.4f MWh) fully", currentAsset.AssetID, currentAsset.AmountMWh))
		} else {
			// Split: transfer the needed portion, remainder stays with sender
			takenAmount := currentAsset.AmountMWh - (transferredMWh - neededAmount)
			taken, remainderPriv, remainderPub, err := util.SplitElectricityGO(ctx, &currentAsset, takenAmount, input.Recipient)
			if err != nil {
				return nil, fmt.Errorf("error splitting eGO %s: %v", currentID, err)
			}

			// Copy consumption declarations to receiver (don't delete — sender's remainder needs them too)
			if err := util.TransferConsumptionDeclarations(ctx, currentAsset.ConsumptionDeclarations, senderCollection, receiverCollection, false); err != nil {
				return nil, err
			}

			// Write taken portion to receiver
			takenBytes, err := json.Marshal(taken)
			if err != nil {
				return nil, fmt.Errorf("error marshaling taken portion: %v", err)
			}
			if err := ctx.GetStub().PutPrivateData(receiverCollection, taken.AssetID, takenBytes); err != nil {
				return nil, fmt.Errorf("error writing taken portion to receiver: %v", err)
			}

			// Delete original from sender
			if err := ctx.GetStub().DelPrivateData(senderCollection, currentAsset.AssetID); err != nil {
				return nil, fmt.Errorf("error deleting original eGO from sender: %v", err)
			}

			// Write remainder (new ID) to sender
			if err := util.WriteEGOToLedger(ctx, remainderPub, remainderPriv, senderCollection); err != nil {
				return nil, fmt.Errorf("error writing remainder eGO: %v", err)
			}

			results = append(results, fmt.Sprintf("Split eGO %s: %.4f MWh transferred, %.4f MWh remainder as %s",
				currentAsset.AssetID, takenAmount, remainderPriv.AmountMWh, remainderPub.AssetID))
		}
	}
	return results, nil
}

// TransferHGOByAmount transfers a specified kilogram amount across one or more hydrogen GOs.
// Transient key: "TransferInput" containing HGOList ("+"-separated), Recipient, NeededKilos.
func (c *TransferContract) TransferHGOByAmount(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleConsumer); err != nil {
		return fmt.Errorf("only producers and consumers can transfer hGOs: %v", err)
	}

	type transferInput struct {
		HGOList     string      `json:"HGOList"`
		Recipient   string      `json:"Recipient"`
		NeededKilos json.Number `json:"NeededKilos"`
	}

	var input transferInput
	if err := util.UnmarshalTransient(ctx, "TransferInput", &input); err != nil {
		return err
	}

	neededKilos, err := input.NeededKilos.Float64()
	if err != nil {
		return fmt.Errorf("error converting NeededKilos: %v", err)
	}
	if err := util.ValidatePositive(map[string]float64{"NeededKilos": neededKilos}); err != nil {
		return err
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	senderCollection := access.GetCollectionForOrg(clientMSP)
	receiverCollection := access.GetCollectionForOrg(input.Recipient)
	hGOList := strings.Split(input.HGOList, "+")

	var transferredKilos float64

	for i := 0; transferredKilos < neededKilos; i++ {
		if i >= len(hGOList) {
			return fmt.Errorf("insufficient hGOs: transferred %.4f kg of %.4f needed", transferredKilos, neededKilos)
		}

		currentID := hGOList[i]
		currentAssetJSON, err := ctx.GetStub().GetPrivateData(senderCollection, currentID)
		if err != nil {
			return fmt.Errorf("error reading hGO %s: %v", currentID, err)
		}
		if currentAssetJSON == nil {
			return fmt.Errorf("hGO %s does not exist in your collection", currentID)
		}

		var currentAsset assets.GreenHydrogenGOPrivateDetails
		if err := json.Unmarshal(currentAssetJSON, &currentAsset); err != nil {
			return fmt.Errorf("error unmarshaling hGO %s: %v", currentID, err)
		}

		transferredKilos += currentAsset.Kilosproduced

		if transferredKilos <= neededKilos {
			// Transfer entire hGO
			if err := util.TransferConsumptionDeclarations(ctx, currentAsset.ConsumptionDeclarations, senderCollection, receiverCollection, true); err != nil {
				return err
			}
			currentAsset.OwnerID = input.Recipient
			updatedBytes, err := json.Marshal(currentAsset)
			if err != nil {
				return fmt.Errorf("error marshaling hGO: %v", err)
			}
			if err := ctx.GetStub().PutPrivateData(receiverCollection, currentAsset.AssetID, updatedBytes); err != nil {
				return fmt.Errorf("error writing hGO to receiver: %v", err)
			}
			if err := ctx.GetStub().DelPrivateData(senderCollection, currentAsset.AssetID); err != nil {
				return fmt.Errorf("error deleting hGO from sender: %v", err)
			}
		} else {
			// Split
			takenKilos := currentAsset.Kilosproduced - (transferredKilos - neededKilos)
			taken, remainderPriv, remainderPub, err := util.SplitHydrogenGO(ctx, &currentAsset, takenKilos, input.Recipient)
			if err != nil {
				return fmt.Errorf("error splitting hGO %s: %v", currentID, err)
			}

			if err := util.TransferConsumptionDeclarations(ctx, currentAsset.ConsumptionDeclarations, senderCollection, receiverCollection, false); err != nil {
				return err
			}

			takenBytes, err := json.Marshal(taken)
			if err != nil {
				return fmt.Errorf("error marshaling taken hGO portion: %v", err)
			}
			if err := ctx.GetStub().PutPrivateData(receiverCollection, taken.AssetID, takenBytes); err != nil {
				return fmt.Errorf("error writing taken hGO portion to receiver: %v", err)
			}
			if err := ctx.GetStub().DelPrivateData(senderCollection, currentAsset.AssetID); err != nil {
				return fmt.Errorf("error deleting original hGO from sender: %v", err)
			}
			if err := util.WriteHGOToLedger(ctx, remainderPub, remainderPriv, senderCollection); err != nil {
				return fmt.Errorf("error writing remainder hGO: %v", err)
			}
		}
	}
	return nil
}
