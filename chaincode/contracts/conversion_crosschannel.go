package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ================================================================================
// v10.1: Cross-Channel Conversion Contract (ADR-033)
// ================================================================================
// This contract enables carrier-to-carrier conversions across separate Fabric channels.
// Example: Convert electricity GO on electricity-de → hydrogen GO on hydrogen-de.
//
// Protocol:
//   Phase 1 (source channel): LockGOForConversion — Lock source GO with tri-party endorsement
//   Phase 2 (dest channel):   MintFromConversion   — Mint destination GO from lock receipt
//   Phase 3 (source channel): FinalizeLock         — Finalize conversion, mark source CONSUMED
//
// Tri-party endorsement: Owner + source issuer + destination issuer
// Backlog integration: Destination channel reads its own backlog (no cross-channel queries)

// ConversionCrossChannelContract handles all cross-channel conversions.
type ConversionCrossChannelContract struct {
	contractapi.Contract
}

// ================================================================================
// Phase 1: Lock Source GO for Conversion (Source Channel)
// ================================================================================

// LockGOForConversion locks a GO on the source channel for cross-channel conversion.
// This is Phase 1 of the lock-mint-finalize protocol.
//
// Transient input key: "LockForConversion"
// Required fields:
//   - GOAssetID: Source GO to lock (e.g., "eGO_123")
//   - DestinationChannel: Target channel (e.g., "hydrogen-de")
//   - DestinationCarrier: Target carrier type (e.g., "hydrogen")
//   - ConversionMethod: Conversion technology (e.g., "electrolysis", "fuel_cell")
//   - ConversionEfficiency: Efficiency factor (e.g., 0.65 = 65%)
//   - OwnerMSP: GO owner MSP (for tri-party endorsement)
//
// Endorsement: Source issuer + GO owner (tri-party)
// Returns: Conversion lock record with lock receipt hash
func (c *ConversionCrossChannelContract) LockGOForConversion(ctx contractapi.TransactionContextInterface) error {
	// Only producers can initiate conversions
	if err := access.RequireRole(ctx, access.RoleProducer); err != nil {
		return fmt.Errorf("only producers can lock GOs for conversion: %v", err)
	}

	type lockInput struct {
		GOAssetID            string  `json:"GOAssetID"`
		DestinationChannel   string  `json:"DestinationChannel"`
		DestinationCarrier   string  `json:"DestinationCarrier"`
		ConversionMethod     string  `json:"ConversionMethod"`
		ConversionEfficiency float64 `json:"ConversionEfficiency"`
		OwnerMSP             string  `json:"OwnerMSP"`
	}

	var input lockInput
	if err := util.UnmarshalTransient(ctx, "LockForConversion", &input); err != nil {
		return err
	}

	// Validate inputs
	if err := util.ValidateNonEmpty("GOAssetID", input.GOAssetID); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("DestinationChannel", input.DestinationChannel); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("DestinationCarrier", input.DestinationCarrier); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("ConversionMethod", input.ConversionMethod); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("OwnerMSP", input.OwnerMSP); err != nil {
		return err
	}
	if input.ConversionEfficiency <= 0 || input.ConversionEfficiency > 10 {
		return fmt.Errorf("invalid conversion efficiency: %f (must be 0 < eff <= 10)", input.ConversionEfficiency)
	}

	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	// Verify client is the owner or issuer
	if clientMSP != input.OwnerMSP {
		issuerRole, err := access.GetRole(ctx, clientMSP)
		if err != nil || issuerRole != access.RoleIssuer {
			return fmt.Errorf("only the owner (%s) or issuer can lock GOs for conversion", input.OwnerMSP)
		}
	}

	// Verify GO ownership
	ownerCollection := access.GetCollectionForOrg(input.OwnerMSP)
	if err := verifyGOOwnership(ctx, input.GOAssetID, input.OwnerMSP, ownerCollection); err != nil {
		return fmt.Errorf("ownership verification failed: %v", err)
	}

	// Read source GO private data to populate lock receipt
	sourceGOJSON, err := ctx.GetStub().GetPrivateData(ownerCollection, input.GOAssetID)
	if err != nil {
		return fmt.Errorf("error reading source GO: %v", err)
	}
	if sourceGOJSON == nil {
		return fmt.Errorf("source GO %s does not exist", input.GOAssetID)
	}

	// Determine source carrier type from GO ID prefix
	sourceCarrier, err := determineCarrierFromGOID(input.GOAssetID)
	if err != nil {
		return err
	}

	// Read source channel name from chaincode environment (channel where this is executing)
	sourceChannel := ctx.GetStub().GetChannelID()

	// Parse source GO data based on carrier type
	var sourceAmount float64
	var sourceAmountUnit string
	var sourceEmissions float64
	var sourceProductionMethod string
	var sourceDeviceID string
	var sourceCreationDateTime int64
	var sourceConsumptionDecls []string

	switch sourceCarrier {
	case "electricity":
		var eGO assets.ElectricityGOPrivateDetails
		if err := json.Unmarshal(sourceGOJSON, &eGO); err != nil {
			return fmt.Errorf("error unmarshaling electricity GO: %v", err)
		}
		sourceAmount = eGO.AmountMWh
		sourceAmountUnit = "MWh"
		sourceEmissions = eGO.Emissions
		sourceProductionMethod = eGO.ElectricityProductionMethod
		sourceDeviceID = eGO.DeviceID
		sourceCreationDateTime = eGO.CreationDateTime
		sourceConsumptionDecls = eGO.ConsumptionDeclarations

	case "hydrogen":
		var hGO assets.GreenHydrogenGOPrivateDetails
		if err := json.Unmarshal(sourceGOJSON, &hGO); err != nil {
			return fmt.Errorf("error unmarshaling hydrogen GO: %v", err)
		}
		sourceAmount = hGO.Kilosproduced
		sourceAmountUnit = "kg"
		sourceEmissions = hGO.EmissionsHydrogen + hGO.InputEmissions
		sourceProductionMethod = hGO.HydrogenProductionMethod
		sourceDeviceID = hGO.DeviceID
		sourceCreationDateTime = hGO.CreationDateTime
		sourceConsumptionDecls = hGO.ConsumptionDeclarations

	case "biogas":
		var bGO assets.BiogasGOPrivateDetails
		if err := json.Unmarshal(sourceGOJSON, &bGO); err != nil {
			return fmt.Errorf("error unmarshaling biogas GO: %v", err)
		}
		sourceAmount = bGO.VolumeNm3
		sourceAmountUnit = "Nm3"
		sourceEmissions = bGO.Emissions
		sourceProductionMethod = bGO.BiogasProductionMethod
		sourceDeviceID = bGO.DeviceID
		sourceCreationDateTime = bGO.CreationDateTime
		sourceConsumptionDecls = bGO.ConsumptionDeclarations

	case "heating_cooling":
		var htGO assets.HeatingCoolingGOPrivateDetails
		if err := json.Unmarshal(sourceGOJSON, &htGO); err != nil {
			return fmt.Errorf("error unmarshaling heating/cooling GO: %v", err)
		}
		sourceAmount = htGO.AmountMWh
		sourceAmountUnit = "MWh"
		sourceEmissions = htGO.Emissions
		sourceProductionMethod = htGO.HeatingCoolingProductionMethod
		sourceDeviceID = htGO.DeviceID
		sourceCreationDateTime = htGO.CreationDateTime
		sourceConsumptionDecls = htGO.ConsumptionDeclarations

	default:
		return fmt.Errorf("unsupported source carrier: %s", sourceCarrier)
	}

	// Read public GO data for metadata
	publicGOJSON, err := ctx.GetStub().GetState(input.GOAssetID)
	if err != nil {
		return fmt.Errorf("error reading public GO: %v", err)
	}
	if publicGOJSON == nil {
		return fmt.Errorf("public GO %s does not exist", input.GOAssetID)
	}

	// Parse public data to extract metadata (works for all carriers since they share common fields)
	var publicGO struct {
		CountryOfOrigin      string `json:"CountryOfOrigin"`
		ProductionPeriodStart int64  `json:"ProductionPeriodStart"`
		ProductionPeriodEnd   int64  `json:"ProductionPeriodEnd"`
		SupportScheme        string `json:"SupportScheme"`
		GridConnectionPoint  string `json:"GridConnectionPoint"`
		Status               string `json:"Status"`
	}
	if err := json.Unmarshal(publicGOJSON, &publicGO); err != nil {
		return fmt.Errorf("error unmarshaling public GO: %v", err)
	}

	// Verify GO is active
	if publicGO.Status != assets.GOStatusActive {
		return fmt.Errorf("cannot lock GO with status %s (must be active)", publicGO.Status)
	}

	// Generate lock ID
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}
	lockID, err := assets.GenerateID(ctx, assets.PrefixConversionLock, 0)
	if err != nil {
		return fmt.Errorf("error generating lock ID: %v", err)
	}

	// Get source issuer MSP
	sourceIssuerMSP, err := getChannelIssuerMSP(ctx)
	if err != nil {
		return fmt.Errorf("error getting source issuer MSP: %v", err)
	}

	// Generate lock receipt hash
	txID := ctx.GetStub().GetTxID()
	lockReceiptHash, err := generateConversionLockReceiptHash(lockID, input.GOAssetID, sourceChannel,
		input.DestinationChannel, input.DestinationCarrier, input.ConversionEfficiency, input.OwnerMSP, txID)
	if err != nil {
		return fmt.Errorf("error generating lock receipt hash: %v", err)
	}

	// Create conversion lock record
	conversionLock := &assets.ConversionLock{
		LockID:               lockID,
		GOAssetID:            input.GOAssetID,
		SourceChannel:        sourceChannel,
		SourceCarrier:        sourceCarrier,
		DestinationChannel:   input.DestinationChannel,
		DestinationCarrier:   input.DestinationCarrier,
		ConversionMethod:     input.ConversionMethod,
		ConversionEfficiency: input.ConversionEfficiency,
		OwnerMSP:             input.OwnerMSP,
		SourceIssuerMSP:      sourceIssuerMSP,
		LockReceiptHash:      lockReceiptHash,
		CreatedAt:            now,
		Status:               assets.ConversionLockStatusLocked,
	}

	lockBytes, err := json.Marshal(conversionLock)
	if err != nil {
		return fmt.Errorf("error marshaling conversion lock: %v", err)
	}

	// Write lock record to ledger
	if err := ctx.GetStub().PutState(lockID, lockBytes); err != nil {
		return fmt.Errorf("error writing conversion lock: %v", err)
	}

	// Set tri-party state-based endorsement policy (source issuer + owner)
	endorsementPolicy, err := statebased.NewStateEP(nil)
	if err != nil {
		return fmt.Errorf("error creating endorsement policy: %v", err)
	}
	if err := endorsementPolicy.AddOrgs(statebased.RoleTypePeer, sourceIssuerMSP, input.OwnerMSP); err != nil {
		return fmt.Errorf("error adding orgs to endorsement policy: %v", err)
	}
	policyBytes, err := endorsementPolicy.Policy()
	if err != nil {
		return fmt.Errorf("error serializing endorsement policy: %v", err)
	}
	if err := ctx.GetStub().SetStateValidationParameter(lockID, policyBytes); err != nil {
		return fmt.Errorf("error setting state validation parameter: %v", err)
	}

	// Update source GO status to LOCKED_CONVERSION
	if err := updateGOStatus(ctx, input.GOAssetID, assets.GOStatusLockedConversion, ownerCollection); err != nil {
		return fmt.Errorf("error updating GO status: %v", err)
	}

	// Create lock receipt for relaying to destination channel
	lockReceipt := &assets.ConversionLockReceipt{
		LockID:                    lockID,
		GOAssetID:                 input.GOAssetID,
		SourceChannel:             sourceChannel,
		SourceCarrier:             sourceCarrier,
		DestinationChannel:        input.DestinationChannel,
		DestinationCarrier:        input.DestinationCarrier,
		ConversionMethod:          input.ConversionMethod,
		ConversionEfficiency:      input.ConversionEfficiency,
		OwnerMSP:                  input.OwnerMSP,
		SourceIssuerMSP:           sourceIssuerMSP,
		LockReceiptHash:           lockReceiptHash,
		TxID:                      txID,
		SourceAmount:              sourceAmount,
		SourceAmountUnit:          sourceAmountUnit,
		SourceEmissions:           sourceEmissions,
		SourceProductionMethod:    sourceProductionMethod,
		SourceDeviceID:            sourceDeviceID,
		SourceCreationDateTime:    sourceCreationDateTime,
		SourceConsumptionDecls:    sourceConsumptionDecls,
		SourceCountryOfOrigin:     publicGO.CountryOfOrigin,
		SourceProductionStart:     publicGO.ProductionPeriodStart,
		SourceProductionEnd:       publicGO.ProductionPeriodEnd,
		SourceSupportScheme:       publicGO.SupportScheme,
		SourceGridConnectionPoint: publicGO.GridConnectionPoint,
	}

	// Emit lifecycle event
	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: util.EventConversionLockCreated,
		AssetID:   lockID,
		GOType:    sourceCarrier,
		Initiator: clientMSP,
		Timestamp: now,
		Details: map[string]interface{}{
			"sourceGOAssetID":      input.GOAssetID,
			"destinationChannel":   input.DestinationChannel,
			"destinationCarrier":   input.DestinationCarrier,
			"conversionMethod":     input.ConversionMethod,
			"conversionEfficiency": input.ConversionEfficiency,
			"ownerMSP":             input.OwnerMSP,
			"lockReceiptHash":      lockReceiptHash,
			"lockReceipt":          lockReceipt, // Full receipt for issuer relay
		},
	})
}

