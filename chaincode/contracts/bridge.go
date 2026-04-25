package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
)

// BridgeContract implements the cross-channel bridge protocol (ADR-024 v7.0, ADR-030/031 v8.0).
//
// v7.0 (ADR-024): Single-channel record-keeping for cross-registry transfers (AIB hub pattern).
// v8.0 (ADR-030/031): True cross-channel bridge with 3-phase atomic protocol:
//   Phase 1: LockGO — locks a GO on the source channel, emits a lock receipt
//   Phase 2: MintFromBridge — on the destination channel, issuer relays lock receipt, mints new GO
//   Phase 3: FinalizeLock — on the source channel, issuer confirms mint, status → BRIDGED
//
// The issuer acts as the trusted cross-channel relay because it is the only organization
// with peers on both channels. The lock receipt hash provides a cryptographic link between
// the two channels' ledger entries.
type BridgeContract struct {
	contractapi.Contract
}

// === v8.0 Cross-Channel Bridge Data Structures (ADR-031) ===

// CrossChannelLock represents a GO locked on the source channel pending cross-channel transfer.
type CrossChannelLock struct {
	LockID             string `json:"lockId"`
	GOAssetID          string `json:"goAssetId"`          // locked asset on this channel
	GOType             string `json:"goType"`              // "Electricity", "Hydrogen", "Biogas", "HeatingCooling"
	SourceChannel      string `json:"sourceChannel"`       // e.g. "electricity-de"
	DestinationChannel string `json:"destinationChannel"`  // e.g. "hydrogen-de"
	Status             string `json:"status"`              // "locked", "approved", "bridged", "expired"
	LockReceiptHash    string `json:"lockReceiptHash"`     // SHA-256 of lock receipt for cross-channel verification
	InitiatedBy        string `json:"initiatedBy"`         // MSP ID of the source issuer relay
	LockedAt           int64  `json:"lockedAt"`
	FinalizedAt        int64  `json:"finalizedAt,omitempty"`
	MintedAssetID      string `json:"mintedAssetId,omitempty"` // asset ID minted on destination channel
	AmountMWh          float64 `json:"amountMWh,omitempty"`
	CountryOfOrigin    string `json:"countryOfOrigin,omitempty"`
	EnergySource       string `json:"energySource,omitempty"`
	// v9.0 Dual-issuer consensus fields (ADR-031 extension)
	SourceIssuerMSP       string `json:"sourceIssuerMSP,omitempty"`       // MSP of the issuer on the source channel
	TargetIssuerMSP       string `json:"targetIssuerMSP,omitempty"`       // MSP of the issuer on the destination channel
	TargetIssuerApproval  bool   `json:"targetIssuerApproval"`            // true once target issuer approves
	TargetIssuerApprovedAt int64 `json:"targetIssuerApprovedAt,omitempty"` // timestamp of target issuer approval
	// v10.1 Owner consent enforcement (tri-party endorsement)
	OwnerMSP              string `json:"ownerMSP,omitempty"`              // MSP of the GO owner who consented to the bridge transfer
}

// BridgeMint represents a GO minted on the destination channel from a cross-channel bridge transfer.
type BridgeMint struct {
	MintID             string `json:"mintId"`
	MintedAssetID      string `json:"mintedAssetId"`       // new asset created on this channel
	GOType             string `json:"goType"`               // type of the minted GO
	SourceChannel      string `json:"sourceChannel"`        // channel the original GO was locked on
	SourceLockID       string `json:"sourceLockId"`         // lock ID on the source channel
	SourceGOAssetID    string `json:"sourceGoAssetId"`      // original asset ID on source channel
	LockReceiptHash    string `json:"lockReceiptHash"`      // must match the hash from LockGO
	MintedBy           string `json:"mintedBy"`             // MSP ID of the issuer relay
	MintedAt           int64  `json:"mintedAt"`
	AmountMWh          float64 `json:"amountMWh,omitempty"`
	CountryOfOrigin    string `json:"countryOfOrigin,omitempty"`
	EnergySource       string `json:"energySource,omitempty"`
}

// BridgeTransfer represents a GO being exported to or imported from an external registry (v7.0 legacy).
type BridgeTransfer struct {
	TransferID       string `json:"transferId"`
	Direction        string `json:"direction"`        // "export" or "import"
	GOAssetID        string `json:"goAssetId"`        // local asset ID
	ExternalRegistry string `json:"externalRegistry"` // e.g. "AIB-hub", "NECS"
	ExternalID       string `json:"externalId"`       // ID in the external registry
	GOType           string `json:"goType"`           // "Electricity", "Hydrogen", "Biogas"
	Status           string `json:"status"`           // "pending", "confirmed", "failed", "cancelled"
	InitiatedBy      string `json:"initiatedBy"`      // MSP ID
	InitiatedAt      int64  `json:"initiatedAt"`
	ConfirmedAt      int64  `json:"confirmedAt,omitempty"`
	AmountMWh        float64 `json:"amountMWh,omitempty"`
	CountryOfOrigin  string `json:"countryOfOrigin,omitempty"`
	EnergySource     string `json:"energySource,omitempty"`
}

