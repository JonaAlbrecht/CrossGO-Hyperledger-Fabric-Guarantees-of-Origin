package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// VersionInfo holds semantic versioning and feature metadata.
// ADR-013: Clients query this before invoking functions to confirm compatibility.
type VersionInfo struct {
	Version        string   `json:"version"`
	ChaincodeID    string   `json:"chaincodeId"`
	SupportedAPIs  []string `json:"supportedApis"`
	BreakingChange bool     `json:"breakingChange"`
}

// AdminContract provides version information and administrative operations.
// ADR-013: API versioning. ADR-014: Dynamic org onboarding.
type AdminContract struct {
	contractapi.Contract
}

// GetVersion returns the current chaincode version and supported API levels.
// ADR-013: Clients call this before invoking other functions to verify compatibility.
func (c *AdminContract) GetVersion(ctx contractapi.TransactionContextInterface) (*VersionInfo, error) {
	return &VersionInfo{
		Version:     "9.0.0",
		ChaincodeID: "golifecycle",
		SupportedAPIs: []string{
			"issuance/v1",
			"transfer/v1",
			"conversion/v2",
			"cancellation/v1",
			"query/v1",
			"device/v1",
			"admin/v2",
			"bridge/v2",
			"oracle/v1",
			"biogas/v1",
			"heating_cooling/v1",
		},
		BreakingChange: false,
	}, nil
}

// RegisteredOrganization represents an on-chain record of a dynamically registered org.
// ADR-014: Tracks organizations without requiring hardcoded crypto material.
type RegisteredOrganization struct {
	OrgMSP         string   `json:"orgMsp"`
	DisplayName    string   `json:"displayName"`
	OrgType        string   `json:"orgType"` // "issuer", "producer", "buyer"
	EnergyCarriers []string `json:"energyCarriers,omitempty"`
	Country        string   `json:"country,omitempty"`    // ISO 3166-1 alpha-2
	RegisteredAt   int64    `json:"registeredAt"`
	RegisteredBy   string   `json:"registeredBy"`         // Issuer MSP that approved
	Status         string   `json:"status"`               // "active", "suspended"
}

// RegisterOrganization records a new organization on-chain.
// ADR-014: Enables dynamic org onboarding. Only issuers can register new organizations.
// The Fabric channel config update (adding the org's MSP) must happen out-of-band;
// this function records the logical registration for the application layer.
//
// Transient key: "OrgRegistration" containing DisplayName, OrgMSP, OrgType, EnergyCarriers, Country.
func (c *AdminContract) RegisterOrganization(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can register organizations: %v", err)
	}

	type orgInput struct {
		DisplayName    string   `json:"DisplayName"`
		OrgMSP         string   `json:"OrgMSP"`
		OrgType        string   `json:"OrgType"`
		EnergyCarriers []string `json:"EnergyCarriers"`
		Country        string   `json:"Country"`
	}

	var input orgInput
	if err := util.UnmarshalTransient(ctx, "OrgRegistration", &input); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("OrgMSP", input.OrgMSP); err != nil {
		return err
	}
	if err := util.ValidateNonEmpty("OrgType", input.OrgType); err != nil {
		return err
	}

	// Check org doesn't already exist
	orgKey := "org_" + input.OrgMSP
	existing, err := ctx.GetStub().GetState(orgKey)
	if err != nil {
		return fmt.Errorf("error checking for existing org: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("organization %s is already registered", input.OrgMSP)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return err
	}

	org := RegisteredOrganization{
		OrgMSP:         input.OrgMSP,
		DisplayName:    input.DisplayName,
		OrgType:        input.OrgType,
		EnergyCarriers: input.EnergyCarriers,
		Country:        input.Country,
		RegisteredAt:   now,
		RegisteredBy:   issuerMSP,
		Status:         "active",
	}

	orgBytes, err := json.Marshal(org)
	if err != nil {
		return fmt.Errorf("failed to marshal org registration: %v", err)
	}
	if err := ctx.GetStub().PutState(orgKey, orgBytes); err != nil {
		return fmt.Errorf("failed to write org registration: %v", err)
	}

	// ADR-016: Emit event for off-chain indexer
	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "ORG_REGISTERED",
		AssetID:   orgKey,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"orgMsp":  input.OrgMSP,
			"orgType": input.OrgType,
			"country": input.Country,
		},
	})
}

// GetOrganization retrieves the registration record for an organization.
func (c *AdminContract) GetOrganization(ctx contractapi.TransactionContextInterface, orgMSP string) (*RegisteredOrganization, error) {
	orgKey := "org_" + orgMSP
	orgBytes, err := ctx.GetStub().GetState(orgKey)
	if err != nil {
		return nil, fmt.Errorf("error reading org %s: %v", orgMSP, err)
	}
	if orgBytes == nil {
		return nil, fmt.Errorf("organization %s is not registered", orgMSP)
	}
	var org RegisteredOrganization
	if err := json.Unmarshal(orgBytes, &org); err != nil {
		return nil, fmt.Errorf("error unmarshalling org: %v", err)
	}
	return &org, nil
}

// ListOrganizations returns all registered organizations (v9).
func (c *AdminContract) ListOrganizations(ctx contractapi.TransactionContextInterface) ([]*RegisteredOrganization, error) {
	iterator, err := ctx.GetStub().GetStateByRange("org_", "org_~")
	if err != nil {
		return nil, fmt.Errorf("failed to get org iterator: %v", err)
	}
	defer iterator.Close()

	var orgs []*RegisteredOrganization
	for iterator.HasNext() {
		kv, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("iterator error: %v", err)
		}
		var org RegisteredOrganization
		if err := json.Unmarshal(kv.Value, &org); err != nil {
			continue
		}
		orgs = append(orgs, &org)
	}
	return orgs, nil
}