// ================================================================================
// Phase 2: Mint Destination GO from Conversion (Destination Channel)
// ================================================================================

// MintFromConversion creates a destination GO on the target channel using a lock receipt
// relayed from the source channel. This is Phase 2 of the lock-mint-finalize protocol.
//
// Transient input key: "MintFromConversion"
// Required fields: ConversionLockReceipt (full lock receipt from source channel)
//
// Endorsement: Destination issuer only (lock receipt hash proves source consent)
// Returns: Newly created destination GO ID
func (c *ConversionCrossChannelContract) MintFromConversion(ctx contractapi.TransactionContextInterface) error {
	// Only issuers can mint from conversion (relay operation)
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can mint from conversion: %v", err)
	}

	var lockReceipt assets.ConversionLockReceipt
	if err := util.UnmarshalTransient(ctx, "MintFromConversion", &lockReceipt); err != nil {
		return err
	}

	// Verify this is the destination channel
	destChannel := ctx.GetStub().GetChannelID()
	if lockReceipt.DestinationChannel != destChannel {
		return fmt.Errorf("lock receipt destination channel (%s) does not match current channel (%s)",
			lockReceipt.DestinationChannel, destChannel)
	}

	// Verify lock receipt hash
	verifyHash, err := generateConversionLockReceiptHash(
		lockReceipt.LockID,
		lockReceipt.GOAssetID,
		lockReceipt.SourceChannel,
		lockReceipt.DestinationChannel,
		lockReceipt.DestinationCarrier,
		lockReceipt.ConversionEfficiency,
		lockReceipt.OwnerMSP,
		lockReceipt.TxID,
	)
	if err != nil {
		return fmt.Errorf("error verifying lock receipt hash: %v", err)
	}
	if verifyHash != lockReceipt.LockReceiptHash {
		return fmt.Errorf("lock receipt hash mismatch (tampering detected)")
	}

	// Check for idempotency — ensure this lock hasn't already been minted
	mintReceiptKey := assets.PrefixConversionMintReceipt + "_" + lockReceipt.LockReceiptHash
	existingMintJSON, err := ctx.GetStub().GetState(mintReceiptKey)
	if err != nil {
		return fmt.Errorf("error checking for existing mint receipt: %v", err)
	}
	if existingMintJSON != nil {
		return fmt.Errorf("conversion already minted (mint receipt exists: %s)", mintReceiptKey)
	}

	// Get owner's collection on destination channel
	ownerCollection := access.GetCollectionForOrg(lockReceipt.OwnerMSP)

	// Read destination backlog for this carrier
	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	// Calculate destination amount
	destAmount := lockReceipt.SourceAmount * lockReceipt.ConversionEfficiency

	// Mint destination GO based on destination carrier type
	var mintedGOID string
	switch lockReceipt.DestinationCarrier {
	case "hydrogen":
		mintedGOID, err = mintHydrogenFromConversion(ctx, &lockReceipt, destAmount, ownerCollection, now)
	case "electricity":
		mintedGOID, err = mintElectricityFromConversion(ctx, &lockReceipt, destAmount, ownerCollection, now)
	case "biogas":
		mintedGOID, err = mintBiogasFromConversion(ctx, &lockReceipt, destAmount, ownerCollection, now)
	case "heating_cooling":
		mintedGOID, err = mintHeatingCoolingFromConversion(ctx, &lockReceipt, destAmount, ownerCollection, now)
	default:
		return fmt.Errorf("unsupported destination carrier: %s", lockReceipt.DestinationCarrier)
	}

	if err != nil {
		return fmt.Errorf("error minting destination GO: %v", err)
	}

	// Create mint receipt to prevent double-minting
	mintReceipt := &assets.ConversionMintReceipt{
		ReceiptKey:          mintReceiptKey,
		LockID:              lockReceipt.LockID,
		LockReceiptHash:     lockReceipt.LockReceiptHash,
		MintedGOAssetID:     mintedGOID,
		DestinationChannel:  destChannel,
		DestinationCarrier:  lockReceipt.DestinationCarrier,
		MintedAt:            now,
		SourceChannel:       lockReceipt.SourceChannel,
		SourceGOAssetID:     lockReceipt.GOAssetID,
	}

	mintReceiptBytes, err := json.Marshal(mintReceipt)
	if err != nil {
		return fmt.Errorf("error marshaling mint receipt: %v", err)
	}
	if err := ctx.GetStub().PutState(mintReceiptKey, mintReceiptBytes); err != nil {
		return fmt.Errorf("error writing mint receipt: %v", err)
	}

	// Emit lifecycle event
	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: util.EventConversionMintCreated,
		AssetID:   mintedGOID,
		GOType:    lockReceipt.DestinationCarrier,
		Initiator: clientMSP,
		Timestamp: now,
		Details: map[string]interface{}{
			"lockID":           lockReceipt.LockID,
			"sourceChannel":    lockReceipt.SourceChannel,
			"sourceGOAssetID":  lockReceipt.GOAssetID,
			"sourceCarrier":    lockReceipt.SourceCarrier,
			"conversionMethod": lockReceipt.ConversionMethod,
			"sourceAmount":     lockReceipt.SourceAmount,
			"destAmount":       destAmount,
			"efficiency":       lockReceipt.ConversionEfficiency,
			"lockReceiptHash":  lockReceipt.LockReceiptHash,
			"mintReceiptKey":   mintReceiptKey,
		},
	})
}