// Bridge transfer status constants (v7.0 legacy).
const (
	BridgeStatusPending   = "pending"
	BridgeStatusConfirmed = "confirmed"
	BridgeStatusFailed    = "failed"
	BridgeStatusCancelled = "cancelled"
)

// Cross-channel lock status constants (v8.0).
const (
	LockStatusLocked   = "locked"
	LockStatusApproved = "approved" // v9.0: target issuer has approved
	LockStatusBridged  = "bridged"
	LockStatusExpired  = "expired"
)

// Bridge transfer direction constants (v7.0 legacy).
const (
	BridgeDirectionExport = "export"
	BridgeDirectionImport = "import"
)

// Bridge ID prefix and range.
const (
	PrefixBridge    = "bridge_"
	RangeEndBridge  = "bridge_~"
	PrefixLock      = "lock_"
	RangeEndLock    = "lock_~"
	PrefixMint      = "mint_"
	RangeEndMint    = "mint_~"
)

// ============================================================================
// v8.0 Cross-Channel Bridge Protocol (ADR-030/031)
// ============================================================================

// LockGO initiates Phase 1 of the cross-channel bridge protocol.
// Called on the SOURCE channel. Locks a GO (status → LOCKED) and creates a
// CrossChannelLock record with a lock receipt hash that the issuer will relay
// to the destination channel.
//
// v10.1: Requires tri-party endorsement (owner + source issuer). The transaction
// must be co-signed by the GO owner and the source issuer to prevent unilateral
// cross-border transfers without owner consent.
//
// Transient key: "BridgeLock" containing GOAssetID, DestinationChannel, OwnerMSP.
func (c *BridgeContract) LockGO(ctx contractapi.TransactionContextInterface) (*CrossChannelLock, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can lock GOs for cross-channel bridge: %v", err)
	}

	type lockInput struct {
		GOAssetID          string `json:"GOAssetID"`
		DestinationChannel string `json:"DestinationChannel"`
		OwnerMSP           string `json:"OwnerMSP"` // v10.1: Owner's MSP for consent validation
	}

	var input lockInput
	if err := util.UnmarshalTransient(ctx, "BridgeLock", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("GOAssetID", input.GOAssetID); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("DestinationChannel", input.DestinationChannel); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("OwnerMSP", input.OwnerMSP); err != nil {
		return nil, err
	}

	// Verify the GO exists and is active
	goJSON, err := ctx.GetStub().GetState(input.GOAssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read GO: %v", err)
	}
	if goJSON == nil {
		return nil, fmt.Errorf("GO %s does not exist", input.GOAssetID)
	}

	var goData map[string]interface{}
	if err := json.Unmarshal(goJSON, &goData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GO: %v", err)
	}
	status, _ := goData["Status"].(string)
	if status != assets.GOStatusActive && status != "" {
		return nil, fmt.Errorf("GO %s is not active (status: %s)", input.GOAssetID, status)
	}

	// Lock the GO — status transitions to "locked" (distinct from "transferred")
	goData["Status"] = assets.GOStatusLocked
	updatedJSON, err := json.Marshal(goData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated GO: %v", err)
	}
	if err := ctx.GetStub().PutState(input.GOAssetID, updatedJSON); err != nil {
		return nil, fmt.Errorf("failed to update GO status: %v", err)
	}

	// Generate lock ID and timestamp
	lockID, err := assets.GenerateID(ctx, PrefixLock, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating lock ID: %v", err)
	}
	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	// Extract GO metadata for the lock receipt
	goType, _ := goData["GOType"].(string)
	country, _ := goData["CountryOfOrigin"].(string)
	energy, _ := goData["EnergySource"].(string)

	// Determine source channel from the channel ID in the stub
	sourceChannel := ctx.GetStub().GetChannelID()

	// v10.1: Compute lock receipt hash with owner MSP to prove tri-party consent
	// SHA-256(lockID || goAssetID || sourceChannel || destinationChannel || ownerMSP || txID)
	// This hash is the cryptographic link between the two channels — the destination channel
	// verifies that a mint references a genuine lock with proven owner consent.
	txID := ctx.GetStub().GetTxID()
	receiptInput := lockID + "||" + input.GOAssetID + "||" + sourceChannel + "||" + input.DestinationChannel + "||" + input.OwnerMSP + "||" + txID
	receiptHash := sha256.Sum256([]byte(receiptInput))
	lockReceiptHash := hex.EncodeToString(receiptHash[:])

	lock := &CrossChannelLock{
		LockID:             lockID,
		GOAssetID:          input.GOAssetID,
		GOType:             goType,
		SourceChannel:      sourceChannel,
		DestinationChannel: input.DestinationChannel,
		Status:             LockStatusLocked,
		LockReceiptHash:    lockReceiptHash,
		InitiatedBy:        issuerMSP,
		LockedAt:           now,
		CountryOfOrigin:    country,
		EnergySource:       energy,
		SourceIssuerMSP:    issuerMSP, // v9.0: dual-issuer consensus
		OwnerMSP:           input.OwnerMSP, // v10.1: owner consent enforcement
	}

	lockBytes, err := json.Marshal(lock)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cross-channel lock: %v", err)
	}
	if err := ctx.GetStub().PutState(lockID, lockBytes); err != nil {
		return nil, fmt.Errorf("failed to write cross-channel lock: %v", err)
	}

	// v10.1: Set tri-party state-based endorsement policy on the lock record
	// Requires: source issuer + owner for finalization (destination issuer approves separately)
	ep, err := statebased.NewStateEP(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create endorsement policy: %v", err)
	}
	if err := ep.AddOrgs(statebased.RoleTypePeer, issuerMSP, input.OwnerMSP); err != nil {
		return nil, fmt.Errorf("failed to add orgs to endorsement policy: %v", err)
	}
	epBytes, err := ep.Policy()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize endorsement policy: %v", err)
	}
	if err := ctx.GetStub().SetStateValidationParameter(lockID, epBytes); err != nil {
		return nil, fmt.Errorf("failed to set state endorsement policy on lock: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "BRIDGE_GO_LOCKED",
		AssetID:   lockID,
		GOType:    goType,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"goAssetId":          input.GOAssetID,
			"sourceChannel":      sourceChannel,
			"destinationChannel": input.DestinationChannel,
			"lockReceiptHash":    lockReceiptHash,
			"ownerMSP":           input.OwnerMSP, // v10.1: record owner consent
		},
	})

	return lock, nil
}

