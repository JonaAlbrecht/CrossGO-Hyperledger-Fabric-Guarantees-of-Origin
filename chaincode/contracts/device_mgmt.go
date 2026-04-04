package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// DeviceContract groups device registration and management functions.
type DeviceContract struct {
	contractapi.Contract
}

// RegisterDevice registers a new metering device on-chain. Only issuers can register devices.
// Transient key: "Device" containing DeviceType, OwnerOrgMSP, EnergyCarriers, Attributes.
func (c *DeviceContract) RegisterDevice(ctx contractapi.TransactionContextInterface) (*assets.Device, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can register devices: %v", err)
	}

	type deviceInput struct {
		DeviceType     string            `json:"deviceType"`
		OwnerOrgMSP    string            `json:"ownerOrgMSP"`
		EnergyCarriers []string          `json:"energyCarriers"`
		Attributes     map[string]string `json:"attributes"`
	}

	var input deviceInput
	if err := util.UnmarshalTransient(ctx, "Device", &input); err != nil {
		return nil, err
	}

	if err := util.ValidateNonEmpty("deviceType", input.DeviceType); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("ownerOrgMSP", input.OwnerOrgMSP); err != nil {
		return nil, err
	}
	if input.DeviceType != assets.DeviceTypeSmartMeter && input.DeviceType != assets.DeviceTypeOutputMeter {
		return nil, fmt.Errorf("invalid device type %q: must be SmartMeter or OutputMeter", input.DeviceType)
	}
	if len(input.EnergyCarriers) == 0 {
		return nil, fmt.Errorf("device must specify at least one energy carrier")
	}

	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	deviceID, err := assets.GenerateID(ctx, assets.PrefixDevice, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating device ID: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}

	device := &assets.Device{
		DeviceID:       deviceID,
		DeviceType:     input.DeviceType,
		OwnerOrgMSP:    input.OwnerOrgMSP,
		EnergyCarriers: input.EnergyCarriers,
		Status:         assets.DeviceStatusActive,
		RegisteredBy:   issuerMSP,
		RegisteredAt:   now,
		Attributes:     input.Attributes,
	}

	deviceBytes, err := json.Marshal(device)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device: %v", err)
	}
	if err := ctx.GetStub().PutState(deviceID, deviceBytes); err != nil {
		return nil, fmt.Errorf("failed to write device to ledger: %v", err)
	}

	return device, nil
}

// GetDevice reads a device by ID from the public world state.
func (c *DeviceContract) GetDevice(ctx contractapi.TransactionContextInterface, deviceID string) (*assets.Device, error) {
	deviceJSON, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to read device: %v", err)
	}
	if deviceJSON == nil {
		return nil, fmt.Errorf("device %s does not exist", deviceID)
	}

	var device assets.Device
	if err := json.Unmarshal(deviceJSON, &device); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device: %v", err)
	}
	return &device, nil
}

// ListDevices returns all registered devices.
// Range covers both legacy IDs (device1, device2, ...) and new IDs (device_<hash>).
func (c *DeviceContract) ListDevices(ctx contractapi.TransactionContextInterface) ([]*assets.Device, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("device", "device~")
	if err != nil {
		return nil, fmt.Errorf("error querying devices: %v", err)
	}
	defer resultsIterator.Close()

	var devices []*assets.Device
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var device assets.Device
		if err := json.Unmarshal(queryResult.Value, &device); err != nil {
			return nil, err
		}
		devices = append(devices, &device)
	}
	return devices, nil
}

// ListDevicesPaginated returns a paginated list of registered devices.
// ADR-006: Accepts pageSize and bookmark for cursor-based pagination.
func (c *DeviceContract) ListDevicesPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination("device", "device~", pageSize, bookmark)
	if err != nil {
		return "", fmt.Errorf("error querying devices with pagination: %v", err)
	}
	defer resultsIterator.Close()

	var devices []*assets.Device
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		var device assets.Device
		if err := json.Unmarshal(queryResult.Value, &device); err != nil {
			return "", err
		}
		devices = append(devices, &device)
	}

	result := struct {
		Devices  []*assets.Device `json:"devices"`
		Bookmark string           `json:"bookmark"`
		Count    int32            `json:"count"`
	}{
		Devices:  devices,
		Bookmark: metadata.GetBookmark(),
		Count:    metadata.GetFetchedRecordsCount(),
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal paginated result: %v", err)
	}
	return string(resultBytes), nil
}

// RevokeDevice changes a device's status to "revoked". Only issuers can revoke.
func (c *DeviceContract) RevokeDevice(ctx contractapi.TransactionContextInterface, deviceID string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can revoke devices: %v", err)
	}

	device, err := c.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	device.Status = assets.DeviceStatusRevoked
	deviceBytes, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}
	return ctx.GetStub().PutState(deviceID, deviceBytes)
}

// SuspendDevice changes a device's status to "suspended". Only issuers can suspend.
func (c *DeviceContract) SuspendDevice(ctx contractapi.TransactionContextInterface, deviceID string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can suspend devices: %v", err)
	}

	device, err := c.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	device.Status = assets.DeviceStatusSuspended
	deviceBytes, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}
	return ctx.GetStub().PutState(deviceID, deviceBytes)
}

// ReactivateDevice changes a device's status back to "active". Only issuers can reactivate.
func (c *DeviceContract) ReactivateDevice(ctx contractapi.TransactionContextInterface, deviceID string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can reactivate devices: %v", err)
	}

	device, err := c.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	device.Status = assets.DeviceStatusActive
	deviceBytes, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}
	return ctx.GetStub().PutState(deviceID, deviceBytes)
}

// RegisterOrgRole registers an organization's role in the network. Only issuers can call this.
// This is a bootstrap function — the initial issuer org must be set via chaincode init.
func (c *DeviceContract) RegisterOrgRole(ctx contractapi.TransactionContextInterface, mspID string, role string) error {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return fmt.Errorf("only issuers can register org roles: %v", err)
	}
	return access.RegisterOrgRole(ctx, mspID, role)
}

// InitLedger bootstraps the ledger with the initial issuer organization.
// This should be called once during chaincode initialization.
func (c *DeviceContract) InitLedger(ctx contractapi.TransactionContextInterface, issuerMSP string) error {
	// ADR-004: callers can only register their own org as issuer
	callerMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get caller MSPID: %v", err)
	}
	if callerMSP != issuerMSP {
		return fmt.Errorf("access denied: caller MSP %s cannot register a different org %s as issuer", callerMSP, issuerMSP)
	}
	return access.RegisterOrgRole(ctx, issuerMSP, access.RoleIssuer)
}