// ================================================================================
// Phase 3: Finalize Conversion Lock (Source Channel)
// ================================================================================

// FinalizeLock marks the source GO as CONSUMED and updates the conversion lock status.
// This is Phase 3 of the lock-mint-finalize protocol.
//
// Transient input key: "FinalizeLock"
// Required fields:
//   - LockID: Conversion lock to finalize
//   - MintedAssetID: Destination GO asset ID (from Phase 2)
//   - DestinationChannel: Destination channel name (verification)
//   - OwnerMSP: Owner MSP (for tri-party endorsement)
//
// Endorsement: Source issuer + GO owner (tri-party)
// Returns: Updated lock status
func (c *ConversionCrossChannelContract) FinalizeLock(ctx contractapi.TransactionContextInterface) error {
	// Only issuers or owners can finalize
	if err := access.RequireRoleOneOf(ctx, access.RoleIssuer, access.RoleProducer); err != nil {
		return fmt.Errorf("only issuers or producers can finalize conversion locks: %v", err)
	}

	type finalizeInput struct {
		LockID             string `json:"LockID"`
		MintedAssetID      string `json:"MintedAssetID"`
		DestinationChannel string `json:"DestinationChannel"`
		OwnerMSP           string `json:"OwnerMSP"`
	}

	var input finalizeInput
	if err := util.UnmarshalTransient(ctx, "FinalizeLock", &input); err != nil {
		return err
	}

	// Validate inputs
	if err := util.ValidateNonEmpty("LockID", input.LockID); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("MintedAssetID", input.MintedAssetID); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("OwnerMSP", input.OwnerMSP); err != nil {
		return err
	}

	// Read lock record
	lockJSON, err := ctx.GetStub().GetState(input.LockID)
	if err != nil {
		return fmt.Errorf("error reading conversion lock: %v", err)
	}
	if lockJSON == nil {
		return fmt.Errorf("conversion lock %s does not exist", input.LockID)
	}

	var lock assets.ConversionLock
	if err := json.Unmarshal(lockJSON, &lock); err != nil {
		return fmt.Errorf("error unmarshaling conversion lock: %v", err)
	}

	// Verify lock status
	if lock.Status != assets.ConversionLockStatusLocked && lock.Status != assets.ConversionLockStatusApproved {
		return fmt.Errorf("cannot finalize lock with status %s (must be locked or approved)", lock.Status)
	}

	// Verify owner MSP matches
	if lock.OwnerMSP != input.OwnerMSP {
		return fmt.Errorf("owner MSP mismatch (lock: %s, input: %s)", lock.OwnerMSP, input.OwnerMSP)
	}

	// Update source GO status to CONSUMED
	ownerCollection := access.GetCollectionForOrg(lock.OwnerMSP)
	if err := updateGOStatus(ctx, lock.GOAssetID, assets.GOStatusConsumed, ownerCollection); err != nil {
		return fmt.Errorf("error updating source GO status: %v", err)
	}

	// Update lock status to consumed
	lock.Status = assets.ConversionLockStatusConsumed
	lockBytes, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("error marshaling updated lock: %v", err)
	}
	if err := ctx.GetStub().PutState(input.LockID, lockBytes); err != nil {
		return fmt.Errorf("error writing updated lock: %v", err)
	}

	// Emit lifecycle event
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}
	clientMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: util.EventConversionFinalized,
		AssetID:   input.LockID,
		GOType:    lock.SourceCarrier,
		Initiator: clientMSP,
		Timestamp: now,
		Details: map[string]interface{}{
			"sourceGOAssetID":    lock.GOAssetID,
			"mintedAssetID":      input.MintedAssetID,
			"destinationChannel": input.DestinationChannel,
			"destinationCarrier": lock.DestinationCarrier,
			"ownerMSP":           lock.OwnerMSP,
		},
	})
}

