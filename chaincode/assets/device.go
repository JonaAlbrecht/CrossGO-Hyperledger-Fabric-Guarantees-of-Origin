package assets

// Device represents a metering device registered on-chain. This replaces the
// X.509 ABAC attribute approach, allowing devices to be registered and revoked
// at runtime without re-enrolling certificates.
type Device struct {
	DeviceID       string            `json:"deviceID"`
	DeviceType     string            `json:"deviceType"`     // "SmartMeter", "OutputMeter"
	OwnerOrgMSP    string            `json:"ownerOrgMSP"`    // MSP ID of the owning producer
	EnergyCarriers []string          `json:"energyCarriers"` // e.g. ["electricity"], ["hydrogen"]
	Status         string            `json:"status"`         // "active", "revoked", "suspended"
	RegisteredBy   string            `json:"registeredBy"`   // Issuer MSP that approved registration
	RegisteredAt   int64             `json:"registeredAt"`
	Attributes     map[string]string `json:"attributes"` // maxEfficiency, emissionIntensity, technologyType, etc.
}

// Device status constants.
const (
	DeviceStatusActive    = "active"
	DeviceStatusRevoked   = "revoked"
	DeviceStatusSuspended = "suspended"
)

// Device type constants.
const (
	DeviceTypeSmartMeter  = "SmartMeter"
	DeviceTypeOutputMeter = "OutputMeter"
)