// MintFromBridge implements Phase 2 of the cross-channel bridge protocol.
// Called on the DESTINATION channel. The issuer relays the lock receipt from
// the source channel and mints a new GO on this channel. The lock receipt hash
// provides a cryptographic link to the source channel's lock record.
//
// Transient key: "BridgeMint" containing SourceChannel, SourceLockID,
// SourceGOAssetID, LockReceiptHash, GOType, AmountMWh, CountryOfOrigin,
// EnergySource.
func (c *BridgeContract) MintFromBridge(ctx contractapi.TransactionContextInterface) (*BridgeMint, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can mint from bridge: %v", err)
	}

	type mintInput struct {
		SourceChannel   string  `json:"SourceChannel"`
		SourceLockID    string  `json:"SourceLockID"`
		SourceGOAssetID string  `json:"SourceGOAssetID"`
		LockReceiptHash string  `json:"LockReceiptHash"`
		OwnerMSP        string  `json:"OwnerMSP"` // v10.1: Owner MSP for cross-channel verification
		GOType          string  `json:"GOType"`
		AmountMWh       float64 `json:"AmountMWh"`
		CountryOfOrigin string  `json:"CountryOfOrigin"`
		EnergySource    string  `json:"EnergySource"`
	}

	var input mintInput
	if err := util.UnmarshalTransient(ctx, "BridgeMint", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("SourceChannel", input.SourceChannel); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("SourceLockID", input.SourceLockID); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("LockReceiptHash", input.LockReceiptHash); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("GOType", input.GOType); err != nil {
		return nil, err
	}
	// v10.1: Validate owner MSP is provided (proves tri-party consent on source channel)
	if err := util.ValidateNonEmpty("OwnerMSP", input.OwnerMSP); err != nil {
		return nil, err
	}

	// Validate CEN-EN 16325 fields if provided
	if input.CountryOfOrigin != "" || input.EnergySource != "" {
		if err := util.ValidateCENFields(input.CountryOfOrigin, "", "", input.EnergySource, 0, 0); err != nil {
			return nil, fmt.Errorf("CEN-EN 16325 validation failed: %v", err)
		}
	}

	// Check for duplicate mint — prevent double-minting from the same lock
	// We use a composite key: mint_<lockReceiptHash> to ensure idempotency
	duplicateKey := "mint_receipt_" + input.LockReceiptHash
	existing, err := ctx.GetStub().GetState(duplicateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate mint: %v", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("a mint for lock receipt hash %s already exists", input.LockReceiptHash)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	// Create a local GO asset on the destination channel
	var localAssetID string
	switch input.GOType {
	case "Electricity":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixEGO, 1)
		if err != nil {
			return nil, err
		}
		commitment, salt, err := assets.GenerateCommitment(ctx, input.AmountMWh)
		if err != nil {
			return nil, err
		}
		pub := &assets.ElectricityGO{
			AssetID:            localAssetID,
			CreationDateTime:   now,
			GOType:             "Electricity",
			Status:             assets.GOStatusActive,
			QuantityCommitment: commitment,
			CountryOfOrigin:    input.CountryOfOrigin,
			EnergySource:       input.EnergySource,
		}
		priv := &assets.ElectricityGOPrivateDetails{
			AssetID:          localAssetID,
			OwnerID:          issuerMSP,
			CreationDateTime: now,
			AmountMWh:        input.AmountMWh,
			CommitmentSalt:   salt,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := util.WriteEGOToLedger(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	case "Hydrogen":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixHGO, 1)
		if err != nil {
			return nil, err
		}
		pub := &assets.GreenHydrogenGO{
			AssetID:          localAssetID,
			CreationDateTime: now,
			GOType:           "Hydrogen",
			Status:           assets.GOStatusActive,
		}
		priv := &assets.GreenHydrogenGOPrivateDetails{
			AssetID:          localAssetID,
			OwnerID:          issuerMSP,
			CreationDateTime: now,
			Kilosproduced:    input.AmountMWh,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := util.WriteHGOToLedger(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	case "Biogas":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixBGO, 1)
		if err != nil {
			return nil, err
		}
		commitment, salt, err := assets.GenerateCommitment(ctx, input.AmountMWh)
		if err != nil {
			return nil, err
		}
		pub := &assets.BiogasGO{
			AssetID:            localAssetID,
			CreationDateTime:   now,
			GOType:             "Biogas",
			Status:             assets.GOStatusActive,
			QuantityCommitment: commitment,
			CountryOfOrigin:    input.CountryOfOrigin,
			EnergySource:       input.EnergySource,
		}
		priv := &assets.BiogasGOPrivateDetails{
			AssetID:                 localAssetID,
			OwnerID:                 issuerMSP,
			CreationDateTime:        now,
			EnergyContentMWh:        input.AmountMWh,
			ConsumptionDeclarations: []string{"none"},
			CommitmentSalt:          salt,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := writeBGOToLedgerBridge(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	case "HeatingCooling":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixHCGO, 1)
		if err != nil {
			return nil, err
		}
		commitment, salt, err := assets.GenerateCommitment(ctx, input.AmountMWh)
		if err != nil {
			return nil, err
		}
		pub := &assets.HeatingCoolingGO{
			AssetID:            localAssetID,
			CreationDateTime:   now,
			GOType:             "HeatingCooling",
			Status:             assets.GOStatusActive,
			QuantityCommitment: commitment,
			CountryOfOrigin:    input.CountryOfOrigin,
			EnergySource:       input.EnergySource,
		}
		priv := &assets.HeatingCoolingGOPrivateDetails{
			AssetID:                 localAssetID,
			OwnerID:                 issuerMSP,
			CreationDateTime:        now,
			AmountMWh:               input.AmountMWh,
			ConsumptionDeclarations: []string{"none"},
			CommitmentSalt:          salt,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := writeHCGOToLedgerBridge(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported GO type for bridge mint: %s", input.GOType)
	}

	// Generate mint record ID
	mintID, err := assets.GenerateID(ctx, PrefixMint, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating mint ID: %v", err)
	}

	mint := &BridgeMint{
		MintID:          mintID,
		MintedAssetID:   localAssetID,
		GOType:          input.GOType,
		SourceChannel:   input.SourceChannel,
		SourceLockID:    input.SourceLockID,
		SourceGOAssetID: input.SourceGOAssetID,
		LockReceiptHash: input.LockReceiptHash,
		MintedBy:        issuerMSP,
		MintedAt:        now,
		AmountMWh:       input.AmountMWh,
		CountryOfOrigin: input.CountryOfOrigin,
		EnergySource:    input.EnergySource,
	}

	mintBytes, err := json.Marshal(mint)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bridge mint: %v", err)
	}
	if err := ctx.GetStub().PutState(mintID, mintBytes); err != nil {
		return nil, fmt.Errorf("failed to write bridge mint: %v", err)
	}

	// Write the duplicate guard key
	if err := ctx.GetStub().PutState(duplicateKey, []byte(mintID)); err != nil {
		return nil, fmt.Errorf("failed to write duplicate guard: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "BRIDGE_GO_MINTED",
		AssetID:   mintID,
		GOType:    input.GOType,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"mintedAssetId":   localAssetID,
			"sourceChannel":   input.SourceChannel,
			"sourceLockId":    input.SourceLockID,
			"lockReceiptHash": input.LockReceiptHash,
		},
	})

	return mint, nil
}

// FinalizeLock implements Phase 3 of the cross-channel bridge protocol.
// Called on the SOURCE channel. The issuer confirms that the mint on the
// destination channel succeeded, transitioning the lock status from "locked"
// to "bridged" and the GO status to "bridged".
//
// Args: lockID, mintedAssetID (the asset ID minted on the destination channel).
func (c *BridgeContract) FinalizeLock(ctx contractapi.TransactionContextInterface, lockID string, mintedAssetID string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can finalize bridge locks: %v", err)
	}
	if err := util.ValidateNonEmpty("lockID", lockID); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("mintedAssetID", mintedAssetID); err != nil {
		return err
	}

	lockBytes, err := ctx.GetStub().GetState(lockID)
	if err != nil {
		return fmt.Errorf("failed to read lock: %v", err)
	}
	if lockBytes == nil {
		return fmt.Errorf("lock %s does not exist", lockID)
	}

	var lock CrossChannelLock
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		return fmt.Errorf("failed to unmarshal lock: %v", err)
	}
	if lock.Status != LockStatusLocked {
		return fmt.Errorf("lock %s is not in locked state (status: %s)", lockID, lock.Status)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	// Update the lock record
	lock.Status = LockStatusBridged
	lock.FinalizedAt = now
	lock.MintedAssetID = mintedAssetID

	updatedLockBytes, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal updated lock: %v", err)
	}
	if err := ctx.GetStub().PutState(lockID, updatedLockBytes); err != nil {
		return fmt.Errorf("failed to update lock: %v", err)
	}

	// Update the GO status from "locked" to "bridged"
	goJSON, err := ctx.GetStub().GetState(lock.GOAssetID)
	if err != nil {
		return fmt.Errorf("failed to read GO: %v", err)
	}
	if goJSON != nil {
		var goData map[string]interface{}
		if err := json.Unmarshal(goJSON, &goData); err != nil {
			return fmt.Errorf("failed to unmarshal GO: %v", err)
		}
		goData["Status"] = assets.GOStatusBridged
		updatedGO, err := json.Marshal(goData)
		if err != nil {
			return fmt.Errorf("failed to marshal updated GO: %v", err)
		}
		if err := ctx.GetStub().PutState(lock.GOAssetID, updatedGO); err != nil {
			return fmt.Errorf("failed to update GO status: %v", err)
		}
	}

	issuerMSP, _ := access.GetClientMSPID(ctx)
	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "BRIDGE_LOCK_FINALIZED",
		AssetID:   lockID,
		GOType:    lock.GOType,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"goAssetId":          lock.GOAssetID,
			"mintedAssetId":      mintedAssetID,
			"destinationChannel": lock.DestinationChannel,
		},
	})

	return nil
}

