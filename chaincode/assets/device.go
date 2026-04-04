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
	Attributes     map[string]string `json:"attributes"`     // maxEfficiency, emissionIntensity, technologyType, etc.
	PublicKeyPEM   string            `json:"publicKeyPEM,omitempty" metadata:",optional"` // ADR-027 (v7.0): ECDSA P-256 public key for device-signed readings
}

// DeviceReading represents a signed meter reading from an IoT device (ADR-027).
type DeviceReading struct {
	DeviceID       string  `json:"deviceId"`
	Timestamp      int64   `json:"timestamp"`
	ReadingMWh     float64 `json:"readingMWh"`
	ReadingType    string  `json:"readingType"`    // "cumulative" or "interval"
	SignatureHex   string  `json:"signatureHex"`   // ECDSA signature over the reading payload
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