// ================================================================================
// Query Functions
// ================================================================================

// GetConversionLock retrieves a conversion lock by ID.
func (c *ConversionCrossChannelContract) GetConversionLock(ctx contractapi.TransactionContextInterface, lockID string) (*assets.ConversionLock, error) {
	lockJSON, err := ctx.GetStub().GetState(lockID)
	if err != nil {
		return nil, fmt.Errorf("error reading conversion lock: %v", err)
	}
	if lockJSON == nil {
		return nil, fmt.Errorf("conversion lock %s does not exist", lockID)
	}

	var lock assets.ConversionLock
	if err := json.Unmarshal(lockJSON, &lock); err != nil {
		return nil, fmt.Errorf("error unmarshaling conversion lock: %v", err)
	}

	return &lock, nil
}

// ListConversionLocks retrieves all conversion locks (paginated).
func (c *ConversionCrossChannelContract) ListConversionLocks(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) ([]*assets.ConversionLock, error) {
	queryString := fmt.Sprintf(`{"selector":{"LockID":{"$regex":"^%s"}}}`, assets.PrefixConversionLock)
	return queryConversionLocks(ctx, queryString, pageSize, bookmark)
}

// ================================================================================
// Helper Functions
// ================================================================================

// generateConversionLockReceiptHash creates a SHA-256 hash of the lock receipt data.
func generateConversionLockReceiptHash(lockID, goAssetID, sourceChannel, destChannel, destCarrier string, efficiency float64, ownerMSP, txID string) (string, error) {
	data := fmt.Sprintf("%s||%s||%s||%s||%s||%.6f||%s||%s",
		lockID, goAssetID, sourceChannel, destChannel, destCarrier, efficiency, ownerMSP, txID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// determineCarrierFromGOID determines the carrier type from the GO asset ID prefix.
func determineCarrierFromGOID(goAssetID string) (string, error) {
	if len(goAssetID) < 3 {
		return "", fmt.Errorf("invalid GO asset ID: %s", goAssetID)
	}
	prefix := goAssetID[:3]
	switch prefix {
	case "eGO":
		return "electricity", nil
	case "hGO":
		return "hydrogen", nil
	case "bGO":
		return "biogas", nil
	case "htG": // heating_cooling GOs start with "htGO"
		return "heating_cooling", nil
	default:
		return "", fmt.Errorf("unknown GO asset ID prefix: %s", prefix)
	}
}

// getChannelIssuerMSP returns the issuer MSP for the current channel by querying the role registry.
func getChannelIssuerMSP(ctx contractapi.TransactionContextInterface) (string, error) {
	return access.GetIssuerMSP(ctx)
}

// updateGOStatus updates the status field of a GO (both public and private data).
func updateGOStatus(ctx contractapi.TransactionContextInterface, goAssetID, newStatus, collection string) error {
	// Read private GO data
	privateJSON, err := ctx.GetStub().GetPrivateData(collection, goAssetID)
	if err != nil {
		return fmt.Errorf("error reading private GO: %v", err)
	}
	if privateJSON == nil {
		return fmt.Errorf("private GO %s does not exist", goAssetID)
	}

	// Unmarshal as generic map to update status field
	var privateData map[string]interface{}
	if err := json.Unmarshal(privateJSON, &privateData); err != nil {
		return fmt.Errorf("error unmarshaling private GO: %v", err)
	}

	// Update status (no change needed in private data for status, it's in public)
	// Read public GO data
	publicJSON, err := ctx.GetStub().GetState(goAssetID)
	if err != nil {
		return fmt.Errorf("error reading public GO: %v", err)
	}
	if publicJSON == nil {
		return fmt.Errorf("public GO %s does not exist", goAssetID)
	}

	var publicData map[string]interface{}
	if err := json.Unmarshal(publicJSON, &publicData); err != nil {
		return fmt.Errorf("error unmarshaling public GO: %v", err)
	}

	publicData["Status"] = newStatus

	updatedPublicJSON, err := json.Marshal(publicData)
	if err != nil {
		return fmt.Errorf("error marshaling updated public GO: %v", err)
	}

	return ctx.GetStub().PutState(goAssetID, updatedPublicJSON)
}

// queryConversionLocks executes a CouchDB query for conversion locks.
func queryConversionLocks(ctx contractapi.TransactionContextInterface, queryString string, pageSize int32, bookmark string) ([]*assets.ConversionLock, error) {
	iterator, _, err := ctx.GetStub().GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer iterator.Close()

	var locks []*assets.ConversionLock
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error iterating query results: %v", err)
		}

		var lock assets.ConversionLock
		if err := json.Unmarshal(queryResponse.Value, &lock); err != nil {
			return nil, fmt.Errorf("error unmarshaling lock: %v", err)
		}

		locks = append(locks, &lock)
	}

	return locks, nil
}