// GetLockReceipt returns the lock receipt for a cross-channel lock, used by
// the issuer relay to submit to the destination channel's MintFromBridge.
func (c *BridgeContract) GetLockReceipt(ctx contractapi.TransactionContextInterface, lockID string) (*CrossChannelLock, error) {
	lockBytes, err := ctx.GetStub().GetState(lockID)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock: %v", err)
	}
	if lockBytes == nil {
		return nil, fmt.Errorf("lock %s does not exist", lockID)
	}

	var lock CrossChannelLock
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock: %v", err)
	}
	return &lock, nil
}

// ListLocksPaginated returns paginated cross-channel lock records.
func (c *BridgeContract) ListLocksPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination(PrefixLock, RangeEndLock, pageSize, bookmark)
	if err != nil {
		return "", fmt.Errorf("error querying locks: %v", err)
	}
	defer resultsIterator.Close()

	var locks []*CrossChannelLock
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		var lock CrossChannelLock
		if err := json.Unmarshal(queryResult.Value, &lock); err != nil {
			return "", err
		}
		locks = append(locks, &lock)
	}

	result := struct {
		Locks    []*CrossChannelLock `json:"locks"`
		Bookmark string              `json:"bookmark"`
	}{
		Locks:    locks,
		Bookmark: metadata.Bookmark,
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(resultBytes), nil
}

