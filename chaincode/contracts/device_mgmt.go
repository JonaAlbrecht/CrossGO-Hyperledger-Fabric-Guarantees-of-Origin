package contracts

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"

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
// DEPRECATED (ADR-022, v6.0): Use ListDevicesPaginated instead. Will be removed in v8.0.
func (c *DeviceContract) ListDevices(ctx contractapi.TransactionContextInterface) ([]*assets.Device, error) {
	fmt.Println("WARNING: ListDevices is deprecated (ADR-022). Use ListDevicesPaginated.")
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

// VerifyDeviceReading validates that a meter reading was signed by the registered device's
// ECDSA P-256 key (ADR-027, v7.0). The reading payload is:
// SHA-256(deviceId || timestamp || readingMWh || readingType)
// The signature is r || s, each 32 bytes for P-256.
func (c *DeviceContract) VerifyDeviceReading(ctx contractapi.TransactionContextInterface, readingJSON string) (bool, error) {
	var reading assets.DeviceReading
	if err := json.Unmarshal([]byte(readingJSON), &reading); err != nil {
		return false, fmt.Errorf("failed to unmarshal device reading: %v", err)
	}

	// Look up the device
	device, err := c.GetDevice(ctx, reading.DeviceID)
	if err != nil {
		return false, err
	}
	if device.Status != assets.DeviceStatusActive {
		return false, fmt.Errorf("device %s is not active (status: %s)", reading.DeviceID, device.Status)
	}
	if device.PublicKeyPEM == "" {
		return false, fmt.Errorf("device %s does not have a registered public key", reading.DeviceID)
	}

	// Parse the PEM-encoded public key
	block, _ := pem.Decode([]byte(device.PublicKeyPEM))
	if block == nil {
		return false, fmt.Errorf("failed to decode device public key PEM")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse device public key: %v", err)
	}
	ecdsaKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("device public key is not ECDSA")
	}

	// Reconstruct the message digest
	payload := fmt.Sprintf("%s||%d||%f||%s", reading.DeviceID, reading.Timestamp, reading.ReadingMWh, reading.ReadingType)
	hash := sha256.Sum256([]byte(payload))

	// Decode the hex signature (r || s, each 32 bytes for P-256)
	sigBytes, err := hex.DecodeString(reading.SignatureHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature hex: %v", err)
	}
	if len(sigBytes) != 64 {
		return false, fmt.Errorf("invalid signature length: expected 64 bytes, got %d", len(sigBytes))
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])
	valid := ecdsa.Verify(ecdsaKey, hash[:], r, s)

	return valid, nil
}

// SubmitSignedReading records a device-signed meter reading on-chain (ADR-027).
// Transient key: "SignedReading" containing DeviceID, Timestamp, ReadingMWh,
// ReadingType, SignatureHex.
func (c *DeviceContract) SubmitSignedReading(ctx contractapi.TransactionContextInterface) error {
	if err := access.RequireAnyRole(ctx, access.RoleProducer, access.RoleIssuer); err != nil {
		return fmt.Errorf("only producers or issuers can submit signed readings: %v", err)
	}

	var reading assets.DeviceReading
	if err := util.UnmarshalTransient(ctx, "SignedReading", &reading); err != nil {
		return err
	}

	// Verify the signature
	readingJSON, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading for verification: %v", err)
	}
	valid, err := c.VerifyDeviceReading(ctx, string(readingJSON))
	if err != nil {
		return fmt.Errorf("reading verification failed: %v", err)
	}
	if !valid {
		return fmt.Errorf("device reading signature is invalid")
	}

	// Store the verified reading
	readingID, err := assets.GenerateID(ctx, "reading_", 0)
	if err != nil {
		return fmt.Errorf("error generating reading ID: %v", err)
	}

	readingBytes, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("failed to marshal reading: %v", err)
	}
	if err := ctx.GetStub().PutState(readingID, readingBytes); err != nil {
		return fmt.Errorf("failed to write reading to ledger: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return err
	}
	msp, _ := access.GetClientMSPID(ctx)

	return util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "DEVICE_READING_SUBMITTED",
		AssetID:   readingID,
		Initiator: msp,
		Timestamp: now,
		Details: map[string]string{
			"deviceId": reading.DeviceID,
		},
	})
}