// ================================================================================
// Carrier-Specific Mint Functions (from Lock Receipt + Backlog)
// ================================================================================

// mintHydrogenFromConversion creates a hydrogen GO from a conversion lock receipt.
// Reads hydrogen backlog from THIS channel (destination channel).
func mintHydrogenFromConversion(ctx contractapi.TransactionContextInterface, lockReceipt *assets.ConversionLockReceipt, destAmount float64, ownerCollection string, now int64) (string, error) {
	// Read hydrogen backlog for the owner
	backlogKey := assets.BacklogKeyHydrogen + "_" + lockReceipt.OwnerMSP
	backlogJSON, err := ctx.GetStub().GetPrivateData(ownerCollection, backlogKey)
	if err != nil {
		return "", fmt.Errorf("error reading hydrogen backlog: %v", err)
	}
	if backlogJSON == nil {
		return "", fmt.Errorf("no hydrogen backlog found for %s (required for conversion)", lockReceipt.OwnerMSP)
	}

	var backlog assets.HydrogenBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return "", fmt.Errorf("error unmarshaling hydrogen backlog: %v", err)
	}

	// Verify backlog has sufficient hydrogen production
	// destAmount is in source units (e.g., MWh). Need to convert to kg H2.
	// Assume conversion efficiency already applied: destAmount = sourceAmount * efficiency
	// For electrolysis: sourceAmount (MWh) * efficiency → kg H2
	// Use backlog's impliedKwhPerKilo to determine how much backlog to consume

	impliedKwhPerKilo := backlog.AccumulatedInputMWh / backlog.AccumulatedKilosProduced
	requiredKilos := destAmount / impliedKwhPerKilo // destAmount is in MWh, convert to kg

	if requiredKilos > backlog.AccumulatedKilosProduced {
		return "", fmt.Errorf("insufficient hydrogen backlog (need %.2f kg, have %.2f kg)", requiredKilos, backlog.AccumulatedKilosProduced)
	}

	// Generate hydrogen GO ID
	hGOID, err := assets.GenerateID(ctx, assets.PrefixHGO, 0)
	if err != nil {
		return "", fmt.Errorf("error generating hydrogen GO ID: %v", err)
	}

	// Calculate emissions (backlog emissions + source emissions)
	backlogEmissionsPortion := (requiredKilos / backlog.AccumulatedKilosProduced) * backlog.AccumulatedEmissions
	totalEmissions := backlogEmissionsPortion + lockReceipt.SourceEmissions

	// Create private hydrogen GO
	hGOPrivate := &assets.GreenHydrogenGOPrivateDetails{
		AssetID:                     hGOID,
		OwnerID:                     lockReceipt.OwnerMSP,
		CreationDateTime:            now,
		Kilosproduced:               requiredKilos,
		EmissionsHydrogen:           backlogEmissionsPortion,
		HydrogenProductionMethod:    backlog.HydrogenProductionMethod,
		InputEmissions:              lockReceipt.SourceEmissions,
		UsedMWh:                     destAmount, // Source energy used for conversion
		ElectricityProductionMethod: []string{lockReceipt.SourceProductionMethod},
		ConsumptionDeclarations:     append(lockReceipt.SourceConsumptionDecls, lockReceipt.GOAssetID),
		DeviceID:                    backlog.DeviceID,
	}

	// Generate commitment
	commitment, salt, err := assets.GenerateCommitment(ctx, hGOPrivate.Kilosproduced)
	if err != nil {
		return "", fmt.Errorf("error generating commitment: %v", err)
	}
	hGOPrivate.CommitmentSalt = salt

	// Create public hydrogen GO
	hGOPublic := &assets.GreenHydrogenGO{
		AssetID:               hGOID,
		CreationDateTime:      now,
		GOType:                "Hydrogen",
		Status:                assets.GOStatusActive,
		QuantityCommitment:    commitment,
		CountryOfOrigin:       lockReceipt.SourceCountryOfOrigin,
		EnergySource:          backlog.HydrogenProductionMethod,
		SupportScheme:         lockReceipt.SourceSupportScheme,
		GridConnectionPoint:   lockReceipt.SourceGridConnectionPoint,
		ProductionPeriodStart: backlog.FirstMeteringTimestamp,
		ProductionPeriodEnd:   backlog.LastMeteringTimestamp,
	}

	// Write hydrogen GO to ledger
	if err := util.WriteHGOToLedger(ctx, hGOPublic, hGOPrivate, ownerCollection); err != nil {
		return "", fmt.Errorf("error writing hydrogen GO: %v", err)
	}

	// Update backlog (consume used portion)
	backlog.AccumulatedKilosProduced -= requiredKilos
	backlog.AccumulatedInputMWh -= destAmount
	backlog.AccumulatedEmissions -= backlogEmissionsPortion

	backlogBytes, err := json.Marshal(backlog)
	if err != nil {
		return "", fmt.Errorf("error marshaling updated backlog: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(ownerCollection, backlogKey, backlogBytes); err != nil {
		return "", fmt.Errorf("error writing updated backlog: %v", err)
	}

	return hGOID, nil
}