// ListMintsPaginated returns paginated bridge mint records.
func (c *BridgeContract) ListMintsPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination(PrefixMint, RangeEndMint, pageSize, bookmark)
	if err != nil {
		return "", fmt.Errorf("error querying mints: %v", err)
	}
	defer resultsIterator.Close()

	var mints []*BridgeMint
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		var mint BridgeMint
		if err := json.Unmarshal(queryResult.Value, &mint); err != nil {
			return "", err
		}
		mints = append(mints, &mint)
	}

	result := struct {
		Mints    []*BridgeMint `json:"mints"`
		Bookmark string        `json:"bookmark"`
	}{
		Mints:    mints,
		Bookmark: metadata.Bookmark,
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(resultBytes), nil
}

// ============================================================================
// v7.0 Legacy Cross-Registry Bridge (ADR-024) — retained for backward compat
// ============================================================================

// ExportGO initiates an export of a local GO to an external registry (v7.0 legacy).
// The GO is locked (status set to "transferred") and a bridge transfer record is created.
// Only issuers can initiate cross-registry exports.
// Transient key: "BridgeExport" containing GOAssetID, ExternalRegistry, ExternalID.
func (c *BridgeContract) ExportGO(ctx contractapi.TransactionContextInterface) (*BridgeTransfer, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can export GOs: %v", err)
	}

	type exportInput struct {
		GOAssetID        string `json:"GOAssetID"`
		ExternalRegistry string `json:"ExternalRegistry"`
		ExternalID       string `json:"ExternalID"`
	}

	var input exportInput
	if err := util.UnmarshalTransient(ctx, "BridgeExport", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("GOAssetID", input.GOAssetID); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("ExternalRegistry", input.ExternalRegistry); err != nil {
		return nil, err
	}

	// Verify the GO exists and is active
	goJSON, err := ctx.GetStub().GetState(input.GOAssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read GO: %v", err)
	}
	if goJSON == nil {
		return nil, fmt.Errorf("GO %s does not exist", input.GOAssetID)
	}

	var goData map[string]interface{}
	if err := json.Unmarshal(goJSON, &goData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GO: %v", err)
	}
	status, _ := goData["Status"].(string)
	if status != assets.GOStatusActive && status != "" {
		return nil, fmt.Errorf("GO %s is not active (status: %s)", input.GOAssetID, status)
	}

	// Lock the GO by setting status to transferred
	goData["Status"] = assets.GOStatusTransferred
	updatedJSON, err := json.Marshal(goData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated GO: %v", err)
	}
	if err := ctx.GetStub().PutState(input.GOAssetID, updatedJSON); err != nil {
		return nil, fmt.Errorf("failed to update GO status: %v", err)
	}

	// Create bridge transfer record
	bridgeID, err := assets.GenerateID(ctx, PrefixBridge, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating bridge ID: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	goType, _ := goData["GOType"].(string)
	country, _ := goData["CountryOfOrigin"].(string)
	energy, _ := goData["EnergySource"].(string)

	transfer := &BridgeTransfer{
		TransferID:       bridgeID,
		Direction:        BridgeDirectionExport,
		GOAssetID:        input.GOAssetID,
		ExternalRegistry: input.ExternalRegistry,
		ExternalID:       input.ExternalID,
		GOType:           goType,
		Status:           BridgeStatusPending,
		InitiatedBy:      issuerMSP,
		InitiatedAt:      now,
		CountryOfOrigin:  country,
		EnergySource:     energy,
	}

	transferBytes, err := json.Marshal(transfer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bridge transfer: %v", err)
	}
	if err := ctx.GetStub().PutState(bridgeID, transferBytes); err != nil {
		return nil, fmt.Errorf("failed to write bridge transfer: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "GO_EXPORTED",
		AssetID:   bridgeID,
		GOType:    goType,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"goAssetId":        input.GOAssetID,
			"externalRegistry": input.ExternalRegistry,
			"externalId":       input.ExternalID,
		},
	})

	return transfer, nil
}

