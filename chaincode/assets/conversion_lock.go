package assets

// ================================================================================
// v10.1: Cross-Channel Conversion Lock Types (ADR-033)
// ================================================================================
// These types support lock-mint-finalize protocol for carrier-to-carrier
// conversions across separate Fabric channels (e.g., electricity-de → hydrogen-de).

// ConversionLock represents a locked GO awaiting cross-channel conversion.
// Stored on source channel (e.g., electricity-de) when Phase 1 completes.
type ConversionLock struct {
	LockID               string  `json:"LockID"`               // conversion_lock_<timestamp>_<suffix>
	GOAssetID            string  `json:"GOAssetID"`            // Source GO ID (e.g., eGO_123)
	SourceChannel        string  `json:"SourceChannel"`        // e.g., "electricity-de"
	SourceCarrier        string  `json:"SourceCarrier"`        // e.g., "electricity"
	DestinationChannel   string  `json:"DestinationChannel"`   // e.g., "hydrogen-de"
	DestinationCarrier   string  `json:"DestinationCarrier"`   // e.g., "hydrogen"
	ConversionMethod     string  `json:"ConversionMethod"`     // e.g., "electrolysis", "fuel_cell", "heat_pump"
	ConversionEfficiency float64 `json:"ConversionEfficiency"` // e.g., 0.65 = 65% efficient
	OwnerMSP             string  `json:"OwnerMSP"`             // Owner of the source GO
	SourceIssuerMSP      string  `json:"SourceIssuerMSP"`      // Source channel issuer
	LockReceiptHash      string  `json:"LockReceiptHash"`      // SHA-256 cryptographic proof
	CreatedAt            int64   `json:"CreatedAt"`            // Unix timestamp
	FinalizedAt          int64   `json:"FinalizedAt,omitempty"` // Set when Phase 3 completes
	MintedAssetID        string  `json:"MintedAssetID,omitempty"` // Destination GO ID, set after Phase 3
	Status               string  `json:"Status"`               // "locked", "approved", "consumed", "expired"
}

// ConversionLockReceipt is the data structure relayed to the destination channel
// during Phase 2 (MintFromConversion). It contains all source GO data needed to
// create the destination GO without requiring cross-channel queries.
type ConversionLockReceipt struct {
	LockID               string  `json:"LockID"`
	GOAssetID            string  `json:"GOAssetID"`
	SourceChannel        string  `json:"SourceChannel"`
	SourceCarrier        string  `json:"SourceCarrier"`
	DestinationChannel   string  `json:"DestinationChannel"`
	DestinationCarrier   string  `json:"DestinationCarrier"`
	ConversionMethod     string  `json:"ConversionMethod"`
	ConversionEfficiency float64 `json:"ConversionEfficiency"`
	OwnerMSP             string  `json:"OwnerMSP"`
	SourceIssuerMSP      string  `json:"SourceIssuerMSP"`
	LockReceiptHash      string  `json:"LockReceiptHash"`
	TxID                 string  `json:"TxID"` // Source transaction ID for verification

	// Source GO data (relayed from source channel to avoid cross-channel queries)
	SourceAmount              float64  `json:"SourceAmount"`              // Amount in source carrier units
	SourceAmountUnit          string   `json:"SourceAmountUnit"`          // "MWh", "kg", "Nm3", etc.
	SourceEmissions           float64  `json:"SourceEmissions"`           // CO2e emissions from source GO
	SourceProductionMethod    string   `json:"SourceProductionMethod"`    // e.g., "solar_pv", "wind_offshore"
	SourceDeviceID            string   `json:"SourceDeviceID,omitempty"`  // Device that created source GO
	SourceCreationDateTime    int64    `json:"SourceCreationDateTime"`    // When source GO was created
	SourceConsumptionDecls    []string `json:"SourceConsumptionDecls"`    // Consumption history
	SourceCountryOfOrigin     string   `json:"SourceCountryOfOrigin"`     // ISO 3166-1 alpha-2
	SourceProductionStart     int64    `json:"SourceProductionStart"`     // Production period start
	SourceProductionEnd       int64    `json:"SourceProductionEnd"`       // Production period end
	SourceSupportScheme       string   `json:"SourceSupportScheme"`       // FIT, FIP, quota, none
	SourceGridConnectionPoint string   `json:"SourceGridConnectionPoint"` // EIC code
}

// ConversionMintReceipt proves that minting occurred on the destination channel.
// Stored on destination channel after Phase 2 completes. Used to prevent double-minting.
type ConversionMintReceipt struct {
	ReceiptKey          string `json:"ReceiptKey"`          // conversion_mint_receipt_<lockReceiptHash>
	LockID              string `json:"LockID"`              // Reference to source lock
	LockReceiptHash     string `json:"LockReceiptHash"`     // Hash from source lock
	MintedGOAssetID     string `json:"MintedGOAssetID"`     // Newly created destination GO ID
	DestinationChannel  string `json:"DestinationChannel"`  // This channel
	DestinationCarrier  string `json:"DestinationCarrier"`  // Carrier type
	MintedAt            int64  `json:"MintedAt"`            // Unix timestamp
	SourceChannel       string `json:"SourceChannel"`       // Origin channel
	SourceGOAssetID     string `json:"SourceGOAssetID"`     // Origin GO ID
}

// Conversion lock status constants
const (
	ConversionLockStatusLocked   = "locked"   // Phase 1 complete, awaiting approval
	ConversionLockStatusApproved = "approved" // Destination issuer approved, awaiting mint
	ConversionLockStatusConsumed = "consumed" // Phase 3 complete, conversion finalized
	ConversionLockStatusExpired  = "expired"  // Lock expired without completion
)

// Conversion lock key prefixes and range ends (for CouchDB range queries).
// Keys are stored as "conversion_lock_<id>" and "conversion_mint_receipt_<hash>".
const (
	PrefixConversionLock             = "conversion_lock_"
	RangeEndConversionLock           = "conversion_lock_~"
	PrefixConversionMintReceipt      = "conversion_mint_receipt_"
	RangeEndConversionMintReceipt    = "conversion_mint_receipt_~"
)