// mintElectricityFromConversion creates an electricity GO from a conversion lock receipt.
// Reads electricity backlog from THIS channel (destination channel).
func mintElectricityFromConversion(ctx contractapi.TransactionContextInterface, lockReceipt *assets.ConversionLockReceipt, destAmount float64, ownerCollection string, now int64) (string, error) {
	// Read electricity backlog for the owner
	backlogKey := assets.BacklogKeyElectricity + "_" + lockReceipt.OwnerMSP
	backlogJSON, err := ctx.GetStub().GetPrivateData(ownerCollection, backlogKey)
	if err != nil {
		return "", fmt.Errorf("error reading electricity backlog: %v", err)
	}
	if backlogJSON == nil {
		return "", fmt.Errorf("no electricity backlog found for %s (required for conversion)", lockReceipt.OwnerMSP)
	}

	var backlog assets.ElectricityBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return "", fmt.Errorf("error unmarshaling electricity backlog: %v", err)
	}

	// Verify backlog has sufficient electricity production
	if destAmount > backlog.AccumulatedMWh {
		return "", fmt.Errorf("insufficient electricity backlog (need %.2f MWh, have %.2f MWh)", destAmount, backlog.AccumulatedMWh)
	}

	// Generate electricity GO ID
	eGOID, err := assets.GenerateID(ctx, assets.PrefixEGO, 0)
	if err != nil {
		return "", fmt.Errorf("error generating electricity GO ID: %v", err)
	}

	// Calculate emissions (backlog emissions + source emissions)
	backlogEmissionsPortion := (destAmount / backlog.AccumulatedMWh) * backlog.AccumulatedEmissions
	totalEmissions := backlogEmissionsPortion + lockReceipt.SourceEmissions

	// Create private electricity GO
	eGOPrivate := &assets.ElectricityGOPrivateDetails{
		AssetID:                     eGOID,
		OwnerID:                     lockReceipt.OwnerMSP,
		CreationDateTime:            now,
		AmountMWh:                   destAmount,
		Emissions:                   totalEmissions,
		ElectricityProductionMethod: fmt.Sprintf("conversion_%s_to_electricity", lockReceipt.SourceCarrier),
		ConsumptionDeclarations:     append(lockReceipt.SourceConsumptionDecls, lockReceipt.GOAssetID),
		DeviceID:                    backlog.DeviceID,
	}

	// Generate commitment
	commitment, salt, err := assets.GenerateCommitment(ctx, eGOPrivate.AmountMWh)
	if err != nil {
		return "", fmt.Errorf("error generating commitment: %v", err)
	}
	eGOPrivate.CommitmentSalt = salt

	// Create public electricity GO
	eGOPublic := &assets.ElectricityGO{
		AssetID:               eGOID,
		CreationDateTime:      now,
		GOType:                "Electricity",
		Status:                assets.GOStatusActive,
		QuantityCommitment:    commitment,
		CountryOfOrigin:       lockReceipt.SourceCountryOfOrigin,
		EnergySource:          backlog.ElectricityProductionMethod,
		SupportScheme:         lockReceipt.SourceSupportScheme,
		GridConnectionPoint:   lockReceipt.SourceGridConnectionPoint,
		ProductionPeriodStart: backlog.FirstMeteringTimestamp,
		ProductionPeriodEnd:   backlog.LastMeteringTimestamp,
	}

	// Write electricity GO to ledger
	if err := util.WriteEGOToLedger(ctx, eGOPublic, eGOPrivate, ownerCollection); err != nil {
		return "", fmt.Errorf("error writing electricity GO: %v", err)
	}

	// Update backlog (consume used portion)
	backlog.AccumulatedMWh -= destAmount
	backlog.AccumulatedEmissions -= backlogEmissionsPortion

	backlogBytes, err := json.Marshal(backlog)
	if err != nil {
		return "", fmt.Errorf("error marshaling updated backlog: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(ownerCollection, backlogKey, backlogBytes); err != nil {
		return "", fmt.Errorf("error writing updated backlog: %v", err)
	}

	return eGOID, nil
}