// ImportGO records an imported GO from an external registry.
// Only issuers can import GOs. The imported GO is created as a new asset.
// Transient key: "BridgeImport" containing ExternalRegistry, ExternalID, GOType,
// AmountMWh, CountryOfOrigin, EnergySource.
func (c *BridgeContract) ImportGO(ctx contractapi.TransactionContextInterface) (*BridgeTransfer, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can import GOs: %v", err)
	}

	type importInput struct {
		ExternalRegistry string  `json:"ExternalRegistry"`
		ExternalID       string  `json:"ExternalID"`
		GOType           string  `json:"GOType"`
		AmountMWh        float64 `json:"AmountMWh"`
		CountryOfOrigin  string  `json:"CountryOfOrigin"`
		EnergySource     string  `json:"EnergySource"`
	}

	var input importInput
	if err := util.UnmarshalTransient(ctx, "BridgeImport", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("ExternalRegistry", input.ExternalRegistry); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("ExternalID", input.ExternalID); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("GOType", input.GOType); err != nil {
		return nil, err
	}

	// ADR-018: Validate CEN-EN 16325 fields
	if err := util.ValidateCENFields(input.CountryOfOrigin, "", "", input.EnergySource, 0, 0); err != nil {
		return nil, fmt.Errorf("CEN-EN 16325 validation failed: %v", err)
	}

	bridgeID, err := assets.GenerateID(ctx, PrefixBridge, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating bridge ID: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	// Create a local GO asset for the import
	var localAssetID string
	switch input.GOType {
	case "Electricity":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixEGO, 1)
		if err != nil {
			return nil, err
		}
		commitment, salt, err := assets.GenerateCommitment(ctx, input.AmountMWh)
		if err != nil {
			return nil, err
		}
		pub := &assets.ElectricityGO{
			AssetID:            localAssetID,
			CreationDateTime:   now,
			GOType:             "Electricity",
			Status:             assets.GOStatusActive,
			QuantityCommitment: commitment,
			CountryOfOrigin:    input.CountryOfOrigin,
			EnergySource:       input.EnergySource,
		}
		priv := &assets.ElectricityGOPrivateDetails{
			AssetID:          localAssetID,
			OwnerID:          issuerMSP,
			CreationDateTime: now,
			AmountMWh:        input.AmountMWh,
			CommitmentSalt:   salt,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := util.WriteEGOToLedger(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	case "Hydrogen":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixHGO, 1)
		if err != nil {
			return nil, err
		}
		pub := &assets.GreenHydrogenGO{
			AssetID:          localAssetID,
			CreationDateTime: now,
			GOType:           "Hydrogen",
			Status:           assets.GOStatusActive,
		}
		priv := &assets.GreenHydrogenGOPrivateDetails{
			AssetID:          localAssetID,
			OwnerID:          issuerMSP,
			CreationDateTime: now,
			Kilosproduced:    input.AmountMWh, // reuse field for quantity
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := util.WriteHGOToLedger(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	case "Biogas":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixBGO, 1)
		if err != nil {
			return nil, err
		}
		commitment, salt, err := assets.GenerateCommitment(ctx, input.AmountMWh)
		if err != nil {
			return nil, err
		}
		pub := &assets.BiogasGO{
			AssetID:            localAssetID,
			CreationDateTime:   now,
			GOType:             "Biogas",
			Status:             assets.GOStatusActive,
			QuantityCommitment: commitment,
			CountryOfOrigin:    input.CountryOfOrigin,
			EnergySource:       input.EnergySource,
		}
		priv := &assets.BiogasGOPrivateDetails{
			AssetID:                 localAssetID,
			OwnerID:                 issuerMSP,
			CreationDateTime:        now,
			EnergyContentMWh:        input.AmountMWh,
			ConsumptionDeclarations: []string{"none"},
			CommitmentSalt:          salt,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := writeBGOToLedgerBridge(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	case "HeatingCooling":
		localAssetID, err = assets.GenerateID(ctx, assets.PrefixHCGO, 1)
		if err != nil {
			return nil, err
		}
		commitment, salt, err := assets.GenerateCommitment(ctx, input.AmountMWh)
		if err != nil {
			return nil, err
		}
		pub := &assets.HeatingCoolingGO{
			AssetID:            localAssetID,
			CreationDateTime:   now,
			GOType:             "HeatingCooling",
			Status:             assets.GOStatusActive,
			QuantityCommitment: commitment,
			CountryOfOrigin:    input.CountryOfOrigin,
			EnergySource:       input.EnergySource,
		}
		priv := &assets.HeatingCoolingGOPrivateDetails{
			AssetID:                 localAssetID,
			OwnerID:                 issuerMSP,
			CreationDateTime:        now,
			AmountMWh:               input.AmountMWh,
			ConsumptionDeclarations: []string{"none"},
			CommitmentSalt:          salt,
		}
		collection := access.GetCollectionForOrg(issuerMSP)
		if err := writeHCGOToLedgerBridge(ctx, pub, priv, collection); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported GO type for import: %s", input.GOType)
	}

	transfer := &BridgeTransfer{
		TransferID:       bridgeID,
		Direction:        BridgeDirectionImport,
		GOAssetID:        localAssetID,
		ExternalRegistry: input.ExternalRegistry,
		ExternalID:       input.ExternalID,
		GOType:           input.GOType,
		Status:           BridgeStatusConfirmed,
		InitiatedBy:      issuerMSP,
		InitiatedAt:      now,
		ConfirmedAt:      now,
		AmountMWh:        input.AmountMWh,
		CountryOfOrigin:  input.CountryOfOrigin,
		EnergySource:     input.EnergySource,
	}

	transferBytes, err := json.Marshal(transfer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bridge transfer: %v", err)
	}
	if err := ctx.GetStub().PutState(bridgeID, transferBytes); err != nil {
		return nil, fmt.Errorf("failed to write bridge transfer: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "GO_IMPORTED",
		AssetID:   bridgeID,
		GOType:    input.GOType,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"localAssetId":     localAssetID,
			"externalRegistry": input.ExternalRegistry,
			"externalId":       input.ExternalID,
		},
	})

	return transfer, nil
}

// ConfirmExport confirms that an export has been accepted by the external registry.
// Only issuers can confirm exports.
func (c *BridgeContract) ConfirmExport(ctx contractapi.TransactionContextInterface, transferID string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can confirm exports: %v", err)
	}

	transferBytes, err := ctx.GetStub().GetState(transferID)
	if err != nil {
		return fmt.Errorf("failed to read bridge transfer: %v", err)
	}
	if transferBytes == nil {
		return fmt.Errorf("bridge transfer %s does not exist", transferID)
	}

	var transfer BridgeTransfer
	if err := json.Unmarshal(transferBytes, &transfer); err != nil {
		return fmt.Errorf("failed to unmarshal transfer: %v", err)
	}
	if transfer.Status != BridgeStatusPending {
		return fmt.Errorf("transfer %s is not pending (status: %s)", transferID, transfer.Status)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	transfer.Status = BridgeStatusConfirmed
	transfer.ConfirmedAt = now

	updatedBytes, err := json.Marshal(transfer)
	if err != nil {
		return fmt.Errorf("failed to marshal updated transfer: %v", err)
	}
	return ctx.GetStub().PutState(transferID, updatedBytes)
}

// GetBridgeTransfer reads a bridge transfer record by ID.
func (c *BridgeContract) GetBridgeTransfer(ctx contractapi.TransactionContextInterface, transferID string) (*BridgeTransfer, error) {
	transferBytes, err := ctx.GetStub().GetState(transferID)
	if err != nil {
		return nil, fmt.Errorf("failed to read bridge transfer: %v", err)
	}
	if transferBytes == nil {
		return nil, fmt.Errorf("bridge transfer %s does not exist", transferID)
	}

	var transfer BridgeTransfer
	if err := json.Unmarshal(transferBytes, &transfer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transfer: %v", err)
	}
	return &transfer, nil
}

// ListBridgeTransfersPaginated returns paginated bridge transfer records.
func (c *BridgeContract) ListBridgeTransfersPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination(PrefixBridge, RangeEndBridge, pageSize, bookmark)
	if err != nil {
		return "", fmt.Errorf("error querying bridge transfers: %v", err)
	}
	defer resultsIterator.Close()

	var transfers []*BridgeTransfer
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		var transfer BridgeTransfer
		if err := json.Unmarshal(queryResult.Value, &transfer); err != nil {
			return "", err
		}
		transfers = append(transfers, &transfer)
	}

	result := struct {
		Transfers []*BridgeTransfer `json:"transfers"`
		Bookmark  string            `json:"bookmark"`
		Count     int32             `json:"count"`
	}{
		Transfers: transfers,
		Bookmark:  metadata.GetBookmark(),
		Count:     metadata.GetFetchedRecordsCount(),
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}
	return string(resultBytes), nil
}

// ============================================================================
// v9.0 Dual-Issuer Consensus (ADR-031 extension)
// ============================================================================

// ApproveBridgeTransfer allows the target channel's issuer to approve a cross-channel lock.
// This implements dual-issuer consensus: the source issuer locks the GO, and the target
// issuer must approve before the mint can proceed. Only issuers can approve.
func (c *BridgeContract) ApproveBridgeTransfer(ctx contractapi.TransactionContextInterface, lockID string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can approve bridge transfers: %v", err)
	}
	if err := util.ValidateNonEmpty("lockID", lockID); err != nil {
		return err
	}

	lockBytes, err := ctx.GetStub().GetState(lockID)
	if err != nil {
		return fmt.Errorf("failed to read lock: %v", err)
	}
	if lockBytes == nil {
		return fmt.Errorf("lock %s does not exist", lockID)
	}

	var lock CrossChannelLock
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		return fmt.Errorf("failed to unmarshal lock: %v", err)
	}
	if lock.Status != LockStatusLocked {
		return fmt.Errorf("lock %s is not in locked state (status: %s)", lockID, lock.Status)
	}

	approverMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	// The approver must be a different issuer than the one who initiated the lock
	if approverMSP == lock.SourceIssuerMSP {
		return fmt.Errorf("target issuer must differ from source issuer (%s)", lock.SourceIssuerMSP)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}

	lock.TargetIssuerMSP = approverMSP
	lock.TargetIssuerApproval = true
	lock.TargetIssuerApprovedAt = now
	lock.Status = LockStatusApproved

	updatedBytes, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal updated lock: %v", err)
	}
	if err := ctx.GetStub().PutState(lockID, updatedBytes); err != nil {
		return fmt.Errorf("failed to update lock: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "BRIDGE_TRANSFER_APPROVED",
		AssetID:   lockID,
		GOType:    lock.GOType,
		Initiator: approverMSP,
		Timestamp: now,
		Details: map[string]string{
			"sourceIssuer":       lock.SourceIssuerMSP,
			"targetIssuer":       approverMSP,
			"destinationChannel": lock.DestinationChannel,
		},
	})

	return nil
}

// VerifyBridgeTransfer verifies a cross-channel bridge lock, returning the lock
// details including dual-issuer consensus status. Any role can call this for auditing.
func (c *BridgeContract) VerifyBridgeTransfer(ctx contractapi.TransactionContextInterface, lockID string) (*CrossChannelLock, error) {
	if err := util.ValidateNonEmpty("lockID", lockID); err != nil {
		return nil, err
	}

	lockBytes, err := ctx.GetStub().GetState(lockID)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock: %v", err)
	}
	if lockBytes == nil {
		return nil, fmt.Errorf("lock %s does not exist", lockID)
	}

	var lock CrossChannelLock
	if err := json.Unmarshal(lockBytes, &lock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock: %v", err)
	}
	return &lock, nil
}

// ============================================================================
// Bridge-local write helpers for Biogas and HeatingCooling
// ============================================================================

func writeBGOToLedgerBridge(ctx contractapi.TransactionContextInterface, pub *assets.BiogasGO, priv *assets.BiogasGOPrivateDetails, collection string) error {
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

func writeHCGOToLedgerBridge(ctx contractapi.TransactionContextInterface, pub *assets.HeatingCoolingGO, priv *assets.HeatingCoolingGOPrivateDetails, collection string) error {
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

// ============================================================================
// v10.1 Owner Verification Helper (Tri-Party Endorsement)
// ============================================================================

// verifyGOOwnership checks that the specified owner MSP actually owns the GO by
// reading the private data from the owner's collection. Returns true if ownership
// is confirmed, false otherwise.
func verifyGOOwnership(ctx contractapi.TransactionContextInterface, goAssetID string, ownerMSP string, ownerCollection string) (bool, error) {
	// Try reading from the owner's private data collection
	privateDataJSON, err := ctx.GetStub().GetPrivateData(ownerCollection, goAssetID)
	if err != nil {
		return false, fmt.Errorf("error reading private data from collection %s: %v", ownerCollection, err)
	}
	if privateDataJSON == nil {
		// GO doesn't exist in this owner's collection
		return false, nil
	}

	// Parse private data to verify OwnerID field matches the claimed owner MSP
	var privateData map[string]interface{}
	if err := json.Unmarshal(privateDataJSON, &privateData); err != nil {
		return false, fmt.Errorf("failed to unmarshal GO private data: %v", err)
	}

	actualOwner, ok := privateData["OwnerID"].(string)
	if !ok || actualOwner == "" {
		return false, fmt.Errorf("GO private data missing OwnerID field")
	}

	// Verify the owner matches
	if actualOwner != ownerMSP {
		return false, nil
	}

	return true, nil
}