// mintBiogasFromConversion creates a biogas GO from a conversion lock receipt.
// Reads biogas backlog from THIS channel (destination channel).
func mintBiogasFromConversion(ctx contractapi.TransactionContextInterface, lockReceipt *assets.ConversionLockReceipt, destAmount float64, ownerCollection string, now int64) (string, error) {
	// Read biogas backlog for the owner
	backlogKey := assets.BacklogKeyBiogas + "_" + lockReceipt.OwnerMSP
	backlogJSON, err := ctx.GetStub().GetPrivateData(ownerCollection, backlogKey)
	if err != nil {
		return "", fmt.Errorf("error reading biogas backlog: %v", err)
	}
	if backlogJSON == nil {
		return "", fmt.Errorf("no biogas backlog found for %s (required for conversion)", lockReceipt.OwnerMSP)
	}

	var backlog assets.BiogasBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return "", fmt.Errorf("error unmarshaling biogas backlog: %v", err)
	}

	// Verify backlog has sufficient biogas production
	// destAmount could be in various units depending on source. For simplicity, assume energy content (MWh).
	if destAmount > backlog.AccumulatedEnergyContentMWh {
		return "", fmt.Errorf("insufficient biogas backlog (need %.2f MWh, have %.2f MWh)", destAmount, backlog.AccumulatedEnergyContentMWh)
	}

	// Generate biogas GO ID
	bGOID, err := assets.GenerateID(ctx, assets.PrefixBGO, 0)
	if err != nil {
		return "", fmt.Errorf("error generating biogas GO ID: %v", err)
	}

	// Calculate emissions
	backlogEmissionsPortion := (destAmount / backlog.AccumulatedEnergyContentMWh) * backlog.AccumulatedEmissions
	totalEmissions := backlogEmissionsPortion + lockReceipt.SourceEmissions

	// Calculate volume (Nm3) from energy content
	// Assume backlog ratio: volumeNm3 / energyContentMWh
	volumeRatio := backlog.AccumulatedVolumeNm3 / backlog.AccumulatedEnergyContentMWh
	destVolumeNm3 := destAmount * volumeRatio

	// Create private biogas GO
	bGOPrivate := &assets.BiogasGOPrivateDetails{
		AssetID:                 bGOID,
		OwnerID:                 lockReceipt.OwnerMSP,
		CreationDateTime:        now,
		VolumeNm3:               destVolumeNm3,
		EnergyContentMWh:        destAmount,
		Emissions:               totalEmissions,
		BiogasProductionMethod:  backlog.BiogasProductionMethod,
		FeedstockType:           backlog.FeedstockType,
		ConsumptionDeclarations: append(lockReceipt.SourceConsumptionDecls, lockReceipt.GOAssetID),
		DeviceID:                backlog.DeviceID,
	}

	// Generate commitment
	commitment, salt, err := assets.GenerateCommitment(ctx, bGOPrivate.VolumeNm3)
	if err != nil {
		return "", fmt.Errorf("error generating commitment: %v", err)
	}
	bGOPrivate.CommitmentSalt = salt

	// Create public biogas GO
	bGOPublic := &assets.BiogasGO{
		AssetID:               bGOID,
		CreationDateTime:      now,
		GOType:                "Biogas",
		Status:                assets.GOStatusActive,
		QuantityCommitment:    commitment,
		CountryOfOrigin:       lockReceipt.SourceCountryOfOrigin,
		EnergySource:          backlog.BiogasProductionMethod,
		SupportScheme:         lockReceipt.SourceSupportScheme,
		GridConnectionPoint:   lockReceipt.SourceGridConnectionPoint,
		ProductionPeriodStart: backlog.FirstMeteringTimestamp,
		ProductionPeriodEnd:   backlog.LastMeteringTimestamp,
	}

	// Write biogas GO to ledger
	if err := util.WriteBGOToLedger(ctx, bGOPublic, bGOPrivate, ownerCollection); err != nil {
		return "", fmt.Errorf("error writing biogas GO: %v", err)
	}

	// Update backlog
	backlog.AccumulatedVolumeNm3 -= destVolumeNm3
	backlog.AccumulatedEnergyContentMWh -= destAmount
	backlog.AccumulatedEmissions -= backlogEmissionsPortion

	backlogBytes, err := json.Marshal(backlog)
	if err != nil {
		return "", fmt.Errorf("error marshaling updated backlog: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(ownerCollection, backlogKey, backlogBytes); err != nil {
		return "", fmt.Errorf("error writing updated backlog: %v", err)
	}

	return bGOID, nil
}

// mintHeatingCoolingFromConversion creates a heating/cooling GO from a conversion lock receipt.
// Reads heating/cooling backlog from THIS channel (destination channel).
func mintHeatingCoolingFromConversion(ctx contractapi.TransactionContextInterface, lockReceipt *assets.ConversionLockReceipt, destAmount float64, ownerCollection string, now int64) (string, error) {
	// Read heating/cooling backlog for the owner
	backlogKey := assets.BacklogKeyHeatingCooling + "_" + lockReceipt.OwnerMSP
	backlogJSON, err := ctx.GetStub().GetPrivateData(ownerCollection, backlogKey)
	if err != nil {
		return "", fmt.Errorf("error reading heating/cooling backlog: %v", err)
	}
	if backlogJSON == nil {
		return "", fmt.Errorf("no heating/cooling backlog found for %s (required for conversion)", lockReceipt.OwnerMSP)
	}

	var backlog assets.HeatingCoolingBacklogPrivateDetails
	if err := json.Unmarshal(backlogJSON, &backlog); err != nil {
		return "", fmt.Errorf("error unmarshaling heating/cooling backlog: %v", err)
	}

	// Verify backlog has sufficient heating/cooling production
	if destAmount > backlog.AccumulatedAmountMWh {
		return "", fmt.Errorf("insufficient heating/cooling backlog (need %.2f MWh, have %.2f MWh)", destAmount, backlog.AccumulatedAmountMWh)
	}

	// Generate heating/cooling GO ID
	htGOID, err := assets.GenerateID(ctx, assets.PrefixHTGO, 0)
	if err != nil {
		return "", fmt.Errorf("error generating heating/cooling GO ID: %v", err)
	}

	// Calculate emissions
	backlogEmissionsPortion := (destAmount / backlog.AccumulatedAmountMWh) * backlog.AccumulatedEmissions
	totalEmissions := backlogEmissionsPortion + lockReceipt.SourceEmissions

	// Create private heating/cooling GO
	htGOPrivate := &assets.HeatingCoolingGOPrivateDetails{
		AssetID:                        htGOID,
		OwnerID:                        lockReceipt.OwnerMSP,
		CreationDateTime:               now,
		AmountMWh:                      destAmount,
		Emissions:                      totalEmissions,
		HeatingCoolingProductionMethod: backlog.HeatingCoolingProductionMethod,
		SupplyTemperature:              backlog.AverageSupplyTemperature,
		ConsumptionDeclarations:        append(lockReceipt.SourceConsumptionDecls, lockReceipt.GOAssetID),
		DeviceID:                       backlog.DeviceID,
	}

	// Generate commitment
	commitment, salt, err := assets.GenerateCommitment(ctx, htGOPrivate.AmountMWh)
	if err != nil {
		return "", fmt.Errorf("error generating commitment: %v", err)
	}
	htGOPrivate.CommitmentSalt = salt

	// Create public heating/cooling GO
	htGOPublic := &assets.HeatingCoolingGO{
		AssetID:               htGOID,
		CreationDateTime:      now,
		GOType:                "HeatingCooling",
		Status:                assets.GOStatusActive,
		QuantityCommitment:    commitment,
		CountryOfOrigin:       lockReceipt.SourceCountryOfOrigin,
		EnergySource:          backlog.HeatingCoolingProductionMethod,
		SupportScheme:         lockReceipt.SourceSupportScheme,
		GridConnectionPoint:   lockReceipt.SourceGridConnectionPoint,
		ProductionPeriodStart: backlog.FirstMeteringTimestamp,
		ProductionPeriodEnd:   backlog.LastMeteringTimestamp,
	}

	// Write heating/cooling GO to ledger
	if err := util.WriteHTGOToLedger(ctx, htGOPublic, htGOPrivate, ownerCollection); err != nil {
		return "", fmt.Errorf("error writing heating/cooling GO: %v", err)
	}

	// Update backlog
	backlog.AccumulatedAmountMWh -= destAmount
	backlog.AccumulatedEmissions -= backlogEmissionsPortion

	backlogBytes, err := json.Marshal(backlog)
	if err != nil {
		return "", fmt.Errorf("error marshaling updated backlog: %v", err)
	}
	if err := ctx.GetStub().PutPrivateData(ownerCollection, backlogKey, backlogBytes); err != nil {
		return "", fmt.Errorf("error writing updated backlog: %v", err)
	}

	return htGOID, nil
}
